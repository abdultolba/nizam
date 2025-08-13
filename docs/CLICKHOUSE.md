# ClickHouse Template Guide

The ClickHouse template provides a complete OLAP database setup for analytics workloads. ClickHouse is a column-oriented database management system optimized for online analytical processing (OLAP) and is perfect for real-time analytics on large datasets.

## Quick Start

```bash
# Add ClickHouse service with interactive configuration
nizam add clickhouse

# Add ClickHouse service using default values
nizam add clickhouse --defaults

# Add ClickHouse with custom service name
nizam add clickhouse --name analytics-db
```

## Template Configuration

The ClickHouse template includes several customizable variables:

### Core Configuration
- **VERSION** (default: `24.1`): ClickHouse server version
- **DB_USER** (default: `admin`): Database username (required)
- **DB_PASSWORD** (default: `password`): Database password (required)  
- **DB_NAME** (default: `analytics`): Initial database to create (required)

### Network Ports
- **HTTP_PORT** (default: `8123`): HTTP interface for queries and administration
- **NATIVE_PORT** (default: `9000`): Native TCP interface for client connections
- **MYSQL_PORT** (default: `9004`): MySQL compatibility interface

### Storage
- **VOLUME_NAME** (default: `clickhousedata`): Docker volume for data persistence

## Default Configuration

When using defaults, the template creates:

```yaml
clickhouse:
  image: clickhouse/clickhouse-server:24.1
  ports:
    - "8123:8123"    # HTTP interface
    - "9000:9000"    # Native TCP interface  
    - "9004:9004"    # MySQL compatibility interface
  env:
    CLICKHOUSE_DB: analytics
    CLICKHOUSE_USER: admin
    CLICKHOUSE_PASSWORD: password
    CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT: "1"
  volume: clickhousedata
  health_check:
    test: ["CMD", "clickhouse-client", "--host=localhost", "--port=9000", "--user=admin", "--password=password", "--query=SELECT 1"]
    interval: 30s
    timeout: 10s
    retries: 3
```

## Connecting to ClickHouse

### HTTP Interface (Port 8123)
The HTTP interface is great for web applications and REST API interactions:

```bash
# Basic query via HTTP
curl 'http://localhost:8123/' -d 'SELECT version()'

# Query with authentication
curl 'http://admin:password@localhost:8123/' -d 'SELECT * FROM system.databases'
```

### Native TCP Interface (Port 9000)
Use the native interface for high-performance client connections:

```bash
# Connect using clickhouse-client (if installed locally)
clickhouse-client --host localhost --port 9000 --user admin --password password

# Or connect via nizam (when available)
nizam exec clickhouse clickhouse-client --user admin --password password
```

### MySQL Compatibility Interface (Port 9004)
Connect using MySQL clients for familiar tooling:

```bash
# Connect using mysql client
mysql -h localhost -P 9004 -u admin -p

# Connect using any MySQL-compatible tool
# Note: Some MySQL features may not be fully supported
```

## Sample Queries

Once connected, try these sample analytics queries:

```sql
-- Create a sample table
CREATE TABLE events (
    timestamp DateTime,
    user_id UInt32,
    event_type String,
    value Float64
) ENGINE = MergeTree()
ORDER BY (timestamp, user_id);

-- Insert sample data
INSERT INTO events VALUES 
    ('2024-01-01 10:00:00', 1, 'click', 1.0),
    ('2024-01-01 10:05:00', 2, 'view', 2.5),
    ('2024-01-01 10:10:00', 1, 'purchase', 99.99);

-- Analytics query: events per hour
SELECT 
    toStartOfHour(timestamp) as hour,
    count() as events,
    uniq(user_id) as unique_users,
    sum(value) as total_value
FROM events 
GROUP BY hour 
ORDER BY hour;
```

## Use Cases

ClickHouse is perfect for:

- **Real-time Analytics**: Fast aggregations on large datasets
- **Time Series Data**: IoT metrics, application logs, user events
- **Business Intelligence**: OLAP cubes, reporting dashboards  
- **Log Analysis**: System logs, application traces, audit trails
- **E-commerce Analytics**: User behavior, sales metrics, inventory tracking

## Integration with Other Services

ClickHouse works great alongside other nizam services:

```bash
# Setup complete analytics stack
nizam add clickhouse --name analytics-db
nizam add grafana       # For visualization
nizam add prometheus    # For metrics collection
nizam add redis --name cache  # For fast query caching

# Start the entire analytics stack
nizam up analytics-db grafana prometheus cache
```

## Performance Tips

1. **Use appropriate storage engines**: MergeTree for analytics, Log for temporary data
2. **Optimize table ordering**: Choose ORDER BY columns based on query patterns
3. **Leverage materialized views**: Pre-aggregate data for common queries
4. **Use proper data types**: Choose efficient types (UInt32 vs String for IDs)
5. **Partition large tables**: Use PARTITION BY for time-based data

## Troubleshooting

Check service health:
```bash
nizam status clickhouse
nizam logs clickhouse
```

Common issues:
- **Connection refused**: Check if ports are available and not conflicting
- **Authentication failed**: Verify CLICKHOUSE_USER and CLICKHOUSE_PASSWORD
- **Out of memory**: Increase Docker memory limits for large datasets

## Further Reading

- [ClickHouse Official Documentation](https://clickhouse.com/docs)
- [ClickHouse SQL Reference](https://clickhouse.com/docs/en/sql-reference/)
- [Performance Optimization Guide](https://clickhouse.com/docs/en/operations/performance/)
