package validation

import (
	"testing"

	"github.com/muidea/magicOrm/metrics"
	metricsvalidation "github.com/muidea/magicOrm/metrics/validation"
	"github.com/muidea/magicOrm/models"
	validationcache "github.com/muidea/magicOrm/validation/cache"
	verrors "github.com/muidea/magicOrm/validation/errors"
	"github.com/stretchr/testify/assert"
)

func TestValidationMetricsRecordedForManagerAndCache(t *testing.T) {
	oldCollector := metricsvalidation.GetValidationMetricsCollector()
	collector := metricsvalidation.NewValidationMetricsCollector()
	metricsvalidation.SetValidationMetricsCollectorForTest(collector)
	defer metricsvalidation.SetValidationMetricsCollectorForTest(oldCollector)

	manager := NewValidationFactory().CreateValidationManager(DefaultConfig())
	model := &testModel{
		name: "User",
		fields: []models.Field{
			&testField{
				name: "Name",
				typ:  &testType{name: "Name", value: models.TypeStringValue},
				spec: &testSpec{
					constraints: testConstraints{
						directives: []models.Directive{
							testDirective{key: models.KeyRequired},
						},
					},
				},
				value: &testValue{value: "alice", valid: true},
			},
		},
	}

	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, nil, "postgresql")
	err := manager.ValidateModel(model, ctx)
	assert.NoError(t, err)
	err = manager.ValidateModel(model, ctx)
	assert.NoError(t, err)

	assert.Equal(t, int64(2), collector.GetValidationCounters()[metrics.BuildKey("create", "User", "insert", "success")])
	assert.Equal(t, int64(1), collector.GetCacheAccessCounters()[metrics.BuildKey("constraint", "miss")])
	assert.Equal(t, int64(1), collector.GetCacheAccessCounters()[metrics.BuildKey("constraint", "hit")])
	assert.Equal(t, int64(2), collector.GetConstraintCheckCounters()[metrics.BuildKey(string(models.KeyRequired), "Name", "passed")])
	assert.Equal(t, 0.5, collector.GetCacheHitRatio("constraint"))
}

func TestValidationMetricsRecordErrors(t *testing.T) {
	oldCollector := metricsvalidation.GetValidationMetricsCollector()
	collector := metricsvalidation.NewValidationMetricsCollector()
	metricsvalidation.SetValidationMetricsCollectorForTest(collector)
	defer metricsvalidation.SetValidationMetricsCollectorForTest(oldCollector)

	manager := NewValidationFactory().CreateValidationManager(DefaultConfig())
	model := &testModel{
		name: "User",
		fields: []models.Field{
			&testField{
				name: "Name",
				typ:  &testType{name: "Name", value: models.TypeStringValue},
				spec: &testSpec{
					constraints: testConstraints{
						directives: []models.Directive{
							testDirective{key: models.KeyRequired},
						},
					},
				},
				value: &testValue{value: nil, valid: true},
			},
		},
	}

	ctx := NewContext(verrors.ScenarioInsert, OperationCreate, nil, "postgresql")
	err := manager.ValidateModel(model, ctx)
	assert.Error(t, err)

	assert.Equal(t, int64(1), collector.GetValidationCounters()[metrics.BuildKey("create", "User", "insert", "error")])
	var errorCount int64
	for key, count := range collector.GetErrorCounters() {
		parts := metrics.ParseKey(key)
		if len(parts) == 4 && parts[0] == "create" && parts[1] == "User" && parts[2] == "insert" {
			errorCount += count
		}
	}
	assert.Equal(t, int64(1), errorCount)
	assert.Equal(t, int64(1), collector.GetConstraintCheckCounters()[metrics.BuildKey(string(models.KeyRequired), "Name", "failed")])
}

func TestValidationMetricsRecordModelCacheAccess(t *testing.T) {
	oldCollector := metricsvalidation.GetValidationMetricsCollector()
	collector := metricsvalidation.NewValidationMetricsCollector()
	metricsvalidation.SetValidationMetricsCollectorForTest(collector)
	defer metricsvalidation.SetValidationMetricsCollectorForTest(oldCollector)

	cache := validationcache.NewValidationCache(validationcache.DefaultCacheConfig())
	model := &testModel{name: "User"}

	_, ok := cache.GetModelResult(model, verrors.ScenarioInsert)
	assert.False(t, ok)

	cache.SetModelResult(model, verrors.ScenarioInsert, nil)
	_, ok = cache.GetModelResult(model, verrors.ScenarioInsert)
	assert.True(t, ok)

	assert.Equal(t, int64(1), collector.GetCacheAccessCounters()[metrics.BuildKey("model", "miss")])
	assert.Equal(t, int64(1), collector.GetCacheAccessCounters()[metrics.BuildKey("model", "hit")])
	assert.Equal(t, 0.5, collector.GetCacheHitRatio("model"))
}

func TestValidationMetricHelpers(t *testing.T) {
	oldCollector := metricsvalidation.GetValidationMetricsCollector()
	collector := metricsvalidation.NewValidationMetricsCollector()
	metricsvalidation.SetValidationMetricsCollectorForTest(collector)
	defer metricsvalidation.SetValidationMetricsCollectorForTest(oldCollector)

	recordValidationMetric(OperationUpdate, "User", verrors.ScenarioUpdate, 0, nil)
	recordValidationCacheAccess("custom", true)
	recordConstraintChecks("Name", []string{string(models.KeyRequired), string(models.KeyMin)}, false)

	assert.Equal(t, int64(1), collector.GetValidationCounters()[metrics.BuildKey("update", "User", "update", "success")])
	assert.Equal(t, int64(1), collector.GetCacheAccessCounters()[metrics.BuildKey("custom", "hit")])
	assert.Equal(t, int64(1), collector.GetConstraintCheckCounters()[metrics.BuildKey(string(models.KeyRequired), "Name", "failed")])
	assert.Equal(t, int64(1), collector.GetConstraintCheckCounters()[metrics.BuildKey(string(models.KeyMin), "Name", "failed")])
}

func TestValidationMetricHelpersWithoutCollector(t *testing.T) {
	oldCollector := metricsvalidation.GetValidationMetricsCollector()
	metricsvalidation.SetValidationMetricsCollectorForTest(nil)
	defer metricsvalidation.SetValidationMetricsCollectorForTest(oldCollector)

	recordValidationMetric(OperationDelete, "User", verrors.ScenarioDelete, 0, nil)
	recordValidationCacheAccess("custom", false)
	recordConstraintChecks("Name", []string{string(models.KeyRequired)}, true)
}
