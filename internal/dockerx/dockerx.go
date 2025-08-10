package dockerx

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
)

// Client wraps the Docker client with nizam-specific functionality
type Client struct {
	cli *client.Client
}

// NewClient creates a new Docker client with API version negotiation
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{cli: cli}, nil
}

// Close closes the Docker client connection
func (c *Client) Close() error {
	return c.cli.Close()
}

// GetClient returns the underlying Docker client
func (c *Client) GetClient() *client.Client {
	return c.cli
}

// ExecResult holds the result of an exec operation
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// ExecCommand executes a command in a container and returns the result
func (c *Client) ExecCommand(ctx context.Context, containerName string, cmd []string) (*ExecResult, error) {
	// Create exec configuration
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create the exec instance
	execResp, err := c.cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to the exec instance
	attachResp, err := c.cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer attachResp.Close()

	// Read output
	output, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read exec output: %w", err)
	}

	// Get exit code
	inspectResp, err := c.cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec instance: %w", err)
	}

	return &ExecResult{
		ExitCode: inspectResp.ExitCode,
		Stdout:   string(output),
		Stderr:   "", // Combined for simplicity
	}, nil
}

// ExecStreaming executes a command in a container with streaming I/O
func (c *Client) ExecStreaming(ctx context.Context, containerName string, cmd []string, stdin io.Reader) (io.ReadCloser, error) {
	// Create exec configuration
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  stdin != nil,
	}

	// Create the exec instance
	execResp, err := c.cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to the exec instance
	attachResp, err := c.cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec instance: %w", err)
	}

	// Stream stdin if provided
	if stdin != nil {
		go func() {
			defer attachResp.CloseWrite()
			_, copyErr := io.Copy(attachResp.Conn, stdin)
			if copyErr != nil {
				log.Debug().Err(copyErr).Msg("Error copying stdin to exec")
			}
		}()
	}

	return io.NopCloser(attachResp.Reader), nil
}

// ExecTTY executes a command in a container with TTY support
func (c *Client) ExecTTY(ctx context.Context, containerName string, cmd []string) error {
	// Create exec configuration
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          true,
	}

	// Create the exec instance
	execResp, err := c.cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to the exec instance
	attachResp, err := c.cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer attachResp.Close()

	// Handle I/O copying
	go func() {
		_, copyErr := io.Copy(attachResp.Conn, os.Stdin)
		if copyErr != nil {
			log.Debug().Err(copyErr).Msg("Error copying stdin")
		}
	}()

	_, err = io.Copy(os.Stdout, attachResp.Reader)
	if err != nil {
		return fmt.Errorf("failed to copy output: %w", err)
	}

	return nil
}

// ContainerExists checks if a container exists
func (c *Client) ContainerExists(ctx context.Context, name string) (bool, error) {
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range containers {
		for _, containerName := range container.Names {
			if strings.TrimPrefix(containerName, "/") == name {
				return true, nil
			}
		}
	}
	return false, nil
}

// ContainerIsRunning checks if a container is currently running
func (c *Client) ContainerIsRunning(ctx context.Context, name string) (bool, error) {
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range containers {
		for _, containerName := range container.Names {
			if strings.TrimPrefix(containerName, "/") == name {
				return container.State == "running", nil
			}
		}
	}
	return false, nil
}

// StopContainer stops a container with a timeout
func (c *Client) StopContainer(ctx context.Context, name string, timeout time.Duration) error {
	timeoutSec := int(timeout.Seconds())
	return c.cli.ContainerStop(ctx, name, container.StopOptions{
		Timeout: &timeoutSec,
	})
}

// StartContainer starts a stopped container
func (c *Client) StartContainer(ctx context.Context, name string) error {
	return c.cli.ContainerStart(ctx, name, types.ContainerStartOptions{})
}

// GetContainerImage returns the image name for a container
func (c *Client) GetContainerImage(ctx context.Context, name string) (string, error) {
	inspect, err := c.cli.ContainerInspect(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}
	return inspect.Config.Image, nil
}

// RedactConnectionString redacts sensitive information from connection strings
func RedactConnectionString(connStr string, debug bool) string {
	if debug {
		return connStr
	}

	// Redact password in connection strings
	if strings.Contains(connStr, "://") {
		parts := strings.Split(connStr, "@")
		if len(parts) == 2 {
			// Handle username:password@host format
			userParts := strings.Split(parts[0], ":")
			if len(userParts) >= 3 { // protocol:username:password
				userParts[2] = "***"
				return strings.Join(userParts, ":") + "@" + parts[1]
			}
		}
	}

	// For other formats, look for password= patterns
	redacted := strings.ReplaceAll(connStr, "password=", "password=***")
	redacted = strings.ReplaceAll(redacted, "pwd=", "pwd=***")

	return redacted
}
