package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/docker"
	"github.com/abdultolba/nizam/internal/templates"
	"github.com/abdultolba/nizam/internal/tui/models"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

// Global channels for log streaming communication
var (
	logStreamChan    chan models.LogLineMsg
	logStreamErrChan chan interface{}
)

// InitLogChannels initializes the log streaming channels
func InitLogChannels() {
	logStreamChan = make(chan models.LogLineMsg, 100)
	logStreamErrChan = make(chan interface{}, 10)
}

// GetLogStreamChan returns the log stream channel
func GetLogStreamChan() chan models.LogLineMsg {
	return logStreamChan
}

// GetLogStreamErrChan returns the log error channel
func GetLogStreamErrChan() chan interface{} {
	return logStreamErrChan
}

// ServiceOperations handles real Docker service operations
type ServiceOperations struct {
	DockerClient *docker.Client
}

// NewServiceOperations creates a new service operations manager
func NewServiceOperations() (*ServiceOperations, error) {
	client, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	
	return &ServiceOperations{
		DockerClient: client,
	}, nil
}

// StartService starts a service using the real Docker client
func (ops *ServiceOperations) StartService(serviceName string, config *config.Config) tea.Cmd {
	return func() tea.Msg {
		serviceConfig, exists := config.Services[serviceName]
		if !exists {
			return models.OperationCompleteMsg{
				Operation: "start",
				Service:   serviceName,
				Success:   false,
				Error:     fmt.Sprintf("Service '%s' not found in configuration", serviceName),
			}
		}

		ctx := context.Background()
		err := ops.DockerClient.StartService(ctx, serviceName, serviceConfig)
		if err != nil {
			return models.OperationCompleteMsg{
				Operation: "start",
				Service:   serviceName,
				Success:   false,
				Error:     err.Error(),
			}
		}

		return models.OperationCompleteMsg{
			Operation: "start",
			Service:   serviceName,
			Success:   true,
			Error:     "",
		}
	}
}

// StopService stops a service using the real Docker client
func (ops *ServiceOperations) StopService(serviceName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := ops.DockerClient.StopService(ctx, serviceName)
		if err != nil {
			return models.OperationCompleteMsg{
				Operation: "stop",
				Service:   serviceName,
				Success:   false,
				Error:     err.Error(),
			}
		}

		return models.OperationCompleteMsg{
			Operation: "stop",
			Service:   serviceName,
			Success:   true,
			Error:     "",
		}
	}
}

// RestartService restarts a service
func (ops *ServiceOperations) RestartService(serviceName string, config *config.Config) tea.Cmd {
	return func() tea.Msg {
		// First stop the service
		ctx := context.Background()
		err := ops.DockerClient.StopService(ctx, serviceName)
		if err != nil {
			return models.OperationCompleteMsg{
				Operation: "restart",
				Service:   serviceName,
				Success:   false,
				Error:     fmt.Sprintf("Failed to stop service: %v", err),
			}
		}

		// Wait a moment for graceful shutdown
		time.Sleep(2 * time.Second)

		// Then start it again
		serviceConfig, exists := config.Services[serviceName]
		if !exists {
			return models.OperationCompleteMsg{
				Operation: "restart",
				Service:   serviceName,
				Success:   false,
				Error:     fmt.Sprintf("Service '%s' not found in configuration", serviceName),
			}
		}

		err = ops.DockerClient.StartService(ctx, serviceName, serviceConfig)
		if err != nil {
			return models.OperationCompleteMsg{
				Operation: "restart",
				Service:   serviceName,
				Success:   false,
				Error:     fmt.Sprintf("Failed to start service: %v", err),
			}
		}

		return models.OperationCompleteMsg{
			Operation: "restart",
			Service:   serviceName,
			Success:   true,
			Error:     "",
		}
	}
}

