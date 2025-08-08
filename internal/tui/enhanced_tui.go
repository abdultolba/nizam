package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abdultolba/nizam/internal/tui/models"
	"github.com/abdultolba/nizam/internal/tui/styles"
)

// EnhancedModel wraps the TUI with full operational capabilities
type EnhancedModel struct {
	App                 models.AppModel
	Operations          *ServiceOperations
	ConfigPath          string
	Viewport            viewport.Model
	Ready               bool
	lastViewportContent string
	contentLines        []string // Store content lines ourselves
	contentSet          bool     // Track if content has been set
	cachedConfigView    string   // Cache config view content
	lastConfigUpdate    time.Time // Track when config was last updated
}

// NewEnhancedModel creates a new enhanced TUI model
func NewEnhancedModel() (*EnhancedModel, error) {
	// Initialize log streaming channels
	InitLogChannels()
	
	// Initialize operations
	ops, err := NewServiceOperations()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize operations: %w", err)
	}

	// Create app model
	app := models.NewAppModel()
	
	// Initialize with real data
	err = app.InitializeDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker client: %w", err)
	}

	err = app.LoadConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	app.LoadTemplates()

	return &EnhancedModel{
		App:        app,
		Operations: ops,
		ConfigPath: GetConfigPath(),
		Ready:      false, // Will be set to true once we receive initial WindowSizeMsg
	}, nil
}

// Init initializes the enhanced model
func (m EnhancedModel) Init() tea.Cmd {
	return tea.Batch(
		m.App.Init(),
		m.Operations.RefreshServices(), // Load real service data immediately
		m.subscribeToLogChannels(), // Start listening for log stream messages
	)
}

