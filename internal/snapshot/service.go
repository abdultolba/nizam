package snapshot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/compress"
	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/paths"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/rs/zerolog/log"
)

// Engine represents a snapshot engine for a specific database type
type Engine interface {
	Create(ctx context.Context, service resolve.ServiceInfo, outputDir string, comp compress.Compression, note string, tag string) (*SnapshotManifest, error)
	Restore(ctx context.Context, service resolve.ServiceInfo, snapshotDir string, manifest *SnapshotManifest, force bool) error
	GetEngineType() string
	CanHandle(engine string) bool
}

// Service provides snapshot operations
type Service struct {
	docker  *dockerx.Client
	engines map[string]Engine
}

// NewService creates a new snapshot service
func NewService(docker *dockerx.Client) *Service {
	engines := make(map[string]Engine)

	// Register engines directly to avoid circular imports
	pgEngine := NewPostgreSQLEngine(docker)
	engines["postgres"] = pgEngine
	engines["postgresql"] = pgEngine

	redisEngine := NewRedisEngine(docker)
	engines["redis"] = redisEngine

	mysqlEngine := NewMySQLEngine(docker)
	engines["mysql"] = mysqlEngine
	engines["mariadb"] = mysqlEngine

	// TODO: Add MongoDB engine
	// mongoEngine := engines.NewMongoEngine(docker)
	// engines["mongo"] = mongoEngine

	return &Service{
		docker:  docker,
		engines: engines,
	}
}

// CreateOptions holds options for creating a snapshot
type CreateOptions struct {
	Tag         string
	Note        string
	Compression compress.Compression
}

