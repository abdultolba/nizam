package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
)

// Client wraps the Docker client with nizam-specific functionality
type Client struct {
	cli *client.Client
}

// ContainerInfo holds information about a running container
type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Ports   []string
	Health  string
	Service string
}

// NewClient creates a new Docker client
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

// StartService starts a Docker container for the given service
func (c *Client) StartService(ctx context.Context, serviceName string, serviceConfig config.Service) error {
	containerName := fmt.Sprintf("nizam_%s", serviceName)

	// Check if container already exists
	if exists, err := c.containerExists(ctx, containerName); err != nil {
		return fmt.Errorf("failed to check if container exists: %w", err)
	} else if exists {
		// Try to start existing container
		if err := c.cli.ContainerStart(ctx, containerName, types.ContainerStartOptions{}); err != nil {
			return fmt.Errorf("failed to start existing container: %w", err)
		}
		log.Info().Str("service", serviceName).Msg("Started existing container")
		return nil
	}

	// Pull image if not present
	if err := c.pullImageIfNeeded(ctx, serviceConfig.Image); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Create port bindings
	portBindings, exposedPorts, err := c.createPortBindings(serviceConfig.Ports)
	if err != nil {
		return fmt.Errorf("failed to create port bindings: %w", err)
	}

	// Create environment variables
	env := make([]string, 0, len(serviceConfig.Environment))
	for key, value := range serviceConfig.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Create container configuration
	containerConfig := &container.Config{
		Image:        serviceConfig.Image,
		ExposedPorts: exposedPorts,
		Env:          env,
		Labels: map[string]string{
			"nizam.service": serviceName,
			"nizam.managed": "true",
		},
	}

	// Add health check if configured
	if serviceConfig.HealthCheck != nil && len(serviceConfig.HealthCheck.Test) > 0 {
		healthConfig := &container.HealthConfig{
			Test: serviceConfig.HealthCheck.Test,
		}

		// Parse interval
		if serviceConfig.HealthCheck.Interval != "" {
			if interval, err := time.ParseDuration(serviceConfig.HealthCheck.Interval); err == nil {
				healthConfig.Interval = interval
			} else {
				healthConfig.Interval = 30 * time.Second // default
			}
		} else {
			healthConfig.Interval = 30 * time.Second // default
		}

		// Parse timeout
		if serviceConfig.HealthCheck.Timeout != "" {
			if timeout, err := time.ParseDuration(serviceConfig.HealthCheck.Timeout); err == nil {
				healthConfig.Timeout = timeout
			} else {
				healthConfig.Timeout = 10 * time.Second // default
			}
		} else {
			healthConfig.Timeout = 10 * time.Second // default
		}

		// Set retries
		if serviceConfig.HealthCheck.Retries > 0 {
			healthConfig.Retries = serviceConfig.HealthCheck.Retries
		} else {
			healthConfig.Retries = 3 // default
		}

		containerConfig.Healthcheck = healthConfig
	}

	if len(serviceConfig.Command) > 0 {
		containerConfig.Cmd = serviceConfig.Command
	}

	// Create host configuration
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		AutoRemove:   false,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Handle volumes
	if serviceConfig.Volume != "" {
		volumeName := fmt.Sprintf("nizam_%s_%s", serviceName, serviceConfig.Volume)
		hostConfig.Binds = []string{fmt.Sprintf("%s:/var/lib/postgresql/data", volumeName)}
	}

	// Create the container
	resp, err := c.cli.ContainerCreate(ctx, containerConfig, hostConfig, &network.NetworkingConfig{}, nil, containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	if err := c.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	log.Info().Str("service", serviceName).Str("container_id", resp.ID[:12]).Msg("Started service")
	return nil
}

// StopService stops and removes a service container
func (c *Client) StopService(ctx context.Context, serviceName string) error {
	containerName := fmt.Sprintf("nizam_%s", serviceName)

	// Check if container exists
	exists, err := c.containerExists(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to check if container exists: %w", err)
	}
	if !exists {
		log.Info().Str("service", serviceName).Msg("Container does not exist, nothing to stop")
		return nil
	}

	// Stop the container
	if err := c.cli.ContainerStop(ctx, containerName, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// Remove the container
	if err := c.cli.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{}); err != nil {
		log.Warn().Err(err).Str("service", serviceName).Msg("Failed to remove container")
	}

	log.Info().Str("service", serviceName).Msg("Stopped service")
	return nil
}

// GetServiceStatus returns the status of all nizam-managed containers
func (c *Client) GetServiceStatus(ctx context.Context) ([]ContainerInfo, error) {
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var nizamContainers []ContainerInfo
	for _, container := range containers {
		if container.Labels["nizam.managed"] == "true" {
			serviceName := container.Labels["nizam.service"]
			if serviceName == "" {
				serviceName = "unknown"
			}

			// Extract port information
			var ports []string
			for _, port := range container.Ports {
				if port.PublicPort > 0 {
					ports = append(ports, fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort))
				}
			}

			// Get container name (remove leading slash)
			name := strings.TrimPrefix(container.Names[0], "/")

			nizamContainers = append(nizamContainers, ContainerInfo{
				ID:      container.ID[:12],
				Name:    name,
				Image:   container.Image,
				Status:  container.Status,
				Ports:   ports,
				Service: serviceName,
			})
		}
	}

	return nizamContainers, nil
}

// GetServiceLogs returns logs for a specific service
func (c *Client) GetServiceLogs(ctx context.Context, serviceName string, follow bool, tail string) (io.ReadCloser, error) {
	containerName := fmt.Sprintf("nizam_%s", serviceName)

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
		Tail:       tail,
	}

	logs, err := c.cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for service %s: %w", serviceName, err)
	}

	return logs, nil
}

// ExecInService executes a command in a service container
func (c *Client) ExecInService(ctx context.Context, serviceName string, cmd []string) error {
	containerName := fmt.Sprintf("nizam_%s", serviceName)

	// Create exec configuration
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          true,
		Cmd:          cmd,
	}

	// Create the exec instance
	execResp, err := c.cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to the exec instance
	attachResp, err := c.cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{
		Tty: true,
	})
	if err != nil {
		return fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer attachResp.Close()

	// Copy streams
	go io.Copy(attachResp.Conn, os.Stdin)
	io.Copy(os.Stdout, attachResp.Reader)

	return nil
}

// Helper functions

func (c *Client) containerExists(ctx context.Context, name string) (bool, error) {
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return false, err
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

func (c *Client) pullImageIfNeeded(ctx context.Context, image string) error {
	// Check if image exists locally
	_, _, err := c.cli.ImageInspectWithRaw(ctx, image)
	if err == nil {
		return nil // Image exists
	}

	log.Info().Str("image", image).Msg("Pulling Docker image")

	reader, err := c.cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Copy the pull output to discard (we could show progress here)
	io.Copy(io.Discard, reader)
	return nil
}

func (c *Client) createPortBindings(ports []string) (nat.PortMap, nat.PortSet, error) {
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for _, portMapping := range ports {
		parts := strings.Split(portMapping, ":")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid port mapping: %s", portMapping)
		}

		hostPort := parts[0]
		containerPort := parts[1]

		// Create nat.Port for the container port
		natPort, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid container port %s: %w", containerPort, err)
		}

		// Add to exposed ports
		exposedPorts[natPort] = struct{}{}

		// Add to port bindings
		portBindings[natPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}

	return portBindings, exposedPorts, nil
}