// Update handles all message types for the enhanced TUI
func (m *EnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.App.Width = msg.Width
		m.App.Height = msg.Height
		
		// Store current viewport content before resize
		var currentContent string
		var currentYOffset int
		if m.Ready {
			currentContent = m.lastViewportContent
			currentYOffset = m.Viewport.YOffset
		}
		
		// Initialize or update viewport
		if !m.Ready {
			m.Viewport = viewport.New(msg.Width, msg.Height-2) // Reserve space for footer
			m.Viewport.HighPerformanceRendering = false
			m.Ready = true
		} else {
			m.Viewport.Width = msg.Width
			m.Viewport.Height = msg.Height - 2
			
			// Restore content after viewport resize
			if currentContent != "" && len(m.contentLines) > 0 {
				m.Viewport.SetContent(currentContent)
				m.contentSet = true
				// Preserve scroll position within bounds using our own line count
				totalLines := len(m.contentLines)
				maxOffset := totalLines - m.Viewport.Height
				if maxOffset < 0 {
					maxOffset = 0
				}
				if currentYOffset > maxOffset {
					currentYOffset = maxOffset
				}
				m.Viewport.YOffset = currentYOffset
			}
		}

	case tea.KeyMsg:
		
		// Handle input modes first
		if m.App.InputMode != models.InputModeNone {
			switch msg.String() {
			case "enter":
				if m.App.InputCallback != nil {
					cmd := m.App.InputCallback(m.App.TextInput.Value())
					m.App.ClearInputMode()
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
				} else {
					m.App.ClearInputMode()
				}
			case "esc":
				m.App.ClearInputMode()
			default:
				var cmd tea.Cmd
				m.App.TextInput, cmd = m.App.TextInput.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
			return m, tea.Batch(cmds...)
		}

		// Handle confirmation dialogs
		if m.App.ConfirmOperation {
			switch msg.String() {
			case "y", "Y", "enter":
				if m.App.ConfirmCallback != nil {
					cmd := m.App.ConfirmCallback()
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
				}
				m.App.HideConfirmDialog()
			case "n", "N", "esc":
				m.App.HideConfirmDialog()
			}
			return m, tea.Batch(cmds...)
		}

		// Global key handlers
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
			// Refresh services and clear config cache
			m.cachedConfigView = "" // Force config view refresh
			m.App.SetStatus("Refreshing services...")
			cmds = append(cmds, m.Operations.RefreshServices())
		case "/":
			// Start search
			m.App.SetInputMode(models.InputModeSearch, "Search: ", func(query string) tea.Cmd {
				m.App.SearchQuery = query
				m.App.ClearInputMode()
				return nil
			})
		case "esc":
			// Clear search or go back
			if m.App.SearchQuery != "" {
				m.App.SearchQuery = ""
			} else if m.App.ShowHelp {
				m.App.ShowHelp = false
			} else if m.App.CurrentView != models.DashboardView {
				m.App.NavigateToView(models.DashboardView)
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
		// Viewport scrolling controls
		case "ctrl+u":
			m.Viewport.YOffset -= 5
			if m.Viewport.YOffset < 0 {
				m.Viewport.YOffset = 0
			}
			return m, nil
		case "ctrl+d":
			totalLines := len(m.contentLines)
			maxOffset := totalLines - m.Viewport.Height
			if maxOffset < 0 {
				maxOffset = 0
			}
			m.Viewport.YOffset += 5
			if m.Viewport.YOffset > maxOffset {
				m.Viewport.YOffset = maxOffset
			}
			return m, nil
		case "ctrl+b":
			m.Viewport.YOffset -= m.Viewport.Height
			if m.Viewport.YOffset < 0 {
				m.Viewport.YOffset = 0
			}
			return m, nil
		case "ctrl+f":
			totalLines := len(m.contentLines)
			maxOffset := totalLines - m.Viewport.Height
			if maxOffset < 0 {
				maxOffset = 0
			}
			m.Viewport.YOffset += m.Viewport.Height
			if m.Viewport.YOffset > maxOffset {
				m.Viewport.YOffset = maxOffset
			}
			return m, nil
		}
		
		switch m.App.CurrentView {
		case models.DashboardView:
			switch msg.String() {
			case "tab":
				m.App.NextPanel()
			case "shift+tab":
				m.App.PrevPanel()
			case "enter", "space":
				switch m.App.ActivePanel {
				case 0: // Start All
					m.App.ShowConfirmDialog("Start all services?", func() tea.Cmd {
						return m.startAllServices()
					})
				case 1: // Stop All
					m.App.ShowConfirmDialog("Stop all services?", func() tea.Cmd {
						return m.stopAllServices()
					})
				case 2: // Refresh
					m.App.SetStatus("Refreshing services...")
					cmds = append(cmds, m.Operations.RefreshServices())
				case 3: // Add Service
					m.App.NavigateToView(models.TemplatesView)
				}
			}
		case models.ServicesView:
			filteredServices := m.App.GetFilteredServices()
			switch msg.String() {
			case "up", "k":
				if m.App.SelectedServiceIndex > 0 {
					m.App.SelectedServiceIndex--
				}
			case "down", "j":
				if m.App.SelectedServiceIndex < len(filteredServices)-1 {
					m.App.SelectedServiceIndex++
				}
			case "s":
				// Start selected service
				if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
					serviceName := filteredServices[m.App.SelectedServiceIndex].Name
					m.App.StartOperation(models.OperationStarting, serviceName)
					cmds = append(cmds, m.Operations.StartService(serviceName, m.App.Config))
				}
			case "x":
				// Stop selected service
				if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
					serviceName := filteredServices[m.App.SelectedServiceIndex].Name
					m.App.ShowConfirmDialog(fmt.Sprintf("Stop service '%s'?", serviceName), func() tea.Cmd {
						m.App.StartOperation(models.OperationStopping, serviceName)
						return m.Operations.StopService(serviceName)
					})
				}
			case "R":
				// Restart selected service
				if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
					serviceName := filteredServices[m.App.SelectedServiceIndex].Name
					m.App.ShowConfirmDialog(fmt.Sprintf("Restart service '%s'?", serviceName), func() tea.Cmd {
						m.App.StartOperation(models.OperationRestarting, serviceName)
						return m.Operations.RestartService(serviceName, m.App.Config)
					})
				}
			case "d", "delete":
				// Remove selected service
				if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
					serviceName := filteredServices[m.App.SelectedServiceIndex].Name
					m.App.ShowConfirmDialog(fmt.Sprintf("Remove service '%s'? This will delete it from configuration.", serviceName), func() tea.Cmd {
						m.App.StartOperation(models.OperationRemoving, serviceName)
						return m.Operations.RemoveService(serviceName, m.ConfigPath)
					})
				}
			case "enter":
				// View service logs
				if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
					m.App.NavigateToView(models.LogsView)
				}
			}
		case models.LogsView:
			filteredServices := m.App.GetFilteredServices()
			switch msg.String() {
			case "left", "h":
				// Navigate services
				if m.App.SelectedServiceIndex > 0 {
					m.App.SelectedServiceIndex--
					// Reset scroll when switching services
					m.App.LogScrollOffset = 0
				}
			case "right", "l":
				// Navigate services
				if m.App.SelectedServiceIndex < len(filteredServices)-1 {
					m.App.SelectedServiceIndex++
					// Reset scroll when switching services
					m.App.LogScrollOffset = 0
				}
			case "up", "k":
				// Scroll up in logs
				if m.App.LogScrollOffset > 0 {
					m.App.LogScrollOffset--
				}
			case "down", "j":
				// Scroll down in logs
				if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
					selectedService := filteredServices[m.App.SelectedServiceIndex]
					serviceLogs := m.App.GetFilteredServiceLogs(selectedService.Name)
					
					// We need to calculate the total log lines including wrapped lines and headers
					// Get the same calculation as in renderRealLogs
					servicesPanelWidth := 30
					logsPanelWidth := m.App.Width - servicesPanelWidth - 6
					if servicesPanelWidth > m.App.Width/3 {
						servicesPanelWidth = m.App.Width / 3
					}
					if logsPanelWidth < 50 {
						logsPanelWidth = 50
					}
					
					// Build the same log structure as renderRealLogs
					var allLogLines []string
					if len(serviceLogs) > 0 {
						timestamp := time.Now().Format("15:04:05")
						header := fmt.Sprintf("[%s] === %s Live Logs (Total: %d) ===", timestamp, selectedService.Name, len(serviceLogs))
						allLogLines = append(allLogLines, header, "")
						
						// Add wrapped log lines
						for _, logLine := range serviceLogs {
							wrappedLines := m.wrapLogLine(logLine, logsPanelWidth-4)
							allLogLines = append(allLogLines, wrappedLines...)
						}
						
						// Add status footer
						allLogLines = append(allLogLines, "", 
							fmt.Sprintf("[%s] 📊 Current Status: %s | Uptime: %s | Health: %s", 
								timestamp, selectedService.Status, formatDuration(selectedService.Uptime), 
								func() string { if selectedService.Healthy { return "✅ Healthy" }; return "❌ Unhealthy" }()))
					}
					
					visibleHeight := m.App.Height - 15
					if visibleHeight < 5 {
						visibleHeight = 5
					}
					
					// Check if we can scroll down
					if len(allLogLines) > visibleHeight && m.App.LogScrollOffset < len(allLogLines)-visibleHeight {
						m.App.LogScrollOffset++
					}
				}
			case "enter":
				// Start/stop following logs for selected service
				if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
					serviceName := filteredServices[m.App.SelectedServiceIndex].Name
					m.App.SetStatus(fmt.Sprintf("Starting log stream for %s...", serviceName))
					cmds = append(cmds, m.Operations.StreamLogs(serviceName, true))
				}
			case "c":
				// Clear logs
				m.App.ClearAllLogs()
				// Reset scroll position
				m.App.LogScrollOffset = 0
			case "f":
				// Filter logs
				m.App.SetInputMode(models.InputModeSearch, "Filter logs: ", func(filter string) tea.Cmd {
					m.App.LogFilter = filter
					m.App.ClearInputMode()
					// Reset scroll when filter changes
					m.App.LogScrollOffset = 0
					return nil
				})
			}
		case models.TemplatesView:
			filteredTemplates := m.App.GetFilteredTemplates()
			switch msg.String() {
			case "up", "k":
				if m.App.SelectedTemplateIndex > 0 {
					m.App.SelectedTemplateIndex--
				}
			case "down", "j":
				if m.App.SelectedTemplateIndex < len(filteredTemplates)-1 {
					m.App.SelectedTemplateIndex++
				}
			case "enter", "a":
				// Add service from selected template
				if len(filteredTemplates) > 0 && m.App.SelectedTemplateIndex < len(filteredTemplates) {
					template := filteredTemplates[m.App.SelectedTemplateIndex]
					m.App.SetInputMode(models.InputModeAddService, fmt.Sprintf("Service name for %s: ", template.Name), func(serviceName string) tea.Cmd {
						if serviceName == "" {
							serviceName = template.Name
						}
						m.App.ClearInputMode()
						m.App.StartOperation(models.OperationAdding, serviceName)
						return m.Operations.AddService(template.Name, serviceName, nil, m.ConfigPath)
					})
				}
			}
		case models.ConfigView:
			switch msg.String() {
			case "e":
				// Edit configuration (placeholder - would open editor or config form)
				m.App.SetStatus("Configuration editing not implemented yet - edit .nizam.yaml manually")
			}
		}

	// Handle real operation results
	case models.OperationCompleteMsg:
		m.App.CompleteOperation(msg.Success, 
			fmt.Sprintf("%s %s: %s", strings.Title(msg.Operation), msg.Service, 
				func() string {
					if msg.Success {
						return "Success"
					}
					return msg.Error
				}()))
		// Refresh services after any operation
		cmds = append(cmds, m.Operations.RefreshServices())

	case models.RealServiceStatusMsg:
		// Update with real service data
		services := make([]models.Service, len(msg))
		for i, enhanced := range msg {
			services[i] = models.Service{
				Name:        enhanced.Name,
				Image:       enhanced.Image,
				Status:      enhanced.Status,
				Ports:       enhanced.Ports,
				Environment: enhanced.Environment,
				Healthy:     enhanced.Healthy,
				Uptime:      enhanced.Uptime,
				CPU:         enhanced.CPU,
				Memory:      enhanced.Memory,
			}
		}
		m.App.UpdateServices(services)
		m.App.EnhancedServices = msg
		m.App.SetStatus("Services refreshed")

	case models.ErrorMsg:
		m.App.SetError(msg.Error)

	// Handle log streaming messages
	case models.LogLineMsg:
		// Add log line to service logs
		m.App.AddLogLine(msg.ServiceName, msg.Line)

	// Handle batched log messages
	case LogBatchMsg:
		// Process each log message in the batch
		for _, logMsg := range msg.Messages {
			m.App.AddLogLine(logMsg.ServiceName, logMsg.Line)
		}
		// Continue listening for more logs immediately
		cmds = append(cmds, m.waitForLogLine())
	
	case models.LogStreamStartMsg:
		m.App.SetStatus(fmt.Sprintf("Started log streaming for %s", msg.ServiceName))
		m.App.StartLogStreaming(msg.ServiceName)
	
	case models.LogStreamStopMsg:
		m.App.SetStatus(fmt.Sprintf("Log streaming stopped for %s", msg.ServiceName))
		m.App.StopLogStreaming()
	
	case models.LogStreamErrorMsg:
		m.App.SetError(fmt.Sprintf("Log error for %s: %s", msg.ServiceName, msg.Error))

	// Handle continue listening messages - restart channel listeners
	case ContinueLogListeningMsg:
		cmds = append(cmds, m.waitForLogLine())
	
	case ContinueLogStartListeningMsg:
		cmds = append(cmds, m.waitForLogStreamStart())
	
	case ContinueLogStopListeningMsg:
		cmds = append(cmds, m.waitForLogStreamStop())
	
	case ContinueLogErrorListeningMsg:
		cmds = append(cmds, m.waitForLogStreamError())

	case models.TickMsg:
		var cmd tea.Cmd
		m.App.Spinner, cmd = m.App.Spinner.Update(msg)
		cmds = append(cmds, cmd)
		// Continue the tick for spinner animation
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return models.TickMsg(t)
		}))
		// Auto-refresh every 30 seconds
		if time.Now().Sub(m.App.LastUpdated) > 30*time.Second {
			cmds = append(cmds, m.Operations.RefreshServices())
		}

	default:
		// Update text input if active
		if m.App.InputMode != models.InputModeNone {
			var cmd tea.Cmd
			m.App.TextInput, cmd = m.App.TextInput.Update(msg)
			cmds = append(cmds, cmd)
		}
		
		// Only pass specific message types to viewport to avoid content resets
		// Viewport should only handle mouse events and specific navigation messages
		if m.Ready {
			switch msg.(type) {
			case tea.MouseMsg:
				// Allow viewport to handle mouse events (scrolling, clicking)
				var cmd tea.Cmd
				m.Viewport, cmd = m.Viewport.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			default:
				// Don't pass other message types to viewport to prevent content resets
				// Viewport content is managed explicitly in the View() function
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// handleInputMode handles input mode interactions
func (m *EnhancedModel) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.App.InputCallback != nil {
			cmd := m.App.InputCallback(m.App.TextInput.Value())
			return m, cmd
		}
		m.App.ClearInputMode()
		return m, nil
	case "esc":
		m.App.ClearInputMode()
		return m, nil
	default:
		var cmd tea.Cmd
		m.App.TextInput, cmd = m.App.TextInput.Update(msg)
		return m, cmd
	}
}

