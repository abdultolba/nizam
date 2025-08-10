package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/abdultolba/nizam/internal/config"
	"github.com/abdultolba/nizam/internal/docker"
	"github.com/rs/zerolog/log"
)

// HealthStatus represents the health status of a service
type HealthStatus string

const (
	HealthStatusHealthy    HealthStatus = "healthy"
	HealthStatusUnhealthy  HealthStatus = "unhealthy"
	HealthStatusStarting   HealthStatus = "starting"
	HealthStatusUnknown    HealthStatus = "unknown"
	HealthStatusNotRunning HealthStatus = "not_running"
)

// HealthCheckType represents different types of health checks
type HealthCheckType string

const (
	HealthCheckTypeCommand HealthCheckType = "command"
	HealthCheckTypeHTTP    HealthCheckType = "http"
	HealthCheckTypeTCP     HealthCheckType = "tcp"
	HealthCheckTypeDocker  HealthCheckType = "docker"
)

// HealthCheckResult contains the result of a health check
type HealthCheckResult struct {
	ServiceName string          `json:"service_name"`
	Status      HealthStatus    `json:"status"`
	Message     string          `json:"message"`
	CheckType   HealthCheckType `json:"check_type"`
	Duration    time.Duration   `json:"duration"`
	Timestamp   time.Time       `json:"timestamp"`
	Error       error           `json:"error,omitempty"`
	Details     interface{}     `json:"details,omitempty"`
}

// ServiceHealthInfo contains comprehensive health information for a service
type ServiceHealthInfo struct {
	ServiceName   string              `json:"service_name"`
	ContainerID   string              `json:"container_id,omitempty"`
	ContainerName string              `json:"container_name,omitempty"`
	Image         string              `json:"image,omitempty"`
	Status        HealthStatus        `json:"status"`
	LastCheck     time.Time           `json:"last_check"`
	Uptime        time.Duration       `json:"uptime,omitempty"`
	CheckHistory  []HealthCheckResult `json:"check_history,omitempty"`
	Configuration *config.HealthCheck `json:"configuration,omitempty"`
	IsRunning     bool                `json:"is_running"`
}

// Engine manages health checks for all services
type Engine struct {
	dockerClient *docker.Client
	services     map[string]*ServiceHealthInfo
	mutex        sync.RWMutex
	ticker       *time.Ticker
	stopChan     chan struct{}
	config       *config.Config
}

// NewEngine creates a new health check engine
func NewEngine(dockerClient *docker.Client, cfg *config.Config) (*Engine, error) {
	if dockerClient == nil {
		return nil, fmt.Errorf("docker client is required")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	return &Engine{
		dockerClient: dockerClient,
		services:     make(map[string]*ServiceHealthInfo),
		stopChan:     make(chan struct{}),
		config:       cfg,
	}, nil
}

// Start begins the health checking loop
func (e *Engine) Start(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 30 * time.Second // Default 30 seconds
	}

	e.ticker = time.NewTicker(interval)
	log.Info().Dur("interval", interval).Msg("Health check engine started")

	go e.healthCheckLoop(ctx)
}

// Stop stops the health check engine
func (e *Engine) Stop() {
	if e.ticker != nil {
		e.ticker.Stop()
	}
	close(e.stopChan)
	log.Info().Msg("Health check engine stopped")
}

