# MagicORM Monitoring System

A comprehensive monitoring system for MagicORM that provides detailed metrics for ORM operations, validation, and database execution.

## Overview

The monitoring system provides:

1. **Unified Monitoring Framework**: Centralized management of all monitoring components
2. **ORM Operation Monitoring**: Track CRUD operations, transactions, and performance
3. **Validation System Monitoring**: Monitor validation performance and cache effectiveness
4. **Database Execution Monitoring**: Track database queries, connections, and transactions
5. **Performance Metrics**: Latency, throughput, error rates, and resource usage
6. **Export Capabilities**: Prometheus and JSON format exports

## Architecture

```
monitoring/
├── core/                    # Core monitoring components
│   ├── config.go           # Configuration system
│   ├── collector.go        # Metrics collector
│   └── exporter.go         # Metrics exporter
├── orm/                    # ORM monitoring
│   ├── metrics.go          # ORM metrics definition
│   ├── decorator.go        # ORM monitoring decorator
│   └── integration.go      # ORM integration utilities
├── validation/             # Validation monitoring
│   └── adapter.go          # Validation monitoring adapter
├── database/               # Database monitoring
│   ├── adapter.go          # Database monitoring adapter
│   └── factory.go          # Database monitoring factory
├── unified/                # Unified management
│   ├── manager.go          # Monitoring manager
│   └── factory.go          # Monitoring factory
└── example/                # Usage examples
    └── example.go          # Comprehensive example
```

## Quick Start

### Basic Usage

```go
import "github.com/muidea/magicOrm/monitoring/unified"

// Create a default monitoring manager
manager := unified.DefaultMonitoringManager()

// Start monitoring
if err := manager.Start(); err != nil {
    log.Fatal(err)
}
defer manager.Stop()

// Get metrics
metrics := manager.GetMetrics()
```

### ORM Monitoring

```go
import "github.com/muidea/magicOrm/monitoring/orm"

// Create ORM monitor
config := core.DefaultMonitoringConfig()
collector := core.NewCollector(&config)
ormMonitor := orm.NewORMMonitor(collector, &config)

// Record ORM operations
ormMonitor.RecordInsert("User", true, 150*time.Millisecond, nil)
ormMonitor.RecordQuery("Product", true, 200*time.Millisecond, nil)
```

### Database Monitoring

```go
import "github.com/muidea/magicOrm/monitoring/database"

// Create database monitor
dbMonitor := database.NewDatabaseMonitor(collector, &config)

// Record database operations
dbMonitor.RecordQuery("postgresql", "select", true, 200*time.Millisecond, 10)
dbMonitor.RecordTransaction("mysql", "begin", true, 50*time.Millisecond)
```

### Validation Monitoring

```go
import "github.com/muidea/magicOrm/monitoring/validation"

// Create validation monitor
validationMonitor := validation.NewValidationMonitor(collector, &config)

// Record validation operations
validationMonitor.RecordValidation(
    "validate_user",
    "User",
    validation.ScenarioInsert,
    50*time.Millisecond,
    nil,
    map[string]string{"field_count": "5"},
)
```

## Configuration

### Default Configuration

```go
config := core.DefaultMonitoringConfig()
```

### Environment-Specific Configurations

```go
// Development environment (verbose, high sampling)
devConfig := core.DevelopmentConfig()

// Production environment (balanced)
prodConfig := core.ProductionConfig()

// High-load environment (minimal overhead)
highLoadConfig := core.HighLoadConfig()
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `Enabled` | Enable/disable monitoring | `true` |
| `SamplingRate` | Percentage of operations to sample | `1.0` (100%) |
| `DetailLevel` | Detail level (basic/detailed/verbose) | `detailed` |
| `EnableORM` | Enable ORM monitoring | `true` |
| `EnableValidation` | Enable validation monitoring | `true` |
| `EnableDatabase` | Enable database monitoring | `true` |
| `EnableCache` | Enable monitoring cache | `true` |
| `AsyncCollection` | Use async metric collection | `true` |
| `CollectionInterval` | Collection interval | `60s` |
| `RetentionPeriod` | Metric retention period | `24h` |

## Metrics

### ORM Metrics

- `orm_operations_total`: Total ORM operations
- `orm_operation_duration_seconds`: ORM operation duration
- `orm_errors_total`: ORM operation errors
- `orm_transactions_total`: ORM transactions
- `orm_cache_hits_total`: ORM cache hits
- `orm_cache_misses_total`: ORM cache misses

### Validation Metrics

- `validation_operations_total`: Validation operations
- `validation_duration_seconds`: Validation duration
- `validation_errors_total`: Validation errors
- `validation_cache_hits_total`: Validation cache hits
- `validation_cache_misses_total`: Validation cache misses
- `validation_layer_performance_seconds`: Layer performance

### Database Metrics

- `database_connections_total`: Database connections
- `database_queries_total`: Database queries
- `database_query_duration_seconds`: Query duration
- `database_transactions_total`: Database transactions
- `database_executions_total`: SQL executions
- `database_errors_total`: Database errors
- `database_connections_active`: Active connections
- `database_connections_idle`: Idle connections

## Export

### Prometheus Export

```go
config := core.DefaultMonitoringConfig()
config.ExportConfig.Enabled = true
config.ExportConfig.Port = 9090

