# Doctor Module Documentation

The doctor module provides comprehensive preflight checks for the nizam environment to ensure Docker and system resources are properly configured.

## Overview

The doctor system is designed to:
- Detect common configuration issues before they cause problems
- Provide actionable suggestions for fixing detected issues
- Run checks concurrently for performance
- Support both human-readable and machine-readable (JSON) output
- Allow selective skipping of checks for flexibility

## Architecture

### Core Components

#### `Runner`
The central orchestrator that manages check execution:
- Controls concurrency using semaphores
- Collects and aggregates results
- Provides formatted output
- Manages check filtering and skipping

#### `Check` Interface
```go
type Check interface {
    ID() string           // Unique identifier for the check
    Name() string         // Human-readable name
    Run() Result          // Execute the check
    CanFix() bool         // Whether auto-fix is supported
    Fix() error           // Attempt automatic fix
}
```

#### `Result` Structure
```go
type Result struct {
    ID          string    // Check identifier
    Name        string    // Human-readable name
    Status      Status    // Pass, Fail, Warning
    Message     string    // Descriptive message
    Suggestion  string    // Fix suggestion
    Details     string    // Additional context
    Timestamp   time.Time // When check was performed
}
```

### Status Types

- **Pass**: Check completed successfully, no issues detected
- **Fail**: Critical issue that prevents proper operation
- **Warning**: Advisory issue that may cause problems

## Implemented Checks

### Docker Checks (`internal/doctor/checks/docker.go`)

#### `docker.daemon`
- **Purpose**: Verify Docker daemon is running and accessible
- **Method**: Execute `docker version` command
- **Failure Impact**: Critical - nizam cannot function without Docker
- **Auto-fix**: Not supported

#### `docker.compose`
- **Purpose**: Verify Docker Compose plugin is available
- **Method**: Execute `docker compose version` command
- **Failure Impact**: Critical - required for multi-service management
- **Auto-fix**: Not supported

### System Checks (`internal/doctor/checks/system.go`)

#### `disk.free`
- **Purpose**: Ensure adequate disk space for Docker operations
- **Method**: Check available space on current filesystem
- **Threshold**: Warns if less than 1GB available
- **Failure Impact**: Warning - low disk space can cause image pull failures
- **Auto-fix**: Not supported (requires manual cleanup)

#### `net.mtu`
- **Purpose**: Check network MTU configuration
- **Method**: Read MTU from default network interface
- **Threshold**: Warns if MTU != 1500 (non-standard)
- **Failure Impact**: Warning - can cause Docker networking issues
- **Auto-fix**: Not supported (requires network configuration)

#### `port.{PORT}` (Dynamic)
- **Purpose**: Check if service ports are available
- **Method**: Attempt TCP connection to localhost:port
- **Source**: Generated dynamically from nizam service configurations
- **Failure Impact**: Critical - port conflicts prevent service startup
- **Auto-fix**: Could be supported via port remapping

## Usage Examples

### Basic Usage
```bash
# Run all checks
nizam doctor

# Skip specific checks
nizam doctor --skip net.mtu,disk.free

# JSON output for automation
nizam doctor --json

# Attempt automatic fixes
nizam doctor --fix
```

### Programmatic Usage
```go
import "github.com/abdultolba/nizam/internal/doctor"

// Create runner with default checks
runner := doctor.NewRunner(10) // max 10 concurrent checks

// Add custom check
runner.AddCheck(&MyCustomCheck{})

// Run all checks
results := runner.Run()

// Check for failures
if runner.HasFailures(results) {
    // Handle failures
}

// Output results
runner.PrintResults(results, false) // human-readable
runner.PrintResults(results, true)  // JSON
```

## Extending the System

### Adding New Checks

1. **Implement the Check Interface**:
```go
type MyCheck struct{}

func (c *MyCheck) ID() string { return "my.check" }
func (c *MyCheck) Name() string { return "My Custom Check" }
func (c *MyCheck) CanFix() bool { return false }
func (c *MyCheck) Fix() error { return errors.New("not implemented") }

func (c *MyCheck) Run() doctor.Result {
    // Perform your check logic
    if someCondition {
        return doctor.Result{
            ID:         c.ID(),
            Name:       c.Name(),
            Status:     doctor.StatusPass,
            Message:    "Check passed successfully",
            Timestamp:  time.Now(),
        }
    }
    
    return doctor.Result{
        ID:         c.ID(),
        Name:       c.Name(),
        Status:     doctor.StatusFail,
        Message:    "Check failed",
        Suggestion: "Fix by doing X",
        Timestamp:  time.Now(),
    }
}
```

2. **Register the Check**:
```go
// In cmd/doctor.go
runner.AddCheck(&MyCheck{})
```

### Adding Auto-Fix Support

1. **Implement CanFix() and Fix()**:
```go
func (c *MyCheck) CanFix() bool { return true }

func (c *MyCheck) Fix() error {
    // Attempt automatic remediation
    if err := performFix(); err != nil {
        return fmt.Errorf("failed to fix: %w", err)
    }
    return nil
}
```

2. **The runner handles the rest** - it will call Fix() when `--fix` is used

## Configuration

### Environment Variables
- `NIZAM_DOCTOR_TIMEOUT` - Override default check timeout
- `NIZAM_DOCTOR_CONCURRENCY` - Override default concurrency limit

### Check-Specific Configuration
Some checks support configuration through the nizam config file:

```yaml
doctor:
  disk:
    min_free_gb: 2  # Override default 1GB threshold
  ports:
    skip_check: false  # Skip port availability checks
```

## Error Handling

The doctor system uses structured error handling:

- **Check Failures**: Captured as Result with StatusFail
- **System Errors**: Logged and reported separately
- **Timeout Handling**: Checks that take too long are cancelled
- **Concurrent Safety**: All operations are thread-safe

## Performance Considerations

- **Concurrency**: Checks run in parallel with configurable limits
- **Timeouts**: Each check has a reasonable timeout to prevent hanging
- **Resource Usage**: Minimal system impact during execution
- **Caching**: Results could be cached for repeated runs (future enhancement)

## Future Enhancements

### Planned Features
1. **Configuration-based checks** - Define custom checks in nizam.yaml
2. **Check dependencies** - Run checks in specific order based on dependencies
3. **Historical tracking** - Track check results over time
4. **Integration testing** - Validate service integration beyond port checks
5. **Performance benchmarking** - Measure Docker and system performance

### Extension Points
1. **Custom check providers** - Plugin system for third-party checks
2. **Notification integration** - Send results to monitoring systems
3. **Remediation workflows** - Complex multi-step fixes
4. **Check scheduling** - Periodic health monitoring

## Troubleshooting

### Common Issues

**Docker daemon check fails**
- Ensure Docker Desktop is running
- Check Docker daemon socket permissions
- Verify Docker installation

**Port checks fail**
- Identify conflicting processes: `lsof -i :PORT`
- Consider using different ports
- Modify port mappings in .nizam.yaml

**MTU warnings**
- Common with VPN connections
- Configure Docker daemon MTU in `/etc/docker/daemon.json`
- Restart Docker daemon after changes

### Debug Mode
Enable verbose output for detailed check information:
```bash
nizam doctor --verbose
```
