package seedpack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/abdultolba/nizam/internal/snapshot"
	"github.com/abdultolba/nizam/internal/version"
)

// SeedPackManifest represents the metadata for a seed pack
type SeedPackManifest struct {
	// Basic pack information
	Name         string    `json:"name"`
	DisplayName  string    `json:"displayName"`
	Description  string    `json:"description"`
	Version      string    `json:"version"`
	Author       string    `json:"author"`
	License      string    `json:"license"`
	Homepage     string    `json:"homepage"`
	Repository   string    `json:"repository"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	
	// Engine and compatibility
	Engine        string   `json:"engine"`
	EngineVersion string   `json:"engineVersion"`
	Images        []string `json:"images"`
	Tags          []string `json:"tags"`
	
	// Data and structure
	DataSize      int64                   `json:"dataSize"`
	RecordCount   int64                   `json:"recordCount"`
	Schema        *SeedPackSchema         `json:"schema,omitempty"`
	Examples      []SeedPackExample       `json:"examples"`
	UseCases      []string                `json:"useCases"`
	
	// Technical details
	Compression   string                  `json:"compression"`
	Encryption    string                  `json:"encryption"`
	Checksum      string                  `json:"checksum"`
	Dependencies  []SeedPackDependency    `json:"dependencies"`
	
	// Metadata from original snapshot
	SourceSnapshot *snapshot.SnapshotManifest `json:"sourceSnapshot"`
	ToolVersion    string                      `json:"toolVersion"`
	Files          []SeedPackFile              `json:"files"`
}

// SeedPackFile extends SnapshotFile with additional metadata
type SeedPackFile struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // data, schema, examples, docs
	Description string `json:"description"`
	Sha256      string `json:"sha256"`
	Size        int64  `json:"size"`
}

// SeedPackSchema describes the data structure
type SeedPackSchema struct {
	Tables      []TableSchema `json:"tables,omitempty"`
	Collections []Collection  `json:"collections,omitempty"`
	Keys        []KeySchema   `json:"keys,omitempty"`
}

// TableSchema for SQL databases
type TableSchema struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Columns     []ColumnInfo  `json:"columns"`
	RowCount    int64         `json:"rowCount"`
	Indexes     []IndexInfo   `json:"indexes"`
}

// ColumnInfo describes a table column
type ColumnInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Nullable    bool   `json:"nullable"`
	PrimaryKey  bool   `json:"primaryKey"`
	Description string `json:"description"`
}

// IndexInfo describes a table index
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// Collection for NoSQL databases
type Collection struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	DocCount    int64       `json:"docCount"`
	SampleDoc   interface{} `json:"sampleDoc"`
	Indexes     []IndexInfo `json:"indexes"`
}

// KeySchema for key-value stores
type KeySchema struct {
	Pattern     string `json:"pattern"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
	Example     string `json:"example"`
}

// SeedPackExample shows usage examples
type SeedPackExample struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Query       string `json:"query"`
	Expected    string `json:"expected"`
}

// SeedPackDependency describes required services or packs
type SeedPackDependency struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // pack, service, image
	Version string `json:"version"`
	Optional bool  `json:"optional"`
}

// NewSeedPackManifest creates a new seed pack manifest
func NewSeedPackManifest(name, displayName, description, author string, sourceSnapshot *snapshot.SnapshotManifest) SeedPackManifest {
	now := time.Now().UTC()
	
	return SeedPackManifest{
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Version:        "1.0.0",
		Author:         author,
		License:        "MIT",
		CreatedAt:      now,
		UpdatedAt:      now,
		Engine:         sourceSnapshot.Engine,
		Images:         []string{sourceSnapshot.Image},
		Tags:           []string{},
		DataSize:       calculateDataSize(sourceSnapshot),
		Compression:    sourceSnapshot.Compression,
		Encryption:     sourceSnapshot.Encryption,
		SourceSnapshot: sourceSnapshot,
		ToolVersion:    version.Version(),
		Files:          convertSnapshotFiles(sourceSnapshot.Files),
		Examples:       []SeedPackExample{},
		UseCases:       []string{},
		Dependencies:   []SeedPackDependency{},
	}
}

// calculateDataSize calculates total data size from snapshot files
func calculateDataSize(snapshot *snapshot.SnapshotManifest) int64 {
	var total int64
	for _, file := range snapshot.Files {
		total += file.Size
	}
	return total
}

// convertSnapshotFiles converts snapshot files to seed pack files
func convertSnapshotFiles(snapshotFiles []snapshot.SnapshotFile) []SeedPackFile {
	files := make([]SeedPackFile, len(snapshotFiles))
	for i, sf := range snapshotFiles {
		files[i] = SeedPackFile{
			Name:   sf.Name,
			Type:   "data",
			Sha256: sf.Sha256,
			Size:   sf.Size,
		}
	}
	return files
}

