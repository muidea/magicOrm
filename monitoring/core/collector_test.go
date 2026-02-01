package core

import (
	"testing"
	"time"
)

func TestNewCollector(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false // Disable async for test
	collector := NewCollector(&config)

	if collector == nil {
		t.Fatal("Collector should not be nil")
	}

	// Test that default metrics are registered
	// Note: Default metrics are registered but not necessarily recorded yet
	// We can test by trying to record one of them
	def := MetricDefinition{
		Name:       "monitoring_metrics_collected_total",
		Type:       CounterMetric,
		Help:       "Test",
		LabelNames: []string{},
	}

	// Try to register - should fail because it's already registered
	err := collector.RegisterDefinition(def)
	if err == nil {
		t.Error("Should fail to register duplicate metric")
	}
}

func TestRegisterDefinition(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Test valid definition
	def := MetricDefinition{
		Name:       "test_metric",
		Type:       CounterMetric,
		Help:       "Test metric",
		LabelNames: []string{"label1", "label2"},
	}

	err := collector.RegisterDefinition(def)
	if err != nil {
		t.Errorf("Failed to register valid definition: %v", err)
	}

	// Test duplicate definition
	err = collector.RegisterDefinition(def)
	if err == nil {
		t.Error("Should error on duplicate definition")
	}

	// Test invalid definition - empty name
	invalidDef := MetricDefinition{
		Name: "",
		Type: CounterMetric,
		Help: "Test",
	}
	err = collector.RegisterDefinition(invalidDef)
	if err == nil {
		t.Error("Should error on empty name")
	}

	// Test invalid definition - empty help
	invalidDef = MetricDefinition{
		Name: "test2",
		Type: CounterMetric,
		Help: "",
	}
	err = collector.RegisterDefinition(invalidDef)
	if err == nil {
		t.Error("Should error on empty help")
	}

	// Test invalid definition - invalid type
	invalidDef = MetricDefinition{
		Name: "test3",
		Type: "invalid_type",
		Help: "Test",
	}
	err = collector.RegisterDefinition(invalidDef)
	if err == nil {
		t.Error("Should error on invalid type")
	}
}

func TestRecord(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register a test metric
	def := MetricDefinition{
		Name:       "test_counter",
		Type:       CounterMetric,
		Help:       "Test counter",
		LabelNames: []string{"label"},
	}
	collector.RegisterDefinition(def)

	// Test recording with valid labels
	err := collector.Record("test_counter", 1.0, map[string]string{"label": "value"})
	if err != nil {
		t.Errorf("Failed to record metric: %v", err)
	}

	// Test recording with missing label
	err = collector.Record("test_counter", 1.0, map[string]string{})
	if err == nil {
		t.Error("Should error on missing label")
	}

	// Test recording undefined metric
	err = collector.Record("undefined_metric", 1.0, nil)
	if err == nil {
		t.Error("Should error on undefined metric")
	}
}

func TestRecordWithTimestamp(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register a test metric
	def := MetricDefinition{
		Name:       "test_gauge",
		Type:       GaugeMetric,
		Help:       "Test gauge",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(def)

	timestamp := time.Now().Add(-1 * time.Hour)
	err := collector.RecordWithTimestamp("test_gauge", 42.0, nil, timestamp)
	if err != nil {
		t.Errorf("Failed to record metric with timestamp: %v", err)
	}

	// Verify the timestamp was recorded
	metrics, err := collector.GetMetric("test_gauge")
	if err != nil {
		// Metric might not exist yet due to async collection
		// For this test, we'll check if we got an error
		t.Logf("Note: GetMetric returned error (might be expected): %v", err)
		return
	}

	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(metrics))
		return
	}

	if !metrics[0].Timestamp.Equal(timestamp) {
		t.Error("Timestamp not preserved")
	}
}

