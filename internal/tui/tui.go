package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abdultolba/nizam/internal/tui/components"
	"github.com/abdultolba/nizam/internal/tui/models"
	"github.com/abdultolba/nizam/internal/tui/styles"
)

type Model struct {
	App models.AppModel
	Operations *ServiceOperations
}

// NewModel creates a new TUI model
func NewModel() Model {
	// Initialize log channels
	InitLogChannels()
	
	// Create service operations
	ops, err := NewServiceOperations()
	if err != nil {
		// Handle error gracefully - could show error in TUI
		panic(fmt.Sprintf("Failed to initialize Docker client: %v", err))
	}
	
	return Model{
		App: models.NewAppModel(),
		Operations: ops,
	}
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.App.Init(),
		// Initial refresh of services
		m.Operations.RefreshServices(),
		// Start listening for log streams
		m.listenForLogStreams(),
	)
}

// listenForLogStreams creates a command to listen for log stream messages
func (m Model) listenForLogStreams() tea.Cmd {
	return func() tea.Msg {
		// This function will listen to the log channels and convert to tea messages
		logChan := GetLogStreamChan()
		errChan := GetLogStreamErrChan()
		
		select {
		case logMsg := <-logChan:
			return logMsg
		case errMsg := <-errChan:
			return errMsg
		default:
			return nil
		}
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.App.Width = msg.Width
		m.App.Height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "h", "?":
			if m.App.CurrentView == models.HelpView {
				m.App.NavigateToView(m.App.PrevView)
			} else {
				m.App.NavigateToView(models.HelpView)
			}
		case "r":
			m.App.SetStatus("Refreshing services...")
			return m, func() tea.Msg {
				// Simulate refresh
				time.Sleep(500 * time.Millisecond)
				return models.RefreshMsg{}
			}
		case "1":
			m.App.NavigateToView(models.DashboardView)
		case "2":
			m.App.NavigateToView(models.ServicesView)
		case "3":
			m.App.NavigateToView(models.LogsView)
		case "4":
			m.App.NavigateToView(models.TemplatesView)
		case "5":
			m.App.NavigateToView(models.ConfigView)
		case "up", "k":
			if m.App.CurrentView == models.LogsView {
				if m.App.SelectedServiceIndex > 0 {
					m.App.SelectedServiceIndex--
					// Start log streaming for newly selected service
					if len(m.App.Services) > 0 {
						selectedService := m.App.Services[m.App.SelectedServiceIndex]
						m.App.StartLogStreaming(selectedService.Name)
						return m, m.Operations.StreamLogs(selectedService.Name, true)
					}
				}
			}
		case "down", "j":
			if m.App.CurrentView == models.LogsView {
				if m.App.SelectedServiceIndex < len(m.App.Services)-1 {
					m.App.SelectedServiceIndex++
					// Start log streaming for newly selected service
					if len(m.App.Services) > 0 {
						selectedService := m.App.Services[m.App.SelectedServiceIndex]
						m.App.StartLogStreaming(selectedService.Name)
						return m, m.Operations.StreamLogs(selectedService.Name, true)
					}
				}
			}
		case "enter", "space":
			if m.App.CurrentView == models.DashboardView {
				// Handle quick actions
				switch m.App.ActivePanel {
				case 0: // Start All
					m.App.SetStatus("Starting all services... (run 'nizam up' in CLI)")
				case 1: // Stop All
					m.App.SetStatus("Stopping all services... (run 'nizam down' in CLI)")
				case 2: // Refresh
					return m, func() tea.Msg {
						time.Sleep(500 * time.Millisecond)
						return models.RefreshMsg{}
					}
				case 3: // Add Service
					m.App.NavigateToView(models.TemplatesView)
				}
			} else if m.App.CurrentView == models.LogsView {
				// Show logs for selected service
				if len(m.App.Services) > 0 {
					selected := m.App.Services[m.App.SelectedServiceIndex]
					m.App.SetStatus(fmt.Sprintf("Viewing logs for %s (run 'nizam logs %s' in CLI)", selected.Name, selected.Name))
				}
			}
		case "tab":
			if m.App.CurrentView == models.DashboardView {
				m.App.NextPanel()
			}
		case "shift+tab":
			if m.App.CurrentView == models.DashboardView {
				m.App.PrevPanel()
			}
		case "esc":
			if m.App.ShowHelp {
				m.App.ShowHelp = false
			} else if m.App.CurrentView != models.DashboardView {
				m.App.NavigateToView(models.DashboardView)
			}
		}

	case models.TickMsg:
		var cmd tea.Cmd
		m.App.Spinner, cmd = m.App.Spinner.Update(msg)
		return m, tea.Batch(cmd, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return models.TickMsg(t)
		}))

	case models.RefreshMsg:
		// Use real Docker service refresh
		return m, m.Operations.RefreshServices()

	case models.RealServiceStatusMsg:
		// Convert enhanced services to regular services for display
		enhancedServices := []models.EnhancedService(msg)
		regularServices := make([]models.Service, len(enhancedServices))
		
		for i, enhanced := range enhancedServices {
			regularServices[i] = models.Service{
				Name:    enhanced.Name,
				Image:   enhanced.Image,
				Status:  enhanced.Status,
				Ports:   enhanced.Ports,
				Healthy: enhanced.Healthy,
				Uptime:  enhanced.Uptime,
				CPU:     enhanced.CPU,
				Memory:  enhanced.Memory,
			}
		}
		
		m.App.UpdateServices(regularServices)
		m.App.SetStatus("Services refreshed from Docker")

	case models.OperationCompleteMsg:
		if msg.Success {
			m.App.SetStatus(fmt.Sprintf("✅ %s %s completed", strings.Title(msg.Operation), msg.Service))
			// Refresh services after successful operation
			return m, m.Operations.RefreshServices()
		} else {
			m.App.SetError(fmt.Sprintf("❌ %s %s failed: %s", strings.Title(msg.Operation), msg.Service, msg.Error))
		}

	case models.LogLineMsg:
		// Handle real-time log messages - add to service logs
		m.App.AddLogLine(msg.ServiceName, msg.Line)
		// Show brief status update
		m.App.SetStatus(fmt.Sprintf("📝 [%s] %s", msg.ServiceName, msg.Line))
		// Continue listening for more logs
		return m, m.listenForLogStreams()

	case models.LogStreamStartMsg:
		m.App.SetStatus(fmt.Sprintf("🔄 Started log stream for %s", msg.ServiceName))
		// Continue listening for log messages
		return m, m.listenForLogStreams()

	case models.LogStreamStopMsg:
		m.App.SetStatus(fmt.Sprintf("⏹️ Log stream stopped for %s", msg.ServiceName))

	case models.LogStreamErrorMsg:
		m.App.SetError(fmt.Sprintf("❌ Log stream error for %s: %s", msg.ServiceName, msg.Error))

	case models.ErrorMsg:
		m.App.SetError(msg.Error)
	}

	return m, nil
}

