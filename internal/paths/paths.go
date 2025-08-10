package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetProjectRoot returns the project root directory (where .nizam directory should be)
func GetProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for .nizam.yaml or .nizam directory in current directory and parents
	dir := cwd
	for {
		// Check for config file or .nizam directory
		if hasNizamMarker(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, use current directory
			return cwd, nil
		}
		dir = parent
	}
}

// hasNizamMarker checks if a directory contains nizam markers
func hasNizamMarker(dir string) bool {
	markers := []string{".nizam.yaml", ".nizam.yml", ".nizam"}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

// GetNizamDir returns the .nizam directory path, creating it if needed
func GetNizamDir() (string, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return "", err
	}

	nizamDir := filepath.Join(root, ".nizam")
	if err := os.MkdirAll(nizamDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create .nizam directory: %w", err)
	}

	return nizamDir, nil
}

// GetSnapshotsDir returns the snapshots directory path
func GetSnapshotsDir() (string, error) {
	nizamDir, err := GetNizamDir()
	if err != nil {
		return "", err
	}

	snapshotsDir := filepath.Join(nizamDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	return snapshotsDir, nil
}

// GetServiceSnapshotsDir returns the snapshots directory for a specific service
func GetServiceSnapshotsDir(service string) (string, error) {
	snapshotsDir, err := GetSnapshotsDir()
	if err != nil {
		return "", err
	}

	serviceDir := filepath.Join(snapshotsDir, service)
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create service snapshots directory: %w", err)
	}

	return serviceDir, nil
}

// GenerateSnapshotDir creates a new snapshot directory with timestamp and optional tag
func GenerateSnapshotDir(service, tag string) (string, error) {
	serviceDir, err := GetServiceSnapshotsDir(service)
	if err != nil {
		return "", err
	}

	// Generate directory name: timestamp-tag or just timestamp
	timestamp := time.Now().UTC().Format("20060102-150405")
	var dirName string
	if tag != "" {
		// Sanitize tag for filesystem
		sanitizedTag := strings.ReplaceAll(tag, "/", "-")
		sanitizedTag = strings.ReplaceAll(sanitizedTag, "\\", "-")
		sanitizedTag = strings.ReplaceAll(sanitizedTag, ":", "-")
		dirName = fmt.Sprintf("%s-%s", timestamp, sanitizedTag)
	} else {
		dirName = timestamp
	}

	snapshotDir := filepath.Join(serviceDir, dirName)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	return snapshotDir, nil
}

// GetSeedsDir returns the seeds directory path
func GetSeedsDir() (string, error) {
	nizamDir, err := GetNizamDir()
	if err != nil {
		return "", err
	}

	seedsDir := filepath.Join(nizamDir, "seeds")
	if err := os.MkdirAll(seedsDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create seeds directory: %w", err)
	}

	return seedsDir, nil
}

// GetServiceSeedsDir returns the seeds directory for a specific service
func GetServiceSeedsDir(service string) (string, error) {
	seedsDir, err := GetSeedsDir()
	if err != nil {
		return "", err
	}

	serviceDir := filepath.Join(seedsDir, service)
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create service seeds directory: %w", err)
	}

	return serviceDir, nil
}

// GetSeedPackDir returns the directory for a specific seed pack
func GetSeedPackDir(service, pack string) (string, error) {
	serviceDir, err := GetServiceSeedsDir(service)
	if err != nil {
		return "", err
	}

	packDir := filepath.Join(serviceDir, pack)
	if err := os.MkdirAll(packDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create seed pack directory: %w", err)
	}

	return packDir, nil
}

// GetSeedPackVersionDir returns the directory for a specific version of a seed pack
func GetSeedPackVersionDir(service, pack, version string) (string, error) {
	packDir, err := GetSeedPackDir(service, pack)
	if err != nil {
		return "", err
	}

	versionDir := filepath.Join(packDir, version)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create seed pack version directory: %w", err)
	}

	return versionDir, nil
}

// ListSnapshots lists all snapshot directories for a service
func ListSnapshots(service string) ([]string, error) {
	serviceDir, err := GetServiceSnapshotsDir(service)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	var snapshots []string
	for _, entry := range entries {
		if entry.IsDir() {
			snapshots = append(snapshots, filepath.Join(serviceDir, entry.Name()))
		}
	}

	return snapshots, nil
}

// ListSeedPacks lists all seed packs for a service
func ListSeedPacks(service string) ([]string, error) {
	serviceDir, err := GetServiceSeedsDir(service)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read seeds directory: %w", err)
	}

	var packs []string
	for _, entry := range entries {
		if entry.IsDir() {
			packs = append(packs, entry.Name())
		}
	}

	return packs, nil
}

// ListSeedPackVersions lists all versions for a seed pack
func ListSeedPackVersions(service, pack string) ([]string, error) {
	packDir, err := GetSeedPackDir(service, pack)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(packDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed pack directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	return versions, nil
}
