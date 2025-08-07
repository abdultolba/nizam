package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/docker"
	"github.com/abdultolba/nizam/internal/templates"
	"github.com/abdultolba/nizam/internal/tui/models"
	"gopkg.in/yaml.v3"
)

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

// StreamLogs streams logs from a service
func (ops *ServiceOperations) StreamLogs(serviceName string, follow bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		
		logs, err := ops.DockerClient.GetServiceLogs(ctx, serviceName, follow, "100")
		if err != nil {
			return models.ErrorMsg{Error: fmt.Sprintf("Failed to get logs: %v", err)}
		}
		defer logs.Close()

		// Read logs and send as messages
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := logs.Read(buf)
				if err != nil {
					if err != io.EOF {
						// Send error
						// Note: This would need a channel or other mechanism to send back to TUI
					}
					break
				}
				
				line := string(buf[:n])
				// Clean up Docker log formatting
				if len(line) > 8 {
					line = line[8:] // Remove Docker log prefix
				}
				line = strings.TrimSpace(line)
				
				if line != "" {
					// Send log line - would need proper message passing mechanism
					// This is a simplified version
				}
			}
		}()

		return models.LogStreamStartMsg{ServiceName: serviceName}
	}
}

// Helper functions

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