// View renders the TUI
func (m Model) View() string {
	if m.App.Width == 0 || m.App.Height == 0 {
		return "Initializing..."
	}

	// Create the main layout
	var content string

	// Render based on current view
	switch m.App.CurrentView {
	case models.DashboardView:
		content = m.renderDashboard()
	case models.ServicesView:
		content = m.renderServices()
	case models.LogsView:
		content = m.renderLogs()
	case models.TemplatesView:
		content = m.renderTemplates()
	case models.ConfigView:
		content = m.renderConfig()
	case models.HelpView:
		content = m.renderHelp()
	}

	// Create main layout with header, content, and footer
	header := m.renderHeader()
	footer := m.renderFooter()
	
	// Calculate content height
	contentHeight := m.App.Height - lipgloss.Height(header) - lipgloss.Height(footer) - 2

	// Apply main app styling
	content = styles.AppStyle.Width(m.App.Width).Height(contentHeight).Render(content)
	
	// Combine all sections
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// renderHeader renders the application header
func (m Model) renderHeader() string {
	// ASCII Art logo
	logo := `
███╗   ██╗██╗███████╗ █████╗ ███╗   ███╗
████╗  ██║██║╚══███╔╝██╔══██╗████╗ ████║
██╔██╗ ██║██║  ███╔╝ ███████║██╔████╔██║
██║╚██╗██║██║ ███╔╝  ██╔══██║██║╚██╔╝██║
██║ ╚████║██║███████╗██║  ██║██║ ╚═╝ ██║
╚═╝  ╚═══╝╚═╝╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝
`

	styledLogo := styles.LogoStyle.Render(logo)
	
	subtitle := styles.HelpStyle.Render("Local Structured Service Manager for Dev Environments")
	
	// Navigation tabs
	tabs := m.renderTabs()
	
	return lipgloss.JoinVertical(lipgloss.Center,
		styledLogo,
		subtitle,
		"",
		tabs,
	)
}

// renderTabs renders the navigation tabs
func (m Model) renderTabs() string {
	var tabs []string
	
	tabItems := []struct {
		name  string
		view  models.ViewState
		key   string
	}{
		{"Dashboard", models.DashboardView, "1"},
		{"Services", models.ServicesView, "2"},
		{"Logs", models.LogsView, "3"},
		{"Templates", models.TemplatesView, "4"},
		{"Config", models.ConfigView, "5"},
	}
	
	for _, item := range tabItems {
		if m.App.CurrentView == item.view {
			tab := styles.ButtonStyle.Render(fmt.Sprintf(" %s (%s) ", item.name, item.key))
			tabs = append(tabs, tab)
		} else {
			tab := styles.ButtonInactiveStyle.Render(fmt.Sprintf(" %s (%s) ", item.name, item.key))
			tabs = append(tabs, tab)
		}
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Center, tabs...)
}

// renderDashboard renders the main dashboard view
func (m Model) renderDashboard() string {
	// Status overview
	running := m.App.GetRunningServices()
	total := m.App.GetTotalServices()
	healthy := m.App.GetHealthyServices()
	
	// Create status cards
	statusCards := []string{
		m.renderStatusCard("Running Services", fmt.Sprintf("%d/%d", running, total), styles.TronCyan),
		m.renderStatusCard("Healthy Services", fmt.Sprintf("%d/%d", healthy, total), styles.TronBlue),
		m.renderStatusCard("Last Updated", m.App.LastUpdated.Format("15:04:05"), styles.TronPurple),
	}
	
	statusRow := lipgloss.JoinHorizontal(lipgloss.Top, statusCards...)
	
	// Quick actions
	actions := m.renderQuickActions()
	
	// Recent services list
	servicesList := m.renderServicesList(true) // compact view
	
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("🚀 Dashboard"),
		"",
		statusRow,
		"",
		actions,
		"",
		servicesList,
	)
}

