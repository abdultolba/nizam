package cmd

import (
	"fmt"
	"os"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/templates"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	serviceName string
	overwrite   bool
)

var addCmd = &cobra.Command{
	Use:   "add <template>",
	Short: "Add a service from a template to your configuration",
	Long: `Add a service from a built-in template to your .nizam.yaml configuration.
This command will add the service configuration without starting it.

Use 'nizam templates' to see all available templates.

Examples:
  nizam add postgres          # Add PostgreSQL with default name 'postgres'  
  nizam add redis --name cache # Add Redis with custom name 'cache'
  nizam add mysql --overwrite # Replace existing MySQL service if it exists`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		if !config.ConfigExists() {
			return fmt.Errorf("no .nizam.yaml configuration found. Run 'nizam init' first")
		}

		template, err := templates.GetTemplate(templateName)
		if err != nil {
			return fmt.Errorf("template not found: %w\nUse 'nizam templates' to see available templates", err)
		}

		// Determine service name
		targetServiceName := serviceName
		if targetServiceName == "" {
			targetServiceName = templateName
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if _, exists := cfg.GetService(targetServiceName); exists && !overwrite {
			return fmt.Errorf("service '%s' already exists in configuration. Use --overwrite to replace it", targetServiceName)
		}

		// Add the service to configuration
		if cfg.Services == nil {
			cfg.Services = make(map[string]config.Service)
		}
		cfg.Services[targetServiceName] = template.Service

		// Save the updated configuration
		if err := saveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		// Show success message
		action := "Added"
		if overwrite {
			action = "Updated"
		}

		fmt.Printf("✅ %s service '%s' from template '%s'\n", action, targetServiceName, templateName)
		fmt.Printf("📝 Configuration saved to %s\n", config.GetConfigPath())
		
		// Show template details
		fmt.Printf("\n📋 Service Details:\n")
		fmt.Printf("   Image: %s\n", template.Service.Image)
		if len(template.Service.Ports) > 0 {
			fmt.Printf("   Ports: %v\n", template.Service.Ports)
		}
		if len(template.Service.Environment) > 0 {
			fmt.Printf("   Environment variables: %d configured\n", len(template.Service.Environment))
		}
		if template.Service.Volume != "" {
			fmt.Printf("   Volume: %s\n", template.Service.Volume)
		}

		fmt.Printf("\n🚀 Run 'nizam up %s' to start the service\n", targetServiceName)

		return nil
	},
}

func saveConfig(cfg *config.Config) error {
	configPath := config.GetConfigPath()
	
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}

func init() {
	addCmd.Flags().StringVarP(&serviceName, "name", "n", "", "Custom name for the service (default: template name)")
	addCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing service with the same name")

	rootCmd.AddCommand(addCmd)
}
