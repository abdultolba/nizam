# nizam 🛠️

> Local Structured Service Manager for Dev Environments

**nizam** is a powerful CLI tool to manage, monitor, and interact with local development services (Postgres, Redis, Meilisearch, etc.) using Docker. It helps you spin up, shut down, and interact with common services without manually writing `docker run` or service-specific commands.

## Features

### Core Service Management
- 🚀 **One-command service management**: `nizam up postgres redis`
- 🎛️ **Interactive template configuration**: Customize ports, credentials, and settings
- 📊 **Service monitoring**: `nizam status` shows health of all services
- 📝 **Log tailing**: `nizam logs redis` to debug issues
- 💻 **Direct service interaction**: `nizam exec postgres psql -U user`
- 🐳 **Docker-native**: Uses Docker containers with sensible defaults

### Development & Operations Tools
- 🩺 **Environment Doctor**: Comprehensive preflight checks with `nizam doctor`
- 🔍 **Configuration Linting**: Best practices validation with `nizam lint`
- ✅ **Config Validation**: Syntax and structure validation with `nizam validate`
- ⚡ **Retry Operations**: Exponential backoff retry for failed operations
- 🕒 **Service Readiness**: Wait for services with `nizam wait-for`
- 🔄 **Self-Update**: Automatic updates from GitHub releases
- 🧩 **Shell Completion**: Multi-shell completion support

### Interactive Interfaces  
- 🖥️ **Interactive TUI**: Full-featured terminal interface for visual service management
- 🏥 **Advanced Health Checks**: Built-in health monitoring with HTTP server and web dashboard
- ⚙️ **Profile support**: Multiple configurations for `dev`, `test`, `ci`

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

## Interactive TUI (Terminal User Interface)

nizam includes a beautiful, cyberpunk-themed terminal interface for visual service management. The TUI provides an immersive experience with real-time monitoring and full operational capabilities.

### Launching the TUI

```bash
# Launch the enhanced TUI (default - with real Docker operations)
nizam tui

# Launch demo mode (for exploration without Docker operations)
nizam tui --demo

# Enable debug mode
nizam tui --debug
```

### TUI Features

🎨 **Tron-Inspired Design**
- Cyberpunk aesthetic with cyan, blue, purple, and pink accents
- Animated status indicators and smooth transitions
- ASCII art logo and professional layout

⚡ **Real Docker Operations**
- Start, stop, restart services directly from the interface
- Add new services from templates with interactive prompts
- Remove services with safety confirmations
- Live monitoring with auto-refresh every 30 seconds
- Viewport scrolling controls (Ctrl+U/D/B/F) for all views
- Config view caching for stable display (5-second intervals)

🖥️ **Multiple Views**
- **Dashboard (1)**: Service overview with quick actions
- **Services (2)**: Detailed service management table
- **Logs (3)**: Real-time log streaming and filtering
- **Templates (4)**: Browse and add services from templates
- **Config (5)**: Live configuration viewing and management

### Navigation & Controls

#### Global Navigation
```
1-5            Switch between views
Tab/Shift+Tab  Navigate panels/buttons
h or ?         Toggle help screen
r              Refresh services (live data)
q or Ctrl+C    Quit application
/              Search services/templates
Esc            Clear search or go back

# Viewport Scrolling (works in all views)
Ctrl+U         Scroll up (5 lines)
Ctrl+D         Scroll down (5 lines)
Ctrl+B         Page up (full screen)
Ctrl+F         Page down (full screen)
```

#### Dashboard View
```
Tab/Shift+Tab  Navigate between quick action buttons
Enter/Space    Execute selected quick action:
               • Start All Services
               • Stop All Services  
               • Refresh Data
               • Add New Service (goes to Templates)

# Viewport Controls
Ctrl+U/D       Scroll service list up/down
Ctrl+B/F       Page up/down in service list
```

#### Services View
```
↑/↓ or j/k     Navigate service list
s              Start selected service
x              Stop selected service
R              Restart selected service
d or Delete    Remove selected service (with confirmation)
Enter          View logs for selected service

# Viewport Controls
Ctrl+U/D       Scroll service list up/down
Ctrl+B/F       Page up/down in service list
```

#### Logs View
```
↑/↓ or j/k     Select service for log viewing
Enter          Start/stop real-time log streaming
c              Clear current logs
f              Filter logs (search within log content)

# Viewport Controls  
Ctrl+U/D       Scroll log content up/down
Ctrl+B/F       Page up/down in log content
```

#### Templates View
```
↑/↓ or j/k     Navigate template list
Enter or a     Add service from selected template
               (opens interactive prompt for service name)

# Viewport Controls
Ctrl+U/D       Scroll template list up/down
Ctrl+B/F       Page up/down in template list
```

