// Package validation provides validation-specific metric collectors for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package validation

import (
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

// ValidationMetricProvider implements types.MetricProvider interface for validation metrics.
type ValidationMetricProvider struct {
	mu sync.RWMutex

	// Validation operation counters
	validationCounts map[string]int64

	// Validation durations
	validationDurations map[string][]time.Duration

	// Error counts
	errorCounts map[string]int64

	// Cache access counts
	cacheAccessCounts map[string]int64

	// Constraint check counts
	constraintCheckCounts map[string]int64

	// Last collection time
	lastCollection time.Time
}

// NewValidationMetricProvider creates a new validation metric provider.
func NewValidationMetricProvider() *ValidationMetricProvider {
	return &ValidationMetricProvider{
		validationCounts:      make(map[string]int64),
		validationDurations:   make(map[string][]time.Duration),
		errorCounts:           make(map[string]int64),
		cacheAccessCounts:     make(map[string]int64),
		constraintCheckCounts: make(map[string]int64),
		lastCollection:        time.Now(),
	}
}

// Metrics returns validation metric definitions.
func (p *ValidationMetricProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		// Validation operation counter
		types.NewCounterDefinition(
			"magicorm_validation_operations_total",
			"Total number of validation operations",
			[]string{"operation", "model", "scenario", "status"},
			map[string]string{"component": "validation"},
		),

		// Validation duration histogram
		types.NewHistogramDefinition(
			"magicorm_validation_duration_seconds",
			"Duration of validation operations in seconds",
			[]string{"operation", "model", "scenario", "status"},
			[]float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			map[string]string{"component": "validation"},
		),

		// Validation error counter
		types.NewCounterDefinition(
			"magicorm_validation_errors_total",
			"Total number of validation errors",
			[]string{"operation", "model", "scenario", "error_type"},
			map[string]string{"component": "validation"},
		),

		// Cache access counter
		types.NewCounterDefinition(
			"magicorm_validation_cache_operations_total",
			"Total number of validation cache operations",
			[]string{"cache_type", "operation", "hit"},
			map[string]string{"component": "validation"},
		),

		// Constraint check counter
		types.NewCounterDefinition(
			"magicorm_validation_constraint_checks_total",
			"Total number of constraint checks",
			[]string{"constraint_type", "field", "result"},
			map[string]string{"component": "validation"},
		),

		// Validation layer performance gauge
		types.NewGaugeDefinition(
			"magicorm_validation_layer_performance_seconds",
			"Validation layer performance in seconds",
			[]string{"layer"},
			map[string]string{"component": "validation"},
		),
	}
}

// Collect collects current validation metrics.
func (p *ValidationMetricProvider) Collect() ([]types.Metric, *types.Error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var metrics []types.Metric

	// Collect validation operation counts
	for key, count := range p.validationCounts {
		parts := splitKey(key)
		if len(parts) >= 4 {
			labels := map[string]string{
				"operation": parts[0],
				"model":     parts[1],
				"scenario":  parts[2],
				"status":    parts[3],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_validation_operations_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect validation error counts
	for key, count := range p.errorCounts {
		parts := splitKey(key)
		if len(parts) >= 4 {
			labels := map[string]string{
				"operation":  parts[0],
				"model":      parts[1],
				"scenario":   parts[2],
				"error_type": parts[3],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_validation_errors_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect cache access counts
	for key, count := range p.cacheAccessCounts {
		parts := splitKey(key)
		if len(parts) >= 3 {
			labels := map[string]string{
				"cache_type": parts[0],
				"operation":  parts[1],
				"hit":        parts[2],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_validation_cache_operations_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect constraint check counts
	for key, count := range p.constraintCheckCounts {
		parts := splitKey(key)
		if len(parts) >= 3 {
			labels := map[string]string{
				"constraint_type": parts[0],
				"field":           parts[1],
				"result":          parts[2],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_validation_constraint_checks_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect validation durations
	for key, durations := range p.validationDurations {
		if len(durations) > 0 {
			parts := splitKey(key)
			if len(parts) >= 4 {
				labels := map[string]string{
					"operation": parts[0],
					"model":     parts[1],
					"scenario":  parts[2],
					"status":    parts[3],
				}

				// Record each duration as histogram observation
				for _, duration := range durations {
					metrics = append(metrics, types.NewMetric(
						"magicorm_validation_duration_seconds",
						types.HistogramMetric,
						duration.Seconds(),
						labels,
					))
				}
			}
		}
	}

	p.lastCollection = time.Now()

	return metrics, nil
}

// RecordValidation records a validation operation.
func (p *ValidationMetricProvider) RecordValidation(operation, model, scenario string, success bool, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}

	key := buildKey(operation, model, scenario, status)

	// Increment validation count
	p.validationCounts[key]++

	// Record duration
	if p.validationDurations[key] == nil {
		p.validationDurations[key] = make([]time.Duration, 0)
	}
	p.validationDurations[key] = append(p.validationDurations[key], duration)
}

// RecordValidationError records a validation error.
func (p *ValidationMetricProvider) RecordValidationError(operation, model, scenario, errorType string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := buildKey(operation, model, scenario, errorType)
	p.errorCounts[key]++
}

// RecordCacheAccess records validation cache access.
func (p *ValidationMetricProvider) RecordCacheAccess(cacheType, operation string, hit bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	hitStr := "false"
	if hit {
		hitStr = "true"
	}

	key := buildKey(cacheType, operation, hitStr)
	p.cacheAccessCounts[key]++
}

// RecordConstraintCheck records a constraint check.
func (p *ValidationMetricProvider) RecordConstraintCheck(constraintType, field string, passed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := "failed"
	if passed {
		result = "passed"
	}

	key := buildKey(constraintType, field, result)
	p.constraintCheckCounts[key]++
}

// Reset clears all collected metrics.
func (p *ValidationMetricProvider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.validationCounts = make(map[string]int64)
	p.validationDurations = make(map[string][]time.Duration)
	p.errorCounts = make(map[string]int64)
	p.cacheAccessCounts = make(map[string]int64)
	p.constraintCheckCounts = make(map[string]int64)
	p.lastCollection = time.Now()
}

// GetLastCollectionTime returns the last collection time.
func (p *ValidationMetricProvider) GetLastCollectionTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.lastCollection
}

// Helper functions

func splitKey(key string) []string {
	parts := strings.Split(key, ":")
	if len(parts) < 3 {
		// Return default values if key format is wrong
		return make([]string, 4) // Return empty strings
	}
	return parts
}

func buildKey(parts ...string) string {
	return strings.Join(parts, ":")
}
