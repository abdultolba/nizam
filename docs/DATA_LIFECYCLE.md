# Data Lifecycle Management

nizam provides comprehensive data lifecycle tools for database snapshots and one-liner database access, making it easy to capture, restore, and work with database states during development.

## Table of Contents

- [Database Snapshots](#database-snapshots)
- [One-liner Database Access](#one-liner-database-access)
- [Architecture & Implementation](#architecture--implementation)
- [Use Cases & Workflows](#use-cases--workflows)
- [Troubleshooting](#troubleshooting)

## Database Snapshots

Create point-in-time snapshots of your databases for backup, testing, or sharing data states.

### Features

- 🎯 **Multi-engine support**: PostgreSQL, MySQL, and Redis (MongoDB planned)
- 🗜️ **Smart compression**: zstd (default), gzip, or none with automatic streaming
- 🔒 **Data integrity**: SHA256 checksums for all snapshot files
- 📋 **Rich metadata**: Tagged snapshots with notes, timestamps, and version tracking
- 📁 **Organized storage**: Structured storage in `.nizam/snapshots/<service>/`
- ⚡ **Atomic operations**: Safe creation and restoration with temporary files

### Quick Start

```bash
# Create a basic snapshot
nizam snapshot create postgres

# Create a tagged snapshot with notes
nizam snapshot create postgres --tag "before-migration" --note "Pre-schema update"

# List all snapshots
nizam snapshot list

# Restore the latest snapshot
nizam snapshot restore postgres --latest

# Clean up old snapshots
nizam snapshot prune postgres --keep 5
```

### Commands Reference

#### `nizam snapshot create <service>`

Create a snapshot of a service database.

```bash
# Basic usage
nizam snapshot create postgres
nizam snapshot create mysql
nizam snapshot create redis

# With options
nizam snapshot create postgres \
  --tag "v1.2.0" \
  --compress zstd \
  --note "Release snapshot"
```

**Flags:**

- `--compress string` - Compression type: `zstd` (default), `gzip`, `none`
- `--note string` - Note/description for the snapshot
- `--tag string` - Tag for the snapshot (default: timestamp)

**Output:**

```
✓ Created snapshot for postgres
  Location: .nizam/snapshots/postgres/20240810-143022-v1.2.0/
  Files: pg.dump.zst (15.2MB)
  Checksum: sha256:a1b2c3d4...
```

#### `nizam snapshot list [service]`

List snapshots for a specific service or all services.

```bash
# List all snapshots
nizam snapshot list

# List for specific service
nizam snapshot list postgres

# JSON output for automation
nizam snapshot list --json
```

**Flags:**

- `--json` - Output in JSON format

**Output (table format):**

```
SERVICE   TAG              CREATED              SIZE    NOTE
postgres  before-migration 2024-08-10 14:30:22  15.2MB  Pre-schema update
postgres  v1.1.0          2024-08-09 09:15:33  14.8MB
redis     cache-backup     2024-08-10 12:45:11   2.1MB  Daily backup
```

**Output (JSON format):**

```json
{
  "snapshots": [
    {
      "service": "postgres",
      "tag": "before-migration",
      "createdAt": "2024-08-10T14:30:22Z",
      "size": 15943680,
      "note": "Pre-schema update",
      "path": ".nizam/snapshots/postgres/20240810-143022-before-migration/",
      "engine": "postgresql",
      "compression": "zstd"
    }
  ]
}
```

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

**Flags:**

- `--force` - Skip confirmation prompts
- `--latest` - Restore the most recent snapshot
- `--tag string` - Restore snapshot with specific tag

**Confirmation prompt:**

```
⚠️  This will replace all data in the postgres database.
   Service: postgres
   Snapshot: before-migration (2024-08-10 14:30:22)
   Size: 15.2MB

Continue? [y/N]:
```

#### `nizam snapshot prune <service>`

Remove old snapshots, keeping the N most recent.

```bash
# Keep 3 most recent snapshots
nizam snapshot prune postgres --keep 3

# Dry run to preview what would be deleted
nizam snapshot prune postgres --keep 3 --dry-run
```

**Flags:**

- `--dry-run` - Show what would be deleted without deleting
- `--keep int` - Number of snapshots to keep (required)

**Dry run output:**

```
Would remove 2 snapshots for postgres (reclaim 28.7MB):
  ✗ v1.0.0 (2024-08-05 16:22:11) - 14.1MB
  ✗ initial (2024-08-03 10:15:44) - 14.6MB

Keeping 3 most recent snapshots:
  ✓ before-migration (2024-08-10 14:30:22) - 15.2MB
  ✓ v1.1.0 (2024-08-09 09:15:33) - 14.8MB
  ✓ staging-data (2024-08-08 11:45:20) - 13.9MB
```

### Storage Structure

Snapshots are organized in a predictable directory structure:

```
.nizam/snapshots/
├── postgres/
│   ├── 20240810-143022-before-migration/
│   │   ├── manifest.json
│   │   └── pg.dump.zst
│   ├── 20240809-091533-v1.1.0/
│   │   ├── manifest.json
│   │   └── pg.dump.zst
│   └── 20240808-114520-staging-data/
│       ├── manifest.json
│       └── pg.dump.zst
└── redis/
    └── 20240810-124511-cache-backup/
        ├── manifest.json
        └── dump.rdb.gz
```

### Manifest Format

Each snapshot includes a `manifest.json` file with metadata:

```json
{
  "service": "postgres",
  "engine": "postgresql",
  "image": "postgres:16.3",
  "createdAt": "2024-08-10T14:30:22Z",
  "tag": "before-migration",
  "toolVersion": "0.7.0",
  "compression": "zstd",
  "encryption": "none",
  "note": "Pre-schema update",
  "files": [
    {
      "name": "pg.dump.zst",
      "sha256": "a1b2c3d4e5f6789012345678901234567890abcdef",
      "size": 15943680
    }
  ]
}
```

## One-liner Database Access

Connect to your databases instantly with auto-resolved connection parameters.

### Features

- 🔧 **Auto-resolution**: Automatically discovers connection details from configuration
- 🔄 **Smart fallback**: Uses host binaries when available, falls back to container execution
- 🎯 **Service detection**: Auto-detects the first service of each type if not specified
- 📋 **Pass-through args**: All arguments after `--` are passed directly to the database CLI

### PostgreSQL Access

#### `nizam psql [service]`

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

**Flags:**

- `--db string` - Database name (override config)
- `--user string` - Username (override config)

**Connection Resolution Example:**

Given this configuration:

```yaml
services:
  postgres:
    image: postgres:16
    ports: ["5432:5432"]
    env:
      POSTGRES_USER: myuser
      POSTGRES_PASSWORD: mypass
      POSTGRES_DB: mydb
```

The command `nizam psql` resolves to:

```bash
# If psql available on host:
psql "postgresql://myuser:mypass@localhost:5432/mydb?sslmode=disable"

# If psql not available on host:
docker exec -it nizam_postgres psql -U myuser -d mydb
```

### MySQL Access

#### `nizam mysql [service]`

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

**Flags:**

- `--db string` - Database name (override config)
- `--user string` - Username (override config)

**Connection Resolution Example:**

Given this configuration:

```yaml
services:
  mysql:
    image: mysql:8.0
    ports: ["3306:3306"]
    env:
      MYSQL_USER: myuser
      MYSQL_PASSWORD: mypass
      MYSQL_DATABASE: mydb
```

The command `nizam mysql` resolves to:

```bash
# If mysql available on host:
mysql -h localhost -P 3306 -u myuser -pmypass mydb

# If mysql not available on host:
docker exec -it nizam_mysql mysql -u myuser -h localhost -pmypass mydb
```

### Redis Access

#### `nizam redis-cli [service]`

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

**Connection Resolution Example:**

Given this configuration:

```yaml
services:
  redis:
    image: redis:7
    ports: ["6379:6379"]
    env:
      REDIS_PASSWORD: mypassword
```

The command `nizam redis-cli` resolves to:

```bash
# If redis-cli available on host:
redis-cli -h localhost -p 6379 -a mypassword

# If redis-cli not available on host:
docker exec -it nizam_redis redis-cli -a mypassword
```

### Smart vs Raw Execution

Understanding the difference between smart connection commands and raw container execution:

| Feature                   | `nizam psql`                     | `nizam exec postgres psql`         |
| ------------------------- | -------------------------------- | ---------------------------------- |
| **Credential resolution** | ✅ Automatic from config         | ❌ Manual specification required   |
| **Connection strings**    | ✅ Auto-built URLs               | ❌ Manual argument construction    |
| **Host binary usage**     | ✅ Uses host `psql` if available | ❌ Always executes in container    |
| **Service discovery**     | ✅ Auto-finds PostgreSQL service | ❌ Must specify exact service name |
| **Ease of use**           | 🟢 Just works                    | 🟡 Requires connection knowledge   |
| **Flexibility**           | 🟡 Opinionated                   | 🟢 Total control                   |

**Smart Connection Examples:**

```bash
nizam psql                           # Auto-connects
nizam psql -- -c "SELECT version()"  # Runs query automatically
nizam mysql                          # Auto-connects to MySQL
nizam mysql -- -e "SHOW DATABASES"   # Runs MySQL query automatically
nizam redis-cli -- ping             # Auto-authenticated ping
```

**Raw Container Execution Examples:**

```bash
nizam exec postgres psql -U user -d mydb -h localhost
nizam exec mysql mysql -u user -pmypass mydb
nizam exec redis redis-cli -a password ping
```

## Architecture & Implementation

### Snapshot Engines

nizam uses engine-specific implementations for different database types:

#### PostgreSQL Engine

- Uses `pg_dump --format=custom` for consistent binary dumps
- Streams output directly to compressed files
- Restores using `pg_restore --clean --if-exists`
- Handles connection parameters from service environment

#### MySQL Engine

- Uses `mysqldump` with comprehensive options for consistent dumps
- Includes routines, triggers, events, and complete inserts
- Restores using `mysql` client with streaming support
- Handles connection parameters from service environment variables
- Supports both MySQL and MariaDB containers

#### Redis Engine

- Uses `BGSAVE` command for consistent point-in-time snapshots
- Copies `dump.rdb` file from container data directory
- Restores by stopping container, replacing file, and restarting
- Preserves Redis configuration and persistence settings

### Compression Pipeline

Snapshots use a streaming compression pipeline:

```
Database Engine → Compressor → Checksum → Atomic Write
    (pg_dump)       (zstd)      (sha256)    (temp + rename)
```

**Compression Options:**

- **zstd** (default): Best compression ratio and speed balance
- **gzip**: Wide compatibility, moderate compression
- **none**: No compression, fastest for small datasets

### Connection Resolution

One-liner commands follow a resolution chain:

```
1. Parse command arguments (service, overrides)
2. Load .nizam.yaml configuration
3. Find matching service by type
4. Extract connection details from environment
5. Check for host binary availability
6. Build connection string or container command
7. Execute with appropriate method
```

**Service Discovery Logic:**

```go
// Pseudo-code for service discovery
func FindDatabaseService(config, serviceType) Service {
    if serviceName := args.ServiceName; serviceName != "" {
        return config.Services[serviceName]
    }

    // Auto-discover first service of matching type
    for name, service := range config.Services {
        if DetectEngine(service) == serviceType {
            return service
        }
    }

    return error("No PostgreSQL/Redis service found")
}
```

### Security Considerations

- **Credential Handling**: Passwords never logged in plain text
- **Container Isolation**: Database operations run in isolated containers
- **File Permissions**: Snapshot files created with 644 permissions
- **Checksum Verification**: All snapshots verified on creation and restore
- **Atomic Operations**: Temporary files prevent partial state corruption

## Use Cases & Workflows

### Development Workflow

**Database State Management:**

```bash
# Save current state before major changes
nizam snapshot create postgres --tag "before-schema-migration"

# Make schema changes...

# If something goes wrong, restore
nizam snapshot restore postgres --tag "before-schema-migration"

# Regular development database access
nizam psql -- -c "\\dt"
nizam redis-cli -- keys "*user*"
```

**Feature Branch Development:**

```bash
# Create feature snapshot
nizam snapshot create postgres --tag "feature-auth"

# Switch to different feature
nizam snapshot restore postgres --tag "feature-payments"

# Quick database exploration
nizam psql postgres -- -c "SELECT * FROM users LIMIT 5"
```

### Testing & QA

**Test Data Setup:**

```bash
# Create test data snapshot
nizam snapshot create postgres --tag "test-suite-data" --note "Complete test dataset"

# Before each test run
nizam snapshot restore postgres --tag "test-suite-data" --force

# Verify test state
nizam psql -- -c "SELECT COUNT(*) FROM test_users"
```

**Integration Testing:**

```bash
# Save clean state
nizam snapshot create postgres --tag "integration-baseline"

# Run integration tests...

# Restore clean state for next test
nizam snapshot restore postgres --tag "integration-baseline" --force
```

### Team Collaboration

**Sharing Database States:**

```bash
# Team member creates useful snapshot
nizam snapshot create postgres --tag "demo-data-v2" --note "Updated demo dataset for Q4"

# Others can restore the same state
nizam snapshot restore postgres --tag "demo-data-v2"

# Verify shared state
nizam psql -- -c "SELECT version_info FROM schema_versions"
```

**Debugging & Support:**

```bash
# Create snapshot for debugging
nizam snapshot create postgres --tag "bug-reproduction" --note "Issue #123 reproduction data"

# Quick data inspection
nizam psql -- -c "SELECT * FROM error_logs WHERE created_at > NOW() - INTERVAL '1 hour'"

# Share snapshot location with team
ls -la .nizam/snapshots/postgres/
```

### Production Debugging

**Safe Local Debugging:**

```bash
# Import sanitized production data (when available)
# NOTE: Production import features planned for future release

# Create local debugging snapshot
nizam snapshot create postgres --tag "production-debug" --note "Sanitized prod data for issue #456"

# Debug locally with production-like data
nizam psql -- -c "EXPLAIN ANALYZE SELECT * FROM slow_query_table"
```

### Backup & Recovery

**Regular Backups:**

```bash
# Daily snapshot with cleanup
nizam snapshot create postgres --tag "daily-$(date +%Y%m%d)"
nizam snapshot prune postgres --keep 7  # Keep 1 week

# Weekly snapshot with different retention
nizam snapshot create postgres --tag "weekly-$(date +%Y-W%U)"
nizam snapshot list postgres | grep weekly | tail -n +5 | cut -f1 | xargs -r nizam snapshot remove postgres
```

**Disaster Recovery:**

```bash
# List available snapshots
nizam snapshot list postgres

# Restore most recent snapshot
nizam snapshot restore postgres --latest

# Verify restore
nizam psql -- -c "SELECT NOW() as restored_at"
```

## Troubleshooting

### Common Issues

#### Snapshot Creation Fails

**Error: "Container not running"**

```bash
# Check service status
nizam status

# Start service if needed
nizam up postgres
```

**Error: "Permission denied accessing container"**

```bash
# Check Docker permissions
docker ps

# Verify nizam container naming
docker ps --filter "name=nizam_"
```

**Error: "Insufficient disk space"**

```bash
# Check available space
df -h .nizam/

# Clean up old snapshots
nizam snapshot prune postgres --keep 3
nizam snapshot prune redis --keep 3
```

#### Restoration Issues

**Error: "Snapshot not found"**

```bash
# List available snapshots
nizam snapshot list postgres

# Verify snapshot path exists
ls -la .nizam/snapshots/postgres/
```

**Error: "Checksum verification failed"**

```bash
# Snapshot may be corrupted - recreate it
nizam snapshot create postgres --tag "replacement-snapshot"
```

#### One-liner Connection Issues

**Error: "Service not found"**

```bash
# Check service configuration
nizam status

# Verify service type
nizam psql postgres  # Use explicit service name
```

**Error: "Connection refused"**

```bash
# Check if service is running
nizam status

# Verify port mapping
docker port nizam_postgres

# Check service health
nizam health postgres
```

**Error: "Authentication failed"**

```bash
# Verify credentials in configuration
cat .nizam.yaml | grep -A 5 postgres

# Try explicit credentials
nizam psql --user postgres --db postgres
```

### Performance Optimization

#### Large Database Snapshots

**Use appropriate compression:**

```bash
# For large databases, zstd provides best balance
nizam snapshot create postgres --compress zstd

# For maximum compatibility, use gzip
nizam snapshot create postgres --compress gzip

# For fastest creation (small DBs), disable compression
nizam snapshot create postgres --compress none
```

**Monitor snapshot sizes:**

```bash
# Check snapshot disk usage
du -sh .nizam/snapshots/*

# List snapshots by size
nizam snapshot list --json | jq -r '.snapshots[] | "\(.size) \(.service) \(.tag)"' | sort -n
```

#### Connection Performance

**Host vs Container Execution:**

```bash
# Install database clients on host for better performance
brew install postgresql redis

# Verify host binaries are found
which psql redis-cli

# Commands will automatically use host binaries when available
nizam psql    # Uses host psql if available
nizam redis-cli  # Uses host redis-cli if available
```

### Debugging Commands

**Verbose snapshot creation:**

```bash
nizam -v snapshot create postgres --tag debug-snapshot
```

**Check snapshot manifest:**

```bash
cat .nizam/snapshots/postgres/20240810-143022-debug-snapshot/manifest.json | jq .
```

**Verify snapshot integrity:**

```bash
# Check file checksums match manifest
cd .nizam/snapshots/postgres/20240810-143022-debug-snapshot/
sha256sum *.zst
# Compare with manifest.json checksums
```

**Connection debugging:**

```bash
# Test manual connection
nizam exec postgres psql -U user -d mydb -c "SELECT 1"

# Compare with smart connection
nizam -v psql -- -c "SELECT 1"
```

### Getting Help

For additional support:

1. **Check logs**: Use `nizam -v` for verbose output
2. **Verify configuration**: Run `nizam validate` and `nizam doctor`
3. **Service status**: Check `nizam status` and `nizam health`
4. **Docker state**: Verify with `docker ps` and `docker logs nizam_<service>`

---

_For more information, see the main [README](../README.md) or [CLI Commands documentation](COMMANDS.md)._
