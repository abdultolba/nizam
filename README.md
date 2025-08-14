<p align="center">
  <img src="https://res.cloudinary.com/friendly-social/image/upload/v1754947220/nizam_logo_blue_kyt9ck.png" alt="Nizam Logo" width="200" style="margin-bottom:-60px;"><br>
  <b>nizam</b> is a powerful CLI tool to manage, monitor, and interact with local development services (Postgres, Redis, Meilisearch, etc.) using Docker. It helps you spin up, shut down, and interact with common services without manually writing <code>docker run</code> or service-specific commands.
</p>

## Features

### Core Service Management

- 🚀 **One-command service management**: `nizam up postgres redis`
- 🎛️ **Interactive template configuration**: Customize ports, credentials, and settings
- 📊 **Service monitoring**: `nizam status` shows health of all services
- 📝 **Log tailing**: `nizam logs redis` to debug issues
- 💻 **Direct service interaction**: `nizam exec postgres psql -U user`
- 🐳 **Docker-native**: Uses Docker containers with sensible defaults

### Data Lifecycle Management

- 📸 **Database Snapshots**: Create, restore, list, and prune database snapshots with `nizam snapshot`
  - **Multi-engine support**: PostgreSQL, MySQL, and Redis (MongoDB planned)
  - **Compression options**: zstd (default), gzip, or none
  - **Atomic operations**: Safe snapshot creation and restoration
  - **Metadata tracking**: Tagged snapshots with notes and checksums
- 🔗 **One-liner Database Access**: Smart CLI tools with auto-resolved connections
  - `nizam psql [service]` - Connect to PostgreSQL with resolved credentials
  - `nizam mysql [service]` - Connect to MySQL with auto-resolved credentials
  - `nizam redis-cli [service]` - Connect to Redis with auto-configuration
  - `nizam mongosh [service]` - Connect to MongoDB with auto-configuration
  - **Fallback execution**: Uses host binaries or container execution automatically

### Development & Operations Tools

- 🩺 **Environment Doctor**: Comprehensive preflight checks with `nizam doctor`
- 🔍 **Configuration Linting**: Best practices validation with `nizam lint`
- ✅ **Config Validation**: Syntax and structure validation with `nizam validate`
- ⚡ **Retry Operations**: Exponential backoff retry for failed operations
- 🕒 **Service Readiness**: Wait for services with `nizam wait-for`
- 🔄 **Self-Update**: Automatic updates from GitHub releases
- 🧩 **Shell Completion**: Multi-shell completion support

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

nizam includes 17+ built-in service templates for popular development tools, with comprehensive configurations, interactive variables, health checks, and organized documentation.

**Databases:**

- `postgres` / `postgres-15` - PostgreSQL database
- `mysql` - MySQL database
- `mongodb` - MongoDB document database
- `redis` / `redis-stack` - Redis cache and data store
- `clickhouse` - ClickHouse OLAP database for analytics
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

For detailed template documentation, configurations, and contribution guidelines, see [`internal/templates/README.md`](internal/templates/README.md).

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

## Health Check System 🏥

nizam includes a comprehensive health check system that monitors your services through multiple check types and provides both CLI and web-based interfaces for monitoring.

### Health Check Features

- 🔍 **Multiple Check Types**: Command execution, HTTP requests, and Docker status checks
- 📊 **Built-in Templates**: Pre-configured health checks for common services (PostgreSQL, MySQL, Redis)
- 🖥️ **CLI Monitoring**: Query health status with multiple output formats
- 🌐 **HTTP Server & Dashboard**: Web-based monitoring with REST API
- 📈 **Health History**: Track health check results over time
- ⚡ **Real-time Updates**: Live monitoring with configurable intervals
- 🎯 **Per-service Status**: Individual service health tracking and management

### Quick Health Check Examples

```bash
# Check health of all services (table format)
nizam health

# Check specific service health
nizam health postgres

# Output in JSON format
nizam health --output json

# Watch health status continuously
nizam health --watch

# Watch with custom interval (5 seconds)
nizam health --watch --interval 5

# Compact status display
nizam health --output compact
```

### Health Check CLI Commands

#### `nizam health` - Health Status Query

