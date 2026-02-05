package validation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewValidationMetricProvider(t *testing.T) {
	provider := NewValidationMetricProvider()
	assert.NotNil(t, provider)
	assert.Equal(t, "magicorm_validation", provider.Name())
}

func TestMetricsDefinitions(t *testing.T) {
	provider := NewValidationMetricProvider()

	metrics := provider.Metrics()
	assert.NotNil(t, metrics)
	assert.True(t, len(metrics) > 0)

	// Check for specific metric definitions
	foundValidationCounter := false
	foundDurationHistogram := false
	foundErrorCounter := false

	for _, metric := range metrics {
		switch metric.Name {
		case "magicorm_validation_operations_total":
			foundValidationCounter = true
		case "magicorm_validation_duration_seconds":
			foundDurationHistogram = true
		case "magicorm_validation_errors_total":
			foundErrorCounter = true
		}
	}

	assert.True(t, foundValidationCounter, "Should have validation counter metric")
	assert.True(t, foundDurationHistogram, "Should have duration histogram metric")
	assert.True(t, foundErrorCounter, "Should have error counter metric")
}

func TestCollectMetrics(t *testing.T) {
	provider := NewValidationMetricProvider()

	// Collect metrics - should return empty since no collector is attached
	metrics, err := provider.Collect()
	assert.Nil(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, len(metrics), "No collector attached - should return empty metrics")

	// Test with collector
	collector := NewValidationMetricsCollector()
	collector.RecordValidation("validate", "User", "insert", 50*time.Millisecond, nil)
	collector.RecordCacheAccess("type", true)
	collector.RecordConstraintCheck("required", "Name", true)

	providerWithCollector := NewValidationMetricProviderWithCollector(collector)
	metrics, err = providerWithCollector.Collect()
	assert.Nil(t, err)
	assert.NotNil(t, metrics)
	assert.True(t, len(metrics) > 0, "Should collect metrics with collector attached")
}

func TestProviderLifecycle(t *testing.T) {
	provider := NewValidationMetricProvider()

	// Test initialization
	err := provider.Init(nil)
	assert.Nil(t, err)

	// Test shutdown
	err = provider.Shutdown()
	assert.Nil(t, err)
}
