package snapshot

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/compress"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/rs/zerolog/log"
)

// MongoDBEngine implements snapshot operations for MongoDB
type MongoDBEngine struct {
	docker *dockerx.Client
}

// NewMongoDBEngine creates a new MongoDB snapshot engine
func NewMongoDBEngine(docker *dockerx.Client) *MongoDBEngine {
	return &MongoDBEngine{docker: docker}
}

// GetEngineType returns the engine type
func (e *MongoDBEngine) GetEngineType() string {
	return "mongo"
}

// CanHandle returns true if this engine can handle the given engine type
func (e *MongoDBEngine) CanHandle(engine string) bool {
	return engine == "mongodb" || engine == "mongo"
}

// Create creates a snapshot of a MongoDB database
func (e *MongoDBEngine) Create(ctx context.Context, service resolve.ServiceInfo, outputDir string, comp compress.Compression, note string, tag string) (*SnapshotManifest, error) {
	log.Info().
		Str("service", service.Name).
		Str("database", service.Database).
		Str("compression", comp.String()).
		Msg("Creating MongoDB snapshot")

	// Create manifest
	manifest := NewSnapshotManifest(service.Name, service.Engine, service.Image, tag, note, comp)

	// Determine output filename
	var extension string
	switch comp {
	case compress.CompZstd:
		extension = ".archive.zst"
	case compress.CompGzip:
		extension = ".archive.gz"
	case compress.CompNone:
		extension = ".archive"
	}

	outputFile := filepath.Join(outputDir, "mongo"+extension)
	tempFile := outputFile + ".tmp"

	// Create compressed writer
	writer, err := compress.NewCompressedWriter(tempFile, comp)
	if err != nil {
		return nil, fmt.Errorf("failed to create compressed writer: %w", err)
	}

	// Build mongodump command
	cmd := []string{
		"mongodump",
		"--host", "localhost:27017", // Connect within container
		"--db", service.Database,
		"--archive",                        // Output to stdout as archive
		"--gzip",                          // Use internal gzip compression
		"--forceTableScan",                // Allow full collection scans
		"--readPreference=secondaryPreferred", // Read from secondary if available
	}

	// Add authentication if provided
	if service.User != "" {
		cmd = append(cmd, "--username", service.User)
	}
	if service.Password != "" {
		cmd = append(cmd, "--password", service.Password)
	}

	log.Debug().
		Str("container", service.Container).
		Strs("command", cmd).
		Msg("Executing mongodump")

	// Execute mongodump and stream to compressed writer
	reader, err := e.docker.ExecStreaming(ctx, service.Container, cmd, nil)
	if err != nil {
		writer.Close() // Close writer to cleanup
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to execute mongodump: %w", err)
	}
	defer reader.Close()

	// Stream data to compressed writer
	written, err := io.Copy(writer, reader)
	if err != nil {
		writer.Close()
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to copy mongodump output: %w", err)
	}

	// Close writer and get checksum
	checksum, err := writer.Close()
	if err != nil {
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to close compressed writer: %w", err)
	}

	// Get file size
	stat, err := os.Stat(tempFile)
	if err != nil {
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to stat output file: %w", err)
	}

	// Atomic move
	if err := os.Rename(tempFile, outputFile); err != nil {
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to move output file: %w", err)
	}

	// Add file to manifest
	manifest.AddFile(filepath.Base(outputFile), checksum, stat.Size())

	log.Info().
		Str("service", service.Name).
		Str("file", outputFile).
		Int64("bytes_written", written).
		Int64("file_size", stat.Size()).
		Str("checksum", checksum[:16]+"...").
		Msg("MongoDB snapshot created successfully")

	return &manifest, nil
}

