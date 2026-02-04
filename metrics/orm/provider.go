// Package orm provides ORM-specific metric definitions for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package orm

import (
	"strings"
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
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
	startTime := time.Now()

	if p.collector == nil {
		// No collector available, return empty metrics
		p.UpdateCollectionStats(true, time.Since(startTime), 0)
		return []types.Metric{}, nil
	}

	var metrics []types.Metric

	// Collect operation counters
	operationCounters := p.collector.GetOperationCounters()
	for key, count := range operationCounters {
		parts := parseKey(key)
		if len(parts) >= 3 {
			operation, model, status := parts[0], parts[1], parts[2]
			metrics = append(metrics, types.NewCounter(
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
		parts := parseKey(key)
		if len(parts) >= 3 {
			operation, model, errorType := parts[0], parts[1], parts[2]
			metrics = append(metrics, types.NewCounter(
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
			parts := parseKey(key)
			if len(parts) >= 3 {
				operation, model, status := parts[0], parts[1], parts[2]

				// Calculate average duration in seconds
				var total time.Duration
				for _, d := range durations {
					total += d
				}
				avgDuration := total.Seconds() / float64(len(durations))

				metrics = append(metrics, types.NewMetric(
					"magicorm_orm_operation_duration_seconds",
					types.HistogramMetric,
					avgDuration,
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
		parts := parseKey(key)
		if len(parts) >= 2 {
			txType, status := parts[0], parts[1]
			metrics = append(metrics, types.NewCounter(
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
		metrics = append(metrics, types.NewGauge(
			"magicorm_orm_cache_hit_ratio",
			hitRatio,
			map[string]string{"cache_type": "default"},
		))
	}

	// Collect active connections
	activeConnections := p.collector.GetActiveConnections()
	metrics = append(metrics, types.NewGauge(
		"magicorm_orm_active_connections",
		float64(activeConnections),
		map[string]string{"database": "default"},
	))

	// Update collection statistics
	duration := time.Since(startTime)
	p.UpdateCollectionStats(true, duration, len(metrics))

	return metrics, nil
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
	return strings.Split(key, "_")
}
