// Package metricsval provides validation-specific metric collection for MagicORM.
package validation

import (
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicOrm/metrics"
)

// ValidationMetricsCollector collects and stores validation operation metrics in a thread-safe manner.
type ValidationMetricsCollector struct {
	mu sync.RWMutex

	// Validation operation counters: operation_model_scenario_status -> count
	validationCounters map[string]int64

	// Error counters: operation_model_scenario_errorType -> count
	errorCounters map[string]int64

	// Validation durations: operation_model_scenario_status -> []duration
	validationDurations map[string][]time.Duration

	// Cache access counters: cacheType_hitMiss -> count
	cacheAccessCounters map[string]int64

	// Constraint check counters: constraintType_field_status -> count
	constraintCheckCounters map[string]int64
}

// NewValidationMetricsCollector creates a new validation metrics collector.
func NewValidationMetricsCollector() *ValidationMetricsCollector {
	return &ValidationMetricsCollector{
		validationCounters:      make(map[string]int64),
		errorCounters:           make(map[string]int64),
		validationDurations:     make(map[string][]time.Duration),
		cacheAccessCounters:     make(map[string]int64),
		constraintCheckCounters: make(map[string]int64),
	}
}

// RecordValidation records a validation operation with its duration and error status.
func (c *ValidationMetricsCollector) RecordValidation(
	operation string,
	model string,
	scenario string,
	duration time.Duration,
	err error,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Determine status and record validation
	status := "success"
	if err != nil {
		status = "error"
		// Record error with classification
		errorType := c.classifyError(err)
		errorKey := metrics.BuildKey(operation, model, scenario, errorType)
		c.errorCounters[errorKey]++
	}

	// Record validation counter
	validationKey := metrics.BuildKey(operation, model, scenario, status)
	c.validationCounters[validationKey]++

	// Record duration (keep last 1000 samples per key to avoid memory leak)
	if c.validationDurations[validationKey] == nil {
		c.validationDurations[validationKey] = make([]time.Duration, 0, 1000)
	}
	durations := c.validationDurations[validationKey]
	if len(durations) >= 1000 {
		// Keep only the last 1000 samples - copy to avoid modifying the slice in place
		newDurations := make([]time.Duration, 999, 1000)
		copy(newDurations, durations[1:])
		durations = newDurations
	}
	c.validationDurations[validationKey] = append(durations, duration)
}

// RecordCacheAccess records a cache access (hit or miss).
func (c *ValidationMetricsCollector) RecordCacheAccess(cacheType string, hit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hitMiss := "miss"
	if hit {
		hitMiss = "hit"
	}
	key := metrics.BuildKey(cacheType, hitMiss)
	c.cacheAccessCounters[key]++
}

// RecordConstraintCheck records a constraint check.
func (c *ValidationMetricsCollector) RecordConstraintCheck(constraintType string, field string, passed bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	status := "passed"
	if !passed {
		status = "failed"
	}
	key := metrics.BuildKey(constraintType, field, status)
	c.constraintCheckCounters[key]++
}

// GetValidationCounters returns a copy of validation counters.
func (c *ValidationMetricsCollector) GetValidationCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.validationCounters))
	for k, v := range c.validationCounters {
		result[k] = v
	}
	return result
}

// GetErrorCounters returns a copy of error counters.
func (c *ValidationMetricsCollector) GetErrorCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.errorCounters))
	for k, v := range c.errorCounters {
		result[k] = v
	}
	return result
}

// GetValidationDurations returns a copy of validation durations.
func (c *ValidationMetricsCollector) GetValidationDurations() map[string][]time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string][]time.Duration, len(c.validationDurations))
	for k, v := range c.validationDurations {
		// Create a copy of the slice
		durations := make([]time.Duration, len(v))
		copy(durations, v)
		result[k] = durations
	}
	return result
}

// GetCacheAccessCounters returns a copy of cache access counters.
func (c *ValidationMetricsCollector) GetCacheAccessCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.cacheAccessCounters))
	for k, v := range c.cacheAccessCounters {
		result[k] = v
	}
	return result
}

// GetConstraintCheckCounters returns a copy of constraint check counters.
func (c *ValidationMetricsCollector) GetConstraintCheckCounters() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]int64, len(c.constraintCheckCounters))
	for k, v := range c.constraintCheckCounters {
		result[k] = v
	}
	return result
}

// GetCacheHitRatio calculates and returns cache hit ratio for a given cache type.
func (c *ValidationMetricsCollector) GetCacheHitRatio(cacheType string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hitKey := metrics.BuildKey(cacheType, "hit")
	missKey := metrics.BuildKey(cacheType, "miss")

	hits := c.cacheAccessCounters[hitKey]
	misses := c.cacheAccessCounters[missKey]

	total := hits + misses
	if total == 0 {
		return 0.0
	}

	return float64(hits) / float64(total)
}

// Clear clears all collected metrics (useful for testing).
func (c *ValidationMetricsCollector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.validationCounters = make(map[string]int64)
	c.errorCounters = make(map[string]int64)
	c.validationDurations = make(map[string][]time.Duration)
	c.cacheAccessCounters = make(map[string]int64)
	c.constraintCheckCounters = make(map[string]int64)
}

// classifyError classifies an error into error types for metrics.
func (c *ValidationMetricsCollector) classifyError(err error) string {
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
	case strings.Contains(errLower, "type"):
		return string(metrics.ErrorTypeValidation)
	case strings.Contains(errLower, "constraint"):
		return string(metrics.ErrorTypeConstraint)
	case strings.Contains(errLower, "required"):
		return string(metrics.ErrorTypeValidation)
	case strings.Contains(errLower, "range"):
		return string(metrics.ErrorTypeValidation)
	case strings.Contains(errLower, "format"):
		return string(metrics.ErrorTypeValidation)
	case strings.Contains(errLower, "unique"):
		return string(metrics.ErrorTypeConstraint)
	case strings.Contains(errLower, "database"):
		return string(metrics.ErrorTypeDatabase)
	default:
		return string(metrics.ErrorTypeUnknown)
	}
}