// RemoveService removes a service from both Docker and configuration
func (ops *ServiceOperations) RemoveService(serviceName string, configPath string) tea.Cmd {
	return func() tea.Msg {
		// First stop and remove from Docker
		ctx := context.Background()
		err := ops.DockerClient.StopService(ctx, serviceName)
		if err != nil {
			// Don't fail if service wasn't running
			if !strings.Contains(err.Error(), "No such container") {
				return models.OperationCompleteMsg{
					Operation: "remove",
					Service:   serviceName,
					Success:   false,
					Error:     fmt.Sprintf("Failed to stop service: %v", err),
				}
			}
		}

		// Remove from configuration
		err = removeServiceFromConfig(serviceName, configPath)
		if err != nil {
			return models.OperationCompleteMsg{
				Operation: "remove",
				Service:   serviceName,
				Success:   false,
				Error:     fmt.Sprintf("Failed to remove from config: %v", err),
			}
		}

		return models.OperationCompleteMsg{
			Operation: "remove",
			Service:   serviceName,
			Success:   true,
			Error:     "",
		}
	}
}

// AddService adds a new service from a template
func (ops *ServiceOperations) AddService(templateName, serviceName string, variables map[string]string, configPath string) tea.Cmd {
	return func() tea.Msg {
		// Get template
		allTemplates := templates.GetAllTemplates()
		template, exists := allTemplates[templateName]
		if !exists {
			return models.OperationCompleteMsg{
				Operation: "add",
				Service:   serviceName,
				Success:   false,
				Error:     fmt.Sprintf("Template '%s' not found", templateName),
			}
		}

		// Process template variables if any
		var serviceConfig config.Service
		var err error
		if len(template.Variables) > 0 && len(variables) > 0 {
			serviceConfig, err = templates.ProcessTemplateWithVariables(template, serviceName)
			if err != nil {
				return models.OperationCompleteMsg{
					Operation: "add",
					Service:   serviceName,
					Success:   false,
					Error:     fmt.Sprintf("Failed to process template: %v", err),
				}
			}
		} else {
			serviceConfig = template.Service
		}

		// Add to configuration file
		err = addServiceToConfig(serviceName, serviceConfig, configPath)
		if err != nil {
			return models.OperationCompleteMsg{
				Operation: "add",
				Service:   serviceName,
				Success:   false,
				Error:     fmt.Sprintf("Failed to add to config: %v", err),
			}
		}

		return models.OperationCompleteMsg{
			Operation: "add",
			Service:   serviceName,
			Success:   true,
			Error:     "",
		}
	}
}

// RefreshServices refreshes the service status from Docker
func (ops *ServiceOperations) RefreshServices() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		
		// Load current configuration
		config, err := config.LoadConfig()
		if err != nil {
			return models.ErrorMsg{Error: fmt.Sprintf("Failed to load config: %v", err)}
		}

		// Get Docker container status
		containers, err := ops.DockerClient.GetServiceStatus(ctx)
		if err != nil {
			return models.ErrorMsg{Error: fmt.Sprintf("Failed to get Docker status: %v", err)}
		}

		// Convert to enhanced services
		enhancedServices := make([]models.EnhancedService, 0)
		
		// Process configured services
		for serviceName, serviceConfig := range config.Services {
			enhancedService := models.EnhancedService{
				Name:        serviceName,
				Image:       serviceConfig.Image,
				Status:      "stopped",
				ContainerID: "",
				Ports:       serviceConfig.Ports,
				Environment: serviceConfig.Environment,
				Healthy:     false,
				Uptime:      0,
				CPU:         0.0,
				Memory:      "0MB",
				Config:      serviceConfig,
				LastError:   "",
			}

			// Find matching Docker container
			for _, container := range containers {
				if container.Service == serviceName {
					enhancedService.ContainerID = container.ID
					enhancedService.Status = parseDockerStatus(container.Status)
					enhancedService.Healthy = !strings.Contains(container.Status, "Exited")
					if len(container.Ports) > 0 {
						enhancedService.Ports = container.Ports
					}
					break
				}
			}

			enhancedServices = append(enhancedServices, enhancedService)
		}

		return models.RealServiceStatusMsg(enhancedServices)
	}
}

// StreamLogs streams logs from a service using proper async handling
func (ops *ServiceOperations) StreamLogs(serviceName string, follow bool) tea.Cmd {
	return func() tea.Msg {
		// Start streaming in a goroutine and return the start message
		go ops.streamLogsAsync(serviceName, follow)
		
		return models.LogStreamStartMsg{ServiceName: serviceName}
	}
}

