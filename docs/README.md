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
  - Configuration management (`init`, `validate`, `lint`, `add`, `remove`, `lint`, `templates`, `export`)
  - Data lifecycle management (`snapshot`, `psql`, `mysql`, `redis-cli`, `mongosh`)
  - Health & monitoring (`doctor`, `health`, `health-server`)
  - Development tools (`wait-for`, `retry`, `completion`, `update`)

### Interface Guides

- **[Health Check System](../README.md#health-check-system-)** - Health monitoring and dashboard
- **[Service Templates](../README.md#service-templates)** - Using and managing service templates
- **[Data Lifecycle Management](DATA_LIFECYCLE.md)** - Database snapshots and one-liner access tools
  - **[MongoDB Interface](MONGODB_SUPPORT.md)** - Describes MongoDB CLI and snapshot support implementation.

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
3. [Basic Commands](COMMANDS.md#core-operations) - Essential CLI commands

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

- **Database Snapshots** - Point-in-time backup and restore for PostgreSQL, MySQL, MongoDB, and Redis
- **One-liner Database Access** - Smart CLI tools with auto-resolved connections
- **Compression & Integrity** - zstd/gzip compression with SHA256 verification
- **Atomic Operations** - Safe creation and restoration with temporary files

### Advanced Features

- **Custom Templates** - Create and share your own service templates
- **Profile Support** - Multiple environment configurations
- **Docker Integration** - Native Docker Compose integration

## 🛠️ Architecture

<!-- ```
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
│   ├── MONGODB_SUPPORT.md # MongoDB snapshot & one-liner implementation
│   └── DOCTOR.md          # Doctor features documentation
└── README.md              # Main project documentation
``` -->

<pre>
<span style="color:#6af">nizam/</span>
├── <span style="color:#6af">cmd/</span>                    <span style="color:#888"># CLI commands implementation</span>
├── <span style="color:#6af">internal/</span>
│   ├── <span style="color:#6af">compress/</span>          <span style="color:#888"># Compression utilities (zstd, gzip)</span>
│   ├── <span style="color:#6af">config/</span>            <span style="color:#888"># Configuration parsing and validation</span>
│   ├── <span style="color:#6af">docker/</span>            <span style="color:#888"># Docker client wrapper</span>
│   ├── <span style="color:#6af">dockerx/</span>           <span style="color:#888"># Lightweight Docker execution</span>
│   ├── <span style="color:#6af">doctor/</span>            <span style="color:#888"># Environment checking system</span>
│   │   ├── <span style="color:#f9f">README.md</span>      <span style="color:#888"># Doctor module documentation</span>
│   │   └── <span style="color:#6af">checks/</span>        <span style="color:#888"># Individual check implementations</span>
│   ├── <span style="color:#6af">lint/</span>              <span style="color:#888"># Configuration linting system</span>
│   │   └── <span style="color:#f9f">README.md</span>      <span style="color:#888"># Lint module documentation</span>
│   ├── <span style="color:#6af">paths/</span>             <span style="color:#888"># Storage path management</span>
│   ├── <span style="color:#6af">resolve/</span>           <span style="color:#888"># Service resolution and detection</span>
│   ├── <span style="color:#6af">snapshot/</span>          <span style="color:#888"># Database snapshot engines</span>
│   └── <span style="color:#6af">version/</span>           <span style="color:#888"># Version management</span>
├── <span style="color:#6af">docs/</span>
│   ├── <span style="color:#f9f">README.md</span>          <span style="color:#888"># This documentation index</span>
│   ├── <span style="color:#f9f">COMMANDS.md</span>        <span style="color:#888"># Complete CLI reference</span>
│   ├── <span style="color:#f9f">DATA_LIFECYCLE.md</span>  <span style="color:#888"># Database snapshots &amp; one-liners</span>
│   ├── <span style="color:#f9f">MONGODB_SUPPORT.md</span> <span style="color:#888"># MongoDB snapshot &amp; one-liner implementation</span>
│   └── <span style="color:#f9f">DOCTOR.md</span>          <span style="color:#888"># Doctor features documentation</span>
└── <span style="color:#f9f">README.md</span>              <span style="color:#888"># Main project documentation</span>
</pre>

## 🤝 Contributing to Documentation

Contributions are welcome to improve nizam's documentation! Here's how you can help:

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

- **Issues** - Report bugs and request features on [Github](https://github.com/abdultolba/nizam/issues/new)
- **Discussions** - Ask questions and share ideas
- **Examples** - Check documentation for usage patterns
- **Contributing** - See the main [README](../README.md) for contribution guidelines

---

_This documentation is actively maintained and updated with each release. For the most current information, always refer to the latest version._
