# Host Binary Detection Implementation

This document describes the host binary detection feature implemented for database CLI commands in nizam.

## Overview

The host binary detection system automatically detects if database client binaries (psql, mysql, redis-cli, mongosh) are available on the host system and uses them when possible, falling back gracefully to container execution when binaries are not found.

## Architecture

### Core Components

1. **`internal/binary` package** - Central binary detection with caching
2. **`internal/resolve` package** - Integration point for service resolution
3. **Database CLI commands** - Updated to use host binaries when available

### Binary Detection Flow

```
User runs command (e.g. `nizam psql`)
        ↓
Check if psql binary exists on host
        ↓
    Available?
   ↙         ↘
 Yes          No
  ↓            ↓
Use host     Use container
binary       execution
```

## Implementation Details

### Binary Package

The `internal/binary` package provides:

- **Thread-safe caching** - Results are cached to avoid repeated filesystem lookups
- **Cross-platform compatibility** - Uses `exec.LookPath` for platform-agnostic detection
- **ClientType enum** - Type-safe constants for supported database clients
- **Cache management** - Functions to clear cache when needed

```go
// Check if PostgreSQL client is available
if binary.HasBinary(binary.PostgreSQL) {
    // Use host binary
} else {
    // Fall back to container
}
```

### Supported Database Clients

| Database   | Binary      | ClientType Constant |
|------------|-------------|---------------------|
| PostgreSQL | `psql`      | `binary.PostgreSQL` |
| MySQL      | `mysql`     | `binary.MySQL`      |
| Redis      | `redis-cli` | `binary.Redis`      |
| MongoDB    | `mongosh`   | `binary.MongoDB`    |

### Command Integration

Each database CLI command follows the same pattern:

1. **Detection** - Check if host binary exists using `binary.HasBinary()`
2. **Host execution** - Execute using `exec.Command()` with proper TTY forwarding
3. **Container fallback** - Use `dockerx.ExecTTY()` if host binary unavailable
4. **Exit code handling** - Proper exit code forwarding for both execution methods

## Benefits

### Performance
- **Faster startup** - No Docker overhead when using host binaries
- **Native feel** - Direct binary execution provides better responsiveness
- **Cached detection** - Binary availability is cached to avoid repeated checks

### User Experience
- **Seamless fallback** - Transparent container execution when binaries missing
- **Consistent interface** - Same command syntax regardless of execution method
- **Debug logging** - Clear indication of which execution method is being used

### Compatibility
- **Cross-platform** - Works on macOS, Linux, and Windows
- **Version independence** - Works with any version of database clients
- **Configuration agnostic** - No configuration changes required

## Usage Examples

```bash
# These commands automatically detect and use host binaries when available

# PostgreSQL
nizam psql                    # Connect to first PostgreSQL service
nizam psql postgres -- -c "SELECT version()"

# MySQL  
nizam mysql                   # Connect to first MySQL service
nizam mysql -- -e "SHOW DATABASES"

# Redis
nizam redis-cli               # Connect to first Redis service
nizam redis-cli -- ping

# MongoDB
nizam mongosh                 # Connect to first MongoDB service
nizam mongosh -- --eval "db.version()"
```

## Debug Information

Use the `--verbose` flag to see which execution method is being used:

```bash
nizam --verbose psql
# Output: "Using host psql binary" or "psql not found on host, using container execution"
```

## Testing

The implementation includes comprehensive tests:

- **Binary detection tests** - Verify correct detection on the current system
- **Cache behavior tests** - Ensure caching works correctly
- **Integration tests** - Verify commands work with both execution methods

Run tests with:
```bash
go test ./internal/binary -v
go test ./internal/resolve -v
```

## Future Enhancements

Potential improvements for the binary detection system:

1. **Version checking** - Detect minimum required versions of client binaries
2. **Custom binary paths** - Allow users to specify custom paths for binaries
3. **Installation suggestions** - Suggest how to install missing binaries
4. **Configuration options** - Allow users to force container execution
5. **Performance metrics** - Track execution time differences between methods
