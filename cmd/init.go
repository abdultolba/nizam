package cmd

import (
	"fmt"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new nizam configuration file",
	Long: `Initialize a new .nizam.yaml configuration file in the current directory.
This will create a default configuration with common services like PostgreSQL, 
Redis, and Meilisearch that you can customize for your project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if config already exists
		if config.ConfigExists() {
			return fmt.Errorf("nizam configuration already exists at %s", config.GetConfigPath())
		}

		// Generate default config
		if err := config.GenerateDefaultConfig(); err != nil {
			return fmt.Errorf("failed to generate configuration: %w", err)
		}

		fmt.Println("✅ Created .nizam.yaml configuration file")
		fmt.Println("📝 Edit the configuration to customize services for your project")
		fmt.Println("🚀 Run 'nizam up' to start your services")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
