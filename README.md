# nizam 🛠️

> Local Structured Service Manager for Dev Environments

**nizam** is a powerful CLI tool to manage, monitor, and interact with local development services (Postgres, Redis, Meilisearch, etc.) using Docker. It helps you spin up, shut down, and interact with common services without manually writing `docker run` or service-specific commands.

## Features

- 🚀 **One-command service management**: `nizam up postgres redis`
- 🖥️ **Interactive TUI**: Full-featured terminal interface for visual service management
- 🎛️ **Interactive template configuration**: Customize ports, credentials, and settings
- 📊 **Service monitoring**: `nizam status` shows health of all services
- 🏥 **Advanced Health Checks**: Built-in health monitoring with HTTP server and web dashboard
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

## Development Status

🚧 **This project is in active development**

- [x] Project structure
- [x] Core CLI commands (`init`, `up`, `down`, `status`, `logs`, `exec`, `add`, `remove`)
- [x] Docker integration
- [x] Config file parsing
- [x] Service definitions
- [x] Basic health checking
- [x] **Advanced Health Check System**: Comprehensive health monitoring with multiple interfaces
  - [x] Health check engine with command, HTTP, and Docker status checks
  - [x] CLI health commands (`nizam health`, `nizam health-server`)
  - [x] HTTP server with REST API endpoints and web dashboard
  - [x] Built-in health checks for common services (PostgreSQL, MySQL, Redis)
  - [x] Health check history tracking and real-time monitoring
  - [x] Multiple output formats (table, JSON, compact) and watch mode
  - [x] Docker native healthcheck integration during container creation
- [x] Log streaming
- [x] Service templates (16 built-in templates)
- [x] Interactive template variables (postgres, mysql, redis, mongodb, rabbitmq)
- [x] Custom user templates (export, import, manage)
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
- [ ] Profile management
- [ ] Network management

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

<!-- ## License

MIT License - see [LICENSE](LICENSE) file for details. -->
