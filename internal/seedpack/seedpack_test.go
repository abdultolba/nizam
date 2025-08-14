package seedpack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/paths"
	"github.com/abdultolba/nizam/internal/snapshot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSeedPackManifest tests the SeedPackManifest structure and methods
func TestSeedPackManifest(t *testing.T) {
	// Create a test snapshot manifest
	snapshotManifest := &snapshot.SnapshotManifest{
		Service:     "test-service",
		Engine:      "postgres",
		Image:       "postgres:15",
		CreatedAt:   time.Now(),
		Compression: "zstd",
		Files: []snapshot.SnapshotFile{
			{Name: "dump.sql", Sha256: "abc123", Size: 1024},
		},
	}

	// Create seed pack manifest
	manifest := NewSeedPackManifest("test-pack", "Test Pack", "A test seed pack", "Test Author", snapshotManifest)

	// Test basic properties
	assert.Equal(t, "test-pack", manifest.Name)
	assert.Equal(t, "Test Pack", manifest.DisplayName)
	assert.Equal(t, "A test seed pack", manifest.Description)
	assert.Equal(t, "Test Author", manifest.Author)
	assert.Equal(t, "1.0.0", manifest.Version)
	assert.Equal(t, "postgres", manifest.Engine)
	assert.Equal(t, int64(1024), manifest.DataSize)

	// Test adding tags
	manifest.AddTag("sql")
	manifest.AddTag("database")
	manifest.AddTag("sql") // duplicate should be ignored
	assert.Equal(t, []string{"sql", "database"}, manifest.Tags)

	// Test adding examples
	manifest.AddExample("List users", "Get all users from the database", "SELECT * FROM users;", "Returns all user records")
	assert.Len(t, manifest.Examples, 1)
	assert.Equal(t, "List users", manifest.Examples[0].Title)

	// Test adding dependencies
	manifest.AddDependency("postgres", "service", "15", false)
	assert.Len(t, manifest.Dependencies, 1)
	assert.Equal(t, "postgres", manifest.Dependencies[0].Name)

	// Test validation
	assert.NoError(t, manifest.Validate())

	// Test invalid manifest
	invalidManifest := SeedPackManifest{}
	assert.Error(t, invalidManifest.Validate())

	// Test GetFullName
	assert.Equal(t, "test-pack@1.0.0", manifest.GetFullName())

	// Test GetDisplayTitle
	assert.Equal(t, "Test Pack", manifest.GetDisplayTitle())

	// Test FormatSize
	assert.NotEmpty(t, manifest.FormatSize())
}

// TestSeedPackInfo tests the SeedPackInfo structure and methods
func TestSeedPackInfo(t *testing.T) {
	info := SeedPackInfo{
		Name:        "test-pack",
		DisplayName: "Test Pack",
		Version:     "1.0.0",
		Author:      "Test Author",
		Engine:      "postgres",
		DataSize:    1024,
		CreatedAt:   time.Now(),
	}

	// Test GetDisplayName
	assert.Equal(t, "Test Pack", info.GetDisplayName())

	// Test GetFullName
	assert.Equal(t, "test-pack@1.0.0", info.GetFullName())

	// Test FormatSize
	assert.NotEmpty(t, info.FormatSize())

	// Test GetAge
	assert.NotEmpty(t, info.GetAge())

	// Test with empty display name
	info.DisplayName = ""
	assert.Equal(t, "test-pack", info.GetDisplayName())
}

// TestSeedPackManifestPersistence tests saving and loading manifests
func TestSeedPackManifestPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "seedpack-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test manifest
	snapshotManifest := &snapshot.SnapshotManifest{
		Service:     "test-service",
		Engine:      "postgres",
		Image:       "postgres:15",
		CreatedAt:   time.Now(),
		Compression: "zstd",
		Files:       []snapshot.SnapshotFile{{Name: "dump.sql", Sha256: "abc123", Size: 1024}},
	}

	manifest := NewSeedPackManifest("test-pack", "Test Pack", "A test seed pack", "Test Author", snapshotManifest)
	manifest.AddTag("test")
	manifest.AddExample("Test query", "A test query", "SELECT 1;", "Returns 1")

	// Save manifest to file
	manifestPath := filepath.Join(tempDir, "seedpack.json")
	err = manifest.WriteToFile(manifestPath)
	require.NoError(t, err)

	// Load manifest from file
	loadedManifest, err := LoadManifestFromFile(manifestPath)
	require.NoError(t, err)

	// Compare loaded manifest with original
	assert.Equal(t, manifest.Name, loadedManifest.Name)
	assert.Equal(t, manifest.DisplayName, loadedManifest.DisplayName)
	assert.Equal(t, manifest.Description, loadedManifest.Description)
	assert.Equal(t, manifest.Engine, loadedManifest.Engine)
	assert.Equal(t, manifest.Tags, loadedManifest.Tags)
	assert.Len(t, loadedManifest.Examples, 1)

	// Test LoadManifestFromDir
	loadedFromDir, err := LoadManifestFromDir(tempDir)
	require.NoError(t, err)
	assert.Equal(t, manifest.Name, loadedFromDir.Name)
}

