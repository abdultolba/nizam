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

// RedisEngine implements snapshot operations for Redis
type RedisEngine struct {
	docker *dockerx.Client
}

// NewRedisEngine creates a new Redis snapshot engine
func NewRedisEngine(docker *dockerx.Client) *RedisEngine {
	return &RedisEngine{docker: docker}
}

// Create creates a snapshot of a Redis database
func (e *RedisEngine) Create(ctx context.Context, service resolve.ServiceInfo, outputDir string, comp compress.Compression, note string, tag string) (*SnapshotManifest, error) {
	log.Info().
		Str("service", service.Name).
		Str("compression", comp.String()).
		Msg("Creating Redis snapshot")

	// Create manifest
	manifest := NewSnapshotManifest(service.Name, service.Engine, service.Image, tag, note, comp)

	// Determine output filename
	var extension string
	switch comp {
	case compress.CompZstd:
		extension = ".rdb.zst"
	case compress.CompGzip:
		extension = ".rdb.gz"
	case compress.CompNone:
		extension = ".rdb"
	}

	outputFile := filepath.Join(outputDir, "redis"+extension)
	tempFile := outputFile + ".tmp"

	// Trigger BGSAVE
	if err := e.triggerBGSave(ctx, service); err != nil {
		return nil, fmt.Errorf("failed to trigger BGSAVE: %w", err)
	}

	// Wait for BGSAVE to complete
	if err := e.waitForBGSave(ctx, service); err != nil {
		return nil, fmt.Errorf("BGSAVE failed: %w", err)
	}

	// Copy dump.rdb from container
	rdbData, err := e.copyRDBFromContainer(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("failed to copy RDB file: %w", err)
	}

	// Create compressed writer
	writer, err := compress.NewCompressedWriter(tempFile, comp)
	if err != nil {
		return nil, fmt.Errorf("failed to create compressed writer: %w", err)
	}

	// Write RDB data to compressed file
	written, err := writer.Write(rdbData)
	if err != nil {
		writer.Close()
		os.Remove(tempFile)
		return nil, fmt.Errorf("failed to write RDB data: %w", err)
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
		Int("bytes_written", written).
		Int64("file_size", stat.Size()).
		Str("checksum", checksum[:16]+"...").
		Msg("Redis snapshot created successfully")

	return &manifest, nil
}

// Restore restores a Redis database from a snapshot
func (e *RedisEngine) Restore(ctx context.Context, service resolve.ServiceInfo, snapshotDir string, manifest *SnapshotManifest, force bool) error {
	log.Info().
		Str("service", service.Name).
		Str("snapshot", snapshotDir).
		Msg("Restoring Redis snapshot")

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

	// Read RDB data
	rdbData, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read RDB data: %w", err)
	}

	// Stop Redis to restore RDB file
	if err := e.flushAndStop(ctx, service); err != nil {
		return fmt.Errorf("failed to prepare Redis for restore: %w", err)
	}

	// Copy RDB data to container
	if err := e.copyRDBToContainer(ctx, service, rdbData); err != nil {
		return fmt.Errorf("failed to copy RDB to container: %w", err)
	}

	// Start Redis
	if err := e.docker.StartContainer(ctx, service.Container); err != nil {
		return fmt.Errorf("failed to start Redis container: %w", err)
	}

	// Wait for Redis to be ready
	if err := e.waitForRedisReady(ctx, service); err != nil {
		return fmt.Errorf("Redis not ready after restore: %w", err)
	}

	log.Info().
		Str("service", service.Name).
		Msg("Redis snapshot restored successfully")

	return nil
}

