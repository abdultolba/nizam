package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/docker"
	"github.com/abdultolba/nizam/internal/healthcheck"
	"github.com/spf13/cobra"
)

var (
	serverAddress   string
	serverInterval  int
	serverAutoStart bool
)

// healthServerCmd represents the health server command
var healthServerCmd = &cobra.Command{
	Use:   "health-server",
	Short: "Start the health check HTTP server",
	Long: `Start the health check HTTP server that provides REST API endpoints and a web dashboard for monitoring service health.

The server provides:
- REST API endpoints for health status (/api/health, /api/services, /api/check/{service})
- Web dashboard at the root URL (/) with real-time monitoring
- Auto-refresh capabilities and manual health check triggers

Examples:
  nizam health-server                           # Start server on :8080 with default settings
  nizam health-server --address :9090          # Start server on port 9090
  nizam health-server --interval 15            # Check health every 15 seconds
  nizam health-server --no-auto-start          # Don't auto-start health checking`,
	RunE: runHealthServer,
}

func init() {
	rootCmd.AddCommand(healthServerCmd)

	// Add flags
	healthServerCmd.Flags().StringVar(&serverAddress, "address", ":8080", "HTTP server address to bind to")
	healthServerCmd.Flags().IntVar(&serverInterval, "interval", 30, "Health check interval in seconds")
	healthServerCmd.Flags().BoolVar(&serverAutoStart, "auto-start", true, "Automatically start health checking")
}

func runHealthServer(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	healthEngine, err := healthcheck.NewEngine(dockerClient, cfg)
	if err != nil {
		return fmt.Errorf("failed to create health check engine: %w", err)
	}

	server := healthcheck.NewServer(healthEngine, serverAddress)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health check engine if auto-start is enabled
	if serverAutoStart {
		interval := time.Duration(serverInterval) * time.Second
		healthEngine.Start(ctx, interval)
		defer healthEngine.Stop()

		fmt.Printf("🔍 Health check engine started (interval: %v)\n", interval)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		fmt.Printf("🌐 Starting health check server on %s\n", serverAddress)
		fmt.Printf("📊 Web dashboard: http://localhost%s\n", getPortFromAddress(serverAddress))
		fmt.Printf("🔗 API endpoints:\n")
		fmt.Printf("   GET  http://localhost%s/api/health           - Health summary\n", getPortFromAddress(serverAddress))
		fmt.Printf("   GET  http://localhost%s/api/services         - All services health\n", getPortFromAddress(serverAddress))
		fmt.Printf("   GET  http://localhost%s/api/health/{service} - Specific service health\n", getPortFromAddress(serverAddress))
		fmt.Printf("   POST http://localhost%s/api/check/{service}  - Trigger health check\n", getPortFromAddress(serverAddress))
		fmt.Printf("\n💡 Press Ctrl+C to stop the server\n\n")

		if err := server.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for termination signal or error
	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		fmt.Printf("\n🛑 Received signal %s, shutting down gracefully...\n", sig)

		// Cancel context to stop all operations
		cancel()

		// Give server time to shut down gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Stop(shutdownCtx); err != nil {
			fmt.Printf("⚠️  Error during server shutdown: %v\n", err)
		}

		fmt.Println("✅ Server stopped successfully")
		return nil
	}
}

// getPortFromAddress extracts port from address string for display purposes
func getPortFromAddress(address string) string {
	if address[0] == ':' {
		return address
	}

	lastColon := -1
	for i := len(address) - 1; i >= 0; i-- {
		if address[i] == ':' {
			lastColon = i
			break
		}
	}

	if lastColon != -1 {
		return address[lastColon:]
	}

	return ":8080" // fallback
}
