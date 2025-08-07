package cmd

import (
	"context"
	"fmt"

	"github.com/abdultolba/nizam/internal/docker"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop all running nizam services",
	Long: `Stop and remove all running nizam-managed containers.
This will gracefully stop all services that were started with 'nizam up'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create Docker client
		dockerClient, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}
		defer dockerClient.Close()

		ctx := context.Background()

		// Get currently running nizam services
		containers, err := dockerClient.GetServiceStatus(ctx)
		if err != nil {
			return fmt.Errorf("failed to get service status: %w", err)
		}

		if len(containers) == 0 {
			fmt.Println("📭 No nizam services are currently running")
			return nil
		}

		fmt.Printf("🛑 Stopping %d service(s)...\n", len(containers))

		var errors []string
		for _, container := range containers {
			fmt.Printf("   Stopping %s...", container.Service)

			if err := dockerClient.StopService(ctx, container.Service); err != nil {
				fmt.Printf(" ❌\n")
				errors = append(errors, fmt.Sprintf("%s: %v", container.Service, err))
				continue
			}

			fmt.Printf(" ✅\n")
		}

		if len(errors) > 0 {
			fmt.Println("\n⚠️  Some services failed to stop:")
			for _, err := range errors {
				fmt.Printf("   • %s\n", err)
			}
			return fmt.Errorf("failed to stop %d service(s)", len(errors))
		}

		fmt.Println("\n🎉 All services stopped successfully!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
