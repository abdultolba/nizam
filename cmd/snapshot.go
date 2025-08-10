package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/compress"
	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/dockerx"
	"github.com/abdultolba/nizam/internal/snapshot"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage database snapshots",
	Long: `Create, restore, list, and prune database snapshots.

Snapshots capture the state of your databases at a specific point in time,
allowing you to quickly restore or share data states.`,
}

// snapshotCreateCmd creates a new snapshot
var snapshotCreateCmd = &cobra.Command{
	Use:   "create <service>",
	Short: "Create a snapshot of a service database",
	Long: `Create a snapshot of a service database.

The snapshot will be stored in .nizam/snapshots/<service>/<timestamp>-<tag>/
with a manifest.json file and compressed database dump.`,
	Example: `  nizam snapshot create postgres
  nizam snapshot create postgres --tag "before-migration"
  nizam snapshot create redis --compress gzip --note "pre-deploy state"`,
	Args: cobra.ExactArgs(1),
	RunE: runSnapshotCreate,
}

// snapshotListCmd lists snapshots
var snapshotListCmd = &cobra.Command{
	Use:   "list [service]",
	Short: "List snapshots",
	Long: `List snapshots for a specific service or all services.

Without a service argument, lists all snapshots across all services.`,
	Example: `  nizam snapshot list
  nizam snapshot list postgres
  nizam snapshot list --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSnapshotList,
}

// snapshotRestoreCmd restores a snapshot
var snapshotRestoreCmd = &cobra.Command{
	Use:   "restore <service>",
	Short: "Restore a snapshot",
	Long: `Restore a snapshot for a service.

By default, restores the latest snapshot. Use --tag, --latest, or --before
to specify which snapshot to restore.`,
	Example: `  nizam snapshot restore postgres
  nizam snapshot restore postgres --tag "before-migration"
  nizam snapshot restore postgres --latest
  nizam snapshot restore postgres --before "2025-08-01 12:00"`,
	Args: cobra.ExactArgs(1),
	RunE: runSnapshotRestore,
}

// snapshotPruneCmd prunes old snapshots
var snapshotPruneCmd = &cobra.Command{
	Use:   "prune <service>",
	Short: "Remove old snapshots",
	Long: `Remove old snapshots, keeping only the N most recent ones.

This helps manage disk space by removing outdated snapshots while
preserving the most recent ones.`,
	Example: `  nizam snapshot prune postgres --keep 5
  nizam snapshot prune postgres --keep 3 --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runSnapshotPrune,
}

func init() {
	rootCmd.AddCommand(snapshotCmd)

	// Add subcommands
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotRestoreCmd)
	snapshotCmd.AddCommand(snapshotPruneCmd)

	// Create command flags
	snapshotCreateCmd.Flags().String("tag", "", "tag for the snapshot")
	snapshotCreateCmd.Flags().String("note", "", "note/description for the snapshot")
	snapshotCreateCmd.Flags().String("compress", "zstd", "compression type: zstd, gzip, none")

	// List command flags
	snapshotListCmd.Flags().Bool("json", false, "output in JSON format")

	// Restore command flags
	snapshotRestoreCmd.Flags().String("tag", "", "restore specific tag")
	snapshotRestoreCmd.Flags().Bool("latest", false, "restore latest snapshot")
	snapshotRestoreCmd.Flags().String("before", "", "restore latest snapshot before timestamp (YYYY-MM-DD HH:MM)")
	snapshotRestoreCmd.Flags().Bool("force", false, "force restore even if errors occur")

	// Prune command flags
	snapshotPruneCmd.Flags().Int("keep", 3, "number of snapshots to keep")
	snapshotPruneCmd.Flags().Bool("dry-run", false, "show what would be removed without removing")
	snapshotPruneCmd.MarkFlagRequired("keep")
}

