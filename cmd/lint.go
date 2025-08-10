package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/lint"
	"github.com/spf13/cobra"
)

func NewLintCmd() *cobra.Command {
	var jsonOut bool
	var file string

	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint configuration for best practices",
		Long: `Lint the nizam configuration file to check for best practices and potential issues.

This command applies various rules to check for:
- Image tags (avoid :latest)
- Port mapping format
- Resource limit recommendations

Rules can be selectively applied and the output can be formatted as JSON.`,
		Example: `  # Lint default config file
  nizam lint

  # Lint specific file
  nizam lint --file ./config.yaml

  # JSON output
  nizam lint --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				file = ".nizam.yaml"
			}
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			rep := lint.Combine(
				lint.NoLatest(cfg),
				lint.PortsShape(cfg),
				lint.LimitsRecommended(cfg),
			)

			if jsonOut {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(rep)
			}

			if len(rep.Findings) == 0 {
				fmt.Println("✔ No linting issues found")
				return nil
			}

			for _, f := range rep.Findings {
				lvl := map[string]string{"error": "✖", "warn": "!"}[f.Severity]
				fmt.Printf("%s %s: %s (%s)\n", lvl, f.Path, f.Message, f.Rule)
				fmt.Printf("  Fix: %s\n", f.Suggestion)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	cmd.Flags().StringVar(&file, "file", "", "config file to lint (default: .nizam.yaml)")

	return cmd
}

func init() {
	rootCmd.AddCommand(NewLintCmd())
}