// handleConfirmation handles confirmation dialog interactions
func (m *EnhancedModel) handleConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.App.ConfirmCallback != nil {
			cmd := m.App.ConfirmCallback()
			m.App.HideConfirmDialog()
			return m, cmd
		}
		m.App.HideConfirmDialog()
		return m, nil
	case "n", "N", "esc":
		m.App.HideConfirmDialog()
		return m, nil
	}
	return m, nil
}

// handleDashboardKeys handles dashboard-specific keys
func (m EnhancedModel) handleDashboardKeys(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd
	
	switch msg.String() {
	case "tab":
		m.App.NextPanel()
	case "shift+tab":
		m.App.PrevPanel()
	case "enter", "space":
		switch m.App.ActivePanel {
		case 0: // Start All
			m.App.ShowConfirmDialog("Start all services?", func() tea.Cmd {
				return m.startAllServices()
			})
		case 1: // Stop All
			m.App.ShowConfirmDialog("Stop all services?", func() tea.Cmd {
				return m.stopAllServices()
			})
		case 2: // Refresh
			m.App.SetStatus("Refreshing services...")
			cmds = append(cmds, m.Operations.RefreshServices())
		case 3: // Add Service
			m.App.NavigateToView(models.TemplatesView)
		}
	}
	
	return cmds
}

// handleServicesKeys handles services view keys
func (m EnhancedModel) handleServicesKeys(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd
	filteredServices := m.App.GetFilteredServices()
	
	switch msg.String() {
	case "up", "k":
		if m.App.SelectedServiceIndex > 0 {
			m.App.SelectedServiceIndex--
		}
	case "down", "j":
		if m.App.SelectedServiceIndex < len(filteredServices)-1 {
			m.App.SelectedServiceIndex++
		}
	case "s":
		// Start selected service
		if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
			serviceName := filteredServices[m.App.SelectedServiceIndex].Name
			m.App.StartOperation(models.OperationStarting, serviceName)
			cmds = append(cmds, m.Operations.StartService(serviceName, m.App.Config))
		}
	case "x":
		// Stop selected service
		if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
			serviceName := filteredServices[m.App.SelectedServiceIndex].Name
			m.App.ShowConfirmDialog(fmt.Sprintf("Stop service '%s'?", serviceName), func() tea.Cmd {
				m.App.StartOperation(models.OperationStopping, serviceName)
				return m.Operations.StopService(serviceName)
			})
		}
	case "R":
		// Restart selected service
		if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
			serviceName := filteredServices[m.App.SelectedServiceIndex].Name
			m.App.ShowConfirmDialog(fmt.Sprintf("Restart service '%s'?", serviceName), func() tea.Cmd {
				m.App.StartOperation(models.OperationRestarting, serviceName)
				return m.Operations.RestartService(serviceName, m.App.Config)
			})
		}
	case "d", "delete":
		// Remove selected service
		if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
			serviceName := filteredServices[m.App.SelectedServiceIndex].Name
			m.App.ShowConfirmDialog(fmt.Sprintf("Remove service '%s'? This will delete it from configuration.", serviceName), func() tea.Cmd {
				m.App.StartOperation(models.OperationRemoving, serviceName)
				return m.Operations.RemoveService(serviceName, m.ConfigPath)
			})
		}
	case "enter":
		// View service logs
		if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
			m.App.NavigateToView(models.LogsView)
		}
	}
	
	return cmds
}

