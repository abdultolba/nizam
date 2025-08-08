package templates

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/abdultolba/nizam/internal/config"
	"gopkg.in/yaml.v3"
)

// Template represents a service template
type Template struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Service     config.Service `json:"service"`
	Tags        []string       `json:"tags"`
	Variables   []Variable     `json:"variables,omitempty"`
}

// Variable represents a customizable template variable
type Variable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Type        string `json:"type,omitempty"` // "string", "int", "bool", "port"
	Validation  string `json:"validation,omitempty"` // regex or validation rules
}

// GetBuiltinTemplates returns all built-in service templates
func GetBuiltinTemplates() map[string]Template {
	return map[string]Template{
		"postgres": {
			Name:        "postgres",
			Description: "PostgreSQL database server (latest stable version)",
			Tags:        []string{"database", "sql", "postgres"},
			Service: config.Service{
				Image: "postgres:16",
				Ports: []string{"{{.PORT}}:5432"},
				Environment: map[string]string{
					"POSTGRES_USER":     "{{.DB_USER}}",
					"POSTGRES_PASSWORD": "{{.DB_PASSWORD}}",
					"POSTGRES_DB":       "{{.DB_NAME}}",
				},
				Volume: "{{.VOLUME_NAME}}",
				HealthCheck: &config.HealthCheck{
					Test:     []string{"CMD-SHELL", "pg_isready -U {{.DB_USER}} -d {{.DB_NAME}}"},
					Interval: "30s",
					Timeout:  "10s",
					Retries:  3,
				},
			},
			Variables: []Variable{
				{
					Name:        "DB_USER",
					Description: "PostgreSQL username",
					Default:     "user",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "DB_PASSWORD",
					Description: "PostgreSQL password",
					Default:     "password",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "DB_NAME",
					Description: "Database name to create",
					Default:     "myapp",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "PORT",
					Description: "Host port to bind PostgreSQL",
					Default:     "5432",
					Required:    false,
					Type:        "port",
					Validation:  "^[1-9][0-9]{0,4}$",
				},
				{
					Name:        "VOLUME_NAME",
					Description: "Docker volume name for data persistence",
					Default:     "pgdata",
					Required:    false,
					Type:        "string",
				},
			},
		},
		"postgres-15": {
			Name:        "postgres-15",
			Description: "PostgreSQL 15 database server",
			Tags:        []string{"database", "sql", "postgres"},
			Service: config.Service{
				Image: "postgres:15",
				Ports: []string{"5432:5432"},
				Environment: map[string]string{
					"POSTGRES_USER":     "user",
					"POSTGRES_PASSWORD": "password",
					"POSTGRES_DB":       "myapp",
				},
				Volume: "pgdata",
			},
		},
		"mysql": {
			Name:        "mysql",
			Description: "MySQL database server",
			Tags:        []string{"database", "sql", "mysql"},
			Service: config.Service{
				Image: "mysql:8.0",
				Ports: []string{"{{.PORT}}:3306"},
				Environment: map[string]string{
					"MYSQL_ROOT_PASSWORD": "{{.ROOT_PASSWORD}}",
					"MYSQL_DATABASE":      "{{.DB_NAME}}",
					"MYSQL_USER":          "{{.DB_USER}}",
					"MYSQL_PASSWORD":      "{{.DB_PASSWORD}}",
				},
				Volume: "{{.VOLUME_NAME}}",
			},
			Variables: []Variable{
				{
					Name:        "DB_USER",
					Description: "MySQL username",
					Default:     "user",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "DB_PASSWORD",
					Description: "MySQL user password",
					Default:     "password",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "ROOT_PASSWORD",
					Description: "MySQL root password",
					Default:     "rootpassword",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "DB_NAME",
					Description: "Database name to create",
					Default:     "myapp",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "PORT",
					Description: "Host port to bind MySQL",
					Default:     "3306",
					Required:    false,
					Type:        "port",
				},
				{
					Name:        "VOLUME_NAME",
					Description: "Docker volume name for data persistence",
					Default:     "mysqldata",
					Required:    false,
					Type:        "string",
				},
			},
		},
		"redis": {
			Name:        "redis",
			Description: "Redis in-memory data store",
			Tags:        []string{"cache", "database", "nosql", "redis"},
			Service: config.Service{
				Image: "redis:{{.VERSION}}",
				Ports: []string{"{{.PORT}}:6379"},
				Volume: "{{.VOLUME_NAME}}",
				HealthCheck: &config.HealthCheck{
					Test:     []string{"CMD", "redis-cli", "ping"},
					Interval: "30s",
					Timeout:  "3s",
					Retries:  3,
				},
			},
			Variables: []Variable{
				{
					Name:        "VERSION",
					Description: "Redis version",
					Default:     "7",
					Required:    false,
					Type:        "string",
				},
				{
					Name:        "PORT",
					Description: "Host port to bind Redis",
					Default:     "6379",
					Required:    false,
					Type:        "port",
				},
				{
					Name:        "PASSWORD",
					Description: "Redis password (optional)",
					Default:     "",
					Required:    false,
					Type:        "string",
				},
				{
					Name:        "VOLUME_NAME",
					Description: "Docker volume name for data persistence",
					Default:     "redisdata",
					Required:    false,
					Type:        "string",
				},
			},
		},
		"redis-stack": {
			Name:        "redis-stack",
			Description: "Redis Stack with modules (RedisJSON, RedisGraph, etc.)",
			Tags:        []string{"cache", "database", "nosql", "redis", "modules"},
			Service: config.Service{
				Image: "redis/redis-stack:latest",
				Ports: []string{"6379:6379", "8001:8001"},
			},
		},
		"mongodb": {
			Name:        "mongodb",
			Description: "MongoDB document database",
			Tags:        []string{"database", "nosql", "mongodb"},
			Service: config.Service{
				Image: "mongo:{{.VERSION}}",
				Ports: []string{"{{.PORT}}:27017"},
				Environment: map[string]string{
					"MONGO_INITDB_ROOT_USERNAME": "{{.ROOT_USERNAME}}",
					"MONGO_INITDB_ROOT_PASSWORD": "{{.ROOT_PASSWORD}}",
				},
				Volume: "{{.VOLUME_NAME}}",
			},
			Variables: []Variable{
				{
					Name:        "VERSION",
					Description: "MongoDB version",
					Default:     "7",
					Required:    false,
					Type:        "string",
				},
				{
					Name:        "ROOT_USERNAME",
					Description: "MongoDB root username",
					Default:     "admin",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "ROOT_PASSWORD",
					Description: "MongoDB root password",
					Default:     "password",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "PORT",
					Description: "Host port to bind MongoDB",
					Default:     "27017",
					Required:    false,
					Type:        "port",
				},
				{
					Name:        "VOLUME_NAME",
					Description: "Docker volume name for data persistence",
					Default:     "mongodata",
					Required:    false,
					Type:        "string",
				},
			},
		},
		"elasticsearch": {
			Name:        "elasticsearch",
			Description: "Elasticsearch search engine",
			Tags:        []string{"search", "elasticsearch", "analytics"},
			Service: config.Service{
				Image: "docker.elastic.co/elasticsearch/elasticsearch:8.11.0",
				Ports: []string{"9200:9200", "9300:9300"},
				Environment: map[string]string{
					"discovery.type":         "single-node",
					"xpack.security.enabled": "false",
					"ES_JAVA_OPTS":           "-Xms512m -Xmx512m",
				},
				Volume: "esdata",
			},
		},
		"meilisearch": {
			Name:        "meilisearch",
			Description: "Meilisearch fast search engine",
			Tags:        []string{"search", "meilisearch"},
			Service: config.Service{
				Image: "getmeili/meilisearch:v1.5",
				Ports: []string{"7700:7700"},
				Environment: map[string]string{
					"MEILI_NO_ANALYTICS": "true",
				},
			},
		},
		"rabbitmq": {
			Name:        "rabbitmq",
			Description: "RabbitMQ message broker",
			Tags:        []string{"messaging", "rabbitmq", "amqp"},
			Service: config.Service{
				Image: "rabbitmq:{{.VERSION}}",
				Ports: []string{"{{.AMQP_PORT}}:5672", "{{.MANAGEMENT_PORT}}:15672"},
				Environment: map[string]string{
					"RABBITMQ_DEFAULT_USER": "{{.DEFAULT_USER}}",
					"RABBITMQ_DEFAULT_PASS": "{{.DEFAULT_PASS}}",
				},
				Volume: "{{.VOLUME_NAME}}",
			},
			Variables: []Variable{
				{
					Name:        "VERSION",
					Description: "RabbitMQ version (with management plugin)",
					Default:     "3-management",
					Required:    false,
					Type:        "string",
				},
				{
					Name:        "DEFAULT_USER",
					Description: "RabbitMQ default username",
					Default:     "admin",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "DEFAULT_PASS",
					Description: "RabbitMQ default password",
					Default:     "password",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "AMQP_PORT",
					Description: "Host port for AMQP protocol",
					Default:     "5672",
					Required:    false,
					Type:        "port",
				},
				{
					Name:        "MANAGEMENT_PORT",
					Description: "Host port for management UI",
					Default:     "15672",
					Required:    false,
					Type:        "port",
				},
				{
					Name:        "VOLUME_NAME",
					Description: "Docker volume name for data persistence",
					Default:     "rabbitmqdata",
					Required:    false,
					Type:        "string",
				},
			},
		},
		"kafka": {
			Name:        "kafka",
			Description: "Apache Kafka (via Redpanda - Kafka-compatible)",
			Tags:        []string{"messaging", "kafka", "streaming"},
			Service: config.Service{
				Image: "docker.redpanda.com/vectorized/redpanda:v23.2.14",
				Ports: []string{"9092:9092", "9644:9644"},
				Command: []string{
					"redpanda", "start",
					"--smp", "1",
					"--memory", "1G",
					"--overprovisioned",
					"--node-id", "0",
					"--kafka-addr", "0.0.0.0:9092",
					"--advertise-kafka-addr", "localhost:9092",
					"--pandaproxy-addr", "0.0.0.0:8082",
					"--advertise-pandaproxy-addr", "localhost:8082",
				},
			},
		},
		"minio": {
			Name:        "minio",
			Description: "MinIO S3-compatible object storage",
			Tags:        []string{"storage", "s3", "minio"},
			Service: config.Service{
				Image: "minio/minio:latest",
				Ports: []string{"9000:9000", "9001:9001"},
				Environment: map[string]string{
					"MINIO_ROOT_USER":     "admin",
					"MINIO_ROOT_PASSWORD": "password123",
				},
				Command: []string{"server", "/data", "--console-address", ":9001"},
				Volume:  "miniodata",
			},
		},
		"nats": {
			Name:        "nats",
			Description: "NATS messaging system",
			Tags:        []string{"messaging", "nats"},
			Service: config.Service{
				Image: "nats:2.10",
				Ports: []string{"4222:4222", "8222:8222"},
			},
		},
		"prometheus": {
			Name:        "prometheus",
			Description: "Prometheus monitoring system",
			Tags:        []string{"monitoring", "metrics", "prometheus"},
			Service: config.Service{
				Image: "prom/prometheus:v2.47.0",
				Ports: []string{"9090:9090"},
				Volume: "prometheusdata",
			},
		},
		"grafana": {
			Name:        "grafana",
			Description: "Grafana visualization platform",
			Tags:        []string{"monitoring", "visualization", "grafana"},
			Service: config.Service{
				Image: "grafana/grafana:10.2.0",
				Ports: []string{"3000:3000"},
				Environment: map[string]string{
					"GF_SECURITY_ADMIN_PASSWORD": "admin",
				},
				Volume: "grafanadata",
			},
		},
		"jaeger": {
			Name:        "jaeger",
			Description: "Jaeger distributed tracing system",
			Tags:        []string{"monitoring", "tracing", "jaeger"},
			Service: config.Service{
				Image: "jaegertracing/all-in-one:1.50",
				Ports: []string{"16686:16686", "14268:14268"},
				Environment: map[string]string{
					"COLLECTOR_ZIPKIN_HOST_PORT": ":9411",
				},
			},
		},
		"mailhog": {
			Name:        "mailhog",
			Description: "MailHog email testing tool",
			Tags:        []string{"email", "testing", "mailhog"},
			Service: config.Service{
				Image: "mailhog/mailhog:v1.0.1",
				Ports: []string{"1025:1025", "8025:8025"},
			},
		},
	}
}

