package cmd

import (
	"fmt"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/templates"
	"github.com/spf13/cobra"
)

var (
	templateName        string
	templateDescription string
	templateTags        []string
)

var exportCmd = &cobra.Command{
	Use:   "export <service>",
	Short: "Export a service configuration as a custom template",
	Long: `Export an existing service configuration from your .nizam.yaml as a custom template.
This allows you to save and reuse service configurations across different projects.

Custom templates are saved to ~/.nizam/templates/ and can be used with 'nizam add'.

Examples:
  nizam export mysql --name my-mysql --description "Custom MySQL setup"
  nizam export postgres --tags database,custom,company`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serviceName := args[0]

		// Check if config exists
		if !config.ConfigExists() {
			return fmt.Errorf("no .nizam.yaml configuration found. Run 'nizam init' first")
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Get the service configuration
		service, exists := cfg.GetService(serviceName)
		if !exists {
			return fmt.Errorf("service '%s' not found in configuration", serviceName)
		}

		// Determine template name
		name := templateName
		if name == "" {
			name = serviceName
		}

		// Check if template already exists
		existingTemplates := templates.GetAllTemplates()
		if _, exists := existingTemplates[name]; exists {
			return fmt.Errorf("template '%s' already exists. Use a different --name or delete the existing template", name)
		}

		// Create template
		template := templates.Template{
			Name:        name,
			Description: templateDescription,
			Service:     service,
			Tags:        templateTags,
		}

		// Set default description if none provided
		if template.Description == "" {
			template.Description = fmt.Sprintf("Custom template exported from service '%s'", serviceName)
		}

		// Set default tags if none provided
		if len(template.Tags) == 0 {
			template.Tags = []string{"custom", "exported"}
		} else {
			// Ensure "custom" tag is included
			if !contains(template.Tags, "custom") {
				template.Tags = append(template.Tags, "custom")
			}
		}

		// Save the template
		if err := templates.SaveCustomTemplate(template); err != nil {
			return fmt.Errorf("failed to save template: %w", err)
		}

		fmt.Printf("✅ Exported service '%s' as template '%s'\n", serviceName, name)
		fmt.Printf("📝 Template saved to %s/%s.yaml\n", templates.GetCustomTemplatesDir(), name)

		// Show template details
		fmt.Printf("\n📋 Template Details:\n")
		fmt.Printf("   Name: %s\n", template.Name)
		fmt.Printf("   Description: %s\n", template.Description)
		fmt.Printf("   Tags: %v\n", template.Tags)
		fmt.Printf("   Image: %s\n", template.Service.Image)
		if len(template.Service.Ports) > 0 {
			fmt.Printf("   Ports: %v\n", template.Service.Ports)
		}

		fmt.Printf("\n💡 Use 'nizam add %s' to use this template in other projects\n", name)

		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&templateName, "name", "n", "", "Name for the template (default: service name)")
	exportCmd.Flags().StringVarP(&templateDescription, "description", "d", "", "Description for the template")
	exportCmd.Flags().StringSliceVarP(&templateTags, "tags", "t", []string{}, "Tags for the template (comma-separated)")

	rootCmd.AddCommand(exportCmd)
}
