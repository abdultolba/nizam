package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/seedpack"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// packCmd represents the pack command
var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Manage seed data packs",
	Long: `Create, install, list, and manage seed data packs.

Seed packs are reusable database snapshots with enhanced metadata,
examples, and documentation. They make it easy to share and distribute
database seeds across projects and teams.`,
}

// packCreateCmd creates a new seed pack
var packCreateCmd = &cobra.Command{
	Use:   "create <service> [snapshot-tag]",
	Short: "Create a seed pack from a snapshot",
	Long: `Create a seed pack from an existing snapshot.

If no snapshot tag is specified, the latest snapshot will be used.
The pack will be stored in .nizam/seeds/<engine>/<pack-name>/<version>/`,
	Example: `  nizam pack create postgres
  nizam pack create postgres my-snapshot --name "ecommerce-data"
  nizam pack create redis --name "session-cache" --author "John Doe"
  nizam pack create postgres --name "blog-content" --description "Sample blog with posts and users"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runPackCreate,
}

// packListCmd lists available seed packs
var packListCmd = &cobra.Command{
	Use:   "list [engine]",
	Short: "List available seed packs",
	Long: `List all available seed packs or packs for a specific engine.

Without an engine argument, lists all packs across all engines.`,
	Example: `  nizam pack list
  nizam pack list postgres
  nizam pack list --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPackList,
}

// packSearchCmd searches for seed packs
var packSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for seed packs",
	Long: `Search for seed packs by name, description, tags, or author.

Use flags to filter by specific criteria.`,
	Example: `  nizam pack search ecommerce
  nizam pack search --tag sql --tag relational
  nizam pack search --author "John Doe"
  nizam pack search --engine postgres`,
	RunE: runPackSearch,
}

// packInstallCmd installs a seed pack
var packInstallCmd = &cobra.Command{
	Use:   "install <service> <pack>",
	Short: "Install a seed pack to a service",
	Long: `Install a seed pack to a running service.

The pack name can include a version (pack@version) or use the latest version.
This will restore the pack's data to the specified service.`,
	Example: `  nizam pack install postgres ecommerce-data
  nizam pack install postgres ecommerce-data@1.0.0
  nizam pack install redis session-cache --force`,
	Args: cobra.ExactArgs(2),
	RunE: runPackInstall,
}

// packInfoCmd shows detailed information about a pack
var packInfoCmd = &cobra.Command{
	Use:   "info <engine> <pack>",
	Short: "Show detailed information about a seed pack",
	Long: `Show detailed information about a seed pack including metadata,
examples, dependencies, and usage information.`,
	Example: `  nizam pack info postgres ecommerce-data
  nizam pack info postgres ecommerce-data@1.0.0
  nizam pack info redis session-cache`,
	Args: cobra.ExactArgs(2),
	RunE: runPackInfo,
}

// packRemoveCmd removes a seed pack
var packRemoveCmd = &cobra.Command{
	Use:   "remove <engine> <pack>",
	Short: "Remove a seed pack",
	Long: `Remove a seed pack from local storage.

If no version is specified, all versions of the pack will be removed.`,
	Example: `  nizam pack remove postgres ecommerce-data
  nizam pack remove postgres ecommerce-data@1.0.0`,
	Args: cobra.ExactArgs(2),
	RunE: runPackRemove,
}

func init() {
	rootCmd.AddCommand(packCmd)

	// Add subcommands
	packCmd.AddCommand(packCreateCmd)
	packCmd.AddCommand(packListCmd)
	packCmd.AddCommand(packSearchCmd)
	packCmd.AddCommand(packInstallCmd)
	packCmd.AddCommand(packInfoCmd)
	packCmd.AddCommand(packRemoveCmd)

	// Create command flags
	packCreateCmd.Flags().String("name", "", "name for the seed pack")
	packCreateCmd.Flags().String("display-name", "", "display name for the seed pack")
	packCreateCmd.Flags().String("description", "", "description of the seed pack")
	packCreateCmd.Flags().String("author", "", "author of the seed pack")
	packCreateCmd.Flags().String("version", "1.0.0", "version of the seed pack")
	packCreateCmd.Flags().String("license", "MIT", "license for the seed pack")
	packCreateCmd.Flags().String("homepage", "", "homepage URL for the seed pack")
	packCreateCmd.Flags().String("repository", "", "repository URL for the seed pack")
	packCreateCmd.Flags().StringSlice("tag", []string{}, "tags for the seed pack (can be used multiple times)")
	packCreateCmd.Flags().StringSlice("use-case", []string{}, "use cases for the seed pack (can be used multiple times)")
	packCreateCmd.Flags().Bool("force", false, "overwrite existing pack")

	// List command flags
	packListCmd.Flags().Bool("json", false, "output in JSON format")

	// Search command flags
	packSearchCmd.Flags().String("engine", "", "filter by engine type")
	packSearchCmd.Flags().StringSlice("tag", []string{}, "filter by tags (can be used multiple times)")
	packSearchCmd.Flags().String("author", "", "filter by author")
	packSearchCmd.Flags().Bool("json", false, "output in JSON format")

	// Install command flags
	packInstallCmd.Flags().Bool("force", false, "force installation even if errors occur")
	packInstallCmd.Flags().Bool("dry-run", false, "show what would be installed without installing")

	// Remove command flags
	packRemoveCmd.Flags().String("version", "", "specific version to remove")
}