// GetTemplate returns a template by name (built-in or custom)
func GetTemplate(name string) (Template, error) {
	templates := GetAllTemplates()
	template, exists := templates[name]
	if !exists {
		return Template{}, fmt.Errorf("template '%s' not found", name)
	}
	return template, nil
}

// GetTemplateNames returns a list of all available template names
func GetTemplateNames() []string {
	templates := GetBuiltinTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}

// GetTemplatesByTag returns templates that have the specified tag
func GetTemplatesByTag(tag string) []Template {
	templates := GetBuiltinTemplates()
	var filtered []Template
	
	for _, template := range templates {
		for _, templateTag := range template.Tags {
			if templateTag == tag {
				filtered = append(filtered, template)
				break
			}
		}
	}
	
	return filtered
}

// GetAllTags returns all unique tags from all templates
func GetAllTags() []string {
	allTemplates := GetAllTemplates()
	tagMap := make(map[string]bool)
	
	for _, template := range allTemplates {
		for _, tag := range template.Tags {
			tagMap[tag] = true
		}
	}
	
	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	
	return tags
}

// GetCustomTemplatesDir returns the path to the custom templates directory
func GetCustomTemplatesDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".nizam/templates" // fallback to local directory
	}
	return filepath.Join(homeDir, ".nizam", "templates")
}

