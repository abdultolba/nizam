package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/doctor"
	"github.com/abdultolba/nizam/internal/doctor/checks"
	"github.com/spf13/cobra"
)

func NewDoctorCmd() *cobra.Command {
	var jsonOut, fix bool
	var verbose bool
	var skipStr string
	var listChecks bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run preflight checks and suggest fixes",
		Long: `Run preflight checks to ensure your Docker environment is ready for nizam.

This command checks:
- Docker daemon connectivity and version
- Docker Compose plugin availability  
- Available disk space
- Memory usage
- Network MTU configuration
- Port conflicts with configured services

The output shows:
- ✓ for passing checks
- ⚠ for warnings (won't prevent Nizam from running)
- ✗ for failures (must be fixed before running Nizam)

Available check IDs for --skip:
- docker.daemon     : Docker daemon connectivity
- docker.compose    : Docker Compose plugin
- disk.free         : Available disk space
- memory.usage      : System memory usage
- net.mtu           : Network MTU configuration
- port.XXXX         : Port availability (e.g., port.5432, port.9000)

Use --json for machine-readable output, --fix to attempt automatic fixes.
Use --skip to bypass specific checks by ID (e.g., net.mtu,disk.free).`,
		Example: `  # Run all checks
  nizam doctor

  # List available check IDs
  nizam doctor --list-checks

  # Skip specific checks
  nizam doctor --skip net.mtu,disk.free

  # Output as JSON
  nizam doctor --json

  # Attempt automatic fixes
  nizam doctor --fix`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skip := map[string]struct{}{}
			if skipStr != "" {
				for _, id := range strings.Split(skipStr, ",") {
					skip[strings.TrimSpace(id)] = struct{}{}
				}
			}

			// Basic checks
			doctorChecks := []doctor.Check{
				checks.DockerDaemon{},
				checks.ComposePlugin{},
				checks.DiskFree{Path: "/", MinBytes: 5 << 30}, // 5GB
				checks.MemoryUsage{},
				checks.MTUCheck{},
			}

			// Add port checks from config if available
			if cfg, err := config.LoadConfig(); err == nil {
				doctorChecks = append(doctorChecks, getPortChecks(cfg)...)
			}

			// If user wants to list checks, show them and exit
			if listChecks {
				printAvailableChecks(doctorChecks)
				return nil
			}

			r := doctor.Runner{
				Checks:      doctorChecks,
				MaxParallel: 6,
				Timeout:     10 * time.Second,
			}

			ctx := context.Background()
			rep, err := r.Run(ctx, skip, fix)
			if err != nil {
				return err
			}

			if jsonOut {
				rep.PrintJSON()
			} else {
				rep.PrintHuman()
			}

			if rep.Summary.RequiredFailed > 0 {
				return doctor.ErrRequiredFailed
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "verbose output")
	cmd.Flags().BoolVar(&fix, "fix", false, "attempt supported fixes")
	cmd.Flags().BoolVar(&listChecks, "list-checks", false, "list available check IDs and exit")
	cmd.Flags().StringVar(&skipStr, "skip", "", "comma-separated check IDs to skip")

	return cmd
}

func getPortChecks(cfg *config.Config) []doctor.Check {
	var portChecks []doctor.Check

	for serviceName, service := range cfg.Services {
		for _, portMapping := range service.Ports {
			parts := strings.Split(portMapping, ":")
			if len(parts) == 2 {
				if hostPort, err := strconv.Atoi(parts[0]); err == nil {
					portChecks = append(portChecks, checks.PortInUse{
						Host: "localhost",
						Port: hostPort,
						Svc:  serviceName,
					})
				}
			}
		}
	}

	return portChecks
}

func printAvailableChecks(checks []doctor.Check) {
	fmt.Println("Available check IDs for --skip:")
	fmt.Println()
	
	// Group checks by type
	coreChecks := []string{}
	portChecks := []string{}
	
	for _, check := range checks {
		id := check.ID()
		if strings.HasPrefix(id, "port.") {
			portChecks = append(portChecks, id)
		} else {
			coreChecks = append(coreChecks, id)
		}
	}
	
	// Print core checks
	fmt.Println("Core checks:")
	for _, id := range coreChecks {
		var description string
		switch id {
		case "docker.daemon":
			description = "Docker daemon connectivity"
		case "docker.compose":
			description = "Docker Compose plugin"
		case "disk.free":
			description = "Available disk space"
		case "memory.usage":
			description = "System memory usage"
		case "net.mtu":
			description = "Network MTU configuration"
		default:
			description = "Unknown check"
		}
		fmt.Printf("  %-20s : %s\n", id, description)
	}
	
	// Print port checks if any
	if len(portChecks) > 0 {
		fmt.Println("\nPort checks:")
		for _, id := range portChecks {
			port := strings.TrimPrefix(id, "port.")
			fmt.Printf("  %-20s : Port %s availability\n", id, port)
		}
	}
	
	fmt.Println("\nExample usage:")
	fmt.Println("  nizam doctor --skip net.mtu,port.5432")
	if len(portChecks) > 2 {
		fmt.Printf("  nizam doctor --skip %s,%s\n", coreChecks[0], portChecks[0])
	}
}

func init() {
	rootCmd.AddCommand(NewDoctorCmd())
}
