package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/abdultolba/nizam/internal/binary"
	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// mysqlCmd represents the mysql command
var mysqlCmd = &cobra.Command{
	Use:     "mysql [service] [-- mysql_args...]",
	Short:   "Connect to MySQL service with auto-resolved credentials",
	Long: `Connect to a MySQL database service with automatically resolved connection parameters.

This command auto-discovers MySQL services from your configuration and extracts
connection details (host, port, username, password, database) to build the
appropriate connection command.

If no service name is provided, it will connect to the first MySQL service found.
If multiple MySQL services exist, specify the service name explicitly.

Examples:
  nizam mysql                          # Connect to first/default MySQL service
  nizam mysql mydb                     # Connect to specific service
  nizam mysql --user root --db mysql   # Override connection parameters
  nizam mysql -- --help               # Pass arguments to mysql client
  nizam mysql -- -e "SHOW DATABASES"  # Execute SQL directly

Connection Resolution:
- Automatically detects MySQL services from configuration
- Extracts credentials from environment variables (MYSQL_USER, MYSQL_PASSWORD, etc.)
- Uses host binaries when available, falls back to container execution
- Supports both mysql and mariadb containers`,
	DisableFlagParsing: true, // Allow -- separator for passing args to mysql
	RunE:               runMysqlCmd,
}

func init() {
	rootCmd.AddCommand(mysqlCmd)
}

func runMysqlCmd(cmd *cobra.Command, args []string) error {
	// Check for help flag explicitly since we use DisableFlagParsing
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}
	}

	// Extract config file path from original arguments before filtering
	configFile := extractConfigFile(args)

	// Filter out global flags since DisableFlagParsing passes everything to us
	filteredArgs := filterGlobalFlags(args)

	// Parse flags manually to handle -- separator
	var serviceName string
	var userOverride, dbOverride string
	var mysqlArgs []string
	var dashDashFound bool

	for i, arg := range filteredArgs {
		if arg == "--" {
			dashDashFound = true
			mysqlArgs = filteredArgs[i+1:]
			break
		}

		if dashDashFound {
			continue
		}

		// Simple flag parsing
		if strings.HasPrefix(arg, "--user=") {
			userOverride = strings.TrimPrefix(arg, "--user=")
		} else if arg == "--user" && i+1 < len(filteredArgs) && !strings.HasPrefix(filteredArgs[i+1], "-") {
			userOverride = filteredArgs[i+1]
			i++ // skip next arg
		} else if strings.HasPrefix(arg, "--db=") || strings.HasPrefix(arg, "--database=") {
			dbOverride = strings.TrimPrefix(arg, "--db=")
			dbOverride = strings.TrimPrefix(dbOverride, "--database=")
		} else if (arg == "--db" || arg == "--database") && i+1 < len(filteredArgs) && !strings.HasPrefix(filteredArgs[i+1], "-") {
			dbOverride = filteredArgs[i+1]
			i++ // skip next arg
		} else if !strings.HasPrefix(arg, "-") && serviceName == "" {
			serviceName = arg
		}
	}

	// Load configuration using extracted config file
	var cfg *config.Config
	var err error
	if configFile != "" {
		cfg, err = config.LoadConfigFromFile(configFile)
	} else {
		cfg, err = config.LoadConfig()
	}
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find MySQL service
	var serviceInfo resolve.ServiceInfo
	if serviceName != "" {
		// Specific service requested
		serviceInfo, err = resolve.GetServiceInfo(cfg, serviceName)
		if err != nil {
			return fmt.Errorf("failed to resolve service info: %w", err)
		}
		
		if serviceInfo.Engine != "mysql" {
			return fmt.Errorf("service '%s' is not a MySQL service (engine: %s)", serviceName, serviceInfo.Engine)
		}
	} else {
		// Auto-discover first MySQL service
		serviceInfo, err = findFirstMySQLService(cfg)
		if err != nil {
			return fmt.Errorf("failed to find MySQL service: %w", err)
		}
	}

	// Apply overrides
	if userOverride != "" {
		serviceInfo.User = userOverride
	}
	if dbOverride != "" {
		serviceInfo.Database = dbOverride
	}

	log.Debug().
		Str("service", serviceInfo.Name).
		Str("host", serviceInfo.Host).
		Int("port", serviceInfo.Port).
		Str("user", serviceInfo.User).
		Str("database", serviceInfo.Database).
		Str("engine", serviceInfo.Engine).
		Msg("Connecting to MySQL service")

	// Try host binary first, fallback to container execution
	if binary.HasBinary(binary.MySQL) {
		log.Debug().Msg("Using host mysql binary")
		return connectWithHostMySQL(serviceInfo, mysqlArgs)
	}

	log.Debug().Msg("mysql not found on host, using container execution")
	return connectWithContainerMySQL(serviceInfo, mysqlArgs)
}

