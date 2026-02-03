// Package database provides database-specific metric collectors for MagicORM.
// This package implements MetricProvider interface for integration with magicCommon/monitoring.
package database

import (
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
)

// DatabaseMetricProvider implements types.MetricProvider interface for database metrics.
type DatabaseMetricProvider struct {
	mu sync.RWMutex

	// Query counters
	queryCounts map[string]int64

	// Query durations
	queryDurations map[string][]time.Duration

	// Transaction counters
	transactionCounts map[string]int64

	// Error counts
	errorCounts map[string]int64

	// Connection statistics
	connectionStats map[string]ConnectionStats

	// Execution counters
	executionCounts map[string]int64

	// Last collection time
	lastCollection time.Time
}

// ConnectionStats holds connection pool statistics
type ConnectionStats struct {
	ActiveConnections  int
	IdleConnections    int
	WaitingConnections int
	MaxConnections     int
	LastUpdated        time.Time
}

// NewDatabaseMetricProvider creates a new database metric provider.
func NewDatabaseMetricProvider() *DatabaseMetricProvider {
	return &DatabaseMetricProvider{
		queryCounts:       make(map[string]int64),
		queryDurations:    make(map[string][]time.Duration),
		transactionCounts: make(map[string]int64),
		errorCounts:       make(map[string]int64),
		connectionStats:   make(map[string]ConnectionStats),
		executionCounts:   make(map[string]int64),
		lastCollection:    time.Now(),
	}
}

// Metrics returns database metric definitions.
func (p *DatabaseMetricProvider) Metrics() []types.MetricDefinition {
	return []types.MetricDefinition{
		// Database query counter
		types.NewCounterDefinition(
			"magicorm_database_queries_total",
			"Total number of database queries",
			[]string{"db_type", "query_type", "status"},
			map[string]string{"component": "database"},
		),

		// Database query duration histogram
		types.NewHistogramDefinition(
			"magicorm_database_query_duration_seconds",
			"Duration of database queries in seconds",
			[]string{"db_type", "query_type", "status"},
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			map[string]string{"component": "database"},
		),

		// Database transaction counter
		types.NewCounterDefinition(
			"magicorm_database_transactions_total",
			"Total number of database transactions",
			[]string{"db_type", "operation", "status"},
			map[string]string{"component": "database"},
		),

		// Database error counter
		types.NewCounterDefinition(
			"magicorm_database_errors_total",
			"Total number of database errors",
			[]string{"db_type", "operation", "error_type"},
			map[string]string{"component": "database"},
		),

		// Database execution counter
		types.NewCounterDefinition(
			"magicorm_database_executions_total",
			"Total number of database executions",
			[]string{"db_type", "operation", "status"},
			map[string]string{"component": "database"},
		),

		// Active connections gauge
		types.NewGaugeDefinition(
			"magicorm_database_connections_active",
			"Number of active database connections",
			[]string{"db_type"},
			map[string]string{"component": "database"},
		),

		// Idle connections gauge
		types.NewGaugeDefinition(
			"magicorm_database_connections_idle",
			"Number of idle database connections",
			[]string{"db_type"},
			map[string]string{"component": "database"},
		),

		// Waiting connections gauge
		types.NewGaugeDefinition(
			"magicorm_database_connections_waiting",
			"Number of waiting database connections",
			[]string{"db_type"},
			map[string]string{"component": "database"},
		),

		// Max connections gauge
		types.NewGaugeDefinition(
			"magicorm_database_connections_max",
			"Maximum number of database connections",
			[]string{"db_type"},
			map[string]string{"component": "database"},
		),
	}
}

