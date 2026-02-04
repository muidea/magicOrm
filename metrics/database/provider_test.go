package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDatabaseMetricProvider(t *testing.T) {
	provider := NewDatabaseMetricProvider()
	assert.NotNil(t, provider)
	assert.Equal(t, "magicorm_database", provider.Name())
}

func TestMetricsDefinitions(t *testing.T) {
	provider := NewDatabaseMetricProvider()

	metrics := provider.Metrics()
	assert.NotNil(t, metrics)
	assert.True(t, len(metrics) > 0)

	// Check for specific metric definitions
	foundQueryCounter := false
	foundDurationHistogram := false
	foundErrorCounter := false

	for _, metric := range metrics {
		switch metric.Name {
		case "magicorm_database_queries_total":
			foundQueryCounter = true
		case "magicorm_database_query_duration_seconds":
			foundDurationHistogram = true
		case "magicorm_database_errors_total":
			foundErrorCounter = true
		}
	}

	assert.True(t, foundQueryCounter, "Should have query counter metric")
	assert.True(t, foundDurationHistogram, "Should have duration histogram metric")
	assert.True(t, foundErrorCounter, "Should have error counter metric")
}

func TestCollectMetrics(t *testing.T) {
	provider := NewDatabaseMetricProvider()

	// Collect metrics - should return empty since MagicORM no longer collects data
	metrics, err := provider.Collect()
	assert.Nil(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, len(metrics), "MagicORM no longer collects data - only provides definitions")
}

func TestProviderLifecycle(t *testing.T) {
	provider := NewDatabaseMetricProvider()

	// Test initialization
	err := provider.Init(nil)
	assert.Nil(t, err)

	// Test shutdown
	err = provider.Shutdown()
	assert.Nil(t, err)
}
