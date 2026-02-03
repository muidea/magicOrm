package core

import (
	"testing"

	"github.com/muidea/magicCommon/monitoring/types"
)

func TestMetricTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		actual   MetricType
		expected string
	}{
		{"CounterMetric", CounterMetric, "counter"},
		{"GaugeMetric", GaugeMetric, "gauge"},
		{"HistogramMetric", HistogramMetric, "histogram"},
		{"SummaryMetric", SummaryMetric, "summary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.actual) != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestTypeAliases(t *testing.T) {
	// Test that type aliases are properly defined
	// These are compile-time checks, but we can verify they exist
	var _ Metric = types.Metric{}
	var _ MetricDefinition = types.MetricDefinition{}

	// Create test instances to verify the types work
	metric := Metric{
		Name:   "test_metric",
		Value:  1.0,
		Labels: map[string]string{"test": "label"},
		Type:   CounterMetric,
	}

	if metric.Name != "test_metric" {
		t.Errorf("Expected metric name 'test_metric', got '%s'", metric.Name)
	}

	def := MetricDefinition{
		Name:       "test_definition",
		Help:       "Test metric definition",
		Type:       CounterMetric,
		LabelNames: []string{"label1", "label2"},
	}

	if def.Name != "test_definition" {
		t.Errorf("Expected definition name 'test_definition', got '%s'", def.Name)
	}
}

func TestConfigCollectorStats(t *testing.T) {
	stats := CollectorStats{
		MetricsCollected: 150,
		BatchOperations:  25,
		Errors:           3,
		LastCollection:   987654321,
	}

	if stats.MetricsCollected != 150 {
		t.Errorf("Expected MetricsCollected 150, got %d", stats.MetricsCollected)
	}

	if stats.BatchOperations != 25 {
		t.Errorf("Expected BatchOperations 25, got %d", stats.BatchOperations)
	}

	if stats.Errors != 3 {
		t.Errorf("Expected Errors 3, got %d", stats.Errors)
	}

	if stats.LastCollection != 987654321 {
		t.Errorf("Expected LastCollection 987654321, got %d", stats.LastCollection)
	}
}

func TestConfigMetricError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "simple error",
			message:  "test error",
			expected: "metric error: test_metric: test error",
		},
		{
			name:     "empty message",
			message:  "",
			expected: "metric error: test_metric: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &MetricError{
				Name:    "test_metric",
				Message: tt.message,
			}

			if err.Error() != tt.expected {
				t.Errorf("Expected error message '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

func TestSimpleCollectorInterface(t *testing.T) {
	// Test that the interface is properly defined
	var collector SimpleCollector = &NoopCollector{}

	// Test all interface methods
	err := collector.Record("test", 1.0, nil)
	if err != nil {
		t.Errorf("Record should not return error: %v", err)
	}

	err = collector.Increment("test", nil)
	if err != nil {
		t.Errorf("Increment should not return error: %v", err)
	}

	err = collector.Decrement("test", nil)
	if err != nil {
		t.Errorf("Decrement should not return error: %v", err)
	}

	err = collector.Observe("test", 1.0, nil)
	if err != nil {
		t.Errorf("Observe should not return error: %v", err)
	}
}

func TestNoopCollectorImplementation(t *testing.T) {
	collector := &NoopCollector{}

	// Verify all methods return nil
	if err := collector.Record("metric", 1.0, map[string]string{"a": "b"}); err != nil {
		t.Errorf("Record returned error: %v", err)
	}

	if err := collector.Increment("counter", map[string]string{"a": "b"}); err != nil {
		t.Errorf("Increment returned error: %v", err)
	}

	if err := collector.Decrement("gauge", map[string]string{"a": "b"}); err != nil {
		t.Errorf("Decrement returned error: %v", err)
	}

	if err := collector.Observe("histogram", 2.5, map[string]string{"a": "b"}); err != nil {
		t.Errorf("Observe returned error: %v", err)
	}
}
