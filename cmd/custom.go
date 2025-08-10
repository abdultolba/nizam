package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/abdultolba/nizam/internal/templates"
	"github.com/spf13/cobra"
)

var customCmd = &cobra.Command{
	Use:   "custom",
	Short: "Manage custom templates",
	Long: `Manage custom service templates including deleting, viewing, and listing custom templates.
Custom templates are stored in ~/.nizam/templates/ and can be shared between projects.`,
}

var customDeleteCmd = &cobra.Command{
	Use:   "delete <template-name>",
	Short: "Delete a custom template",
	Long: `Delete a custom template from your templates directory.
Note: You cannot delete built-in templates.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		// Delete the template
		if err := templates.DeleteCustomTemplate(templateName); err != nil {
			return fmt.Errorf("failed to delete template: %w", err)
		}

		fmt.Printf("✅ Deleted custom template '%s'\n", templateName)
		return nil
	},
}

var customListCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom templates only",
	Long:  `List only your custom templates. Use 'nizam templates' to see all templates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		customTemplates, err := templates.GetCustomTemplates()
		if err != nil {
			return fmt.Errorf("failed to load custom templates: %w", err)
		}

		if len(customTemplates) == 0 {
			fmt.Println("📭 No custom templates found")
			fmt.Printf("💡 Custom templates are stored in: %s\n", templates.GetCustomTemplatesDir())
			fmt.Println("💡 Use 'nizam export <service>' to create a custom template")
			return nil
		}

		fmt.Printf("🔧 Custom Templates (%d found)\n", len(customTemplates))
		fmt.Printf("📂 Location: %s\n\n", templates.GetCustomTemplatesDir())

		for name, template := range customTemplates {
			// Filter out "custom" tag for display
			displayTags := make([]string, 0, len(template.Tags))
			for _, tag := range template.Tags {
				if tag != "custom" {
					displayTags = append(displayTags, tag)
				}
			}

			tagStr := ""
			if len(displayTags) > 0 {
				tagStr = fmt.Sprintf(" [%s]", fmt.Sprintf("%v", displayTags))
			}

			fmt.Printf("  %-15s - %s%s\n", name, template.Description, tagStr)
		}

		fmt.Println("\n💡 Use 'nizam add <template>' to use a custom template")
		fmt.Println("💡 Use 'nizam custom delete <name>' to remove a custom template")

		return nil
	},
}

var customShowCmd = &cobra.Command{
	Use:   "show <template-name>",
	Short: "Show detailed information about a template",
	Long:  `Display detailed configuration information for a template (built-in or custom).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]

		// Get the template
		template, err := templates.GetTemplate(templateName)
		if err != nil {
			return fmt.Errorf("template not found: %w", err)
		}

		// Check if it's a custom template
		isCustom := templateContains(template.Tags, "custom")
		templateType := "Built-in"
		if isCustom {
			templateType = "Custom"
		}

		// Display template information
		fmt.Printf("📋 Template: %s (%s)\n\n", template.Name, templateType)
		fmt.Printf("Description: %s\n", template.Description)

		if len(template.Tags) > 0 {
			fmt.Printf("Tags: %v\n", template.Tags)
		}

		fmt.Println("\n🐳 Service Configuration:")
		fmt.Printf("  Image: %s\n", template.Service.Image)

		if len(template.Service.Ports) > 0 {
			fmt.Printf("  Ports: %v\n", template.Service.Ports)
		}

		if len(template.Service.Environment) > 0 {
			fmt.Printf("  Environment Variables:\n")
			for key, value := range template.Service.Environment {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}

		if template.Service.Volume != "" {
			fmt.Printf("  Volume: %s\n", template.Service.Volume)
		}

		if len(template.Service.Command) > 0 {
			fmt.Printf("  Command: %v\n", template.Service.Command)
		}

		if len(template.Service.Networks) > 0 {
			fmt.Printf("  Networks: %v\n", template.Service.Networks)
		}

		if isCustom {
			templatePath := filepath.Join(templates.GetCustomTemplatesDir(), templateName+".yaml")
			fmt.Printf("\n📂 Template file: %s\n", templatePath)
		}

		fmt.Printf("\n💡 Use 'nizam add %s' to add this service to your project\n", templateName)

		return nil
	},
}

var customDirCmd = &cobra.Command{
	Use:   "dir",
	Short: "Show the custom templates directory path",
	Long:  `Display the path where custom templates are stored and optionally open it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		templatesDir := templates.GetCustomTemplatesDir()
		fmt.Printf("📂 Custom templates directory:\n")
		fmt.Printf("   %s\n", templatesDir)

		// Check if directory exists
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			fmt.Printf("\n⚠️  Directory does not exist yet\n")
			fmt.Printf("💡 It will be created when you export your first template\n")
		} else {
			// Count templates
			files, err := filepath.Glob(filepath.Join(templatesDir, "*.yaml"))
			if err == nil {
				fmt.Printf("\n📊 Contains %d custom template(s)\n", len(files))
			}
		}

		return nil
	},
}

func templateContains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func init() {
	// Add subcommands to custom command
	customCmd.AddCommand(customDeleteCmd)
	customCmd.AddCommand(customListCmd)
	customCmd.AddCommand(customShowCmd)
	customCmd.AddCommand(customDirCmd)

	// Add custom command to root
	rootCmd.AddCommand(customCmd)
}
