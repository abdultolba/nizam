package templates

import (
	"slices"
	"strings"
	"testing"

	"github.com/abdultolba/nizam/internal/config"
)

func TestAllBuiltinTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Expected template names
	expectedTemplates := []string{
		"postgres", "postgres-15", "mysql", "mongodb", "redis", "redis-stack",
		"clickhouse", "elasticsearch", "meilisearch", "rabbitmq", "kafka",
		"minio", "nats", "prometheus", "grafana", "jaeger", "mailhog",
		"pinecone-local", "pinecone-index",
	}

	// Check that all expected templates exist
	for _, expectedName := range expectedTemplates {
		if _, exists := templates[expectedName]; !exists {
			t.Errorf("Expected template '%s' not found", expectedName)
		}
	}

	// Check that template count matches expectation
	if len(templates) != len(expectedTemplates) {
		t.Errorf("Expected %d templates, got %d", len(expectedTemplates), len(templates))
	}

	// Validate each template structure
	for name, template := range templates {
		t.Run(name, func(t *testing.T) {
			validateTemplateStructure(t, template)
		})
	}
}

func validateTemplateStructure(t *testing.T, template Template) {
	// Name should not be empty
	if template.Name == "" {
		t.Error("Template name should not be empty")
	}

	// Description should not be empty
	if template.Description == "" {
		t.Error("Template description should not be empty")
	}

	// Should have at least one tag
	if len(template.Tags) == 0 {
		t.Error("Template should have at least one tag")
	}

	// Image should not be empty
	if template.Service.Image == "" {
		t.Error("Template service image should not be empty")
	}

	// Validate variables if present
	for _, variable := range template.Variables {
		validateTemplateVariable(t, variable)
	}

	// Validate health check if present
	if template.Service.HealthCheck != nil {
		validateHealthCheck(t, *template.Service.HealthCheck)
	}
}

func validateTemplateVariable(t *testing.T, variable Variable) {
	// Name should not be empty
	if variable.Name == "" {
		t.Error("Variable name should not be empty")
	}

	// Description should not be empty
	if variable.Description == "" {
		t.Errorf("Variable '%s' should have a description", variable.Name)
	}

	// Type should be valid if specified
	if variable.Type != "" {
		validTypes := []string{"string", "int", "bool", "port"}
		found := slices.Contains(validTypes, variable.Type)
		if !found {
			t.Errorf("Variable '%s' has invalid type '%s'", variable.Name, variable.Type)
		}
	}

	// Port validation is recommended for port type (but not required for legacy templates)
	if variable.Type == "port" && variable.Validation == "" {
		// Only warn for new templates that should have validation
		// ClickHouse has validation as an example of best practice
		t.Logf("Port variable '%s' could benefit from validation pattern", variable.Name)
	}
}

func validateHealthCheck(t *testing.T, healthCheck config.HealthCheck) {
	// Should have test command
	if len(healthCheck.Test) == 0 {
		t.Error("Health check should have test command")
	}

	// Interval should be specified
	if healthCheck.Interval == "" {
		t.Error("Health check should have interval")
	}

	// Timeout should be specified
	if healthCheck.Timeout == "" {
		t.Error("Health check should have timeout")
	}

	// Retries should be positive
	if healthCheck.Retries <= 0 {
		t.Error("Health check should have positive retries count")
	}
}

// Test specific templates for their unique characteristics

func TestPostgresTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test main postgres template
	postgres := templates["postgres"]
	if postgres.Service.Image != "postgres:16" {
		t.Errorf("Expected postgres image 'postgres:16', got %s", postgres.Service.Image)
	}

	// Should have interactive variables
	if len(postgres.Variables) == 0 {
		t.Error("Postgres template should have interactive variables")
	}

	// Check required variables exist
	requiredVars := []string{"DB_USER", "DB_PASSWORD", "DB_NAME"}
	for _, reqVar := range requiredVars {
		found := false
		for _, variable := range postgres.Variables {
			if variable.Name == reqVar && variable.Required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Postgres template missing required variable: %s", reqVar)
		}
	}

	// Test postgres-15 template
	postgres15 := templates["postgres-15"]
	if postgres15.Service.Image != "postgres:15" {
		t.Errorf("Expected postgres-15 image 'postgres:15', got %s", postgres15.Service.Image)
	}
}

func TestMySQLTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	mysql := templates["mysql"]

	// Check image
	if mysql.Service.Image != "mysql:8.0" {
		t.Errorf("Expected mysql image 'mysql:8.0', got %s", mysql.Service.Image)
	}

	// Should have health check
	if mysql.Service.HealthCheck == nil {
		t.Error("MySQL template should have health check")
	}

	// Check for ROOT_PASSWORD variable (unique to MySQL)
	hasRootPassword := false
	for _, variable := range mysql.Variables {
		if variable.Name == "ROOT_PASSWORD" {
			hasRootPassword = true
			break
		}
	}
	if !hasRootPassword {
		t.Error("MySQL template should have ROOT_PASSWORD variable")
	}
}

func TestMongoDBTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	mongodb := templates["mongodb"]

	// Should have mongo image template
	if !strings.HasPrefix(mongodb.Service.Image, "mongo:") {
		t.Errorf("Expected MongoDB image to start with 'mongo:', got %s", mongodb.Service.Image)
	}

	// Check tags include nosql
	hasNoSQLTag := false
	for _, tag := range mongodb.Tags {
		if tag == "nosql" {
			hasNoSQLTag = true
			break
		}
	}
	if !hasNoSQLTag {
		t.Error("MongoDB template should have 'nosql' tag")
	}
}

func TestRedisTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test redis template
	redis := templates["redis"]
	if !strings.Contains(redis.Service.Image, "redis:") && !strings.Contains(redis.Service.Image, "{{.VERSION}}") {
		t.Errorf("Expected redis image to contain 'redis:', got %s", redis.Service.Image)
	}

	// Should have health check
	if redis.Service.HealthCheck == nil {
		t.Error("Redis template should have health check")
	}

	// Test redis-stack template
	redisStack := templates["redis-stack"]
	if redisStack.Service.Image != "redis/redis-stack:latest" {
		t.Errorf("Expected redis-stack image 'redis/redis-stack:latest', got %s", redisStack.Service.Image)
	}

	// Redis-stack should have modules tag
	hasModulesTag := false
	for _, tag := range redisStack.Tags {
		if tag == "modules" {
			hasModulesTag = true
			break
		}
	}
	if !hasModulesTag {
		t.Error("Redis-stack template should have 'modules' tag")
	}
}

func TestClickHouseTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	clickhouse := templates["clickhouse"]

	// Should have multiple ports
	if len(clickhouse.Service.Ports) != 3 {
		t.Errorf("ClickHouse should have 3 ports, got %d", len(clickhouse.Service.Ports))
	}

	// Should have olap tag
	hasOLAPTag := slices.Contains(clickhouse.Tags, "olap")
	if !hasOLAPTag {
		t.Error("ClickHouse template should have 'olap' tag")
	}

	// Should have multiple port variables
	portVars := 0
	for _, variable := range clickhouse.Variables {
		if variable.Type == "port" {
			portVars++
		}
	}
	if portVars != 3 {
		t.Errorf("ClickHouse should have 3 port variables, got %d", portVars)
	}
}

func TestElasticsearchTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()
	elasticsearch := templates["elasticsearch"]

	// Should have elastic.co image
	if !strings.Contains(elasticsearch.Service.Image, "elastic.co") {
		t.Errorf("Expected Elasticsearch image to contain 'elastic.co', got %s", elasticsearch.Service.Image)
	}

	// Should have search tag
	hasSearchTag := slices.Contains(elasticsearch.Tags, "search")
	if !hasSearchTag {
		t.Error("Elasticsearch template should have 'search' tag")
	}
}

func TestMessagingTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test RabbitMQ
	rabbitmq := templates["rabbitmq"]
	if !strings.Contains(rabbitmq.Service.Image, "rabbitmq") {
		t.Errorf("Expected RabbitMQ image to contain 'rabbitmq', got %s", rabbitmq.Service.Image)
	}

	// Should have AMQP tag
	hasAMQPTag := slices.Contains(rabbitmq.Tags, "amqp")
	if !hasAMQPTag {
		t.Error("RabbitMQ template should have 'amqp' tag")
	}

	// Test Kafka (Redpanda)
	kafka := templates["kafka"]
	if !strings.Contains(kafka.Service.Image, "redpanda") {
		t.Errorf("Expected Kafka image to contain 'redpanda', got %s", kafka.Service.Image)
	}

	// Should have streaming tag
	hasStreamingTag := slices.Contains(kafka.Tags, "streaming")
	if !hasStreamingTag {
		t.Error("Kafka template should have 'streaming' tag")
	}

	// Test NATS
	nats := templates["nats"]
	if !strings.Contains(nats.Service.Image, "nats") {
		t.Errorf("Expected NATS image to contain 'nats', got %s", nats.Service.Image)
	}
}

func TestMonitoringTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test Prometheus
	prometheus := templates["prometheus"]
	if !strings.Contains(prometheus.Service.Image, "prometheus") {
		t.Errorf("Expected Prometheus image to contain 'prometheus', got %s", prometheus.Service.Image)
	}

	// Should have metrics tag
	hasMetricsTag := slices.Contains(prometheus.Tags, "metrics")
	if !hasMetricsTag {
		t.Error("Prometheus template should have 'metrics' tag")
	}

	// Test Grafana
	grafana := templates["grafana"]
	if !strings.Contains(grafana.Service.Image, "grafana") {
		t.Errorf("Expected Grafana image to contain 'grafana', got %s", grafana.Service.Image)
	}

	// Should have visualization tag
	hasVisualizationTag := slices.Contains(grafana.Tags, "visualization")
	if !hasVisualizationTag {
		t.Error("Grafana template should have 'visualization' tag")
	}

	// Test Jaeger
	jaeger := templates["jaeger"]
	if !strings.Contains(jaeger.Service.Image, "jaeger") {
		t.Errorf("Expected Jaeger image to contain 'jaeger', got %s", jaeger.Service.Image)
	}

	// Should have tracing tag
	hasTracingTag := slices.Contains(jaeger.Tags, "tracing")
	if !hasTracingTag {
		t.Error("Jaeger template should have 'tracing' tag")
	}
}

func TestSearchTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test Meilisearch
	meilisearch := templates["meilisearch"]
	if !strings.Contains(meilisearch.Service.Image, "meilisearch") {
		t.Errorf("Expected Meilisearch image to contain 'meilisearch', got %s", meilisearch.Service.Image)
	}

	// Should have search tag
	hasSearchTag := slices.Contains(meilisearch.Tags, "search")
	if !hasSearchTag {
		t.Error("Meilisearch template should have 'search' tag")
	}
}

func TestStorageTemplate(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test MinIO
	minio := templates["minio"]
	if !strings.Contains(minio.Service.Image, "minio") {
		t.Errorf("Expected MinIO image to contain 'minio', got %s", minio.Service.Image)
	}

	// Should have s3 tag
	hasS3Tag := slices.Contains(minio.Tags, "s3")
	if !hasS3Tag {
		t.Error("MinIO template should have 's3' tag")
	}

	// Should have multiple ports
	if len(minio.Service.Ports) != 2 {
		t.Errorf("MinIO should have 2 ports, got %d", len(minio.Service.Ports))
	}
}

func TestDevelopmentToolTemplates(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test MailHog
	mailhog := templates["mailhog"]
	if !strings.Contains(mailhog.Service.Image, "mailhog") {
		t.Errorf("Expected MailHog image to contain 'mailhog', got %s", mailhog.Service.Image)
	}

	// Should have email and testing tags
	hasEmailTag := false
	hasTestingTag := false
	for _, tag := range mailhog.Tags {
		if tag == "email" {
			hasEmailTag = true
		}
		if tag == "testing" {
			hasTestingTag = true
		}
	}
	if !hasEmailTag {
		t.Error("MailHog template should have 'email' tag")
	}
	if !hasTestingTag {
		t.Error("MailHog template should have 'testing' tag")
	}

	// Should have two ports (SMTP and HTTP)
	if len(mailhog.Service.Ports) != 2 {
		t.Errorf("MailHog should have 2 ports, got %d", len(mailhog.Service.Ports))
	}
}

func TestTemplateTagConsistency(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test that database templates have database tag
	databaseTemplates := []string{"postgres", "postgres-15", "mysql", "mongodb", "redis", "redis-stack", "clickhouse"}
	for _, templateName := range databaseTemplates {
		template := templates[templateName]
		hasDatabaseTag := slices.Contains(template.Tags, "database")
		if !hasDatabaseTag {
			t.Errorf("Template '%s' should have 'database' tag", templateName)
		}
	}

	// Test that messaging templates have messaging tag
	messagingTemplates := []string{"rabbitmq", "kafka", "nats"}
	for _, templateName := range messagingTemplates {
		template := templates[templateName]
		hasMessagingTag := slices.Contains(template.Tags, "messaging")
		if !hasMessagingTag {
			t.Errorf("Template '%s' should have 'messaging' tag", templateName)
		}
	}

	// Test that monitoring templates have monitoring tag
	monitoringTemplates := []string{"prometheus", "grafana", "jaeger"}
	for _, templateName := range monitoringTemplates {
		template := templates[templateName]
		hasMonitoringTag := slices.Contains(template.Tags, "monitoring")
		if !hasMonitoringTag {
			t.Errorf("Template '%s' should have 'monitoring' tag", templateName)
		}
	}
}