// handleLogsKeys handles logs view keys
func (m EnhancedModel) handleLogsKeys(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd
	filteredServices := m.App.GetFilteredServices()
	
	switch msg.String() {
	case "up", "k":
		if m.App.SelectedServiceIndex > 0 {
			m.App.SelectedServiceIndex--
		}
	case "down", "j":
		if m.App.SelectedServiceIndex < len(filteredServices)-1 {
			m.App.SelectedServiceIndex++
		}
	case "enter":
		// Start/stop following logs for selected service
		if len(filteredServices) > 0 && m.App.SelectedServiceIndex < len(filteredServices) {
			serviceName := filteredServices[m.App.SelectedServiceIndex].Name
			cmds = append(cmds, m.Operations.StreamLogs(serviceName, true))
		}
	case "c":
		// Clear logs
		m.App.ClearAllLogs()
	case "f":
		// Filter logs
		m.App.SetInputMode(models.InputModeSearch, "Filter logs: ", func(filter string) tea.Cmd {
			m.App.LogFilter = filter
			m.App.ClearInputMode()
			return nil
		})
	}
	
	return cmds
}

// handleTemplatesKeys handles templates view keys
func (m EnhancedModel) handleTemplatesKeys(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd
	filteredTemplates := m.App.GetFilteredTemplates()
	
	switch msg.String() {
	case "up", "k":
		if m.App.SelectedTemplateIndex > 0 {
			m.App.SelectedTemplateIndex--
		}
	case "down", "j":
		if m.App.SelectedTemplateIndex < len(filteredTemplates)-1 {
			m.App.SelectedTemplateIndex++
		}
	case "enter", "a":
		// Add service from selected template
		if len(filteredTemplates) > 0 && m.App.SelectedTemplateIndex < len(filteredTemplates) {
			template := filteredTemplates[m.App.SelectedTemplateIndex]
			m.App.SetInputMode(models.InputModeAddService, fmt.Sprintf("Service name for %s: ", template.Name), func(serviceName string) tea.Cmd {
				if serviceName == "" {
					serviceName = template.Name
				}
				m.App.ClearInputMode()
				m.App.StartOperation(models.OperationAdding, serviceName)
				return m.Operations.AddService(template.Name, serviceName, nil, m.ConfigPath)
			})
		}
	}
	
	return cmds
}

// handleConfigKeys handles config view keys
func (m EnhancedModel) handleConfigKeys(msg tea.KeyMsg) []tea.Cmd {
	var cmds []tea.Cmd
	
	switch msg.String() {
	case "e":
		// Edit configuration (placeholder - would open editor or config form)
		m.App.SetStatus("Configuration editing not implemented yet - edit .nizam.yaml manually")
	}
	
	return cmds
}

// startAllServices starts all configured services
func (m EnhancedModel) startAllServices() tea.Cmd {
	return func() tea.Msg {
		var errors []string
		successCount := 0
		
		for serviceName, serviceConfig := range m.App.Config.Services {
			err := m.Operations.DockerClient.StartService(context.Background(), serviceName, serviceConfig)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", serviceName, err))
			} else {
				successCount++
			}
		}
		
		if len(errors) > 0 {
			return models.OperationCompleteMsg{
				Operation: "start all",
				Service:   "all services",
				Success:   false,
				Error:     fmt.Sprintf("Started %d services, failed %d: %s", successCount, len(errors), strings.Join(errors, "; ")),
			}
		}
		
		return models.OperationCompleteMsg{
			Operation: "start all",
			Service:   "all services",
			Success:   true,
			Error:     "",
		}
	}
}

// stopAllServices stops all running services
func (m EnhancedModel) stopAllServices() tea.Cmd {
	return func() tea.Msg {
		var errors []string
		successCount := 0
		
		for serviceName := range m.App.Config.Services {
			err := m.Operations.DockerClient.StopService(context.Background(), serviceName)
			if err != nil {
				// Don't count "not found" as errors
				if !strings.Contains(err.Error(), "No such container") {
					errors = append(errors, fmt.Sprintf("%s: %v", serviceName, err))
				}
			} else {
				successCount++
			}
		}
		
		if len(errors) > 0 {
			return models.OperationCompleteMsg{
				Operation: "stop all",
				Service:   "all services",
				Success:   false,
				Error:     fmt.Sprintf("Stopped %d services, failed %d: %s", successCount, len(errors), strings.Join(errors, "; ")),
			}
		}
		
		return models.OperationCompleteMsg{
			Operation: "stop all",
			Service:   "all services",
			Success:   true,
			Error:     "",
		}
	}
}

// View renders the enhanced TUI
func (m *EnhancedModel) View() string {
	if !m.Ready || m.App.Width == 0 || m.App.Height == 0 {
		return "Initializing enhanced TUI..."
	}

	// Show confirmation dialog if active
	if m.App.ConfirmOperation {
		return m.renderConfirmationDialog()
	}

	// Show input mode if active
	if m.App.InputMode != models.InputModeNone {
		return m.renderInputMode()
	}

	// Create the full content that will be scrollable
	header := m.renderEnhancedHeader()
	
	// Render main view content (no height constraints - let it be natural)
	var content string
	switch m.App.CurrentView {
	case models.DashboardView:
		content = m.renderEnhancedDashboard()
	case models.ServicesView:
		content = m.renderEnhancedServices()
	case models.LogsView:
		content = m.renderEnhancedLogs()
	case models.TemplatesView:
		content = m.renderEnhancedTemplates()
	case models.ConfigView:
		content = m.renderEnhancedConfig()
	case models.HelpView:
		content = m.renderEnhancedHelp()
	}
	
	// Combine header and content for the viewport
	fullContent := lipgloss.JoinVertical(lipgloss.Left, header, "", content)
	
	
	// Manage viewport content robustly
	newContentLines := strings.Split(fullContent, "\n")
	contentChanged := fullContent != m.lastViewportContent
	
	// Only update if content actually changed or we haven't set content yet
	if !m.contentSet || contentChanged {
		m.Viewport.SetContent(fullContent)
		m.lastViewportContent = fullContent
		m.contentLines = newContentLines
		m.contentSet = true
	}
	
	// Render footer separately (always visible at bottom)
	footer := m.renderEnhancedFooter()
	
	// Return viewport view with footer
	return lipgloss.JoinVertical(lipgloss.Left, m.Viewport.View(), footer)
}

