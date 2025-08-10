package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetServiceSnapshotsDir(t *testing.T) {
	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory and use it as working directory
	tmpDir, err := os.MkdirTemp("", "nizam-paths-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test GetServiceSnapshotsDir
	snapDir, err := GetServiceSnapshotsDir("test-service")
	if err != nil {
		t.Fatalf("failed to get service snapshots dir: %v", err)
	}

	// Resolve both paths to handle symlinks (like macOS /var -> /private/var)
	expected := filepath.Join(tmpDir, ".nizam", "snapshots", "test-service")
	resolvedExpected, _ := filepath.EvalSymlinks(expected)
	resolvedActual, _ := filepath.EvalSymlinks(snapDir)

	if resolvedActual != resolvedExpected {
		t.Errorf("expected %q, got %q", resolvedExpected, resolvedActual)
	}

	// Test that it creates the directory
	_, err = os.Stat(snapDir)
	if err != nil {
		t.Errorf("expected directory to be created, got error: %v", err)
	}
}

func TestGetServiceSeedsDir(t *testing.T) {
	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory and use it as working directory
	tmpDir, err := os.MkdirTemp("", "nizam-paths-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test GetServiceSeedsDir
	seedsDir, err := GetServiceSeedsDir("test-service")
	if err != nil {
		t.Fatalf("failed to get service seeds dir: %v", err)
	}

	// Resolve both paths to handle symlinks (like macOS /var -> /private/var)
	expected := filepath.Join(tmpDir, ".nizam", "seeds", "test-service")
	resolvedExpected, _ := filepath.EvalSymlinks(expected)
	resolvedActual, _ := filepath.EvalSymlinks(seedsDir)

	if resolvedActual != resolvedExpected {
		t.Errorf("expected %q, got %q", resolvedExpected, resolvedActual)
	}

	// Test that it creates the directory
	_, err = os.Stat(seedsDir)
	if err != nil {
		t.Errorf("expected directory to be created, got error: %v", err)
	}
}

func TestGenerateSnapshotDir(t *testing.T) {
	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory and use it as working directory
	tmpDir, err := os.MkdirTemp("", "nizam-generate-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test with tag
	snapshotDir, err := GenerateSnapshotDir("test-service", "backup")
	if err != nil {
		t.Fatalf("failed to generate snapshot dir: %v", err)
	}

	// Should exist and be a directory
	info, err := os.Stat(snapshotDir)
	if err != nil {
		t.Errorf("expected directory to exist, got error: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected path to be a directory")
	}

	// Should contain the tag
	if !strings.Contains(filepath.Base(snapshotDir), "backup") {
		t.Errorf("expected directory name to contain 'backup', got %q", filepath.Base(snapshotDir))
	}

	// Test without tag (timestamp only)
	snapshotDir2, err := GenerateSnapshotDir("test-service", "")
	if err != nil {
		t.Fatalf("failed to generate timestamp snapshot dir: %v", err)
	}

	// Should be different directories
	if snapshotDir == snapshotDir2 {
		t.Error("expected different directories for different snapshots")
	}
}

func TestListSnapshots(t *testing.T) {
	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory and use it as working directory
	tmpDir, err := os.MkdirTemp("", "nizam-list-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create some snapshot directories
	_, err = GenerateSnapshotDir("test-service", "backup1")
	if err != nil {
		t.Fatalf("failed to create test snapshot: %v", err)
	}

	_, err = GenerateSnapshotDir("test-service", "backup2")
	if err != nil {
		t.Fatalf("failed to create test snapshot: %v", err)
	}

	// List snapshots
	snapshots, err := ListSnapshots("test-service")
	if err != nil {
		t.Fatalf("failed to list snapshots: %v", err)
	}

	if len(snapshots) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(snapshots))
	}
}
