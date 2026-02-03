// Package validation provides validation monitoring adapters.
// This is a simplified version that only provides monitoring collection,
// relying on external monitoring systems for management and export.
package validation

import (
	"time"
)

// ValidationMonitor is a simplified monitor for validation operations.
type ValidationMonitor struct {
	collector ValidationCollector
	enabled   bool
}

// NewValidationMonitor creates a new validation monitor.
func NewValidationMonitor(collector ValidationCollector) *ValidationMonitor {
	return &ValidationMonitor{
		collector: collector,
		enabled:   collector != nil,
	}
}

// RecordValidation records a validation operation.
func (m *ValidationMonitor) RecordValidation(
	operation string,
	modelName string,
	scenario string,
	duration time.Duration,
	err error,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	// Calculate start time from duration
	startTime := time.Now().Add(-duration)

	// Merge labels
	labels := make(map[string]string)
	if additionalLabels != nil {
		for k, v := range additionalLabels {
			labels[k] = v
		}
	}

	m.collector.RecordValidation(operation, modelName, scenario, startTime, err, labels)
}

// RecordCacheAccess records cache access for validation.
func (m *ValidationMonitor) RecordCacheAccess(
	operation string,
	cacheType string,
	hit bool,
	duration time.Duration,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	// Merge labels
	labels := make(map[string]string)
	if additionalLabels != nil {
		for k, v := range additionalLabels {
			labels[k] = v
		}
	}

	m.collector.RecordCacheAccess(cacheType, operation, hit, duration, labels)
}

// RecordConstraintCheck records a constraint check.
func (m *ValidationMonitor) RecordConstraintCheck(
	constraintType string,
	fieldName string,
	passed bool,
	duration time.Duration,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	// Merge labels
	labels := make(map[string]string)
	if additionalLabels != nil {
		for k, v := range additionalLabels {
			labels[k] = v
		}
	}

	// Use the collector's RecordConstraintCheck method
	if collector, ok := m.collector.(interface {
		RecordConstraintCheck(constraintType string, fieldName string, passed bool, duration time.Duration, labels map[string]string)
	}); ok {
		collector.RecordConstraintCheck(constraintType, fieldName, passed, duration, labels)
	}
}

// RecordLayerPerformance records validation layer performance.
func (m *ValidationMonitor) RecordLayerPerformance(
	layer string,
	duration time.Duration,
	additionalLabels map[string]string,
) {
	if !m.enabled {
		return
	}

	// This is a simplified implementation
	// In real code, this would record layer-specific metrics
	labels := make(map[string]string)
	if additionalLabels != nil {
		for k, v := range additionalLabels {
			labels[k] = v
		}
	}
	labels["layer"] = layer

	// Record as a validation operation
	m.collector.RecordValidation("layer_performance", "", layer, time.Now().Add(-duration), nil, labels)
}

// classifyError classifies validation errors.
func (m *ValidationMonitor) classifyError(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()

	// Simple error classification
	switch {
	case contains(errStr, "type"):
		return "type_error"
	case contains(errStr, "constraint"):
		return "constraint_error"
	case contains(errStr, "required"):
		return "required_error"
	case contains(errStr, "range"):
		return "range_error"
	case contains(errStr, "format"):
		return "format_error"
	default:
		return "validation_error"
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}

// buildLabels builds labels for validation metrics.
func (m *ValidationMonitor) buildLabels(modelName string, scenario string, additional map[string]string) map[string]string {
	labels := make(map[string]string)

	// Base labels
	if modelName != "" {
		labels["model"] = modelName
	}
	if scenario != "" {
		labels["scenario"] = scenario
	}

	// Additional labels
	for k, v := range additional {
		labels[k] = v
	}

	return labels
}

// SimpleValidationMonitor is a wrapper for backward compatibility.
type SimpleValidationMonitor struct {
	*ValidationMonitor
}

// NewSimpleValidationMonitor creates a simple validation monitor.
func NewSimpleValidationMonitor(collector ValidationCollector) *SimpleValidationMonitor {
	return &SimpleValidationMonitor{
		ValidationMonitor: NewValidationMonitor(collector),
	}
}
