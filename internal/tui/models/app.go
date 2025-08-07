package models

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/docker"
	"github.com/abdultolba/nizam/internal/templates"
)

// ViewState represents the current view/screen
type ViewState int

const (
	DashboardView ViewState = iota
	ServicesView
	LogsView
	TemplatesView
	ConfigView
	HelpView
)

// Service represents a service with its status
type Service struct {
	Name        string
	Image       string
	Status      string
	Ports       []string
	Environment map[string]string
	Healthy     bool
	Uptime      time.Duration
	CPU         float64
	Memory      string
}

// OperationState represents the current operation being performed
type OperationState int

const (
	OperationIdle OperationState = iota
	OperationStarting
	OperationStopping
	OperationRestarting
	OperationRemoving
	OperationAdding
	OperationConfiguring
	OperationRefreshing
)

// InputMode represents different input modes
type InputMode int

const (
	InputModeNone InputMode = iota
	InputModeSearch
	InputModeConfigEdit
	InputModeAddService
	InputModeConfirm
)

// Enhanced Service with real Docker integration
type EnhancedService struct {
	Name        string
	Image       string
	Status      string
	ContainerID string
	Ports       []string
	Environment map[string]string
	Healthy     bool
	Uptime      time.Duration
	CPU         float64
	Memory      string
	Config      config.Service
	LastError   string
}

// ServiceTemplate represents available templates
type ServiceTemplate struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	Config      config.Service
	Variables   []templates.Variable
}

// AppModel represents the main application model with enhanced functionality
type AppModel struct {
	// Current view state
	CurrentView ViewState
	PrevView    ViewState

	// Core data
	Services         []Service
	EnhancedServices []EnhancedService
	Templates        []ServiceTemplate
	Config           *config.Config
	LastUpdated      time.Time

	// UI state
	Width      int
	Height     int
	Loading    bool
	Error      string
	StatusMsg  string
	SuccessMsg string
	WarningMsg string

	// Navigation and selection
	ActivePanel           int
	MaxPanels             int
	SelectedServiceIndex  int
	SelectedTemplateIndex int
	SelectedConfigKey     int
	ScrollOffset          int

	// Search and filtering
	SearchQuery    string
	ShowSearch     bool
	FilterCategory string
	FilterStatus   string

	// Real operations
	DockerClient      *docker.Client
	OperationState    OperationState
	OperatingOn       string // Service name being operated on
	OperationProgress string

	// Input handling
	TextInput     textinput.Model
	InputMode     InputMode
	InputPrompt   string
	InputValue    string
	InputCallback func(string) tea.Cmd

	// Log management
	LogLines      []string
	LogFollowing  bool
	LogFilter     string
	MaxLogLines   int
	LogCtxCancel  context.CancelFunc

	// Components
	Spinner spinner.Model

	// Flags
	ShowHelp         bool
	Debug            bool
	ConfirmOperation bool
	ConfirmMessage   string
	ConfirmCallback  func() tea.Cmd
}

// NewAppModel creates a new application model with enhanced functionality
func NewAppModel() AppModel {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFF7"))

	// Initialize text input
	textInput := textinput.New()
	textInput.Placeholder = "Type to search..."
	textInput.CharLimit = 256

	return AppModel{
		// View state
		CurrentView: DashboardView,
		PrevView:    DashboardView,

		// Data
		Services:         []Service{},
		EnhancedServices: []EnhancedService{},
		Templates:        []ServiceTemplate{},
		LastUpdated:      time.Now(),

		// UI state
		Loading:    false,
		Error:      "",
		StatusMsg:  "",
		SuccessMsg: "",
		WarningMsg: "",

		// Navigation
		ActivePanel:           0,
		MaxPanels:             4, // Start All, Stop All, Refresh, Add Service
		SelectedServiceIndex:  0,
		SelectedTemplateIndex: 0,
		SelectedConfigKey:     0,
		ScrollOffset:          0,

		// Search and filtering
		SearchQuery:    "",
		ShowSearch:     false,
		FilterCategory: "all",
		FilterStatus:   "all",

		// Operations
		OperationState:    OperationIdle,
		OperatingOn:       "",
		OperationProgress: "",

		// Input handling
		TextInput:     textInput,
		InputMode:     InputModeNone,
		InputPrompt:   "",
		InputValue:    "",
		InputCallback: nil,

		// Log management
		LogLines:      make([]string, 0),
		LogFollowing:  false,
		LogFilter:     "",
		MaxLogLines:   1000,
		LogCtxCancel:  nil,

		// Components
		Spinner: s,

		// Flags
		ShowHelp:         false,
		Debug:            false,
		ConfirmOperation: false,
		ConfirmMessage:   "",
		ConfirmCallback:  nil,
	}
}