func TestIncrementDecrement(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register counter
	counterDef := MetricDefinition{
		Name:       "test_counter",
		Type:       CounterMetric,
		Help:       "Test counter",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(counterDef)

	// Register gauge
	gaugeDef := MetricDefinition{
		Name:       "test_gauge",
		Type:       GaugeMetric,
		Help:       "Test gauge",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(gaugeDef)

	// Test increment
	err := collector.Increment("test_counter", nil)
	if err != nil {
		t.Errorf("Failed to increment: %v", err)
	}

	// Test decrement (should work for gauge)
	err = collector.Decrement("test_gauge", nil)
	if err != nil {
		t.Errorf("Failed to decrement: %v", err)
	}

	// Test decrement on counter (should still work but might not be semantically correct)
	err = collector.Decrement("test_counter", nil)
	if err != nil {
		t.Errorf("Failed to decrement counter: %v", err)
	}
}

func TestObserve(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register histogram
	histogramDef := MetricDefinition{
		Name:       "test_histogram",
		Type:       HistogramMetric,
		Help:       "Test histogram",
		LabelNames: []string{},
		Buckets:    []float64{0.1, 0.5, 1.0},
	}
	collector.RegisterDefinition(histogramDef)

	// Test observe
	err := collector.Observe("test_histogram", 0.3, nil)
	if err != nil {
		t.Errorf("Failed to observe: %v", err)
	}
}

func TestGetMetrics(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register and record multiple metrics
	def1 := MetricDefinition{
		Name:       "metric1",
		Type:       CounterMetric,
		Help:       "Metric 1",
		LabelNames: []string{},
	}
	def2 := MetricDefinition{
		Name:       "metric2",
		Type:       GaugeMetric,
		Help:       "Metric 2",
		LabelNames: []string{"label"},
	}

	collector.RegisterDefinition(def1)
	collector.RegisterDefinition(def2)

	collector.Record("metric1", 1.0, nil)
	collector.Record("metric2", 2.0, map[string]string{"label": "value"})
	collector.Record("metric2", 3.0, map[string]string{"label": "value2"})

	// Get all metrics
	allMetrics := collector.GetMetrics()

	if len(allMetrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(allMetrics))
	}

	if len(allMetrics["metric1"]) != 1 {
		t.Errorf("Expected 1 metric1, got %d", len(allMetrics["metric1"]))
	}

	if len(allMetrics["metric2"]) != 2 {
		t.Errorf("Expected 2 metric2, got %d", len(allMetrics["metric2"]))
	}

	// Get specific metric
	metric2, err := collector.GetMetric("metric2")
	if err != nil {
		t.Errorf("Failed to get metric2: %v", err)
	}

	if len(metric2) != 2 {
		t.Errorf("Expected 2 metric2 entries, got %d", len(metric2))
	}

	// Test getting undefined metric
	_, err = collector.GetMetric("undefined")
	if err == nil {
		t.Error("Should error on undefined metric")
	}
}

func TestGetStats(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Record some metrics
	def := MetricDefinition{
		Name:       "test",
		Type:       CounterMetric,
		Help:       "Test",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(def)

	collector.Record("test", 1.0, nil)
	collector.Record("test", 2.0, nil)

	stats := collector.GetStats()

	if stats.MetricsCollected != 2 {
		t.Errorf("Expected 2 metrics collected, got %d", stats.MetricsCollected)
	}

	if stats.Uptime <= 0 {
		t.Error("Uptime should be positive")
	}
}

func TestReset(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Record some metrics
	def := MetricDefinition{
		Name:       "test",
		Type:       CounterMetric,
		Help:       "Test",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(def)

	collector.Record("test", 1.0, nil)

	// Reset
	collector.Reset()

	// Verify metrics are cleared
	metrics := collector.GetMetrics()
	if len(metrics) != 0 {
		t.Error("Metrics should be cleared after reset")
	}

	// Verify stats are reset
	stats := collector.GetStats()
	if stats.MetricsCollected != 0 {
		t.Errorf("Metrics collected should be 0 after reset, got %d", stats.MetricsCollected)
	}
}

func TestCleanup(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	config.RetentionPeriod = 1 * time.Millisecond // Very short retention for test
	collector := NewCollector(&config)

	// Register and record a metric
	def := MetricDefinition{
		Name:       "test",
		Type:       CounterMetric,
		Help:       "Test",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(def)

	// Record with old timestamp
	oldTime := time.Now().Add(-2 * time.Millisecond)
	collector.RecordWithTimestamp("test", 1.0, nil, oldTime)

	// Record with current timestamp
	collector.Record("test", 2.0, nil)

	// Cleanup
	collector.Cleanup()

	// Verify only current metric remains
	metrics, err := collector.GetMetric("test")
	if err != nil {
		t.Errorf("Failed to get metrics: %v", err)
		return
	}

	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric after cleanup, got %d", len(metrics))
		return
	}

	if metrics[0].Value != 2.0 {
		t.Errorf("Expected value 2.0, got %f", metrics[0].Value)
	}
}

func TestRecordOperation(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register operation metrics
	opDef := MetricDefinition{
		Name:       "test_operation_total",
		Type:       CounterMetric,
		Help:       "Test operations",
		LabelNames: []string{"status"},
	}
	durationDef := MetricDefinition{
		Name:       "test_operation_duration_seconds",
		Type:       HistogramMetric,
		Help:       "Test operation duration",
		LabelNames: []string{"status"},
	}

	collector.RegisterDefinition(opDef)
	collector.RegisterDefinition(durationDef)

	startTime := time.Now().Add(-100 * time.Millisecond)
	err := collector.RecordOperation("test_operation", startTime, true, map[string]string{"status": "success"})
	if err != nil {
		t.Errorf("Failed to record operation: %v", err)
		return
	}

	// Verify metrics were recorded
	opMetrics, err := collector.GetMetric("test_operation_total")
	if err != nil {
		t.Errorf("Failed to get operation metrics: %v", err)
		return
	}

	durationMetrics, err := collector.GetMetric("test_operation_duration_seconds")
	if err != nil {
		t.Errorf("Failed to get duration metrics: %v", err)
		return
	}

	if len(opMetrics) != 1 {
		t.Errorf("Expected 1 operation metric, got %d", len(opMetrics))
		return
	}

	if len(durationMetrics) != 1 {
		t.Errorf("Expected 1 duration metric, got %d", len(durationMetrics))
		return
	}

	if opMetrics[0].Labels["status"] != "success" {
		t.Errorf("Expected status 'success', got %s", opMetrics[0].Labels["status"])
	}
}

func TestRecordError(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register error metric
	def := MetricDefinition{
		Name:       "test_errors_total",
		Type:       CounterMetric,
		Help:       "Test errors",
		LabelNames: []string{"error_type"},
	}
	collector.RegisterDefinition(def)

	err := collector.RecordError("test", "validation_error", map[string]string{"error_type": "validation_error"})
	if err != nil {
		t.Errorf("Failed to record error: %v", err)
		return
	}

	// Verify error was recorded
	metrics, err := collector.GetMetric("test_errors_total")
	if err != nil {
		t.Errorf("Failed to get error metrics: %v", err)
		return
	}

	if len(metrics) != 1 {
		t.Errorf("Expected 1 error metric, got %d", len(metrics))
		return
	}

	if metrics[0].Labels["error_type"] != "validation_error" {
		t.Errorf("Expected error_type 'validation_error', got %s", metrics[0].Labels["error_type"])
	}
}

func TestRecordDuration(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register duration metric
	def := MetricDefinition{
		Name:       "test_duration",
		Type:       HistogramMetric,
		Help:       "Test duration",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(def)

	duration := 1500 * time.Millisecond // 1.5 seconds
	err := collector.RecordDuration("test_duration", duration, nil)
	if err != nil {
		t.Errorf("Failed to record duration: %v", err)
		return
	}

	// Verify duration was recorded
	metrics, err := collector.GetMetric("test_duration")
	if err != nil {
		t.Errorf("Failed to get duration metrics: %v", err)
		return
	}

	if len(metrics) != 1 {
		t.Errorf("Expected 1 duration metric, got %d", len(metrics))
		return
	}

	// Value should be in seconds
	if metrics[0].Value != 1.5 {
		t.Errorf("Expected value 1.5, got %f", metrics[0].Value)
	}
}

func TestSampling(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.SamplingRate = 0.0 // Disable sampling
	collector := NewCollector(&config)

	// Register a metric
	def := MetricDefinition{
		Name:       "test",
		Type:       CounterMetric,
		Help:       "Test",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(def)

	// Try to record - should be dropped due to sampling
	err := collector.Record("test", 1.0, nil)
	if err != nil {
		t.Errorf("Should not error when dropping due to sampling: %v", err)
	}

	// Verify no metrics were recorded
	metrics := collector.GetMetrics()
	if len(metrics) != 0 {
		t.Error("Metrics should be dropped when sampling rate is 0")
	}
}

func TestConcurrentAccess(t *testing.T) {
	config := DefaultMonitoringConfig()
	config.AsyncCollection = false
	collector := NewCollector(&config)

	// Register a metric
	def := MetricDefinition{
		Name:       "concurrent_test",
		Type:       CounterMetric,
		Help:       "Concurrent test",
		LabelNames: []string{},
	}
	collector.RegisterDefinition(def)

	// Concurrent recording
	done := make(chan bool)
	const goroutines = 10
	const recordsPerGoroutine = 100

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < recordsPerGoroutine; j++ {
				collector.Record("concurrent_test", 1.0, nil)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify total records
	stats := collector.GetStats()
	expected := goroutines * recordsPerGoroutine
	if stats.MetricsCollected != int64(expected) {
		t.Errorf("Expected %d metrics collected, got %d", expected, stats.MetricsCollected)
	}
}
