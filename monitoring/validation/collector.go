// Package validation provides validation-specific metric collectors.
// This package focuses on collecting validation operation metrics only.
package validation

import (
	"time"

	"github.com/muidea/magicCommon/monitoring/types"
	"github.com/muidea/magicOrm/monitoring"
)

// ValidationCollector defines the interface for validation-specific metric collectors.
type ValidationCollector interface {
	// RecordValidation records a validation operation
	RecordValidation(operation string, modelName string, scenario string, startTime time.Time, err error, labels map[string]string)

	// RecordCacheAccess records cache access for validation
	RecordCacheAccess(cacheType string, operation string, hit bool, duration time.Duration, labels map[string]string)

	// RecordConstraintCheck records a constraint check operation
	RecordConstraintCheck(constraintType string, fieldName string, passed bool, duration time.Duration, labels map[string]string)

	// GetMetrics returns collected validation metrics
	GetMetrics() ([]types.Metric, error)
}

// SimpleValidationCollector is a simplified collector that only records metrics without complex logic
type SimpleValidationCollector struct {
	*monitoring.BaseCollector

	// Additional validation-specific metrics
	constraintChecks []ConstraintCheckMetric
}

// ConstraintCheckMetric represents a recorded constraint check
type ConstraintCheckMetric struct {
	ConstraintType string
	FieldName      string
	Passed         bool
	Duration       time.Duration
	Labels         map[string]string
}

// NewCollector creates a new validation collector
func NewCollector() ValidationCollector {
	return &SimpleValidationCollector{
		BaseCollector:    monitoring.NewBaseCollector(),
		constraintChecks: make([]ConstraintCheckMetric, 0),
	}
}

// RecordValidation implements ValidationCollector interface
func (c *SimpleValidationCollector) RecordValidation(operation string, modelName string, scenario string, startTime time.Time, err error, labels map[string]string) {
	c.BaseCollector.RecordValidation(operation, modelName, scenario, startTime, err, labels)
}

// RecordCacheAccess implements ValidationCollector interface
func (c *SimpleValidationCollector) RecordCacheAccess(cacheType string, operation string, hit bool, duration time.Duration, labels map[string]string) {
	c.BaseCollector.RecordCacheAccess(cacheType, operation, hit, duration, labels)
}

// RecordConstraintCheck implements ValidationCollector interface
func (c *SimpleValidationCollector) RecordConstraintCheck(constraintType string, fieldName string, passed bool, duration time.Duration, labels map[string]string) {
	c.constraintChecks = append(c.constraintChecks, ConstraintCheckMetric{
		ConstraintType: constraintType,
		FieldName:      fieldName,
		Passed:         passed,
		Duration:       duration,
		Labels:         monitoring.MergeLabels(monitoring.DefaultLabels(), labels),
	})
}

// GetMetrics implements ValidationCollector interface
func (c *SimpleValidationCollector) GetMetrics() ([]types.Metric, error) {
	// Get base metrics
	baseMetrics, err := c.BaseCollector.GetMetrics()
	if err != nil {
		return nil, err
	}

	// For now, return only base metrics
	// In real implementation, this would also include constraint check metrics
	return baseMetrics, nil
}

// Validation scenarios
const (
	ScenarioInsert = "insert"
	ScenarioUpdate = "update"
	ScenarioQuery  = "query"
	ScenarioDelete = "delete"
	ScenarioCreate = "create"
	ScenarioDrop   = "drop"
)

// Validation operations
const (
	OperationValidateModel   = "validate_model"
	OperationValidateField   = "validate_field"
	OperationCheckConstraint = "check_constraint"
	OperationTypeConversion  = "type_conversion"
)

// Constraint types
const (
	ConstraintRequired = "required"
	ConstraintMin      = "min"
	ConstraintMax      = "max"
	ConstraintRange    = "range"
	ConstraintRegex    = "regex"
	ConstraintIn       = "in"
)