// streamLogsAsync handles the actual log streaming in a separate goroutine
func (ops *ServiceOperations) streamLogsAsync(serviceName string, follow bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	logs, err := ops.DockerClient.GetServiceLogs(ctx, serviceName, follow, "50")
	if err != nil {
		// Send error message back to TUI
		logStreamErrChan <- models.LogStreamErrorMsg{
			ServiceName: serviceName,
			Error:       fmt.Sprintf("Failed to get logs: %v", err),
		}
		return
	}
	defer logs.Close()

	// Read logs line by line
	buf := make([]byte, 4096)
	var leftover []byte

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := logs.Read(buf)
			if err != nil {
				if err == io.EOF {
					// End of stream
					logStreamErrChan <- models.LogStreamStopMsg{ServiceName: serviceName}
				} else {
					// Error reading
					logStreamErrChan <- models.LogStreamErrorMsg{
						ServiceName: serviceName,
						Error:       fmt.Sprintf("Error reading logs: %v", err),
					}
				}
				return
			}

			// Combine with leftover data
			data := append(leftover, buf[:n]...)
			leftover = nil

			// Process lines
			lines := strings.Split(string(data), "\n")
			
			// Keep the last incomplete line as leftover
			if len(lines) > 0 && !strings.HasSuffix(string(data), "\n") {
				leftover = []byte(lines[len(lines)-1])
				lines = lines[:len(lines)-1]
			}

			for _, line := range lines {
				cleanLine := cleanDockerLogLine(line)
				if cleanLine != "" {
					// Send log line to TUI
					select {
					case logStreamChan <- models.LogLineMsg{
						ServiceName: serviceName,
						Line:        cleanLine,
						Timestamp:   time.Now(),
					}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}

// cleanDockerLogLine removes Docker log prefixes and cleans up the line
func cleanDockerLogLine(line string) string {
	// Docker log format often has 8-byte header we need to skip
	if len(line) > 8 {
		// Check if it starts with Docker log prefix (usually has control characters)
		if line[0] <= 2 { // Docker uses 0, 1, 2 for stdout, stdin, stderr
			line = line[8:]
		}
	}
	
	// Clean up the line
	line = strings.TrimSpace(line)
	line = strings.ReplaceAll(line, "\r", "")
	line = strings.ReplaceAll(line, "\x00", "")
	
	return line
}

// Helper functions

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseDockerStatus converts Docker status to simplified status
func parseDockerStatus(dockerStatus string) string {
	dockerStatus = strings.ToLower(dockerStatus)
	if strings.Contains(dockerStatus, "up") && !strings.Contains(dockerStatus, "unhealthy") {
		return "running"
	} else if strings.Contains(dockerStatus, "restarting") {
		return "restarting"
	} else if strings.Contains(dockerStatus, "exited") {
		return "stopped"
	} else if strings.Contains(dockerStatus, "created") {
		return "created"
	} else {
		return "unknown"
	}
}

// addServiceToConfig adds a service to the configuration file
func addServiceToConfig(serviceName string, serviceConfig config.Service, configPath string) error {
	// Read current config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Add the new service
	if cfg.Services == nil {
		cfg.Services = make(map[string]config.Service)
	}
	cfg.Services[serviceName] = serviceConfig

	// Write back to file
	updatedData, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// removeServiceFromConfig removes a service from the configuration file
func removeServiceFromConfig(serviceName string, configPath string) error {
	// Read current config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Remove the service
	if cfg.Services != nil {
		delete(cfg.Services, serviceName)
	}

	// Write back to file
	updatedData, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// updateConfigValue updates a specific configuration value
func updateConfigValue(configPath string, key string, value interface{}) error {
	// Read current config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configData map[string]interface{}
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Update the value using key path (e.g., "services.postgres.image")
	parts := strings.Split(key, ".")
	current := configData
	
	for i, part := range parts[:len(parts)-1] {
		if current[part] == nil {
			current[part] = make(map[string]interface{})
		}
		if nested, ok := current[part].(map[string]interface{}); ok {
			current = nested
		} else {
			return fmt.Errorf("invalid key path: %s at part %d", key, i)
		}
	}
	
	current[parts[len(parts)-1]] = value

	// Write back to file
	updatedData, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	configPaths := []string{".nizam.yaml", ".nizam.yml"}
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			abs, _ := filepath.Abs(path)
			return abs
		}
	}
	return ".nizam.yaml" // default
}