// TestSeedPackService tests the core functionality (integration test)
func TestSeedPackService(t *testing.T) {
	// Skip if Docker is not available
	docker, err := dockerx.NewClient()
	if err != nil {
		t.Skip("Docker not available, skipping integration test")
	}
	defer docker.Close()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "seedpack-service-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create seed pack service
	service := NewService(docker)
	assert.NotNil(t, service)

	// Test parsePackName function
	name, version := parsePackName("test-pack@1.0.0")
	assert.Equal(t, "test-pack", name)
	assert.Equal(t, "1.0.0", version)

	name, version = parsePackName("test-pack")
	assert.Equal(t, "test-pack", name)
	assert.Equal(t, "", version)

	// Test search functionality with empty result
	packs, err := service.Search(SearchOptions{
		Engine: "nonexistent",
		Query:  "nonexistent",
	})
	require.NoError(t, err)
	assert.Empty(t, packs)

	// Test list functionality with empty result
	packs, err = service.List("")
	require.NoError(t, err)
	assert.Empty(t, packs) // Should be empty since no packs are installed
}

// TestFormatSize tests the size formatting function
func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, test := range tests {
		result := formatSize(test.size)
		assert.Equal(t, test.expected, result, "formatSize(%d) should return %s", test.size, test.expected)
	}
}

// TestCreateOptions tests the CreateOptions structure
func TestCreateOptions(t *testing.T) {
	opts := CreateOptions{
		Name:        "test-pack",
		DisplayName: "Test Pack",
		Description: "A test pack",
		Author:      "Test Author",
		Version:     "1.0.0",
		Tags:        []string{"test", "example"},
		UseCases:    []string{"testing", "examples"},
		Force:       true,
	}

	assert.Equal(t, "test-pack", opts.Name)
	assert.Equal(t, "Test Pack", opts.DisplayName)
	assert.Len(t, opts.Tags, 2)
	assert.Len(t, opts.UseCases, 2)
	assert.True(t, opts.Force)
}

// TestInstallOptions tests the InstallOptions structure
func TestInstallOptions(t *testing.T) {
	opts := InstallOptions{
		Force:  true,
		DryRun: false,
	}

	assert.True(t, opts.Force)
	assert.False(t, opts.DryRun)
}

// TestSearchOptions tests the SearchOptions structure
func TestSearchOptions(t *testing.T) {
	opts := SearchOptions{
		Engine: "postgres",
		Tags:   []string{"database", "sql"},
		Author: "Test Author",
		Query:  "test query",
	}

	assert.Equal(t, "postgres", opts.Engine)
	assert.Len(t, opts.Tags, 2)
	assert.Equal(t, "Test Author", opts.Author)
	assert.Equal(t, "test query", opts.Query)
}

// TestSchemaTypes tests the schema-related structures
func TestSchemaTypes(t *testing.T) {
	// Test TableSchema
	table := TableSchema{
		Name:        "users",
		Description: "User accounts table",
		RowCount:    100,
		Columns: []ColumnInfo{
			{
				Name:        "id",
				Type:        "integer",
				PrimaryKey:  true,
				Nullable:    false,
				Description: "User ID",
			},
			{
				Name:        "email",
				Type:        "varchar(255)",
				PrimaryKey:  false,
				Nullable:    false,
				Description: "User email address",
			},
		},
		Indexes: []IndexInfo{
			{
				Name:    "idx_users_email",
				Columns: []string{"email"},
				Unique:  true,
			},
		},
	}

	assert.Equal(t, "users", table.Name)
	assert.Equal(t, int64(100), table.RowCount)
	assert.Len(t, table.Columns, 2)
	assert.Len(t, table.Indexes, 1)
	assert.True(t, table.Columns[0].PrimaryKey)
	assert.False(t, table.Columns[1].PrimaryKey)

	// Test Collection
	collection := Collection{
		Name:        "posts",
		Description: "Blog posts collection",
		DocCount:    50,
		SampleDoc:   map[string]interface{}{"title": "Test Post", "content": "Test content"},
		Indexes: []IndexInfo{
			{
				Name:    "idx_posts_title",
				Columns: []string{"title"},
				Unique:  false,
			},
		},
	}

	assert.Equal(t, "posts", collection.Name)
	assert.Equal(t, int64(50), collection.DocCount)
	assert.NotNil(t, collection.SampleDoc)
	assert.Len(t, collection.Indexes, 1)

	// Test KeySchema
	keySchema := KeySchema{
		Pattern:     "user:*",
		Type:        "hash",
		Description: "User data stored as hash",
		Count:       25,
		Example:     "user:123",
	}

	assert.Equal(t, "user:*", keySchema.Pattern)
	assert.Equal(t, "hash", keySchema.Type)
	assert.Equal(t, int64(25), keySchema.Count)
}