// renderConfirmationDialog renders a confirmation dialog
func (m EnhancedModel) renderConfirmationDialog() string {
	dialog := styles.PanelStyle.Copy().
		Width(60).
		Height(8).
		BorderForeground(styles.TronPink).
		Render(fmt.Sprintf("⚠️  Confirmation\n\n%s\n\n[Y]es / [N]o", m.App.ConfirmMessage))
	
	return lipgloss.Place(m.App.Width, m.App.Height, lipgloss.Center, lipgloss.Center, dialog)
}

// renderInputMode renders the input mode overlay
func (m EnhancedModel) renderInputMode() string {
	prompt := m.App.InputPrompt + m.App.TextInput.View()
	
	dialog := styles.PanelStyle.Copy().
		Width(60).
		Height(5).
		BorderForeground(styles.TronCyan).
		Render(prompt + "\n\nPress Enter to submit, Esc to cancel")
	
	return lipgloss.Place(m.App.Width, m.App.Height, lipgloss.Center, lipgloss.Center, dialog)
}

// Enhanced rendering methods - Full implementations

func (m EnhancedModel) renderEnhancedHeader() string {
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
	
	subtitle := styles.HelpStyle.Render("Enhanced Service Manager - Full Docker Operations")
	
	// Navigation tabs
	tabs := m.renderEnhancedTabs()
	
	return lipgloss.JoinVertical(lipgloss.Center,
		styledLogo,
		subtitle,
		"",
		tabs,
	)
}

func (m EnhancedModel) renderEnhancedTabs() string {
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

func (m EnhancedModel) renderEnhancedDashboard() string {
	// Status overview
	running := m.App.GetRunningServices()
	total := m.App.GetTotalServices()
	healthy := m.App.GetHealthyServices()
	
	// Create status cards
	statusCards := []string{
		m.renderStatusCard("Running", fmt.Sprintf("%d/%d", running, total), styles.TronCyan),
		m.renderStatusCard("Healthy", fmt.Sprintf("%d/%d", healthy, total), styles.TronBlue),
		m.renderStatusCard("Last Updated", m.App.LastUpdated.Format("15:04:05"), styles.TronPurple),
	}
	
	statusRow := lipgloss.JoinHorizontal(lipgloss.Top, statusCards...)
	
	// Quick actions
	actions := m.renderEnhancedQuickActions()
	
	// Recent services list
	servicesList := m.renderEnhancedServicesList(true) // compact view
	
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("🚀 Enhanced Dashboard"),
		"",
		statusRow,
		"",
		actions,
		"",
		servicesList,
	)
}

func (m EnhancedModel) renderStatusCard(title, value string, color lipgloss.Color) string {
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

func (m EnhancedModel) renderEnhancedQuickActions() string {
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
		"Navigation: Tab/Shift+Tab to select | Enter/Space to execute | Ctrl+U/D half-page | Ctrl+B/F full-page | Real Docker operations enabled")
	
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("⚡ Quick Actions (Live Operations)"),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, actions...),
		"",
		instructions,
	)
}

func (m EnhancedModel) renderEnhancedServicesList(compact bool) string {
	if len(m.App.Services) == 0 {
		emptyMsg := styles.HelpStyle.Render("No services configured. Run 'nizam init' or use Templates (4) to add services.")
		return styles.PanelStyle.Render(emptyMsg)
	}
	
	filteredServices := m.App.GetFilteredServices()
	var rows []string
	
	// Header
	if !compact {
		header := fmt.Sprintf("%-15s %-20s %-12s %-15s %-10s %-8s",
			"NAME", "IMAGE", "STATUS", "PORTS", "UPTIME", "HEALTH")
		rows = append(rows, styles.TableHeaderStyle.Render(header))
	} else {
		header := fmt.Sprintf("%-15s %-12s %-10s %-8s",
			"NAME", "STATUS", "UPTIME", "HEALTH")
		rows = append(rows, styles.TableHeaderStyle.Render(header))
	}
	
	// Service rows
	for i, service := range filteredServices {
		// Status styling
		var statusStyle lipgloss.Style
		var statusIcon string
		switch service.Status {
		case "running":
			statusStyle = lipgloss.NewStyle().Foreground(styles.TronCyan)
			statusIcon = "🟢"
		case "stopped":
			statusStyle = lipgloss.NewStyle().Foreground(styles.TronGrayLight)
			statusIcon = "🔴"
		default:
			statusStyle = lipgloss.NewStyle().Foreground(styles.TronPink)
			statusIcon = "🟡"
		}
		
		healthIcon := "❌"
		if service.Healthy {
			healthIcon = "✅"
		}
		
		uptimeStr := "0s"
		if service.Uptime > 0 {
			uptimeStr = formatDuration(service.Uptime)
		}
		
		// Highlight selected row if this is the services view
		var rowStyle lipgloss.Style
		if m.App.CurrentView == models.ServicesView && i == m.App.SelectedServiceIndex {
			rowStyle = styles.TableRowSelectedStyle
		} else {
			rowStyle = styles.TableRowStyle
		}
		
		var row string
		if !compact {
			ports := strings.Join(service.Ports, ",")
			if len(ports) > 12 {
				ports = ports[:12] + "..."
			}
			row = fmt.Sprintf("%s %-15s %-20s %s %-15s %-10s %s",
				statusIcon,
				service.Name,
				service.Image,
				statusStyle.Render(service.Status),
				ports,
				uptimeStr,
				healthIcon,
			)
		} else {
			row = fmt.Sprintf("%s %-12s %s %-8s %s",
				statusIcon,
				service.Name,
				statusStyle.Render(service.Status),
				uptimeStr,
				healthIcon,
			)
		}
		
		rows = append(rows, rowStyle.Render(row))
	}
	
	content := strings.Join(rows, "\n")
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("📊 Live Services Status"),
		"",
		styles.PanelStyle.Render(content),
	)
}

func (m EnhancedModel) renderEnhancedServices() string {
	// Service management instructions
	instructions := styles.HelpStyle.Render(
		"Controls: ↑/↓ Navigate | s=Start | x=Stop | R=Restart | d=Remove | Enter=View Logs | Ctrl+U/D half-page | Ctrl+B/F full-page")
	
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("🐳 Enhanced Service Management"),
		"",
		m.renderEnhancedServicesList(false),
		"",
		instructions,
	)
}