// renderStatusCard renders a status card
func (m Model) renderStatusCard(title, value string, color lipgloss.Color) string {
	cardStyle := styles.PanelStyle.Copy().
		BorderForeground(color).
		Width(20).
		Height(5)
		
	titleStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Align(lipgloss.Center)
		
	valueStyle := lipgloss.NewStyle().
		Foreground(styles.TronWhite).
		Bold(true).
		Align(lipgloss.Center).
		Padding(1, 0)
	
	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render(title),
		valueStyle.Render(value),
	)
	
	return cardStyle.Render(content)
}

// renderQuickActions renders quick action buttons
func (m Model) renderQuickActions() string {
	// Highlight active panel
	var actions []string
	buttonData := []struct{
		text string
		index int
	}{
		{" ▶ Start All ", 0},
		{" ⏸ Stop All ", 1},
		{" 🔄 Refresh (r) ", 2},
		{" ➕ Add Service ", 3},
	}

	for _, button := range buttonData {
		if m.App.ActivePanel == button.index {
			// Active button with cursor
			actions = append(actions, styles.ButtonStyle.Render("▶ "+button.text))
		} else {
			actions = append(actions, styles.ButtonInactiveStyle.Render(button.text))
		}
	}
	
	// Interactive instructions
	instructions := styles.HelpStyle.Render(
		"Navigation: Tab/Shift+Tab to select | Enter/Space to execute | CLI: 'nizam up/down/add <template>'")
	
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("⚡ Quick Actions"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, actions...),
		"",
		instructions,
	)
}

