package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/abdultolba/nizam/internal/tui/styles"
)

// PulseAnimation creates a pulsing effect for text
func PulseAnimation(text string, phase float64) string {
	intensity := (math.Sin(phase) + 1) / 2 // 0 to 1
	
	// Create gradient effect based on intensity
	if intensity > 0.7 {
		return lipgloss.NewStyle().Foreground(styles.TronCyan).Render(text)
	} else if intensity > 0.4 {
		return lipgloss.NewStyle().Foreground(styles.TronBlue).Render(text)
	} else {
		return lipgloss.NewStyle().Foreground(styles.TronPurple).Render(text)
	}
}

// ProgressBar creates an animated progress bar
func ProgressBar(width int, progress float64, animated bool, phase float64) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	
	filled := int(float64(width) * progress)
	empty := width - filled
	
	var bar strings.Builder
	
	// Filled portion
	for i := 0; i < filled; i++ {
		if animated && float64(i)/float64(width) > progress-0.1 {
			// Animate the leading edge
			intensity := math.Sin(phase + float64(i)*0.5)
			if intensity > 0 {
				bar.WriteString(lipgloss.NewStyle().Foreground(styles.TronCyan).Render("█"))
			} else {
				bar.WriteString(lipgloss.NewStyle().Foreground(styles.TronBlue).Render("█"))
			}
		} else {
			bar.WriteString(lipgloss.NewStyle().Foreground(styles.TronCyan).Render("█"))
		}
	}
	
	// Empty portion
	for i := 0; i < empty; i++ {
		bar.WriteString(lipgloss.NewStyle().Foreground(styles.TronGrayDark).Render("░"))
	}
	
	return bar.String()
}

// MatrixRain creates a matrix-style rain effect
func MatrixRain(width, height int, phase float64) string {
	var lines []string
	
	chars := []rune{'0', '1', '⚡', '●', '◆', '▲', '▼', '◄', '►', '○'}
	
	for row := 0; row < height; row++ {
		var line strings.Builder
		for col := 0; col < width; col++ {
			// Create pseudo-random pattern based on position and phase
			seed := float64(row*width+col) + phase*2
			intensity := math.Sin(seed) * math.Cos(seed*1.3)
			
			if intensity > 0.8 {
				char := chars[int(seed)%len(chars)]
				line.WriteString(lipgloss.NewStyle().Foreground(styles.TronCyan).Render(string(char)))
			} else if intensity > 0.6 {
				char := chars[int(seed*1.7)%len(chars)]
				line.WriteString(lipgloss.NewStyle().Foreground(styles.TronBlue).Render(string(char)))
			} else if intensity > 0.4 {
				char := chars[int(seed*0.3)%len(chars)]
				line.WriteString(lipgloss.NewStyle().Foreground(styles.TronPurple).Render(string(char)))
			} else {
				line.WriteString(" ")
			}
		}
		lines = append(lines, line.String())
	}
	
	return strings.Join(lines, "\n")
}

// GlitchText creates a glitch effect on text
func GlitchText(text string, phase float64, intensity float64) string {
	if intensity == 0 {
		return text
	}
	
	glitchChars := []rune{'█', '▓', '▒', '░', '▄', '▀', '▐', '▌'}
	runes := []rune(text)
	
	var result strings.Builder
	for i, r := range runes {
		// Create pseudo-random glitch based on position and phase
		seed := float64(i) + phase*10
		glitchProb := math.Sin(seed) * intensity
		
		if glitchProb > 0.7 {
			glitchChar := glitchChars[int(seed)%len(glitchChars)]
			result.WriteString(lipgloss.NewStyle().Foreground(styles.TronPink).Render(string(glitchChar)))
		} else if glitchProb > 0.4 {
			// Color shift
			result.WriteString(lipgloss.NewStyle().Foreground(styles.TronPurple).Render(string(r)))
		} else {
			result.WriteString(lipgloss.NewStyle().Foreground(styles.TronCyan).Render(string(r)))
		}
	}
	
	return result.String()
}

// SpinningIcon creates a spinning icon animation
func SpinningIcon(phase float64, size int) string {
	icons := []string{"◐", "◓", "◑", "◒"}
	index := int(phase*4) % len(icons)
	icon := icons[index]
	
	// Add glow effect
	glowStyle := lipgloss.NewStyle().
		Foreground(styles.TronCyan).
		Bold(true)
	
	return glowStyle.Render(icon)
}