```bash
# Usage patterns
nizam health [service] [flags]

# Examples
nizam health                    # All services, table format
nizam health postgres          # Specific service
nizam health --output json     # JSON output
nizam health --watch           # Continuous monitoring
nizam health --watch --interval 5  # Custom watch interval

# Available flags
-o, --output string   Output format (table, json, compact)
-w, --watch           Watch health status continuously
    --interval int    Watch interval in seconds (default 10)
```

**Output Formats:**

- **table**: Formatted table with service details, status, and timestamps
- **json**: Complete health data in JSON format for automation
- **compact**: Minimal status display with emoji indicators

#### `nizam health-server` - HTTP Health Monitor

```bash
# Launch health monitoring server
nizam health-server [flags]

# Examples
nizam health-server                      # Start on :8080
nizam health-server --address :9090     # Custom port
nizam health-server --interval 15       # 15-second check interval
nizam health-server --no-auto-start     # Manual health check start

# Available flags
    --address string   HTTP server address (default ":8080")
    --interval int     Health check interval in seconds (default 30)
    --auto-start       Auto-start health checking (default true)
```

### HTTP API Endpoints

The health server provides REST API endpoints for integration:

```bash
# Get overall health summary
GET /api/health

# Get specific service health
GET /api/services/{service}

# Trigger immediate health check
POST /api/check/{service}

# Get all services health status
GET /api/services
```

**Example API Response:**

```json
{
  "service": "postgres",
  "status": "healthy",
  "is_running": true,
  "container_name": "nizam-postgres",
  "image": "postgres:16",
  "last_check": "2024-08-08T03:45:30Z",
  "check_history": [
    {
      "status": "healthy",
      "message": "pg_isready check passed",
      "timestamp": "2024-08-08T03:45:30Z",
      "duration": "12ms"
    }
  ]
}
```

### Web Dashboard

Access the web dashboard at `http://localhost:8080` when running the health server:

- 📊 **Live Status Overview**: Real-time service health monitoring
- 🔄 **Auto-refresh**: Configurable automatic status updates
- 🎯 **Manual Triggers**: On-demand health check execution
- 📈 **Health History**: Visual timeline of health check results
- 🎨 **Responsive UI**: Clean, modern interface with status indicators

### Health Check Configuration

Services can include health check configurations in their templates:

```yaml
# Example service with health checks
services:
  postgres:
    image: postgres:16
    ports:
      - 5432:5432
    env:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "user"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

**Health Check Types:**

1. **Command Checks**: Execute commands inside containers

   ```yaml
   test: ["CMD", "pg_isready", "-U", "user"]
   test: ["CMD-SHELL", "curl -f http://localhost:8080/health"]
   ```

2. **HTTP Checks**: Automatically detected from curl/wget commands

   ```yaml
   test: ["CMD", "curl", "-f", "http://localhost:3000/health"]
   ```

3. **Docker Status**: Default fallback using container running status

### Built-in Health Checks

Common service templates include pre-configured health checks:

| Service           | Health Check      | Command                  |
| ----------------- | ----------------- | ------------------------ |
| **PostgreSQL**    | `pg_isready`      | Database connection test |
| **MySQL**         | `mysqladmin ping` | Database ping test       |
| **Redis**         | `redis-cli ping`  | Redis ping command       |
| **MongoDB**       | `mongosh --eval`  | Database status check    |
| **Elasticsearch** | HTTP health API   | `GET /_health` endpoint  |

### Health Status Types

- 🟢 **healthy**: Service is running and responding correctly
- 🔴 **unhealthy**: Service is running but health check failed
- 🟡 **starting**: Service is starting up (within start_period)
- ⚫ **not_running**: Docker container is not running
- 🟣 **unknown**: Health check status could not be determined

### Use Cases

**Development Workflow:**

```bash
# Start services
nizam up postgres redis

# Monitor health during startup
nizam health --watch

# Check specific service issues
nizam health postgres

# Launch web dashboard for team monitoring
nizam health-server --address :8080
```

**CI/CD Integration:**

```bash
# Wait for services to be healthy
nizam health --output json | jq '.status == "healthy"'

