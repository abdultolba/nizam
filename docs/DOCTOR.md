# Doctor Features Documentation

The nizam doctor system provides comprehensive development environment validation and operational reliability tools. This document covers all doctor-related features and their usage.

## Overview

The doctor system consists of several interconnected tools designed to ensure your development environment is properly configured and your services are running reliably:

- **Environment Doctor** (`nizam doctor`) - Preflight environment validation
- **Configuration Validation** (`nizam validate`) - Config file syntax and structure checking  
- **Configuration Linting** (`nizam lint`) - Best practices enforcement
- **Service Readiness** (`nizam wait-for`) - Service availability checking
- **Retry Operations** (`nizam retry`) - Intelligent failure recovery
- **Self-Update** (`nizam update`) - Automatic tool updates
- **Shell Completion** (`nizam completion`) - Enhanced CLI productivity

## Environment Doctor (`nizam doctor`)

### Purpose
Comprehensive preflight checks to validate your Docker environment and detect common configuration issues before they cause problems.

### Quick Start
```bash
# Run all environment checks
nizam doctor

# Skip specific checks
nizam doctor --skip net.mtu,disk.free

# JSON output for automation
nizam doctor --json

# Attempt automatic fixes
nizam doctor --fix

# Verbose diagnostics
nizam doctor --verbose
```

### Available Checks

#### Docker Infrastructure
- **`docker.daemon`** - Verify Docker daemon connectivity
- **`docker.compose`** - Ensure Docker Compose plugin availability

#### System Resources  
- **`disk.free`** - Check available disk space (warns if <1GB)
- **`net.mtu`** - Validate network MTU configuration

#### Service Ports
- **`port.{PORT}`** - Dynamic port availability checks for all configured services

### Sample Output
```
✔ docker.daemon       Docker daemon responding
✔ docker.compose      Docker Compose plugin available
! net.mtu              non-standard MTU detected
  VPNs may lower MTU; if Docker networking is flaky, align MTU in daemon.json
✖ port.5432            port in use
  Change host port for service postgres in .nizam.yaml
  Or stop the process using the port

Summary: required_failed=1 advisory_failed=1
```

### Advanced Usage

#### Selective Check Execution
```bash
# Skip network and disk checks
nizam doctor --skip net.mtu,disk.free

# Skip all port checks
nizam doctor --skip port.*

# Only run Docker checks
nizam doctor --skip disk.free,net.mtu,port.*
```

#### Automation Integration
```bash
# CI/CD pipeline usage
nizam doctor --json | jq '.summary.required_failed == 0'

# Pre-commit hook
if ! nizam doctor --json >/dev/null 2>&1; then
    echo "Environment check failed. Run 'nizam doctor' for details."
    exit 1
fi
```

#### Fix Attempts
```bash
# Attempt automatic fixes where supported
nizam doctor --fix

# Verbose fix output
nizam doctor --fix --verbose
```

### Configuration
Environment variables for customization:
- `NIZAM_DOCTOR_TIMEOUT` - Override default check timeout
- `NIZAM_DOCTOR_CONCURRENCY` - Override default concurrency limit

## Configuration Validation (`nizam validate`)

### Purpose
Validate nizam configuration file syntax and structure before deployment or service startup.

### Usage
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

### Validation Features
- ✅ YAML syntax validation with detailed error reporting
- 🔍 Service structure verification
- 📋 Required field checking (services, image, etc.)
- 🔢 Profile validation
- 📊 Multiple output formats (human-readable, JSON)

### Sample Output
```bash
# Success
✔ Configuration is valid
  Profile: dev
  Services: 7

# Failure
✖ Configuration validation failed
  Error: services.postgres.image is required
  Line: 5, Column: 3
```

### Integration Examples

#### Pre-commit Hook
```bash
#!/bin/bash
# .git/hooks/pre-commit
nizam validate --strict
if [ $? -ne 0 ]; then
    echo "Configuration validation failed!"
    exit 1
fi
```

