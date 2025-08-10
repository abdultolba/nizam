package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/docker"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up [services...]",
	Short: "Start one or more services",
	Long: `Start one or more services defined in your .nizam.yaml configuration.
If no services are specified, all services will be started.

Examples:
  nizam up                    # Start all services
  nizam up postgres           # Start only postgres
  nizam up postgres redis     # Start postgres and redis`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if config exists
		if !config.ConfigExists() {
			return fmt.Errorf("no .nizam.yaml configuration found. Run 'nizam init' first")
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create Docker client
		dockerClient, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}
		defer dockerClient.Close()

		// Determine which services to start
		servicesToStart := make(map[string]config.Service)

		if len(args) == 0 {
			// Start all services
			servicesToStart = cfg.GetAllServices()
			fmt.Printf("🚀 Starting all services (%s)...\n", strings.Join(cfg.GetServiceNames(), ", "))
		} else {
			// Start specific services
			for _, serviceName := range args {
				service, exists := cfg.GetService(serviceName)
				if !exists {
					return fmt.Errorf("service '%s' not found in configuration", serviceName)
				}
				servicesToStart[serviceName] = service
			}
			fmt.Printf("🚀 Starting services: %s...\n", strings.Join(args, ", "))
		}

		// Start services
		ctx := context.Background()
		var errors []string

		for serviceName, serviceConfig := range servicesToStart {
			fmt.Printf("   Starting %s...", serviceName)

			if err := dockerClient.StartService(ctx, serviceName, serviceConfig); err != nil {
				fmt.Printf(" ❌\n")
				errors = append(errors, fmt.Sprintf("%s: %v", serviceName, err))
				continue
			}

			fmt.Printf(" ✅\n")
		}

		if len(errors) > 0 {
			fmt.Println("\n⚠️  Some services failed to start:")
			for _, err := range errors {
				fmt.Printf("   • %s\n", err)
			}
			return fmt.Errorf("failed to start %d service(s)", len(errors))
		}

		fmt.Println("\n🎉 All services started successfully!")
		fmt.Println("💡 Use 'nizam status' to check service health")
		fmt.Println("📝 Use 'nizam logs <service>' to view service logs")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