# Automated health monitoring
nizam health-server --no-auto-start &
curl http://localhost:8080/api/health
```

**Team Monitoring:**

```bash
# Shared health dashboard
nizam health-server --address :3030

# Team members access: http://dev-server:3030
```

## Data Lifecycle Management 📸

nizam provides comprehensive data lifecycle tools for database snapshots and one-liner database access, making it easy to capture, restore, and work with database states during development.

### Database Snapshots

Create point-in-time snapshots of your databases for backup, testing, or sharing data states.

#### Snapshot Features

- 🎯 **Multi-engine support**: PostgreSQL, MySQL, and Redis (MongoDB planned)
- 🗜️ **Smart compression**: zstd (default), gzip, or none with automatic streaming
- 🔒 **Data integrity**: SHA256 checksums for all snapshot files
- 📋 **Rich metadata**: Tagged snapshots with notes, timestamps, and version tracking
- 📁 **Organized storage**: Structured storage in `.nizam/snapshots/<service>/`
- ⚡ **Atomic operations**: Safe creation and restoration with temporary files

#### Quick Snapshot Examples

```bash
# Create a snapshot with automatic timestamping
nizam snapshot create postgres

# Create a tagged snapshot with notes
nizam snapshot create postgres --tag "before-migration" --note "Pre-schema update"

# Create with different compression
nizam snapshot create redis --compress gzip

# List all snapshots
nizam snapshot list

# List snapshots for specific service
nizam snapshot list postgres

# Restore latest snapshot
nizam snapshot restore postgres --latest

# Restore specific tagged snapshot
nizam snapshot restore postgres --tag "before-migration"

# Clean up old snapshots (keep 5 most recent)
nizam snapshot prune postgres --keep 5
```

#### Snapshot Commands

**`nizam snapshot create <service>`**

```bash
# Basic snapshot creation
nizam snapshot create postgres
nizam snapshot create mysql
nizam snapshot create redis

# With custom options
nizam snapshot create postgres --tag "v1.2.0" --compress zstd --note "Release snapshot"

# Available flags:
    --compress string   Compression type: zstd, gzip, none (default "zstd")
    --note string      Note/description for the snapshot
    --tag string       Tag for the snapshot (default: timestamp)
```

**`nizam snapshot list [service]`**

```bash
# List all snapshots across all services
nizam snapshot list

# List snapshots for specific service
nizam snapshot list postgres

# JSON output for automation
nizam snapshot list --json
```

**`nizam snapshot restore <service>`**

```bash
# Restore latest snapshot
nizam snapshot restore postgres --latest

# Restore specific tagged snapshot
nizam snapshot restore postgres --tag "before-migration"

# Available flags:
    --force          Skip confirmation prompts
    --latest         Restore the most recent snapshot
    --tag string     Restore snapshot with specific tag
```

**`nizam snapshot prune <service>`**

```bash
# Remove old snapshots, keeping 3 most recent
nizam snapshot prune postgres --keep 3

# Dry run to see what would be deleted
nizam snapshot prune postgres --keep 5 --dry-run

# Available flags:
    --dry-run        Show what would be deleted without actually deleting
    --keep int       Number of snapshots to keep (required)
```

### One-liner Database Access

Connect to your databases instantly with auto-resolved connection parameters.

#### Features

- 🔧 **Auto-resolution**: Automatically discovers connection details from configuration
- 🔄 **Smart fallback**: Uses host binaries when available, falls back to container execution
- 🎯 **Service detection**: Auto-detects the first service of each type if not specified
- 📋 **Pass-through args**: All arguments after `--` are passed directly to the database CLI

#### PostgreSQL Access

**`nizam psql [service]`**

```bash
# Connect to first/default PostgreSQL service
nizam psql

# Connect to specific service
nizam psql postgres
nizam psql api-db

# Override connection parameters
nizam psql --user admin --db production

# Pass arguments to psql
nizam psql -- --help
nizam psql -- -c "SELECT version()"
nizam psql postgres -- -c "\\l"

# Available flags:
    --db string       Database name (override config)
    --user string     Username (override config)
```

#### MySQL Access

**`nizam mysql [service]`**

```bash
# Connect to first/default MySQL service
nizam mysql

