package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/templates"
	"github.com/spf13/cobra"
)

var (
	addServices     string
	initUseDefaults bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new nizam configuration file",
	Long: `Initialize a new .nizam.yaml configuration file in the current directory.
By default, this creates a configuration with PostgreSQL, Redis, and Meilisearch.

Use --add to specify custom services instead of the defaults:
  nizam init --add postgres,mysql,redis
  nizam init --add "mongodb, prometheus, mailhog"

Note: The init command always uses default values for template variables to ensure
quick setup. Use 'nizam add' for interactive variable configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if config already exists
		if config.ConfigExists() {
			return fmt.Errorf("nizam configuration already exists at %s", config.GetConfigPath())
		}

		var serviceNames []string

		// Parse custom services if provided
		if addServices != "" {
			// Split by comma and trim spaces
			for _, service := range strings.Split(addServices, ",") {
				service = strings.TrimSpace(service)
				if service != "" {
					serviceNames = append(serviceNames, service)
				}
			}
		} else {
			// Use default services
			serviceNames = []string{"postgres", "redis", "meilisearch"}
		}

		// Validate all template names exist
		for _, serviceName := range serviceNames {
			if _, err := templates.GetTemplate(serviceName); err != nil {
				return fmt.Errorf("template '%s' not found: %w\nUse 'nizam templates' to see available templates", serviceName, err)
			}
		}

		// Generate config with custom or default services
		if err := generateConfigWithServices(serviceNames, initUseDefaults); err != nil {
			return fmt.Errorf("failed to generate configuration: %w", err)
		}

		fmt.Printf("✅ Created .nizam.yaml configuration file with services: %s\n", strings.Join(serviceNames, ", "))
		if !initUseDefaults {
			fmt.Println("💡 Services with variables were configured with default values")
		}
		fmt.Println("📝 Edit the configuration to customize services for your project")
		fmt.Println("🚀 Run 'nizam up' to start your services")

		return nil
	},
}

// generateConfigWithServices creates a config file with specified services
func generateConfigWithServices(serviceNames []string, useDefaults bool) error {
	configPath := ".nizam.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file %s already exists", configPath)
	}

	// Create config structure
	cfg := &config.Config{
		Profile:  "dev",
		Services: make(map[string]config.Service),
	}

	// Process each service template
	for _, serviceName := range serviceNames {
		template, err := templates.GetTemplate(serviceName)
		if err != nil {
			return fmt.Errorf("failed to get template '%s': %w", serviceName, err)
		}

		// Process template with defaults (init always uses defaults)
		processedService, err := templates.ProcessTemplateWithDefaults(template)
		if err != nil {
			return fmt.Errorf("failed to process template '%s': %w", serviceName, err)
		}

		cfg.Services[serviceName] = processedService
	}

	// Save configuration
	return saveConfig(cfg)
}

func init() {
	initCmd.Flags().StringVar(&addServices, "add", "", "Comma-separated list of services to add instead of defaults (e.g., 'postgres,mysql,redis')")

	rootCmd.AddCommand(initCmd)
}
