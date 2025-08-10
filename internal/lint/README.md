# Lint Module Documentation

The lint module provides comprehensive configuration analysis to enforce best practices and identify potential issues in nizam service configurations.

## Overview

The linting system is designed to:
- Enforce Docker and container best practices
- Identify common configuration mistakes
- Provide actionable suggestions for improvements
- Support extensible rule framework
- Output both human-readable and JSON formats

## Architecture

### Core Components

#### `Report` Structure
```go
type Report struct {
    Findings []Finding `json:"findings"`
    Summary  Summary   `json:"summary"`
}
```

#### `Finding` Structure
```go
type Finding struct {
    Rule       string `json:"rule"`       // Rule identifier
    Severity   string `json:"severity"`   // "error", "warn", "info"
    Path       string `json:"path"`       // Configuration path
    Message    string `json:"message"`    // Description of issue
    Suggestion string `json:"suggestion"` // How to fix the issue
}
```

#### `Summary` Structure
```go
type Summary struct {
    Total    int `json:"total"`
    Errors   int `json:"errors"`
    Warnings int `json:"warnings"`
    Info     int `json:"info"`
}
```

## Implemented Rules

### `no-latest` (Error)
**Purpose**: Prevent usage of `:latest` image tags in production environments

**Rationale**: 
- Latest tags are mutable and can change unexpectedly
- Makes deployments non-reproducible
- Can lead to different behavior across environments

**Detection**: 
- Images ending with `:latest`
- Images without explicit tags (defaults to `:latest`)

**Example**:
```yaml
# ❌ Bad
services:
  web:
    image: nginx:latest  # Will trigger rule

  api:
    image: myapp         # Defaults to :latest, will trigger rule

# ✅ Good  
services:
  web:
    image: nginx:1.21.6  # Specific version

  api:
    image: myapp:v2.1.0  # Semantic version
```

**Fix Suggestion**: "pin to a specific tag, e.g. 'redis:7.2'"

### `ports-shape` (Error)
**Purpose**: Validate port mapping format and detect common mistakes

**Detection**:
- Invalid port format (non-numeric)
- Missing port mappings where expected
- Malformed host:container port syntax

**Example**:
```yaml
# ❌ Bad
services:
  web:
    ports: 
      - "abc:80"      # Invalid host port
      - "8080-80"     # Wrong separator

# ✅ Good
services:
  web:
    ports:
      - "8080:80"     # Correct format
      - "443:443"     # Host and container same
```

**Fix Suggestion**: "use format 'host:container', e.g. '8080:80'"

### `limits` (Warning)
**Purpose**: Recommend resource limits for consistent development environments

**Rationale**:
- Prevents resource exhaustion
- Makes environments more predictable
- Helps identify resource-hungry services
- Matches production-like constraints

**Detection**: Services without CPU or memory limits defined

**Example**:
```yaml
# ❌ Missing limits (will trigger warning)
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

**Fix Suggestion**: "add 'resources: { cpus: \"1.0\", memory: \"512m\" }'"

## Usage Examples

### Basic Usage
```bash
# Lint default configuration
nizam lint

# Lint specific file
nizam lint --file ./production.yaml

# JSON output for CI/CD integration
nizam lint --json
```

### Integration Examples

#### GitHub Actions
```yaml
- name: Lint nizam configuration
  run: |
    nizam lint --json > lint-results.json
    if [ $(jq '.summary.errors' lint-results.json) -gt 0 ]; then
      echo "Linting errors found"
      exit 1
    fi
```

#### Pre-commit Hook
```bash
#!/bin/bash
nizam lint --json | jq -e '.summary.errors == 0' > /dev/null
if [ $? -ne 0 ]; then
    echo "Configuration has linting errors. Run 'nizam lint' for details."
    exit 1
