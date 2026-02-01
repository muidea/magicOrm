package validation

import (
	"time"

	"github.com/muidea/magicOrm/monitoring/core"
)

// Scenario represents the operation scenario (compatible with validation/errors.Scenario)
type Scenario string

const (
	ScenarioInsert Scenario = "insert"
	ScenarioUpdate Scenario = "update"
	ScenarioQuery  Scenario = "query"
	ScenarioDelete Scenario = "delete"
)

// ValidationMonitor provides monitoring for validation system
type ValidationMonitor struct {
	collector *core.Collector
	config    *core.MonitoringConfig
}

// NewValidationMonitor creates a new validation monitor
func NewValidationMonitor(collector *core.Collector, config *core.MonitoringConfig) *ValidationMonitor {
	if config == nil {
		defaultConfig := core.DefaultMonitoringConfig()
		config = &defaultConfig
	}

	monitor := &ValidationMonitor{
		collector: collector,
		config:    config,
	}

	// Register validation-specific metrics
	monitor.registerMetrics()

	return monitor
}

// RecordValidation records a validation operation
func (m *ValidationMonitor) RecordValidation(
	operation string,
	modelName string,
	scenario Scenario,
	duration time.Duration,
	err error,
	additionalLabels map[string]string,
) {
	if !m.config.IsValidationEnabled() {
		return
	}

	labels := m.buildLabels(modelName, scenario, additionalLabels)

	// Record operation
	status := "success"
	if err != nil {
		status = "error"
	}
	labels["status"] = status

	m.collector.RecordOperation(
		"validation_operation",
		time.Now().Add(-duration), // start time
		err == nil,
		labels,
	)

	// Record duration
	m.collector.RecordDuration(
		"validation_duration_seconds",
		duration,
		labels,
	)

	// Record error if any
	if err != nil {
		errorType := m.classifyError(err)
		errorLabels := make(map[string]string)
		for k, v := range labels {
			errorLabels[k] = v
		}
		errorLabels["error_type"] = errorType

		m.collector.RecordError(
			"validation",
			errorType,
			errorLabels,
		)
	}
}

// RecordCacheAccess records cache access for validation
func (m *ValidationMonitor) RecordCacheAccess(
	operation string,
	cacheType string,
	key string,
	hit bool,
	duration time.Duration,
	additionalLabels map[string]string,
) {
	if !m.config.IsValidationEnabled() || !m.config.IsCacheEnabled() {
		return
	}

	labels := m.buildCacheLabels(cacheType, additionalLabels)
	labels["operation"] = operation
	labels["hit"] = "false"
	if hit {
		labels["hit"] = "true"
	}

	// Record cache operation
	m.collector.Increment(
		"validation_cache_operations_total",
		labels,
	)

	// Record cache hit/miss
	if hit {
		m.collector.Increment(
			"validation_cache_hits_total",
			labels,
		)
	} else {
		m.collector.Increment(
			"validation_cache_misses_total",
			labels,
		)
	}

	// Record cache duration
	m.collector.RecordDuration(
		"validation_cache_duration_seconds",
		duration,
		labels,
	)
}

// RecordLayerPerformance records performance of a validation layer
func (m *ValidationMonitor) RecordLayerPerformance(
	layer string,
	operation string,
	duration time.Duration,
	success bool,
	additionalLabels map[string]string,
) {
	if !m.config.IsValidationEnabled() {
		return
	}

	labels := m.buildLayerLabels(layer, additionalLabels)
	labels["operation"] = operation
	labels["success"] = "false"
	if success {
		labels["success"] = "true"
	}

	// Record layer operation
	m.collector.Increment(
		"validation_layer_operations_total",
		labels,
	)

	// Record layer duration
	m.collector.RecordDuration(
		"validation_layer_duration_seconds",
		duration,
		labels,
	)

	// Record error if failed
	if !success {
		m.collector.RecordError(
			"validation_layer",
			layer,
			labels,
		)
	}
}

// RecordConstraintValidation records constraint validation
func (m *ValidationMonitor) RecordConstraintValidation(
	constraintType string,
	fieldName string,
	value interface{},
	duration time.Duration,
	success bool,
	additionalLabels map[string]string,
) {
	if !m.config.IsValidationEnabled() {
		return
	}

	labels := m.buildConstraintLabels(constraintType, fieldName, additionalLabels)
	labels["success"] = "false"
	if success {
		labels["success"] = "true"
	}

	// Record constraint validation
	m.collector.Increment(
		"validation_constraint_operations_total",
		labels,
	)

	// Record constraint duration
	m.collector.RecordDuration(
		"validation_constraint_duration_seconds",
		duration,
		labels,
	)

	// Record error if failed
	if !success {
		m.collector.RecordError(
			"validation_constraint",
			constraintType,
			labels,
		)
	}
}