# Connect to specific service
nizam mysql mysql
nizam mysql api-db

# Override connection parameters
nizam mysql --user root --db mysql

# Pass arguments to mysql client
nizam mysql -- --help
nizam mysql -- -e "SHOW DATABASES"
nizam mysql api-db -- -e "SELECT version()"
```

**Available flags:**

- `--db string` - Database name (override config)
- `--user string` - Username (override config)

#### Redis Access

**`nizam redis-cli [service]`**

```bash
# Connect to first/default Redis service
nizam redis-cli

# Connect to specific service
nizam redis-cli redis
nizam redis-cli cache

# Pass arguments to redis-cli
nizam redis-cli -- --help
nizam redis-cli -- ping
nizam redis-cli cache -- info server
```

#### Connection Resolution

The one-liner commands automatically resolve connection details from your configuration:

1. **Service Discovery**: If no service specified, uses the first service of matching type
2. **Credential Extraction**: Pulls username, password, database, and port from service environment
3. **Host Binary Detection**: Checks if `psql`, `redis-cli`, etc. are available on the host
4. **Fallback Execution**: Uses `docker exec` if host binaries are not found
5. **Connection String Building**: Constructs proper connection URLs with credentials

#### vs. Raw Container Execution

**Key Difference**: `nizam psql` is a **smart database client** that auto-resolves connections, while `nizam exec postgres psql` is **raw container command execution**.

| Feature                   | `nizam psql`                     | `nizam exec postgres psql`         |
| ------------------------- | -------------------------------- | ---------------------------------- |
| **Credential resolution** | ✅ Automatic from config         | ❌ Manual specification required   |
| **Connection strings**    | ✅ Auto-built URLs               | ❌ Manual argument construction    |
| **Host binary usage**     | ✅ Uses host `psql` if available | ❌ Always executes in container    |
| **Service discovery**     | ✅ Auto-finds PostgreSQL service | ❌ Must specify exact service name |
| **Ease of use**           | 🟢 Just works                    | 🟡 Requires connection knowledge   |

**Examples:**

```bash
# Smart connection (auto-resolves everything)
nizam psql                           # Connects automatically
nizam psql -- -c "SELECT version()"   # Runs query with auto-connection
nizam mysql                          # Connects to MySQL automatically
nizam mysql -- -e "SHOW DATABASES"    # Runs MySQL query with auto-connection

# Raw container execution (manual specification required)
nizam exec postgres psql -U user -d mydb -h localhost
nizam exec mysql mysql -u user -pmypass mydb
```

**Example Resolution:**

```yaml
# .nizam.yaml
services:
  postgres:
    image: postgres:16
    ports: ["5432:5432"]
    env:
      POSTGRES_USER: myuser
      POSTGRES_PASSWORD: mypass
      POSTGRES_DB: mydb
```

```bash
# This command:
nizam psql

# Resolves to:
psql "postgresql://myuser:mypass@localhost:5432/mydb?sslmode=disable"

# Or if psql not on host:
docker exec -it nizam_postgres psql -U myuser -d mydb

# For MySQL:
nizam mysql

# Resolves to:
mysql -h localhost -P 3306 -u myuser -pmypass mydb

# Or if mysql not on host:
docker exec -it nizam_mysql mysql -u myuser -h localhost -pmypass mydb
```

## Development & Operations Tools 🛠️

nizam includes comprehensive tooling for development workflow optimization, environment validation, and operational reliability.

### Environment Doctor (`nizam doctor`)

Comprehensive preflight environment checks to ensure your Docker setup is ready for development.

```bash
# Run all environment checks
nizam doctor

# Skip specific checks
nizam doctor --skip net.mtu,disk.free

# JSON output for CI/CD integration
nizam doctor --json

# Attempt automatic fixes
nizam doctor --fix

# Verbose output with detailed diagnostics
nizam doctor --verbose
```

**Checks Performed:**

- 🐳 **Docker connectivity** - Verify Docker daemon is running
- 🔧 **Docker Compose** - Ensure compose plugin is available
- 💾 **Disk space** - Check available storage (warns if <1GB)
- 🌐 **Network MTU** - Validate network configuration
- 🚪 **Port conflicts** - Dynamic port availability for all configured services

**Sample Output:**

```
✔ docker.daemon       Docker daemon responding
✔ docker.compose      Docker Compose plugin available
! net.mtu              non-standard MTU detected
  VPNs may lower MTU; if Docker networking is flaky, align MTU in daemon.json
