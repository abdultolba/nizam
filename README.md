# nizam 🛠️

> Local Structured Service Manager for Dev Environments

**nizam** is a powerful CLI tool to manage, monitor, and interact with local development services (Postgres, Redis, Meilisearch, etc.) using Docker. It helps you spin up, shut down, and interact with common services without manually writing `docker run` or service-specific commands.

## Features

- 🚀 **One-command service management**: `nizam up postgres redis`
- 📊 **Service monitoring**: `nizam status` shows health of all services
- 📝 **Log tailing**: `nizam logs redis` to debug issues
- 💻 **Direct service interaction**: `nizam exec postgres psql -U user`
- ⚙️ **Profile support**: Multiple configurations for `dev`, `test`, `ci`
- 🐳 **Docker-native**: Uses Docker containers with sensible defaults

## Quick Start

```bash
# Initialize a new nizam config
nizam init

# Browse available service templates
nizam templates

# Add services from templates
nizam add mysql
nizam add redis --name cache

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
# List all available templates
nizam templates

# Filter templates by category
nizam templates --tag database

# Add a service from a template
nizam add postgres
nizam add redis --name cache
```

## Development Status

🚧 **This project is in active development**

- [x] Project structure
- [x] Core CLI commands (`init`, `up`, `down`, `status`, `logs`, `exec`)
- [x] Docker integration
- [x] Config file parsing
- [x] Service definitions
- [x] Basic health checking
- [x] Log streaming
- [x] Service templates (16 built-in templates)
- [ ] Profile management
- [ ] Advanced health checks
- [ ] Network management

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

<!-- ## License

MIT License - see [LICENSE](LICENSE) file for details. -->