#### CI/CD Pipeline
```yaml
# GitHub Actions
- name: Validate Configuration
  run: |
    nizam validate --json --strict
    echo "Configuration is valid"
```

## Configuration Linting (`nizam lint`)

### Purpose
Analyze configurations for best practices, security issues, and potential problems using an extensible rule framework.

### Usage
```bash
# Lint default configuration
nizam lint

# Lint specific file
nizam lint --file ./config.yaml

# JSON output for CI/CD
nizam lint --json
```

### Built-in Rules

#### `no-latest` (Error)
**Problem:** `:latest` image tags are mutable and non-reproducible
**Detection:** Images without explicit tags or with `:latest` suffix
**Fix:** Pin to specific version tags

```yaml
# ❌ Problematic
services:
  web:
    image: nginx:latest
  api:
    image: myapp  # defaults to :latest

# ✅ Recommended  
services:
  web:
    image: nginx:1.21.6
  api:
    image: myapp:v2.1.0
```

#### `ports-shape` (Error)
**Problem:** Invalid port mapping format
**Detection:** Malformed host:container port syntax
**Fix:** Use correct `host:container` format

```yaml
# ❌ Problematic
services:
  web:
    ports:
      - "abc:80"    # Invalid host port
      - "8080-80"   # Wrong separator

# ✅ Recommended
services:
  web:
    ports:
      - "8080:80"
      - "443:443"
```

#### `limits` (Warning)
**Problem:** Missing resource limits can cause resource exhaustion
**Detection:** Services without CPU or memory limits
**Fix:** Add resource constraints for predictable behavior

```yaml
# ❌ No limits (will trigger warning)
services:
  database:
    image: postgres:14
    ports: ["5432:5432"]

# ✅ With limits
services:
  database:
    image: postgres:14
    ports: ["5432:5432"]
    resources:
      cpus: "1.0"
      memory: "512m"
```

### Sample Output
```
✖ services.web.image: image tag missing or ':latest' not allowed (no-latest)
  Fix: pin to a specific tag, e.g. 'nginx:1.21'

! services.database: consider setting CPU/memory limits (limits)
  Fix: add 'resources: { cpus: "1.0", memory: "512m" }'

Summary: 1 error, 1 warning
```

### CI/CD Integration
```bash
# Fail build on linting errors
nizam lint --json | jq -e '.summary.errors == 0'

# Generate lint report
nizam lint --json > lint-results.json
```

## Service Readiness (`nizam wait-for`)

### Purpose
Wait for services to become ready before proceeding with dependent operations, preventing race conditions in startup sequences.

### Usage
```bash
# Wait for specific service
nizam wait-for database

# Wait for multiple services  
nizam wait-for web database cache

# Wait for all configured services
nizam wait-for

# Custom timeout and check interval
nizam wait-for --timeout 60s --interval 2s database
```

### Readiness Check Types

#### Port Connectivity
TCP connection tests to verify service ports are accepting connections:
```yaml
services:
  postgres:
    ports: ["5432:5432"]  # Will check localhost:5432
```

#### HTTP Health Checks
HTTP endpoint availability testing:
```yaml  
services:
  web:
    health_check:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
```

#### Container Status
Docker container running state as fallback.

### Sample Output
```
Waiting for 3 service(s) to become ready (timeout: 30s)...
⏳ database: waiting for ports...
✔ cache: port 6379 is ready
⏳ web: waiting for health check...
✔ All services are ready (took 12.3s)
```

### Integration Patterns

#### Dependent Service Startup
```bash
# Start database first, then dependent services
nizam up postgres
nizam wait-for postgres --timeout 30s
nizam up web api
```

#### CI/CD Testing
```bash
# Ensure services are ready before running tests
nizam up
nizam wait-for --timeout 120s
npm test
```

## Retry Operations (`nizam retry`)

### Purpose
Retry failed operations with intelligent exponential backoff to handle transient failures in Docker operations, network issues, or resource conflicts.

