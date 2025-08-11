package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/abdultolba/nizam/internal/binary"
	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/rs/zerolog/log"
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

	// Try host binary first, fallback to container execution
	if binary.HasBinary(binary.Redis) {
		log.Debug().Msg("Using host redis-cli binary")
		return connectWithHostRedis(serviceInfo, passthroughArgs)
	}

	log.Debug().Msg("redis-cli not found on host, using container execution")
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

// connectWithHostRedis connects using the host's redis-cli binary
func connectWithHostRedis(service resolve.ServiceInfo, extraArgs []string) error {
	args := []string{}
	
	// Add connection parameters
	if service.Host != "" && service.Host != "localhost" {
		args = append(args, "-h", service.Host)
	}
	if service.Port != 0 && service.Port != 6379 {
		args = append(args, "-p", strconv.Itoa(service.Port))
	}
	if service.Password != "" {
		args = append(args, "-a", service.Password)
	}

	// Add extra arguments
	args = append(args, extraArgs...)

	if viper.GetBool("verbose") {
		connStr := fmt.Sprintf("redis://%s:%d", service.Host, service.Port)
		fmt.Fprintf(os.Stderr, "Connecting to: %s\n", dockerx.RedactConnectionString(connStr, false))
	}

	log.Debug().
		Str("command", "redis-cli").
		Strs("args", redactRedisCredentials(args)).
		Msg("Executing host redis-cli client")

	// Execute redis-cli
	cmd := exec.Command("redis-cli", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("redis-cli command failed: %w", err)
	}

	return nil
}

// redactRedisCredentials removes sensitive information from redis-cli arguments for logging
func redactRedisCredentials(args []string) []string {
	redacted := make([]string, len(args))
	copy(redacted, args)

	for i := 0; i < len(redacted); i++ {
		if redacted[i] == "-a" && i+1 < len(redacted) {
			redacted[i+1] = "[REDACTED]"
			i++ // skip the password value
		}
	}

	return redacted
}