func (m EnhancedModel) renderEnhancedLogs() string {
	filteredServices := m.App.GetFilteredServices()
	if len(filteredServices) == 0 {
		emptyMsg := styles.HelpStyle.Render("No services available for log viewing.")
		return lipgloss.JoinVertical(lipgloss.Left,
			styles.HeaderStyle.Render("📝 Service Logs"),
			"",
			styles.PanelStyle.Render(emptyMsg),
		)
	}

	// Service selection list - make it more compact
	var serviceRows []string
	serviceRows = append(serviceRows, styles.TableHeaderStyle.Render("Services:"))

	for i, service := range filteredServices {
		statusIcon := "🟢"
		if service.Status != "running" {
			statusIcon = "🔴"
		}
		
		// Highlight selected service with compact formatting
		if i == m.App.SelectedServiceIndex {
			cursor := "▶ "
			// Compact format: just name and status
			row := fmt.Sprintf("%s%s %s", cursor, statusIcon, service.Name)
			serviceRows = append(serviceRows, styles.TableRowSelectedStyle.Render(row))
		} else {
			// Compact format without cursor padding
			row := fmt.Sprintf("  %s %s", statusIcon, service.Name)
			serviceRows = append(serviceRows, styles.TableRowStyle.Render(row))
		}
	}

	serviceList := strings.Join(serviceRows, "\n")
	servicesPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronCyan).Render(serviceList)

	// Mock log content for selected service
	selectedService := filteredServices[m.App.SelectedServiceIndex]
	logContent := m.renderRealLogs(selectedService)
	logsPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronBlue).Render(logContent)

	// Instructions with scrolling controls
	instructions := styles.HelpStyle.Render(
		"Navigation: ←/→ h/l select service | ↑/↓ j/k scroll logs | Ctrl+U/D half-page | Ctrl+B/F full-page | Enter=Stream | c=Clear | f=Filter")

	// Calculate layout dimensions to prevent overlap
	servicesPanelWidth := 30 // Reduced width for services panel
	logsPanelWidth := m.App.Width - servicesPanelWidth - 6 // Remaining width minus spacing and borders
	
	// Ensure reasonable widths
	if servicesPanelWidth > m.App.Width/3 {
		servicesPanelWidth = m.App.Width / 3
	}
	if logsPanelWidth < 50 {
		logsPanelWidth = 50
	}
	
	// Apply fixed widths to panels
	servicesPanel = styles.PanelStyle.Copy().BorderForeground(styles.TronCyan).Width(servicesPanelWidth).Render(serviceList)
	logsPanel = styles.PanelStyle.Copy().BorderForeground(styles.TronBlue).Width(logsPanelWidth).Render(logContent)
	
	// Arrange side by side with proper spacing
	content := lipgloss.JoinHorizontal(lipgloss.Top, servicesPanel, "  ", logsPanel)

	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("📝 Enhanced Service Logs"),
		"",
		content,
		"",
		instructions,
	)
}

func (m EnhancedModel) renderRealLogs(service models.Service) string {
	// Calculate logs panel width for wrapping
	servicesPanelWidth := 30 // Reduced width for services panel
	logsPanelWidth := m.App.Width - servicesPanelWidth - 6 // Remaining width minus spacing and borders
	
	// Ensure reasonable widths
	if servicesPanelWidth > m.App.Width/3 {
		servicesPanelWidth = m.App.Width / 3
	}
	if logsPanelWidth < 50 {
		logsPanelWidth = 50
	}
	
	// Get actual stored logs from the AppModel
	serviceLogs := m.App.GetFilteredServiceLogs(service.Name)
	
	if len(serviceLogs) == 0 {
		// Show helpful message when no logs are available
		timestamp := time.Now().Format("15:04:05")
		emptyMsg := []string{
			fmt.Sprintf("[%s] === %s Logs ===", timestamp, service.Name),
			"",
		}
		
		if service.Status == "running" {
			emptyMsg = append(emptyMsg, 
				fmt.Sprintf("[%s] ℹ️  No logs captured yet for %s", timestamp, service.Name),
				fmt.Sprintf("[%s] 🔄 Press Enter to start streaming logs", timestamp),
				fmt.Sprintf("[%s] 📊 Service Status: %s, Ports: %s", timestamp, service.Status, strings.Join(service.Ports, ", ")),
			)
		} else {
			emptyMsg = append(emptyMsg,
				fmt.Sprintf("[%s] ❌ Service %s is %s", timestamp, service.Name, service.Status),
				fmt.Sprintf("[%s] 💡 Start the service first to view logs", timestamp),
			)
		}
		
		return strings.Join(emptyMsg, "\n")
	}
	
	// Display actual logs with a header
	timestamp := time.Now().Format("15:04:05")
	header := fmt.Sprintf("[%s] === %s Live Logs (Total: %d) ===", timestamp, service.Name, len(serviceLogs))
	
	// Show filter info if active
	filterInfo := ""
	if m.App.LogFilter != "" {
		filterInfo = fmt.Sprintf("[%s] 🔍 Filtered by: '%s'", timestamp, m.App.LogFilter)
	}
	
	// Combine header, filter info, and actual logs
	allLogs := []string{header}
	if filterInfo != "" {
		allLogs = append(allLogs, filterInfo, "")
	} else {
		allLogs = append(allLogs, "")
	}
	
	// Add the actual service logs with proper wrapping for long lines
	for _, logLine := range serviceLogs {
		// Wrap long log lines to fit in the panel
		wrappedLines := m.wrapLogLine(logLine, logsPanelWidth-4) // Account for panel borders
		allLogs = append(allLogs, wrappedLines...)
	}
	
	// Show current service status at the end if logs exist
	if len(serviceLogs) > 0 {
		allLogs = append(allLogs, "", 
			fmt.Sprintf("[%s] 📊 Current Status: %s | Uptime: %s | Health: %s", 
				timestamp, service.Status, formatDuration(service.Uptime), 
				func() string { if service.Healthy { return "✅ Healthy" }; return "❌ Unhealthy" }()))
	}
	
	// Apply scrolling to display only the visible portion
	visibleHeight := m.App.Height - 15 // Account for header, footer, instructions
	if visibleHeight < 5 {
		visibleHeight = 5 // Minimum visible height
	}
	
	// Calculate the visible slice of logs based on scroll offset
	scrollOffset := m.App.LogScrollOffset
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	
	// Make sure we don't scroll past the end
	if len(allLogs) > visibleHeight && scrollOffset > len(allLogs)-visibleHeight {
		scrollOffset = len(allLogs) - visibleHeight
		m.App.LogScrollOffset = scrollOffset // Update the model
	}
	
	// Extract the visible portion of logs
	var visibleLogs []string
	if len(allLogs) <= visibleHeight {
		// All logs fit on screen, no scrolling needed
		visibleLogs = allLogs
	} else {
		// Show only the visible portion
		start := scrollOffset
		end := scrollOffset + visibleHeight
		if end > len(allLogs) {
			end = len(allLogs)
		}
		visibleLogs = allLogs[start:end]
		
		// Add scroll indicators if needed
		if scrollOffset > 0 {
			// Add indicator at the top that there are more logs above
			visibleLogs[0] = fmt.Sprintf("⬆️  (%d more lines above) %s", scrollOffset, visibleLogs[0])
		}
		if end < len(allLogs) {
			// Add indicator at the bottom that there are more logs below
			remaining := len(allLogs) - end
			if len(visibleLogs) > 0 {
				visibleLogs = append(visibleLogs, fmt.Sprintf("⬇️  (%d more lines below) Use ↓/j to scroll down", remaining))
			}
		}
	}
	
	return strings.Join(visibleLogs, "\n")
}

