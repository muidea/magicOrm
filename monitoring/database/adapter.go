package database

import (
	"context"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/monitoring/core"
)

// DatabaseMonitor monitors database executor operations
type DatabaseMonitor struct {
	collector *core.Collector
	config    *core.MonitoringConfig
}

// NewDatabaseMonitor creates a new database monitor
func NewDatabaseMonitor(collector *core.Collector, config *core.MonitoringConfig) *DatabaseMonitor {
	if collector == nil {
		collector = core.NewCollector(config)
	}

	// Register database-specific metrics
	registerDatabaseMetrics(collector)

	return &DatabaseMonitor{
		collector: collector,
		config:    config,
	}
}

// registerDatabaseMetrics registers all database-specific metrics
func registerDatabaseMetrics(collector *core.Collector) {
	// Connection pool metrics
	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_connections_total",
		Type:       core.CounterMetric,
		Help:       "Total number of database connections",
		LabelNames: []string{"database_type", "operation"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_connections_active",
		Type:       core.GaugeMetric,
		Help:       "Number of active database connections",
		LabelNames: []string{"database_type"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_connections_idle",
		Type:       core.GaugeMetric,
		Help:       "Number of idle database connections",
		LabelNames: []string{"database_type"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_connection_errors_total",
		Type:       core.CounterMetric,
		Help:       "Total number of database connection errors",
		LabelNames: []string{"database_type", "error_type"},
	})

	// Query execution metrics
	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_queries_total",
		Type:       core.CounterMetric,
		Help:       "Total number of database queries",
		LabelNames: []string{"database_type", "query_type", "success"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_query_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Duration of database queries in seconds",
		LabelNames: []string{"database_type", "query_type"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_query_rows_processed",
		Type:       core.HistogramMetric,
		Help:       "Number of rows processed by queries",
		LabelNames: []string{"database_type", "query_type"},
	})

	// Transaction metrics
	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_transactions_total",
		Type:       core.CounterMetric,
		Help:       "Total number of database transactions",
		LabelNames: []string{"database_type", "operation", "success"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_transaction_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Duration of database transactions in seconds",
		LabelNames: []string{"database_type", "operation"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_transaction_active",
		Type:       core.GaugeMetric,
		Help:       "Number of active database transactions",
		LabelNames: []string{"database_type"},
	})

	// SQL execution metrics
	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_executions_total",
		Type:       core.CounterMetric,
		Help:       "Total number of SQL executions",
		LabelNames: []string{"database_type", "operation", "success"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_execution_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Duration of SQL executions in seconds",
		LabelNames: []string{"database_type", "operation"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_execution_rows_affected",
		Type:       core.HistogramMetric,
		Help:       "Number of rows affected by SQL executions",
		LabelNames: []string{"database_type", "operation"},
	})

	// Resource usage metrics
	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_memory_bytes",
		Type:       core.GaugeMetric,
		Help:       "Database memory usage in bytes",
		LabelNames: []string{"database_type"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_cpu_seconds_total",
		Type:       core.CounterMetric,
		Help:       "Total CPU time used by database operations",
		LabelNames: []string{"database_type"},
	})

	// Error metrics
	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_errors_total",
		Type:       core.CounterMetric,
		Help:       "Total number of database errors",
		LabelNames: []string{"database_type", "error_type", "operation"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_timeouts_total",
		Type:       core.CounterMetric,
		Help:       "Total number of database timeouts",
		LabelNames: []string{"database_type", "operation"},
	})

	// Performance metrics
	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_latency_seconds",
		Type:       core.HistogramMetric,
		Help:       "Database operation latency in seconds",
		LabelNames: []string{"database_type", "operation"},
	})

	collector.RegisterDefinition(core.MetricDefinition{
		Name:       "database_throughput_ops_per_second",
		Type:       core.GaugeMetric,
		Help:       "Database operations per second",
		LabelNames: []string{"database_type", "operation"},
	})
}

// RecordConnection records a database connection event
func (m *DatabaseMonitor) RecordConnection(databaseType string, operation string, success bool, duration time.Duration) {
	if !m.shouldSample() {
		return
	}

	labels := map[string]string{
		"database_type": databaseType,
		"operation":     operation,
	}

	// Record connection count
	m.collector.Increment("database_connections_total", labels)

	// Record connection duration
	if success {
		m.collector.RecordDuration("database_query_duration_seconds", duration, labels)
	}
}

// RecordQuery records a database query execution
func (m *DatabaseMonitor) RecordQuery(databaseType string, queryType string, success bool, duration time.Duration, rowsProcessed int64) {
	if !m.shouldSample() {
		return
	}

	labels := map[string]string{
		"database_type": databaseType,
		"query_type":    queryType,
		"success":       boolToString(success),
	}

	// Record query count
	m.collector.Increment("database_queries_total", labels)

	// Record query duration
	if success {
		m.collector.RecordDuration("database_query_duration_seconds", duration, labels)

		// Record rows processed
		m.collector.Record("database_query_rows_processed", float64(rowsProcessed), map[string]string{
			"database_type": databaseType,
			"query_type":    queryType,
		})
	}
}