// renderServicesList renders the services list
func (m Model) renderServicesList(compact bool) string {
	if len(m.App.Services) == 0 {
		emptyMsg := styles.HelpStyle.Render("No services configured. Run 'nizam init' to get started.")
		return styles.PanelStyle.Render(emptyMsg)
	}
	
	var rows []string
	
	// Header
	if !compact {
		header := fmt.Sprintf("%-15s %-20s %-10s %-15s %-10s %-8s",
			"NAME", "IMAGE", "STATUS", "PORTS", "UPTIME", "HEALTH")
		rows = append(rows, styles.TableHeaderStyle.Render(header))
	} else {
		header := fmt.Sprintf("%-15s %-15s %-10s %-10s",
			"NAME", "STATUS", "UPTIME", "HEALTH")
		rows = append(rows, styles.TableHeaderStyle.Render(header))
	}
	
	// Service rows
	phase := components.GetCurrentPhase()
	for _, service := range m.App.Services {
		// Use animated status indicator instead of plain text
		statusIndicator := components.CreateServiceStatusIndicator(service.Status, phase)
		statusBadge := components.CreateStatusBadge(service.Status, phase)
		
		healthIcon := "❌"
		if service.Healthy {
			healthIcon = "✅"
		}
		
		uptimeStr := "0s"
		if service.Uptime > 0 {
			uptimeStr = formatDuration(service.Uptime)
		}
		
		var row string
		if !compact {
			ports := strings.Join(service.Ports, ",")
			if len(ports) > 12 {
				ports = ports[:12] + "..."
			}
			row = fmt.Sprintf("%s %-15s %-20s %-15s %-10s %s",
				statusIndicator,
				service.Name,
				service.Image,
				ports,
				uptimeStr,
				healthIcon,
			)
		} else {
			row = fmt.Sprintf("%s %-12s %s %-8s %s",
				statusIndicator,
				service.Name,
				statusBadge,
				uptimeStr,
				healthIcon,
			)
		}
		
		rows = append(rows, styles.TableRowStyle.Render(row))
	}
	
	content := strings.Join(rows, "\n")
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("📊 Services Overview"),
		"",
		styles.PanelStyle.Render(content),
	)
}

// renderServices renders the detailed services view
func (m Model) renderServices() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("🐳 Services Management"),
		"",
		m.renderServicesList(false),
	)
}