// AddFile adds a file to the seed pack
func (m *SeedPackManifest) AddFile(name, fileType, description, sha256 string, size int64) {
	m.Files = append(m.Files, SeedPackFile{
		Name:        name,
		Type:        fileType,
		Description: description,
		Sha256:      sha256,
		Size:        size,
	})
}

// AddExample adds an example to the seed pack
func (m *SeedPackManifest) AddExample(title, description, query, expected string) {
	m.Examples = append(m.Examples, SeedPackExample{
		Title:       title,
		Description: description,
		Query:       query,
		Expected:    expected,
	})
}

// AddTag adds a tag to the seed pack
func (m *SeedPackManifest) AddTag(tag string) {
	for _, existing := range m.Tags {
		if existing == tag {
			return // Tag already exists
		}
	}
	m.Tags = append(m.Tags, tag)
}

// AddDependency adds a dependency to the seed pack
func (m *SeedPackManifest) AddDependency(name, depType, version string, optional bool) {
	m.Dependencies = append(m.Dependencies, SeedPackDependency{
		Name:     name,
		Type:     depType,
		Version:  version,
		Optional: optional,
	})
}

// WriteToFile writes the manifest to a JSON file
func (m *SeedPackManifest) WriteToFile(path string) error {
	m.UpdatedAt = time.Now().UTC()
	
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// LoadManifestFromFile loads a seed pack manifest from a JSON file
func LoadManifestFromFile(path string) (*SeedPackManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest SeedPackManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// LoadManifestFromDir loads a seed pack manifest from a pack directory
func LoadManifestFromDir(dir string) (*SeedPackManifest, error) {
	manifestPath := filepath.Join(dir, "seedpack.json")
	return LoadManifestFromFile(manifestPath)
}

// Validate checks if the seed pack manifest is valid
func (m *SeedPackManifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("pack name is required")
	}
	if m.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}
	if m.Description == "" {
		return fmt.Errorf("description is required")
	}
	if m.Version == "" {
		return fmt.Errorf("version is required")
	}
	if m.Engine == "" {
		return fmt.Errorf("engine is required")
	}
	if len(m.Files) == 0 {
		return fmt.Errorf("at least one file is required")
	}
	if m.SourceSnapshot == nil {
		return fmt.Errorf("source snapshot reference is required")
	}
	return nil
}

// GetMainDataFile returns the primary data file from the manifest
func (m *SeedPackManifest) GetMainDataFile() (SeedPackFile, error) {
	for _, file := range m.Files {
		if file.Type == "data" {
			return file, nil
		}
	}
	return SeedPackFile{}, fmt.Errorf("no data file found in manifest")
}

// GetFullName returns the full name including version
func (m *SeedPackManifest) GetFullName() string {
	return fmt.Sprintf("%s@%s", m.Name, m.Version)
}

// GetDisplayTitle returns a user-friendly title
func (m *SeedPackManifest) GetDisplayTitle() string {
	if m.DisplayName != "" {
		return m.DisplayName
	}
	return m.Name
}

// FormatSize returns a human-readable size string
func (m *SeedPackManifest) FormatSize() string {
	return formatSize(m.DataSize)
}

// GetAge returns the age of the seed pack as a human-readable string
func (m *SeedPackManifest) GetAge() string {
	duration := time.Since(m.CreatedAt)

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

// SeedPackInfo holds information about a seed pack for listing
type SeedPackInfo struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Engine      string    `json:"engine"`
	Tags        []string  `json:"tags"`
	DataSize    int64     `json:"dataSize"`
	RecordCount int64     `json:"recordCount"`
	CreatedAt   time.Time `json:"createdAt"`
	Path        string    `json:"path"`
	Installed   bool      `json:"installed"`
}

// GetDisplayName returns a display name for the pack
func (pi *SeedPackInfo) GetDisplayName() string {
	if pi.DisplayName != "" {
		return pi.DisplayName
	}
	return pi.Name
}

// GetFullName returns the full name including version
func (pi *SeedPackInfo) GetFullName() string {
	return fmt.Sprintf("%s@%s", pi.Name, pi.Version)
}

// FormatSize returns a human-readable size string
func (pi *SeedPackInfo) FormatSize() string {
	return formatSize(pi.DataSize)
}

// GetAge returns the age of the seed pack as a human-readable string
func (pi *SeedPackInfo) GetAge() string {
	duration := time.Since(pi.CreatedAt)

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