func runPackCreate(cmd *cobra.Command, args []string) error {
	serviceName := args[0]
	var snapshotTag string
	if len(args) > 1 {
		snapshotTag = args[1]
	}

	// Parse flags
	name, _ := cmd.Flags().GetString("name")
	displayName, _ := cmd.Flags().GetString("display-name")
	description, _ := cmd.Flags().GetString("description")
	author, _ := cmd.Flags().GetString("author")
	version, _ := cmd.Flags().GetString("version")
	license, _ := cmd.Flags().GetString("license")
	homepage, _ := cmd.Flags().GetString("homepage")
	repository, _ := cmd.Flags().GetString("repository")
	tags, _ := cmd.Flags().GetStringSlice("tag")
	useCases, _ := cmd.Flags().GetStringSlice("use-case")
	force, _ := cmd.Flags().GetBool("force")

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Docker client
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close()

	// Create seed pack service
	packSvc := seedpack.NewService(docker)

	// Create seed pack
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	opts := seedpack.CreateOptions{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Author:      author,
		Version:     version,
		License:     license,
		Homepage:    homepage,
		Repository:  repository,
		Tags:        tags,
		UseCases:    useCases,
		Force:       force,
	}

	manifest, err := packSvc.Create(ctx, cfg, serviceName, snapshotTag, opts)
	if err != nil {
		return fmt.Errorf("failed to create seed pack: %w", err)
	}

	fmt.Printf("Seed pack created successfully:\n")
	fmt.Printf("  Name: %s\n", manifest.GetFullName())
	fmt.Printf("  Display Name: %s\n", manifest.GetDisplayTitle())
	fmt.Printf("  Engine: %s\n", manifest.Engine)
	fmt.Printf("  Author: %s\n", manifest.Author)
	fmt.Printf("  Data Size: %s\n", manifest.FormatSize())
	if len(manifest.Tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(manifest.Tags, ", "))
	}

	return nil
}

func runPackList(cmd *cobra.Command, args []string) error {
	// Parse flags
	jsonOutput, _ := cmd.Flags().GetBool("json")

	var engine string
	if len(args) > 0 {
		engine = args[0]
	}

	// Create Docker client
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close()

	// Create seed pack service
	packSvc := seedpack.NewService(docker)

	// List packs
	packs, err := packSvc.List(engine)
	if err != nil {
		return fmt.Errorf("failed to list seed packs: %w", err)
	}

	if len(packs) == 0 {
		if engine != "" {
			fmt.Printf("No seed packs found for engine '%s'\n", engine)
		} else {
			fmt.Println("No seed packs found")
		}
		return nil
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(packs)
	}

	// Table output
	headers := []string{"Name", "Version", "Engine", "Author", "Size", "Age", "Tags"}
	rows := [][]string{}

	for _, pack := range packs {
		tags := strings.Join(pack.Tags, ",")
		if len(tags) > 20 {
			tags = tags[:17] + "..."
		}
		if tags == "" {
			tags = "-"
		}

		rows = append(rows, []string{
			pack.Name,
			pack.Version,
			pack.Engine,
			pack.Author,
			pack.FormatSize(),
			pack.GetAge(),
			tags,
		})
	}

	// Create and render table
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader(headers),
	)

	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
	fmt.Printf("\nTotal: %d seed packs\n", len(packs))

	return nil
}

func runPackSearch(cmd *cobra.Command, args []string) error {
	// Parse flags
	engine, _ := cmd.Flags().GetString("engine")
	tags, _ := cmd.Flags().GetStringSlice("tag")
	author, _ := cmd.Flags().GetString("author")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	var query string
	if len(args) > 0 {
		query = args[0]
	}

	// Create Docker client
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close()

	// Create seed pack service
	packSvc := seedpack.NewService(docker)

	// Search packs
	opts := seedpack.SearchOptions{
		Engine: engine,
		Tags:   tags,
		Author: author,
		Query:  query,
	}

	packs, err := packSvc.Search(opts)
	if err != nil {
		return fmt.Errorf("failed to search seed packs: %w", err)
	}

	if len(packs) == 0 {
		fmt.Println("No seed packs found matching criteria")
		return nil
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(packs)
	}

	// Table output
	headers := []string{"Name", "Version", "Engine", "Author", "Size", "Description"}
	rows := [][]string{}

	for _, pack := range packs {
		description := pack.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		rows = append(rows, []string{
			pack.Name,
			pack.Version,
			pack.Engine,
			pack.Author,
			pack.FormatSize(),
			description,
		})
	}

	// Create and render table
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader(headers),
	)

	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
	fmt.Printf("\nFound: %d seed packs\n", len(packs))

	return nil
}