// renderLogs renders the logs view
func (m Model) renderLogs() string {
	if len(m.App.Services) == 0 {
		emptyMsg := styles.HelpStyle.Render("No services available. Run 'nizam init' to get started.")
		return lipgloss.JoinVertical(lipgloss.Left,
			styles.HeaderStyle.Render("📝 Service Logs"),
			"",
			styles.PanelStyle.Render(emptyMsg),
		)
	}

	// Ensure selected index is within bounds
	if m.App.SelectedServiceIndex >= len(m.App.Services) {
		m.App.SelectedServiceIndex = len(m.App.Services) - 1
	}
	if m.App.SelectedServiceIndex < 0 {
		m.App.SelectedServiceIndex = 0
	}

	// Service selection list
	var serviceRows []string
	serviceRows = append(serviceRows, styles.TableHeaderStyle.Render("Select a service to view logs:"))

	phase := components.GetCurrentPhase()
	for i, service := range m.App.Services {
		statusIndicator := components.CreateServiceStatusIndicator(service.Status, phase)
		
		// Highlight selected service
		var rowStyle lipgloss.Style
		if i == m.App.SelectedServiceIndex {
			rowStyle = styles.TableRowSelectedStyle
			cursor := "▶ "
			row := fmt.Sprintf("%s %s%-15s %-15s %-20s",
				cursor, statusIndicator, service.Name, service.Status, service.Image)
			serviceRows = append(serviceRows, rowStyle.Render(row))
		} else {
			rowStyle = styles.TableRowStyle
			row := fmt.Sprintf("  %s %-15s %-15s %-20s",
				statusIndicator, service.Name, service.Status, service.Image)
			serviceRows = append(serviceRows, rowStyle.Render(row))
		}
	}

	serviceList := strings.Join(serviceRows, "\n")
	servicesPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronCyan).Render(serviceList)

	// Real log content for selected service
	selectedService := m.App.Services[m.App.SelectedServiceIndex]
	logContent := m.renderRealLogs(selectedService)
	logsPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronBlue).Render(logContent)

	// Instructions
	instructions := styles.HelpStyle.Render(
		"Navigation: ↑/↓ or j/k to select service | Enter to view logs | Press 'h' for help")

	// Arrange side by side
	content := lipgloss.JoinHorizontal(lipgloss.Top, servicesPanel, "  ", logsPanel)

	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("📝 Service Logs"),
		"",
		content,
		"",
		instructions,
	)
}

// renderTemplates renders the templates view
func (m Model) renderTemplates() string {
	// Available templates
	templates := []struct {
		Name        string
		Category    string
		Description string
		Ports       string
		Status      string
	}{
		{"postgres", "Database", "PostgreSQL relational database", "5432", "Available"},
		{"postgres-15", "Database", "PostgreSQL 15 with extensions", "5432", "Available"},
		{"mysql", "Database", "MySQL relational database", "3306", "Available"},
		{"mongodb", "Database", "MongoDB document database", "27017", "Available"},
		{"redis", "Cache", "Redis in-memory data store", "6379", "Available"},
		{"redis-stack", "Cache", "Redis with RedisInsight", "6379,8001", "Available"},
		{"elasticsearch", "Search", "Elasticsearch search engine", "9200", "Available"},
		{"meilisearch", "Search", "Fast search engine", "7700", "Available"},
		{"rabbitmq", "Message Queue", "RabbitMQ message broker", "5672,15672", "Available"},
		{"kafka", "Streaming", "Apache Kafka (via Redpanda)", "9092", "Available"},
		{"prometheus", "Monitoring", "Prometheus metrics collection", "9090", "Available"},
		{"grafana", "Monitoring", "Grafana visualization", "3000", "Available"},
		{"jaeger", "Tracing", "Distributed tracing", "16686", "Available"},
		{"minio", "Storage", "S3-compatible object storage", "9000,9001", "Available"},
		{"mailhog", "Development", "Email testing tool", "1025,8025", "Available"},
	}

	// Create header
	var rows []string
	header := fmt.Sprintf("%-15s %-12s %-30s %-12s %-10s",
		"NAME", "CATEGORY", "DESCRIPTION", "PORTS", "STATUS")
	rows = append(rows, styles.TableHeaderStyle.Render(header))

	// Template rows with colors
	phase := components.GetCurrentPhase()
	for i, template := range templates {
		// Alternate row colors for better readability
		rowStyle := styles.TableRowStyle
		if i%2 == 0 {
			rowStyle = styles.TableRowStyle.Copy().Background(styles.TronGrayDark)
		}

		// Category colors
		var categoryStyle lipgloss.Style
		switch template.Category {
		case "Database":
			categoryStyle = lipgloss.NewStyle().Foreground(styles.TronCyan)
		case "Cache":
			categoryStyle = lipgloss.NewStyle().Foreground(styles.TronBlue)
		case "Search":
			categoryStyle = lipgloss.NewStyle().Foreground(styles.TronPurple)
		case "Message Queue", "Streaming":
			categoryStyle = lipgloss.NewStyle().Foreground(styles.TronPink)
		default:
			categoryStyle = lipgloss.NewStyle().Foreground(styles.TronWhite)
		}

		// Status indicator
		statusIndicator := components.CreateServiceStatusIndicator("running", phase)

		row := fmt.Sprintf("%s %-12s %-12s %-30s %-12s",
			statusIndicator,
			template.Name,
			categoryStyle.Render(template.Category),
			template.Description,
			template.Ports,
		)
		rows = append(rows, rowStyle.Render(row))
	}

	// Instructions
	instructions := styles.HelpStyle.Render(
		"\nInstructions:\n" +
		"• Use 'nizam add <template-name>' to add a service\n" +
		"• Example: nizam add postgres\n" +
		"• Use 'nizam add <template-name> --name <custom-name>' for custom names\n" +
		"• Press 'r' to refresh templates")

	content := strings.Join(rows, "\n")
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("📦 Service Templates"),
		"",
		styles.PanelStyle.Render(content),
		"",
		instructions,
	)
}