// RecordTypeValidation records type validation
func (m *ValidationMonitor) RecordTypeValidation(
	typeName string,
	operation string,
	duration time.Duration,
	success bool,
	additionalLabels map[string]string,
) {
	if !m.config.IsValidationEnabled() {
		return
	}

	labels := m.buildTypeLabels(typeName, additionalLabels)
	labels["operation"] = operation
	labels["success"] = "false"
	if success {
		labels["success"] = "true"
	}

	// Record type validation
	m.collector.Increment(
		"validation_type_operations_total",
		labels,
	)

	// Record type duration
	m.collector.RecordDuration(
		"validation_type_duration_seconds",
		duration,
		labels,
	)

	// Record error if failed
	if !success {
		m.collector.RecordError(
			"validation_type",
			typeName,
			labels,
		)
	}
}

// RecordDatabaseValidation records database validation
func (m *ValidationMonitor) RecordDatabaseValidation(
	dbType string,
	operation string,
	duration time.Duration,
	success bool,
	additionalLabels map[string]string,
) {
	if !m.config.IsValidationEnabled() || !m.config.IsDatabaseEnabled() {
		return
	}

	labels := m.buildDatabaseLabels(dbType, additionalLabels)
	labels["operation"] = operation
	labels["success"] = "false"
	if success {
		labels["success"] = "true"
	}

	// Record database validation
	m.collector.Increment(
		"validation_database_operations_total",
		labels,
	)

	// Record database duration
	m.collector.RecordDuration(
		"validation_database_duration_seconds",
		duration,
		labels,
	)

	// Record error if failed
	if !success {
		m.collector.RecordError(
			"validation_database",
			dbType,
			labels,
		)
	}
}

// RecordScenarioAdaptation records scenario adaptation
func (m *ValidationMonitor) RecordScenarioAdaptation(
	scenario Scenario,
	operation string,
	duration time.Duration,
	success bool,
	additionalLabels map[string]string,
) {
	if !m.config.IsValidationEnabled() {
		return
	}

	labels := m.buildScenarioLabels(scenario, additionalLabels)
	labels["operation"] = operation
	labels["success"] = "false"
	if success {
		labels["success"] = "true"
	}

	// Record scenario adaptation
	m.collector.Increment(
		"validation_scenario_operations_total",
		labels,
	)

	// Record scenario duration
	m.collector.RecordDuration(
		"validation_scenario_duration_seconds",
		duration,
		labels,
	)

	// Record error if failed
	if !success {
		m.collector.RecordError(
			"validation_scenario",
			string(scenario),
			labels,
		)
	}
}

// GetStats returns validation monitoring statistics
func (m *ValidationMonitor) GetStats() map[string]interface{} {
	// Get metrics for validation
	metrics, err := m.collector.GetMetric("validation_operation_total")
	if err != nil {
		metrics = []core.Metric{}
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"total_operations": len(metrics),
		"enabled":          m.config.IsValidationEnabled(),
	}

	return stats
}

// Private methods

