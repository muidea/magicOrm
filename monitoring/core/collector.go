// Package core provides core monitoring types and interfaces.
// This is a simplified version that only provides type definitions and basic interfaces.
package core

import (
	"github.com/muidea/magicCommon/monitoring/types"
)

// MetricType represents the type of metric
type MetricType = types.MetricType

const (
	// CounterMetric represents a cumulative metric that only increases
	CounterMetric MetricType = types.CounterMetric
	// GaugeMetric represents a metric that can go up and down
	GaugeMetric MetricType = types.GaugeMetric
	// HistogramMetric represents a metric that samples observations
	HistogramMetric MetricType = types.HistogramMetric
	// SummaryMetric represents a metric that calculates quantiles
	SummaryMetric MetricType = types.SummaryMetric
)

// Metric represents a single metric with labels and value
type Metric = types.Metric

// MetricDefinition defines a metric's structure and behavior
type MetricDefinition = types.MetricDefinition

// CollectorStats holds collector statistics
type CollectorStats struct {
	MetricsCollected int64 `json:"metrics_collected"`
	BatchOperations  int64 `json:"batch_operations"`
	Errors           int64 `json:"errors"`
	LastCollection   int64 `json:"last_collection"`
}

// MetricError represents a metric-related error
type MetricError struct {
	Name    string
	Message string
}

func (e *MetricError) Error() string {
	return "metric error: " + e.Name + ": " + e.Message
}

// SimpleCollector is a minimal collector interface for backward compatibility
type SimpleCollector interface {
	// Record records a metric value
	Record(name string, value float64, labels map[string]string) error

	// Increment increments a counter metric
	Increment(name string, labels map[string]string) error

	// Decrement decrements a gauge metric
	Decrement(name string, labels map[string]string) error

	// Observe observes a value for a histogram or summary metric
	Observe(name string, value float64, labels map[string]string) error
}

// NoopCollector is a collector that does nothing (for backward compatibility)
type NoopCollector struct{}

// Record implements SimpleCollector interface
func (c *NoopCollector) Record(name string, value float64, labels map[string]string) error {
	return nil
}

// Increment implements SimpleCollector interface
func (c *NoopCollector) Increment(name string, labels map[string]string) error {
	return nil
}

// Decrement implements SimpleCollector interface
func (c *NoopCollector) Decrement(name string, labels map[string]string) error {
	return nil
}

// Observe implements SimpleCollector interface
func (c *NoopCollector) Observe(name string, value float64, labels map[string]string) error {
	return nil
}

// NewNoopCollector creates a new no-op collector
func NewNoopCollector() *NoopCollector {
	return &NoopCollector{}
}
