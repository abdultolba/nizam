# nizam 🛠️

> Local Structured Service Manager for Dev Environments

**nizam** is a powerful CLI tool to manage, monitor, and interact with local development services (Postgres, Redis, Meilisearch, etc.) using Docker. It helps you spin up, shut down, and interact with common services without manually writing `docker run` or service-specific commands.

## Features

- 🚀 **One-command service management**: `nizam up postgres redis`
- 🎛️ **Interactive template configuration**: Customize ports, credentials, and settings
- 📊 **Service monitoring**: `nizam status` shows health of all services
- 📝 **Log tailing**: `nizam logs redis` to debug issues
- 💻 **Direct service interaction**: `nizam exec postgres psql -U user`
- ⚙️ **Profile support**: Multiple configurations for `dev`, `test`, `ci`
- 🐳 **Docker-native**: Uses Docker containers with sensible defaults

## Quick Start

```bash
# Initialize a new nizam config (default: postgres, redis, meilisearch)
nizam init

# Or initialize with custom services
nizam init --add "mysql, mongodb, prometheus"

# Browse available service templates
nizam templates

# Add services from templates
nizam add mysql
nizam add redis --name cache

# Remove services from configuration
nizam remove mysql
nizam remove redis postgres --force

# Start services
nizam up mysql cache

# Check service status
nizam status

# View logs
nizam logs mysql

# Execute commands in service containers
nizam exec mysql mysql -u user -p

# Stop all services
nizam down
```

## Installation

### From Source

```bash
git clone https://github.com/abdultolba/nizam.git
cd nizam
go build -o nizam
sudo mv nizam /usr/local/bin/
```

### Homebrew (Coming Soon)

```bash
brew install abdultolba/tap/nizam
```

## Configuration

nizam uses a `.nizam.yaml` file to define your services:

```yaml
profile: dev
services:
  postgres:
    image: postgres:16
    ports:
      - 5432:5432
    env:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volume: pgdata

  redis:
    image: redis:7
    ports:
      - 6379:6379

  meilisearch:
    image: getmeili/meilisearch
    ports:
      - 7700:7700
```

## Service Templates

nizam includes 16+ built-in service templates for popular development tools:

**Databases:**

- `postgres` / `postgres-15` - PostgreSQL database
- `mysql` - MySQL database
- `mongodb` - MongoDB document database
- `redis` / `redis-stack` - Redis cache and data store
- `elasticsearch` - Elasticsearch search engine

**Messaging & Streaming:**

- `rabbitmq` - RabbitMQ message broker
- `kafka` - Apache Kafka (via Redpanda)
- `nats` - NATS messaging system

**Monitoring & Observability:**

- `prometheus` - Prometheus metrics collection
- `grafana` - Grafana visualization
- `jaeger` - Distributed tracing

**Storage & Search:**

- `minio` - S3-compatible object storage
- `meilisearch` - Fast search engine

**Development Tools:**

- `mailhog` - Email testing

### Using Templates

```bash
# List all available templates (built-in + custom)
nizam templates

# Filter templates by category
nizam templates --tag database

# Add a service from a template
nizam add postgres
nizam add redis --name cache
```

### Interactive Template Variables

Key templates support interactive configuration of ports, credentials, and settings:

```bash
# Add with interactive prompts (PostgreSQL, MySQL, Redis, MongoDB, RabbitMQ, etc.)
nizam add postgres    # You'll be prompted for username, password, port, etc.

# Skip prompts and use default values
nizam add postgres --defaults

# Add with custom name and interactive config
nizam add mysql --name production-db
```

Interactive features include:

- Clear variable descriptions with purpose and usage
- Default value suggestions shown in brackets
- Required field indicators and type validation
- Real-time validation with helpful error messages

### Custom Templates

Create and manage your own reusable service templates:

```bash
# Export existing service as custom template
nizam export mysql --name company-mysql --description "Our standard MySQL setup"

# List custom templates only
nizam custom list

# View detailed template information
nizam custom show company-mysql

# Use custom template in another project
nizam add company-mysql

# Delete custom template
nizam custom delete company-mysql

# Show custom templates directory
nizam custom dir
```

Custom templates are stored in `~/.nizam/templates/` and can be shared between projects or with your team.

## Service Management Commands

### Initialization

```bash
# Initialize with default services (postgres, redis, meilisearch)
nizam init

# Initialize with custom services
nizam init --add postgres,mysql,redis
nizam init --add "mongodb, prometheus, mailhog"
```

The `init` command always uses default values for template variables to ensure quick setup. Use `nizam add` afterward for interactive configuration.

### Adding Services

```bash
# Add with interactive configuration
nizam add postgres

# Add with default values
nizam add mysql --defaults

# Add with custom name
nizam add redis --name cache
```

### Removing Services

```bash
# Remove single service (stops container and removes from config)
nizam remove postgres

# Remove multiple services
nizam remove redis mysql

# Remove all services
nizam remove --all

# Force removal without confirmation
nizam remove postgres --force

# Using alias
nizam rm postgres
```

The `remove` command automatically stops running Docker containers before removing services from the configuration.

## Development Status

🚧 **This project is in active development**

- [x] Project structure
- [x] Core CLI commands (`init`, `up`, `down`, `status`, `logs`, `exec`, `add`, `remove`)
- [x] Docker integration
- [x] Config file parsing
- [x] Service definitions
- [x] Basic health checking
- [x] Log streaming
- [x] Service templates (16 built-in templates)
- [x] Interactive template variables (postgres, mysql, redis, mongodb, rabbitmq)
- [x] Custom user templates (export, import, manage)
- [ ] Profile management
- [ ] Advanced health checks
- [ ] Network management

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

<!-- ## License

MIT License - see [LICENSE](LICENSE) file for details. -->
