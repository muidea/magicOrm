package orm

import (
	"fmt"
	"time"

	"github.com/muidea/magicOrm/monitoring/core"
)

// OperationType represents the type of ORM operation
type OperationType string

const (
	OperationInsert OperationType = "insert"
	OperationUpdate OperationType = "update"
	OperationQuery  OperationType = "query"
	OperationDelete OperationType = "delete"
	OperationCreate OperationType = "create"
	OperationDrop   OperationType = "drop"
	OperationCount  OperationType = "count"
	OperationBatch  OperationType = "batch"
)

// ErrorType represents the type of ORM error
type ErrorType string

const (
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeDatabase    ErrorType = "database"
	ErrorTypeConnection  ErrorType = "connection"
	ErrorTypeTimeout     ErrorType = "timeout"
	ErrorTypeConstraint  ErrorType = "constraint"
	ErrorTypeTransaction ErrorType = "transaction"
	ErrorTypeUnknown     ErrorType = "unknown"
)

// QueryType represents the type of query
type QueryType string

const (
	QueryTypeSimple   QueryType = "simple"
	QueryTypeFilter   QueryType = "filter"
	QueryTypeRelation QueryType = "relation"
	QueryTypeBatch    QueryType = "batch"
)

// ORMMonitor provides monitoring for ORM operations
type ORMMonitor struct {
	collector *core.Collector
	config    *core.MonitoringConfig
}

// NewORMMonitor creates a new ORM monitor
func NewORMMonitor(collector *core.Collector, config *core.MonitoringConfig) *ORMMonitor {
	if config == nil {
		defaultConfig := core.DefaultMonitoringConfig()
		config = &defaultConfig
	}

	monitor := &ORMMonitor{
		collector: collector,
		config:    config,
	}

	// Register ORM-specific metrics
	monitor.registerMetrics()

	return monitor
}

// RecordOperation records an ORM operation
func (m *ORMMonitor) RecordOperation(
	operation OperationType,
	modelName string,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	if !m.config.IsORMEnabled() {
		return
	}

	duration := time.Since(startTime)
	labels := m.buildOperationLabels(operation, modelName, additionalLabels)

	// Record operation with status
	status := "success"
	if err != nil {
		status = "error"
	}
	labels["status"] = status

	m.collector.RecordOperation(
		"orm_operation",
		startTime,
		err == nil,
		labels,
	)

	// Record operation duration
	m.collector.RecordDuration(
		"orm_operation_duration_seconds",
		duration,
		labels,
	)

	// Record error if any
	if err != nil {
		errorType := m.classifyError(err)
		errorLabels := make(map[string]string)
		for k, v := range labels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = string(errorType)

		m.collector.RecordError(
			"orm",
			string(errorType),
			errorLabels,
		)
	}
}

// RecordQuery records a query operation with additional details
func (m *ORMMonitor) RecordQuery(
	modelName string,
	queryType QueryType,
	rowsReturned int,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	if !m.config.IsORMEnabled() {
		return
	}

	duration := time.Since(startTime)
	labels := m.buildQueryLabels(modelName, queryType, additionalLabels)

	// Record query operation
	status := "success"
	if err != nil {
		status = "error"
	}
	labels["status"] = status

	m.collector.RecordOperation(
		"orm_query",
		startTime,
		err == nil,
		labels,
	)

	// Record query duration
	m.collector.RecordDuration(
		"orm_query_duration_seconds",
		duration,
		labels,
	)

	// Record rows returned
	m.collector.Record(
		"orm_query_rows_returned",
		float64(rowsReturned),
		labels,
	)

	// Record error if any
	if err != nil {
		errorType := m.classifyError(err)
		errorLabels := make(map[string]string)
		for k, v := range labels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = string(errorType)

		m.collector.RecordError(
			"orm_query",
			string(errorType),
			errorLabels,
		)
	}
}

