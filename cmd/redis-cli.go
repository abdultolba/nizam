package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// redisCliCmd represents the redis-cli command
var redisCliCmd = &cobra.Command{
	Use:   "redis-cli [service] [-- redis-cli-args...]",
	Short: "Connect to Redis service with redis-cli",
	Long: `Connect to a Redis service using redis-cli with auto-resolved connection parameters.

If no service is specified, uses the first Redis service found in config.
All arguments after '--' are passed directly to redis-cli.`,
	Example: `  nizam redis-cli
  nizam redis-cli redis
  nizam redis-cli cache -- --help
  nizam redis-cli redis -- ping`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRedisCliCmd,
}

func init() {
	rootCmd.AddCommand(redisCliCmd)
}

func runRedisCliCmd(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine service name
	serviceName := ""
	if len(args) > 0 {
		serviceName = args[0]
	} else {
		// Find first Redis service
		for name, service := range cfg.GetAllServices() {
			engine := resolve.DetermineEngine(service.Image, name)
			if engine == "redis" {
				serviceName = name
				break
			}
		}
		if serviceName == "" {
			return fmt.Errorf("no Redis service found in config")
		}
	}

	// Resolve service info
	serviceInfo, err := resolve.GetServiceInfo(cfg, serviceName)
	if err != nil {
		return fmt.Errorf("failed to resolve service info: %w", err)
	}

	if serviceInfo.Engine != "redis" {
		return fmt.Errorf("service '%s' is not a Redis service (engine: %s)", serviceName, serviceInfo.Engine)
	}

	// Get passthrough arguments
	var passthroughArgs []string
	if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 {
		passthroughArgs = os.Args[dashIdx+1:]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return connectRedisViaDocker(ctx, serviceInfo, passthroughArgs)
}

func connectRedisViaDocker(ctx context.Context, serviceInfo resolve.ServiceInfo, extraArgs []string) error {
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer docker.Close()

	// Check if container is running
	running, err := docker.ContainerIsRunning(ctx, serviceInfo.Container)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	if !running {
		return fmt.Errorf("container %s is not running", serviceInfo.Container)
	}

	// Build redis-cli command
	cmd := []string{"redis-cli"}

	// Add connection parameters
	if serviceInfo.Host != "" && serviceInfo.Host != "localhost" {
		cmd = append(cmd, "-h", serviceInfo.Host)
	}
	if serviceInfo.Port != 0 && serviceInfo.Port != 6379 {
		cmd = append(cmd, "-p", fmt.Sprintf("%d", serviceInfo.Port))
	}
	if serviceInfo.Password != "" {
		cmd = append(cmd, "-a", serviceInfo.Password)
	}

	if viper.GetBool("verbose") {
		connStr := fmt.Sprintf("redis://%s:%d", serviceInfo.Host, serviceInfo.Port)
		fmt.Fprintf(os.Stderr, "Connecting to: %s\n", dockerx.RedactConnectionString(connStr, false))
	}

	// Add extra arguments
	cmd = append(cmd, extraArgs...)

	// Execute with TTY
	return docker.ExecTTY(ctx, serviceInfo.Container, cmd)
}
