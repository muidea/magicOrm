package validation

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/metrics"
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

	for _, metric := range metrics {
		assert.Equal(t, "1.0.0", metric.ConstLabels["version"])
		assert.Equal(t, "validation", metric.ConstLabels["component"])
	}
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
	collector.RecordCacheAccess("constraint", false)
	collector.RecordConstraintCheck("required", "Name", true)

	providerWithCollector := NewValidationMetricProviderWithCollector(collector)
	metrics, err = providerWithCollector.Collect()
	assert.Nil(t, err)
	assert.NotNil(t, metrics)
	assert.True(t, len(metrics) > 0, "Should collect metrics with collector attached")

	cacheTypes := make([]string, 0)
	foundConstraintCounter := false
	for _, metric := range metrics {
		if metric.Name == "magicorm_validation_cache_hit_ratio" {
			if cacheType, ok := metric.Labels["cache_type"]; ok {
				cacheTypes = append(cacheTypes, cacheType)
			}
		}
		if metric.Name == "magicorm_validation_constraint_checks_total" &&
			metric.Labels["constraint_type"] == "required" &&
			metric.Labels["field"] == "Name" &&
			metric.Labels["status"] == "passed" {
			foundConstraintCounter = true
		}
	}
	assert.ElementsMatch(t, []string{"type", "constraint"}, cacheTypes)
	assert.True(t, foundConstraintCounter, "Should export constraint check metrics")
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

func TestCollectMetricsSkipsMalformedKeys(t *testing.T) {
	collector := NewValidationMetricsCollector()
	collector.validationCounters["invalid"] = 1
	collector.errorCounters["too_short"] = 2
	collector.validationDurations["bad"] = []time.Duration{time.Second}
	collector.cacheAccessCounters["broken"] = 3
	collector.constraintCheckCounters["oops"] = 4

	validDurationKey := metrics.BuildKey("validate", "User", "insert", "success")
	collector.validationCounters[validDurationKey] = 1
	collector.validationDurations[validDurationKey] = []time.Duration{100 * time.Millisecond, 300 * time.Millisecond}
	collector.cacheAccessCounters[metrics.BuildKey("type", "hit")] = 1

	metricsList, err := NewValidationMetricProviderWithCollector(collector).Collect()
	assert.Nil(t, err)

	foundDuration := false
	for _, metric := range metricsList {
		if metric.Name == "magicorm_validation_duration_seconds" {
			foundDuration = true
			assert.InDelta(t, 0.2, metric.Value, 0.001)
			assert.Equal(t, "User", metric.Labels["model"])
		}
	}
	assert.True(t, foundDuration)
}

func TestParseKey(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, parseKey(metrics.BuildKey("a", "b", "c")))
}

func TestRegisterValidationMetricsWithoutGlobalManager(t *testing.T) {
	oldCollector := validationMetricCollector
	oldProvider := validationMetricProvider
	defer func() {
		validationMetricCollector = oldCollector
		validationMetricProvider = oldProvider
	}()

	validationMetricCollector = nil
	validationMetricProvider = nil

	RegisterValidationMetrics()

	assert.NotNil(t, GetValidationMetricsCollector())
	assert.Nil(t, validationMetricProvider)
}

func TestEnsureValidationMetricProviderRegisteredWithoutCollector(t *testing.T) {
	oldCollector := validationMetricCollector
	oldProvider := validationMetricProvider
	defer func() {
		validationMetricCollector = oldCollector
		validationMetricProvider = oldProvider
	}()

	validationMetricCollector = nil
	validationMetricProvider = nil

	EnsureValidationMetricProviderRegistered()

	assert.Nil(t, validationMetricProvider)
}
