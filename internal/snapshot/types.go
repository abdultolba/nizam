package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/abdultolba/nizam/internal/compress"
	"github.com/abdultolba/nizam/internal/version"
)

// SnapshotManifest represents the metadata for a snapshot
type SnapshotManifest struct {
	Service     string         `json:"service"`
	Engine      string         `json:"engine"`
	Image       string         `json:"image"`
	CreatedAt   time.Time      `json:"createdAt"`
	Tag         string         `json:"tag"`
	ToolVersion string         `json:"toolVersion"`
	Compression string         `json:"compression"`
	Encryption  string         `json:"encryption"`
	Note        string         `json:"note"`
	Files       []SnapshotFile `json:"files"`
}

// SnapshotFile represents a file within a snapshot
type SnapshotFile struct {
	Name   string `json:"name"`
	Sha256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// NewSnapshotManifest creates a new snapshot manifest
func NewSnapshotManifest(service, engine, image, tag, note string, comp compress.Compression) SnapshotManifest {
	return SnapshotManifest{
		Service:     service,
		Engine:      engine,
		Image:       image,
		CreatedAt:   time.Now().UTC(),
		Tag:         tag,
		ToolVersion: version.Version(),
		Compression: comp.String(),
		Encryption:  "none", // TODO: Implement age encryption
		Note:        note,
		Files:       []SnapshotFile{},
	}
}

// AddFile adds a file to the manifest
func (m *SnapshotManifest) AddFile(name, sha256 string, size int64) {
	m.Files = append(m.Files, SnapshotFile{
		Name:   name,
		Sha256: sha256,
		Size:   size,
	})
}

// WriteToFile writes the manifest to a JSON file
func (m *SnapshotManifest) WriteToFile(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// LoadManifestFromFile loads a manifest from a JSON file
func LoadManifestFromFile(path string) (*SnapshotManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest SnapshotManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// LoadManifestFromDir loads a manifest from a snapshot directory
func LoadManifestFromDir(dir string) (*SnapshotManifest, error) {
	manifestPath := filepath.Join(dir, "manifest.json")
	return LoadManifestFromFile(manifestPath)
}

// GetCompression returns the compression type from the manifest
func (m *SnapshotManifest) GetCompression() compress.Compression {
	switch m.Compression {
	case "zstd":
		return compress.CompZstd
	case "gzip":
		return compress.CompGzip
	case "none":
		return compress.CompNone
	default:
		return compress.CompZstd // default
	}
}

// GetMainFile returns the primary data file from the manifest
func (m *SnapshotManifest) GetMainFile() (SnapshotFile, error) {
	if len(m.Files) == 0 {
		return SnapshotFile{}, fmt.Errorf("no files in manifest")
	}
	return m.Files[0], nil // First file is the main data file
}

// Validate checks if the manifest is valid
func (m *SnapshotManifest) Validate() error {
	if m.Service == "" {
		return fmt.Errorf("service name is required")
	}
	if m.Engine == "" {
		return fmt.Errorf("engine is required")
	}
	if len(m.Files) == 0 {
		return fmt.Errorf("at least one file is required")
	}
	return nil
}

// SnapshotInfo holds information about a snapshot for listing
type SnapshotInfo struct {
	Service   string    `json:"service"`
	Tag       string    `json:"tag"`
	CreatedAt time.Time `json:"createdAt"`
	Size      int64     `json:"size"`
	Path      string    `json:"path"`
	Engine    string    `json:"engine"`
	Image     string    `json:"image"`
	Note      string    `json:"note"`
}

// GetDisplayName returns a display name for the snapshot
func (si *SnapshotInfo) GetDisplayName() string {
	if si.Tag != "" {
		return si.Tag
	}
	return si.CreatedAt.Format("2006-01-02 15:04:05")
}

// GetAge returns the age of the snapshot as a human-readable string
func (si *SnapshotInfo) GetAge() string {
	duration := time.Since(si.CreatedAt)

	if duration.Hours() < 1 {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	if duration.Hours() < 24 {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	days := int(duration.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%dd", days)
	}
	if days < 365 {
		return fmt.Sprintf("%dM", days/30)
	}
	return fmt.Sprintf("%dy", days/365)
}

// FormatSize returns a human-readable size string
func (si *SnapshotInfo) FormatSize() string {
	size := float64(si.Size)
	units := []string{"B", "KB", "MB", "GB", "TB"}

	for _, unit := range units {
		if size < 1024 {
			if unit == "B" {
				return fmt.Sprintf("%.0f%s", size, unit)
			}
			return fmt.Sprintf("%.1f%s", size, unit)
		}
		size /= 1024
	}

	return fmt.Sprintf("%.1f%s", size*1024, units[len(units)-1])
}
