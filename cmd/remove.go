package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/docker"
	"github.com/spf13/cobra"
)

var (
	removeForce bool
	removeAll   bool
)

var removeCmd = &cobra.Command{
	Use:   "remove <service...>",
	Short: "Remove services from your configuration",
	Long: `Remove one or more services from your .nizam.yaml configuration.
This command will stop the service if it's running and remove it from the configuration file.

Examples:
  nizam remove postgres                   # Remove PostgreSQL service
  nizam remove redis mysql               # Remove multiple services
  nizam remove --all                     # Remove all services (keeps empty config)
  nizam remove postgres --force          # Remove without confirmation prompts`,
	Aliases: []string{"rm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.ConfigExists() {
			return fmt.Errorf("no .nizam.yaml configuration found. Run 'nizam init' first")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		var servicesToRemove []string

		if removeAll {
			// Remove all services
			for serviceName := range cfg.Services {
				servicesToRemove = append(servicesToRemove, serviceName)
			}
			if len(servicesToRemove) == 0 {
				fmt.Println("ℹ️  No services found in configuration")
				return nil
			}
		} else {
			// Validate service names provided
			if len(args) == 0 {
				return fmt.Errorf("please specify services to remove or use --all flag")
			}

			for _, serviceName := range args {
				if _, exists := cfg.GetService(serviceName); !exists {
					return fmt.Errorf("service '%s' not found in configuration", serviceName)
				}
				servicesToRemove = append(servicesToRemove, serviceName)
			}
		}

		// Confirm removal unless force flag is used
		if !removeForce {
			fmt.Printf("⚠️  This will remove the following services from your configuration: %s\n", strings.Join(servicesToRemove, ", "))
			fmt.Print("Do you want to continue? (y/N): ")
			
			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			
			if response != "y" && response != "yes" {
				fmt.Println("❌ Operation cancelled")
				return nil
			}
		}

		// Initialize Docker client to stop services
		dockerClient, err := docker.NewClient()
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not connect to Docker: %v\n", err)
			fmt.Println("Services will be removed from configuration but may still be running")
		}

		removedServices := []string{}
		stoppedServices := []string{}

		// Create context for Docker operations
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Stop and remove each service
		for _, serviceName := range servicesToRemove {
			// Try to stop the service if Docker is available
			if dockerClient != nil {
				if err := dockerClient.StopService(ctx, serviceName); err != nil {
					fmt.Printf("⚠️  Could not stop service '%s': %v\n", serviceName, err)
				} else {
					stoppedServices = append(stoppedServices, serviceName)
				}
			}

			// Remove from configuration
			delete(cfg.Services, serviceName)
			removedServices = append(removedServices, serviceName)
		}

		// Save updated configuration
		if err := saveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		// Show results
		fmt.Printf("✅ Removed services from configuration: %s\n", strings.Join(removedServices, ", "))
		if len(stoppedServices) > 0 {
			fmt.Printf("🛑 Stopped running services: %s\n", strings.Join(stoppedServices, ", "))
		}
		fmt.Printf("📝 Configuration saved to %s\n", config.GetConfigPath())

		if len(cfg.Services) == 0 {
			fmt.Println("ℹ️  Configuration now contains no services")
			fmt.Println("🚀 Run 'nizam add <template>' to add new services")
		}

		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Remove services without confirmation prompt")
	removeCmd.Flags().BoolVar(&removeAll, "all", false, "Remove all services from configuration")

	rootCmd.AddCommand(removeCmd)
}