// GetCustomTemplates returns all custom user templates
func GetCustomTemplates() (map[string]Template, error) {
	templatesDir := GetCustomTemplatesDir()
	customTemplates := make(map[string]Template)
	
	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return customTemplates, nil // No custom templates yet
	}
	
	// Read all .yaml files in the templates directory
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}
	
	for _, file := range files {
		template, err := loadTemplateFromFile(file)
		if err != nil {
			// Log error but continue with other templates
			continue
		}
		
		// Use filename (without extension) as template name if not specified
		if template.Name == "" {
			template.Name = strings.TrimSuffix(filepath.Base(file), ".yaml")
		}
		
		customTemplates[template.Name] = template
	}
	
	return customTemplates, nil
}

// GetAllTemplates returns both built-in and custom templates
func GetAllTemplates() map[string]Template {
	allTemplates := GetBuiltinTemplates()
	
	// Add custom templates
	customTemplates, err := GetCustomTemplates()
	if err == nil {
		for name, template := range customTemplates {
			// Add "custom" tag to distinguish from built-in templates
			if !contains(template.Tags, "custom") {
				template.Tags = append(template.Tags, "custom")
			}
			allTemplates[name] = template
		}
	}
	
	return allTemplates
}

// GetAllTemplateNames returns names of all templates (built-in + custom)
func GetAllTemplateNames() []string {
	templates := GetAllTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}

