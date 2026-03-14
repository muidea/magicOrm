// Package metricsval provides validation-specific metric definitions for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package validation

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/muidea/magicOrm/metrics"
)

// ValidationMetricProvider implements types.MetricProvider interface for validation metrics.
type ValidationMetricProvider struct {
	*types.BaseProvider
	collector *ValidationMetricsCollector
}

// NewValidationMetricProvider creates a new validation metric provider.
func NewValidationMetricProvider() *ValidationMetricProvider {
	base := types.NewBaseProvider(
		"magicorm_validation",
		"1.0.0",
		"MagicORM validation monitoring provider",
	)
	base.AddTag("validation")
	base.AddTag("magicorm")

	return &ValidationMetricProvider{
		BaseProvider: base,
	}
}

// NewValidationMetricProviderWithCollector creates a new validation metric provider with a collector.
func NewValidationMetricProviderWithCollector(collector *ValidationMetricsCollector) *ValidationMetricProvider {
	base := types.NewBaseProvider(
		"magicorm_validation",
		"1.0.0",
		"MagicORM validation monitoring provider",
	)
	base.AddTag("validation")
	base.AddTag("magicorm")

	return &ValidationMetricProvider{
		BaseProvider: base,
		collector:    collector,
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

		// Validation duration gauge (average)
		types.NewGaugeDefinition(
			"magicorm_validation_duration_seconds",
			"Average duration of validation operations in seconds",
			[]string{"operation", "model", "scenario", "status"},
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
			"magicorm_validation_cache_access_total",
			"Total number of cache accesses",
			[]string{"cache_type", "hit_miss"},
			map[string]string{"component": "validation"},
		),

		// Constraint check counter
		types.NewCounterDefinition(
			"magicorm_validation_constraint_checks_total",
			"Total number of constraint checks",
			[]string{"constraint_type", "field", "status"},
			map[string]string{"component": "validation"},
		),

		// Cache hit ratio gauge
		types.NewGaugeDefinition(
			"magicorm_validation_cache_hit_ratio",
			"Validation cache hit ratio",
			[]string{"cache_type"},
			map[string]string{"component": "validation"},
		),
	}
}

// Collect collects current validation metrics.
func (p *ValidationMetricProvider) Collect() ([]types.Metric, *types.Error) {
	startTime := time.Now()

	if p.collector == nil {
		// No collector available, return empty metrics
		p.UpdateCollectionStats(true, time.Since(startTime), 0)
		return []types.Metric{}, nil
	}

	var metricList []types.Metric

	// Collect validation counters
	validationCounters := p.collector.GetValidationCounters()
	for key, count := range validationCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 4 {
			operation, model, scenario, status := parts[0], parts[1], parts[2], parts[3]
			metricList = append(metricList, types.NewCounter(
				"magicorm_validation_operations_total",
				float64(count),
				map[string]string{
					"operation": operation,
					"model":     model,
					"scenario":  scenario,
					"status":    status,
				},
			))
		}
	}

	// Collect error counters
	errorCounters := p.collector.GetErrorCounters()
	for key, count := range errorCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 4 {
			operation, model, scenario, errorType := parts[0], parts[1], parts[2], parts[3]
			metricList = append(metricList, types.NewCounter(
				"magicorm_validation_errors_total",
				float64(count),
				map[string]string{
					"operation":  operation,
					"model":      model,
					"scenario":   scenario,
					"error_type": errorType,
				},
			))
		}
	}

	// Collect validation durations (calculate average)
	validationDurations := p.collector.GetValidationDurations()
	for key, durations := range validationDurations {
		if len(durations) > 0 {
			parts := metrics.ParseKey(key)
			if len(parts) >= 4 {
				operation, model, scenario, status := parts[0], parts[1], parts[2], parts[3]

				metricList = append(metricList, types.NewGauge(
					"magicorm_validation_duration_seconds",
					metrics.AverageDurationSeconds(durations),
					map[string]string{
						"operation": operation,
						"model":     model,
						"scenario":  scenario,
						"status":    status,
					},
				))
			}
		}
	}

	// Collect cache access counters
	cacheAccessCounters := p.collector.GetCacheAccessCounters()
	for key, count := range cacheAccessCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 2 {
			cacheType, hitMiss := parts[0], parts[1]
			metricList = append(metricList, types.NewCounter(
				"magicorm_validation_cache_access_total",
				float64(count),
				map[string]string{
					"cache_type": cacheType,
					"hit_miss":   hitMiss,
				},
			))
		}
	}

	// Collect constraint check counters
	constraintCheckCounters := p.collector.GetConstraintCheckCounters()
	for key, count := range constraintCheckCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 3 {
			constraintType, field, status := parts[0], parts[1], parts[2]
			metricList = append(metricList, types.NewCounter(
				"magicorm_validation_constraint_checks_total",
				float64(count),
				map[string]string{
					"constraint_type": constraintType,
					"field":           field,
					"status":          status,
				},
			))
		}
	}

	// Collect cache hit ratio for each observed cache type.
	for _, cacheType := range p.collector.GetCacheTypes() {
		metricList = append(metricList, types.NewGauge(
			"magicorm_validation_cache_hit_ratio",
			p.collector.GetCacheHitRatio(cacheType),
			map[string]string{"cache_type": cacheType},
		))
	}

	// Update collection statistics
	duration := time.Since(startTime)
	p.UpdateCollectionStats(true, duration, len(metricList))

	return metricList, nil
}

// Name returns the provider name.
func (p *ValidationMetricProvider) Name() string {
	return "magicorm_validation"
}

// Init initializes the provider.
func (p *ValidationMetricProvider) Init(collector interface{}) *types.Error {
	// Optional initialization logic
	p.UpdateHealthStatus(types.ProviderHealthy)
	return nil
}

// Shutdown cleans up provider resources.
func (p *ValidationMetricProvider) Shutdown() *types.Error {
	// No resources to clean up
	return nil
}

// parseKey parses a metric key into its components.
func parseKey(key string) []string {
	return metrics.ParseKey(key)
}
