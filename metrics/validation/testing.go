package validation

// SetValidationMetricsCollectorForTest replaces the global collector for tests.
func SetValidationMetricsCollectorForTest(collector *ValidationMetricsCollector) {
	validationMetricCollector = collector
	validationMetricProvider = nil
}