// triggerBGSave triggers a Redis BGSAVE operation
func (e *RedisEngine) triggerBGSave(ctx context.Context, service resolve.ServiceInfo) error {
	cmd := []string{"redis-cli"}
	if service.Password != "" {
		cmd = append(cmd, "-a", service.Password)
	}
	cmd = append(cmd, "BGSAVE")

	result, err := e.docker.ExecCommand(ctx, service.Container, cmd)
	if err != nil {
		return fmt.Errorf("failed to execute BGSAVE: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("BGSAVE failed with exit code %d: %s", result.ExitCode, result.Stdout)
	}

	if !strings.Contains(result.Stdout, "Background saving started") {
		return fmt.Errorf("unexpected BGSAVE response: %s", result.Stdout)
	}

	return nil
}

// waitForBGSave waits for the BGSAVE operation to complete
func (e *RedisEngine) waitForBGSave(ctx context.Context, service resolve.ServiceInfo) error {
	cmd := []string{"redis-cli"}
	if service.Password != "" {
		cmd = append(cmd, "-a", service.Password)
	}

	timeout := time.Now().Add(5 * time.Minute) // 5 minute timeout

	for time.Now().Before(timeout) {
		// Check LASTSAVE time
		lastSaveCmd := append(cmd, "LASTSAVE")
		initialResult, err := e.docker.ExecCommand(ctx, service.Container, lastSaveCmd)
		if err != nil {
			return fmt.Errorf("failed to get LASTSAVE: %w", err)
		}

		time.Sleep(1 * time.Second)

		// Check LASTSAVE again
		finalResult, err := e.docker.ExecCommand(ctx, service.Container, lastSaveCmd)
		if err != nil {
			return fmt.Errorf("failed to get LASTSAVE: %w", err)
		}

		// If LASTSAVE timestamp changed, BGSAVE completed
		if strings.TrimSpace(finalResult.Stdout) != strings.TrimSpace(initialResult.Stdout) {
			return nil
		}

		// Also check if BGSAVE is still running
		infoCmd := append(cmd, "INFO", "persistence")
		infoResult, err := e.docker.ExecCommand(ctx, service.Container, infoCmd)
		if err == nil && !strings.Contains(infoResult.Stdout, "rdb_bgsave_in_progress:1") {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("BGSAVE timeout after 5 minutes")
}

// copyRDBFromContainer copies the RDB file from the Redis container
func (e *RedisEngine) copyRDBFromContainer(ctx context.Context, service resolve.ServiceInfo) ([]byte, error) {
	// Try to read the RDB file using cat
	cmd := []string{"cat", "/data/dump.rdb"}
	reader, err := e.docker.ExecStreaming(ctx, service.Container, cmd, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read RDB file: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read RDB data: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("RDB file is empty")
	}

	return data, nil
}

// flushAndStop flushes Redis data and stops the container
func (e *RedisEngine) flushAndStop(ctx context.Context, service resolve.ServiceInfo) error {
	// Flush any pending writes
	cmd := []string{"redis-cli"}
	if service.Password != "" {
		cmd = append(cmd, "-a", service.Password)
	}
	cmd = append(cmd, "BGSAVE")

	// Trigger one final save before stopping
	_, err := e.docker.ExecCommand(ctx, service.Container, cmd)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to trigger final BGSAVE before restore")
	}

	// Stop the container
	return e.docker.StopContainer(ctx, service.Container, 10*time.Second)
}

// copyRDBToContainer copies RDB data to the Redis container
func (e *RedisEngine) copyRDBToContainer(ctx context.Context, service resolve.ServiceInfo, data []byte) error {
	// For now, we'll use a simple approach - write to a temp file and copy
	// In a more sophisticated implementation, we could use Docker's copy API

	// Create a temporary file with the RDB data
	tempFile, err := os.CreateTemp("", "redis-restore-*.rdb")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	if _, err := tempFile.Write(data); err != nil {
		return fmt.Errorf("failed to write temp RDB file: %w", err)
	}

	tempFile.Close()

	// Copy the file into the container using docker cp-like functionality
	// For simplicity, we'll use exec to write the data
	// In production, you'd want to use the Docker API's CopyToContainer

	// Write data via stdin to container
	cmd := []string{"tee", "/data/dump.rdb"}
	reader, err := e.docker.ExecStreaming(ctx, service.Container, cmd, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to copy RDB to container: %w", err)
	}
	defer reader.Close()

	// Read any output
	_, err = io.ReadAll(reader)
	return err
}

// waitForRedisReady waits for Redis to be ready after restart
func (e *RedisEngine) waitForRedisReady(ctx context.Context, service resolve.ServiceInfo) error {
	cmd := []string{"redis-cli"}
	if service.Password != "" {
		cmd = append(cmd, "-a", service.Password)
	}
	cmd = append(cmd, "ping")

	timeout := time.Now().Add(30 * time.Second)

	for time.Now().Before(timeout) {
		result, err := e.docker.ExecCommand(ctx, service.Container, cmd)
		if err == nil && result.ExitCode == 0 && strings.Contains(result.Stdout, "PONG") {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("Redis not ready after 30 seconds")
}

// verifyChecksum verifies the SHA256 checksum of a file
func (e *RedisEngine) verifyChecksum(path, expectedChecksum string) error {
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
func (e *RedisEngine) GetEngineType() string {
	return "redis"
}

// CanHandle checks if this engine can handle the given service
func (e *RedisEngine) CanHandle(engine string) bool {
	return engine == "redis"
}
