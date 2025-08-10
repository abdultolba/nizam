package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/spf13/cobra"
)

func NewValidateCmd() *cobra.Command {
	var jsonOut bool
	var strict bool
	var file string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long: `Validate the nizam configuration file for syntax and basic structure.

This command loads and validates the configuration file to ensure it's 
properly formatted and contains required fields.`,
		Example: `  # Validate default config file
  nizam validate

  # Validate specific file
  nizam validate --file ./config.yaml

  # JSON output
  nizam validate --json

  # Strict mode (exit non-zero on failure)
  nizam validate --strict`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				file = ".nizam.yaml"
			}
			cfg, err := config.LoadConfig()
			if err != nil {
				if jsonOut {
					_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": false, "error": err.Error()})
				} else {
					fmt.Printf("Configuration validation failed: %v\n", err)
				}
				if strict {
					return err
				}
				return nil
			}

			// Basic validation
			if len(cfg.Services) == 0 {
				err := fmt.Errorf("no services defined in configuration")
				if jsonOut {
					_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": false, "error": err.Error()})
				} else {
					fmt.Printf("Configuration validation failed: %v\n", err)
				}
				if strict {
					return err
				}
				return nil
			}

			if jsonOut {
				_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
					"ok":       true,
					"services": len(cfg.Services),
					"profile":  cfg.Profile,
				})
			} else {
				fmt.Printf("✔ Configuration is valid\n")
				fmt.Printf("  Profile: %s\n", cfg.Profile)
				fmt.Printf("  Services: %d\n", len(cfg.Services))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	cmd.Flags().BoolVar(&strict, "strict", false, "exit non-zero on validation failure")
	cmd.Flags().StringVar(&file, "file", "", "config file to validate (default: .nizam.yaml)")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewValidateCmd())
}