fi
```

## Extending the System

### Adding New Rules

1. **Create Rule Function**:
```go
// internal/lint/rules.go
func MyCustomRule(cfg *config.Config) *Report {
    report := &Report{}
    
    for serviceName, service := range cfg.Services {
        // Check your condition
        if violatesRule(service) {
            report.Findings = append(report.Findings, Finding{
                Rule:       "my-rule",
                Severity:   "warn",
                Path:       fmt.Sprintf("services.%s", serviceName),
                Message:    "Service violates custom rule",
                Suggestion: "Fix by doing X",
            })
        }
    }
    
    return report
}
```

2. **Register Rule**:
```go
// cmd/lint.go
rep := lint.Combine(
    lint.NoLatest(cfg),
    lint.PortsShape(cfg),
    lint.LimitsRecommended(cfg),
    lint.MyCustomRule(cfg), // Add your rule
)
```

### Rule Categories

#### Security Rules
- **secrets-in-env**: Detect hardcoded secrets in environment variables
- **privileged-containers**: Warn about privileged container usage
- **root-user**: Detect containers running as root

#### Performance Rules  
- **resource-ratios**: Validate CPU/memory ratios
- **volume-performance**: Check for performance-critical volume mounts
- **network-optimization**: Suggest network optimizations

#### Reliability Rules
- **health-checks**: Require health check configuration
- **restart-policies**: Validate restart policy settings
- **dependency-order**: Check service dependency ordering

### Rule Implementation Patterns

#### Simple Validation
```go
func ValidateFormat(cfg *config.Config) *Report {
    report := &Report{}
    
    for name, svc := range cfg.Services {
        if !isValidFormat(svc.SomeField) {
            report.AddFinding("format-check", "error", 
                fmt.Sprintf("services.%s.field", name),
                "Invalid format", "Use correct format")
        }
    }
    
    return report
}
```

#### Complex Analysis
```go
func CrossServiceValidation(cfg *config.Config) *Report {
    report := &Report{}
    
    // Analyze relationships between services
    dependencies := analyzeDependencies(cfg.Services)
    
    for service, deps := range dependencies {
        if hasCyclicDependency(deps) {
            report.AddFinding("cyclic-deps", "error",
                fmt.Sprintf("services.%s", service),
                "Cyclic dependency detected",
                "Restructure service dependencies")
        }
    }
    
    return report
}
```

## Configuration

### Rule-Specific Settings
Rules can be configured through the nizam configuration:

```yaml
# .nizam.yaml
lint:
  rules:
    no-latest:
      enabled: true
      exceptions: ["development-*"]  # Allow latest for dev images
    
    limits:
      enabled: true
      default_cpu: "0.5"
      default_memory: "256m"
      require_both: false  # Don't require both CPU and memory
    
    ports-shape:
      enabled: true
      allow_range: false  # Disallow port ranges
```

### Global Settings
```yaml
lint:
  severity_exit_codes:
    error: 1    # Exit code for errors
    warn: 0     # Don't fail on warnings
    info: 0     # Don't fail on info
  
  output:
    color: true
    verbose: false
```

## Output Formats

### Human-Readable Output
```
✖ services.web.image: image tag missing or ':latest' not allowed (no-latest)
  Fix: pin to a specific tag, e.g. 'nginx:1.21'

! services.database: consider setting CPU/memory limits (limits)  
  Fix: add 'resources: { cpus: "1.0", memory: "512m" }'
```

### JSON Output
```json
{
  "findings": [
    {
      "rule": "no-latest",
      "severity": "error", 
      "path": "services.web.image",
      "message": "image tag missing or ':latest' not allowed",
      "suggestion": "pin to a specific tag, e.g. 'nginx:1.21'"
    }
  ],
  "summary": {
    "total": 1,
    "errors": 1,
    "warnings": 0,
    "info": 0
  }
}
```

## Best Practices

### Rule Design
1. **Clear Messages**: Explain what's wrong and why
2. **Actionable Suggestions**: Provide specific fix instructions  
3. **Appropriate Severity**: Use errors for critical issues, warnings for recommendations
4. **Performance**: Keep rule execution fast for large configurations

### Testing Rules
```go
func TestMyRule(t *testing.T) {
    cfg := &config.Config{
        Services: map[string]config.Service{
            "test": {
                Image: "nginx:latest", // Should trigger rule
            },
        },
    }
    
    report := MyRule(cfg)
    
    assert.Len(t, report.Findings, 1)
    assert.Equal(t, "my-rule", report.Findings[0].Rule)
    assert.Equal(t, "error", report.Findings[0].Severity)
}
```

## Integration with CI/CD

### Exit Codes
- `0`: No issues or only warnings/info
- `1`: Errors found (configurable per severity)

### Automation-Friendly Features
- JSON output for parsing
- Configurable exit codes
- Rule filtering and selection
- Batch processing support

## Future Enhancements

### Planned Features
1. **Rule Profiles**: Predefined rule sets (development, staging, production)
2. **Custom Rule Plugins**: External rule loading system
3. **Fix Suggestions**: Automatic configuration fixes
4. **Historical Analysis**: Track lint results over time
5. **Integration Tests**: Validate rules against real configurations

### Advanced Rules
1. **Security Scanning**: Integration with security vulnerability databases
2. **Performance Analysis**: Resource usage predictions
3. **Compliance Checking**: Industry standard compliance (CIS, NIST)
4. **Cost Optimization**: Cloud resource cost analysis
