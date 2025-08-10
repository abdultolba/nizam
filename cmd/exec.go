package cmd

import (
	"context"
	"fmt"

	"github.com/abdultolba/nizam/internal/docker"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <service> <command> [args...]",
	Short: "Execute a command in a running service container",
	Long: `Execute a command in a running nizam-managed service container.
This allows you to interact directly with services, such as:
- Running psql commands in PostgreSQL
- Accessing Redis CLI
- Running shell commands

Examples:
  nizam exec postgres psql -U user -d myapp
  nizam exec redis redis-cli
  nizam exec postgres bash

Note: Use -- to separate nizam flags from container command flags:
  nizam exec postgres -- psql -U user -d myapp`,
	Args:               cobra.MinimumNArgs(2),
	DisableFlagParsing: true, // Disable flag parsing to pass through all args
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]
		command := args[1:]

		dockerClient, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}
		defer dockerClient.Close()

		ctx := context.Background()

		// Verify service exists and is running
		containers, err := dockerClient.GetServiceStatus(ctx)
		if err != nil {
			return fmt.Errorf("failed to get service status: %w", err)
		}

		var serviceFound bool
		for _, container := range containers {
			if container.Service == serviceName {
				serviceFound = true
				break
			}
		}

		if !serviceFound {
			return fmt.Errorf("service '%s' not found or not running. Use 'nizam status' to check running services", serviceName)
		}

		fmt.Printf("💻 Executing command in '%s': %v\n", serviceName, command)

		if err := dockerClient.ExecInService(ctx, serviceName, command); err != nil {
			return fmt.Errorf("failed to execute command: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
