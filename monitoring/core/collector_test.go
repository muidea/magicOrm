package core

import (
	"testing"
)

// TestCollector is a simple test implementation
type TestCollector struct {
	records    []Record
	increments []Increment
	decrements []Decrement
	observes   []Observe
}

type Record struct {
	name   string
	value  float64
	labels map[string]string
}

type Increment struct {
	name   string
	labels map[string]string
}

type Decrement struct {
	name   string
	labels map[string]string
}

type Observe struct {
	name   string
	value  float64
	labels map[string]string
}

func NewTestCollector() *TestCollector {
	return &TestCollector{
		records:    make([]Record, 0),
		increments: make([]Increment, 0),
		decrements: make([]Decrement, 0),
		observes:   make([]Observe, 0),
	}
}

func (c *TestCollector) Record(name string, value float64, labels map[string]string) error {
	c.records = append(c.records, Record{name, value, labels})
	return nil
}

func (c *TestCollector) Increment(name string, labels map[string]string) error {
	c.increments = append(c.increments, Increment{name, labels})
	return nil
}

func (c *TestCollector) Decrement(name string, labels map[string]string) error {
	c.decrements = append(c.decrements, Decrement{name, labels})
	return nil
}

func (c *TestCollector) Observe(name string, value float64, labels map[string]string) error {
	c.observes = append(c.observes, Observe{name, value, labels})
	return nil
}

func TestNoopCollector(t *testing.T) {
	collector := &NoopCollector{}

	// Test Record
	err := collector.Record("test_metric", 1.0, map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Record should not return error: %v", err)
	}

	// Test Increment
	err = collector.Increment("test_counter", map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Increment should not return error: %v", err)
	}

	// Test Decrement
	err = collector.Decrement("test_gauge", map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Decrement should not return error: %v", err)
	}

	// Test Observe
	err = collector.Observe("test_histogram", 2.5, map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Observe should not return error: %v", err)
	}
}

func TestTestCollector(t *testing.T) {
	collector := NewTestCollector()

	// Test Record
	err := collector.Record("test_metric", 1.0, map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Record should not return error: %v", err)
	}

	if len(collector.records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(collector.records))
	}

	record := collector.records[0]
	if record.name != "test_metric" {
		t.Errorf("Expected name 'test_metric', got '%s'", record.name)
	}

	if record.value != 1.0 {
		t.Errorf("Expected value 1.0, got %f", record.value)
	}

	if record.labels["label"] != "value" {
		t.Errorf("Expected label 'value', got '%s'", record.labels["label"])
	}

	// Test Increment
	err = collector.Increment("test_counter", map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Increment should not return error: %v", err)
	}

	if len(collector.increments) != 1 {
		t.Errorf("Expected 1 increment, got %d", len(collector.increments))
	}

	// Test Decrement
	err = collector.Decrement("test_gauge", map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Decrement should not return error: %v", err)
	}

	if len(collector.decrements) != 1 {
		t.Errorf("Expected 1 decrement, got %d", len(collector.decrements))
	}

	// Test Observe
	err = collector.Observe("test_histogram", 2.5, map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Observe should not return error: %v", err)
	}

	if len(collector.observes) != 1 {
		t.Errorf("Expected 1 observe, got %d", len(collector.observes))
	}
}

func TestMetricTypes(t *testing.T) {
	// Test that metric type constants are defined
	if CounterMetric != "counter" {
		t.Errorf("CounterMetric should be 'counter', got '%s'", CounterMetric)
	}

	if GaugeMetric != "gauge" {
		t.Errorf("GaugeMetric should be 'gauge', got '%s'", GaugeMetric)
	}

	if HistogramMetric != "histogram" {
		t.Errorf("HistogramMetric should be 'histogram', got '%s'", HistogramMetric)
	}

	if SummaryMetric != "summary" {
		t.Errorf("SummaryMetric should be 'summary', got '%s'", SummaryMetric)
	}
}

func TestCollectorStats(t *testing.T) {
	stats := CollectorStats{
		MetricsCollected: 100,
		BatchOperations:  10,
		Errors:           2,
		LastCollection:   1234567890,
	}

	if stats.MetricsCollected != 100 {
		t.Errorf("Expected MetricsCollected 100, got %d", stats.MetricsCollected)
	}

	if stats.BatchOperations != 10 {
		t.Errorf("Expected BatchOperations 10, got %d", stats.BatchOperations)
	}

	if stats.Errors != 2 {
		t.Errorf("Expected Errors 2, got %d", stats.Errors)
	}

	if stats.LastCollection != 1234567890 {
		t.Errorf("Expected LastCollection 1234567890, got %d", stats.LastCollection)
	}
}

func TestMetricError(t *testing.T) {
	err := &MetricError{
		Name:    "test_metric",
		Message: "test error",
	}

	expected := "metric error: test_metric: test error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}