// healthCheckLoop runs the continuous health checking
func (e *Engine) healthCheckLoop(ctx context.Context) {
	// Initial check
	e.checkAllServices(ctx)

	for {
		select {
		case <-e.ticker.C:
			e.checkAllServices(ctx)
		case <-e.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkAllServices performs health checks on all configured services
func (e *Engine) checkAllServices(ctx context.Context) {
	// Get current container status
	containers, err := e.dockerClient.GetServiceStatus(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get service status for health checks")
		return
	}

	// Check each configured service
	for serviceName, serviceConfig := range e.config.Services {
		e.checkService(ctx, serviceName, serviceConfig, containers)
	}
}

// checkService performs a health check on a specific service
func (e *Engine) checkService(ctx context.Context, serviceName string, serviceConfig config.Service, containers []docker.ContainerInfo) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Find container for this service
	var containerInfo *docker.ContainerInfo
	for _, container := range containers {
		if container.Service == serviceName {
			containerInfo = &container
			break
		}
	}

	// Get or create service health info
	healthInfo, exists := e.services[serviceName]
	if !exists {
		healthInfo = &ServiceHealthInfo{
			ServiceName:   serviceName,
			Status:        HealthStatusUnknown,
			Configuration: serviceConfig.HealthCheck,
			CheckHistory:  make([]HealthCheckResult, 0, 10), // Keep last 10 checks
		}
		e.services[serviceName] = healthInfo
	}

	// Update container information
	if containerInfo != nil {
		healthInfo.ContainerID = containerInfo.ID
		healthInfo.ContainerName = containerInfo.Name
		healthInfo.Image = containerInfo.Image
		healthInfo.IsRunning = strings.Contains(strings.ToLower(containerInfo.Status), "up")
	} else {
		healthInfo.IsRunning = false
	}

	// If container is not running, mark as not running
	if !healthInfo.IsRunning {
		result := HealthCheckResult{
			ServiceName: serviceName,
			Status:      HealthStatusNotRunning,
			Message:     "Container is not running",
			CheckType:   HealthCheckTypeDocker,
			Duration:    0,
			Timestamp:   time.Now(),
		}
		healthInfo.Status = HealthStatusNotRunning
		healthInfo.LastCheck = result.Timestamp
		e.addCheckResult(healthInfo, result)
		return
	}

	// Perform health check based on configuration
	var result HealthCheckResult
	if serviceConfig.HealthCheck != nil && len(serviceConfig.HealthCheck.Test) > 0 {
		result = e.performConfiguredHealthCheck(ctx, serviceName, serviceConfig.HealthCheck, containerInfo)
	} else {
		result = e.performDefaultHealthCheck(ctx, serviceName, containerInfo)
	}

	// Update health info
	healthInfo.Status = result.Status
	healthInfo.LastCheck = result.Timestamp
	e.addCheckResult(healthInfo, result)
}

// performConfiguredHealthCheck runs a configured health check
func (e *Engine) performConfiguredHealthCheck(ctx context.Context, serviceName string, healthCheck *config.HealthCheck, containerInfo *docker.ContainerInfo) HealthCheckResult {
	start := time.Now()

	// Parse timeout
	timeout := 10 * time.Second // default
	if healthCheck.Timeout != "" {
		if parsedTimeout, err := time.ParseDuration(healthCheck.Timeout); err == nil {
			timeout = parsedTimeout
		}
	}

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := HealthCheckResult{
		ServiceName: serviceName,
		CheckType:   HealthCheckTypeCommand,
		Timestamp:   time.Now(),
	}

	// Determine check type based on test command
	if len(healthCheck.Test) == 0 {
		result.Status = HealthStatusUnknown
		result.Message = "No health check test configured"
		result.Duration = time.Since(start)
		return result
	}

	firstArg := strings.ToLower(healthCheck.Test[0])
	switch firstArg {
	case "cmd", "cmd-shell":
		result = e.performCommandHealthCheck(checkCtx, serviceName, healthCheck.Test[1:], containerInfo)
	case "curl", "wget":
		result = e.performHTTPHealthCheck(checkCtx, serviceName, healthCheck.Test, containerInfo)
	default:
		// Default to command execution
		result = e.performCommandHealthCheck(checkCtx, serviceName, healthCheck.Test, containerInfo)
	}

	result.Duration = time.Since(start)
	return result
}

// performCommandHealthCheck executes a command-based health check
func (e *Engine) performCommandHealthCheck(ctx context.Context, serviceName string, command []string, containerInfo *docker.ContainerInfo) HealthCheckResult {
	result := HealthCheckResult{
		ServiceName: serviceName,
		CheckType:   HealthCheckTypeCommand,
		Timestamp:   time.Now(),
	}

	if len(command) == 0 {
		result.Status = HealthStatusUnknown
		result.Message = "Empty command"
		return result
	}

	// If we have container info, execute command inside the container
	if containerInfo != nil {
		// For now, we'll use docker exec to run the health check
		// This could be enhanced to use the Docker API directly
		dockerCmd := []string{"docker", "exec", containerInfo.Name}
		dockerCmd = append(dockerCmd, command...)

		cmd := exec.CommandContext(ctx, dockerCmd[0], dockerCmd[1:]...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			result.Status = HealthStatusUnhealthy
			result.Message = fmt.Sprintf("Command failed: %v", err)
			result.Error = err
			result.Details = map[string]interface{}{
				"command": command,
				"output":  string(output),
			}
		} else {
			result.Status = HealthStatusHealthy
			result.Message = "Command executed successfully"
			result.Details = map[string]interface{}{
				"command": command,
				"output":  string(output),
			}
		}
	} else {
		// Execute command on host (fallback)
		cmd := exec.CommandContext(ctx, command[0], command[1:]...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			result.Status = HealthStatusUnhealthy
			result.Message = fmt.Sprintf("Host command failed: %v", err)
			result.Error = err
		} else {
			result.Status = HealthStatusHealthy
			result.Message = "Host command executed successfully"
		}

		result.Details = map[string]interface{}{
			"command": command,
			"output":  string(output),
		}
	}

	return result
}

// performHTTPHealthCheck performs an HTTP-based health check
func (e *Engine) performHTTPHealthCheck(ctx context.Context, serviceName string, command []string, containerInfo *docker.ContainerInfo) HealthCheckResult {
	result := HealthCheckResult{
		ServiceName: serviceName,
		CheckType:   HealthCheckTypeHTTP,
		Timestamp:   time.Now(),
	}

	if len(command) < 2 {
		result.Status = HealthStatusUnknown
		result.Message = "Invalid HTTP command format"
		return result
	}

	url := command[1]

	// Create HTTP client with timeout from context
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Failed to create HTTP request: %v", err)
		result.Error = err
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("HTTP request failed: %v", err)
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Status = HealthStatusHealthy
		result.Message = fmt.Sprintf("HTTP check successful (status: %d)", resp.StatusCode)
	} else {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("HTTP check failed (status: %d)", resp.StatusCode)
	}

	result.Details = map[string]interface{}{
		"url":         url,
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
	}

	return result
}

// performDefaultHealthCheck performs a default health check when none is configured
func (e *Engine) performDefaultHealthCheck(ctx context.Context, serviceName string, containerInfo *docker.ContainerInfo) HealthCheckResult {
	result := HealthCheckResult{
		ServiceName: serviceName,
		CheckType:   HealthCheckTypeDocker,
		Timestamp:   time.Now(),
	}

	if containerInfo == nil {
		result.Status = HealthStatusNotRunning
		result.Message = "Container not found"
		return result
	}

	// Check if container is running
	if strings.Contains(strings.ToLower(containerInfo.Status), "up") {
		result.Status = HealthStatusHealthy
		result.Message = fmt.Sprintf("Container is running (%s)", containerInfo.Status)
	} else {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("Container status: %s", containerInfo.Status)
	}

	result.Details = map[string]interface{}{
		"container_id":     containerInfo.ID,
		"container_name":   containerInfo.Name,
		"container_status": containerInfo.Status,
		"image":            containerInfo.Image,
		"ports":            containerInfo.Ports,
	}

	return result
}

// addCheckResult adds a health check result to the service history
func (e *Engine) addCheckResult(healthInfo *ServiceHealthInfo, result HealthCheckResult) {
	healthInfo.CheckHistory = append(healthInfo.CheckHistory, result)

	// Keep only the last 10 results
	if len(healthInfo.CheckHistory) > 10 {
		healthInfo.CheckHistory = healthInfo.CheckHistory[1:]
	}
}

// GetServiceHealth returns the health information for a specific service
func (e *Engine) GetServiceHealth(serviceName string) (*ServiceHealthInfo, bool) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	info, exists := e.services[serviceName]
	if !exists {
		return nil, false
	}

	// Create a copy to avoid race conditions
	infoCopy := *info
	infoCopy.CheckHistory = make([]HealthCheckResult, len(info.CheckHistory))
	copy(infoCopy.CheckHistory, info.CheckHistory)

	return &infoCopy, true
}