// findFirstMySQLService finds the first MySQL service in the configuration
func findFirstMySQLService(cfg *config.Config) (resolve.ServiceInfo, error) {
	for serviceName, service := range cfg.GetAllServices() {
		engine := resolve.DetermineEngine(service.Image, serviceName)
		if engine == "mysql" {
			return resolve.GetServiceInfo(cfg, serviceName)
		}
	}

	return resolve.ServiceInfo{}, fmt.Errorf("no MySQL services found in configuration")
}


// connectWithHostMySQL connects using the host's mysql binary
func connectWithHostMySQL(service resolve.ServiceInfo, extraArgs []string) error {
	args := []string{
		"-h", service.Host,
		"-P", fmt.Sprintf("%d", service.Port),
		"-u", service.User,
	}

	// Add password if provided
	if service.Password != "" {
		args = append(args, fmt.Sprintf("-p%s", service.Password))
	}

	// Add database if specified
	if service.Database != "" {
		args = append(args, service.Database)
	}

	// Add extra arguments
	args = append(args, extraArgs...)

	log.Debug().
		Str("command", "mysql").
		Strs("args", redactCredentials(args)).
		Msg("Executing host mysql client")

	// Execute mysql command
	cmd := exec.Command("mysql", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("mysql command failed: %w", err)
	}

	return nil
}

// connectWithContainerMySQL connects using mysql inside the container
func connectWithContainerMySQL(service resolve.ServiceInfo, extraArgs []string) error {
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Build mysql command for container execution
	cmd := []string{"mysql"}

	// Add connection parameters
	cmd = append(cmd, "-u", service.User)
	cmd = append(cmd, "-h", "localhost") // Connect within container

	// Add password if provided
	if service.Password != "" {
		cmd = append(cmd, fmt.Sprintf("-p%s", service.Password))
	}

	// Add database if specified
	if service.Database != "" {
		cmd = append(cmd, service.Database)
	}

	// Add extra arguments
	cmd = append(cmd, extraArgs...)

	log.Debug().
		Str("container", service.Container).
		Str("command", "mysql").
		Strs("args", redactCredentials(cmd[1:])).
		Msg("Executing mysql in container")

	// Execute with TTY for interactive session
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return docker.ExecTTY(ctx, service.Container, cmd)
}

// extractConfigFile extracts the config file path from command arguments
func extractConfigFile(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check for --config=value format
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}

		// Check for --config value format
		if arg == "--config" && i+1 < len(args) {
			return args[i+1]
		}
	}

	return ""
}

// filterGlobalFlags removes global flags from args since DisableFlagParsing passes everything
func filterGlobalFlags(args []string) []string {
	filtered := make([]string, 0, len(args))
	globalFlags := map[string]bool{
		"--config":    true,
		"--profile":   true,
		"--verbose":   true,
		"-p":          true, // short form of --profile
		"-v":          true, // short form of --verbose
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check for flag=value format
		if strings.Contains(arg, "=") {
			flagName := strings.Split(arg, "=")[0]
			if globalFlags[flagName] {
				continue // skip this global flag
			}
		}

		// Check for flag value format
		if globalFlags[arg] {
			// Skip flag and its value (if next arg doesn't start with -)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++ // skip the value too
			}
			continue
		}

		// Keep non-global arguments
		filtered = append(filtered, arg)
	}

	return filtered
}

// redactCredentials redacts sensitive information from command arguments for logging
func redactCredentials(args []string) []string {
	redacted := make([]string, len(args))
	copy(redacted, args)

	for i, arg := range redacted {
		// Redact password arguments
		if strings.HasPrefix(arg, "-p") && len(arg) > 2 {
			redacted[i] = "-p***"
		}
	}

	return redacted
}