✖ port.5432            port in use
  Change host port for service postgres in .nizam.yaml
  Or stop the process using the port
```

### Configuration Validation (`nizam validate`)

Validate configuration file syntax and structure before deployment.

```bash
# Validate default configuration
nizam validate

# Validate specific file
nizam validate --file ./production.yaml

# JSON output for automation
nizam validate --json

# Strict mode (exit non-zero on any issues)
nizam validate --strict
```

**Validation Features:**

- ✅ YAML syntax validation
- 🔍 Service structure verification
- 📋 Required field checking
- 🔢 Profile validation
- 📊 Multiple output formats

### Configuration Linting (`nizam lint`)

Analyze configurations for best practices and potential issues.

```bash
# Lint default configuration
nizam lint

# Lint specific file
nizam lint --file ./config.yaml

# JSON output for CI/CD pipelines
nizam lint --json
```

**Linting Rules:**

- 🚫 **no-latest**: Prevents `:latest` image tags (reproducibility)
- 🔌 **ports-shape**: Validates port mapping format
- ⚡ **limits**: Recommends resource limits for consistency

**Sample Output:**

```
✖ services.web.image: image tag missing or ':latest' not allowed (no-latest)
  Fix: pin to a specific tag, e.g. 'nginx:1.21'

! services.database: consider setting CPU/memory limits (limits)
  Fix: add 'resources: { cpus: "1.0", memory: "512m" }'
```

### Service Readiness (`nizam wait-for`)

Wait for services to become ready before proceeding with dependent operations.

```bash
# Wait for specific service
nizam wait-for database

# Wait for multiple services
nizam wait-for web database cache

# Wait for all services
nizam wait-for

# Custom timeout and check interval
nizam wait-for --timeout 60s --interval 2s database
```

**Readiness Checks:**

- 🔌 **Port connectivity** - TCP connection tests
- 🌐 **HTTP health checks** - Endpoint availability
- 🐳 **Container status** - Docker container state
- ⏱️ **Configurable timeouts** - Flexible waiting strategies

### Retry Operations (`nizam retry`)

Retry failed operations with intelligent exponential backoff.

```bash
# Retry start command with defaults
nizam retry start

# Custom retry attempts and delay
nizam retry start --attempts 5 --delay 2s

# Retry specific services
nizam retry start web database --attempts 3
```

**Supported Operations:**

- 🚀 `start` - Start services with retry
- ⏹️ `stop` - Stop services with retry
- 🔄 `restart` - Restart services with retry
- 📥 `pull` - Pull images with retry
- 🏗️ `build` - Build images with retry

**Retry Features:**

- 📈 Exponential backoff (1s → 2s → 4s → 8s)
- 🎯 Configurable attempts and delays
- 📊 Progress reporting with attempt counters
- 🔄 Graceful failure handling

### Self-Update (`nizam update`)

Keep nizam up-to-date with the latest features and fixes.

```bash
# Check for updates without installing
nizam update --check

# Update to latest stable version
nizam update

# Include prerelease versions
nizam update --prerelease
```

**Update Features:**

- 🔍 GitHub releases integration
- 🖥️ Platform-specific binary detection
- 🔄 Safe binary replacement with rollback
- 🚀 Cross-platform support (Windows, macOS, Linux)
- 📦 Prerelease channel support

### Shell Completion (`nizam completion`)

Generate completion scripts for faster command-line usage.

```bash
# Bash completion
source <(nizam completion bash)

# Zsh completion
source <(nizam completion zsh)

# Fish completion
nizam completion fish | source

# PowerShell completion
nizam completion powershell | Out-String | Invoke-Expression
```

**Installation Examples:**

```bash
# Bash (add to ~/.bashrc)
echo 'source <(nizam completion bash)' >> ~/.bashrc