manager := unified.NewMonitoringManager(&config)
manager.Start()

// Metrics available at http://localhost:9090/metrics
```

### JSON Export

```go
// Get metrics as JSON
metrics := manager.GetMetrics()
jsonData, _ := json.Marshal(metrics)
```

## Integration with Existing Code

### Wrap Existing ORM

```go
import "github.com/muidea/magicOrm/monitoring/orm"

// Create monitored ORM
monitoredOrm := orm.NewMonitoredOrm(existingOrm, ormMonitor, &config)

// Use as normal
result, err := monitoredOrm.Insert(userModel)
```

### Wrap Database Executor

```go
import "github.com/muidea/magicOrm/monitoring/database"

// Create monitored executor
monitoredExecutor := database.NewMonitoredExecutor(
    existingExecutor,
    dbMonitor,
    "postgresql",
)

// Use as normal
rows, err := monitoredExecutor.Query("SELECT * FROM users", true)
```

## Performance Considerations

### Sampling

Control monitoring overhead with sampling:

```go
config := core.HighLoadConfig()
config.SamplingRate = 0.1 // Sample 10% of operations
```

### Async Collection

Reduce impact on application performance:

```go
config.AsyncCollection = true
config.CollectionInterval = 30 * time.Second
```

### Cache Configuration

```go
config.EnableCache = true
config.CacheTTL = 5 * time.Minute
```

## Testing

Run monitoring tests:

```bash
# Run all monitoring tests
go test ./monitoring/... -v

# Run specific component tests
go test ./monitoring/core/... -v
go test ./monitoring/orm/... -v
go test ./monitoring/database/... -v
```

## Examples

See the `example/` directory for comprehensive usage examples:

```bash
cd monitoring/example
go run example.go
```

## Best Practices

1. **Start Simple**: Begin with default configuration
2. **Monitor Production**: Use production configuration for live systems
3. **Sample Appropriately**: Adjust sampling rate based on load
4. **Use Labels**: Add custom labels for better metric organization
5. **Regular Cleanup**: Enable metric retention to prevent memory issues
6. **Secure Exports**: Enable TLS and authentication for production exports

## Troubleshooting

### Common Issues

1. **High Memory Usage**: Reduce retention period or increase sampling rate
2. **Performance Impact**: Enable async collection or reduce detail level
3. **Missing Metrics**: Check if monitoring is enabled for specific components
4. **Export Issues**: Verify port availability and firewall settings

### Debugging

```go
// Get detailed statistics
stats := manager.GetManagerStats()
fmt.Printf("%+v\n", stats)

// Check component status
fmt.Printf("ORM monitoring: %v\n", manager.GetORMMonitor() != nil)
fmt.Printf("Validation monitoring: %v\n", manager.GetValidationMonitor() != nil)
fmt.Printf("Database monitoring: %v\n", manager.GetDatabaseMonitor() != nil)
```

## API Reference

### Core Components

- `core.Collector`: Central metrics collection
- `core.Exporter`: Metrics export (Prometheus/JSON)
- `core.MonitoringConfig`: Configuration management

### Monitoring Adapters

- `orm.ORMMonitor`: ORM operation monitoring
- `validation.ValidationMonitor`: Validation system monitoring
- `database.DatabaseMonitor`: Database execution monitoring

### Unified Management

- `unified.MonitoringManager`: Unified monitoring management
- `unified.MonitoringFactory`: Component factory

## License

Part of the MagicORM project. See main project LICENSE for details.