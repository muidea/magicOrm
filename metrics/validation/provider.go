// Package validation provides validation-specific metric definitions for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package validation

import (
	"github.com/muidea/magicCommon/monitoring/types"
)

// ValidationMetricProvider implements types.MetricProvider interface for validation metrics.
type ValidationMetricProvider struct {
	*types.BaseProvider
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
			[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
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
	// MagicORM no longer collects data - only provides metric definitions
	// Data collection is handled by magicCommon/monitoring system
	return []types.Metric{}, nil
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
