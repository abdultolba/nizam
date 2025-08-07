package templates

import (
	"fmt"

	"github.com/abdultolba/nizam/internal/config"
)

// Template represents a service template
type Template struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Service     config.Service `json:"service"`
	Tags        []string       `json:"tags"`
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
				Ports: []string{"5432:5432"},
				Environment: map[string]string{
					"POSTGRES_USER":     "user",
					"POSTGRES_PASSWORD": "password",
					"POSTGRES_DB":       "myapp",
				},
				Volume: "pgdata",
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
				Ports: []string{"3306:3306"},
				Environment: map[string]string{
					"MYSQL_ROOT_PASSWORD": "rootpassword",
					"MYSQL_DATABASE":      "myapp",
					"MYSQL_USER":          "user",
					"MYSQL_PASSWORD":      "password",
				},
				Volume: "mysqldata",
			},
		},
		"redis": {
			Name:        "redis",
			Description: "Redis in-memory data store",
			Tags:        []string{"cache", "database", "nosql", "redis"},
			Service: config.Service{
				Image: "redis:7",
				Ports: []string{"6379:6379"},
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
				Image: "mongo:7",
				Ports: []string{"27017:27017"},
				Environment: map[string]string{
					"MONGO_INITDB_ROOT_USERNAME": "admin",
					"MONGO_INITDB_ROOT_PASSWORD": "password",
				},
				Volume: "mongodata",
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
				Image: "rabbitmq:3-management",
				Ports: []string{"5672:5672", "15672:15672"},
				Environment: map[string]string{
					"RABBITMQ_DEFAULT_USER": "admin",
					"RABBITMQ_DEFAULT_PASS": "password",
				},
				Volume: "rabbitmqdata",
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

// GetTemplate returns a template by name
func GetTemplate(name string) (Template, error) {
	templates := GetBuiltinTemplates()
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
	templates := GetBuiltinTemplates()
	tagMap := make(map[string]bool)
	
	for _, template := range templates {
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
