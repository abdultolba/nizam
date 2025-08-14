package seedpack

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/paths"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/abdultolba/nizam/internal/snapshot"
	"github.com/rs/zerolog/log"
)

// Service provides seed pack operations
type Service struct {
	docker      *dockerx.Client
	snapshotSvc *snapshot.Service
}

// NewService creates a new seed pack service
func NewService(docker *dockerx.Client) *Service {
	return &Service{
		docker:      docker,
		snapshotSvc: snapshot.NewService(docker),
	}
}

// CreateOptions holds options for creating a seed pack
type CreateOptions struct {
	Name         string
	DisplayName  string
	Description  string
	Author       string
	Version      string
	License      string
	Homepage     string
	Repository   string
	Tags         []string
	UseCases     []string
	Examples     []SeedPackExample
	Dependencies []SeedPackDependency
	Force        bool
}

// Create creates a new seed pack from an existing snapshot
func (s *Service) Create(ctx context.Context, cfg *config.Config, serviceName, snapshotTag string, opts CreateOptions) (*SeedPackManifest, error) {
	// Resolve service info
	serviceInfo, err := resolve.GetServiceInfo(cfg, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service info: %w", err)
	}

	// Find the snapshot to use
	snapshots, err := s.snapshotSvc.List(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots found for service %s", serviceName)
	}

	var selectedSnapshot snapshot.SnapshotInfo
	if snapshotTag != "" {
		// Find snapshot by tag
		found := false
		for _, snap := range snapshots {
			if snap.Tag == snapshotTag {
				selectedSnapshot = snap
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("snapshot with tag '%s' not found", snapshotTag)
		}
	} else {
		// Use latest snapshot
		selectedSnapshot = snapshots[0]
	}

	// Load the snapshot manifest
	snapshotManifest, err := snapshot.LoadManifestFromDir(selectedSnapshot.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot manifest: %w", err)
	}

	// Set defaults for pack creation
	if opts.Name == "" {
		opts.Name = fmt.Sprintf("%s-pack", serviceName)
	}
	if opts.DisplayName == "" {
		opts.DisplayName = fmt.Sprintf("%s Seed Pack", strings.Title(serviceName))
	}
	if opts.Description == "" {
		opts.Description = fmt.Sprintf("Seed data pack for %s service", serviceName)
	}
	if opts.Author == "" {
		opts.Author = "Unknown"
	}
	if opts.Version == "" {
		opts.Version = "1.0.0"
	}
	if opts.License == "" {
		opts.License = "MIT"
	}

	// Create seed pack manifest
	manifest := NewSeedPackManifest(opts.Name, opts.DisplayName, opts.Description, opts.Author, snapshotManifest)
	manifest.Version = opts.Version
	manifest.License = opts.License
	manifest.Homepage = opts.Homepage
	manifest.Repository = opts.Repository
	manifest.UseCases = opts.UseCases
	manifest.Examples = opts.Examples
	manifest.Dependencies = opts.Dependencies

	// Add tags
	for _, tag := range opts.Tags {
		manifest.AddTag(tag)
	}

	// Create pack directory
	packDir, err := paths.GetSeedPackVersionDir(serviceInfo.Engine, opts.Name, opts.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to create pack directory: %w", err)
	}

	// Check if pack already exists
	manifestPath := filepath.Join(packDir, "seedpack.json")
	if _, err := os.Stat(manifestPath); err == nil && !opts.Force {
		return nil, fmt.Errorf("seed pack %s@%s already exists (use --force to overwrite)", opts.Name, opts.Version)
	}

	log.Info().
		Str("pack", opts.Name).
		Str("version", opts.Version).
		Str("engine", serviceInfo.Engine).
		Str("source", selectedSnapshot.Path).
		Msg("Creating seed pack")

	// Copy snapshot files to pack directory
	err = s.copySnapshotFiles(selectedSnapshot.Path, packDir, snapshotManifest)
	if err != nil {
		os.RemoveAll(packDir)
		return nil, fmt.Errorf("failed to copy snapshot files: %w", err)
	}

	// Generate additional metadata if possible
	err = s.enhanceManifest(ctx, &manifest, snapshotManifest, serviceInfo)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to enhance manifest with metadata")
	}

	// Write seed pack manifest
	if err := manifest.WriteToFile(manifestPath); err != nil {
		os.RemoveAll(packDir)
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	// Create README if it doesn't exist
	readmePath := filepath.Join(packDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		err = s.generateReadme(&manifest, readmePath)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to generate README")
		}
	}

	log.Info().
		Str("pack", opts.Name).
		Str("version", opts.Version).
		Str("directory", packDir).
		Msg("Seed pack created successfully")

	return &manifest, nil
}