// Collect collects current database metrics.
func (p *DatabaseMetricProvider) Collect() ([]types.Metric, *types.Error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var metrics []types.Metric

	// Collect query counts
	for key, count := range p.queryCounts {
		parts := splitKey(key)
		if len(parts) >= 3 {
			labels := map[string]string{
				"db_type":    parts[0],
				"query_type": parts[1],
				"status":     parts[2],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_database_queries_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect transaction counts
	for key, count := range p.transactionCounts {
		parts := splitKey(key)
		if len(parts) >= 3 {
			labels := map[string]string{
				"db_type":   parts[0],
				"operation": parts[1],
				"status":    parts[2],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_database_transactions_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect error counts
	for key, count := range p.errorCounts {
		parts := splitKey(key)
		if len(parts) >= 3 {
			labels := map[string]string{
				"db_type":    parts[0],
				"operation":  parts[1],
				"error_type": parts[2],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_database_errors_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect execution counts
	for key, count := range p.executionCounts {
		parts := splitKey(key)
		if len(parts) >= 3 {
			labels := map[string]string{
				"db_type":   parts[0],
				"operation": parts[1],
				"status":    parts[2],
			}

			metrics = append(metrics, types.NewCounter(
				"magicorm_database_executions_total",
				float64(count),
				labels,
			))
		}
	}

	// Collect connection statistics
	for dbType, stats := range p.connectionStats {
		// Active connections
		metrics = append(metrics, types.NewGauge(
			"magicorm_database_connections_active",
			float64(stats.ActiveConnections),
			map[string]string{"db_type": dbType},
		))

		// Idle connections
		metrics = append(metrics, types.NewGauge(
			"magicorm_database_connections_idle",
			float64(stats.IdleConnections),
			map[string]string{"db_type": dbType},
		))

		// Waiting connections
		metrics = append(metrics, types.NewGauge(
			"magicorm_database_connections_waiting",
			float64(stats.WaitingConnections),
			map[string]string{"db_type": dbType},
		))

		// Max connections
		metrics = append(metrics, types.NewGauge(
			"magicorm_database_connections_max",
			float64(stats.MaxConnections),
			map[string]string{"db_type": dbType},
		))
	}

	// Collect query durations
	for key, durations := range p.queryDurations {
		if len(durations) > 0 {
			parts := splitKey(key)
			if len(parts) >= 3 {
				labels := map[string]string{
					"db_type":    parts[0],
					"query_type": parts[1],
					"status":     parts[2],
				}

				// Record each duration as histogram observation
				for _, duration := range durations {
					metrics = append(metrics, types.NewMetric(
						"magicorm_database_query_duration_seconds",
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

// RecordQuery records a database query.
func (p *DatabaseMetricProvider) RecordQuery(dbType, queryType string, success bool, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}

	key := buildKey(dbType, queryType, status)

	// Increment query count
	p.queryCounts[key]++

	// Record duration
	if p.queryDurations[key] == nil {
		p.queryDurations[key] = make([]time.Duration, 0)
	}
	p.queryDurations[key] = append(p.queryDurations[key], duration)
}

// RecordTransaction records a database transaction.
func (p *DatabaseMetricProvider) RecordTransaction(dbType, operation string, success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}

	key := buildKey(dbType, operation, status)
	p.transactionCounts[key]++
}

// RecordError records a database error.
func (p *DatabaseMetricProvider) RecordError(dbType, operation, errorType string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := buildKey(dbType, operation, errorType)
	p.errorCounts[key]++
}

// RecordExecution records a database execution.
func (p *DatabaseMetricProvider) RecordExecution(dbType, operation string, success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := "success"
	if !success {
		status = "error"
	}

	key := buildKey(dbType, operation, status)
	p.executionCounts[key]++
}

// RecordConnectionStats records connection pool statistics.
func (p *DatabaseMetricProvider) RecordConnectionStats(dbType string, active, idle, waiting, max int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connectionStats[dbType] = ConnectionStats{
		ActiveConnections:  active,
		IdleConnections:    idle,
		WaitingConnections: waiting,
		MaxConnections:     max,
		LastUpdated:        time.Now(),
	}
}

// Reset clears all collected metrics.
func (p *DatabaseMetricProvider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queryCounts = make(map[string]int64)
	p.queryDurations = make(map[string][]time.Duration)
	p.transactionCounts = make(map[string]int64)
	p.errorCounts = make(map[string]int64)
	p.connectionStats = make(map[string]ConnectionStats)
	p.executionCounts = make(map[string]int64)
	p.lastCollection = time.Now()
}

// GetLastCollectionTime returns the last collection time.
func (p *DatabaseMetricProvider) GetLastCollectionTime() time.Time {
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
