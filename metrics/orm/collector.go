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
	durationKeyLRU  []string
	maxDurationKeys int
}

// NewORMMetricsCollector creates a new ORM metrics collector.
func NewORMMetricsCollector() *ORMMetricsCollector {
	return &ORMMetricsCollector{
		operationCounters:   make(map[string]int64),
		errorCounters:       make(map[string]int64),
		operationDurations:  make(map[string][]time.Duration),
		transactionCounters: make(map[string]int64),
		durationKeyLRU:      make([]string, 0, 1000),
		maxDurationKeys:     1000,
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
	// Initialize if needed
	if c.operationDurations[key] == nil {
		c.operationDurations[key] = make([]time.Duration, 0, 1000)

		// Check if we need to evict old keys
		if len(c.durationKeyLRU) >= c.maxDurationKeys {
			// Remove oldest key
			oldestKey := c.durationKeyLRU[0]
			delete(c.operationDurations, oldestKey)
			c.durationKeyLRU = c.durationKeyLRU[1:]
		}

		// Add new key to LRU
		c.durationKeyLRU = append(c.durationKeyLRU, key)
	} else {
		// Move key to end of LRU (most recently used)
		for i, k := range c.durationKeyLRU {
			if k == key {
				// Remove from current position
				c.durationKeyLRU = append(c.durationKeyLRU[:i], c.durationKeyLRU[i+1:]...)
				// Add to end
				c.durationKeyLRU = append(c.durationKeyLRU, key)
				break
			}
		}
	}

	// Record duration (keep last 1000 samples per key)
	durations := c.operationDurations[key]
	if len(durations) >= 1000 {
		// Keep only the last 1000 samples - copy to avoid modifying the slice in place
		newDurations := make([]time.Duration, 999, 1000)
		copy(newDurations, durations[1:])
		durations = newDurations
	}
	c.operationDurations[key] = append(durations, duration)
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
		return "none"
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
		return "unknown"
	}

	switch {
	case strings.Contains(errStr, "validation"):
		return "validation"
	case strings.Contains(errStr, "database"):
		return "database"
	case strings.Contains(errStr, "connection"):
		return "connection"
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "constraint"):
		return "constraint"
	case strings.Contains(errStr, "transaction"):
		return "transaction"
	default:
		return "unknown"
	}
}
