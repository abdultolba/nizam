# Nizam Documentation

Welcome to the comprehensive documentation for nizam, the Local Structured Service Manager for Development Environments.

## 📚 Documentation Overview

### Getting Started
- **[README](../README.md)** - Complete project overview, features, and quick start guide
- **[Installation & Setup](../README.md#installation)** - How to install and configure nizam
- **[Configuration Guide](../README.md#configuration)** - Setting up your `.nizam.yaml` file

### Command Reference
- **[CLI Commands Documentation](COMMANDS.md)** - Complete reference for all nizam commands
  - Core operations (`up`, `down`, `status`, `logs`, `exec`)
  - Configuration management (`init`, `validate`, `lint`, `add`, `remove`)
  - Data lifecycle management (`snapshot`, `psql`, `mysql`, `redis-cli`)
  - Health & monitoring (`doctor`, `health`, `health-server`)
  - Development tools (`wait-for`, `retry`, `completion`, `update`)

### Interface Guides
- **[Interactive TUI Guide](../README.md#interactive-tui-terminal-user-interface)** - Terminal User Interface usage
- **[Health Check System](../README.md#health-check-system-)** - Health monitoring and dashboard
- **[Service Templates](../README.md#service-templates)** - Using and managing service templates
- **[Data Lifecycle Management](DATA_LIFECYCLE.md)** - Database snapshots and one-liner access tools

### Technical Documentation

#### Core Modules
- **[Doctor Module](../internal/doctor/README.md)** - Environment checking and validation system
  - Architecture and check implementation
  - Adding custom checks and auto-fix support
  - Performance considerations and configuration

- **[Lint Module](../internal/lint/README.md)** - Configuration linting and best practices
  - Rule framework and implementation patterns
  - Adding custom linting rules
  - CI/CD integration examples

#### Implementation Details
- **[Doctor Features Guide](DOCTOR.md)** - Complete doctor system documentation and usage
- **[Development Status](../README.md#development-status)** - Current progress and planned features

## 🚀 Quick Navigation

### For New Users
1. [Installation](../README.md#installation) - Get nizam installed
2. [Quick Start](../README.md#quick-start) - Your first nizam project
3. [TUI Guide](../README.md#interactive-tui-terminal-user-interface) - Visual interface overview
4. [Basic Commands](COMMANDS.md#core-operations) - Essential CLI commands

### For Developers
1. [Development Tools](../README.md#development--operations-tools-) - DevOps tooling overview
2. [Doctor Module](../internal/doctor/README.md) - Environment validation
3. [Lint System](../internal/lint/README.md) - Configuration best practices
4. [CLI Reference](COMMANDS.md) - Complete command documentation

### For Operations Teams
1. [Health Monitoring](../README.md#health-check-system-) - Service health tracking
2. [Environment Doctor](../README.md#environment-doctor-nizam-doctor) - Infrastructure validation  
3. [CI/CD Integration](COMMANDS.md#examples) - Automation examples
4. [Production Patterns](../README.md#development-workflow-integration) - Best practices

## 📖 Documentation Sections

### Core Features
- **Service Management** - Start, stop, and manage development services
- **Template System** - Reusable service configurations with 16+ built-in templates
- **Interactive Configuration** - Guided setup with validation and defaults
- **Health Monitoring** - Multi-interface health checking with web dashboard

### Development Tools
- **Environment Doctor** - Comprehensive preflight environment validation
- **Configuration Linting** - Best practices enforcement with extensible rules
- **Service Readiness** - Wait for service availability with multiple check types
- **Retry Operations** - Intelligent exponential backoff for failed operations
- **Self-Update** - Automatic updates from GitHub releases
- **Shell Completion** - Multi-shell completion support

### Data Lifecycle Management
- **Database Snapshots** - Point-in-time backup and restore for PostgreSQL, MySQL, and Redis
- **One-liner Database Access** - Smart CLI tools with auto-resolved connections
- **Compression & Integrity** - zstd/gzip compression with SHA256 verification
- **Atomic Operations** - Safe creation and restoration with temporary files

### Advanced Features
- **Interactive TUI** - Full-featured terminal interface with real-time monitoring
- **Custom Templates** - Create and share your own service templates
- **Profile Support** - Multiple environment configurations
- **Docker Integration** - Native Docker Compose integration

## 🛠️ Architecture

```
nizam/
├── cmd/                    # CLI commands implementation
├── internal/
│   ├── compress/          # Compression utilities (zstd, gzip)
│   ├── config/            # Configuration parsing and validation
│   ├── docker/            # Docker client wrapper
│   ├── dockerx/           # Lightweight Docker execution
│   ├── doctor/            # Environment checking system
│   │   ├── README.md      # Doctor module documentation
│   │   └── checks/        # Individual check implementations
│   ├── lint/              # Configuration linting system
│   │   └── README.md      # Lint module documentation
│   ├── paths/             # Storage path management
│   ├── resolve/           # Service resolution and detection
│   ├── snapshot/          # Database snapshot engines
│   └── version/           # Version management
├── docs/
│   ├── README.md          # This documentation index
│   ├── COMMANDS.md        # Complete CLI reference
│   ├── DATA_LIFECYCLE.md  # Database snapshots & one-liners
│   └── DOCTOR.md          # Doctor features documentation
└── README.md              # Main project documentation
```

## 🤝 Contributing to Documentation

We welcome contributions to improve our documentation! Here's how you can help:

### Documentation Types
- **User Guides** - Help users understand and use features
- **API Documentation** - Technical reference for developers
- **Examples** - Real-world usage patterns and integrations
- **Troubleshooting** - Common issues and solutions

### Documentation Standards
- **Clear Examples** - Include working code examples
- **Complete Context** - Provide necessary background information
- **Update Status** - Keep implementation status current
- **Cross-References** - Link to related documentation

### File Conventions
- **README.md files** - Overview and getting started information
- **Module READMEs** - Technical implementation details
- **COMMANDS.md** - Complete CLI reference
- **Integration examples** - Real-world usage patterns

## 📞 Support

- **Issues** - Report bugs and request features on GitHub
- **Discussions** - Ask questions and share ideas
- **Examples** - Check documentation for usage patterns
- **Contributing** - See the main README for contribution guidelines

---

*This documentation is actively maintained and updated with each release. For the most current information, always refer to the latest version.*
