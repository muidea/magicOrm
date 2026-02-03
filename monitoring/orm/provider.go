// Package orm provides ORM-specific metric collectors for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package orm

import (
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

// ORMMetricProvider implements types.MetricProvider interface for ORM metrics.
type ORMMetricProvider struct {
	mu sync.RWMutex

	// Operation counters
	operationCounts map[string]int64

	// Operation durations
	operationDurations map[string][]time.Duration

	// Error counts
	errorCounts map[string]int64

	// Last collection time
	lastCollection time.Time
}

// NewORMMetricProvider creates a new ORM metric provider.
func NewORMMetricProvider() *ORMMetricProvider {
	return &ORMMetricProvider{
		operationCounts:    make(map[string]int64),
		operationDurations: make(map[string][]time.Duration),
		errorCounts:        make(map[string]int64),
		lastCollection:     time.Now(),
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
			[]float64{0.001, 0.005, -0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
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
	p.mu.Lock()
	defer p.mu.Unlock()

	var metrics []types.Metric

	// Collect operation counts
	for key, count := range p.operationCounts {
		// Parse key format: "operation:model:status"
		parts := splitKey(key)
		if len(parts) >= 3 {
			labels := map[string]string{
				"operation": parts[0],
				"model":     parts[1],
				"status":    parts[2],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_orm_operations_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect error counts
	for key, count := range p.errorCounts {
		// Parse key format: "operation:model:error_type"
		parts := splitKey(key)
		if len(parts) >= 3 {
			labels := map[string]string{
				"operation":  parts[0],
				"model":      parts[1],
				"error_type": parts[2],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_orm_errors_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect operation durations
	for key, durations := range p.operationDurations {
		if len(durations) > 0 {
			// Parse key format: "operation:model:status"
			parts := splitKey(key)
			if len(parts) >= 3 {
				labels := map[string]string{
					"operation": parts[0],
					"model":     parts[1],
					"status":    parts[2],
				}

				// Record each duration as histogram observation
				for _, duration := range durations {
					metrics = append(metrics, types.NewMetric(
						"magicorm_orm_operation_duration_seconds",
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

// RecordOperation records an ORM operation.
func (p *ORMMetricProvider) RecordOperation(operation, model string, success bool, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}

	key := buildKey(operation, model, status)

	// Increment operation count
	p.operationCounts[key]++

	// Record duration
	if p.operationDurations[key] == nil {
		p.operationDurations[key] = make([]time.Duration, 0)
	}
	p.operationDurations[key] = append(p.operationDurations[key], duration)
}

// RecordError records an ORM error.
func (p *ORMMetricProvider) RecordError(operation, model, errorType string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := buildKey(operation, model, errorType)
	p.errorCounts[key]++
}

// RecordTransaction records an ORM transaction.
func (p *ORMMetricProvider) RecordTransaction(txType string, success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}

	key := buildKey(txType, "transaction", status)
	p.operationCounts[key]++
}

// RecordCacheAccess records ORM cache access.
func (p *ORMMetricProvider) RecordCacheAccess(cacheType string, hit bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// This would update cache hit ratio metrics
	// For simplicity, we're just tracking counts here
	operation := "cache_access"
	status := "miss"
	if hit {
		status = "hit"
	}

	key := buildKey(operation, cacheType, status)
	p.operationCounts[key]++
}

// RecordConnection records ORM connection statistics.
func (p *ORMMetricProvider) RecordConnection(dbType string, active int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// This would update active connections gauge
	// For now, we're just tracking the count
	key := buildKey("connection", dbType, "active")
	p.operationCounts[key] = int64(active)
}

// Reset clears all collected metrics.
func (p *ORMMetricProvider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.operationCounts = make(map[string]int64)
	p.operationDurations = make(map[string][]time.Duration)
	p.errorCounts = make(map[string]int64)
	p.lastCollection = time.Now()
}

// GetLastCollectionTime returns the last collection time.
func (p *ORMMetricProvider) GetLastCollectionTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.lastCollection
}

// Helper functions

func splitKey(key string) []string {
	parts := strings.Split(key, ":")
	if len(parts) < 3 {
		// Return default values if key format is wrong
		return []string{"unknown", "unknown", "unknown"}
	}
	return parts
}

func buildKey(parts ...string) string {
	return strings.Join(parts, ":")
}
