package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/abdultolba/nizam/internal/binary"
	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/resolve"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// psqlCmd represents the psql command
var psqlCmd = &cobra.Command{
	Use:   "psql [service] [-- psql-args...]",
	Short: "Connect to PostgreSQL service with psql",
	Long: `Connect to a PostgreSQL service using psql with auto-resolved connection parameters.

If no service is specified, uses the first PostgreSQL service found in config.
All arguments after '--' are passed directly to psql.`,
	Example: `  nizam psql
  nizam psql postgres
  nizam psql api-db -- --help
  nizam psql postgres -- -c "SELECT version()"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPSQL,
}

func init() {
	rootCmd.AddCommand(psqlCmd)

	psqlCmd.Flags().String("db", "", "database name (override config)")
	psqlCmd.Flags().String("user", "", "username (override config)")
}

func runPSQL(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine service name
	serviceName := ""
	if len(args) > 0 {
		serviceName = args[0]
	} else {
		// Find first PostgreSQL service
		for name, service := range cfg.GetAllServices() {
			engine := resolve.DetermineEngine(service.Image, name)
			if engine == "postgres" {
				serviceName = name
				break
			}
		}
		if serviceName == "" {
			return fmt.Errorf("no PostgreSQL service found in config")
		}
	}

	// Resolve service info
	serviceInfo, err := resolve.GetServiceInfo(cfg, serviceName)
	if err != nil {
		return fmt.Errorf("failed to resolve service info: %w", err)
	}

	if serviceInfo.Engine != "postgres" {
		return fmt.Errorf("service '%s' is not a PostgreSQL service (engine: %s)", serviceName, serviceInfo.Engine)
	}

	// Override database/user if specified via flags
	if dbFlag := cmd.Flags().Lookup("db"); dbFlag.Changed {
		serviceInfo.Database = dbFlag.Value.String()
	}
	if userFlag := cmd.Flags().Lookup("user"); userFlag.Changed {
		serviceInfo.User = userFlag.Value.String()
	}

	// Get passthrough arguments
	var passthroughArgs []string
	if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 {
		passthroughArgs = os.Args[dashIdx+1:]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try host binary first, fallback to container execution
	if binary.HasBinary(binary.PostgreSQL) {
		log.Debug().Msg("Using host psql binary")
		return connectWithHostPsql(serviceInfo, passthroughArgs)
	}

	log.Debug().Msg("psql not found on host, using container execution")
	return connectViaDocker(ctx, serviceInfo, passthroughArgs)
}

func connectViaDocker(ctx context.Context, serviceInfo resolve.ServiceInfo, extraArgs []string) error {
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer docker.Close()

	// Check if container is running
	running, err := docker.ContainerIsRunning(ctx, serviceInfo.Container)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}
	if !running {
		return fmt.Errorf("container %s is not running", serviceInfo.Container)
	}

	// Build psql command
	cmd := []string{"psql"}

	// Add connection string
	connStr := serviceInfo.GetConnectionString()
	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Connecting to: %s\n", dockerx.RedactConnectionString(connStr, false))
	}
	cmd = append(cmd, connStr)

	// Add extra arguments
	cmd = append(cmd, extraArgs...)

	// Execute with TTY
	return docker.ExecTTY(ctx, serviceInfo.Container, cmd)
}

// connectWithHostPsql connects using the host's psql binary
func connectWithHostPsql(service resolve.ServiceInfo, extraArgs []string) error {
	connStr := service.GetConnectionString()
	args := []string{connStr}
	args = append(args, extraArgs...)

	if viper.GetBool("verbose") {
		fmt.Fprintf(os.Stderr, "Connecting to: %s\n", dockerx.RedactConnectionString(connStr, false))
	}

	log.Debug().
		Str("command", "psql").
		Strs("args", []string{dockerx.RedactConnectionString(connStr, true)}).
		Msg("Executing host psql client")

	// Execute psql with connection string
	cmd := exec.Command("psql", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("psql command failed: %w", err)
	}

	return nil
}
