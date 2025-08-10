package snapshot

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/abdultolba/nizam/internal/compress"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/rs/zerolog/log"
)

// PostgreSQLEngine implements snapshot operations for PostgreSQL
type PostgreSQLEngine struct {
	docker *dockerx.Client
}

// NewPostgreSQLEngine creates a new PostgreSQL snapshot engine
func NewPostgreSQLEngine(docker *dockerx.Client) *PostgreSQLEngine {
	return &PostgreSQLEngine{docker: docker}
}

// Create creates a snapshot of a PostgreSQL database
func (e *PostgreSQLEngine) Create(ctx context.Context, service resolve.ServiceInfo, outputDir string, comp compress.Compression, note string, tag string) (*SnapshotManifest, error) {
	log.Info().
		Str("service", service.Name).
		Str("database", service.Database).
		Str("compression", comp.String()).
		Msg("Creating PostgreSQL snapshot")

	// Create manifest
	manifest := NewSnapshotManifest(service.Name, service.Engine, service.Image, tag, note, comp)

	// Determine output filename
	var extension string
	switch comp {
	case compress.CompZstd:
		extension = ".dump.zst"
	case compress.CompGzip:
		extension = ".dump.gz"
	case compress.CompNone:
		extension = ".dump"
	}

	outputFile := filepath.Join(outputDir, "pg"+extension)
	tempFile := outputFile + ".tmp"

	// Create compressed writer
	writer, err := compress.NewCompressedWriter(tempFile, comp)
	if err != nil {
		return nil, fmt.Errorf("failed to create compressed writer: %w", err)
	}

	// Build pg_dump command
	cmd := []string{
		"pg_dump",
		"--format=custom",
		"--no-owner",
		"--no-privileges",
		"-U", service.User,
		"-d", service.Database,
	}

	log.Debug().
		Str("container", service.Container).
		Strs("command", cmd).
		Msg("Executing pg_dump")

	// Execute pg_dump and stream to compressed writer
	reader, err := e.docker.ExecStreaming(ctx, service.Container, cmd, nil)
	if err != nil {
		writer.Close() // Close writer to cleanup
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to execute pg_dump: %w", err)
	}
	defer reader.Close()

	// Stream data to compressed writer
	written, err := io.Copy(writer, reader)
	if err != nil {
		writer.Close()
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to copy pg_dump output: %w", err)
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
		Msg("PostgreSQL snapshot created successfully")

	return &manifest, nil
}

// Restore restores a PostgreSQL database from a snapshot
func (e *PostgreSQLEngine) Restore(ctx context.Context, service resolve.ServiceInfo, snapshotDir string, manifest *SnapshotManifest, force bool) error {
	log.Info().
		Str("service", service.Name).
		Str("database", service.Database).
		Str("snapshot", snapshotDir).
		Msg("Restoring PostgreSQL snapshot")

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

	// Open compressed reader
	reader, err := compress.NewCompressedReader(snapshotFile, manifest.GetCompression())
	if err != nil {
		return fmt.Errorf("failed to create compressed reader: %w", err)
	}
	defer reader.Close()

	// Build pg_restore command
	cmd := []string{
		"pg_restore",
		"--clean",
		"--if-exists",
		"--no-owner",
		"-U", service.User,
		"-d", service.Database,
	}

	if force {
		// Add --single-transaction for atomic restore
		cmd = append(cmd, "--single-transaction")
	}

	log.Debug().
		Str("container", service.Container).
		Strs("command", cmd).
		Msg("Executing pg_restore")

	// Execute pg_restore with streaming input
	execReader, err := e.docker.ExecStreaming(ctx, service.Container, cmd, reader)
	if err != nil {
		return fmt.Errorf("failed to execute pg_restore: %w", err)
	}
	defer execReader.Close()

	// Read output (for potential error messages)
	output, err := io.ReadAll(execReader)
	if err != nil {
		return fmt.Errorf("failed to read pg_restore output: %w", err)
	}

	// Check for errors in output
	outputStr := string(output)
	if strings.Contains(outputStr, "ERROR") || strings.Contains(outputStr, "FATAL") {
		log.Warn().Str("output", outputStr).Msg("pg_restore reported errors")
		if !force {
			return fmt.Errorf("pg_restore failed, output: %s", outputStr)
		}
	}

	log.Info().
		Str("service", service.Name).
		Str("database", service.Database).
		Msg("PostgreSQL snapshot restored successfully")

	return nil
}

// verifyChecksum verifies the SHA256 checksum of a file
func (e *PostgreSQLEngine) verifyChecksum(path, expectedChecksum string) error {
	actualChecksum, err := compress.CalculateSHA256(path)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// GetEngineType returns the engine type
func (e *PostgreSQLEngine) GetEngineType() string {
	return "postgres"
}

// CanHandle checks if this engine can handle the given service
func (e *PostgreSQLEngine) CanHandle(engine string) bool {
	return engine == "postgres" || engine == "postgresql"
}