// renderConfig renders the configuration view
func (m Model) renderConfig() string {
	// Configuration sections
	generalConfig := `
General Settings:
• Profile: dev (active)
• Config File: .nizam.yaml
• Verbose Logging: false
• Auto-refresh: 30s
`

	dockerConfig := `
Docker Settings:
• Docker Host: unix:///var/run/docker.sock
• Network: nizam-network
• Volume Prefix: nizam-
• Container Prefix: nizam_
`

	tuiConfig := `
TUI Settings:
• Theme: Tron (Cyberpunk)
• Animation: enabled
• Refresh Rate: 1s
• Mouse Support: enabled
`

	serviceConfig := `
Service Defaults:
• Restart Policy: unless-stopped
• Memory Limit: 512MB
• CPU Limit: 1.0
• Health Check: enabled
`

	// Style the sections
	generalPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronCyan).Render(generalConfig)
	dockerPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronBlue).Render(dockerConfig)
	tuiPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronPurple).Render(tuiConfig)
	servicePanel := styles.PanelStyle.Copy().BorderForeground(styles.TronPink).Render(serviceConfig)

	// Arrange in two columns
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, generalPanel, "", dockerPanel)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, tuiPanel, "", servicePanel)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, "  ", rightColumn)

	// Instructions
	instructions := styles.HelpStyle.Render(
		"\nConfiguration Management:\n" +
		"• Edit .nizam.yaml to modify settings\n" +
		"• Use --config flag to specify custom config file\n" +
		"• Use --profile flag to switch between profiles (dev, test, prod)\n" +
		"• Press 'r' to reload configuration")

	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("⚙️ Configuration"),
		"",
		columns,
		"",
		instructions,
	)
}

// renderHelp renders the help view
func (m Model) renderHelp() string {
	// Split help into sections for better organization
	navigation := `
Keyboard Navigation:
• 1-5         - Switch between views
• Tab/Shift+Tab - Navigate panels
• h/?         - Toggle this help
• r           - Refresh services/data
• q/Ctrl+C    - Quit application
• Esc         - Go back/close help
`

	views := `
View Details:
• Dashboard (1) - Service overview + quick actions
• Services (2)  - Full service management table
• Logs (3)     - Service logs (future feature)
• Templates (4) - Browse available service templates
• Config (5)    - Configuration and settings
`

	commands := `
Main CLI Commands:
• nizam init              - Initialize configuration
• nizam add <template>    - Add service from template
• nizam up [services...]  - Start services
• nizam down              - Stop all services
• nizam status            - Show service status
• nizam logs <service>    - View service logs
• nizam remove <service>  - Remove service
• nizam templates         - List templates
`

	colors := `
Tron Color Theme:
• Cyan (#00FFF7)    - Active/Running states
• Blue (#0088FF)    - Info/Secondary elements
• Purple (#B300FF)  - System/Metadata
• Pink (#FF0059)    - Errors/Warnings/Alerts
• Gray (#4B4B4B)    - Inactive/Disabled
• White (#E0E0E0)   - Primary text content
`

	// Style each section with different colors
	navPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronCyan).Render(navigation)
	viewsPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronBlue).Render(views)
	commandsPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronPurple).Render(commands)
	colorsPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronPink).Render(colors)

	// Arrange in two columns
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, navPanel, "", viewsPanel)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, commandsPanel, "", colorsPanel)
	helpContent := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, "  ", rightColumn)

	// Footer with additional info
	footerText := styles.HelpStyle.Render(
		"💡 Tip: The TUI is for monitoring - use CLI commands in another terminal for actions")

	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("❓ Help & Usage Guide"),
		"",
		helpContent,
		"",
		footerText,
	)
}

