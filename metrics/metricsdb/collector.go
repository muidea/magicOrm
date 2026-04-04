// Package metricsdb provides database-specific metric collection for MagicORM.
package metricsdb

import (
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicOrm/metrics"
)

type DurationAggregate struct {
	Total time.Duration
	Count int64
}

// DatabaseMetricsCollector collects and stores database operation metrics in a thread-safe manner.
type DatabaseMetricsCollector struct {
	queryMu sync.RWMutex
	txMu    sync.RWMutex
	execMu  sync.RWMutex
	connMu  sync.RWMutex

	// Query counters: database_queryType_status -> count
	queryCounters map[string]int64

	// Error counters: database_operation_errorType -> count
	errorCounters map[string]int64

	// Query duration aggregates: database_queryType_status -> {total,count}
	queryDurationAggregates map[string]DurationAggregate

	// Transaction counters: database_type_status -> count
	transactionCounters map[string]int64

	// Execution counters: database_operation_status -> count
	executionCounters map[string]int64

	// Connection pool statistics
	connectionStats map[string]int64
}

// NewDatabaseMetricsCollector creates a new database metrics collector.
func NewDatabaseMetricsCollector() *DatabaseMetricsCollector {
	return &DatabaseMetricsCollector{
		queryCounters:           make(map[string]int64),
		errorCounters:           make(map[string]int64),
		queryDurationAggregates: make(map[string]DurationAggregate),
		transactionCounters:     make(map[string]int64),
		executionCounters:       make(map[string]int64),
		connectionStats:         make(map[string]int64),
	}
}

// RecordQuery records a database query with its duration and error status.
func (c *DatabaseMetricsCollector) RecordQuery(
	database string,
	queryType string,
	duration time.Duration,
	err error,
) {
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	// Determine status and record query
	status := "success"
	if err != nil {
		status = "error"
		// Record error with classification
		errorType := c.classifyError(err)
		errorKey := metrics.BuildKey(database, queryType, errorType)
		c.errorCounters[errorKey]++
	}

	// Record query counter
	queryKey := metrics.BuildKey(database, queryType, status)
	c.queryCounters[queryKey]++

	aggregate := c.queryDurationAggregates[queryKey]
	aggregate.Total += duration
	aggregate.Count++
	c.queryDurationAggregates[queryKey] = aggregate
}

// RecordTransaction records a database transaction operation.
func (c *DatabaseMetricsCollector) RecordTransaction(database string, txType string, success bool) {
	c.txMu.Lock()
	defer c.txMu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}
	key := metrics.BuildKey(database, txType, status)
	c.transactionCounters[key]++
}

// RecordExecution records a database execution (INSERT, UPDATE, DELETE, etc.).
func (c *DatabaseMetricsCollector) RecordExecution(database string, operation string, success bool) {
	c.execMu.Lock()
	defer c.execMu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}
	key := metrics.BuildKey(database, operation, status)
	c.executionCounters[key]++
}

// UpdateConnectionStats updates connection pool statistics.
func (c *DatabaseMetricsCollector) UpdateConnectionStats(database string, state string, count int64) {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	key := metrics.BuildKey(database, state)
	if pre, ok := c.connectionStats[key]; ok && pre == count {
		return
	}
	c.connectionStats[key] = count
}

// GetQueryCounters returns a copy of query counters.
func (c *DatabaseMetricsCollector) GetQueryCounters() map[string]int64 {
	c.queryMu.RLock()
	defer c.queryMu.RUnlock()

	result := make(map[string]int64, len(c.queryCounters))
	for k, v := range c.queryCounters {
		result[k] = v
	}
	return result
}

// GetErrorCounters returns a copy of error counters.
func (c *DatabaseMetricsCollector) GetErrorCounters() map[string]int64 {
	c.queryMu.RLock()
	defer c.queryMu.RUnlock()

	result := make(map[string]int64, len(c.errorCounters))
	for k, v := range c.errorCounters {
		result[k] = v
	}
	return result
}

// GetQueryDurationAggregates returns a copy of duration aggregates.
func (c *DatabaseMetricsCollector) GetQueryDurationAggregates() map[string]DurationAggregate {
	c.queryMu.RLock()
	defer c.queryMu.RUnlock()

	result := make(map[string]DurationAggregate, len(c.queryDurationAggregates))
	for k, v := range c.queryDurationAggregates {
		result[k] = v
	}
	return result
}

// GetTransactionCounters returns a copy of transaction counters.
func (c *DatabaseMetricsCollector) GetTransactionCounters() map[string]int64 {
	c.txMu.RLock()
	defer c.txMu.RUnlock()

	result := make(map[string]int64, len(c.transactionCounters))
	for k, v := range c.transactionCounters {
		result[k] = v
	}
	return result
}

// GetExecutionCounters returns a copy of execution counters.
func (c *DatabaseMetricsCollector) GetExecutionCounters() map[string]int64 {
	c.execMu.RLock()
	defer c.execMu.RUnlock()

	result := make(map[string]int64, len(c.executionCounters))
	for k, v := range c.executionCounters {
		result[k] = v
	}
	return result
}

// GetConnectionStats returns a copy of connection statistics.
func (c *DatabaseMetricsCollector) GetConnectionStats() map[string]int64 {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	result := make(map[string]int64, len(c.connectionStats))
	for k, v := range c.connectionStats {
		result[k] = v
	}
	return result
}

// Clear clears all collected metrics (useful for testing).
func (c *DatabaseMetricsCollector) Clear() {
	c.queryMu.Lock()
	c.txMu.Lock()
	c.execMu.Lock()
	c.connMu.Lock()
	defer c.connMu.Unlock()
	defer c.execMu.Unlock()
	defer c.txMu.Unlock()
	defer c.queryMu.Unlock()

	c.queryCounters = make(map[string]int64)
	c.errorCounters = make(map[string]int64)
	c.queryDurationAggregates = make(map[string]DurationAggregate)
	c.transactionCounters = make(map[string]int64)
	c.executionCounters = make(map[string]int64)
	c.connectionStats = make(map[string]int64)
}

// classifyError classifies an error into error types for metrics.
func (c *DatabaseMetricsCollector) classifyError(err error) string {
	if err == nil {
		return string(metrics.ErrorTypeUnknown)
	}

	// 使用recover安全地获取错误字符串
	var errStr string
	func() {
		defer func() {
			if r := recover(); r != nil {
				// 如果获取错误字符串时发生panic，设置errStr为空
				errStr = ""
			}
		}()
		errStr = err.Error()
	}()

	if errStr == "" {
		return string(metrics.ErrorTypeUnknown)
	}

	errLower := strings.ToLower(errStr)

	switch {
	case strings.Contains(errLower, "connection"):
		return string(metrics.ErrorTypeConnection)
	case strings.Contains(errLower, "timeout"):
		return string(metrics.ErrorTypeTimeout)
	case strings.Contains(errLower, "deadlock"):
		return string(metrics.ErrorTypeDatabase)
	case strings.Contains(errLower, "constraint"):
		return string(metrics.ErrorTypeConstraint)
	case strings.Contains(errLower, "syntax"):
		return string(metrics.ErrorTypeDatabase)
	case strings.Contains(errLower, "permission"):
		return string(metrics.ErrorTypeDatabase)
	case strings.Contains(errLower, "duplicate"):
		return string(metrics.ErrorTypeConstraint)
	case strings.Contains(errLower, "not found"):
		return string(metrics.ErrorTypeDatabase)
	default:
		return string(metrics.ErrorTypeUnknown)
	}
}