// WaveText creates a wave animation on text
func WaveText(text string, phase float64, amplitude float64) string {
	runes := []rune(text)
	var result strings.Builder
	
	for i, r := range runes {
		// Calculate wave offset for each character
		wave := math.Sin(phase + float64(i)*0.5) * amplitude
		
		// Apply color based on wave position
		if wave > 0.5 {
			result.WriteString(lipgloss.NewStyle().Foreground(styles.TronCyan).Render(string(r)))
		} else if wave > 0 {
			result.WriteString(lipgloss.NewStyle().Foreground(styles.TronBlue).Render(string(r)))
		} else if wave > -0.5 {
			result.WriteString(lipgloss.NewStyle().Foreground(styles.TronPurple).Render(string(r)))
		} else {
			result.WriteString(lipgloss.NewStyle().Foreground(styles.TronWhite).Render(string(r)))
		}
	}
	
	return result.String()
}

// CreateServiceStatusIndicator creates an animated status indicator
func CreateServiceStatusIndicator(status string, phase float64) string {
	switch status {
	case "running":
		// Pulsing green dot
		intensity := (math.Sin(phase*3) + 1) / 2
		if intensity > 0.7 {
			return lipgloss.NewStyle().Foreground(styles.TronCyan).Render("●")
		} else {
			return lipgloss.NewStyle().Foreground(styles.TronCyanDark).Render("●")
		}
		
	case "starting":
		// Spinning indicator
		return SpinningIcon(phase, 1)
		
	case "stopped":
		return lipgloss.NewStyle().Foreground(styles.TronGray).Render("●")
		
	case "error", "failed":
		// Blinking red
		if math.Sin(phase*4) > 0 {
			return lipgloss.NewStyle().Foreground(styles.TronPink).Render("●")
		} else {
			return lipgloss.NewStyle().Foreground(styles.TronGrayDark).Render("●")
		}
		
	default:
		return lipgloss.NewStyle().Foreground(styles.TronWhite).Render("●")
	}
}

// CreateBorderGlow creates an animated border glow effect
func CreateBorderGlow(width, height int, phase float64) string {
	intensity := (math.Sin(phase*2) + 1) / 2
	
	var color lipgloss.Color
	if intensity > 0.7 {
		color = styles.TronCyan
	} else if intensity > 0.4 {
		color = styles.TronBlue
	} else {
		color = styles.TronPurple
	}
	
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Width(width).
		Height(height)
	
	return style.Render("")
}

// TypewriterEffect simulates typing text
func TypewriterEffect(text string, progress float64) string {
	if progress <= 0 {
		return ""
	}
	if progress >= 1 {
		return text
	}
	
	runes := []rune(text)
	length := len(runes)
	visibleLength := int(float64(length) * progress)
	
	if visibleLength > length {
		visibleLength = length
	}
	
	result := string(runes[:visibleLength])
	
	// Add cursor
	if visibleLength < length {
		cursor := lipgloss.NewStyle().Foreground(styles.TronCyan).Render("▊")
		result += cursor
	}
	
	return result
}

// CreateStatusBadge creates an animated status badge
func CreateStatusBadge(status string, phase float64) string {
	var badge string
	var style lipgloss.Style
	
	switch status {
	case "running":
		badge = "ONLINE"
		intensity := (math.Sin(phase*2) + 1) / 2
		if intensity > 0.5 {
			style = styles.ButtonStyle.Copy().Background(styles.TronCyan)
		} else {
			style = styles.ButtonStyle.Copy().Background(styles.TronCyanDark)
		}
		
	case "stopped":
		badge = "OFFLINE"
		style = styles.ButtonInactiveStyle.Copy()
		
	case "starting":
		badge = "INIT..."
		style = styles.ButtonHoverStyle.Copy()
		
	case "error":
		badge = "ERROR"
		// Blinking effect
		if math.Sin(phase*4) > 0 {
			style = styles.ButtonStyle.Copy().Background(styles.TronPink)
		} else {
			style = styles.ButtonStyle.Copy().Background(styles.TronGrayDark)
		}
		
	default:
		badge = "UNKNOWN"
		style = styles.ButtonInactiveStyle.Copy()
	}
	
	return style.Render(fmt.Sprintf(" %s ", badge))
}

// GetCurrentPhase calculates the current animation phase
func GetCurrentPhase() float64 {
	return float64(time.Now().UnixNano()) / float64(time.Second)
}