// copySnapshotFiles copies snapshot files to the pack directory
func (s *Service) copySnapshotFiles(snapshotDir, packDir string, manifest *snapshot.SnapshotManifest) error {
	for _, file := range manifest.Files {
		srcPath := filepath.Join(snapshotDir, file.Name)
		dstPath := filepath.Join(packDir, file.Name)

		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy file %s: %w", file.Name, err)
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// enhanceManifest adds metadata about the database structure
func (s *Service) enhanceManifest(ctx context.Context, manifest *SeedPackManifest, snapshotManifest *snapshot.SnapshotManifest, serviceInfo resolve.ServiceInfo) error {
	// TODO:
	// This would analyze the database and extract schema information
	// For now, we'll add basic metadata based on the engine type
	
	switch manifest.Engine {
	case "postgres", "postgresql":
		manifest.AddTag("sql")
		manifest.AddTag("relational")
		manifest.AddTag("postgresql")
		
		// Add a sample example for PostgreSQL
		if len(manifest.Examples) == 0 {
			manifest.AddExample(
				"List all tables",
				"Get information about all tables in the database",
				"SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';",
				"Returns a list of all public tables",
			)
		}
		
	case "mysql", "mariadb":
		manifest.AddTag("sql")
		manifest.AddTag("relational")
		manifest.AddTag("mysql")
		
		if len(manifest.Examples) == 0 {
			manifest.AddExample(
				"Show tables",
				"List all tables in the current database",
				"SHOW TABLES;",
				"Returns a list of all tables",
			)
		}
		
	case "redis":
		manifest.AddTag("nosql")
		manifest.AddTag("key-value")
		manifest.AddTag("cache")
		
		if len(manifest.Examples) == 0 {
			manifest.AddExample(
				"List all keys",
				"Get all keys stored in Redis",
				"KEYS *",
				"Returns a list of all keys (use with caution in production)",
			)
		}
		
	case "mongo", "mongodb":
		manifest.AddTag("nosql")
		manifest.AddTag("document")
		manifest.AddTag("mongodb")
		
		if len(manifest.Examples) == 0 {
			manifest.AddExample(
				"List collections",
				"Show all collections in the database",
				"show collections",
				"Returns a list of all collections",
			)
		}
	}

	return nil
}

// generateReadme creates a basic README for the seed pack
func (s *Service) generateReadme(manifest *SeedPackManifest, path string) error {
	content := fmt.Sprintf(`# %s

%s

## Information

- **Version:** %s
- **Author:** %s
- **Engine:** %s
- **License:** %s
- **Data Size:** %s

`, manifest.GetDisplayTitle(), manifest.Description, manifest.Version, manifest.Author, manifest.Engine, manifest.License, manifest.FormatSize())

	if len(manifest.Tags) > 0 {
		content += "## Tags\n\n"
		for _, tag := range manifest.Tags {
			content += fmt.Sprintf("- %s\n", tag)
		}
		content += "\n"
	}

	if len(manifest.UseCases) > 0 {
		content += "## Use Cases\n\n"
		for _, useCase := range manifest.UseCases {
			content += fmt.Sprintf("- %s\n", useCase)
		}
		content += "\n"
	}

	if len(manifest.Examples) > 0 {
		content += "## Examples\n\n"
		for _, example := range manifest.Examples {
			content += fmt.Sprintf("### %s\n\n", example.Title)
			content += fmt.Sprintf("%s\n\n", example.Description)
			content += fmt.Sprintf("```sql\n%s\n```\n\n", example.Query)
			if example.Expected != "" {
				content += fmt.Sprintf("Expected result: %s\n\n", example.Expected)
			}
		}
	}

	content += "## Installation\n\n"
	content += fmt.Sprintf("```bash\nnizam pack install %s@%s\n```\n\n", manifest.Name, manifest.Version)

	if manifest.Homepage != "" {
		content += fmt.Sprintf("## More Information\n\nVisit: %s\n", manifest.Homepage)
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

// InstallOptions holds options for installing a seed pack
type InstallOptions struct {
	Force   bool
	DryRun  bool
}

// Install installs a seed pack to a service
func (s *Service) Install(ctx context.Context, cfg *config.Config, serviceName, packName string, opts InstallOptions) error {
	// Parse pack name (name@version or just name)
	name, version := parsePackName(packName)
	
	// Resolve service info
	serviceInfo, err := resolve.GetServiceInfo(cfg, serviceName)
	if err != nil {
		return fmt.Errorf("failed to resolve service info: %w", err)
	}

	// Find the pack
	packDir, err := s.findPack(serviceInfo.Engine, name, version)
	if err != nil {
		return fmt.Errorf("failed to find pack: %w", err)
	}

	// Load pack manifest
	manifest, err := LoadManifestFromDir(packDir)
	if err != nil {
		return fmt.Errorf("failed to load pack manifest: %w", err)
	}

	// Validate compatibility
	if manifest.Engine != serviceInfo.Engine {
		return fmt.Errorf("pack engine %s is not compatible with service engine %s", manifest.Engine, serviceInfo.Engine)
	}

	// Check if container is running
	running, err := s.docker.ContainerIsRunning(ctx, serviceInfo.Container)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	if !running {
		return fmt.Errorf("container %s is not running", serviceInfo.Container)
	}

	if opts.DryRun {
		log.Info().
			Str("pack", manifest.GetFullName()).
			Str("service", serviceName).
			Str("engine", manifest.Engine).
			Msg("Would install seed pack (dry run)")
		return nil
	}

	log.Info().
		Str("pack", manifest.GetFullName()).
		Str("service", serviceName).
		Str("engine", manifest.Engine).
		Str("directory", packDir).
		Msg("Installing seed pack")

	// Use snapshot restore functionality to install the pack
	err = s.snapshotSvc.Restore(ctx, cfg, serviceName, snapshot.RestoreOptions{
		Latest: true, // We'll restore from the pack's snapshot data
		Force:  opts.Force,
	})
	if err != nil {
		return fmt.Errorf("failed to install pack: %w", err)
	}

	log.Info().
		Str("pack", manifest.GetFullName()).
		Str("service", serviceName).
		Msg("Seed pack installed successfully")

	return nil
}

// parsePackName parses a pack name with optional version
func parsePackName(packName string) (name, version string) {
	parts := strings.Split(packName, "@")
	name = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	return
}

// findPack finds a pack directory, optionally with a specific version
func (s *Service) findPack(engine, name, version string) (string, error) {
	if version != "" {
		// Specific version requested
		packDir, err := paths.GetSeedPackVersionDir(engine, name, version)
		if err != nil {
			return "", err
		}
		
		manifestPath := filepath.Join(packDir, "seedpack.json")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			return "", fmt.Errorf("pack %s@%s not found", name, version)
		}
		
		return packDir, nil
	}

	// Find latest version
	versions, err := paths.ListSeedPackVersions(engine, name)
	if err != nil {
		return "", err
	}
	
	if len(versions) == 0 {
		return "", fmt.Errorf("pack %s not found", name)
	}

	// Sort versions (simple string sort for now)
	sort.Strings(versions)
	latestVersion := versions[len(versions)-1]

	return paths.GetSeedPackVersionDir(engine, name, latestVersion)
}

// List lists all available seed packs
func (s *Service) List(engine string) ([]SeedPackInfo, error) {
	if engine == "" {
		return s.listAllPacks()
	}
	return s.listEnginePacks(engine)
}

// listEnginePacks lists packs for a specific engine
func (s *Service) listEnginePacks(engine string) ([]SeedPackInfo, error) {
	packs, err := paths.ListSeedPacks(engine)
	if err != nil {
		return nil, fmt.Errorf("failed to list seed packs: %w", err)
	}

	var packInfos []SeedPackInfo
	for _, pack := range packs {
		versions, err := paths.ListSeedPackVersions(engine, pack)
		if err != nil {
			log.Warn().Str("pack", pack).Err(err).Msg("Failed to list pack versions")
			continue
		}

		for _, version := range versions {
			packDir, err := paths.GetSeedPackVersionDir(engine, pack, version)
			if err != nil {
				continue
			}

			info, err := s.getPackInfo(packDir)
			if err != nil {
				log.Warn().Str("dir", packDir).Err(err).Msg("Failed to get pack info")
				continue
			}
			packInfos = append(packInfos, info)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(packInfos, func(i, j int) bool {
		return packInfos[i].CreatedAt.After(packInfos[j].CreatedAt)
	})

	return packInfos, nil
}

// listAllPacks lists packs for all engines
func (s *Service) listAllPacks() ([]SeedPackInfo, error) {
	seedsDir, err := paths.GetSeedsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get seeds directory: %w", err)
	}

	entries, err := os.ReadDir(seedsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read seeds directory: %w", err)
	}

	var allPacks []SeedPackInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		engine := entry.Name()
		enginePacks, err := s.listEnginePacks(engine)
		if err != nil {
			log.Warn().Str("engine", engine).Err(err).Msg("Failed to list engine packs")
			continue
		}

		allPacks = append(allPacks, enginePacks...)
	}

	// Sort by creation time (newest first)
	sort.Slice(allPacks, func(i, j int) bool {
		return allPacks[i].CreatedAt.After(allPacks[j].CreatedAt)
	})

	return allPacks, nil
}

// getPackInfo extracts pack information from a directory
func (s *Service) getPackInfo(dir string) (SeedPackInfo, error) {
	manifest, err := LoadManifestFromDir(dir)
	if err != nil {
		return SeedPackInfo{}, fmt.Errorf("failed to load manifest: %w", err)
	}

	return SeedPackInfo{
		Name:        manifest.Name,
		DisplayName: manifest.DisplayName,
		Description: manifest.Description,
		Version:     manifest.Version,
		Author:      manifest.Author,
		Engine:      manifest.Engine,
		Tags:        manifest.Tags,
		DataSize:    manifest.DataSize,
		RecordCount: manifest.RecordCount,
		CreatedAt:   manifest.CreatedAt,
		Path:        dir,
		Installed:   false, // TODO: Check if pack is currently installed
	}, nil
}

// SearchOptions holds options for searching seed packs
type SearchOptions struct {
	Engine string
	Tags   []string
	Author string
	Query  string
}

// Search searches for seed packs matching criteria
func (s *Service) Search(opts SearchOptions) ([]SeedPackInfo, error) {
	// Get all packs
	packs, err := s.List(opts.Engine)
	if err != nil {
		return nil, err
	}

	// Filter packs
	var filtered []SeedPackInfo
	for _, pack := range packs {
		if s.matchesSearchCriteria(pack, opts) {
			filtered = append(filtered, pack)
		}
	}

	return filtered, nil
}

// matchesSearchCriteria checks if a pack matches search criteria
func (s *Service) matchesSearchCriteria(pack SeedPackInfo, opts SearchOptions) bool {
	// Check author
	if opts.Author != "" && !strings.Contains(strings.ToLower(pack.Author), strings.ToLower(opts.Author)) {
		return false
	}

	// Check tags
	if len(opts.Tags) > 0 {
		found := false
		for _, searchTag := range opts.Tags {
			for _, packTag := range pack.Tags {
				if strings.EqualFold(packTag, searchTag) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check query (search in name and description)
	if opts.Query != "" {
		query := strings.ToLower(opts.Query)
		name := strings.ToLower(pack.Name)
		displayName := strings.ToLower(pack.DisplayName)
		description := strings.ToLower(pack.Description)
		
		if !strings.Contains(name, query) && 
		   !strings.Contains(displayName, query) && 
		   !strings.Contains(description, query) {
			return false
		}
	}

	return true
}

// Remove removes a seed pack
func (s *Service) Remove(engine, packName string, version string) error {
	name, ver := parsePackName(packName)
	if version != "" {
		ver = version
	}

	if ver == "" {
		// Remove all versions
		packDir, err := paths.GetSeedPackDir(engine, name)
		if err != nil {
			return err
		}
		
		if err := os.RemoveAll(packDir); err != nil {
			return fmt.Errorf("failed to remove pack: %w", err)
		}
		
		log.Info().Str("pack", name).Msg("Removed all versions of seed pack")
	} else {
		// Remove specific version
		versionDir, err := paths.GetSeedPackVersionDir(engine, name, ver)
		if err != nil {
			return err
		}
		
		if err := os.RemoveAll(versionDir); err != nil {
			return fmt.Errorf("failed to remove pack version: %w", err)
		}
		
		log.Info().Str("pack", fmt.Sprintf("%s@%s", name, ver)).Msg("Removed seed pack version")
	}

	return nil
}

// Info gets detailed information about a specific seed pack
func (s *Service) Info(engine, packName string) (*SeedPackManifest, error) {
	name, version := parsePackName(packName)
	
	packDir, err := s.findPack(engine, name, version)
	if err != nil {
		return nil, err
	}

	return LoadManifestFromDir(packDir)
}