// GetAllServicesHealth returns health information for all services
func (e *Engine) GetAllServicesHealth() map[string]*ServiceHealthInfo {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	result := make(map[string]*ServiceHealthInfo)
	for name, info := range e.services {
		// Create a copy to avoid race conditions
		infoCopy := *info
		infoCopy.CheckHistory = make([]HealthCheckResult, len(info.CheckHistory))
		copy(infoCopy.CheckHistory, info.CheckHistory)
		result[name] = &infoCopy
	}

	return result
}

// CheckServiceNow performs an immediate health check on a specific service
func (e *Engine) CheckServiceNow(ctx context.Context, serviceName string) (*HealthCheckResult, error) {
	serviceConfig, exists := e.config.Services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found in configuration", serviceName)
	}

	// Get container info
	containers, err := e.dockerClient.GetServiceStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	var containerInfo *docker.ContainerInfo
	for _, container := range containers {
		if container.Service == serviceName {
			containerInfo = &container
			break
		}
	}

	// Perform check
	var result HealthCheckResult
	if serviceConfig.HealthCheck != nil && len(serviceConfig.HealthCheck.Test) > 0 {
		result = e.performConfiguredHealthCheck(ctx, serviceName, serviceConfig.HealthCheck, containerInfo)
	} else {
		result = e.performDefaultHealthCheck(ctx, serviceName, containerInfo)
	}

	// Update stored health info
	e.mutex.Lock()
	healthInfo, exists := e.services[serviceName]
	if !exists {
		healthInfo = &ServiceHealthInfo{
			ServiceName:   serviceName,
			Status:        result.Status,
			Configuration: serviceConfig.HealthCheck,
			CheckHistory:  make([]HealthCheckResult, 0, 10),
		}
		e.services[serviceName] = healthInfo
	}

	healthInfo.Status = result.Status
	healthInfo.LastCheck = result.Timestamp
	if containerInfo != nil {
		healthInfo.ContainerID = containerInfo.ID
		healthInfo.ContainerName = containerInfo.Name
		healthInfo.Image = containerInfo.Image
		healthInfo.IsRunning = strings.Contains(strings.ToLower(containerInfo.Status), "up")
	}

	e.addCheckResult(healthInfo, result)
	e.mutex.Unlock()

	return &result, nil
}

// GetHealthSummary returns a summary of health status across all services
func (e *Engine) GetHealthSummary() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	summary := map[string]interface{}{
		"total_services": len(e.services),
		"healthy":        0,
		"unhealthy":      0,
		"starting":       0,
		"not_running":    0,
		"unknown":        0,
		"last_updated":   time.Now(),
	}

	for _, info := range e.services {
		switch info.Status {
		case HealthStatusHealthy:
			summary["healthy"] = summary["healthy"].(int) + 1
		case HealthStatusUnhealthy:
			summary["unhealthy"] = summary["unhealthy"].(int) + 1
		case HealthStatusStarting:
			summary["starting"] = summary["starting"].(int) + 1
		case HealthStatusNotRunning:
			summary["not_running"] = summary["not_running"].(int) + 1
		default:
			summary["unknown"] = summary["unknown"].(int) + 1
		}
	}

	return summary
}

// GetDockerClient returns the Docker client
func (e *Engine) GetDockerClient() *docker.Client {
	return e.dockerClient
}

// GetConfig returns the configuration
func (e *Engine) GetConfig() *config.Config {
	return e.config
}
