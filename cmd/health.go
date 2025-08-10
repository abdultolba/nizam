package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/docker"
	"github.com/abdultolba/nizam/internal/healthcheck"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	healthOutputFormat  string
	healthWatchMode     bool
	healthWatchInterval int
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health [service]",
	Short: "Check health status of services",
	Long: `Check the health status of services managed by nizam.

Examples:
  nizam health                    # Check health of all services
  nizam health postgres          # Check health of postgres service
  nizam health --output json     # Output health status in JSON format
  nizam health --watch           # Watch health status continuously
  nizam health --watch --interval 5  # Watch with 5 second intervals`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHealthCheck,
}

func init() {
	rootCmd.AddCommand(healthCmd)

	// Add flags
	healthCmd.Flags().StringVarP(&healthOutputFormat, "output", "o", "table", "Output format (table, json, compact)")
	healthCmd.Flags().BoolVarP(&healthWatchMode, "watch", "w", false, "Watch health status continuously")
	healthCmd.Flags().IntVar(&healthWatchInterval, "interval", 10, "Watch interval in seconds")
}

func runHealthCheck(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	// Create health check engine
	healthEngine, err := healthcheck.NewEngine(dockerClient, cfg)
	if err != nil {
		return fmt.Errorf("failed to create health check engine: %w", err)
	}

	ctx := context.Background()

	if healthWatchMode {
		return watchHealthStatus(ctx, healthEngine, args)
	}

	return checkHealthOnce(ctx, healthEngine, args)
}

func checkHealthOnce(ctx context.Context, engine *healthcheck.Engine, args []string) error {
	if len(args) == 1 {
		// Check specific service
		serviceName := args[0]
		result, err := engine.CheckServiceNow(ctx, serviceName)
		if err != nil {
			return fmt.Errorf("failed to check health of service '%s': %w", serviceName, err)
		}

		if healthOutputFormat == "json" {
			return outputHealthResultJSON(result)
		}

		outputHealthResult(result)
		return nil
	}

	// Check all services
	// Get all services health
	allHealth := make(map[string]*healthcheck.ServiceHealthInfo)
	for serviceName := range engine.GetConfig().Services {
		_, err := engine.CheckServiceNow(ctx, serviceName)
		if err != nil {
			fmt.Printf("Warning: failed to check health of service '%s': %v\n", serviceName, err)
			continue
		}

		healthInfo, exists := engine.GetServiceHealth(serviceName)
		if exists {
			allHealth[serviceName] = healthInfo
		}
	}

	if healthOutputFormat == "json" {
		return outputAllHealthJSON(allHealth)
	}

	outputAllHealth(allHealth)
	return nil
}

