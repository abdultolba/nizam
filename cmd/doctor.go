package cmd

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/doctor"
	"github.com/abdultolba/nizam/internal/doctor/checks"
)

func NewDoctorCmd() *cobra.Command {
	var jsonOut, fix bool
	var verbose bool
	var skipStr string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run preflight checks and suggest fixes",
		Long: `Run preflight checks to ensure your Docker environment is ready for nizam.

This command checks:
- Docker daemon connectivity and version
- Docker Compose plugin availability
- Available disk space
- Network MTU configuration
- Port conflicts with configured services

Use --json for machine-readable output, --fix to attempt automatic fixes.`,
		Example: `  # Run all checks
  nizam doctor

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
				checks.MTUCheck{},
			}

			// Add port checks from config if available
			if cfg, err := config.LoadConfig(); err == nil {
				doctorChecks = append(doctorChecks, getPortChecks(cfg)...)
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

func init() {
	rootCmd.AddCommand(NewDoctorCmd())
}