// GetAllTemplatesByTag returns all templates (built-in + custom) that have the specified tag
func GetAllTemplatesByTag(tag string) []Template {
	templates := GetAllTemplates()
	var filtered []Template
	
	for _, template := range templates {
		for _, templateTag := range template.Tags {
			if templateTag == tag {
				filtered = append(filtered, template)
				break
			}
		}
	}
	
	return filtered
}

// SaveCustomTemplate saves a template to the custom templates directory
func SaveCustomTemplate(template Template) error {
	templatesDir := GetCustomTemplatesDir()
	
	// Create templates directory if it doesn't exist
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}
	
	// Generate filename from template name
	filename := fmt.Sprintf("%s.yaml", template.Name)
	filepath := filepath.Join(templatesDir, filename)
	
	// Marshal template to YAML
	yamlData, err := yaml.Marshal(&template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(filepath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}
	
	return nil
}

// DeleteCustomTemplate removes a custom template
func DeleteCustomTemplate(name string) error {
	templatesDir := GetCustomTemplatesDir()
	filename := fmt.Sprintf("%s.yaml", name)
	filepath := filepath.Join(templatesDir, filename)
	
	// Check if it's a built-in template
	builtinTemplates := GetBuiltinTemplates()
	if _, isBuiltin := builtinTemplates[name]; isBuiltin {
		return fmt.Errorf("cannot delete built-in template '%s'", name)
	}
	
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return fmt.Errorf("custom template '%s' not found", name)
	}
	
	// Delete the file
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete template file: %w", err)
	}
	
	return nil
}

// loadTemplateFromFile loads a template from a YAML file
func loadTemplateFromFile(filepath string) (Template, error) {
	var template Template
	
	data, err := os.ReadFile(filepath)
	if err != nil {
		return template, fmt.Errorf("failed to read template file: %w", err)
	}
	
	if err := yaml.Unmarshal(data, &template); err != nil {
		return template, fmt.Errorf("failed to parse template file: %w", err)
	}
	
	return template, nil
}

// contains checks if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// ProcessTemplateWithVariables processes a template by prompting for variables and substituting them
func ProcessTemplateWithVariables(tmpl Template, serviceName string) (config.Service, error) {
	if len(tmpl.Variables) == 0 {
		// No variables to process, return service as-is
		return tmpl.Service, nil
	}

	fmt.Printf("\n⚙️  Configuring template '%s' for service '%s'\n", tmpl.Name, serviceName)
	fmt.Printf("📝 Please provide values for the following variables:\n\n")

	// Collect variable values from user
	variableValues := make(map[string]string)
	reader := bufio.NewReader(os.Stdin)

	for _, variable := range tmpl.Variables {
		value, err := promptForVariable(variable, reader)
		if err != nil {
			return config.Service{}, fmt.Errorf("failed to get variable '%s': %w", variable.Name, err)
		}
		variableValues[variable.Name] = value
	}

	// Process the template with the collected values
	processedService, err := substituteVariables(tmpl.Service, variableValues)
	if err != nil {
		return config.Service{}, fmt.Errorf("failed to process template: %w", err)
	}

	fmt.Printf("\n✅ Template configured successfully!\n")
	return processedService, nil
}