func (m EnhancedModel) renderEnhancedTemplates() string {
	// Get available templates
	allTemplates := m.App.Templates
	filteredTemplates := m.App.GetFilteredTemplates()
	
	// Create header
	var rows []string
	header := fmt.Sprintf("%-15s %-12s %-30s %-12s",
		"NAME", "CATEGORY", "DESCRIPTION", "PORTS")
	rows = append(rows, styles.TableHeaderStyle.Render(header))
	
	// Template rows with selection highlighting
	for i, template := range filteredTemplates {
		// Highlight selected template
		var rowStyle lipgloss.Style
		if i == m.App.SelectedTemplateIndex {
			rowStyle = styles.TableRowSelectedStyle
		} else {
			rowStyle = styles.TableRowStyle
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
		default:
			categoryStyle = lipgloss.NewStyle().Foreground(styles.TronWhite)
		}
		
		cursor := ""
		if i == m.App.SelectedTemplateIndex {
			cursor = "▶ "
		} else {
			cursor = "  "
		}
		
		ports := strings.Join(template.Config.Ports, ",")
		if len(ports) > 10 {
			ports = ports[:10] + "..."
		}
		
		row := fmt.Sprintf("%s%-12s %s %-30s %-12s",
			cursor,
			template.Name,
			categoryStyle.Render(template.Category),
			template.Description,
			ports,
		)
		rows = append(rows, rowStyle.Render(row))
	}
	
	// Instructions
	instructions := styles.HelpStyle.Render(
		"Navigation: ↑/↓ Navigate | Enter/a=Add Service | Ctrl+U/D half-page | Ctrl+B/F full-page | Real service creation enabled")
	
	// Search info
	searchInfo := ""
	if m.App.SearchQuery != "" {
		searchInfo = styles.HelpStyle.Render(fmt.Sprintf("Search: %s (%d/%d)", 
			m.App.SearchQuery, len(filteredTemplates), len(allTemplates)))
	}
	
	content := strings.Join(rows, "\n")
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("📦 Enhanced Service Templates"),
		"",
		styles.PanelStyle.Render(content),
		"",
		instructions,
		searchInfo,
	)
}

func (m *EnhancedModel) renderEnhancedConfig() string {
	// Cache config view content and only update every 5 seconds or on explicit refresh
	now := time.Now()
	if m.cachedConfigView != "" && now.Sub(m.lastConfigUpdate) < 5*time.Second {
		return m.cachedConfigView
	}
	
	// Configuration sections with stable data
	generalConfig := fmt.Sprintf(`
General Settings:
• Profile: %s (active)
• Config File: %s
• Total Services: %d
• Docker Client: Connected`,
		m.App.Config.Profile,
		m.ConfigPath,
		len(m.App.Config.Services))
	
	dockerConfig := `
Docker Settings:
• Docker Host: unix:///var/run/docker.sock
• Network: nizam-network
• Volume Prefix: nizam-
• Container Prefix: nizam_
• Live Operations: ✅ Enabled`
	
	tuiConfig := `
Enhanced TUI Settings:
• Theme: Tron (Cyberpunk)
• Real-time Updates: ✅ Enabled
• Service Operations: ✅ Live
• Auto-refresh: Every 30s`
	
	// Current services in config (stable list)
	var servicesList []string
	for name, service := range m.App.Config.Services {
		servicesList = append(servicesList, fmt.Sprintf("• %s: %s", name, service.Image))
	}
	servicesConfig := fmt.Sprintf("\nConfigured Services:\n%s", strings.Join(servicesList, "\n"))
	
	// Style the sections
	generalPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronCyan).Render(generalConfig)
	dockerPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronBlue).Render(dockerConfig)
	tuiPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronPurple).Render(tuiConfig)
	servicePanel := styles.PanelStyle.Copy().BorderForeground(styles.TronPink).Render(servicesConfig)
	
	// Arrange in two columns
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, generalPanel, "", dockerPanel)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, tuiPanel, "", servicePanel)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, "  ", rightColumn)
	
	// Instructions
	instructions := styles.HelpStyle.Render(
		"Configuration Management:\n" +
		"• Edit .nizam.yaml directly to modify settings\n" +
		"• Use Templates (4) to add new services\n" +
		"• Press 'r' to reload configuration\n" +
		"• Ctrl+U/D half-page | Ctrl+B/F full-page scrolling")
	
	content := lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("⚙️ Enhanced Configuration"),
		"",
		columns,
		"",
		instructions,
	)
	
	// Cache the content and update timestamp
	m.cachedConfigView = content
	m.lastConfigUpdate = now
	
	return content
}

func (m EnhancedModel) renderEnhancedHelp() string {
	// Enhanced help with real operations
	navigation := `
Keyboard Navigation:
• 1-5           - Switch between views
• Tab/Shift+Tab - Navigate panels
• h/?           - Toggle this help
• r             - Refresh services (live)
• q/Ctrl+C      - Quit application
• /             - Search services/templates
• Esc           - Clear search/go back

Viewport Scrolling:
• Ctrl+U        - Scroll up (5 lines)
• Ctrl+D        - Scroll down (5 lines)  
• Ctrl+B        - Page up (full screen)
• Ctrl+F        - Page down (full screen)`
	
	operations := `
Live Service Operations:
• s           - Start selected service
• x           - Stop selected service  
• R           - Restart selected service
• d/Delete    - Remove service (with confirmation)
• Enter       - View logs/execute action
• Space       - Execute quick action`
	
	views := `
View Details:
• Dashboard (1) - Live status + quick actions
• Services (2)  - Full service management
• Logs (3)     - Real-time log streaming
• Templates (4) - Add services directly
• Config (5)    - Live configuration view`
	
	features := `
Enhanced Features:
• ✅ Real Docker operations
• ✅ Live service monitoring
• ✅ Direct service management
• ✅ Template-based service creation
• ✅ Configuration file management
• ✅ Confirmation dialogs
• ✅ Search and filtering`
	
	// Style each section
	navPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronCyan).Render(navigation)
	opsPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronBlue).Render(operations)
	viewsPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronPurple).Render(views)
	featuresPanel := styles.PanelStyle.Copy().BorderForeground(styles.TronPink).Render(features)
	
	// Arrange in two columns
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, navPanel, "", opsPanel)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, viewsPanel, "", featuresPanel)
	helpContent := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, "  ", rightColumn)
	
	// Footer with additional info
	footerText := styles.HelpStyle.Render(
		"💡 This Enhanced TUI performs real Docker operations - no CLI commands needed!")
	
	return lipgloss.JoinVertical(lipgloss.Left,
		styles.HeaderStyle.Render("❓ Enhanced TUI Help & Operations Guide"),
		"",
		helpContent,
		"",
		footerText,
	)
}

