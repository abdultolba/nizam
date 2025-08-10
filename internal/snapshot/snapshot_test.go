package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/abdultolba/nizam/internal/compress"
)

func TestNewSnapshotManifest(t *testing.T) {
	manifest := NewSnapshotManifest("test-service", "postgres", "postgres:13", "test-tag", "test note", compress.CompNone)

	if manifest.Service != "test-service" {
		t.Errorf("expected service 'test-service', got %q", manifest.Service)
	}
	if manifest.Engine != "postgres" {
		t.Errorf("expected engine 'postgres', got %q", manifest.Engine)
	}
	if manifest.Image != "postgres:13" {
		t.Errorf("expected image 'postgres:13', got %q", manifest.Image)
	}
	if manifest.Tag != "test-tag" {
		t.Errorf("expected tag 'test-tag', got %q", manifest.Tag)
	}
	if manifest.Note != "test note" {
		t.Errorf("expected note 'test note', got %q", manifest.Note)
	}
	if manifest.Compression != "none" {
		t.Errorf("expected compression 'none', got %q", manifest.Compression)
	}
	if len(manifest.Files) != 0 {
		t.Errorf("expected empty files list, got %d files", len(manifest.Files))
	}
	if manifest.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestManifest_AddFile(t *testing.T) {
	manifest := NewSnapshotManifest("test-service", "postgres", "postgres:13", "test-tag", "", compress.CompNone)

	manifest.AddFile("dump.sql", "abc123", 1024)

	if len(manifest.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(manifest.Files))
	}
	if manifest.Files[0].Name != "dump.sql" {
		t.Errorf("expected file name 'dump.sql', got %q", manifest.Files[0].Name)
	}
	if manifest.Files[0].Sha256 != "abc123" {
		t.Errorf("expected file sha256 'abc123', got %q", manifest.Files[0].Sha256)
	}
	if manifest.Files[0].Size != 1024 {
		t.Errorf("expected file size 1024, got %d", manifest.Files[0].Size)
	}
}

func TestWriteAndReadManifest(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "nizam-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manifest
	manifest := NewSnapshotManifest("test-service", "postgres", "postgres:13", "test-tag", "test note", compress.CompGzip)
	manifest.AddFile("dump.sql.gz", "def456", 2048)

	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Write manifest
	err = manifest.WriteToFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to write manifest: %v", err)
	}

	// Read manifest back
	readManifest, err := LoadManifestFromFile(manifestPath)
	if err != nil {
		t.Fatalf("failed to read manifest: %v", err)
	}

	// Verify contents
	if readManifest.Service != manifest.Service {
		t.Errorf("service mismatch: expected %q, got %q", manifest.Service, readManifest.Service)
	}
	if readManifest.Engine != manifest.Engine {
		t.Errorf("engine mismatch: expected %q, got %q", manifest.Engine, readManifest.Engine)
	}
	if readManifest.Tag != manifest.Tag {
		t.Errorf("tag mismatch: expected %q, got %q", manifest.Tag, readManifest.Tag)
	}
	if readManifest.Compression != manifest.Compression {
		t.Errorf("compression mismatch: expected %q, got %q", manifest.Compression, readManifest.Compression)
	}
	if readManifest.Note != manifest.Note {
		t.Errorf("note mismatch: expected %q, got %q", manifest.Note, readManifest.Note)
	}
	if len(readManifest.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(readManifest.Files))
	}
	if readManifest.Files[0].Name != "dump.sql.gz" {
		t.Errorf("file name mismatch: expected 'dump.sql.gz', got %q", readManifest.Files[0].Name)
	}
	if readManifest.Files[0].Size != 2048 {
		t.Errorf("file size mismatch: expected 2048, got %d", readManifest.Files[0].Size)
	}
	if readManifest.Files[0].Sha256 != "def456" {
		t.Errorf("sha256 mismatch: expected 'def456', got %q", readManifest.Files[0].Sha256)
	}
}

func TestManifestValidation(t *testing.T) {
	// Valid manifest
	manifest := NewSnapshotManifest("test-service", "postgres", "postgres:13", "test-tag", "", compress.CompNone)
	manifest.AddFile("dump.sql", "abc123", 1024)

	err := manifest.Validate()
	if err != nil {
		t.Errorf("expected valid manifest to pass validation, got: %v", err)
	}

	// Invalid - no service
	badManifest := manifest
	badManifest.Service = ""
	err = badManifest.Validate()
	if err == nil {
		t.Error("expected validation error for missing service")
	}

	// Invalid - no engine
	badManifest = manifest
	badManifest.Engine = ""
	err = badManifest.Validate()
	if err == nil {
		t.Error("expected validation error for missing engine")
	}

	// Invalid - no files
	badManifest = manifest
	badManifest.Files = []SnapshotFile{}
	err = badManifest.Validate()
	if err == nil {
		t.Error("expected validation error for missing files")
	}
}

func TestSnapshotInfo(t *testing.T) {
	createdAt := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	info := SnapshotInfo{
		Service:   "test-service",
		Tag:       "backup",
		CreatedAt: createdAt,
		Size:      2048000, // 2MB
		Path:      "/tmp/snapshot",
		Engine:    "postgres",
		Note:      "test snapshot",
	}

	// Test GetDisplayName with tag
	displayName := info.GetDisplayName()
	if displayName != "backup" {
		t.Errorf("expected display name 'backup', got %q", displayName)
	}

	// Test GetDisplayName without tag
	info.Tag = ""
	displayName = info.GetDisplayName()
	expected := "2024-01-01 12:00:00"
	if displayName != expected {
		t.Errorf("expected display name %q, got %q", expected, displayName)
	}

	// Test FormatSize
	formatted := info.FormatSize()
	if !strings.Contains(formatted, "MB") {
		t.Errorf("expected size to be formatted in MB, got %q", formatted)
	}
}