// RecordBatchOperation records a batch operation
func (m *ORMMonitor) RecordBatchOperation(
	operation OperationType,
	modelName string,
	batchSize int,
	startTime time.Time,
	successCount int,
	failedCount int,
	err error,
	additionalLabels map[string]string,
) {
	if !m.config.IsORMEnabled() {
		return
	}

	duration := time.Since(startTime)
	labels := m.buildBatchLabels(operation, modelName, additionalLabels)

	// Record batch operation
	status := "success"
	if err != nil || failedCount > 0 {
		status = "partial"
		if successCount == 0 {
			status = "error"
		}
	}
	labels["status"] = status
	labels["batch_size"] = fmt.Sprintf("%d", batchSize)

	m.collector.RecordOperation(
		"orm_batch_operation",
		startTime,
		err == nil && failedCount == 0,
		labels,
	)

	// Record batch duration
	m.collector.RecordDuration(
		"orm_batch_duration_seconds",
		duration,
		labels,
	)

	// Record batch statistics
	m.collector.Record(
		"orm_batch_size",
		float64(batchSize),
		labels,
	)

	m.collector.Record(
		"orm_batch_success_count",
		float64(successCount),
		labels,
	)

	m.collector.Record(
		"orm_batch_failed_count",
		float64(failedCount),
		labels,
	)

	// Record error if any
	if err != nil || failedCount > 0 {
		errorType := m.classifyError(err)
		if errorType == ErrorTypeUnknown && failedCount > 0 {
			errorType = ErrorTypeValidation
		}

		errorLabels := make(map[string]string)
		for k, v := range labels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = string(errorType)

		m.collector.RecordError(
			"orm_batch",
			string(errorType),
			errorLabels,
		)
	}
}

// RecordTransaction records a transaction operation
func (m *ORMMonitor) RecordTransaction(
	operation string, // begin, commit, rollback
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	if !m.config.IsORMEnabled() {
		return
	}

	duration := time.Since(startTime)
	labels := m.buildTransactionLabels(operation, additionalLabels)

	// Record transaction operation
	status := "success"
	if err != nil {
		status = "error"
	}
	labels["status"] = status

	m.collector.RecordOperation(
		"orm_transaction",
		startTime,
		err == nil,
		labels,
	)

	// Record transaction duration
	m.collector.RecordDuration(
		"orm_transaction_duration_seconds",
		duration,
		labels,
	)

	// Record error if any
	if err != nil {
		errorType := m.classifyError(err)
		errorLabels := make(map[string]string)
		for k, v := range labels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = string(errorType)

		m.collector.RecordError(
			"orm_transaction",
			string(errorType),
			errorLabels,
		)
	}
}

// RecordCacheAccess records cache access for ORM
func (m *ORMMonitor) RecordCacheAccess(
	cacheType string,
	operation string,
	hit bool,
	duration time.Duration,
	additionalLabels map[string]string,
) {
	if !m.config.IsORMEnabled() || !m.config.IsCacheEnabled() {
		return
	}

	labels := m.buildCacheLabels(cacheType, additionalLabels)
	labels["operation"] = operation
	labels["hit"] = "false"
	if hit {
		labels["hit"] = "true"
	}

	// Record cache operation
	m.collector.Increment(
		"orm_cache_operations_total",
		labels,
	)

	// Record cache hit/miss
	if hit {
		m.collector.Increment(
			"orm_cache_hits_total",
			labels,
		)
	} else {
		m.collector.Increment(
			"orm_cache_misses_total",
			labels,
		)
	}

	// Record cache duration
	m.collector.RecordDuration(
		"orm_cache_duration_seconds",
		duration,
		labels,
	)
}

// RecordDatabaseOperation records a database-level operation
func (m *ORMMonitor) RecordDatabaseOperation(
	dbType string,
	operation string,
	startTime time.Time,
	err error,
	additionalLabels map[string]string,
) {
	if !m.config.IsORMEnabled() || !m.config.IsDatabaseEnabled() {
		return
	}

	duration := time.Since(startTime)
	labels := m.buildDatabaseLabels(dbType, operation, additionalLabels)

	// Record database operation
	status := "success"
	if err != nil {
		status = "error"
	}
	labels["status"] = status

	m.collector.RecordOperation(
		"orm_database_operation",
		startTime,
		err == nil,
		labels,
	)

	// Record database duration
	m.collector.RecordDuration(
		"orm_database_duration_seconds",
		duration,
		labels,
	)

	// Record error if any
	if err != nil {
		errorType := m.classifyError(err)
		errorLabels := make(map[string]string)
		for k, v := range labels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = string(errorType)

		m.collector.RecordError(
			"orm_database",
			string(errorType),
			errorLabels,
		)
	}
}