// promptForVariable prompts the user for a variable value with validation
func promptForVariable(variable Variable, reader *bufio.Reader) (string, error) {
	// Display prompt
	requiredIndicator := ""
	if variable.Required {
		requiredIndicator = " *"
	}

	defaultDisplay := ""
	if variable.Default != "" {
		defaultDisplay = fmt.Sprintf(" [%s]", variable.Default)
	}

	fmt.Printf("  %s%s: %s%s\n", variable.Name, requiredIndicator, variable.Description, defaultDisplay)
	if variable.Type != "" {
		fmt.Printf("    Type: %s", variable.Type)
		if variable.Validation != "" {
			fmt.Printf(" (pattern: %s)", variable.Validation)
		}
		fmt.Println()
	}

	for {
		fmt.Printf("    > ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		value := strings.TrimSpace(input)

		// Use default if empty and default exists
		if value == "" {
			if variable.Default != "" {
				value = variable.Default
			} else if variable.Required {
				fmt.Printf("    ❌ This field is required. Please provide a value.\n")
				continue
			}
		}

		// Validate the value
		if err := validateVariable(variable, value); err != nil {
			fmt.Printf("    ❌ %s Please try again.\n", err)
			continue
		}

		fmt.Printf("    ✅ %s = %s\n\n", variable.Name, value)
		return value, nil
	}
}

// validateVariable validates a variable value based on its type and validation rules
func validateVariable(variable Variable, value string) error {
	if value == "" && !variable.Required {
		return nil // Empty is OK for non-required variables
	}

	// Type-based validation
	switch variable.Type {
	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid integer value '%s'", value)
		}
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port number '%s'", value)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("port number must be between 1 and 65535")
		}
	case "bool":
		lowerValue := strings.ToLower(value)
		if lowerValue != "true" && lowerValue != "false" && lowerValue != "yes" && lowerValue != "no" {
			return fmt.Errorf("boolean value must be true/false or yes/no")
		}
	}

	// Regex validation if provided
	if variable.Validation != "" {
		matched, err := regexp.MatchString(variable.Validation, value)
		if err != nil {
			return fmt.Errorf("invalid validation pattern: %w", err)
		}
		if !matched {
			return fmt.Errorf("value doesn't match required pattern")
		}
	}

	return nil
}

// substituteVariables substitutes template variables in a service configuration
func substituteVariables(service config.Service, variables map[string]string) (config.Service, error) {
	// Convert service to YAML, substitute variables, then convert back
	yamlData, err := yaml.Marshal(&service)
	if err != nil {
		return service, fmt.Errorf("failed to marshal service: %w", err)
	}

	// Create template and execute substitution
	tmpl, err := template.New("service").Parse(string(yamlData))
	if err != nil {
		return service, fmt.Errorf("failed to parse service template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return service, fmt.Errorf("failed to execute template: %w", err)
	}

	// Unmarshal back to service struct
	var processedService config.Service
	if err := yaml.Unmarshal(buf.Bytes(), &processedService); err != nil {
		return service, fmt.Errorf("failed to unmarshal processed service: %w", err)
	}

	return processedService, nil
}

// HasVariables returns true if the template has customizable variables
func (t Template) HasVariables() bool {
	return len(t.Variables) > 0
}

// ProcessTemplateWithDefaults processes a template using only default values
func ProcessTemplateWithDefaults(tmpl Template) (config.Service, error) {
	if len(tmpl.Variables) == 0 {
		// No variables to process, return service as-is
		return tmpl.Service, nil
	}

	// Collect default values
	variableValues := make(map[string]string)
	for _, variable := range tmpl.Variables {
		if variable.Default != "" {
			variableValues[variable.Name] = variable.Default
		} else if variable.Required {
			return config.Service{}, fmt.Errorf("required variable '%s' has no default value", variable.Name)
		}
		// For non-required variables without defaults, we leave them empty
	}

	// Process the template with default values
	processedService, err := substituteVariables(tmpl.Service, variableValues)
	if err != nil {
		return config.Service{}, fmt.Errorf("failed to process template with defaults: %w", err)
	}

	return processedService, nil
}