# Zsh (add to ~/.zshrc)
echo 'source <(nizam completion zsh)' >> ~/.zshrc
```

### Development Workflow Integration

**Pre-commit Checks:**

```bash
#!/bin/bash
# .git/hooks/pre-commit
nizam validate --strict && nizam lint && nizam doctor --json
```

**CI/CD Pipeline:**

```yaml
# .github/workflows/validate.yml
- name: Validate nizam configuration
  run: |
    nizam doctor --json
    nizam validate --strict
    nizam lint --json
```

**Development Environment Setup:**

```bash
# Reliable environment startup
nizam doctor                    # Check environment
nizam validate                  # Validate config
nizam retry start --attempts 3  # Start with retry
nizam wait-for --timeout 60s    # Wait for readiness
```

**Production Deployment:**

```bash
# Production-ready checks
nizam lint --json > lint-report.json
nizam validate --strict --file production.yaml
nizam doctor --fix
```

## Development Status

🚧 **This project is in active development**

### Core Infrastructure ✅

- [x] Project structure and modern Go standards
- [x] Core CLI commands (`init`, `up`, `down`, `status`, `logs`, `exec`, `add`, `remove`)
- [x] Docker integration with Compose support
- [x] Configuration file parsing and validation
- [x] Service definition system

### Service Management ✅

- [x] Service templates (16+ built-in templates)
- [x] Interactive template variables (postgres, mysql, redis, mongodb, rabbitmq)
- [x] Custom user templates (export, import, manage)
- [x] Log streaming and real-time monitoring

### Health & Monitoring ✅

- [x] **Advanced Health Check System**: Comprehensive health monitoring with multiple interfaces
  - [x] Health check engine with command, HTTP, and Docker status checks
  - [x] CLI health commands (`nizam health`, `nizam health-server`)
  - [x] HTTP server with REST API endpoints and web dashboard
  - [x] Built-in health checks for common services (PostgreSQL, MySQL, Redis)
  - [x] Health check history tracking and real-time monitoring
  - [x] Multiple output formats (table, JSON, compact) and watch mode
  - [x] Docker native healthcheck integration during container creation

### Development & Operations Tools ✅

- [x] **Environment Doctor** (`nizam doctor`): Comprehensive preflight checks
  - [x] Docker daemon and Compose plugin verification
  - [x] System resource checks (disk space, network MTU)
  - [x] Dynamic port conflict detection
  - [x] JSON output and automatic fix attempts
  - [x] Concurrent check execution with semaphores
- [x] **Configuration Validation** (`nizam validate`): Syntax and structure validation
  - [x] YAML parsing with detailed error reporting
  - [x] Service structure verification
  - [x] Multiple output formats and strict mode
- [x] **Configuration Linting** (`nizam lint`): Best practices enforcement
  - [x] Extensible rule framework with severity levels
  - [x] Built-in rules (no-latest, ports-shape, limits)
  - [x] JSON output for CI/CD integration
- [x] **Service Readiness** (`nizam wait-for`): Wait for service availability
  - [x] Port connectivity and HTTP health check support
  - [x] Configurable timeouts and check intervals
  - [x] Multi-service waiting with progress reporting
- [x] **Retry Operations** (`nizam retry`): Exponential backoff for failed operations
  - [x] Support for all major operations (start, stop, restart, pull, build)
  - [x] Configurable attempts and delay intervals
  - [x] Progress reporting with attempt counters
- [x] **Self-Update** (`nizam update`): Automatic updates from GitHub releases
  - [x] Platform-specific binary detection and safe replacement
  - [x] Version comparison and prerelease support
  - [x] Cross-platform compatibility (Windows, macOS, Linux)
- [x] **Shell Completion** (`nizam completion`): Multi-shell completion support
  - [x] Bash, Zsh, Fish, and PowerShell support
  - [x] Dynamic command and flag completion

### Data Lifecycle Management ✅

- [x] **Database Snapshots** (`nizam snapshot`): Complete snapshot lifecycle management
  - [x] PostgreSQL, MySQL, Redis, and MongoDB snapshot engines with streaming dumps
  - [x] Multi-compression support (zstd, gzip, none) with checksum verification
  - [x] Rich manifest system with metadata, tags, and notes
  - [x] Atomic operations with temporary files and safe renames
  - [x] Organized storage in `.nizam/snapshots/<service>/` structure
  - [x] Create, list, restore, and prune operations with comprehensive CLI
- [x] **One-liner Database Access**: Smart database CLI tools
  - [x] `nizam psql [service]` - Auto-resolved PostgreSQL connections
  - [x] `nizam mysql [service]` - Auto-resolved MySQL connections
  - [x] `nizam redis-cli [service]` - Auto-resolved Redis connections
  - [x] `nizam mongosh [service]` - Auto-resolved MongoDB connections
  - [x] Service auto-discovery and credential resolution from configuration
  - [x] Host binary detection with container execution fallback
  - [x] Pass-through argument support for native CLI tools

### Documentation & Examples ✅

- [x] Comprehensive README with feature documentation
- [x] CLI commands documentation (`docs/COMMANDS.md`)
- [x] Module-specific documentation (`internal/doctor/README.md`, `internal/lint/README.md`)
- [x] Data lifecycle specification (`.docs/data-lifecycle.md`)
- [x] Usage examples and integration patterns
- [x] Complete unit test coverage with Makefile integration

### Seed Pack System ✅

nizam includes a comprehensive seed pack system for creating, sharing, and managing reusable database datasets with rich metadata.

#### Seed Pack Features

- 🎯 **Enhanced Snapshots**: Convert snapshots into reusable seed packs with rich metadata
- 📋 **Rich Metadata**: Author, version, license, homepage, tags, use cases, and examples
- 🔍 **Discovery & Search**: Find packs by name, tags, author, or engine type
- 📦 **Versioning**: Multiple versions of the same pack with semantic versioning
- 🏗️ **Template Integration**: Templates can reference seed packs for auto-installation
- 📁 **Organized Storage**: Structured storage in `.nizam/seeds/<engine>/<pack>/<version>/`

#### Quick Seed Pack Examples

```bash
# Create a seed pack from a snapshot
nizam pack create postgres my-snapshot \
  --name "ecommerce-starter" \
  --description "Sample e-commerce database with products and users" \
  --author "Your Name" \
  --tag "ecommerce" --tag "sample-data"

