package metricsdb

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/metrics"
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

	// Collect metrics - should return empty since no collector is attached
	metrics, err := provider.Collect()
	assert.Nil(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0, len(metrics), "No collector attached - should return empty metrics")

	// Test with collector
	collector := NewDatabaseMetricsCollector()
	collector.RecordQuery("postgresql", "select", 100*time.Millisecond, nil)
	collector.RecordTransaction("postgresql", "begin", true)
	collector.UpdateConnectionStats("postgresql", "active", 5)

	providerWithCollector := NewDatabaseMetricProviderWithCollector(collector)
	metrics, err = providerWithCollector.Collect()
	assert.Nil(t, err)
	assert.NotNil(t, metrics)
	assert.True(t, len(metrics) > 0, "Should collect metrics with collector attached")
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

func TestCollectMetricsSkipsMalformedKeys(t *testing.T) {
	collector := NewDatabaseMetricsCollector()
	collector.queryCounters["invalid"] = 1
	collector.errorCounters["too_short"] = 2
	collector.queryDurations["bad"] = []time.Duration{time.Second}
	collector.transactionCounters["broken"] = 3
	collector.executionCounters["oops"] = 4
	collector.connectionStats["bad_state_extra"] = 5

	validKey := metrics.BuildKey("postgresql", "select", "success")
	collector.queryCounters[validKey] = 1
	collector.queryDurations[validKey] = []time.Duration{100 * time.Millisecond, 300 * time.Millisecond}

	metricsList, err := NewDatabaseMetricProviderWithCollector(collector).Collect()
	assert.Nil(t, err)

	foundDuration := false
	for _, metric := range metricsList {
		if metric.Name == "magicorm_database_query_duration_seconds" {
			foundDuration = true
			assert.InDelta(t, 0.2, metric.Value, 0.001)
			assert.Equal(t, "postgresql", metric.Labels["database"])
		}
	}
	assert.True(t, foundDuration)
}

func TestParseKey(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, parseKey(metrics.BuildKey("a", "b", "c")))
}

func TestRegisterDatabaseMetricsWithoutGlobalManager(t *testing.T) {
	oldCollector := databaseMetricCollector
	oldProvider := databaseMetricProvider
	defer func() {
		databaseMetricCollector = oldCollector
		databaseMetricProvider = oldProvider
	}()

	databaseMetricCollector = nil
	databaseMetricProvider = nil

	RegisterDatabaseMetrics()

	assert.NotNil(t, GetDatabaseMetricsCollector())
	assert.Nil(t, databaseMetricProvider)
}
