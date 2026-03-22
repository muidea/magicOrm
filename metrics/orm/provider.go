// Package orm provides ORM-specific metric definitions for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package orm

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/muidea/magicOrm/metrics"
)

// ORMMetricProvider implements types.MetricProvider interface for ORM metrics.
type ORMMetricProvider struct {
	*types.BaseProvider
	collector *ORMMetricsCollector
}

// NewORMMetricProvider creates a new ORM metric provider.
func NewORMMetricProvider(collector *ORMMetricsCollector) *ORMMetricProvider {
	base := types.NewBaseProvider(
		"magicorm_orm",
		"1.0.0",
		"MagicORM ORM operation monitoring provider",
	)
	base.AddTag("orm")
	base.AddTag("magicorm")

	return &ORMMetricProvider{
		BaseProvider: base,
		collector:    collector,
	}
}

// Metrics returns ORM metric definitions.
func (p *ORMMetricProvider) Metrics() []types.MetricDefinition {
	ormLabels := metrics.MergeLabels(metrics.DefaultLabels(), map[string]string{"component": "orm"})

	return []types.MetricDefinition{
		// ORM operation counter
		types.NewCounterDefinition(
			"magicorm_orm_operations_total",
			"Total number of ORM operations",
			[]string{"operation", "model", "status"},
			ormLabels,
		),

		// ORM operation duration gauge (average)
		types.NewGaugeDefinition(
			"magicorm_orm_operation_duration_seconds",
			"Average duration of ORM operations in seconds",
			[]string{"operation", "model", "status"},
			ormLabels,
		),

		// ORM error counter
		types.NewCounterDefinition(
			"magicorm_orm_errors_total",
			"Total number of ORM errors",
			[]string{"operation", "model", "error_type"},
			ormLabels,
		),

		// ORM transaction counter
		types.NewCounterDefinition(
			"magicorm_orm_transactions_total",
			"Total number of ORM transactions",
			[]string{"type", "status"},
			ormLabels,
		),

		// ORM cache hit ratio gauge
		types.NewGaugeDefinition(
			"magicorm_orm_cache_hit_ratio",
			"ORM cache hit ratio",
			[]string{"cache_type"},
			ormLabels,
		),

		// Active connections gauge
		types.NewGaugeDefinition(
			"magicorm_orm_active_connections",
			"Number of active ORM connections",
			[]string{"database"},
			ormLabels,
		),
	}
}

// Collect collects current ORM metrics.
func (p *ORMMetricProvider) Collect() ([]types.Metric, *types.Error) {
	startTime := time.Now()

	if p.collector == nil {
		// No collector available, return empty metrics
		p.UpdateCollectionStats(true, time.Since(startTime), 0)
		return []types.Metric{}, nil
	}

	var metricList []types.Metric

	// Collect operation counters
	operationCounters := p.collector.GetOperationCounters()
	for key, count := range operationCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 3 {
			operation, model, status := parts[0], parts[1], parts[2]
			metricList = append(metricList, types.NewCounter(
				"magicorm_orm_operations_total",
				float64(count),
				map[string]string{
					"operation": operation,
					"model":     model,
					"status":    status,
				},
			))
		}
	}

	// Collect error counters
	errorCounters := p.collector.GetErrorCounters()
	for key, count := range errorCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 3 {
			operation, model, errorType := parts[0], parts[1], parts[2]
			metricList = append(metricList, types.NewCounter(
				"magicorm_orm_errors_total",
				float64(count),
				map[string]string{
					"operation":  operation,
					"model":      model,
					"error_type": errorType,
				},
			))
		}
	}

	// Collect operation durations (calculate average)
	operationDurations := p.collector.GetOperationDurations()
	for key, durations := range operationDurations {
		if len(durations) > 0 {
			parts := metrics.ParseKey(key)
			if len(parts) >= 3 {
				operation, model, status := parts[0], parts[1], parts[2]

				metricList = append(metricList, types.NewGauge(
					"magicorm_orm_operation_duration_seconds",
					metrics.AverageDurationSeconds(durations),
					map[string]string{
						"operation": operation,
						"model":     model,
						"status":    status,
					},
				))
			}
		}
	}

	// Collect transaction counters
	transactionCounters := p.collector.GetTransactionCounters()
	for key, count := range transactionCounters {
		parts := metrics.ParseKey(key)
		if len(parts) >= 2 {
			txType, status := parts[0], parts[1]
			metricList = append(metricList, types.NewCounter(
				"magicorm_orm_transactions_total",
				float64(count),
				map[string]string{
					"type":   txType,
					"status": status,
				},
			))
		}
	}

	// Collect cache statistics
	cacheHits, cacheMisses := p.collector.GetCacheStats()
	if cacheHits+cacheMisses > 0 {
		hitRatio := float64(cacheHits) / float64(cacheHits+cacheMisses)
		metricList = append(metricList, types.NewGauge(
			"magicorm_orm_cache_hit_ratio",
			hitRatio,
			map[string]string{"cache_type": "default"},
		))
	}

	// Collect active connections
	activeConnections := p.collector.GetActiveConnections()
	metricList = append(metricList, types.NewGauge(
		"magicorm_orm_active_connections",
		float64(activeConnections),
		map[string]string{"database": "default"},
	))

	// Update collection statistics
	duration := time.Since(startTime)
	p.UpdateCollectionStats(true, duration, len(metricList))

	return metricList, nil
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

// parseKey parses a metric key into its components.
func parseKey(key string) []string {
	return metrics.ParseKey(key)
}
