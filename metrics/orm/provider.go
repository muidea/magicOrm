// Package orm provides ORM-specific metric definitions for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package orm

import (
	"github.com/muidea/magicCommon/monitoring/types"
)

// ORMMetricProvider implements types.MetricProvider interface for ORM metrics.
type ORMMetricProvider struct {
	*types.BaseProvider
}

// NewORMMetricProvider creates a new ORM metric provider.
func NewORMMetricProvider() *ORMMetricProvider {
	base := types.NewBaseProvider(
		"magicorm_orm",
		"1.0.0",
		"MagicORM ORM operation monitoring provider",
	)
	base.AddTag("orm")
	base.AddTag("magicorm")

	return &ORMMetricProvider{
		BaseProvider: base,
	}
}

// Metrics returns ORM metric definitions.
func (p *ORMMetricProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		// ORM operation counter
		types.NewCounterDefinition(
			"magicorm_orm_operations_total",
			"Total number of ORM operations",
			[]string{"operation", "model", "status"},
			map[string]string{"component": "orm"},
		),

		// ORM operation duration histogram
		types.NewHistogramDefinition(
			"magicorm_orm_operation_duration_seconds",
			"Duration of ORM operations in seconds",
			[]string{"operation", "model", "status"},
			[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
			map[string]string{"component": "orm"},
		),

		// ORM error counter
		types.NewCounterDefinition(
			"magicorm_orm_errors_total",
			"Total number of ORM errors",
			[]string{"operation", "model", "error_type"},
			map[string]string{"component": "orm"},
		),

		// ORM transaction counter
		types.NewCounterDefinition(
			"magicorm_orm_transactions_total",
			"Total number of ORM transactions",
			[]string{"type", "status"},
			map[string]string{"component": "orm"},
		),

		// ORM cache hit ratio gauge
		types.NewGaugeDefinition(
			"magicorm_orm_cache_hit_ratio",
			"ORM cache hit ratio",
			[]string{"cache_type"},
			map[string]string{"component": "orm"},
		),

		// Active connections gauge
		types.NewGaugeDefinition(
			"magicorm_orm_active_connections",
			"Number of active ORM connections",
			[]string{"database"},
			map[string]string{"component": "orm"},
		),
	}
}

// Collect collects current ORM metrics.
func (p *ORMMetricProvider) Collect() ([]types.Metric, *types.Error) {
	// MagicORM no longer collects data - only provides metric definitions
	// Data collection is handled by magicCommon/monitoring system
	return []types.Metric{}, nil
}

// Name returns the provider name.
func (p *ORMMetricProvider) Name() string {
	return "magicorm_orm"
}

// Init initializes the provider.
func (p *ORMMetricProvider) Init(collector interface{}) *types.Error {
	// Optional initialization logic
	p.UpdateHealthStatus(types.ProviderHealthy)
	return nil
}

// Shutdown cleans up provider resources.
func (p *ORMMetricProvider) Shutdown() *types.Error {
	// No resources to clean up
	return nil
}
