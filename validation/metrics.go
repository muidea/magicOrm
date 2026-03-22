package validation

import (
	"time"

	metricsvalidation "github.com/muidea/magicOrm/metrics/validation"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

func recordValidationMetric(operation OperationType, modelName string, scenario verrors.Scenario, duration time.Duration, err error) {
	collector := metricsvalidation.GetValidationMetricsCollector()
	if collector == nil {
		return
	}

	collector.RecordValidation(string(operation), modelName, string(scenario), duration, err)
}

func recordValidationCacheAccess(cacheType string, hit bool) {
	collector := metricsvalidation.GetValidationMetricsCollector()
	if collector == nil {
		return
	}

	collector.RecordCacheAccess(cacheType, hit)
}

func recordConstraintChecks(fieldName string, constraints []string, passed bool) {
	collector := metricsvalidation.GetValidationMetricsCollector()
	if collector == nil {
		return
	}

	for _, constraintType := range constraints {
		collector.RecordConstraintCheck(constraintType, fieldName, passed)
	}
}
