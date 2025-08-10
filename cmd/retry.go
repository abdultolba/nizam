package cmd

import (
	"fmt"
	"time"

	"github.com/abdultolba/nizam/internal/operations"
	"github.com/spf13/cobra"
)

func NewRetryCmd() *cobra.Command {
	var attempts int
	var delay string

	cmd := &cobra.Command{
		Use:   "retry [command]",
		Short: "Retry a failed command with exponential backoff",
		Long: `Retry a failed nizam command with configurable attempts and delay.

This command is useful for handling transient failures in Docker operations,
network issues, or resource conflicts. It uses exponential backoff between attempts.`,
		Example: `  # Retry starting services up to 5 times
  nizam retry start --attempts 5

  # Retry with custom initial delay
  nizam retry start --delay 2s

  # Retry specific services
  nizam retry start web database --attempts 3`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			delayDuration, err := time.ParseDuration(delay)
			if err != nil {
				return fmt.Errorf("invalid delay format: %w", err)
			}

			operation := args[0]
			operationArgs := args[1:]

			for attempt := 1; attempt <= attempts; attempt++ {
				fmt.Printf("Attempt %d/%d: Running 'nizam %s'\n", attempt, attempts, operation)

				err := runOperation(operation, operationArgs)
				if err == nil {
					fmt.Printf("✔ Command succeeded on attempt %d\n", attempt)
					return nil
				}

				fmt.Printf("✖ Attempt %d failed: %v\n", attempt, err)

				if attempt < attempts {
					waitTime := time.Duration(attempt) * delayDuration
					fmt.Printf("Waiting %v before next attempt...\n", waitTime)
					time.Sleep(waitTime)
				}
			}

			return fmt.Errorf("command failed after %d attempts", attempts)
		},
	}

	cmd.Flags().IntVar(&attempts, "attempts", 3, "maximum number of retry attempts")
	cmd.Flags().StringVar(&delay, "delay", "1s", "initial delay between retries (exponential backoff)")

	return cmd
}

func runOperation(operation string, args []string) error {
	switch operation {
	case "start":
		return operations.Start(args)
	case "stop":
		return operations.Stop(args)
	case "restart":
		return operations.Restart(args)
	case "pull":
		return operations.Pull(args)
	case "build":
		return operations.Build(args)
	default:
		return fmt.Errorf("unsupported operation: %s", operation)
	}
}

func init() {
	rootCmd.AddCommand(NewRetryCmd())
}