// renderRealLogs renders real log content from the stored service logs
func (m Model) renderRealLogs(service models.Service) string {
	// Get real logs for the service
	logs := m.App.GetServiceLogs(service.Name)
	
	if len(logs) == 0 {
		// Show helpful message if no logs yet
		var message strings.Builder
		message.WriteString(fmt.Sprintf("=== %s Logs ===\n\n", service.Name))
		
		if service.Status == "running" {
			message.WriteString("🔄 Starting log stream...\n")
			message.WriteString("📝 Real-time logs will appear here\n\n")
			message.WriteString(styles.HelpStyle.Render(
				"💡 Tip: Navigate up/down to select different services and view their logs"))
		} else {
			message.WriteString(fmt.Sprintf("⏹️  Service '%s' is not running\n", service.Name))
			message.WriteString("📝 No logs available for stopped services\n\n")
			message.WriteString(styles.HelpStyle.Render(
				fmt.Sprintf("💡 Tip: Start the service with 'nizam up %s' to see logs", service.Name)))
		}
		
		return message.String()
	}
	
	// Build the log display
	var logDisplay strings.Builder
	logDisplay.WriteString(fmt.Sprintf("=== %s Real-Time Logs ===\n\n", service.Name))
	
	// Show recent logs (limit to fit in panel)
	maxLines := 20 // Adjust based on panel height
	startIdx := 0
	if len(logs) > maxLines {
		startIdx = len(logs) - maxLines
	}
	
	for i := startIdx; i < len(logs); i++ {
		logDisplay.WriteString(logs[i])
		logDisplay.WriteString("\n")
	}
	
	// Add status information
	logDisplay.WriteString("\n")
	if m.App.LogFollowing && m.App.CurrentLogService == service.Name {
		logDisplay.WriteString(styles.HelpStyle.Render("🔄 Following logs... (live stream active)"))
	} else {
		logDisplay.WriteString(styles.HelpStyle.Render(
			fmt.Sprintf("📝 Showing %d log lines | Use 'nizam logs %s' for full output", 
				len(logs), service.Name)))
	}
	
	return logDisplay.String()
}

