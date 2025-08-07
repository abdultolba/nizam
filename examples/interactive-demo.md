# Interactive Template Variables Demo

This document demonstrates how to use the new interactive template variables feature in Nizam.

## Overview

The following templates now support customizable variables with interactive prompts:

- **postgres**: PostgreSQL database with custom credentials, ports, and volumes
- **mysql**: MySQL database with custom user/root credentials, database name, and ports  
- **redis**: Redis cache with optional password, custom port, and volume
- **mongodb**: MongoDB with custom root credentials and port
- **rabbitmq**: RabbitMQ with custom credentials and ports for AMQP + management UI

## Usage Examples

### 1. Interactive Mode (Default)
```bash
# Add PostgreSQL with interactive configuration
nizam add postgres

# You'll be prompted for:
# - DB_USER (default: user)
# - DB_PASSWORD (default: password) 
# - DB_NAME (default: myapp)
# - PORT (default: 5432)
# - VOLUME_NAME (default: pgdata)
```

### 2. Default Values Mode
```bash
# Skip prompts and use all default values
nizam add postgres --defaults
```

### 3. Custom Service Names
```bash
# Add with custom name and interactive config
nizam add mysql --name production-db

# Add with custom name and defaults
nizam add redis --name cache --defaults
```

## Interactive Prompt Features

When adding a template with variables in interactive mode, you'll see:

- **Clear variable descriptions** with purpose and usage
- **Default value suggestions** shown in brackets `[default]`
- **Required field indicators** marked with `*`
- **Type validation** (e.g., port numbers 1-65535)
- **Pattern validation** where applicable
- **Easy defaults** - just press Enter to use the default value

## Example Interactive Session

```
⚙️  Configuring template 'postgres' for service 'mydb'
📝 Please provide values for the following variables:

  DB_USER *: PostgreSQL username [user]
    Type: string
    > myuser
    ✅ DB_USER = myuser

  DB_PASSWORD *: PostgreSQL password [password]
    Type: string
    > securepass123
    ✅ DB_PASSWORD = securepass123

  DB_NAME *: Database name to create [myapp]
    Type: string
    > 
    ✅ DB_NAME = myapp

  PORT: Host port to bind PostgreSQL [5432]
    Type: port (pattern: ^[1-9][0-9]{0,4}$)
    > 5433
    ✅ PORT = 5433

  VOLUME_NAME: Docker volume name for data persistence [pgdata]
    Type: string
    > 
    ✅ VOLUME_NAME = pgdata

✅ Template configured successfully!

✅ Added service 'mydb' from template 'postgres'
📝 Configuration saved to .nizam.yaml

📋 Service Details:
   Image: postgres:16
   Ports: [5433:5432]
   Environment variables: 3 configured
   Volume: pgdata

💡 Template variables have been configured interactively

🚀 Run 'nizam up mydb' to start the service
```

## Validation Features

- **Port validation**: Ensures ports are between 1-65535
- **Required fields**: Won't accept empty values for required variables
- **Type checking**: Validates integers, booleans, etc.
- **Pattern matching**: Regex validation where specified

## Template Variable Details

### PostgreSQL
- `DB_USER` (required): Database username
- `DB_PASSWORD` (required): Database password  
- `DB_NAME` (required): Database name to create
- `PORT` (optional): Host port binding (default: 5432)
- `VOLUME_NAME` (optional): Docker volume name (default: pgdata)

### MySQL
- `DB_USER` (required): MySQL username
- `DB_PASSWORD` (required): MySQL user password
- `ROOT_PASSWORD` (required): MySQL root password
- `DB_NAME` (required): Database name to create
- `PORT` (optional): Host port binding (default: 3306)
- `VOLUME_NAME` (optional): Docker volume name (default: mysqldata)

### Redis
- `VERSION` (optional): Redis version (default: 7)
- `PORT` (optional): Host port binding (default: 6379)
- `PASSWORD` (optional): Redis password (empty = no auth)
- `VOLUME_NAME` (optional): Docker volume name (default: redisdata)

### MongoDB
- `VERSION` (optional): MongoDB version (default: 7)
- `ROOT_USERNAME` (required): MongoDB root username (default: admin)
- `ROOT_PASSWORD` (required): MongoDB root password (default: password)
- `PORT` (optional): Host port binding (default: 27017)
- `VOLUME_NAME` (optional): Docker volume name (default: mongodata)

### RabbitMQ
- `VERSION` (optional): RabbitMQ version (default: 3-management)
- `DEFAULT_USER` (required): RabbitMQ username (default: admin)
- `DEFAULT_PASS` (required): RabbitMQ password (default: password)
- `AMQP_PORT` (optional): AMQP protocol port (default: 5672)
- `MANAGEMENT_PORT` (optional): Management UI port (default: 15672)
- `VOLUME_NAME` (optional): Docker volume name (default: rabbitmqdata)