func watchHealthStatus(ctx context.Context, engine *healthcheck.Engine, args []string) error {
	interval := time.Duration(healthWatchInterval) * time.Second

	// Start the health check engine
	engine.Start(ctx, interval)
	defer engine.Stop()

	fmt.Printf("🔍 Watching health status (interval: %v, press Ctrl+C to stop)\n\n", interval)

	// Create a ticker for display updates
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial check
	if err := displayCurrentHealth(ctx, engine, args); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			// Clear screen and show updated status
			fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
			fmt.Printf("🔍 Health Status - %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

			if err := displayCurrentHealth(ctx, engine, args); err != nil {
				return err
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func displayCurrentHealth(ctx context.Context, engine *healthcheck.Engine, args []string) error {
	if len(args) == 1 {
		// Display specific service
		serviceName := args[0]
		healthInfo, exists := engine.GetServiceHealth(serviceName)
		if !exists {
			return fmt.Errorf("service '%s' not found", serviceName)
		}

		outputHealthInfo(serviceName, healthInfo)
		return nil
	}

	// Display all services
	allHealth := engine.GetAllServicesHealth()
	outputAllHealth(allHealth)

	// Show summary
	summary := engine.GetHealthSummary()
	fmt.Printf("\nSummary: %d total, %d healthy, %d unhealthy, %d not running, %d unknown\n",
		summary["total_services"].(int),
		summary["healthy"].(int),
		summary["unhealthy"].(int),
		summary["not_running"].(int),
		summary["unknown"].(int))

	return nil
}

func outputHealthResult(result *healthcheck.HealthCheckResult) {
	statusColor := getStatusColor(result.Status)

	fmt.Printf("Service: %s\n", result.ServiceName)
	fmt.Printf("Status:  %s%s%s\n", statusColor, result.Status, "\033[0m")
	fmt.Printf("Message: %s\n", result.Message)
	fmt.Printf("Type:    %s\n", result.CheckType)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Time:    %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))

	if result.Error != nil {
		fmt.Printf("Error:   %v\n", result.Error)
	}

	if result.Details != nil {
		fmt.Printf("Details:\n")
		detailsJSON, _ := json.MarshalIndent(result.Details, "  ", "  ")
		fmt.Printf("  %s\n", string(detailsJSON))
	}
}

func outputHealthResultJSON(result *healthcheck.HealthCheckResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health result: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func outputHealthInfo(serviceName string, healthInfo *healthcheck.ServiceHealthInfo) {
	statusColor := getStatusColor(healthInfo.Status)

	fmt.Printf("Service: %s\n", serviceName)
	fmt.Printf("Status:  %s%s%s\n", statusColor, healthInfo.Status, "\033[0m")
	fmt.Printf("Running: %t\n", healthInfo.IsRunning)

	if healthInfo.ContainerName != "" {
		fmt.Printf("Container: %s\n", healthInfo.ContainerName)
	}

	if healthInfo.Image != "" {
		fmt.Printf("Image: %s\n", healthInfo.Image)
	}

	fmt.Printf("Last Check: %s\n", healthInfo.LastCheck.Format("2006-01-02 15:04:05"))

	if len(healthInfo.CheckHistory) > 0 {
		fmt.Printf("\nRecent Checks:\n")
		for i, check := range healthInfo.CheckHistory {
			if i >= 3 { // Show only last 3 checks
				break
			}
			checkColor := getStatusColor(check.Status)
			fmt.Printf("  %s%s%s - %s (%v)\n",
				checkColor,
				check.Status,
				"\033[0m",
				check.Timestamp.Format("15:04:05"),
				check.Duration)
		}
	}

	fmt.Println()
}

func outputAllHealth(allHealth map[string]*healthcheck.ServiceHealthInfo) {
	if healthOutputFormat == "compact" {
		outputAllHealthCompact(allHealth)
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Service", "Status", "Running", "Container", "Image", "Last Check")

	for serviceName, healthInfo := range allHealth {
		status := string(healthInfo.Status)
		running := "No"
		if healthInfo.IsRunning {
			running = "Yes"
		}

		container := healthInfo.ContainerName
		if container == "" {
			container = "-"
		}

		image := healthInfo.Image
		if image == "" {
			image = "-"
		} else if len(image) > 30 {
			image = image[:27] + "..."
		}

		lastCheck := healthInfo.LastCheck.Format("15:04:05")
		if healthInfo.LastCheck.IsZero() {
			lastCheck = "-"
		}

		// Add color to status
		switch healthInfo.Status {
		case healthcheck.HealthStatusHealthy:
			status = fmt.Sprintf("\033[32m%s\033[0m", status)
		case healthcheck.HealthStatusUnhealthy:
			status = fmt.Sprintf("\033[31m%s\033[0m", status)
		case healthcheck.HealthStatusStarting:
			status = fmt.Sprintf("\033[33m%s\033[0m", status)
		case healthcheck.HealthStatusNotRunning:
			status = fmt.Sprintf("\033[90m%s\033[0m", status)
		default:
			status = fmt.Sprintf("\033[35m%s\033[0m", status)
		}

		table.Append([]string{serviceName, status, running, container, image, lastCheck})
	}

	table.Render()
}

func outputAllHealthCompact(allHealth map[string]*healthcheck.ServiceHealthInfo) {
	for serviceName, healthInfo := range allHealth {
		statusColor := getStatusColor(healthInfo.Status)
		runningStatus := "🛑"
		if healthInfo.IsRunning {
			runningStatus = "🟢"
		}

		fmt.Printf("%s %s%8s%s %s\n",
			runningStatus,
			statusColor,
			healthInfo.Status,
			"\033[0m",
			serviceName)
	}
}

func outputAllHealthJSON(allHealth map[string]*healthcheck.ServiceHealthInfo) error {
	jsonData, err := json.MarshalIndent(allHealth, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health data: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func getStatusColor(status healthcheck.HealthStatus) string {
	switch status {
	case healthcheck.HealthStatusHealthy:
		return "\033[32m" // Green
	case healthcheck.HealthStatusUnhealthy:
		return "\033[31m" // Red
	case healthcheck.HealthStatusStarting:
		return "\033[33m" // Yellow
	case healthcheck.HealthStatusNotRunning:
		return "\033[90m" // Gray
	default:
		return "\033[35m" // Magenta
	}
}