# List all available seed packs
nizam pack list

# Search for specific packs
nizam pack search ecommerce
nizam pack search --tag "sample-data" --engine postgres

# Install a seed pack to a service
nizam pack install postgres ecommerce-starter
nizam pack install postgres ecommerce-starter@1.0.0

# Get detailed pack information
nizam pack info postgres ecommerce-starter

# Remove old packs
nizam pack remove postgres ecommerce-starter@1.0.0
```

#### Seed Pack Commands

**`nizam pack create <service> [snapshot-tag]`**

```bash
# Create from latest snapshot
nizam pack create postgres

# Create from specific snapshot with metadata
nizam pack create postgres my-snapshot \
  --name "blog-content" \
  --display-name "Blog Content Pack" \
  --description "Sample blog with posts, users, and comments" \
  --author "John Doe" \
  --version "1.0.0" \
  --license "MIT" \
  --homepage "https://github.com/johndoe/blog-seeds" \
  --tag "blog" --tag "cms" --tag "sample-data" \
  --use-case "Development and testing" \
  --use-case "Demo applications"

# Available flags:
    --name string           Pack name
    --display-name string   Human-readable pack name
    --description string    Pack description
    --author string         Pack author
    --version string        Pack version (default "1.0.0")
    --license string        Pack license (default "MIT")
    --homepage string       Homepage URL
    --repository string     Repository URL
    --tag strings           Tags (can be used multiple times)
    --use-case strings      Use cases (can be used multiple times)
    --force                 Overwrite existing pack
```

**`nizam pack list [engine]`**

```bash
# List all seed packs
nizam pack list

# List packs for specific engine
nizam pack list postgres
nizam pack list redis

# JSON output for automation
nizam pack list --json
```

**`nizam pack search [query]`**

```bash
# Search by name or description
nizam pack search ecommerce
nizam pack search blog

# Filter by specific criteria
nizam pack search --tag "sample-data" --engine postgres
nizam pack search --author "John Doe"
nizam pack search --engine redis --tag "cache"

