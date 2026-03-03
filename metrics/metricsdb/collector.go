// Package metricsdb provides database-specific metric collection for MagicORM.
package metricsdb

import (
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicOrm/metrics"
)

// DatabaseMetricsCollector collects and stores database operation metrics in a thread-safe manner.
type DatabaseMetricsCollector struct {
	mu sync.RWMutex

	// Query counters: database_queryType_status -> count
	queryCounters map[string]int64

	// Error counters: database_operation_errorType -> count
	errorCounters map[string]int64

	// Query durations: database_queryType_status -> []duration
	queryDurations map[string][]time.Duration

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
		queryCounters:       make(map[string]int64),
		errorCounters:       make(map[string]int64),
		queryDurations:      make(map[string][]time.Duration),
		transactionCounters: make(map[string]int64),
		executionCounters:   make(map[string]int64),
		connectionStats:     make(map[string]int64),
	}
}

// RecordQuery records a database query with its duration and error status.
func (c *DatabaseMetricsCollector) RecordQuery(
	database string,
	queryType string,
	duration time.Duration,
	err error,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

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

	// Record duration (keep last 1000 samples per key to avoid memory leak)
	if c.queryDurations[queryKey] == nil {
		c.queryDurations[queryKey] = make([]time.Duration, 0, 1000)
	}
	durations := c.queryDurations[queryKey]
	if len(durations) >= 1000 {
		// Keep only the last 1000 samples - copy to avoid modifying the slice in place
		newDurations := make([]time.Duration, 999, 1000)
		copy(newDurations, durations[1:])
		durations = newDurations
	}
	c.queryDurations[queryKey] = append(durations, duration)
}

// RecordTransaction records a database transaction operation.
func (c *DatabaseMetricsCollector) RecordTransaction(database string, txType string, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}
	key := metrics.BuildKey(database, txType, status)
	c.transactionCounters[key]++
}

// RecordExecution records a database execution (INSERT, UPDATE, DELETE, etc.).
func (c *DatabaseMetricsCollector) RecordExecution(database string, operation string, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}
	key := metrics.BuildKey(database, operation, status)
	c.executionCounters[key]++
}

// UpdateConnectionStats updates connection pool statistics.
func (c *DatabaseMetricsCollector) UpdateConnectionStats(database string, state string, count int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := metrics.BuildKey(database, state)
	c.connectionStats[key] = count
}

// GetQueryCounters returns a copy of query counters.
func (c *DatabaseMetricsCollector) GetQueryCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.queryCounters))
	for k, v := range c.queryCounters {
		result[k] = v
	}
	return result
}

// GetErrorCounters returns a copy of error counters.
func (c *DatabaseMetricsCollector) GetErrorCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.errorCounters))
	for k, v := range c.errorCounters {
		result[k] = v
	}
	return result
}

// GetQueryDurations returns a copy of query durations.
func (c *DatabaseMetricsCollector) GetQueryDurations() map[string][]time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string][]time.Duration, len(c.queryDurations))
	for k, v := range c.queryDurations {
		// Create a copy of the slice
		durations := make([]time.Duration, len(v))
		copy(durations, v)
		result[k] = durations
	}
	return result
}

// GetTransactionCounters returns a copy of transaction counters.
func (c *DatabaseMetricsCollector) GetTransactionCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.transactionCounters))
	for k, v := range c.transactionCounters {
		result[k] = v
	}
	return result
}

// GetExecutionCounters returns a copy of execution counters.
func (c *DatabaseMetricsCollector) GetExecutionCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.executionCounters))
	for k, v := range c.executionCounters {
		result[k] = v
	}
	return result
}

// GetConnectionStats returns a copy of connection statistics.
func (c *DatabaseMetricsCollector) GetConnectionStats() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.connectionStats))
	for k, v := range c.connectionStats {
		result[k] = v
	}
	return result
}

// Clear clears all collected metrics (useful for testing).
func (c *DatabaseMetricsCollector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.queryCounters = make(map[string]int64)
	c.errorCounters = make(map[string]int64)
	c.queryDurations = make(map[string][]time.Duration)
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