func runPackInstall(cmd *cobra.Command, args []string) error {
	serviceName := args[0]
	packName := args[1]

	// Parse flags
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Docker client
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close()

	// Create seed pack service
	packSvc := seedpack.NewService(docker)

	// Install pack
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	opts := seedpack.InstallOptions{
		Force:  force,
		DryRun: dryRun,
	}

	if err := packSvc.Install(ctx, cfg, serviceName, packName, opts); err != nil {
		return fmt.Errorf("failed to install seed pack: %w", err)
	}

	if !dryRun {
		fmt.Printf("Seed pack '%s' installed successfully to service '%s'\n", packName, serviceName)
	}

	return nil
}

func runPackInfo(cmd *cobra.Command, args []string) error {
	engine := args[0]
	packName := args[1]

	// Create Docker client
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close()

	// Create seed pack service
	packSvc := seedpack.NewService(docker)

	// Get pack info
	manifest, err := packSvc.Info(engine, packName)
	if err != nil {
		return fmt.Errorf("failed to get pack info: %w", err)
	}

	// Display pack information
	fmt.Printf("# %s\n\n", manifest.GetDisplayTitle())
	fmt.Printf("%s\n\n", manifest.Description)

	fmt.Printf("## Information\n\n")
	fmt.Printf("- **Name:** %s\n", manifest.GetFullName())
	fmt.Printf("- **Author:** %s\n", manifest.Author)
	fmt.Printf("- **Engine:** %s\n", manifest.Engine)
	fmt.Printf("- **License:** %s\n", manifest.License)
	fmt.Printf("- **Data Size:** %s\n", manifest.FormatSize())
	fmt.Printf("- **Created:** %s\n", manifest.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("- **Updated:** %s\n", manifest.UpdatedAt.Format("2006-01-02 15:04:05"))

	if manifest.Homepage != "" {
		fmt.Printf("- **Homepage:** %s\n", manifest.Homepage)
	}
	if manifest.Repository != "" {
		fmt.Printf("- **Repository:** %s\n", manifest.Repository)
	}

	if len(manifest.Tags) > 0 {
		fmt.Printf("\n## Tags\n\n")
		for _, tag := range manifest.Tags {
			fmt.Printf("- %s\n", tag)
		}
	}

	if len(manifest.UseCases) > 0 {
		fmt.Printf("\n## Use Cases\n\n")
		for _, useCase := range manifest.UseCases {
			fmt.Printf("- %s\n", useCase)
		}
	}

	if len(manifest.Examples) > 0 {
		fmt.Printf("\n## Examples\n\n")
		for _, example := range manifest.Examples {
			fmt.Printf("### %s\n\n", example.Title)
			fmt.Printf("%s\n\n", example.Description)
			fmt.Printf("```sql\n%s\n```\n\n", example.Query)
			if example.Expected != "" {
				fmt.Printf("Expected result: %s\n\n", example.Expected)
			}
		}
	}

	if len(manifest.Dependencies) > 0 {
		fmt.Printf("\n## Dependencies\n\n")
		for _, dep := range manifest.Dependencies {
			optional := ""
			if dep.Optional {
				optional = " (optional)"
			}
			fmt.Printf("- %s@%s%s\n", dep.Name, dep.Version, optional)
		}
	}

	fmt.Printf("\n## Installation\n\n")
	fmt.Printf("```bash\nnizam pack install <service> %s\n```\n", manifest.GetFullName())

	return nil
}

func runPackRemove(cmd *cobra.Command, args []string) error {
	engine := args[0]
	packName := args[1]

	// Parse flags
	version, _ := cmd.Flags().GetString("version")

	// Create Docker client
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close()

	// Create seed pack service
	packSvc := seedpack.NewService(docker)

	// Remove pack
	if err := packSvc.Remove(engine, packName, version); err != nil {
		return fmt.Errorf("failed to remove seed pack: %w", err)
	}

	if version != "" {
		fmt.Printf("Seed pack '%s@%s' removed successfully\n", packName, version)
	} else {
		fmt.Printf("All versions of seed pack '%s' removed successfully\n", packName)
	}

	return nil
}