func TestTemplateWithDefaults(t *testing.T) {
	templates := GetBuiltinTemplates()

	// Test templates with variables can be processed with defaults
	templatesWithVariables := []string{"postgres", "mysql", "redis", "mongodb", "rabbitmq", "clickhouse"}
	
	for _, templateName := range templatesWithVariables {
		template := templates[templateName]
		t.Run(templateName+"_defaults", func(t *testing.T) {
			processedService, err := ProcessTemplateWithDefaults(template)
			if err != nil {
				t.Errorf("Failed to process template '%s' with defaults: %v", templateName, err)
				return
			}

			// Processed service should have concrete values (no template variables)
			if strings.Contains(processedService.Image, "{{") {
				t.Errorf("Template '%s' image still contains template variables after processing: %s", templateName, processedService.Image)
			}

			// Environment variables should be processed
			for key, value := range processedService.Environment {
				if strings.Contains(value, "{{") {
					t.Errorf("Template '%s' environment variable %s still contains template variables after processing: %s", templateName, key, value)
				}
			}

			// Ports should be processed
			for _, port := range processedService.Ports {
				if strings.Contains(port, "{{") {
					t.Errorf("Template '%s' port still contains template variables after processing: %s", templateName, port)
				}
			}

			// Volume should be processed
			if strings.Contains(processedService.Volume, "{{") {
				t.Errorf("Template '%s' volume still contains template variables after processing: %s", templateName, processedService.Volume)
			}
		})
	}
}

func TestTemplateTagFiltering(t *testing.T) {
	// Test that each major category returns expected templates
	testCases := []struct {
		tag               string
		minExpectedCount  int
		maxExpectedCount  int
		expectedTemplates []string
	}{
		{"database", 6, 8, []string{"postgres", "mysql", "redis", "clickhouse"}},
		{"messaging", 3, 3, []string{"rabbitmq", "kafka", "nats"}},
		{"monitoring", 3, 3, []string{"prometheus", "grafana", "jaeger"}},
		{"search", 2, 2, []string{"elasticsearch", "meilisearch"}},
		{"nosql", 3, 3, []string{"mongodb", "redis", "redis-stack"}},
		{"sql", 3, 3, []string{"postgres", "mysql"}},
		{"analytics", 2, 2, []string{"elasticsearch", "clickhouse"}},
	}

	for _, testCase := range testCases {
		t.Run("tag_"+testCase.tag, func(t *testing.T) {
			filteredTemplates := GetTemplatesByTag(testCase.tag)
			
			if len(filteredTemplates) < testCase.minExpectedCount {
				t.Errorf("Tag '%s' returned %d templates, expected at least %d", testCase.tag, len(filteredTemplates), testCase.minExpectedCount)
			}
			
			if len(filteredTemplates) > testCase.maxExpectedCount {
				t.Errorf("Tag '%s' returned %d templates, expected at most %d", testCase.tag, len(filteredTemplates), testCase.maxExpectedCount)
			}

			// Check that expected templates are present
			for _, expectedTemplate := range testCase.expectedTemplates {
				found := false
				for _, template := range filteredTemplates {
					if template.Name == expectedTemplate {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Tag '%s' should include template '%s'", testCase.tag, expectedTemplate)
				}
			}
		})
	}
}

func TestUniqueTemplateNames(t *testing.T) {
	templates := GetBuiltinTemplates()
	seen := make(map[string]bool)

	for name := range templates {
		if seen[name] {
			t.Errorf("Duplicate template name found: %s", name)
		}
		seen[name] = true
	}
}

func TestTemplateImageValidity(t *testing.T) {
	templates := GetBuiltinTemplates()

	for name, template := range templates {
		t.Run(name+"_image", func(t *testing.T) {
			image := template.Service.Image
			
			// Image should not be empty
			if image == "" {
				t.Error("Template image should not be empty")
				return
			}

			// Image should have a reasonable format
			if !strings.Contains(image, ":") && !strings.Contains(image, "{{") {
				t.Errorf("Image '%s' should contain ':' for tag or template variable", image)
			}

			// Should not use latest tag unless it's explicit
			if strings.HasSuffix(image, ":latest") {
				t.Logf("Template '%s' uses :latest tag - consider pinning to specific version", name)
			}
		})
	}
}
