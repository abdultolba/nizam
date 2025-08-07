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

# Start services defined in .nizam.yaml
nizam up postgres redis

# Check service status
nizam status

# View logs
nizam logs postgres

# Execute commands in service containers
nizam exec postgres psql -U user

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

## Supported Services

- PostgreSQL
- Redis
- Meilisearch
- MongoDB
- MySQL
- Elasticsearch
- RabbitMQ
- Kafka (via Redpanda)
- MinIO
- NATS

## Development Status

🚧 **This project is in active development**

- [x] Project structure
- [x] Core CLI commands
- [x] Docker integration
- [ ] Config file parsing
- [ ] Service definitions
- [ ] Health checking
- [ ] Log streaming
- [ ] Profile management

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

<!-- ## License

MIT License - see [LICENSE](LICENSE) file for details. -->
