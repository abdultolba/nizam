# Nizam Service Templates

This directory contains the built-in service templates that power `nizam`'s quick service deployment capabilities. Each template provides a pre-configured Docker service with sensible defaults and optional interactive customization.

## Overview

Nizam includes **17 built-in templates** across different categories:
- **7 Database templates** for different storage and analytics needs
- **3 Messaging templates** for event streaming and queuing
- **3 Monitoring templates** for observability and metrics
- **2 Search templates** for full-text and analytics search
- **1 Storage template** for object storage
- **1 Development tool** for email testing

## Available Templates

### 🗄️ Databases

#### PostgreSQL
- **Templates**: `postgres`, `postgres-15`
- **Image**: [postgres:16](https://hub.docker.com/_/postgres), [postgres:15](https://hub.docker.com/_/postgres)
- **Description**: Industry-standard relational database with ACID compliance
- **Ports**: 5432 (configurable)
- **Interactive Variables**: ✅ Username, password, database name, port, volume
- **Health Check**: `pg_isready` command
- **Tags**: `database`, `sql`, `postgres`

#### MySQL
- **Template**: `mysql`
- **Image**: [mysql:8.0](https://hub.docker.com/_/mysql)
- **Description**: Popular open-source relational database
- **Ports**: 3306 (configurable)
- **Interactive Variables**: ✅ Username, password, root password, database name, port, volume
- **Health Check**: `mysqladmin ping`
- **Tags**: `database`, `sql`, `mysql`

#### MongoDB
- **Template**: `mongodb`
- **Image**: [mongo:7](https://hub.docker.com/_/mongo)
- **Description**: Document-oriented NoSQL database
- **Ports**: 27017 (configurable)
- **Interactive Variables**: ✅ Version, root username, root password, port, volume
- **Tags**: `database`, `nosql`, `mongodb`

#### Redis
- **Templates**: `redis`, `redis-stack`
- **Images**: [redis:7](https://hub.docker.com/_/redis), [redis/redis-stack:latest](https://hub.docker.com/r/redis/redis-stack)
- **Description**: In-memory data store and cache
- **Ports**: 6379 (redis), 6379+8001 (redis-stack)
- **Interactive Variables**: ✅ Version, port, password (optional), volume
- **Health Check**: `redis-cli ping`
- **Tags**: `cache`, `database`, `nosql`, `redis` (+`modules` for redis-stack)

#### ClickHouse
- **Template**: `clickhouse`
- **Image**: [clickhouse/clickhouse-server:24.1](https://hub.docker.com/r/clickhouse/clickhouse-server)
- **Description**: Column-oriented OLAP database for analytics
- **Ports**: 8123 (HTTP), 9000 (Native), 9004 (MySQL compatibility)
- **Interactive Variables**: ✅ Version, username, password, database, all ports, volume
- **Health Check**: `clickhouse-client` query
- **Tags**: `database`, `analytics`, `olap`, `clickhouse`
- **Documentation**: [📖 ClickHouse Guide](../../docs/CLICKHOUSE.md)

#### Elasticsearch
- **Template**: `elasticsearch`
- **Image**: [docker.elastic.co/elasticsearch/elasticsearch:8.11.0](https://www.docker.elastic.co/r/elasticsearch)
- **Description**: Distributed search and analytics engine
- **Ports**: 9200, 9300
- **Configuration**: Single-node mode, security disabled for development
- **Tags**: `search`, `elasticsearch`, `analytics`

### 📨 Messaging & Streaming

#### RabbitMQ
- **Template**: `rabbitmq`
- **Image**: [rabbitmq:3-management](https://hub.docker.com/_/rabbitmq)
- **Description**: Message broker with AMQP protocol support
- **Ports**: 5672 (AMQP), 15672 (Management UI) - both configurable
- **Interactive Variables**: ✅ Version, username, password, ports, volume
- **Tags**: `messaging`, `rabbitmq`, `amqp`

#### Apache Kafka (via Redpanda)
- **Template**: `kafka`
- **Image**: [docker.redpanda.com/vectorized/redpanda:v23.2.14](https://hub.docker.com/r/redpandadata/redpanda)
- **Description**: Kafka-compatible streaming platform
- **Ports**: 9092 (Kafka), 9644 (Admin)
- **Configuration**: Single-node development setup
- **Tags**: `messaging`, `kafka`, `streaming`

#### NATS
- **Template**: `nats`
- **Image**: [nats:2.10](https://hub.docker.com/_/nats)
- **Description**: Lightweight messaging system
- **Ports**: 4222 (Client), 8222 (HTTP monitoring)
- **Tags**: `messaging`, `nats`

### 📊 Monitoring & Observability

#### Prometheus
- **Template**: `prometheus`
- **Image**: [prom/prometheus:v2.47.0](https://hub.docker.com/r/prom/prometheus)
- **Description**: Metrics collection and monitoring system
- **Ports**: 9090
- **Volume**: Persistent data storage
- **Tags**: `monitoring`, `metrics`, `prometheus`

#### Grafana
- **Template**: `grafana`
- **Image**: [grafana/grafana:10.2.0](https://hub.docker.com/r/grafana/grafana)
- **Description**: Visualization and dashboarding platform
- **Ports**: 3000
- **Default Credentials**: admin/admin
- **Volume**: Dashboard and configuration persistence
- **Tags**: `monitoring`, `visualization`, `grafana`

#### Jaeger
- **Template**: `jaeger`
- **Image**: [jaegertracing/all-in-one:1.50](https://hub.docker.com/r/jaegertracing/jaeger)
- **Description**: Distributed tracing system
- **Ports**: 16686 (UI), 14268 (Collector)
- **Configuration**: All-in-one development setup
- **Tags**: `monitoring`, `tracing`, `jaeger`

### 🔍 Search Engines

#### Meilisearch
- **Template**: `meilisearch`
- **Image**: [getmeili/meilisearch:v1.5](https://hub.docker.com/r/getmeili/meilisearch)
- **Description**: Fast, typo-tolerant search engine
- **Ports**: 7700
- **Configuration**: Analytics disabled for privacy
- **Tags**: `search`, `meilisearch`

### 📦 Storage

#### MinIO
- **Template**: `minio`
- **Image**: [minio/minio:latest](https://hub.docker.com/r/minio/minio)
- **Description**: S3-compatible object storage
- **Ports**: 9000 (API), 9001 (Console)
- **Default Credentials**: admin/password123
- **Volume**: Object data persistence
- **Tags**: `storage`, `s3`, `minio`

### 🛠️ Development Tools

#### MailHog
- **Template**: `mailhog`
- **Image**: [mailhog/mailhog:v1.0.1](https://hub.docker.com/r/mailhog/mailhog)
- **Description**: Email testing tool for development
- **Ports**: 1025 (SMTP), 8025 (Web UI)
- **Use Case**: Capture and inspect outgoing emails during development
- **Tags**: `email`, `testing`, `mailhog`

## Template Features

### Interactive Variables
Templates marked with ✅ support interactive configuration during `nizam add <template>`. This allows you to customize:
- **Credentials**: Usernames, passwords, API keys
- **Network**: Port mappings and exposures
- **Storage**: Volume names and persistence
- **Versions**: Specific image tags and variants

### Health Checks
Many templates include built-in health checks that integrate with:
- Docker's native healthcheck system
- Nizam's `health` and `health-server` commands
- Service readiness verification during startup

### Default Configurations
All templates provide sensible defaults for:
- **Development use**: Optimized for local development workflows
- **Security**: Basic authentication with changeable defaults
- **Networking**: Standard ports with conflict resolution
- **Persistence**: Data volumes for stateful services

## Usage Examples

### Quick Start (Default Values)
```bash
# Use defaults for rapid development setup
nizam add postgres --defaults
nizam add redis --defaults
nizam add clickhouse --defaults
```

### Interactive Configuration
```bash
# Customize settings during addition
nizam add mysql          # Prompts for username, password, port, etc.
nizam add rabbitmq       # Configure AMQP and management ports
nizam add clickhouse     # Set up analytics database with custom ports
```

### Template Discovery
```bash
# Browse all available templates
nizam templates

# Filter by category
nizam templates --tag database
nizam templates --tag monitoring
nizam templates --tag messaging

# View all available tags
nizam templates --show-tags
```

### Service Management
```bash
# Start services from templates
nizam up postgres redis prometheus

# Check health status
nizam health
nizam health postgres

# View logs and execute commands
nizam logs clickhouse
nizam exec mysql mysql -u root -p
```

## Template Architecture

### File Structure
- `templates.go` - Main template definitions and management functions
- `templates_test.go` - Comprehensive template validation tests for all templates
- `README.md` - This documentation file

### Template Definition
Each template is defined with:
```go
type Template struct {
    Name        string         // Template identifier
    Description string         // Human-readable description  
    Tags        []string       // Category and search tags
    Service     config.Service // Docker service configuration
    Variables   []Variable     // Interactive customization options
}
```

### Variable Types
Templates support various variable types:
- **string** - Text values (usernames, names)
- **port** - Network ports with validation (1-65535)
- **int** - Numeric values
- **bool** - Boolean flags

### Health Check Integration
Templates can define health checks that integrate with Docker and nizam's monitoring:
```go
HealthCheck: &config.HealthCheck{
    Test:     []string{"CMD", "pg_isready", "-U", "{{.DB_USER}}"},
    Interval: "30s",
    Timeout:  "10s", 
    Retries:  3,
}
```

## Adding New Templates

### Template Requirements
1. **Unique name** - Must not conflict with existing templates
2. **Clear description** - Explain the service's purpose
3. **Appropriate tags** - Use existing tags when possible
4. **Docker image** - Use official or well-maintained images
5. **Health checks** - Include when possible for better monitoring
6. **Documentation** - Add to this README

### Variable Guidelines
- Use **descriptive names** (DB_USER vs USER)
- Provide **sensible defaults** for development
- Mark **required variables** appropriately
- Add **validation patterns** for structured data (ports, emails)
- Include **clear descriptions** explaining each variable's purpose

### Testing
All templates should include comprehensive tests:
- Template structure validation
- Variable configuration testing
- Default value processing
- Tag filtering verification

See `templates_test.go` for examples of comprehensive template testing across all templates.

## Contributing

When adding new templates:
1. Define the template in `templates.go`
2. Add comprehensive tests
3. Update this README
4. Consider adding detailed documentation for complex services
5. Test with both default and interactive configurations

## Links

### Documentation
- [Main README](../../README.md) - Full project documentation
- [ClickHouse Guide](../../docs/CLICKHOUSE.md) - Detailed ClickHouse usage
- [Commands Reference](../../docs/COMMANDS.md) - All nizam commands

### Docker Images
All templates use official or widely-trusted Docker images. Links to Docker Hub repositories are provided in each template section above for reference and version information.
