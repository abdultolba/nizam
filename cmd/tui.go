package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/abdultolba/nizam/internal/tui"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI (Terminal User Interface)",
	Long: `Launch the interactive TUI for managing your services with a beautiful, 
Tron-themed terminal interface.

The TUI provides:
• Real-time service monitoring
• Interactive service management
• Live log viewing
• Template browsing
• Configuration editing

Navigation:
• 1-5: Switch between views (Dashboard, Services, Logs, Templates, Config)  
• Tab/Shift+Tab: Navigate panels
• r: Refresh services
• h/?: Toggle help
• q/Ctrl+C: Quit

The TUI uses a futuristic Tron theme with cyan, blue, and purple accents
for an immersive development experience.`,
	Example: `  # Launch the TUI
  nizam tui

  # Launch TUI with verbose logging
  nizam tui --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check for demo flag
		demoFlag, _ := cmd.Flags().GetBool("demo")
		debugFlag, _ := cmd.Flags().GetBool("debug")
		
		fmt.Println("🚀 Launching Nizam Enhanced TUI...")
		fmt.Println("💡 Press 'h' for help, 'q' to quit")
		fmt.Println("")
		
		// Run the enhanced TUI with real operations
		if err := tui.RunEnhancedTUI(demoFlag, debugFlag); err != nil {
			fmt.Printf("❌ Error running TUI: %v\n", err)
			return
		}
		
		fmt.Println("👋 Thanks for using Nizam TUI!")
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	
	// Add TUI-specific flags if needed
	tuiCmd.Flags().BoolP("demo", "d", false, "Run with demo data (for development)")
	tuiCmd.Flags().BoolP("debug", "", false, "Enable debug mode")
}
