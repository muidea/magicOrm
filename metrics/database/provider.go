// Package database provides database-specific metric definitions for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package database

import (
	"github.com/muidea/magicCommon/monitoring/types"
)

// DatabaseMetricProvider implements types.MetricProvider interface for database metrics.
type DatabaseMetricProvider struct {
	*types.BaseProvider
}

// NewDatabaseMetricProvider creates a new database metric provider.
func NewDatabaseMetricProvider() *DatabaseMetricProvider {
	base := types.NewBaseProvider(
		"magicorm_database",
		"1.0.0",
		"MagicORM database monitoring provider",
	)
	base.AddTag("database")
	base.AddTag("magicorm")

	return &DatabaseMetricProvider{
		BaseProvider: base,
	}
}

// Metrics returns database metric definitions.
func (p *DatabaseMetricProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		// Database query counter
		types.NewCounterDefinition(
			"magicorm_database_queries_total",
			"Total number of database queries",
			[]string{"database", "query_type", "status"},
			map[string]string{"component": "database"},
		),

		// Query duration histogram
		types.NewHistogramDefinition(
			"magicorm_database_query_duration_seconds",
			"Duration of database queries in seconds",
			[]string{"database", "query_type", "status"},
			[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0, 10.0},
			map[string]string{"component": "database"},
		),

		// Database error counter
		types.NewCounterDefinition(
			"magicorm_database_errors_total",
			"Total number of database errors",
			[]string{"database", "operation", "error_type"},
			map[string]string{"component": "database"},
		),

		// Transaction counter
		types.NewCounterDefinition(
			"magicorm_database_transactions_total",
			"Total number of database transactions",
			[]string{"database", "type", "status"},
			map[string]string{"component": "database"},
		),

		// Connection pool gauge
		types.NewGaugeDefinition(
			"magicorm_database_connections",
			"Database connection pool statistics",
			[]string{"database", "state"},
			map[string]string{"component": "database"},
		),

		// Execution counter
		types.NewCounterDefinition(
			"magicorm_database_executions_total",
			"Total number of database executions",
			[]string{"database", "operation", "status"},
			map[string]string{"component": "database"},
		),
	}
}

// Collect collects current database metrics.
func (p *DatabaseMetricProvider) Collect() ([]types.Metric, *types.Error) {
	// MagicORM no longer collects data - only provides metric definitions
	// Data collection is handled by magicCommon/monitoring system
	return []types.Metric{}, nil
}

// Name returns the provider name.
func (p *DatabaseMetricProvider) Name() string {
	return "magicorm_database"
}

// Init initializes the provider.
func (p *DatabaseMetricProvider) Init(collector interface{}) *types.Error {
	// Optional initialization logic
	p.UpdateHealthStatus(types.ProviderHealthy)
	return nil
}

// Shutdown cleans up provider resources.
func (p *DatabaseMetricProvider) Shutdown() *types.Error {
	// No resources to clean up
	return nil
}