// RecordConnectionPool records connection pool statistics
func (m *ORMMonitor) RecordConnectionPool(
	dbType string,
	activeConnections int,
	idleConnections int,
	waitingConnections int,
	maxConnections int,
	additionalLabels map[string]string,
) {
	if !m.config.IsORMEnabled() || !m.config.IsDatabaseEnabled() {
		return
	}

	labels := m.buildConnectionPoolLabels(dbType, additionalLabels)

	// Record connection pool metrics
	m.collector.Record(
		"orm_connections_active",
		float64(activeConnections),
		labels,
	)

	m.collector.Record(
		"orm_connections_idle",
		float64(idleConnections),
		labels,
	)

	m.collector.Record(
		"orm_connections_waiting",
		float64(waitingConnections),
		labels,
	)

	m.collector.Record(
		"orm_connections_max",
		float64(maxConnections),
		labels,
	)

	// Calculate utilization
	if maxConnections > 0 {
		utilization := float64(activeConnections) / float64(maxConnections)
		m.collector.Record(
			"orm_connections_utilization",
			utilization,
			labels,
		)
	}
}

// GetStats returns ORM monitoring statistics
func (m *ORMMonitor) GetStats() map[string]interface{} {
	// Get metrics for ORM operations
	metrics, err := m.collector.GetMetric("orm_operation_total")
	if err != nil {
		metrics = []core.Metric{}
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"total_operations": len(metrics),
		"enabled":          m.config.IsORMEnabled(),
	}

	return stats
}

// Private methods

