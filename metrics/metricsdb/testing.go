package metricsdb

// SetDatabaseMetricsCollectorForTest replaces the global collector for tests.
func SetDatabaseMetricsCollectorForTest(collector *DatabaseMetricsCollector) {
	databaseMetricCollector = collector
	databaseMetricProvider = nil
}