// Create creates a snapshot for a service
func (s *Service) Create(ctx context.Context, cfg *config.Config, serviceName string, opts CreateOptions) (*SnapshotManifest, error) {
	// Resolve service info
	serviceInfo, err := resolve.GetServiceInfo(cfg, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service info: %w", err)
	}

	// Get appropriate engine
	engine, exists := s.engines[serviceInfo.Engine]
	if !exists {
		return nil, fmt.Errorf("unsupported engine: %s (TODO: implement %s snapshot support)", serviceInfo.Engine, serviceInfo.Engine)
	}

	// Validate compression
	if !opts.Compression.IsValid() {
		opts.Compression = compress.CompZstd // default
	}

	// Check if container is running
	running, err := s.docker.ContainerIsRunning(ctx, serviceInfo.Container)
	if err != nil {
		return nil, fmt.Errorf("failed to check container status: %w", err)
	}
	if !running {
		return nil, fmt.Errorf("container %s is not running", serviceInfo.Container)
	}

	// Create snapshot directory
	snapshotDir, err := paths.GenerateSnapshotDir(serviceName, opts.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	log.Info().
		Str("service", serviceName).
		Str("engine", serviceInfo.Engine).
		Str("directory", snapshotDir).
		Str("compression", opts.Compression.String()).
		Msg("Creating snapshot")

	// Create snapshot using appropriate engine
	manifest, err := engine.Create(ctx, serviceInfo, snapshotDir, opts.Compression, opts.Note, opts.Tag)
	if err != nil {
		// Cleanup on error
		os.RemoveAll(snapshotDir)
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Write manifest
	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	if err := manifest.WriteToFile(manifestPath); err != nil {
		os.RemoveAll(snapshotDir)
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	log.Info().
		Str("service", serviceName).
		Str("directory", snapshotDir).
		Msg("Snapshot created successfully")

	return manifest, nil
}

// RestoreOptions holds options for restoring a snapshot
type RestoreOptions struct {
	Tag    string
	Latest bool
	Before *time.Time
	Force  bool
}

// Restore restores a snapshot for a service
func (s *Service) Restore(ctx context.Context, cfg *config.Config, serviceName string, opts RestoreOptions) error {
	// Resolve service info
	serviceInfo, err := resolve.GetServiceInfo(cfg, serviceName)
	if err != nil {
		return fmt.Errorf("failed to resolve service info: %w", err)
	}

	// Get appropriate engine
	engine, exists := s.engines[serviceInfo.Engine]
	if !exists {
		return fmt.Errorf("unsupported engine: %s", serviceInfo.Engine)
	}

	// Find snapshot to restore
	snapshotDir, err := s.findSnapshotToRestore(serviceName, opts)
	if err != nil {
		return fmt.Errorf("failed to find snapshot: %w", err)
	}

	// Load manifest
	manifest, err := LoadManifestFromDir(snapshotDir)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Validate manifest
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	log.Info().
		Str("service", serviceName).
		Str("snapshot", snapshotDir).
		Str("created", manifest.CreatedAt.Format("2006-01-02 15:04:05")).
		Msg("Restoring snapshot")

	// Restore using appropriate engine
	if err := engine.Restore(ctx, serviceInfo, snapshotDir, manifest, opts.Force); err != nil {
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	return nil
}

// List lists snapshots for a service
func (s *Service) List(serviceName string) ([]SnapshotInfo, error) {
	if serviceName == "" {
		return s.listAllSnapshots()
	}
	return s.listServiceSnapshots(serviceName)
}

// listServiceSnapshots lists snapshots for a specific service
func (s *Service) listServiceSnapshots(serviceName string) ([]SnapshotInfo, error) {
	snapshotDirs, err := paths.ListSnapshots(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	var snapshots []SnapshotInfo
	for _, dir := range snapshotDirs {
		info, err := s.getSnapshotInfo(dir)
		if err != nil {
			log.Warn().Str("dir", dir).Err(err).Msg("Failed to get snapshot info")
			continue
		}
		snapshots = append(snapshots, info)
	}

	// Sort by creation time (newest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.After(snapshots[j].CreatedAt)
	})

	return snapshots, nil
}

// listAllSnapshots lists snapshots for all services
func (s *Service) listAllSnapshots() ([]SnapshotInfo, error) {
	snapshotsDir, err := paths.GetSnapshotsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots directory: %w", err)
	}

	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	var allSnapshots []SnapshotInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		serviceName := entry.Name()
		serviceSnapshots, err := s.listServiceSnapshots(serviceName)
		if err != nil {
			log.Warn().Str("service", serviceName).Err(err).Msg("Failed to list service snapshots")
			continue
		}

		allSnapshots = append(allSnapshots, serviceSnapshots...)
	}

	// Sort by creation time (newest first)
	sort.Slice(allSnapshots, func(i, j int) bool {
		return allSnapshots[i].CreatedAt.After(allSnapshots[j].CreatedAt)
	})

	return allSnapshots, nil
}

// getSnapshotInfo extracts snapshot information from a directory
func (s *Service) getSnapshotInfo(dir string) (SnapshotInfo, error) {
	manifest, err := LoadManifestFromDir(dir)
	if err != nil {
		return SnapshotInfo{}, fmt.Errorf("failed to load manifest: %w", err)
	}

	// Calculate total size
	var totalSize int64
	for _, file := range manifest.Files {
		totalSize += file.Size
	}

	// Extract tag from directory name if available
	dirName := filepath.Base(dir)
	parts := strings.Split(dirName, "-")
	var tag string
	if len(parts) > 2 {
		// Format: YYYYMMDD-HHMMSS-tag
		tag = strings.Join(parts[2:], "-")
	}

	return SnapshotInfo{
		Service:   manifest.Service,
		Tag:       tag,
		CreatedAt: manifest.CreatedAt,
		Size:      totalSize,
		Path:      dir,
		Engine:    manifest.Engine,
		Image:     manifest.Image,
		Note:      manifest.Note,
	}, nil
}

// findSnapshotToRestore finds the appropriate snapshot directory to restore
func (s *Service) findSnapshotToRestore(serviceName string, opts RestoreOptions) (string, error) {
	snapshots, err := s.listServiceSnapshots(serviceName)
	if err != nil {
		return "", err
	}

	if len(snapshots) == 0 {
		return "", fmt.Errorf("no snapshots found for service %s", serviceName)
	}

	// If tag specified, find exact match
	if opts.Tag != "" {
		for _, snapshot := range snapshots {
			if snapshot.Tag == opts.Tag {
				return snapshot.Path, nil
			}
		}
		return "", fmt.Errorf("snapshot with tag '%s' not found", opts.Tag)
	}

	// If latest requested, return newest
	if opts.Latest {
		return snapshots[0].Path, nil
	}

	// If before time specified, find newest before that time
	if opts.Before != nil {
		for _, snapshot := range snapshots {
			if snapshot.CreatedAt.Before(*opts.Before) {
				return snapshot.Path, nil
			}
		}
		return "", fmt.Errorf("no snapshots found before %s", opts.Before.Format("2006-01-02 15:04:05"))
	}

	// Default to latest
	return snapshots[0].Path, nil
}

// PruneOptions holds options for pruning snapshots
type PruneOptions struct {
	Keep   int
	DryRun bool
}

// Prune removes old snapshots, keeping the N most recent
func (s *Service) Prune(serviceName string, opts PruneOptions) error {
	if opts.Keep <= 0 {
		return fmt.Errorf("keep value must be positive")
	}

	snapshots, err := s.listServiceSnapshots(serviceName)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) <= opts.Keep {
		log.Info().
			Str("service", serviceName).
			Int("total", len(snapshots)).
			Int("keep", opts.Keep).
			Msg("No snapshots to prune")
		return nil
	}

	// Calculate which snapshots to remove (everything after opts.Keep)
	toRemove := snapshots[opts.Keep:]

	var totalSize int64
	for _, snapshot := range toRemove {
		totalSize += snapshot.Size
	}

	if opts.DryRun {
		log.Info().
			Str("service", serviceName).
			Int("count", len(toRemove)).
			Str("size", formatSize(totalSize)).
			Msg("Would remove snapshots (dry run)")

		for _, snapshot := range toRemove {
			log.Info().
				Str("path", snapshot.Path).
				Str("tag", snapshot.Tag).
				Time("created", snapshot.CreatedAt).
				Str("size", snapshot.FormatSize()).
				Msg("Would remove")
		}
		return nil
	}

	// Remove snapshots
	log.Info().
		Str("service", serviceName).
		Int("count", len(toRemove)).
		Str("size", formatSize(totalSize)).
		Msg("Removing old snapshots")

	var removed int
	for _, snapshot := range toRemove {
		if err := os.RemoveAll(snapshot.Path); err != nil {
			log.Warn().Str("path", snapshot.Path).Err(err).Msg("Failed to remove snapshot")
		} else {
			removed++
			log.Debug().Str("path", snapshot.Path).Msg("Removed snapshot")
		}
	}

	log.Info().
		Str("service", serviceName).
		Int("removed", removed).
		Int("kept", len(snapshots)-removed).
		Msg("Prune completed")

	return nil
}

// formatSize formats a size in bytes as a human-readable string
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