// Restore restores a MongoDB database from a snapshot
func (e *MongoDBEngine) Restore(ctx context.Context, service resolve.ServiceInfo, snapshotDir string, manifest *SnapshotManifest, force bool) error {
	log.Info().
		Str("service", service.Name).
		Str("database", service.Database).
		Str("snapshot", snapshotDir).
		Msg("Restoring MongoDB snapshot")

	// Get main file from manifest
	mainFile, err := manifest.GetMainFile()
	if err != nil {
		return fmt.Errorf("failed to get main file from manifest: %w", err)
	}

	snapshotFile := filepath.Join(snapshotDir, mainFile.Name)

	// Verify file exists
	if _, err := os.Stat(snapshotFile); err != nil {
		return fmt.Errorf("snapshot file not found: %w", err)
	}

	// Verify checksum
	if err := e.verifyChecksum(snapshotFile, mainFile.Sha256); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Check if container is running
	running, err := e.docker.ContainerIsRunning(ctx, service.Container)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	if !running {
		return fmt.Errorf("container %s is not running", service.Container)
	}

	// Drop database if force is enabled
	if force {
		if err := e.dropDatabase(ctx, service); err != nil {
			return fmt.Errorf("failed to drop database: %w", err)
		}
	}

	// Open compressed reader
	reader, err := compress.NewCompressedReader(snapshotFile, manifest.GetCompression())
	if err != nil {
		return fmt.Errorf("failed to create compressed reader: %w", err)
	}
	defer reader.Close()

	// Build mongorestore command
	cmd := []string{
		"mongorestore",
		"--host", "localhost:27017", // Connect within container
		"--db", service.Database,
		"--archive",    // Read from stdin as archive
		"--gzip",       // Handle gzip decompression
		"--drop",       // Drop collections before restoring
		"--stopOnError", // Stop on first error
	}

	// Add authentication if provided
	if service.User != "" {
		cmd = append(cmd, "--username", service.User)
	}
	if service.Password != "" {
		cmd = append(cmd, "--password", service.Password)
	}

	log.Debug().
		Str("container", service.Container).
		Strs("command", cmd).
		Msg("Executing mongorestore")

	// Execute mongorestore with streaming input
	execReader, err := e.docker.ExecStreaming(ctx, service.Container, cmd, reader)
	if err != nil {
		return fmt.Errorf("failed to execute mongorestore: %w", err)
	}
	defer execReader.Close()

	// Read output (for potential error messages)
	output, err := io.ReadAll(execReader)
	if err != nil {
		return fmt.Errorf("failed to read mongorestore output: %w", err)
	}

	// Check for errors in output
	outputStr := string(output)
	if strings.Contains(outputStr, "error") || strings.Contains(outputStr, "failed") {
		log.Warn().Str("mongorestore_output", outputStr).Msg("MongoDB restore completed with warnings")
		if !force {
			return fmt.Errorf("mongorestore failed: %s", outputStr)
		}
	}

	log.Info().
		Str("service", service.Name).
		Msg("MongoDB snapshot restored successfully")

	return nil
}

// dropDatabase drops the target database for a clean restore
func (e *MongoDBEngine) dropDatabase(ctx context.Context, service resolve.ServiceInfo) error {
	log.Debug().
		Str("service", service.Name).
		Str("database", service.Database).
		Msg("Dropping database for clean restore")

	// Build command to drop database
	dropScript := fmt.Sprintf("db = db.getSiblingDB('%s'); db.dropDatabase();", service.Database)

	cmd := []string{
		"mongosh",
		"--host", "localhost:27017",
		"--eval", dropScript,
	}

	// Add authentication if provided
	if service.User != "" {
		cmd = append(cmd, "--username", service.User)
	}
	if service.Password != "" {
		cmd = append(cmd, "--password", service.Password)
	}

	// Execute command
	result, err := e.docker.ExecCommand(ctx, service.Container, cmd)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("database drop failed (exit code %d): %s", result.ExitCode, result.Stdout+result.Stderr)
	}

	// Check for error patterns in output
	outputStr := result.Stdout + result.Stderr
	if strings.Contains(strings.ToLower(outputStr), "error") {
		log.Warn().Str("drop_output", outputStr).Msg("Database drop completed with warnings")
	}

	return nil
}

// waitForMongoReady waits for MongoDB to be ready after operations
func (e *MongoDBEngine) waitForMongoReady(ctx context.Context, service resolve.ServiceInfo) error {
	cmd := []string{
		"mongosh",
		"--host", "localhost:27017",
		"--eval", "db.adminCommand('ping')",
	}

	// Add authentication if provided
	if service.User != "" {
		cmd = append(cmd, "--username", service.User)
	}
	if service.Password != "" {
		cmd = append(cmd, "--password", service.Password)
	}

	timeout := time.Now().Add(30 * time.Second)

	for time.Now().Before(timeout) {
		result, err := e.docker.ExecCommand(ctx, service.Container, cmd)
		if err == nil && result.ExitCode == 0 && strings.Contains(result.Stdout, "ok") {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("MongoDB not ready after 30 seconds")
}

// verifyChecksum verifies the SHA256 checksum of a file
func (e *MongoDBEngine) verifyChecksum(path, expectedChecksum string) error {
	actualChecksum, err := compress.CalculateSHA256(path)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}
