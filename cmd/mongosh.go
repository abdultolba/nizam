package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// mongoshCmd represents the mongosh command
var mongoshCmd = &cobra.Command{
	Use:     "mongosh [service] [-- mongosh_args...]",
	Short:   "Connect to MongoDB service with auto-resolved credentials",
	Long: `Connect to a MongoDB database service with automatically resolved connection parameters.

This command auto-discovers MongoDB services from your configuration and extracts
connection details (host, port, username, password, database) to build the
appropriate connection command.

If no service name is provided, it will connect to the first MongoDB service found.
If multiple MongoDB services exist, specify the service name explicitly.

Examples:
  nizam mongosh                        # Connect to first/default MongoDB service  
  nizam mongosh mydb                   # Connect to specific service
  nizam mongosh --user admin --db app  # Override connection parameters
  nizam mongosh -- --help             # Pass arguments to mongosh client
  nizam mongosh -- --eval "db.version()" # Execute JavaScript directly

Connection Resolution:
- Automatically detects MongoDB services from configuration
- Extracts credentials from environment variables (MONGO_INITDB_ROOT_USERNAME, MONGO_INITDB_ROOT_PASSWORD, etc.)
- Uses host binaries when available, falls back to container execution
- Supports both mongodb and mongo containers`,
	DisableFlagParsing: true, // Allow -- separator for passing args to mongosh
	RunE:               runMongoshCmd,
}

func init() {
	rootCmd.AddCommand(mongoshCmd)
}

func runMongoshCmd(cmd *cobra.Command, args []string) error {
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
	var mongoshArgs []string
	var dashDashFound bool

	for i, arg := range filteredArgs {
		if arg == "--" {
			dashDashFound = true
			mongoshArgs = filteredArgs[i+1:]
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

	// Find MongoDB service
	var serviceInfo resolve.ServiceInfo
	if serviceName != "" {
		// Specific service requested
		serviceInfo, err = resolve.GetServiceInfo(cfg, serviceName)
		if err != nil {
			return fmt.Errorf("failed to resolve service info: %w", err)
		}
		
		if serviceInfo.Engine != "mongo" {
			return fmt.Errorf("service '%s' is not a MongoDB service (engine: %s)", serviceName, serviceInfo.Engine)
		}
	} else {
		// Auto-discover first MongoDB service
		serviceInfo, err = findFirstMongoDBService(cfg)
		if err != nil {
			return fmt.Errorf("failed to find MongoDB service: %w", err)
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
		Msg("Connecting to MongoDB service")

	// Try host binary first, fallback to container execution
	if hasHostMongoshClient() {
		return connectWithHostMongosh(serviceInfo, mongoshArgs)
	}

	log.Debug().Msg("mongosh client not found on host, using container execution")
	return connectWithContainerMongosh(serviceInfo, mongoshArgs)
}

// findFirstMongoDBService finds the first MongoDB service in the configuration
func findFirstMongoDBService(cfg *config.Config) (resolve.ServiceInfo, error) {
	for serviceName, service := range cfg.GetAllServices() {
		engine := resolve.DetermineEngine(service.Image, serviceName)
		if engine == "mongo" {
			return resolve.GetServiceInfo(cfg, serviceName)
		}
	}

	return resolve.ServiceInfo{}, fmt.Errorf("no MongoDB services found in configuration")
}

// hasHostMongoshClient checks if mongosh client is available on the host
func hasHostMongoshClient() bool {
	_, err := exec.LookPath("mongosh")
	return err == nil
}

// connectWithHostMongosh connects using the host's mongosh binary
func connectWithHostMongosh(service resolve.ServiceInfo, extraArgs []string) error {
	var connectionStr string
	
	// Build MongoDB connection string
	if service.User != "" && service.Password != "" {
		// With authentication
		connectionStr = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", 
			service.User, service.Password, service.Host, service.Port, service.Database)
	} else {
		// Without authentication
		connectionStr = fmt.Sprintf("mongodb://%s:%d/%s", 
			service.Host, service.Port, service.Database)
	}

	args := []string{connectionStr}
	
	// Add extra arguments
	args = append(args, extraArgs...)

	log.Debug().
		Str("command", "mongosh").
		Strs("args", redactMongoCredentials(args)).
		Msg("Executing host mongosh client")

	// Execute mongosh command
	cmd := exec.Command("mongosh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// connectWithContainerMongosh connects using mongosh inside the container
func connectWithContainerMongosh(service resolve.ServiceInfo, extraArgs []string) error {
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Build mongosh command for container execution
	cmd := []string{"mongosh"}

	// Add connection parameters
	cmd = append(cmd, "--host", "localhost:27017") // Connect within container

	// Add authentication if provided
	if service.User != "" {
		cmd = append(cmd, "--username", service.User)
	}
	if service.Password != "" {
		cmd = append(cmd, "--password", service.Password)
	}

	// Add database if specified
	if service.Database != "" {
		cmd = append(cmd, service.Database)
	}

	// Add extra arguments
	cmd = append(cmd, extraArgs...)

	log.Debug().
		Str("container", service.Container).
		Str("command", "mongosh").
		Strs("args", redactMongoCredentials(cmd[1:])).
		Msg("Executing mongosh in container")

	// Execute with TTY for interactive session
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return docker.ExecTTY(ctx, service.Container, cmd)
}

// redactMongoCredentials redacts sensitive information from command arguments for logging
func redactMongoCredentials(args []string) []string {
	redacted := make([]string, len(args))
	copy(redacted, args)

	for i, arg := range redacted {
		// Redact password arguments
		if arg == "--password" && i+1 < len(redacted) {
			redacted[i+1] = "***"
		}
		
		// Redact connection strings containing credentials
		if strings.HasPrefix(arg, "mongodb://") && strings.Contains(arg, "@") {
			parts := strings.Split(arg, "@")
			if len(parts) == 2 {
				credentialsPart := strings.Split(parts[0], "://")[1]
				if strings.Contains(credentialsPart, ":") {
					userPart := strings.Split(credentialsPart, ":")[0]
					redacted[i] = fmt.Sprintf("mongodb://%s:***@%s", userPart, parts[1])
				}
			}
		}
	}

	return redacted
}
