// Package metricsdb provides database-specific metric definitions for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package metricsdb

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/muidea/magicOrm/metrics"
)

// DatabaseMetricProvider implements types.MetricProvider interface for database metrics.
type DatabaseMetricProvider struct {
	*types.BaseProvider
	collector *DatabaseMetricsCollector
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

// NewDatabaseMetricProviderWithCollector creates a new database metric provider with a collector.
func NewDatabaseMetricProviderWithCollector(collector *DatabaseMetricsCollector) *DatabaseMetricProvider {
	base := types.NewBaseProvider(
		"magicorm_database",
		"1.0.0",
		"MagicORM database monitoring provider",
	)
	base.AddTag("database")
	base.AddTag("magicorm")

	return &DatabaseMetricProvider{
		BaseProvider: base,
		collector:    collector,
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

		// Query duration gauge (average)
		types.NewGaugeDefinition(
			"magicorm_database_query_duration_seconds",
			"Average duration of database queries in seconds",
			[]string{"database", "query_type", "status"},
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
	startTime := time.Now()

	if p.collector == nil {
		// No collector available, return empty metrics
		p.UpdateCollectionStats(true, time.Since(startTime), 0)
		return []types.Metric{}, nil
	}

	var metricList []types.Metric

	// Collect query counters
	queryCounters := p.collector.GetQueryCounters()
	for key, count := range queryCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 3 {
			database, queryType, status := parts[0], parts[1], parts[2]
			metricList = append(metricList, types.NewCounter(
				"magicorm_database_queries_total",
				float64(count),
				map[string]string{
					"database":   database,
					"query_type": queryType,
					"status":     status,
				},
			))
		}
	}

	// Collect error counters
	errorCounters := p.collector.GetErrorCounters()
	for key, count := range errorCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 3 {
			database, operation, errorType := parts[0], parts[1], parts[2]
			metricList = append(metricList, types.NewCounter(
				"magicorm_database_errors_total",
				float64(count),
				map[string]string{
					"database":   database,
					"operation":  operation,
					"error_type": errorType,
				},
			))
		}
	}

	// Collect query durations (calculate average)
	queryDurations := p.collector.GetQueryDurations()
	for key, durations := range queryDurations {
		if len(durations) > 0 {
			parts := metrics.ParseKey(key)
			if len(parts) >= 3 {
				database, queryType, status := parts[0], parts[1], parts[2]

				// Calculate average duration in seconds
				var total time.Duration
				for _, d := range durations {
					total += d
				}
				avgDuration := total.Seconds() / float64(len(durations))

				metricList = append(metricList, types.NewGauge(
					"magicorm_database_query_duration_seconds",
					avgDuration,
					map[string]string{
						"database":   database,
						"query_type": queryType,
						"status":     status,
					},
				))
			}
		}
	}

	// Collect transaction counters
	transactionCounters := p.collector.GetTransactionCounters()
	for key, count := range transactionCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 3 {
			database, txType, status := parts[0], parts[1], parts[2]
			metricList = append(metricList, types.NewCounter(
				"magicorm_database_transactions_total",
				float64(count),
				map[string]string{
					"database": database,
					"type":     txType,
					"status":   status,
				},
			))
		}
	}

	// Collect execution counters
	executionCounters := p.collector.GetExecutionCounters()
	for key, count := range executionCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 3 {
			database, operation, status := parts[0], parts[1], parts[2]
			metricList = append(metricList, types.NewCounter(
				"magicorm_database_executions_total",
				float64(count),
				map[string]string{
					"database":  database,
					"operation": operation,
					"status":    status,
				},
			))
		}
	}

	// Collect connection statistics
	connectionStats := p.collector.GetConnectionStats()
	for key, count := range connectionStats {
		parts := metrics.ParseKey(key)
		if len(parts) >= 2 {
			database, state := parts[0], parts[1]
			metricList = append(metricList, types.NewGauge(
				"magicorm_database_connections",
				float64(count),
				map[string]string{
					"database": database,
					"state":    state,
				},
			))
		}
	}

	// Update collection statistics
	duration := time.Since(startTime)
	p.UpdateCollectionStats(true, duration, len(metricList))

	return metricList, nil
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

// parseKey parses a metric key into its components.
func parseKey(key string) []string {
	return metrics.ParseKey(key)
}
