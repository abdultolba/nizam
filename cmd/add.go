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
	useDefaults bool
)

var addCmd = &cobra.Command{
	Use:   "add <template>",
	Short: "Add a service from a template to your configuration",
	Long: `Add a service from a built-in template to your .nizam.yaml configuration.
This command will add the service configuration without starting it.

For templates with customizable variables (like postgres, mysql, redis, mongodb, rabbitmq),
you'll be prompted to configure ports, credentials, and other settings interactively.
Use --defaults to skip prompts and use default values.

Use 'nizam templates' to see all available templates.

Examples:
  nizam add postgres                    # Add PostgreSQL with interactive configuration
  nizam add postgres --defaults         # Add PostgreSQL with default values  
  nizam add redis --name cache          # Add Redis with custom name 'cache'
  nizam add mysql --overwrite           # Replace existing MySQL service
  nizam add rabbitmq --name broker      # Add RabbitMQ with custom name`,
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

		// Process template with variables if needed
		var processedService config.Service
		if useDefaults || !template.HasVariables() {
			// Use defaults or no variables to process
			processedService, err = templates.ProcessTemplateWithDefaults(template)
		} else {
			// Interactive mode with prompts
			processedService, err = templates.ProcessTemplateWithVariables(template, targetServiceName)
		}
		if err != nil {
			return fmt.Errorf("failed to process template: %w", err)
		}

		// Add the processed service to configuration
		if cfg.Services == nil {
			cfg.Services = make(map[string]config.Service)
		}
		cfg.Services[targetServiceName] = processedService

		// Save the updated configuration
		if err := saveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		// Show success message
		action := "Added"
		if overwrite {
			action = "Updated"
		}

		fmt.Printf("\n✅ %s service '%s' from template '%s'\n", action, targetServiceName, templateName)
		fmt.Printf("📝 Configuration saved to %s\n", config.GetConfigPath())
		
		// Show processed service details
		fmt.Printf("\n📋 Service Details:\n")
		fmt.Printf("   Image: %s\n", processedService.Image)
		if len(processedService.Ports) > 0 {
			fmt.Printf("   Ports: %v\n", processedService.Ports)
		}
		if len(processedService.Environment) > 0 {
			fmt.Printf("   Environment variables: %d configured\n", len(processedService.Environment))
		}
		if processedService.Volume != "" {
			fmt.Printf("   Volume: %s\n", processedService.Volume)
		}
		if len(processedService.Command) > 0 {
			fmt.Printf("   Command: %v\n", processedService.Command)
		}

		// Show template info if variables were processed
		if template.HasVariables() {
			if useDefaults {
				fmt.Printf("\n💡 Used default values for template variables\n")
			} else {
				fmt.Printf("\n💡 Template variables have been configured interactively\n")
			}
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
	addCmd.Flags().BoolVar(&useDefaults, "defaults", false, "Skip interactive prompts and use default values for all template variables")

	rootCmd.AddCommand(addCmd)
}
