# Seed Packs Guide

Seed packs are reusable database snapshots with enhanced metadata, examples, and documentation. They make it easy to share and distribute database seeds across projects and teams.

## Table of Contents

- [Overview](#overview)
- [Creating Seed Packs](#creating-seed-packs)
- [Installing Seed Packs](#installing-seed-packs)
- [Managing Seed Packs](#managing-seed-packs)
- [Template Integration](#template-integration)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

Seed packs extend the existing snapshot system with:

- **Rich Metadata**: Author, version, license, homepage, repository links
- **Documentation**: Descriptions, use cases, examples with queries
- **Schema Information**: Table/collection structures, indexes, sample data
- **Dependencies**: Required services or other seed packs
- **Tagging**: Organize and search packs by categories
- **Versioning**: Multiple versions of the same pack

### Architecture

```
.nizam/
├── snapshots/          # Raw snapshots (managed by snapshot system)
│   └── service/
│       └── timestamp/
└── seeds/              # Organized seed packs
    └── engine/         # postgres, mysql, redis, etc.
        └── pack-name/
            └── version/
                ├── seedpack.json    # Manifest with metadata
                ├── README.md        # Auto-generated documentation
                └── dump.sql         # Data files (copied from snapshot)
```

## Creating Seed Packs

### Prerequisites

1. **Existing Snapshot**: You need a snapshot of your service first
2. **Running Service**: The service should be running to create snapshots

### Create a Snapshot

First, create a snapshot of your service:

```bash
# Create a snapshot with a descriptive tag
nizam snapshot create postgres --tag "ecommerce-sample-data"

# Or create without a tag (uses latest)
nizam snapshot create postgres
```

### Create a Seed Pack

Convert the snapshot into a seed pack with enhanced metadata:

```bash
# Basic seed pack creation
nizam pack create postgres

# Create with custom metadata
nizam pack create postgres my-snapshot \
  --name "ecommerce-starter" \
  --display-name "E-commerce Starter Data" \
  --description "Sample e-commerce database with products, users, and orders" \
  --author "Your Name" \
  --version "1.0.0" \
  --license "MIT" \
  --homepage "https://github.com/yourorg/ecommerce-seeds" \
  --tag "ecommerce" \
  --tag "sample-data" \
  --tag "starter" \
  --use-case "Development and testing" \
  --use-case "Demo applications"
```

### Seed Pack Manifest

Each seed pack includes a `seedpack.json` manifest with comprehensive metadata:

```json
{
  "name": "ecommerce-starter",
  "displayName": "E-commerce Starter Data",
  "description": "Sample e-commerce database with products, users, and orders",
  "version": "1.0.0",
  "author": "Your Name",
  "license": "MIT",
  "homepage": "https://github.com/yourorg/ecommerce-seeds",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z",
  "engine": "postgres",
  "images": ["postgres:16"],
  "tags": ["ecommerce", "sample-data", "starter"],
  "dataSize": 2048576,
  "recordCount": 1500,
  "compression": "zstd",
  "useCases": [
    "Development and testing",
    "Demo applications"
  ],
  "examples": [
    {
      "title": "List all products",
      "description": "Get all products with their categories",
      "query": "SELECT p.name, p.price, c.name as category FROM products p JOIN categories c ON p.category_id = c.id;",
      "expected": "Returns product names, prices, and categories"
    }
  ],
  "dependencies": [
    {
      "name": "postgres",
      "type": "service",
      "version": "15+",
      "optional": false
    }
  ],
  "sourceSnapshot": {
    "service": "postgres",
    "engine": "postgres",
    "createdAt": "2024-01-15T10:25:00Z",
    "tag": "ecommerce-sample-data"
  }
}
```

## Installing Seed Packs

### List Available Packs

```bash
# List all seed packs
nizam pack list

# List packs for specific engine
nizam pack list postgres

# Search for packs
nizam pack search ecommerce
nizam pack search --tag "sample-data" --engine postgres
nizam pack search --author "Your Name"
```

### Install a Seed Pack

```bash
# Install latest version
nizam pack install postgres ecommerce-starter

# Install specific version
nizam pack install postgres ecommerce-starter@1.0.0

# Dry run to see what would be installed
nizam pack install postgres ecommerce-starter --dry-run

# Force install even if service has data
nizam pack install postgres ecommerce-starter --force
```

### Get Pack Information

```bash
# View detailed pack information
nizam pack info postgres ecommerce-starter

# This shows:
# - Description and metadata
# - Use cases and examples
# - Schema information
# - Dependencies
# - Installation instructions
```

## Managing Seed Packs

### Remove Seed Packs

```bash
# Remove specific version
nizam pack remove postgres ecommerce-starter --version 1.0.0

# Remove all versions
nizam pack remove postgres ecommerce-starter
```

### Update Pack Metadata

You can manually edit the `seedpack.json` file to update metadata, examples, or documentation.

## Template Integration

Templates can reference seed packs for automatic installation during service creation.

### Template with Seed Packs

```yaml
name: "ecommerce-dev"
description: "E-commerce development environment with sample data"
tags: ["development", "ecommerce"]
service:
  image: "postgres:16"
  ports: ["5432:5432"]
  environment:
    POSTGRES_USER: "admin"
    POSTGRES_PASSWORD: "password"
    POSTGRES_DB: "ecommerce"
  volume: "ecommerce_data"
seedPacks:
  - name: "ecommerce-starter"
    version: "1.0.0"
    description: "Sample e-commerce data for development"
    autoInstall: true
    optional: false
  - name: "analytics-sample"
    description: "Additional analytics data"
    autoInstall: false
    optional: true
variables:
  - name: "DB_USER"
    description: "Database username"
    default: "admin"
    required: true
```

### Auto-Installation

When creating a service from a template with seed packs:

```bash
nizam create ecommerce-dev mystore
# This will:
# 1. Create the service
# 2. Wait for it to be healthy
# 3. Automatically install required seed packs
# 4. Prompt for optional seed packs
```

## Best Practices

### Creating Quality Seed Packs

1. **Descriptive Names**: Use clear, descriptive names for your packs
2. **Good Documentation**: Include comprehensive descriptions and use cases
3. **Useful Examples**: Provide query examples that demonstrate the data
4. **Realistic Data**: Use realistic but anonymized sample data
5. **Consistent Versioning**: Follow semantic versioning (1.0.0, 1.1.0, 2.0.0)
6. **Proper Tagging**: Use relevant tags for discoverability

### Data Considerations

1. **Size**: Keep packs reasonably sized (< 100MB for most use cases)
2. **Privacy**: Never include real user data or sensitive information
3. **Licensing**: Clearly specify the license for your data
4. **Dependencies**: Document any required services or other packs

### Organization

1. **Naming Convention**: Use consistent naming like `{domain}-{type}` (e.g., `blog-starter`, `ecommerce-demo`)
2. **Versioning**: Update versions when making significant changes
3. **Documentation**: Keep README files up to date
4. **Testing**: Test your seed packs before sharing

## Examples

### E-commerce Starter Pack

A comprehensive e-commerce database with:
- Users and authentication
- Products and categories
- Shopping carts and orders
- Payment information (anonymized)

```bash
nizam pack create postgres ecommerce-snapshot \
  --name "ecommerce-starter" \
  --display-name "E-commerce Starter Pack" \
  --description "Complete e-commerce database with users, products, orders" \
  --author "E-commerce Team" \
  --tag "ecommerce" \
  --tag "starter" \
  --tag "demo" \
  --use-case "Development and testing" \
  --use-case "Demo applications" \
  --use-case "Training and tutorials"
```

### Blog Content Pack

A simple blog database with:
- Posts and comments
- Users and authors
- Categories and tags

```bash
nizam pack create postgres blog-snapshot \
  --name "blog-content" \
  --display-name "Blog Content Pack" \
  --description "Sample blog with posts, comments, and users" \
  --tag "blog" \
  --tag "cms" \
  --tag "content"
```

### Analytics Sample Pack

Analytics data for testing:
- Events and metrics
- User behavior data
- Time-series data

```bash
nizam pack create clickhouse analytics-snapshot \
  --name "analytics-sample" \
  --display-name "Analytics Sample Data" \
  --description "Sample analytics data for testing and development" \
  --tag "analytics" \
  --tag "olap" \
  --tag "time-series"
```

### Redis Cache Pack

Redis data for session management:
- Session data
- Cache entries
- Rate limiting data

```bash
nizam pack create redis session-snapshot \
  --name "session-cache" \
  --display-name "Session Cache Pack" \
  --description "Sample session and cache data" \
  --tag "cache" \
  --tag "session" \
  --tag "redis"
```

## Advanced Usage

### Custom Schema Information

You can manually add detailed schema information to your seed pack manifest:

```json
{
  "schema": {
    "tables": [
      {
        "name": "users",
        "description": "User accounts",
        "rowCount": 100,
        "columns": [
          {
            "name": "id",
            "type": "integer",
            "primaryKey": true,
            "nullable": false,
            "description": "User ID"
          },
          {
            "name": "email",
            "type": "varchar(255)",
            "nullable": false,
            "description": "User email address"
          }
        ],
        "indexes": [
          {
            "name": "idx_users_email",
            "columns": ["email"],
            "unique": true
          }
        ]
      }
    ]
  }
}
```

### Multi-Engine Support

Some seed packs might work with multiple database engines:

```json
{
  "name": "user-data",
  "engine": "postgres",
  "images": ["postgres:15", "postgres:16"],
  "dependencies": [
    {
      "name": "postgres",
      "type": "service",
      "version": "15+",
      "optional": false
    }
  ]
}
```

### Sharing Seed Packs

1. **Export**: Copy the entire pack directory structure
2. **Version Control**: Store packs in Git repositories
3. **Documentation**: Include comprehensive README files
4. **Testing**: Verify packs work across different environments

## Troubleshooting

### Common Issues

1. **Pack Not Found**: Check engine name and pack availability
   ```bash
   nizam pack list postgres  # List available postgres packs
   ```

2. **Installation Fails**: Check service is running and healthy
   ```bash
   nizam status postgres  # Check service status
   ```

3. **Permission Errors**: Ensure proper file permissions in .nizam directory

4. **Size Issues**: Large packs may take time to install
   ```bash
   nizam pack info postgres large-pack  # Check pack size
   ```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
export LOG_LEVEL=debug
nizam pack install postgres my-pack
```

## Migration from Snapshots

If you have existing snapshots, you can convert them to seed packs:

```bash
# List existing snapshots
nizam snapshot list postgres

# Convert a snapshot to a seed pack
nizam pack create postgres existing-snapshot-tag \
  --name "converted-pack" \
  --description "Converted from snapshot"
```

This guide covers the complete seed pack system in Nizam. Seed packs make it easy to share, version, and manage database seeds across your development workflow.
