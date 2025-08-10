package operations

import (
	"fmt"
	"os/exec"
	"strings"
)

// Start starts the specified services or all services if none specified
func Start(serviceNames []string) error {
	args := []string{"compose", "up", "-d"}
	if len(serviceNames) > 0 {
		args = append(args, serviceNames...)
	}
	
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// Stop stops the specified services or all services if none specified
func Stop(serviceNames []string) error {
	args := []string{"compose", "stop"}
	if len(serviceNames) > 0 {
		args = append(args, serviceNames...)
	}
	
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose stop failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// Restart restarts the specified services or all services if none specified
func Restart(serviceNames []string) error {
	args := []string{"compose", "restart"}
	if len(serviceNames) > 0 {
		args = append(args, serviceNames...)
	}
	
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose restart failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// Pull pulls images for the specified services or all services if none specified
func Pull(serviceNames []string) error {
	args := []string{"compose", "pull"}
	if len(serviceNames) > 0 {
		args = append(args, serviceNames...)
	}
	
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose pull failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// Build builds images for the specified services or all services if none specified
func Build(serviceNames []string) error {
	args := []string{"compose", "build"}
	if len(serviceNames) > 0 {
		args = append(args, serviceNames...)
	}
	
	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose build failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// GetServiceStatus returns the status of all services
func GetServiceStatus() (map[string]string, error) {
	cmd := exec.Command("docker", "compose", "ps", "--format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker compose ps failed: %w\nOutput: %s", err, output)
	}
	
	// Simple parsing - for a more robust implementation, use JSON parsing
	status := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Name") && strings.Contains(line, "State") {
			// This is a simplified parser - in production, use proper JSON unmarshaling
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				status[parts[0]] = parts[1]
			}
		}
	}
	
	return status, nil
}
