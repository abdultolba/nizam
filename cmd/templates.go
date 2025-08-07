package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/abdultolba/nizam/internal/templates"
	"github.com/spf13/cobra"
)

var (
	filterTag string
	showTags  bool
)

var templatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "List available service templates",
	Long: `List all available service templates that can be used with 'nizam add'.
Templates provide pre-configured service definitions for popular services
like databases, caches, message queues, and monitoring tools.

Use --tag to filter templates by category (e.g., database, monitoring).
Use --show-tags to see all available tags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showTags {
			return listTags()
		}

		if filterTag != "" {
			return listTemplatesByTag(filterTag)
		}

		return listAllTemplates()
	},
}

func listAllTemplates() error {
	allTemplates := templates.GetBuiltinTemplates()
	
	if len(allTemplates) == 0 {
		fmt.Println("No templates available.")
		return nil
	}

	fmt.Printf("📋 Available Service Templates (%d total)\n\n", len(allTemplates))

	// Sort templates by name for consistent output
	templateNames := templates.GetTemplateNames()
	sort.Strings(templateNames)

	for _, name := range templateNames {
		template := allTemplates[name]
		
		// Format tags
		tagStr := ""
		if len(template.Tags) > 0 {
			tagStr = fmt.Sprintf(" [%s]", strings.Join(template.Tags, ", "))
		}

		fmt.Printf("  %-15s - %s%s\n", name, template.Description, tagStr)
	}

	fmt.Println("\n💡 Use 'nizam add <template>' to add a service from a template")
	fmt.Println("💡 Use 'nizam templates --tag <tag>' to filter by category")
	fmt.Println("💡 Use 'nizam templates --show-tags' to see all available tags")

	return nil
}

func listTemplatesByTag(tag string) error {
	filteredTemplates := templates.GetTemplatesByTag(tag)
	
	if len(filteredTemplates) == 0 {
		fmt.Printf("No templates found with tag '%s'\n", tag)
		fmt.Println("\n💡 Use 'nizam templates --show-tags' to see all available tags")
		return nil
	}

	fmt.Printf("📋 Service Templates with tag '%s' (%d found)\n\n", tag, len(filteredTemplates))

	// Sort by name
	sort.Slice(filteredTemplates, func(i, j int) bool {
		return filteredTemplates[i].Name < filteredTemplates[j].Name
	})

	for _, template := range filteredTemplates {
		// Format other tags (excluding the filter tag)
		var otherTags []string
		for _, t := range template.Tags {
			if t != tag {
				otherTags = append(otherTags, t)
			}
		}
		
		tagStr := ""
		if len(otherTags) > 0 {
			tagStr = fmt.Sprintf(" [%s]", strings.Join(otherTags, ", "))
		}

		fmt.Printf("  %-15s - %s%s\n", template.Name, template.Description, tagStr)
	}

	fmt.Println("\n💡 Use 'nizam add <template>' to add a service from a template")

	return nil
}

func listTags() error {
	allTags := templates.GetAllTags()
	sort.Strings(allTags)

	if len(allTags) == 0 {
		fmt.Println("No tags available.")
		return nil
	}

	fmt.Printf("🏷️  Available Template Tags (%d total)\n\n", len(allTags))

	for _, tag := range allTags {
		templateCount := len(templates.GetTemplatesByTag(tag))
		fmt.Printf("  %-15s (%d template%s)\n", tag, templateCount, func() string {
			if templateCount != 1 {
				return "s"
			}
			return ""
		}())
	}

	fmt.Println("\n💡 Use 'nizam templates --tag <tag>' to filter templates by tag")

	return nil
}

func init() {
	templatesCmd.Flags().StringVarP(&filterTag, "tag", "t", "", "Filter templates by tag (e.g., database, monitoring)")
	templatesCmd.Flags().BoolVar(&showTags, "show-tags", false, "Show all available tags")

	rootCmd.AddCommand(templatesCmd)
}
