package cmd

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/abdultolba/nizam/internal/config"
)

func NewWaitForCmd() *cobra.Command {
	var timeout string
	var interval string

	cmd := &cobra.Command{
		Use:     "wait-for [service...]",
		Aliases: []string{"wait"},
		Short:   "Wait for services to become ready",
		Long: `Wait for one or more services to become ready before proceeding.

This command checks service readiness by attempting to connect to configured ports
or health check endpoints. It's useful for ensuring dependencies are available
before starting dependent services.`,
		Example: `  # Wait for database service
  nizam wait-for database

  # Wait for multiple services with custom timeout
  nizam wait-for web database --timeout 60s

  # Wait for all services
  nizam wait-for`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			timeoutDuration, err := time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout format: %w", err)
			}

			intervalDuration, err := time.ParseDuration(interval)
			if err != nil {
				return fmt.Errorf("invalid interval format: %w", err)
			}

			var servicesToWait []string
			if len(args) == 0 {
				// Wait for all services
				for name := range cfg.Services {
					servicesToWait = append(servicesToWait, name)
				}
			} else {
				servicesToWait = args
			}

			fmt.Printf("Waiting for %d service(s) to become ready (timeout: %v)...\n", 
				len(servicesToWait), timeoutDuration)

			start := time.Now()
			for {
				allReady := true
				for _, serviceName := range servicesToWait {
					service, exists := cfg.Services[serviceName]
					if !exists {
						return fmt.Errorf("service %s not found in configuration", serviceName)
					}

					ready := checkServiceReadiness(service, serviceName)
					if !ready {
						allReady = false
					}
				}

				if allReady {
					fmt.Printf("✔ All services are ready (took %v)\n", time.Since(start))
					return nil
				}

				if time.Since(start) > timeoutDuration {
					return fmt.Errorf("timeout waiting for services after %v", timeoutDuration)
				}

				time.Sleep(intervalDuration)
			}
		},
	}

	cmd.Flags().StringVar(&timeout, "timeout", "30s", "maximum time to wait for services")
	cmd.Flags().StringVar(&interval, "interval", "1s", "interval between readiness checks")

	return cmd
}

func checkServiceReadiness(service config.Service, serviceName string) bool {
	// Check if service has health check configuration
	if service.HealthCheck != nil && len(service.HealthCheck.Test) > 0 {
		// For simplicity, assume first test command is an HTTP endpoint if it starts with "http"
		test := service.HealthCheck.Test[0]
		if len(test) > 4 && test[:4] == "http" {
			if checkHTTPEndpoint(test) {
				fmt.Printf("✔ %s: health check passed\n", serviceName)
				return true
			}
			fmt.Printf("⏳ %s: waiting for health check...\n", serviceName)
			return false
		}
		// For other health check types, assume ready (could be enhanced)
		fmt.Printf("✔ %s: health check configured\n", serviceName)
		return true
	}

	// Check if service has exposed ports
	if len(service.Ports) > 0 {
		for _, portMapping := range service.Ports {
			// Parse port mapping (e.g., "8080:80")
			hostPort, _, err := parsePortMapping(portMapping)
			if err != nil {
				fmt.Printf("! %s: invalid port mapping %s\n", serviceName, portMapping)
				continue
			}

			if checkTCPPort(fmt.Sprintf("localhost:%s", hostPort)) {
				fmt.Printf("✔ %s: port %s is ready\n", serviceName, hostPort)
				return true
			}
		}
		fmt.Printf("⏳ %s: waiting for ports...\n", serviceName)
		return false
	}

	// If no health check or ports, assume ready
	fmt.Printf("✔ %s: no readiness checks configured\n", serviceName)
	return true
}

func checkHTTPEndpoint(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func checkTCPPort(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func parsePortMapping(portMapping string) (hostPort, containerPort string, err error) {
	// Simple parsing for "host:container" format
	if len(portMapping) == 0 {
		return "", "", fmt.Errorf("empty port mapping")
	}
	
	// For simplicity, assume format is "host:container" or just "port"
	parts := []rune(portMapping)
	colonIdx := -1
	for i, r := range parts {
		if r == ':' {
			colonIdx = i
			break
		}
	}
	
	if colonIdx == -1 {
		// Just a port number
		return portMapping, portMapping, nil
	}
	
	hostPort = string(parts[:colonIdx])
	containerPort = string(parts[colonIdx+1:])
	return hostPort, containerPort, nil
}

func init() {
	rootCmd.AddCommand(NewWaitForCmd())
}
