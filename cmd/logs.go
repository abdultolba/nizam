package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/abdultolba/nizam/internal/docker"
	"github.com/spf13/cobra"
)

var (
	follow bool
	tail   string
)

var logsCmd = &cobra.Command{
	Use:   "logs <service>",
	Short: "Show logs for a specific service",
	Long: `Display logs for a specific nizam-managed service.
By default, shows the last 50 lines of logs.
Use --follow to continuously stream new logs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		// Create Docker client
		dockerClient, err := docker.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}
		defer dockerClient.Close()

		ctx := context.Background()

		// Get logs
		fmt.Printf("📝 Showing logs for service '%s'%s\n", serviceName, func() string {
			if follow {
				return " (following)"
			}
			return ""
		}())
		fmt.Println(fmt.Sprintf("   Press Ctrl+C to stop%s", func() string {
			if follow {
				return " following"
			}
			return ""
		}()))
		fmt.Println()

		logsReader, err := dockerClient.GetServiceLogs(ctx, serviceName, follow, tail)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}
		defer logsReader.Close()

		// Handle interruption signals for graceful shutdown when following
		if follow {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			// Stream logs in a goroutine
			go func() {
				io.Copy(os.Stdout, logsReader)
			}()

			// Wait for signal
			<-sigChan
			fmt.Println("\n\n📝 Stopped following logs")
			return nil
		}

		// For non-follow mode, just copy all logs
		_, err = io.Copy(os.Stdout, logsReader)
		if err != nil {
			return fmt.Errorf("failed to read logs: %w", err)
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().StringVarP(&tail, "tail", "t", "50", "Number of lines to show from the end of the logs")

	rootCmd.AddCommand(logsCmd)
}
