# Nizam CLI Commands Documentation

This document provides comprehensive documentation for all nizam CLI commands, organized by category.

## Table of Contents
- [Core Operations](#core-operations)
- [Configuration Management](#configuration-management) 
- [Data Lifecycle Management](#data-lifecycle-management)
- [Health & Monitoring](#health--monitoring)
- [Development Tools](#development-tools)
- [Utility Commands](#utility-commands)

## Core Operations

### `nizam up`
Start one or more services defined in your configuration.

```bash
# Start all services
nizam up

# Start specific services
nizam up postgres redis
```

### `nizam down`
Stop all running nizam services and clean up resources.

```bash
# Stop all services
nizam down
```

### `nizam status`
Show the current status of all configured services.

```bash
# Show status of all services
nizam status
```

### `nizam logs`
Display logs from running services.

```bash
# Show logs for a specific service
nizam logs postgres

# Follow logs in real-time
nizam logs --follow postgres

# Show last 100 lines
nizam logs --tail 100 postgres
```

**Options:**
- `--follow, -f` - Follow log output in real-time
- `--tail N` - Show last N lines of logs (default: 50)

### `nizam exec`
Execute commands inside running service containers.

```bash
# Open interactive shell
nizam exec postgres bash

# Run single command
nizam exec postgres psql -U user -d myapp
```

## Configuration Management

### `nizam init`
Initialize a new nizam configuration file in the current directory.

```bash
# Create default configuration
nizam init

# Initialize with custom services
nizam init --add postgres,mysql,redis
```

**Options:**
- `--add SERVICES` - Comma-separated list of services to add instead of defaults

### `nizam validate`
Validate configuration file syntax and structure.

```bash
# Validate default config
nizam validate

# Validate specific file
nizam validate --file ./production.yaml

# JSON output for automation
nizam validate --json

# Strict mode (exit non-zero on any issues)
nizam validate --strict
```

**Options:**
- `--file FILE` - Specify configuration file to validate
- `--json` - Output results in JSON format
- `--strict` - Exit with non-zero code on validation failures

### `nizam lint`
Analyze configuration for best practices and potential issues.

```bash
# Lint default configuration
nizam lint

# Lint specific file
nizam lint --file ./config.yaml

# JSON output for CI/CD
nizam lint --json
```

**Options:**
- `--file FILE` - Configuration file to analyze
- `--json` - Output results in JSON format

**Rules Checked:**
- **no-latest**: Prevents usage of `:latest` image tags
- **ports-shape**: Validates port mapping format
- **limits**: Recommends resource limits for services

### `nizam add`
Add a service from a template to your configuration.

```bash
# Add service with interactive configuration
nizam add postgres

# Add service with default values
nizam add postgres --defaults

# Add service with custom name
nizam add redis --name cache
```

**Options:**
- `--defaults` - Skip interactive prompts and use default values
- `--name NAME` - Custom name for the service (default: template name)
- `--overwrite` - Overwrite existing service with the same name

### `nizam remove`
Remove services from your configuration.

```bash
# Remove a service
nizam remove postgres

# Remove multiple services
nizam remove postgres redis

# Remove with confirmation
nizam remove --confirm postgres
```

**Options:**
- `--confirm` - Require confirmation before removal

## Data Lifecycle Management

### `nizam snapshot`
Manage database snapshots for backup, testing, and data sharing.

#### `nizam snapshot create <service>`
Create a snapshot of a service database.

```bash
# Basic snapshot creation
nizam snapshot create postgres
nizam snapshot create redis

# With custom options
nizam snapshot create postgres --tag "v1.2.0" --compress zstd --note "Release snapshot"
```

**Options:**
- `--compress string` - Compression type: `zstd` (default), `gzip`, `none`
- `--note string` - Note/description for the snapshot
- `--tag string` - Tag for the snapshot (default: timestamp)

#### `nizam snapshot list [service]`
List snapshots for a specific service or all services.

```bash
# List all snapshots across all services
nizam snapshot list

# List snapshots for specific service
nizam snapshot list postgres

# JSON output for automation
nizam snapshot list --json
```

**Options:**
- `--json` - Output in JSON format

#### `nizam snapshot restore <service>`
Restore a snapshot for a service.

```bash
# Restore latest snapshot
nizam snapshot restore postgres --latest

# Restore specific tagged snapshot
nizam snapshot restore postgres --tag "before-migration"

# Force restore without confirmation
nizam snapshot restore postgres --latest --force
```

**Options:**
- `--force` - Skip confirmation prompts
- `--latest` - Restore the most recent snapshot
- `--tag string` - Restore snapshot with specific tag

#### `nizam snapshot prune <service>`
Remove old snapshots, keeping the N most recent.

```bash
# Remove old snapshots, keeping 3 most recent
nizam snapshot prune postgres --keep 3

# Dry run to see what would be deleted
nizam snapshot prune postgres --keep 5 --dry-run
```

**Options:**
- `--dry-run` - Show what would be deleted without actually deleting
- `--keep int` - Number of snapshots to keep (required)

### `nizam psql`
Connect to PostgreSQL services with auto-resolved credentials.

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
```

**Options:**
- `--db string` - Database name (override config)
- `--user string` - Username (override config)

**Key Difference from `nizam exec`:**
- `nizam psql` auto-resolves credentials and builds connection strings
- `nizam exec postgres psql` requires manual specification of all connection details

### `nizam mysql`
Connect to MySQL services with auto-resolved credentials.

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

**Options:**
- `--db string` - Database name (override config)
- `--user string` - Username (override config)

**Key Features:**
- Auto-discovers MySQL services from configuration
- Extracts credentials from environment variables (MYSQL_USER, MYSQL_PASSWORD, etc.)
- Uses host binaries when available, falls back to container execution
- Supports both MySQL and MariaDB containers
- Supports pass-through arguments after `--`

### `nizam redis-cli`
Connect to Redis services with auto-configuration.

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

**Key Features:**
- Auto-discovers Redis services from configuration
- Extracts authentication details automatically
- Uses host binaries when available, falls back to container execution
- Supports pass-through arguments after `--`

## Health & Monitoring

### `nizam doctor`
Run comprehensive preflight checks on your Docker environment.

```bash
# Run all checks
nizam doctor

# Skip specific checks
nizam doctor --skip net.mtu,disk.free

# JSON output for automation
nizam doctor --json

# Attempt automatic fixes
nizam doctor --fix

# Verbose output
nizam doctor --verbose
```

**Options:**
- `--skip CHECKS` - Comma-separated list of check IDs to skip
- `--json` - Output results in JSON format
- `--fix` - Attempt automatic fixes for supported issues
- `--verbose` - Show detailed check information

**Checks Performed:**
- `docker.daemon` - Docker daemon connectivity
- `docker.compose` - Docker Compose availability
- `disk.free` - Available disk space
- `net.mtu` - Network MTU configuration
- `port.{PORT}` - Port availability for each service

### `nizam health`
Check health status of running services.

```bash
# Check all services
nizam health

# Check specific service
nizam health postgres

# Wait for services to become healthy
nizam health --wait

# JSON output
nizam health --json
```

**Options:**
- `--wait` - Wait for services to become healthy
- `--timeout DURATION` - Maximum wait time (default: 30s)
- `--json` - Output status in JSON format

### `nizam health-server`
Start HTTP health check server for monitoring integration.

```bash
# Start on default port (8080)
nizam health-server

# Start on custom port
nizam health-server --port 9090

# Enable metrics endpoint
nizam health-server --metrics
```

**Options:**
- `--port PORT` - HTTP server port (default: 8080)
- `--metrics` - Enable Prometheus metrics endpoint

**Endpoints:**
- `/health` - Overall health status
- `/health/{service}` - Individual service health
- `/metrics` - Prometheus metrics (if enabled)

## Development Tools

### `nizam wait-for`
Wait for services to become ready before proceeding.

```bash
# Wait for specific service
nizam wait-for database

# Wait for multiple services
nizam wait-for web database cache

# Wait for all services
nizam wait-for

# Custom timeout
nizam wait-for --timeout 60s database

# Custom check interval
nizam wait-for --interval 2s database
```

**Aliases:** `nizam wait`

**Options:**
- `--timeout DURATION` - Maximum wait time (default: 30s)
- `--interval DURATION` - Check interval (default: 1s)

**Readiness Checks:**
- Port connectivity for services with exposed ports
- HTTP health checks for services with health check URLs
- Assumes ready if no checks are configured

### `nizam retry`
Retry failed operations with exponential backoff.

```bash
# Retry start command
nizam retry start

# Retry with custom attempts
nizam retry start --attempts 5

# Custom initial delay
nizam retry start --delay 2s

# Retry specific services
nizam retry start web database --attempts 3
```

**Options:**
- `--attempts N` - Maximum retry attempts (default: 3)
- `--delay DURATION` - Initial delay between retries (default: 1s)

**Supported Operations:**
- `start` - Start services
- `stop` - Stop services  
- `restart` - Restart services
- `pull` - Pull images
- `build` - Build images

## Utility Commands

### `nizam completion`
Generate shell completion scripts.

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

**Supported Shells:**
- bash
- zsh
- fish
- powershell

**Installation:**
```bash
# Bash (add to ~/.bashrc)
echo 'source <(nizam completion bash)' >> ~/.bashrc

# Zsh (add to ~/.zshrc)
echo 'source <(nizam completion zsh)' >> ~/.zshrc
```

### `nizam update`
Update nizam to the latest version.

```bash
# Check for updates
nizam update --check

# Update to latest version
nizam update

# Include prerelease versions
nizam update --prerelease
```

**Options:**
- `--check` - Only check for updates, don't install
- `--prerelease` - Include prerelease versions

### `nizam templates`
List available service templates.

```bash
# Show all templates
nizam templates

# Show template details
nizam templates --details

# Filter by category
nizam templates --category database
```

**Options:**
- `--details` - Show detailed template information
- `--category CATEGORY` - Filter by template category

### `nizam custom`
Manage custom service templates.

```bash
# List custom templates
nizam custom list

# Create custom template
nizam custom create mytemplate

# Import template from file
nizam custom import ./template.yaml

# Export template
nizam custom export mytemplate > template.yaml
```

**Subcommands:**
- `list` - List custom templates
- `create NAME` - Create new custom template
- `import FILE` - Import template from file
- `export NAME` - Export template to stdout

### `nizam export`
Export service configuration as a custom template.

```bash
# Export service as template
nizam export postgres mypostgres

# Export with description
nizam export postgres mypostgres --description "Custom Postgres setup"
```

**Options:**
- `--description TEXT` - Template description

## Global Options

All commands support these global options:

- `--config FILE` - Configuration file path (default: .nizam.yaml)
- `--profile PROFILE` - Configuration profile to use (default: dev)
- `--verbose, -v` - Enable verbose logging
- `--help, -h` - Show help information
- `--version` - Show version information

## Exit Codes

Standard exit codes used by nizam commands:

- `0` - Success
- `1` - General error
- `2` - Invalid arguments
- `3` - Configuration error
- `4` - Docker/system error

## Environment Variables

- `NIZAM_CONFIG` - Override default configuration file path
- `NIZAM_PROFILE` - Override default profile
- `NIZAM_VERBOSE` - Enable verbose logging (true/false)
- `NIZAM_DOCTOR_TIMEOUT` - Override doctor check timeout
- `NIZAM_DOCTOR_CONCURRENCY` - Override doctor concurrency limit

## Examples

### Development Workflow
```bash
# Initialize new project
nizam init

# Add required services
nizam add postgres redis

# Start development environment
nizam up

# Check everything is working
nizam doctor
nizam health

# View service logs
nizam logs --follow postgres

# Clean up
nizam down
```

### CI/CD Integration
```bash
# Validate configuration
nizam validate --json --strict

# Lint for best practices
nizam lint --json

# Environment check
nizam doctor --json

# Start services for testing
nizam up

# Wait for services to be ready
nizam wait-for --timeout 60s

# Run tests...

# Clean up
nizam down --force
```

### Production Health Monitoring
```bash
# Start health server
nizam health-server --port 8080 --metrics &

# Check service health
curl http://localhost:8080/health

# Get metrics
curl http://localhost:8080/metrics
```