### Usage
```bash
# Retry start command with defaults (3 attempts, 1s initial delay)
nizam retry start

# Custom retry configuration
nizam retry start --attempts 5 --delay 2s

# Retry specific services
nizam retry start web database --attempts 3
```

### Supported Operations
- **`start`** - Start services with retry logic
- **`stop`** - Stop services with retry logic  
- **`restart`** - Restart services with retry logic
- **`pull`** - Pull images with retry logic
- **`build`** - Build images with retry logic

### Retry Behavior
- **Exponential Backoff:** 1s → 2s → 4s → 8s → 16s
- **Progress Reporting:** Shows attempt numbers and wait times
- **Graceful Failure:** Clear error messages after all attempts fail

### Sample Output
```
Attempt 1/3: Running 'nizam start'
✖ Attempt 1 failed: port 5432 already in use
Waiting 1s before next attempt...
Attempt 2/3: Running 'nizam start'  
✖ Attempt 2 failed: port 5432 already in use
Waiting 2s before next attempt...
Attempt 3/3: Running 'nizam start'
✔ Command succeeded on attempt 3
```

### Use Cases
- **Network Issues:** Transient connectivity problems
- **Resource Conflicts:** Temporary port or resource unavailability
- **Image Pulls:** Unreliable registry connections
- **Service Dependencies:** Services starting in wrong order

## Self-Update (`nizam update`)

### Purpose
Keep nizam up-to-date with the latest features, bug fixes, and security updates through automatic GitHub releases integration.

### Usage
```bash
# Check for updates without installing
nizam update --check

# Update to latest stable version
nizam update

# Include prerelease versions
nizam update --prerelease
```

### Update Process
1. **Version Check:** Compare current version with GitHub releases
2. **Binary Detection:** Find appropriate binary for your platform
3. **Safe Download:** Download to temporary location
4. **Atomic Replace:** Replace current binary safely with rollback capability
5. **Verification:** Confirm update success

### Sample Output
```
Current version: v1.2.0
Latest version: v1.3.0
📦 Update available
Downloading v1.3.0...
✔ Successfully updated to v1.3.0
```

### Platform Support
- **Linux:** amd64, arm64
- **macOS:** amd64, arm64 (Apple Silicon)
- **Windows:** amd64

## Shell Completion (`nizam completion`)

### Purpose
Generate shell completion scripts for faster and more accurate command-line usage.

### Installation

#### Bash
```bash
# Temporary (current session)
source <(nizam completion bash)

# Permanent (add to ~/.bashrc)
echo 'source <(nizam completion bash)' >> ~/.bashrc
```

#### Zsh
```bash
# Temporary (current session)  
source <(nizam completion zsh)

# Permanent (add to ~/.zshrc)
echo 'source <(nizam completion zsh)' >> ~/.zshrc
```

#### Fish
```bash
# Temporary (current session)
nizam completion fish | source

# Permanent
nizam completion fish > ~/.config/fish/completions/nizam.fish
```

#### PowerShell
```powershell
# Current session
nizam completion powershell | Out-String | Invoke-Expression

# Permanent (add to profile)
nizam completion powershell >> $PROFILE
```

### Features
- **Command Completion:** All nizam commands and subcommands
- **Flag Completion:** Command-line options and their values
- **Service Names:** Dynamic completion of configured service names
- **File Paths:** Intelligent file and directory completion

## Workflow Integration

### Development Environment Setup
```bash
#!/bin/bash
# setup-dev.sh - Reliable development environment setup

echo "🔍 Checking environment..."
nizam doctor --fix || exit 1

echo "✅ Validating configuration..."  
nizam validate --strict || exit 1

echo "🔍 Linting configuration..."
nizam lint || echo "⚠️ Linting warnings detected"

echo "🚀 Starting services with retry..."
nizam retry start --attempts 3 || exit 1

echo "⏳ Waiting for services to be ready..."
nizam wait-for --timeout 60s || exit 1

echo "✅ Development environment ready!"
```