// Messages for the TUI
type (
	// TickMsg is sent on every tick
	TickMsg time.Time

	// RefreshMsg triggers a refresh of service data
	RefreshMsg struct{}

	// ServicesUpdatedMsg is sent when services data is updated
	ServicesUpdatedMsg []Service

	// ErrorMsg is sent when an error occurs
	ErrorMsg struct {
		Error string
	}

	// StatusMsg is sent to update the status bar
	StatusMsg string

	// ViewChangeMsg is sent when the view changes
	ViewChangeMsg ViewState

	// Enhanced operation messages
	ServiceStartMsg struct {
		ServiceName string
	}

	ServiceStopMsg struct {
		ServiceName string
	}

	ServiceRestartMsg struct {
		ServiceName string
	}

	ServiceRemoveMsg struct {
		ServiceName string
	}

	ServiceAddMsg struct {
		TemplateName string
		ServiceName  string
		Variables    map[string]string
	}

	// Operation completed messages
	OperationCompleteMsg struct {
		Operation string
		Service   string
		Success   bool
		Error     string
	}

	// Config update messages
	ConfigUpdateMsg struct {
		Config *config.Config
	}

	ConfigSavedMsg struct{}

	// Log streaming messages
	LogLineMsg struct {
		ServiceName string
		Line        string
	}

	LogStreamStartMsg struct {
		ServiceName string
	}

	LogStreamStopMsg struct {
		ServiceName string
	}

	// Input handling messages
	InputSubmitMsg struct {
		Value string
	}

	InputCancelMsg struct{}

	// Search and filter messages
	SearchUpdateMsg struct {
		Query string
	}

	FilterUpdateMsg struct {
		Category string
		Status   string
	}

	// Template loading
	TemplatesLoadedMsg []ServiceTemplate

	// Real service status update
	RealServiceStatusMsg []EnhancedService
)

// Init initializes the model
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
}

// NavigateToView changes the current view
func (m *AppModel) NavigateToView(view ViewState) {
	m.PrevView = m.CurrentView
	m.CurrentView = view
}

// NextPanel moves to the next panel
func (m *AppModel) NextPanel() {
	m.ActivePanel = (m.ActivePanel + 1) % m.MaxPanels
}

// PrevPanel moves to the previous panel
func (m *AppModel) PrevPanel() {
	m.ActivePanel = (m.ActivePanel - 1 + m.MaxPanels) % m.MaxPanels
}

// SetLoading sets the loading state
func (m *AppModel) SetLoading(loading bool) {
	m.Loading = loading
}

// SetError sets an error message
func (m *AppModel) SetError(err string) {
	m.Error = err
	m.Loading = false
}

// ClearError clears the error message
func (m *AppModel) ClearError() {
	m.Error = ""
}

// SetStatus sets a status message
func (m *AppModel) SetStatus(status string) {
	m.StatusMsg = status
}

// UpdateServices updates the services list
func (m *AppModel) UpdateServices(services []Service) {
	m.Services = services
	m.LastUpdated = time.Now()
}

// GetRunningServices returns the count of running services
func (m *AppModel) GetRunningServices() int {
	count := 0
	for _, service := range m.Services {
		if service.Status == "running" {
			count++
		}
	}
	return count
}

// GetTotalServices returns the total count of services
func (m *AppModel) GetTotalServices() int {
	return len(m.Services)
}

// GetHealthyServices returns the count of healthy services
func (m *AppModel) GetHealthyServices() int {
	count := 0
	for _, service := range m.Services {
		if service.Healthy {
			count++
		}
	}
	return count
}

// Enhanced methods for full functionality

// InitializeDockerClient initializes the Docker client
func (m *AppModel) InitializeDockerClient() error {
	client, err := docker.NewClient()
	if err != nil {
		return err
	}
	m.DockerClient = client
	return nil
}

// LoadConfiguration loads the configuration from file
func (m *AppModel) LoadConfiguration() error {
	config, err := config.LoadConfig()
	if err != nil {
		return err
	}
	m.Config = config
	return nil
}

