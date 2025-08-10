package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/abdultolba/nizam/internal/docker"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of all nizam services",
	Long: `Display the current status of all nizam-managed containers including:
- Service name and container ID
- Current status (running, stopped, etc.)
- Port mappings
- Docker image information`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dockerClient, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}
		defer dockerClient.Close()

		ctx := context.Background()

		containers, err := dockerClient.GetServiceStatus(ctx)
		if err != nil {
			return fmt.Errorf("failed to get service status: %w", err)
		}

		if len(containers) == 0 {
			fmt.Println("📭 No nizam services are currently managed")
			fmt.Println("💡 Run 'nizam up' to start services")
			return nil
		}

		fmt.Printf("📊 Nizam Services Status (%d service(s))\n\n", len(containers))

		// Print table header
		fmt.Printf("%-15s %-12s %-20s %-20s %s\n", "SERVICE", "STATUS", "CONTAINER ID", "PORTS", "IMAGE")
		fmt.Printf("%-15s %-12s %-20s %-20s %s\n",
			strings.Repeat("-", 15),
			strings.Repeat("-", 12),
			strings.Repeat("-", 20),
			strings.Repeat("-", 20),
			strings.Repeat("-", 20))

		// Print service information
		for _, container := range containers {
			status := getStatusEmoji(container.Status) + " " + container.Status
			ports := strings.Join(container.Ports, ", ")
			if ports == "" {
				ports = "none"
			}

			// Truncate image name if too long
			image := container.Image
			if len(image) > 20 {
				image = image[:17] + "..."
			}

			fmt.Printf("%-15s %-12s %-20s %-20s %s\n",
				container.Service,
				status,
				container.ID,
				ports,
				image)
		}

		fmt.Println("\n💡 Use 'nizam logs <service>' to view logs")
		fmt.Println("💡 Use 'nizam exec <service> <command>' to execute commands")

		return nil
	},
}

func getStatusEmoji(status string) string {
	statusLower := strings.ToLower(status)

	if strings.Contains(statusLower, "up") {
		return "🟢"
	} else if strings.Contains(statusLower, "exited") {
		return "🔴"
	} else if strings.Contains(statusLower, "created") {
		return "🟡"
	} else if strings.Contains(statusLower, "restarting") {
		return "🟠"
	}

	return "⚪"
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
