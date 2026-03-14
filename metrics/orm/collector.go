// Package orm provides ORM-specific metric collection for MagicORM.
package orm

import (
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicOrm/metrics"
	"github.com/muidea/magicOrm/models"
)

// ORMMetricsCollector collects and stores ORM operation metrics in a thread-safe manner.
type ORMMetricsCollector struct {
	mu sync.RWMutex

	// Operation counters: operation_model_status -> count
	operationCounters map[string]int64

	// Error counters: operation_model_errorType -> count
	errorCounters map[string]int64

	// Duration records: operation_model_status -> []duration
	operationDurations map[string][]time.Duration

	// Transaction counters: type_status -> count
	transactionCounters map[string]int64

	// Cache statistics
	cacheHits   int64
	cacheMisses int64

	// Connection statistics
	activeConnections int64

	// LRU tracking for duration keys to prevent unlimited growth
	durationKeyLRU     []string
	maxDurationKeys    int
	maxDurationSamples int
}

// NewORMMetricsCollector creates a new ORM metrics collector.
func NewORMMetricsCollector() *ORMMetricsCollector {
	return &ORMMetricsCollector{
		operationCounters:   make(map[string]int64),
		errorCounters:       make(map[string]int64),
		operationDurations:  make(map[string][]time.Duration),
		transactionCounters: make(map[string]int64),
		durationKeyLRU:      make([]string, 0, 1000),
		maxDurationKeys:     metrics.DefaultMaxDurationKeys,
		maxDurationSamples:  metrics.DefaultMaxDurationSamples,
	}
}

// RecordOperation records an ORM operation with its duration and error status.
func (c *ORMMetricsCollector) RecordOperation(
	operation string,
	model models.Model,
	duration time.Duration,
	err error,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get model name
	modelName := "unknown"
	if model != nil {
		modelName = model.GetPkgKey()
	}

	// Determine status and record operation
	status := "success"
	if err != nil {
		status = "error"
		// Record error with classification
		errorType := c.classifyError(err)
		errorKey := metrics.BuildKey(operation, modelName, errorType)
		c.errorCounters[errorKey]++
	}

	// Record operation counter
	opKey := metrics.BuildKey(operation, modelName, status)
	c.operationCounters[opKey]++

	// Record duration with LRU management
	c.recordDurationWithLRU(opKey, duration)
}

// RecordTransaction records a transaction operation.
func (c *ORMMetricsCollector) RecordTransaction(txType string, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}
	key := metrics.BuildKey(txType, status)
	c.transactionCounters[key]++
}

// RecordCacheHit records a cache hit.
func (c *ORMMetricsCollector) RecordCacheHit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheHits++
}

// RecordCacheMiss records a cache miss.
func (c *ORMMetricsCollector) RecordCacheMiss() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheMisses++
}

// UpdateActiveConnections updates the active connections count.
func (c *ORMMetricsCollector) UpdateActiveConnections(count int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.activeConnections = count
}

// GetOperationCounters returns a copy of operation counters.
func (c *ORMMetricsCollector) GetOperationCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.operationCounters))
	for k, v := range c.operationCounters {
		result[k] = v
	}
	return result
}

// GetErrorCounters returns a copy of error counters.
func (c *ORMMetricsCollector) GetErrorCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.errorCounters))
	for k, v := range c.errorCounters {
		result[k] = v
	}
	return result
}

// GetOperationDurations returns a copy of operation durations.
func (c *ORMMetricsCollector) GetOperationDurations() map[string][]time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string][]time.Duration, len(c.operationDurations))
	for k, v := range c.operationDurations {
		// Create a copy of the slice
		durations := make([]time.Duration, len(v))
		copy(durations, v)
		result[k] = durations
	}
	return result
}

// GetTransactionCounters returns a copy of transaction counters.
func (c *ORMMetricsCollector) GetTransactionCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.transactionCounters))
	for k, v := range c.transactionCounters {
		result[k] = v
	}
	return result
}

// GetCacheStats returns cache hit and miss counts.
func (c *ORMMetricsCollector) GetCacheStats() (hits, misses int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cacheHits, c.cacheMisses
}

// GetActiveConnections returns the active connections count.
func (c *ORMMetricsCollector) GetActiveConnections() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.activeConnections
}

// recordDurationWithLRU records a duration with LRU key management.
func (c *ORMMetricsCollector) recordDurationWithLRU(key string, duration time.Duration) {
	metrics.RecordDurationSample(
		c.operationDurations,
		&c.durationKeyLRU,
		c.maxDurationKeys,
		c.maxDurationSamples,
		key,
		duration,
	)
}

// Clear clears all collected metrics (useful for testing).
func (c *ORMMetricsCollector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.operationCounters = make(map[string]int64)
	c.errorCounters = make(map[string]int64)
	c.operationDurations = make(map[string][]time.Duration)
	c.transactionCounters = make(map[string]int64)
	c.cacheHits = 0
	c.cacheMisses = 0
	c.activeConnections = 0
	c.durationKeyLRU = make([]string, 0, 1000)
}

// classifyError classifies an error into error types for metrics.
func (c *ORMMetricsCollector) classifyError(err error) string {
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
	case strings.Contains(errLower, "validation"):
		return string(metrics.ErrorTypeValidation)
	case strings.Contains(errLower, "database"):
		return string(metrics.ErrorTypeDatabase)
	case strings.Contains(errLower, "connection"):
		return string(metrics.ErrorTypeConnection)
	case strings.Contains(errLower, "timeout"):
		return string(metrics.ErrorTypeTimeout)
	case strings.Contains(errLower, "constraint"):
		return string(metrics.ErrorTypeConstraint)
	case strings.Contains(errLower, "transaction"):
		return string(metrics.ErrorTypeTransaction)
	default:
		return string(metrics.ErrorTypeUnknown)
	}
}
