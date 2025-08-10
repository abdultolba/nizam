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

// MySQLEngine implements snapshot operations for MySQL
type MySQLEngine struct {
	docker *dockerx.Client
}

// NewMySQLEngine creates a new MySQL snapshot engine
func NewMySQLEngine(docker *dockerx.Client) *MySQLEngine {
	return &MySQLEngine{docker: docker}
}

// GetEngineType returns the engine type
func (e *MySQLEngine) GetEngineType() string {
	return "mysql"
}

// CanHandle returns true if this engine can handle the given engine type
func (e *MySQLEngine) CanHandle(engine string) bool {
	return engine == "mysql" || engine == "mariadb"
}

// Create creates a snapshot of a MySQL database
func (e *MySQLEngine) Create(ctx context.Context, service resolve.ServiceInfo, outputDir string, comp compress.Compression, note string, tag string) (*SnapshotManifest, error) {
	log.Info().
		Str("service", service.Name).
		Str("database", service.Database).
		Str("compression", comp.String()).
		Msg("Creating MySQL snapshot")

	// Create manifest
	manifest := NewSnapshotManifest(service.Name, service.Engine, service.Image, tag, note, comp)

	// Determine output filename
	var extension string
	switch comp {
	case compress.CompZstd:
		extension = ".sql.zst"
	case compress.CompGzip:
		extension = ".sql.gz"
	case compress.CompNone:
		extension = ".sql"
	}

	outputFile := filepath.Join(outputDir, "mysql"+extension)
	tempFile := outputFile + ".tmp"

	// Create compressed writer
	writer, err := compress.NewCompressedWriter(tempFile, comp)
	if err != nil {
		return nil, fmt.Errorf("failed to create compressed writer: %w", err)
	}

	// Build mysqldump command
	cmd := []string{
		"mysqldump",
		"--single-transaction",   // Consistent backup with InnoDB
		"--routines",            // Include stored procedures and functions
		"--triggers",            // Include triggers
		"--events",              // Include events
		"--complete-insert",     // Full INSERT statements
		"--extended-insert",     // Multi-row INSERT statements
		"--default-character-set=utf8mb4", // Ensure proper charset
		"-u", service.User,
		"-h", "localhost", // Connect within container
	}

	// Add password if provided
	if service.Password != "" {
		cmd = append(cmd, fmt.Sprintf("-p%s", service.Password))
	}

	// Add database name
	cmd = append(cmd, service.Database)

	log.Debug().
		Str("container", service.Container).
		Strs("command", cmd).
		Msg("Executing mysqldump")

	// Execute mysqldump and stream to compressed writer
	reader, err := e.docker.ExecStreaming(ctx, service.Container, cmd, nil)
	if err != nil {
		writer.Close() // Close writer to cleanup
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to execute mysqldump: %w", err)
	}
	defer reader.Close()

	// Stream data to compressed writer
	written, err := io.Copy(writer, reader)
	if err != nil {
		writer.Close()
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to copy mysqldump output: %w", err)
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
		Msg("MySQL snapshot created successfully")

	return &manifest, nil
}

// Restore restores a MySQL database from a snapshot
func (e *MySQLEngine) Restore(ctx context.Context, service resolve.ServiceInfo, snapshotDir string, manifest *SnapshotManifest, force bool) error {
	log.Info().
		Str("service", service.Name).
		Str("database", service.Database).
		Str("snapshot", snapshotDir).
		Msg("Restoring MySQL snapshot")

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

	// Drop and recreate database if force is enabled
	if force {
		if err := e.recreateDatabase(ctx, service); err != nil {
			return fmt.Errorf("failed to recreate database: %w", err)
		}
	}

	// Open compressed reader
	reader, err := compress.NewCompressedReader(snapshotFile, manifest.GetCompression())
	if err != nil {
		return fmt.Errorf("failed to create compressed reader: %w", err)
	}
	defer reader.Close()

	// Build mysql command
	cmd := []string{
		"mysql",
		"-u", service.User,
		"-h", "localhost", // Connect within container
	}

	// Add password if provided
	if service.Password != "" {
		cmd = append(cmd, fmt.Sprintf("-p%s", service.Password))
	}

	// Add database name
	cmd = append(cmd, service.Database)

	log.Debug().
		Str("container", service.Container).
		Strs("command", cmd).
		Msg("Executing mysql")

	// Execute mysql with streaming input
	execReader, err := e.docker.ExecStreaming(ctx, service.Container, cmd, reader)
	if err != nil {
		return fmt.Errorf("failed to execute mysql: %w", err)
	}
	defer execReader.Close()

	// Read output (for potential error messages)
	output, err := io.ReadAll(execReader)
	if err != nil {
		return fmt.Errorf("failed to read mysql output: %w", err)
	}

	// Check for errors in output
	outputStr := string(output)
	if strings.Contains(outputStr, "ERROR") {
		log.Warn().Str("mysql_output", outputStr).Msg("MySQL restore completed with warnings")
		if !force {
			return fmt.Errorf("mysql restore failed: %s", outputStr)
		}
	}

	log.Info().
		Str("service", service.Name).
		Msg("MySQL snapshot restored successfully")

	return nil
}

// recreateDatabase drops and recreates the database for a clean restore
func (e *MySQLEngine) recreateDatabase(ctx context.Context, service resolve.ServiceInfo) error {
	log.Debug().
		Str("service", service.Name).
		Str("database", service.Database).
		Msg("Recreating database for clean restore")

	// Build command to drop and recreate database
	dropCreateSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`; CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", 
		service.Database, service.Database)

	cmd := []string{
		"mysql",
		"-u", service.User,
		"-h", "localhost",
		"-e", dropCreateSQL,
	}

	// Add password if provided
	if service.Password != "" {
		cmd = append(cmd, fmt.Sprintf("-p%s", service.Password))
	}

	// Execute command
	result, err := e.docker.ExecCommand(ctx, service.Container, cmd)
	if err != nil {
		return fmt.Errorf("failed to recreate database: %w", err)
	}

	if result.ExitCode != 0 || strings.Contains(result.Stdout+result.Stderr, "ERROR") {
		return fmt.Errorf("database recreation failed (exit code %d): %s", result.ExitCode, result.Stdout+result.Stderr)
	}

	return nil
}

// verifyChecksum verifies the SHA256 checksum of a file
func (e *MySQLEngine) verifyChecksum(path, expectedChecksum string) error {
	actualChecksum, err := compress.CalculateSHA256(path)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}