func (m EnhancedModel) renderEnhancedFooter() string {
	leftSection := fmt.Sprintf("Services: %d | Running: %d | Enhanced Mode: ON",
		m.App.GetTotalServices(),
		m.App.GetRunningServices())
		
	middleSection := ""
	if m.App.Loading {
		middleSection = fmt.Sprintf("%s %s", m.App.Spinner.View(), m.App.StatusMsg)
	} else if m.App.Error != "" {
		middleSection = "❌ " + m.App.Error
	} else if m.App.SuccessMsg != "" {
		middleSection = "✅ " + m.App.SuccessMsg
	} else if m.App.StatusMsg != "" {
		middleSection = m.App.StatusMsg
	} else {
		middleSection = "Ready for operations"
	}
	
	rightSection := "'?' Help | 'q' Quit"
	
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

// subscribeToLogChannels starts listening to log streaming channels
func (m EnhancedModel) subscribeToLogChannels() tea.Cmd {
	return tea.Batch(
		m.waitForLogLine(),
		m.waitForLogStreamStart(),
		m.waitForLogStreamStop(),
		m.waitForLogStreamError(),
	)
}

// waitForLogLine waits for log line messages and collects them in batches
func (m EnhancedModel) waitForLogLine() tea.Cmd {
	return func() tea.Msg {
		// Try to collect multiple messages at once
		logChan := GetLogStreamChan()
		
		// First, try to get at least one message
		select {
		case firstMsg := <-logChan:
			// Got first message, now try to collect more quickly
			collectedMsgs := []models.LogLineMsg{firstMsg}
			
			// Try to collect additional messages for up to 5ms
			timeout := time.After(5 * time.Millisecond)
			for len(collectedMsgs) < 50 { // Limit batch size
				select {
				case additionalMsg := <-logChan:
					collectedMsgs = append(collectedMsgs, additionalMsg)
				case <-timeout:
					// Timeout reached, return what we have
					goto returnBatch
				}
			}
			
			returnBatch:
			// Return the batch - the Update function will handle continuing to listen
			return LogBatchMsg{Messages: collectedMsgs}
			
		case <-time.After(10 * time.Millisecond):
			// No messages available, continue listening
			return ContinueLogListeningMsg{}
		}
	}
}

// waitForLogStreamStart waits for a log stream start message (placeholder)
func (m EnhancedModel) waitForLogStreamStart() tea.Cmd {
	return func() tea.Msg {
		// For now, just return continue message since we don't have separate start channel
		time.Sleep(50 * time.Millisecond)
		return ContinueLogStartListeningMsg{}
	}
}

// waitForLogStreamStop waits for a log stream stop message (placeholder)
func (m EnhancedModel) waitForLogStreamStop() tea.Cmd {
	return func() tea.Msg {
		// For now, just return continue message since we don't have separate stop channel
		time.Sleep(50 * time.Millisecond)
		return ContinueLogStopListeningMsg{}
	}
}

// waitForLogStreamError waits for a log stream error message
func (m EnhancedModel) waitForLogStreamError() tea.Cmd {
	return func() tea.Msg {
		select {
		case errMsg := <-GetLogStreamErrChan():
			// Type assert to the expected message types
			switch msg := errMsg.(type) {
			case models.LogStreamErrorMsg:
				return msg
			case models.LogStreamStopMsg:
				return msg
			default:
				// Handle unexpected message type
				return models.LogStreamErrorMsg{
					ServiceName: "unknown",
					Error:       "Unknown error message type",
				}
			}
		case <-time.After(50 * time.Millisecond):
			return ContinueLogErrorListeningMsg{}
		}
	}
}

// Message types for continuing to listen to channels
type ContinueLogListeningMsg struct{}
type ContinueLogStartListeningMsg struct{}
type ContinueLogStopListeningMsg struct{}
type ContinueLogErrorListeningMsg struct{}

// LogBatchMsg represents a batch of log messages
type LogBatchMsg struct {
	Messages []models.LogLineMsg
}

// wrapLogLine wraps a single log line to fit within the specified width
// It breaks long lines at word boundaries when possible, or at character boundaries if necessary
func (m EnhancedModel) wrapLogLine(line string, width int) []string {
	if len(line) <= width {
		return []string{line}
	}
	
	var wrappedLines []string
	currentLine := line
	
	for len(currentLine) > width {
		// Try to break at a word boundary within the width limit
		breakPoint := width
		
		// Look for the last space or other word boundary character within the limit
		for i := width - 1; i >= width/2; i-- {
			if currentLine[i] == ' ' || currentLine[i] == '\t' || currentLine[i] == ',' || 
			   currentLine[i] == ';' || currentLine[i] == '|' || currentLine[i] == '=' {
				breakPoint = i
				break
			}
		}
		
		// If we couldn't find a good break point, just break at the width limit
		if breakPoint == width {
			// Look for any reasonable breaking character in the first half
			for i := width/2; i < width; i++ {
				if currentLine[i] == ' ' || currentLine[i] == '\t' {
					breakPoint = i
					break
				}
			}
		}
		
		// Extract the line segment
		segment := strings.TrimRight(currentLine[:breakPoint], " \t")
		wrappedLines = append(wrappedLines, segment)
		
		// Continue with the remainder, skipping any leading whitespace
		if breakPoint < len(currentLine) {
			currentLine = strings.TrimLeft(currentLine[breakPoint:], " \t")
		} else {
			break
		}
	}
	
	// Add any remaining text
	if len(currentLine) > 0 {
		wrappedLines = append(wrappedLines, currentLine)
	}
	
	return wrappedLines
}

// RunEnhancedTUI starts the enhanced TUI application with real Docker operations
func RunEnhancedTUI(demo bool, debug bool) error {
	// Use demo mode to fall back to original TUI if requested
	if demo {
		return RunTUI()
	}
	model, err := NewEnhancedModel()
	if err != nil {
		return fmt.Errorf("failed to create enhanced model: %w", err)
	}
	
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}