// LoadTemplates loads available service templates
func (m *AppModel) LoadTemplates() {
	allTemplates := templates.GetAllTemplates()
	m.Templates = make([]ServiceTemplate, 0, len(allTemplates))
	
	for name, template := range allTemplates {
		category := "Other"
		if len(template.Tags) > 0 {
			category = template.Tags[0]
		}
		
		m.Templates = append(m.Templates, ServiceTemplate{
			Name:        name,
			Description: template.Description,
			Category:    category,
			Tags:        template.Tags,
			Config:      template.Service,
			Variables:   template.Variables,
		})
	}
}

// SetInputMode sets the input mode and configures the text input
func (m *AppModel) SetInputMode(mode InputMode, prompt string, callback func(string) tea.Cmd) {
	m.InputMode = mode
	m.InputPrompt = prompt
	m.InputCallback = callback
	m.TextInput.SetValue("")
	m.TextInput.Focus()
}

// ClearInputMode clears the input mode
func (m *AppModel) ClearInputMode() {
	m.InputMode = InputModeNone
	m.TextInput.Blur()
	m.InputCallback = nil
}

// StartOperation starts an operation on a service
func (m *AppModel) StartOperation(op OperationState, serviceName string) {
	m.OperationState = op
	m.OperatingOn = serviceName
	m.SetLoading(true)
}

// CompleteOperation completes the current operation
func (m *AppModel) CompleteOperation(success bool, message string) {
	m.OperationState = OperationIdle
	m.OperatingOn = ""
	m.SetLoading(false)
	
	if success {
		m.SuccessMsg = message
		m.Error = ""
	} else {
		m.Error = message
		m.SuccessMsg = ""
	}
}

// GetFilteredServices returns services filtered by search and status
func (m *AppModel) GetFilteredServices() []Service {
	filtered := make([]Service, 0)
	
	for _, service := range m.Services {
		// Apply search filter
		if m.SearchQuery != "" {
			if !contains(service.Name, m.SearchQuery) && !contains(service.Image, m.SearchQuery) {
				continue
			}
		}
		
		// Apply status filter
		if m.FilterStatus != "" && m.FilterStatus != "all" {
			if service.Status != m.FilterStatus {
				continue
			}
		}
		
		filtered = append(filtered, service)
	}
	
	return filtered
}

// GetFilteredTemplates returns templates filtered by search and category
func (m *AppModel) GetFilteredTemplates() []ServiceTemplate {
	filtered := make([]ServiceTemplate, 0)
	
	for _, template := range m.Templates {
		// Apply search filter
		if m.SearchQuery != "" {
			if !contains(template.Name, m.SearchQuery) && !contains(template.Description, m.SearchQuery) {
				continue
			}
		}
		
		// Apply category filter
		if m.FilterCategory != "" && m.FilterCategory != "all" {
			if template.Category != m.FilterCategory {
				continue
			}
		}
		
		filtered = append(filtered, template)
	}
	
	return filtered
}

// AddLogLine adds a line to the log buffer
func (m *AppModel) AddLogLine(serviceName, line string) {
	timestamp := time.Now().Format("15:04:05")
	logLine := fmt.Sprintf("[%s] %s: %s", timestamp, serviceName, line)
	
	m.LogLines = append(m.LogLines, logLine)
	
	// Limit log buffer size
	if len(m.LogLines) > m.MaxLogLines {
		m.LogLines = m.LogLines[1:]
	}
}

// ClearLogs clears the log buffer
func (m *AppModel) ClearLogs() {
	m.LogLines = make([]string, 0)
}

// GetFilteredLogs returns logs filtered by the current filter
func (m *AppModel) GetFilteredLogs() []string {
	if m.LogFilter == "" {
		return m.LogLines
	}
	
	filtered := make([]string, 0)
	for _, line := range m.LogLines {
		if contains(line, m.LogFilter) {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

// ShowConfirmDialog shows a confirmation dialog
func (m *AppModel) ShowConfirmDialog(message string, callback func() tea.Cmd) {
	m.ConfirmOperation = true
	m.ConfirmMessage = message
	m.ConfirmCallback = callback
}

// HideConfirmDialog hides the confirmation dialog
func (m *AppModel) HideConfirmDialog() {
	m.ConfirmOperation = false
	m.ConfirmMessage = ""
	m.ConfirmCallback = nil
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && s[0:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		(len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