### CI/CD Pipeline Integration
```yaml
# .github/workflows/test.yml
name: Test Environment
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install nizam
        run: |
          # Install nizam binary
          curl -L https://github.com/user/nizam/releases/latest/download/nizam_linux_amd64 -o nizam
          chmod +x nizam
          sudo mv nizam /usr/local/bin/
          
      - name: Validate Environment
        run: |
          nizam doctor --json
          nizam validate --strict
          nizam lint --json
          
      - name: Start Services
        run: |
          nizam retry start --attempts 3
          nizam wait-for --timeout 120s
          
      - name: Run Tests
        run: |
          # Your test commands here
          npm test
          
      - name: Cleanup
        if: always()
        run: nizam down --force
```

### Production Deployment Checks
```bash
#!/bin/bash
# deploy-check.sh - Production readiness validation

echo "🔍 Production environment check..."
nizam doctor --json > doctor-report.json

if ! jq -e '.summary.required_failed == 0' doctor-report.json; then
    echo "❌ Critical environment issues detected"
    jq '.results[] | select(.status == "fail")' doctor-report.json
    exit 1
fi

echo "✅ Validating production config..."
nizam validate --file production.yaml --strict || exit 1

echo "🔍 Linting production config..."
nizam lint --file production.yaml --json > lint-report.json

if ! jq -e '.summary.errors == 0' lint-report.json; then
    echo "❌ Configuration errors detected"
    jq '.findings[] | select(.severity == "error")' lint-report.json
    exit 1
fi

echo "✅ Production deployment checks passed"
```

## Troubleshooting

### Common Doctor Issues

#### Docker Daemon Connection Failed
```
✖ docker.daemon        Docker daemon not responding
```
**Solutions:**
- Start Docker Desktop or Docker daemon
- Check Docker daemon socket permissions
- Verify Docker installation

#### Port Conflicts  
```
✖ port.5432            port in use
```
**Solutions:**
- Identify conflicting process: `lsof -i :5432`
- Use different host port in configuration
- Stop conflicting service

#### MTU Configuration Warnings
```
! net.mtu              non-standard MTU detected
```
**Solutions:**
- Common with VPN connections
- Configure Docker daemon MTU in `/etc/docker/daemon.json`:
  ```json
  {
    "mtu": 1350
  }
  ```
- Restart Docker daemon after changes

### Configuration Issues

#### YAML Syntax Errors
- Use proper indentation (spaces, not tabs)
- Validate YAML syntax with online validators
- Check for missing colons or brackets

#### Service Definition Problems
- Ensure required fields (image, ports) are present
- Validate port mapping format ("host:container")
- Check image tag format and availability

### Performance Considerations

#### Doctor Check Performance
- Use `--skip` to exclude unnecessary checks
- Set `NIZAM_DOCTOR_CONCURRENCY` for parallel execution
- Cache results for repeated runs in CI/CD

#### Large Configuration Files
- Lint incrementally during development
- Use profiles to separate environments
- Split large configurations into multiple files

## Best Practices

### Environment Validation
1. **Run doctor checks early** - Before any development work
2. **Automate in CI/CD** - Catch issues before deployment  
3. **Fix systematically** - Address errors before warnings
4. **Monitor trends** - Track recurring issues

### Configuration Management
1. **Validate continuously** - On every change
2. **Lint aggressively** - Enforce best practices early
3. **Version configurations** - Track changes over time
4. **Profile separation** - Different rules for different environments

### Service Dependencies
1. **Wait explicitly** - Don't assume services are ready
2. **Retry strategically** - Handle transient failures gracefully
3. **Timeout appropriately** - Balance reliability with speed
4. **Monitor readiness** - Track startup performance

### Tool Maintenance
1. **Update regularly** - Keep nizam current
2. **Enable completion** - Improve CLI productivity
3. **Document workflows** - Share knowledge with team
4. **Monitor usage** - Optimize based on actual patterns