# Available flags:
    --engine string     Filter by engine type
    --tag strings       Filter by tags
    --author string     Filter by author
    --json              Output in JSON format
```

**`nizam pack install <service> <pack>`**

```bash
# Install latest version
nizam pack install postgres ecommerce-starter

# Install specific version
nizam pack install postgres ecommerce-starter@1.0.0

# Preview installation
nizam pack install postgres ecommerce-starter --dry-run

# Force install even if service has data
nizam pack install postgres ecommerce-starter --force

# Available flags:
    --dry-run           Show what would be installed
    --force             Force installation even if errors occur
```

**`nizam pack info <engine> <pack>`**

```bash
# Get detailed pack information
nizam pack info postgres ecommerce-starter
nizam pack info postgres ecommerce-starter@1.0.0

# Shows:
# - Description and metadata
# - Use cases and examples
# - Dependencies and requirements
# - Installation instructions
# - Data size and record counts
```

**`nizam pack remove <engine> <pack>`**

```bash
# Remove specific version
nizam pack remove postgres ecommerce-starter --version 1.0.0

# Remove all versions
nizam pack remove postgres ecommerce-starter
```

#### Seed Pack Manifest

Each seed pack includes a comprehensive manifest with metadata:

```json
{
  "name": "ecommerce-starter",
  "displayName": "E-commerce Starter Data",
  "description": "Sample e-commerce database with products, users, and orders",
  "version": "1.0.0",
  "author": "Your Name",
  "license": "MIT",
  "homepage": "https://github.com/yourorg/ecommerce-seeds",
  "createdAt": "2024-01-15T10:30:00Z",
  "engine": "postgres",
  "images": ["postgres:16"],
  "tags": ["ecommerce", "sample-data", "starter"],
  "dataSize": 2048576,
  "recordCount": 1500,
  "compression": "zstd",
  "useCases": ["Development and testing", "Demo applications"],
  "examples": [
    {
      "title": "List all products",
      "description": "Get all products with their categories",
      "query": "SELECT p.name, p.price, c.name as category FROM products p JOIN categories c ON p.category_id = c.id;",
      "expected": "Returns product names, prices, and categories"
    }
  ],
  "dependencies": [
    {
      "name": "postgres",
      "type": "service",
      "version": "15+",
      "optional": false
    }
  ]
}
```

#### Template Integration

Templates can reference seed packs for automatic installation:

```yaml
# Template with seed pack references
seedPacks:
  - name: "ecommerce-starter"
    version: "1.0.0"
    description: "Sample e-commerce data"
    optional: false
    autoInstall: true
  - name: "test-data"
    description: "Additional test data"
    optional: true
    autoInstall: false
```

For complete seed pack documentation and examples, see [`docs/SEED_PACKS.md`](docs/SEED_PACKS.md).

### Planned Data Lifecycle Features 🔄

- [x] **MySQL Snapshots & CLI**: MySQL database snapshot and one-liner access support ✅
- [x] **MongoDB Snapshots & CLI**: MongoDB snapshot support and one-liner access ✅
- [x] **Seed Pack System**: Versioned, shareable dataset management ✅
  - [x] Local seed pack registry with versioning ✅
  - [x] Rich metadata with tags, use cases, and examples ✅
  - [x] Template integration for auto-installation ✅
  - [ ] Team/remote registry support (Git, URL-based)
  - [ ] Seed pack creation from snapshots with data masking
- [ ] **Safe Production Imports**: Data masking and sanitization
  - [ ] Built-in masking profiles (minimal-pii, full-pii, payments-safe)
  - [ ] Custom YAML-based masking rule definitions
  - [ ] Deterministic faker for consistent data transformation
- [ ] **Encryption Support**: Age-based snapshot encryption
- [ ] **S3 Integration**: Remote snapshot storage and registries

### Other Planned Features 🔄

- [ ] **Profile Management**: Multi-environment configuration support
- [ ] **Network Management**: Custom Docker network creation and management
- [ ] **Plugin System**: Extensible architecture for third-party integrations
- [ ] **Performance Monitoring**: Resource usage tracking and optimization
- [ ] **Secret Management**: Secure credential handling and rotation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details.