func (m *ORMMonitor) registerMetrics() {
	// ORM operation metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_operation_total",
		Type:       core.CounterMetric,
		Help:       "Total number of ORM operations",
		LabelNames: []string{"operation", "model", "status"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_operation_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "ORM operation duration in seconds",
		LabelNames: []string{"operation", "model", "status"},
		Buckets:    []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	})

	// ORM error metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_errors_total",
		Type:       core.CounterMetric,
		Help:       "Total number of ORM errors",
		LabelNames: []string{"operation", "model", "error_type"},
	})

	// Query metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_query_total",
		Type:       core.CounterMetric,
		Help:       "Total number of ORM queries",
		LabelNames: []string{"model", "query_type", "status"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_query_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "ORM query duration in seconds",
		LabelNames: []string{"model", "query_type", "status"},
		Buckets:    []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_query_rows_returned",
		Type:       core.HistogramMetric,
		Help:       "Number of rows returned by queries",
		LabelNames: []string{"model", "query_type", "status"},
		Buckets:    []float64{1, 5, 10, 50, 100, 500, 1000, 5000, 10000},
	})

	// Batch operation metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_batch_operation_total",
		Type:       core.CounterMetric,
		Help:       "Total number of batch operations",
		LabelNames: []string{"operation", "model", "status", "batch_size"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_batch_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Batch operation duration in seconds",
		LabelNames: []string{"operation", "model", "status", "batch_size"},
		Buckets:    []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_batch_size",
		Type:       core.GaugeMetric,
		Help:       "Size of batch operations",
		LabelNames: []string{"operation", "model", "status"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_batch_success_count",
		Type:       core.GaugeMetric,
		Help:       "Number of successful operations in batch",
		LabelNames: []string{"operation", "model", "status"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_batch_failed_count",
		Type:       core.GaugeMetric,
		Help:       "Number of failed operations in batch",
		LabelNames: []string{"operation", "model", "status"},
	})

	// Transaction metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_transaction_total",
		Type:       core.CounterMetric,
		Help:       "Total number of transactions",
		LabelNames: []string{"operation", "status"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_transaction_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Transaction duration in seconds",
		LabelNames: []string{"operation", "status"},
		Buckets:    []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
	})

	// Cache metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_cache_operations_total",
		Type:       core.CounterMetric,
		Help:       "Total number of cache operations",
		LabelNames: []string{"cache_type", "operation", "hit"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_cache_hits_total",
		Type:       core.CounterMetric,
		Help:       "Total number of cache hits",
		LabelNames: []string{"cache_type", "operation"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_cache_misses_total",
		Type:       core.CounterMetric,
		Help:       "Total number of cache misses",
		LabelNames: []string{"cache_type", "operation"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_cache_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Cache operation duration in seconds",
		LabelNames: []string{"cache_type", "operation", "hit"},
		Buckets:    []float64{.0001, .0005, .001, .005, .01, .025, .05},
	})

	// Database metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_database_operation_total",
		Type:       core.CounterMetric,
		Help:       "Total number of database operations",
		LabelNames: []string{"db_type", "operation", "status"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_database_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Database operation duration in seconds",
		LabelNames: []string{"db_type", "operation", "status"},
		Buckets:    []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	})

	// Connection pool metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_connections_active",
		Type:       core.GaugeMetric,
		Help:       "Number of active database connections",
		LabelNames: []string{"db_type"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_connections_idle",
		Type:       core.GaugeMetric,
		Help:       "Number of idle database connections",
		LabelNames: []string{"db_type"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_connections_waiting",
		Type:       core.GaugeMetric,
		Help:       "Number of waiting database connections",
		LabelNames: []string{"db_type"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_connections_max",
		Type:       core.GaugeMetric,
		Help:       "Maximum number of database connections",
		LabelNames: []string{"db_type"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "orm_connections_utilization",
		Type:       core.GaugeMetric,
		Help:       "Database connection pool utilization (0-1)",
		LabelNames: []string{"db_type"},
	})
}

func (m *ORMMonitor) buildOperationLabels(operation OperationType, modelName string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	labels["operation"] = string(operation)
	if modelName != "" {
		labels["model"] = modelName
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ORMMonitor) buildQueryLabels(modelName string, queryType QueryType, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if modelName != "" {
		labels["model"] = modelName
	}
	labels["query_type"] = string(queryType)

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ORMMonitor) buildBatchLabels(operation OperationType, modelName string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	labels["operation"] = string(operation)
	if modelName != "" {
		labels["model"] = modelName
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ORMMonitor) buildTransactionLabels(operation string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	labels["operation"] = operation

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ORMMonitor) buildCacheLabels(cacheType string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if cacheType != "" {
		labels["cache_type"] = cacheType
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ORMMonitor) buildDatabaseLabels(dbType, operation string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if dbType != "" {
		labels["db_type"] = dbType
	}
	if operation != "" {
		labels["operation"] = operation
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ORMMonitor) buildConnectionPoolLabels(dbType string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if dbType != "" {
		labels["db_type"] = dbType
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ORMMonitor) classifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	// Classify error based on error message
	errStr := err.Error()

	switch {
	case contains(errStr, "validation"):
		return ErrorTypeValidation
	case contains(errStr, "database"):
		return ErrorTypeDatabase
	case contains(errStr, "connection"):
		return ErrorTypeConnection
	case contains(errStr, "timeout"):
		return ErrorTypeTimeout
	case contains(errStr, "constraint"):
		return ErrorTypeConstraint
	case contains(errStr, "transaction"):
		return ErrorTypeTransaction
	default:
		return ErrorTypeUnknown
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}

// Convenience functions

// DefaultORMMonitor creates an ORM monitor with default configuration
func DefaultORMMonitor() *ORMMonitor {
	config := core.DefaultMonitoringConfig()
	collector := core.NewCollector(&config)
	return NewORMMonitor(collector, &config)
}

// ORMMonitorWithConfig creates an ORM monitor with custom configuration
func ORMMonitorWithConfig(config *core.MonitoringConfig) *ORMMonitor {
	collector := core.NewCollector(config)
	return NewORMMonitor(collector, config)
}