func (m *ValidationMonitor) registerMetrics() {
	// Validation operation metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_operation_total",
		Type:       core.CounterMetric,
		Help:       "Total number of validation operations",
		LabelNames: []string{"model", "scenario", "status"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Validation operation duration in seconds",
		LabelNames: []string{"model", "scenario", "status"},
		Buckets:    []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	})

	// Validation error metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_errors_total",
		Type:       core.CounterMetric,
		Help:       "Total number of validation errors",
		LabelNames: []string{"model", "scenario", "error_type"},
	})

	// Cache metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_cache_operations_total",
		Type:       core.CounterMetric,
		Help:       "Total number of cache operations",
		LabelNames: []string{"cache_type", "operation", "hit"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_cache_hits_total",
		Type:       core.CounterMetric,
		Help:       "Total number of cache hits",
		LabelNames: []string{"cache_type", "operation"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_cache_misses_total",
		Type:       core.CounterMetric,
		Help:       "Total number of cache misses",
		LabelNames: []string{"cache_type", "operation"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_cache_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Cache operation duration in seconds",
		LabelNames: []string{"cache_type", "operation", "hit"},
		Buckets:    []float64{.0001, .0005, .001, .005, .01, .025, .05},
	})

	// Layer metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_layer_operations_total",
		Type:       core.CounterMetric,
		Help:       "Total number of layer operations",
		LabelNames: []string{"layer", "operation", "success"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_layer_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Layer operation duration in seconds",
		LabelNames: []string{"layer", "operation", "success"},
		Buckets:    []float64{.0001, .0005, .001, .005, .01, .025, .05},
	})

	// Constraint metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_constraint_operations_total",
		Type:       core.CounterMetric,
		Help:       "Total number of constraint validations",
		LabelNames: []string{"constraint_type", "field", "success"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_constraint_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Constraint validation duration in seconds",
		LabelNames: []string{"constraint_type", "field", "success"},
		Buckets:    []float64{.0001, .0005, .001, .005, .01, .025, .05},
	})

	// Type metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_type_operations_total",
		Type:       core.CounterMetric,
		Help:       "Total number of type validations",
		LabelNames: []string{"type", "operation", "success"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_type_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Type validation duration in seconds",
		LabelNames: []string{"type", "operation", "success"},
		Buckets:    []float64{.0001, .0005, .001, .005, .01, .025, .05},
	})

	// Database metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_database_operations_total",
		Type:       core.CounterMetric,
		Help:       "Total number of database validations",
		LabelNames: []string{"db_type", "operation", "success"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_database_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Database validation duration in seconds",
		LabelNames: []string{"db_type", "operation", "success"},
		Buckets:    []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	})

	// Scenario metrics
	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_scenario_operations_total",
		Type:       core.CounterMetric,
		Help:       "Total number of scenario adaptations",
		LabelNames: []string{"scenario", "operation", "success"},
	})

	m.collector.RegisterDefinition(core.MetricDefinition{
		Name:       "validation_scenario_duration_seconds",
		Type:       core.HistogramMetric,
		Help:       "Scenario adaptation duration in seconds",
		LabelNames: []string{"scenario", "operation", "success"},
		Buckets:    []float64{.0001, .0005, .001, .005, .01, .025, .05},
	})
}

func (m *ValidationMonitor) buildLabels(modelName string, scenario Scenario, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if modelName != "" {
		labels["model"] = modelName
	}
	if scenario != "" {
		labels["scenario"] = string(scenario)
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ValidationMonitor) buildCacheLabels(cacheType string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if cacheType != "" {
		labels["cache_type"] = cacheType
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ValidationMonitor) buildLayerLabels(layer string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if layer != "" {
		labels["layer"] = layer
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ValidationMonitor) buildConstraintLabels(constraintType, fieldName string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if constraintType != "" {
		labels["constraint_type"] = constraintType
	}
	if fieldName != "" {
		labels["field"] = fieldName
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ValidationMonitor) buildTypeLabels(typeName string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if typeName != "" {
		labels["type"] = typeName
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ValidationMonitor) buildDatabaseLabels(dbType string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if dbType != "" {
		labels["db_type"] = dbType
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ValidationMonitor) buildScenarioLabels(scenario Scenario, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if scenario != "" {
		labels["scenario"] = string(scenario)
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

func (m *ValidationMonitor) classifyError(err error) string {
	// Classify error based on error message or type
	// This is a simplified implementation
	errStr := err.Error()

	switch {
	case contains(errStr, "type"):
		return "type_validation"
	case contains(errStr, "constraint"):
		return "constraint_validation"
	case contains(errStr, "database"):
		return "database_validation"
	case contains(errStr, "scenario"):
		return "scenario_adaptation"
	case contains(errStr, "cache"):
		return "cache_error"
	default:
		return "unknown"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}

// Convenience functions

// DefaultValidationMonitor creates a validation monitor with default configuration
func DefaultValidationMonitor() *ValidationMonitor {
	config := core.DefaultMonitoringConfig()
	collector := core.NewCollector(&config)
	return NewValidationMonitor(collector, &config)
}

// ValidationMonitorWithConfig creates a validation monitor with custom configuration
func ValidationMonitorWithConfig(config *core.MonitoringConfig) *ValidationMonitor {
	collector := core.NewCollector(config)
	return NewValidationMonitor(collector, config)
}
