package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewORMMetricProvider(t *testing.T) {
	provider := NewORMMetricProvider()
	assert.NotNil(t, provider)
	assert.Equal(t, "magicorm_orm", provider.Name())
}

func TestMetricsDefinitions(t *testing.T) {
	provider := NewORMMetricProvider()

	metrics := provider.Metrics()
	assert.NotNil(t, metrics)
	assert.True(t, len(metrics) > 0)

	// Check for specific metric definitions
	foundOperationCounter := false
	foundDurationHistogram := false
	foundErrorCounter := false

	for _, metric := range metrics {
		switch metric.Name {
		case "magicorm_orm_operations_total":
			foundOperationCounter = true
		case "magicorm_orm_operation_duration_seconds":
			foundDurationHistogram = true
		case "magicorm_orm_errors_total":
			foundErrorCounter = true
		}
	}

	assert.True(t, foundOperationCounter, "Should have operation counter metric")
	assert.True(t, foundDurationHistogram, "Should have duration histogram metric")
	assert.True(t, foundErrorCounter, "Should have error counter metric")
}

func TestCollectMetrics(t *testing.T) {
	provider := NewORMMetricProvider()

	// Collect metrics - should return empty since MagicORM no longer collects data
	metrics, err := provider.Collect()
	assert.Nil(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, len(metrics), "MagicORM no longer collects data - only provides definitions")
}

func TestProviderLifecycle(t *testing.T) {
	provider := NewORMMetricProvider()

	// Test initialization
	err := provider.Init(nil)
	assert.Nil(t, err)

	// Test shutdown
	err = provider.Shutdown()
	assert.Nil(t, err)
}