// renderMockLogs renders mock log content for a service
func (m Model) renderMockLogs(service models.Service) string {
	// Generate mock logs based on service status and type
	var logs []string
	timestamp := time.Now().Format("15:04:05")
	
	logs = append(logs, fmt.Sprintf("[%s] === %s Logs ===", timestamp, service.Name))
	
	if service.Status == "running" {
		logs = append(logs, 
			fmt.Sprintf("[%s] Container %s started successfully", timestamp, service.Name),
			fmt.Sprintf("[%s] Listening on port(s): %s", timestamp, strings.Join(service.Ports, ", ")),
			fmt.Sprintf("[%s] Health check passed ✓", timestamp),
		)
		
		// Service-specific logs
		switch service.Name {
		case "postgres":
			logs = append(logs,
				fmt.Sprintf("[%s] PostgreSQL init process complete; ready for connections", timestamp),
				fmt.Sprintf("[%s] database system is ready to accept connections", timestamp),
				fmt.Sprintf("[%s] autovacuum launcher started", timestamp),
			)
		case "redis":
			logs = append(logs,
				fmt.Sprintf("[%s] Redis server started, version 7.0.11", timestamp),
				fmt.Sprintf("[%s] The server is now ready to accept connections", timestamp),
				fmt.Sprintf("[%s] Background saving started by pid 42", timestamp),
			)
		case "meilisearch":
			if service.Status == "running" {
				logs = append(logs,
					fmt.Sprintf("[%s] Meilisearch server started on 0.0.0.0:7700", timestamp),
					fmt.Sprintf("[%s] No master key found; The server will accept unidentified requests", timestamp),
					fmt.Sprintf("[%s] Server is listening on: http://0.0.0.0:7700", timestamp),
				)
			}
		default:
			logs = append(logs,
				fmt.Sprintf("[%s] Application initialized", timestamp),
				fmt.Sprintf("[%s] Ready to serve requests", timestamp),
			)
		}
		
		// Add some recent activity
		logs = append(logs,
			fmt.Sprintf("[%s] CPU: %.1f%%, Memory: %s", timestamp, service.CPU, service.Memory),
			fmt.Sprintf("[%s] Uptime: %s", timestamp, formatDuration(service.Uptime)),
		)
	} else {
		// Stopped service logs
		logs = append(logs,
			fmt.Sprintf("[%s] Container %s is stopped", timestamp, service.Name),
			fmt.Sprintf("[%s] Last seen: %s ago", timestamp, "2h30m"),
			fmt.Sprintf("[%s] Exit code: 0 (normal shutdown)", timestamp),
			fmt.Sprintf("[%s] Use 'nizam up %s' to start", timestamp, service.Name),
		)
	}
	
	// Add instructions
	logs = append(logs, "",
		styles.HelpStyle.Render("💡 Tip: Use 'nizam logs "+service.Name+"' in CLI for real-time logs"),
	)
	
	return strings.Join(logs, "\n")
}

// renderFooter renders the application footer
func (m Model) renderFooter() string {
	leftSection := fmt.Sprintf("Services: %d | Running: %d",
		m.App.GetTotalServices(),
		m.App.GetRunningServices())
		
	middleSection := ""
	if m.App.Loading {
		middleSection = fmt.Sprintf("%s Loading...", m.App.Spinner.View())
	} else if m.App.StatusMsg != "" {
		middleSection = m.App.StatusMsg
	}
	
	rightSection := "Press 'h' for help | 'q' to quit"
	
	// Create footer sections
	left := styles.HelpStyle.Render(leftSection)
	middle := styles.KeyStyle.Render(middleSection)
	right := styles.HelpStyle.Render(rightSection)
	
	// Calculate spacing
	middleWidth := lipgloss.Width(middle)
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	totalPadding := m.App.Width - leftWidth - rightWidth - middleWidth
	
	if totalPadding < 0 {
		totalPadding = 0
	}
	
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding
	
	footer := left + strings.Repeat(" ", leftPadding) + middle + strings.Repeat(" ", rightPadding) + right
	
	return styles.SeparatorStyle.Render(strings.Repeat("─", m.App.Width)) + "\n" + footer
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
}

// RunTUI starts the TUI application
func RunTUI() error {
	m := NewModel()
	
	// Initialize with some mock data
	m.App.UpdateServices([]models.Service{
		{
			Name:    "postgres",
			Image:   "postgres:16",
			Status:  "running",
			Ports:   []string{"5432:5432"},
			Healthy: true,
			Uptime:  2*time.Hour + 30*time.Minute,
			CPU:     12.5,
			Memory:  "256MB",
		},
		{
			Name:    "redis",
			Image:   "redis:7",
			Status:  "running",
			Ports:   []string{"6379:6379"},
			Healthy: true,
			Uptime:  1*time.Hour + 45*time.Minute,
			CPU:     3.2,
			Memory:  "128MB",
		},
		{
			Name:    "meilisearch",
			Image:   "getmeili/meilisearch",
			Status:  "stopped",
			Ports:   []string{"7700:7700"},
			Healthy: false,
			Uptime:  0,
			CPU:     0.0,
			Memory:  "0MB",
		},
	})
	
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