// RecordTransaction records a database transaction
func (m *DatabaseMonitor) RecordTransaction(databaseType string, operation string, success bool, duration time.Duration) {
	if !m.shouldSample() {
		return
	}

	labels := map[string]string{
		"database_type": databaseType,
		"operation":     operation,
		"success":       boolToString(success),
	}

	// Record transaction count
	m.collector.Increment("database_transactions_total", labels)

	// Record transaction duration
	if success {
		m.collector.RecordDuration("database_transaction_duration_seconds", duration, labels)
	}
}

// RecordExecution records a SQL execution (INSERT, UPDATE, DELETE)
func (m *DatabaseMonitor) RecordExecution(databaseType string, operation string, success bool, duration time.Duration, rowsAffected int64) {
	if !m.shouldSample() {
		return
	}

	labels := map[string]string{
		"database_type": databaseType,
		"operation":     operation,
		"success":       boolToString(success),
	}

	// Record execution count
	m.collector.Increment("database_executions_total", labels)

	// Record execution duration
	if success {
		m.collector.RecordDuration("database_execution_duration_seconds", duration, labels)

		// Record rows affected
		m.collector.Record("database_execution_rows_affected", float64(rowsAffected), map[string]string{
			"database_type": databaseType,
			"operation":     operation,
		})
	}
}

// RecordError records a database error
func (m *DatabaseMonitor) RecordError(databaseType string, errorType string, operation string) {
	if !m.shouldSample() {
		return
	}

	labels := map[string]string{
		"database_type": databaseType,
		"error_type":    errorType,
		"operation":     operation,
	}

	m.collector.Increment("database_errors_total", labels)
}

// RecordTimeout records a database timeout
func (m *DatabaseMonitor) RecordTimeout(databaseType string, operation string) {
	if !m.shouldSample() {
		return
	}

	labels := map[string]string{
		"database_type": databaseType,
		"operation":     operation,
	}

	m.collector.Increment("database_timeouts_total", labels)
}

// UpdateConnectionPool updates connection pool metrics
func (m *DatabaseMonitor) UpdateConnectionPool(databaseType string, activeConnections, idleConnections int) {
	if !m.shouldSample() {
		return
	}

	m.collector.Record("database_connections_active", float64(activeConnections), map[string]string{
		"database_type": databaseType,
	})

	m.collector.Record("database_connections_idle", float64(idleConnections), map[string]string{
		"database_type": databaseType,
	})
}

// UpdateActiveTransactions updates active transaction count
func (m *DatabaseMonitor) UpdateActiveTransactions(databaseType string, count int) {
	if !m.shouldSample() {
		return
	}

	m.collector.Record("database_transaction_active", float64(count), map[string]string{
		"database_type": databaseType,
	})
}

// UpdateResourceUsage updates database resource usage metrics
func (m *DatabaseMonitor) UpdateResourceUsage(databaseType string, memoryBytes int64, cpuSeconds float64) {
	if !m.shouldSample() {
		return
	}

	m.collector.Record("database_memory_bytes", float64(memoryBytes), map[string]string{
		"database_type": databaseType,
	})

	m.collector.Record("database_cpu_seconds_total", cpuSeconds, map[string]string{
		"database_type": databaseType,
	})
}

// RecordOperation records a complete database operation with timing
func (m *DatabaseMonitor) RecordOperation(operation string, startTime time.Time, success bool, labels map[string]string) {
	if !m.shouldSample() {
		return
	}

	m.collector.RecordOperation(operation, startTime, success, labels)
}

// shouldSample determines if the current operation should be sampled
func (m *DatabaseMonitor) shouldSample() bool {
	if m.config == nil {
		return true
	}
	return m.config.ShouldSample()
}

// GetCollector returns the collector (for internal use)
func (m *DatabaseMonitor) GetCollector() *core.Collector {
	return m.collector
}

// boolToString converts boolean to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// stringifyErrorCode converts error code to string
func stringifyErrorCode(code cd.Code) string {
	// Map common error codes to strings
	switch code {
	case cd.Success:
		return "Success"
	case cd.UnKnownError:
		return "UnKnownError"
	case cd.NotFound:
		return "NotFound"
	case cd.InvalidParameter:
		return "InvalidParameter"
	case cd.IllegalParam:
		return "IllegalParam"
	case cd.InvalidAuthority:
		return "InvalidAuthority"
	case cd.Unexpected:
		return "Unexpected"
	case cd.Duplicated:
		return "Duplicated"
	case cd.DatabaseError:
		return "DatabaseError"
	case cd.Timeout:
		return "Timeout"
	case cd.NetworkError:
		return "NetworkError"
	default:
		return "Unknown"
	}
}

// MonitoredExecutor wraps a database executor with monitoring
type MonitoredExecutor struct {
	database.Executor
	monitor *DatabaseMonitor
	dbType  string
}