### TUI Screenshots

The TUI features a distinctive Tron-themed interface:

```
███╗   ██╗██╗███████╗ █████╗ ███╗   ███╗
████╗  ██║██║╚══███╔╝██╔══██╗████╗ ████║
██╔██╗ ██║██║  ███╔╝ ███████║██╔████╔██║
██║╚██╗██║██║ ███╔╝  ██╔══██║██║╚██╔╝██║
██║ ╚████║██║███████╗██║  ██║██║ ╚═╝ ██║
╚═╝  ╚═══╝╚═╝╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝

Enhanced Service Manager - Full Docker Operations

┌─ Dashboard (1) ─┬─ Services (2) ─┬─ Logs (3) ─┬─ Templates (4) ─┬─ Config (5) ─┐
```

### Safety Features

The TUI includes built-in safety mechanisms:

- **Confirmation Dialogs**: All destructive operations require confirmation
- **Input Validation**: Service names and parameters are validated in real-time
- **Error Handling**: Clear error messages with helpful suggestions
- **Graceful Fallback**: Demo mode available if Docker is unavailable

### TUI vs CLI

| Feature | CLI Commands | Enhanced TUI |
|---------|-------------|--------------|
| **Service Operations** | `nizam up/down/restart` | ✅ Direct interface operations |
| **Real-time Monitoring** | `nizam status` (snapshot) | ✅ Live updates every 30s |
| **Log Viewing** | `nizam logs <service>` | ✅ Interactive log streaming |
| **Service Creation** | `nizam add <template>` | ✅ Visual template browser |
| **Configuration** | Edit `.nizam.yaml` | ✅ Live config viewing |
| **Batch Operations** | Multiple commands | ✅ Single interface |
| **Learning Curve** | Command memorization | ✅ Visual guidance |

**When to use TUI:**
- 🎮 Prefer visual interface over command line
- 🔄 Need real-time monitoring
- 🚀 Want one-stop service management
- 📚 Learning nizam features
- 🎯 Managing multiple services frequently

**When to use CLI:**
- 🤖 Scripting and automation
- 🚀 Single, quick operations
- 📱 Working in constrained terminals
- 🔧 Integration with other tools

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

| Service | Health Check | Command |
|---------|-------------|----------|
| **PostgreSQL** | `pg_isready` | Database connection test |
| **MySQL** | `mysqladmin ping` | Database ping test |
| **Redis** | `redis-cli ping` | Redis ping command |
| **MongoDB** | `mongosh --eval` | Database status check |
| **Elasticsearch** | HTTP health API | `GET /_health` endpoint |

### Health Status Types

- 🟢 **healthy**: Service is running and responding correctly
- 🔴 **unhealthy**: Service is running but health check failed
- 🟡 **starting**: Service is starting up (within start_period)
- ⚫ **not_running**: Docker container is not running
- 🟣 **unknown**: Health check status could not be determined

### Integration with TUI

The health check system integrates with the TUI for visual monitoring:

- Health status indicators in service listings
- Real-time health updates in dashboard view
- Health check history in service details
- Manual health check triggers from interface

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
  Change host port for service postgres
  Or run: nizam up --resolve-ports
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

### Interactive Interfaces ✅
- [x] **Interactive TUI**: Full-featured terminal interface with real Docker operations
  - [x] Tron-themed cyberpunk design with animated status indicators
  - [x] Live service monitoring with auto-refresh
  - [x] Direct service operations (start, stop, restart, remove)
  - [x] Interactive service creation from templates
  - [x] Real-time log streaming and filtering
  - [x] Safety confirmations and input validation
  - [x] Search and filtering capabilities
  - [x] Demo mode for exploration without Docker
  - [x] Viewport scrolling controls (Ctrl+U/D/B/F) for all views
  - [x] Config view caching to prevent rapid refreshing (5-second intervals)

### Documentation & Examples ✅
- [x] Comprehensive README with feature documentation
- [x] CLI commands documentation (`docs/COMMANDS.md`)
- [x] Module-specific documentation (`internal/doctor/README.md`, `internal/lint/README.md`)
- [x] Implementation status tracking (`FEATURE_IMPLEMENTATION.md`)
- [x] Usage examples and integration patterns

### Planned Features 🔄
- [ ] **Profile Management**: Multi-environment configuration support
- [ ] **Network Management**: Custom Docker network creation and management
- [ ] **Plugin System**: Extensible architecture for third-party integrations
- [ ] **Backup & Restore**: Service data backup and restoration
- [ ] **Performance Monitoring**: Resource usage tracking and optimization
- [ ] **Secret Management**: Secure credential handling and rotation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

<!-- ## License

MIT License - see [LICENSE](LICENSE) file for details. -->
