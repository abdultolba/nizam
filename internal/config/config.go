package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Profile  string             `yaml:"profile" mapstructure:"profile"`
	Services map[string]Service `yaml:"services" mapstructure:"services"`
}

// Service represents a single service configuration
type Service struct {
	Image       string            `yaml:"image" mapstructure:"image"`
	Ports       []string          `yaml:"ports" mapstructure:"ports"`
	Environment map[string]string `yaml:"env" mapstructure:"env"`
	Volume      string            `yaml:"volume" mapstructure:"volume"`
	Networks    []string          `yaml:"networks" mapstructure:"networks"`
	Command     []string          `yaml:"command" mapstructure:"command"`
	HealthCheck *HealthCheck      `yaml:"health_check" mapstructure:"health_check"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	Test     []string `yaml:"test" mapstructure:"test"`
	Interval string   `yaml:"interval" mapstructure:"interval"`
	Timeout  string   `yaml:"timeout" mapstructure:"timeout"`
	Retries  int      `yaml:"retries" mapstructure:"retries"`
}

// LoadConfig loads the configuration from file or returns defaults
func LoadConfig() (*Config, error) {
	return LoadConfigFromFile("")
}

// LoadConfigFromFile loads configuration from the specified file path,
// or uses default detection if path is empty
func LoadConfigFromFile(configPath string) (*Config, error) {
	if configPath == "" {
		// Use viper to get the config file path (respects --config flag)
		configPath = viper.ConfigFileUsed()
		if configPath == "" {
			// If no config file was found by viper, try default paths
			configPath = GetConfigPath()
		}
	}

	// Read the YAML file directly to preserve key case
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set default profile if not specified
	if config.Profile == "" {
		config.Profile = "dev"
	}

	return &config, nil
}

// GenerateDefaultConfig creates a default .nizam.yaml file
func GenerateDefaultConfig() error {
	configPath := ".nizam.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file %s already exists", configPath)
	}

	defaultConfig := Config{
		Profile: "dev",
		Services: map[string]Service{
			"postgres": {
				Image: "postgres:16",
				Ports: []string{"5432:5432"},
				Environment: map[string]string{
					"POSTGRES_USER":     "user",
					"POSTGRES_PASSWORD": "password",
					"POSTGRES_DB":       "myapp",
				},
				Volume: "pgdata",
			},
			"redis": {
				Image: "redis:7",
				Ports: []string{"6379:6379"},
			},
			"meilisearch": {
				Image: "getmeili/meilisearch:v1.5",
				Ports: []string{"7700:7700"},
				Environment: map[string]string{
					"MEILI_NO_ANALYTICS": "true",
				},
			},
		},
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetService returns a specific service configuration
func (c *Config) GetService(name string) (Service, bool) {
	service, exists := c.Services[name]
	return service, exists
}

// GetAllServices returns all configured services
func (c *Config) GetAllServices() map[string]Service {
	return c.Services
}

// GetServiceNames returns a list of all service names
func (c *Config) GetServiceNames() []string {
	names := make([]string, 0, len(c.Services))
	for name := range c.Services {
		names = append(names, name)
	}
	return names
}

// ConfigExists checks if a nizam config file exists in the current directory
func ConfigExists() bool {
	configPaths := []string{".nizam.yaml", ".nizam.yml"}
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
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