// NewMonitoredExecutor creates a new monitored executor
func NewMonitoredExecutor(executor database.Executor, monitor *DatabaseMonitor, dbType string) *MonitoredExecutor {
	return &MonitoredExecutor{
		Executor: executor,
		monitor:  monitor,
		dbType:   dbType,
	}
}

// Query executes a query with monitoring
func (e *MonitoredExecutor) Query(sql string, needCols bool, args ...any) (ret []string, err *cd.Error) {
	startTime := time.Now()
	success := false

	defer func() {
		duration := time.Since(startTime)
		e.monitor.RecordQuery(e.dbType, "query", success, duration, 0)

		if err != nil {
			e.monitor.RecordError(e.dbType, stringifyErrorCode(err.Code), "query")
		}
	}()

	ret, err = e.Executor.Query(sql, needCols, args...)
	success = err == nil
	return
}

// Execute executes SQL with monitoring
func (e *MonitoredExecutor) Execute(sql string, args ...any) (rowsAffected int64, err *cd.Error) {
	startTime := time.Now()
	success := false

	defer func() {
		duration := time.Since(startTime)
		e.monitor.RecordExecution(e.dbType, "execute", success, duration, rowsAffected)

		if err != nil {
			e.monitor.RecordError(e.dbType, stringifyErrorCode(err.Code), "execute")
		}
	}()

	rowsAffected, err = e.Executor.Execute(sql, args...)
	success = err == nil
	return
}

// ExecuteInsert executes INSERT with monitoring
func (e *MonitoredExecutor) ExecuteInsert(sql string, pkValOut any, args ...any) (err *cd.Error) {
	startTime := time.Now()
	success := false

	defer func() {
		duration := time.Since(startTime)
		e.monitor.RecordExecution(e.dbType, "insert", success, duration, 1)

		if err != nil {
			e.monitor.RecordError(e.dbType, stringifyErrorCode(err.Code), "insert")
		}
	}()

	err = e.Executor.ExecuteInsert(sql, pkValOut, args...)
	success = err == nil
	return
}

// BeginTransaction begins a transaction with monitoring
func (e *MonitoredExecutor) BeginTransaction() (err *cd.Error) {
	startTime := time.Now()
	success := false

	defer func() {
		duration := time.Since(startTime)
		e.monitor.RecordTransaction(e.dbType, "begin", success, duration)

		if err != nil {
			e.monitor.RecordError(e.dbType, stringifyErrorCode(err.Code), "begin_transaction")
		} else {
			e.monitor.UpdateActiveTransactions(e.dbType, 1)
		}
	}()

	err = e.Executor.BeginTransaction()
	success = err == nil
	return
}

// CommitTransaction commits a transaction with monitoring
func (e *MonitoredExecutor) CommitTransaction() (err *cd.Error) {
	startTime := time.Now()
	success := false

	defer func() {
		duration := time.Since(startTime)
		e.monitor.RecordTransaction(e.dbType, "commit", success, duration)

		if err != nil {
			e.monitor.RecordError(e.dbType, stringifyErrorCode(err.Code), "commit_transaction")
		} else {
			e.monitor.UpdateActiveTransactions(e.dbType, -1)
		}
	}()

	err = e.Executor.CommitTransaction()
	success = err == nil
	return
}

// RollbackTransaction rolls back a transaction with monitoring
func (e *MonitoredExecutor) RollbackTransaction() (err *cd.Error) {
	startTime := time.Now()
	success := false

	defer func() {
		duration := time.Since(startTime)
		e.monitor.RecordTransaction(e.dbType, "rollback", success, duration)

		if err != nil {
			e.monitor.RecordError(e.dbType, stringifyErrorCode(err.Code), "rollback_transaction")
		} else {
			e.monitor.UpdateActiveTransactions(e.dbType, -1)
		}
	}()

	err = e.Executor.RollbackTransaction()
	success = err == nil
	return
}

// MonitoredPool wraps a database pool with monitoring
type MonitoredPool struct {
	database.Pool
	monitor *DatabaseMonitor
	dbType  string
}

// NewMonitoredPool creates a new monitored pool
func NewMonitoredPool(pool database.Pool, monitor *DatabaseMonitor, dbType string) *MonitoredPool {
	return &MonitoredPool{
		Pool:    pool,
		monitor: monitor,
		dbType:  dbType,
	}
}

// GetExecutor gets an executor with monitoring
func (p *MonitoredPool) GetExecutor(ctx context.Context) (executor database.Executor, err *cd.Error) {
	startTime := time.Now()
	success := false

	defer func() {
		duration := time.Since(startTime)
		p.monitor.RecordConnection(p.dbType, "get_executor", success, duration)

		if err != nil {
			p.monitor.RecordError(p.dbType, stringifyErrorCode(err.Code), "get_executor")
		}
	}()

	executor, err = p.Pool.GetExecutor(ctx)
	success = err == nil

	if success && executor != nil {
		// Wrap the executor with monitoring
		executor = NewMonitoredExecutor(executor, p.monitor, p.dbType)
	}

	return
}
