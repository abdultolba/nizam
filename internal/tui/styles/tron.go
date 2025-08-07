package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Tron Color Palette
var (
	// Primary colors
	TronBlack   = lipgloss.Color("#0D0D0D") // Deep black background
	TronCyan    = lipgloss.Color("#00FFF7") // Bright cyan - primary accent
	TronBlue    = lipgloss.Color("#0088FF") // Electric blue - secondary accent
	TronPurple  = lipgloss.Color("#B300FF") // Neon purple - tertiary accent
	TronPink    = lipgloss.Color("#FF0059") // Hot pink - warning/error
	TronGray    = lipgloss.Color("#4B4B4B") // Mid gray - borders/dividers
	TronWhite   = lipgloss.Color("#E0E0E0") // Off-white - main text

	// Gradients and variants
	TronCyanDark  = lipgloss.Color("#005A57")
	TronBlueDark  = lipgloss.Color("#003366")
	TronGrayDark  = lipgloss.Color("#2A2A2A")
	TronGrayLight = lipgloss.Color("#666666")
)

// Base styles
var (
	// Main application styles
	AppStyle = lipgloss.NewStyle().
		Background(TronBlack).
		Foreground(TronWhite).
		Padding(0, 1)

	// Title and header styles
	TitleStyle = lipgloss.NewStyle().
		Foreground(TronCyan).
		Bold(true).
		Background(TronBlack).
		Padding(1, 2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(TronCyan).
		Align(lipgloss.Center)

	HeaderStyle = lipgloss.NewStyle().
		Foreground(TronCyan).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(TronGray).
		Padding(0, 1, 1, 1)

	// Panel styles
	PanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(TronGray).
		Background(TronBlack).
		Foreground(TronWhite).
		Padding(1, 2)

	ActivePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(TronCyan).
		Background(TronBlack).
		Foreground(TronWhite).
		Padding(1, 2)

	// Status styles
	RunningStatusStyle = lipgloss.NewStyle().
		Foreground(TronCyan).
		Bold(true)

	StoppedStatusStyle = lipgloss.NewStyle().
		Foreground(TronGray).
		Bold(true)

	ErrorStatusStyle = lipgloss.NewStyle().
		Foreground(TronPink).
		Bold(true)

	WarningStatusStyle = lipgloss.NewStyle().
		Foreground(TronBlue).
		Bold(true)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
		Foreground(TronBlack).
		Background(TronCyan).
		Padding(0, 2).
		Bold(true).
		Border(lipgloss.NormalBorder()).
		BorderForeground(TronCyan)

	ButtonInactiveStyle = lipgloss.NewStyle().
		Foreground(TronWhite).
		Background(TronGrayDark).
		Padding(0, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(TronGray)

	ButtonHoverStyle = lipgloss.NewStyle().
		Foreground(TronBlack).
		Background(TronBlue).
		Padding(0, 2).
		Bold(true).
		Border(lipgloss.NormalBorder()).
		BorderForeground(TronBlue)

	// List item styles
	ListItemStyle = lipgloss.NewStyle().
		Foreground(TronWhite).
		Padding(0, 2)

	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(TronBlack).
		Background(TronCyan).
		Padding(0, 2).
		Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
		Foreground(TronWhite).
		Background(TronGrayDark).
		Border(lipgloss.NormalBorder()).
		BorderForeground(TronGray).
		Padding(0, 1)

	InputFocusStyle = lipgloss.NewStyle().
		Foreground(TronWhite).
		Background(TronGrayDark).
		Border(lipgloss.NormalBorder()).
		BorderForeground(TronCyan).
		Padding(0, 1)

	// Help text styles
	HelpStyle = lipgloss.NewStyle().
		Foreground(TronGrayLight).
		Italic(true)

	KeyStyle = lipgloss.NewStyle().
		Foreground(TronCyan).
		Bold(true)

	// Progress bar styles
	ProgressBarStyle = lipgloss.NewStyle().
		Background(TronGrayDark).
		Foreground(TronCyan)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
		Foreground(TronCyan).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(TronGray).
		Padding(0, 1)

	TableRowStyle = lipgloss.NewStyle().
		Foreground(TronWhite).
		Padding(0, 1)

	TableRowSelectedStyle = lipgloss.NewStyle().
		Foreground(TronBlack).
		Background(TronCyan).
		Padding(0, 1).
		Bold(true)

	// Logo/ASCII art style
	LogoStyle = lipgloss.NewStyle().
		Foreground(TronCyan).
		Bold(true).
		Align(lipgloss.Center)

	// Separator styles
	SeparatorStyle = lipgloss.NewStyle().
		Foreground(TronGray).
		Bold(true)
)

// Utility functions
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "running", "healthy", "up":
		return RunningStatusStyle
	case "stopped", "down", "exited":
		return StoppedStatusStyle
	case "error", "failed", "unhealthy":
		return ErrorStatusStyle
	case "starting", "restarting", "warning":
		return WarningStatusStyle
	default:
		return lipgloss.NewStyle().Foreground(TronWhite)
	}
}

func GetServiceColor(serviceName string) lipgloss.Color {
	// Assign colors based on service name hash for consistency
	colors := []lipgloss.Color{TronCyan, TronBlue, TronPurple, TronPink}
	hash := 0
	for _, char := range serviceName {
		hash += int(char)
	}
	return colors[hash%len(colors)]
}

// Create a glowing effect by combining styles
func GlowStyle(baseStyle lipgloss.Style, glowColor lipgloss.Color) lipgloss.Style {
	return baseStyle.Copy().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(glowColor)
}

// Create a subtle gradient effect
func GradientStyle(from, to lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(from).
		Foreground(to)
}