func runSnapshotCreate(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Parse flags
	tag, _ := cmd.Flags().GetString("tag")
	note, _ := cmd.Flags().GetString("note")
	compressFlag, _ := cmd.Flags().GetString("compress")

	// Validate compression
	var compression compress.Compression
	switch strings.ToLower(compressFlag) {
	case "zstd":
		compression = compress.CompZstd
	case "gzip":
		compression = compress.CompGzip
	case "none":
		compression = compress.CompNone
	default:
		return fmt.Errorf("invalid compression type: %s (must be: zstd, gzip, none)", compressFlag)
	}

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

	// Create snapshot service
	snapshotSvc := snapshot.NewService(docker)

	// Create snapshot
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	opts := snapshot.CreateOptions{
		Tag:         tag,
		Note:        note,
		Compression: compression,
	}

	manifest, err := snapshotSvc.Create(ctx, cfg, serviceName, opts)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	fmt.Printf("Snapshot created successfully:\n")
	fmt.Printf("  Service: %s\n", manifest.Service)
	fmt.Printf("  Tag: %s\n", manifest.Tag)
	fmt.Printf("  Created: %s\n", manifest.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Compression: %s\n", manifest.Compression)
	if manifest.Note != "" {
		fmt.Printf("  Note: %s\n", manifest.Note)
	}

	return nil
}

func runSnapshotList(cmd *cobra.Command, args []string) error {
	// Parse flags
	jsonOutput, _ := cmd.Flags().GetBool("json")

	var serviceName string
	if len(args) > 0 {
		serviceName = args[0]
	}

	// Create Docker client (needed for snapshot service)
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close()

	// Create snapshot service
	snapshotSvc := snapshot.NewService(docker)

	// List snapshots
	snapshots, err := snapshotSvc.List(serviceName)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		if serviceName != "" {
			fmt.Printf("No snapshots found for service '%s'\n", serviceName)
		} else {
			fmt.Println("No snapshots found")
		}
		return nil
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(snapshots)
	}

	// Table output - prepare data
	headers := []string{"Service", "Tag", "Created", "Age", "Size", "Engine", "Note"}
	rows := [][]string{}

	for _, snapshot := range snapshots {
		tag := snapshot.Tag
		if tag == "" {
			tag = "-"
		}

		note := snapshot.Note
		if len(note) > 30 {
			note = note[:27] + "..."
		}
		if note == "" {
			note = "-"
		}

		rows = append(rows, []string{
			snapshot.Service,
			tag,
			snapshot.CreatedAt.Format("2006-01-02 15:04"),
			snapshot.GetAge(),
			snapshot.FormatSize(),
			snapshot.Engine,
			note,
		})
	}

	// Create and render table with new API
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader(headers),
	)

	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
	fmt.Printf("\nTotal: %d snapshots\n", len(snapshots))

	return nil
}

func runSnapshotRestore(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Parse flags
	tag, _ := cmd.Flags().GetString("tag")
	latest, _ := cmd.Flags().GetBool("latest")
	beforeStr, _ := cmd.Flags().GetString("before")
	force, _ := cmd.Flags().GetBool("force")

	// Parse before timestamp
	var beforeTime *time.Time
	if beforeStr != "" {
		// Try multiple formats
		formats := []string{
			"2006-01-02 15:04",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}

		var parsed time.Time
		var parseErr error
		for _, format := range formats {
			parsed, parseErr = time.Parse(format, beforeStr)
			if parseErr == nil {
				beforeTime = &parsed
				break
			}
		}

		if parseErr != nil {
			return fmt.Errorf("invalid timestamp format: %s (use YYYY-MM-DD HH:MM)", beforeStr)
		}
	}

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

	// Create snapshot service
	snapshotSvc := snapshot.NewService(docker)

	// Restore snapshot
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	opts := snapshot.RestoreOptions{
		Tag:    tag,
		Latest: latest,
		Before: beforeTime,
		Force:  force,
	}

	if err := snapshotSvc.Restore(ctx, cfg, serviceName, opts); err != nil {
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	fmt.Printf("Snapshot restored successfully for service '%s'\n", serviceName)
	return nil
}

func runSnapshotPrune(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Parse flags
	keep, _ := cmd.Flags().GetInt("keep")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Create Docker client (needed for snapshot service)
	docker, err := dockerx.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer docker.Close()

	// Create snapshot service
	snapshotSvc := snapshot.NewService(docker)

	// Prune snapshots
	opts := snapshot.PruneOptions{
		Keep:   keep,
		DryRun: dryRun,
	}

	if err := snapshotSvc.Prune(serviceName, opts); err != nil {
		return fmt.Errorf("failed to prune snapshots: %w", err)
	}

	return nil
}
