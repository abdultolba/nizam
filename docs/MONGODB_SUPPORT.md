# MongoDB Support Implementation

## Overview

This document summarizes the implementation of MongoDB snapshot and CLI support in `nizam`, completing a major data lifecycle management milestone.

## Features

### 1. MongoDB Snapshot Engine (`internal/snapshot/mongodb_engine.go`)

**Core Functionality:**

- Full `mongodump`/`mongorestore` integration with streaming support
- Multi-compression support (zstd, gzip, none)
- Atomic operations with temporary files and checksums
- Rich error handling and logging
- Database drop functionality for clean restores

**Key Methods:**

- `Create()` - Creates MongoDB snapshots using mongodump with archive format
- `Restore()` - Restores MongoDB snapshots using mongorestore
- `CanHandle()` - Handles both "mongo" and "mongodb" engine types
- `GetEngineType()` - Returns "mongo" for consistency with resolve package

**Technical Details:**

- Uses `mongodump --archive --gzip` for efficient streaming backups
- Supports authentication via `--username`/`--password` flags
- Implements force restore with database recreation
- Integrated checksum verification for data integrity

### 2. MongoDB One-liner CLI (`cmd/mongosh.go`)

**Core Functionality:**

- Auto-discovery of MongoDB services from configuration
- Smart credential resolution from environment variables
- Host binary detection with container execution fallback
- Pass-through argument support for native mongosh commands

**Key Features:**

- Service auto-discovery when no service name specified
- Connection string building with proper MongoDB URLs
- Credential redaction in logs for security
- Support for both host and container execution modes

**Usage Examples:**

```bash
# Auto-connect to first MongoDB service
nizam mongosh

# Connect to specific service
nizam mongosh mydb

# Override connection parameters
nizam mongosh --user admin --db myapp

# Pass arguments to mongosh
nizam mongosh -- --eval "db.version()"
```

### 3. Engine Registration and Integration

**Service Registration:**

- MongoDB engine registered in snapshot service (`internal/snapshot/service.go`)
- Supports both "mongo" and "mongodb" engine identifiers
- Fully integrated with existing snapshot lifecycle operations

**Resolution Integration:**

- Enhanced resolve package to handle MongoDB environment variables
- Added support for `MONGO_INITDB_ROOT_USERNAME`, `MONGO_INITDB_ROOT_PASSWORD`
- Proper engine detection from image names and service names

### 4. Testing and Quality Assurance

**Unit Tests:**

- Comprehensive test coverage for MongoDB engine (`mongodb_engine_test.go`)
- Engine type and capability testing
- Integration with existing test infrastructure

**Build Verification:**

- All code compiles successfully
- All existing tests continue to pass
- New MongoDB-specific tests validate core functionality

## Architecture Decisions

### 1. Engine Type Consistency

- Used "mongo" as primary engine identifier for consistency with existing resolve package
- Maintained backward compatibility with "mongodb" identifier
- Aligned with existing PostgreSQL/MySQL patterns

### 2. Command Integration

- Follows existing patterns from `psql.go` and `mysql.go` implementations
- Maintains consistent argument parsing and credential handling
- Preserves security practices with credential redaction

### 3. Snapshot Implementation

- Leverages existing compression and manifest infrastructure
- Maintains atomic operations pattern with temporary files
- Integrated seamlessly with existing snapshot lifecycle commands

## Updated Features

### Engine Support Matrix

| Engine      | Snapshot | One-liner CLI            | Health Checks |
| ----------- | -------- | ------------------------ | ------------- |
| PostgreSQL  | ✅       | ✅ (`nizam psql`)        | ✅            |
| MySQL       | ✅       | ✅ (`nizam mysql`)       | ✅            |
| Redis       | ✅       | ✅ (`nizam redis-cli`)   | ✅            |
| **MongoDB** | **✅**   | **✅** (`nizam mongosh`) | ✅            |

## Usage Examples

### MongoDB Snapshot Operations

```bash
# Create MongoDB snapshot
nizam snapshot create mongodb --tag "before-migration"

# List MongoDB snapshots
nizam snapshot list mongodb

# Restore latest MongoDB snapshot
nizam snapshot restore mongodb --latest

# Prune old MongoDB snapshots
nizam snapshot prune mongodb --keep 5
```

### MongoDB CLI Access

```bash
# Connect to MongoDB service
nizam mongosh

# Run MongoDB commands
nizam mongosh -- --eval "db.stats()"

# Connect with overrides
nizam mongosh --user admin --db production
```

## Verification

All implementation has been verified through:

- ✅ Successful Go build (`go build`)
- ✅ All unit tests passing (`go test ./internal/snapshot/...`)
- ✅ CLI help output verification
- ✅ Command registration verification