// TestSeedPackExample tests the SeedPackExample structure
func TestSeedPackExample(t *testing.T) {
	example := SeedPackExample{
		Title:       "Get all users",
		Description: "Retrieves all users from the users table",
		Query:       "SELECT * FROM users;",
		Expected:    "Returns all user records with id, email, and created_at fields",
	}

	assert.Equal(t, "Get all users", example.Title)
	assert.Equal(t, "Retrieves all users from the users table", example.Description)
	assert.Equal(t, "SELECT * FROM users;", example.Query)
	assert.Equal(t, "Returns all user records with id, email, and created_at fields", example.Expected)
}

// TestSeedPackDependency tests the SeedPackDependency structure
func TestSeedPackDependency(t *testing.T) {
	dependency := SeedPackDependency{
		Name:     "postgres",
		Type:     "service",
		Version:  "15",
		Optional: false,
	}

	assert.Equal(t, "postgres", dependency.Name)
	assert.Equal(t, "service", dependency.Type)
	assert.Equal(t, "15", dependency.Version)
	assert.False(t, dependency.Optional)
}

// TestJSON tests JSON marshaling and unmarshaling
func TestJSON(t *testing.T) {
	// Create a test manifest
	snapshotManifest := &snapshot.SnapshotManifest{
		Service:     "test-service",
		Engine:      "postgres",
		Image:       "postgres:15",
		CreatedAt:   time.Now(),
		Compression: "zstd",
		Files:       []snapshot.SnapshotFile{{Name: "dump.sql", Sha256: "abc123", Size: 1024}},
	}

	original := NewSeedPackManifest("test-pack", "Test Pack", "A test seed pack", "Test Author", snapshotManifest)
	original.AddTag("test")
	original.AddExample("Test query", "A test query", "SELECT 1;", "Returns 1")

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal from JSON
	var restored SeedPackManifest
	err = json.Unmarshal(jsonData, &restored)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.Name, restored.Name)
	assert.Equal(t, original.DisplayName, restored.DisplayName)
	assert.Equal(t, original.Engine, restored.Engine)
	assert.Equal(t, original.Tags, restored.Tags)
	assert.Len(t, restored.Examples, 1)
}

// Mock test for paths functionality
func TestMockPaths(t *testing.T) {
	// Test that paths functions exist and are callable
	// These are integration points that would be tested with real file system operations
	_, err := paths.GetSeedsDir()
	assert.NoError(t, err) // Should not error even if directory doesn't exist

	// Test with mock service name
	_, err = paths.GetServiceSeedsDir("test-service")
	assert.NoError(t, err)

	_, err = paths.GetSeedPackDir("test-service", "test-pack")
	assert.NoError(t, err)

	_, err = paths.GetSeedPackVersionDir("test-service", "test-pack", "1.0.0")
	assert.NoError(t, err)
}

// BenchmarkFormatSize benchmarks the formatSize function
func BenchmarkFormatSize(b *testing.B) {
	sizes := []int64{0, 512, 1024, 1048576, 1073741824}
	
	for i := 0; i < b.N; i++ {
		for _, size := range sizes {
			formatSize(size)
		}
	}
}

// BenchmarkNewSeedPackManifest benchmarks seed pack manifest creation
func BenchmarkNewSeedPackManifest(b *testing.B) {
	snapshotManifest := &snapshot.SnapshotManifest{
		Service:     "test-service",
		Engine:      "postgres",
		Image:       "postgres:15",
		CreatedAt:   time.Now(),
		Compression: "zstd",
		Files:       []snapshot.SnapshotFile{{Name: "dump.sql", Sha256: "abc123", Size: 1024}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewSeedPackManifest("test-pack", "Test Pack", "A test seed pack", "Test Author", snapshotManifest)
	}
}
