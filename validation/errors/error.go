package errors

import (
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"
)

// ValidationLayer represents which validation layer generated the error
type ValidationLayer string

const (
	LayerType       ValidationLayer = "type"
	LayerConstraint ValidationLayer = "constraint"
	LayerDatabase   ValidationLayer = "database"
	LayerScenario   ValidationLayer = "scenario"
)

// Scenario represents the operation scenario
type Scenario string

const (
	ScenarioInsert Scenario = "insert"
	ScenarioUpdate Scenario = "update"
	ScenarioQuery  Scenario = "query"
	ScenarioDelete Scenario = "delete"
)

// ValidationError represents a validation error with rich context
type ValidationError interface {
	error
	GetField() string
	GetConstraint() string
	GetValue() any
	GetExpected() any
	GetLayer() ValidationLayer
	GetScenario() Scenario
	ToRichError() *cd.Error
	WithField(field string) ValidationError
	WithConstraint(constraint string) ValidationError
	WithScenario(scenario Scenario) ValidationError
}

// validationErrorImpl implements ValidationError
type validationErrorImpl struct {
	field      string
	constraint string
	value      any
	expected   any
	layer      ValidationLayer
	scenario   Scenario
	message    string
}

// NewValidationError creates a new validation error
func NewValidationError(message string) ValidationError {
	return &validationErrorImpl{
		message: message,
		layer:   LayerConstraint, // Default layer
	}
}

// NewTypeError creates a type validation error
func NewTypeError(field string, value any, expectedType string) ValidationError {
	return &validationErrorImpl{
		field:    field,
		value:    value,
		expected: expectedType,
		layer:    LayerType,
		message:  fmt.Sprintf("type mismatch for field '%s': got %T, expected %s", field, value, expectedType),
	}
}

// NewConstraintError creates a constraint validation error
func NewConstraintError(field, constraint string, value, expected any) ValidationError {
	return &validationErrorImpl{
		field:      field,
		constraint: constraint,
		value:      value,
		expected:   expected,
		layer:      LayerConstraint,
		message:    fmt.Sprintf("constraint '%s' violation for field '%s': got %v", constraint, field, value),
	}
}

// NewDatabaseError creates a database validation error
func NewDatabaseError(field string, value any, dbConstraint string) ValidationError {
	return &validationErrorImpl{
		field:      field,
		constraint: dbConstraint,
		value:      value,
		layer:      LayerDatabase,
		message:    fmt.Sprintf("database constraint '%s' violation for field '%s': %v", dbConstraint, field, value),
	}
}

// Error returns the error message
func (e *validationErrorImpl) Error() string {
	return e.message
}

// GetField returns the field name
func (e *validationErrorImpl) GetField() string {
	return e.field
}

// GetConstraint returns the constraint name
func (e *validationErrorImpl) GetConstraint() string {
	return e.constraint
}

// GetValue returns the actual value
func (e *validationErrorImpl) GetValue() any {
	return e.value
}

// GetExpected returns the expected value or type
func (e *validationErrorImpl) GetExpected() any {
	return e.expected
}

// GetLayer returns the validation layer
func (e *validationErrorImpl) GetLayer() ValidationLayer {
	return e.layer
}

// GetScenario returns the operation scenario
func (e *validationErrorImpl) GetScenario() Scenario {
	return e.scenario
}

// ToRichError converts to magicCommon error type
func (e *validationErrorImpl) ToRichError() *cd.Error {
	var code cd.Code = cd.IllegalParam
	if e.layer == LayerDatabase {
		code = cd.DatabaseError
	}

	// Build detailed message
	message := e.message
	if e.field != "" {
		message = fmt.Sprintf("field '%s': %s", e.field, e.message)
	}

	return cd.NewError(code, message)
}

// WithField sets the field name
func (e *validationErrorImpl) WithField(field string) ValidationError {
	e.field = field
	return e
}

// WithConstraint sets the constraint name
func (e *validationErrorImpl) WithConstraint(constraint string) ValidationError {
	e.constraint = constraint
	return e
}

// WithScenario sets the operation scenario
func (e *validationErrorImpl) WithScenario(scenario Scenario) ValidationError {
	e.scenario = scenario
	return e
}

// ErrorCollector collects multiple validation errors
type ErrorCollector interface {
	// AddError adds a validation error
	AddError(err ValidationError)

	// HasErrors returns true if there are any errors
	HasErrors() bool

	// GetErrors returns all collected errors
	GetErrors() []ValidationError

	// GetErrorsByField returns errors for a specific field
	GetErrorsByField(field string) []ValidationError

	// GetErrorsByLayer returns errors for a specific validation layer
	GetErrorsByLayer(layer ValidationLayer) []ValidationError

	// GetErrorSummary returns a summary of all errors
	GetErrorSummary() string

	// Clear clears all collected errors
	Clear()

	// ToRichError converts all errors to a single rich error
	ToRichError() *cd.Error
}

// errorCollectorImpl implements ErrorCollector
type errorCollectorImpl struct {
	errors []ValidationError
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() ErrorCollector {
	return &errorCollectorImpl{
		errors: make([]ValidationError, 0),
	}
}

// AddError adds a validation error
func (c *errorCollectorImpl) AddError(err ValidationError) {
	c.errors = append(c.errors, err)
}

// HasErrors returns true if there are any errors
func (c *errorCollectorImpl) HasErrors() bool {
	return len(c.errors) > 0
}

// GetErrors returns all collected errors
func (c *errorCollectorImpl) GetErrors() []ValidationError {
	return c.errors
}

// GetErrorsByField returns errors for a specific field
func (c *errorCollectorImpl) GetErrorsByField(field string) []ValidationError {
	result := make([]ValidationError, 0)
	for _, err := range c.errors {
		if err.GetField() == field {
			result = append(result, err)
		}
	}
	return result
}

// GetErrorsByLayer returns errors for a specific validation layer
func (c *errorCollectorImpl) GetErrorsByLayer(layer ValidationLayer) []ValidationError {
	result := make([]ValidationError, 0)
	for _, err := range c.errors {
		if err.GetLayer() == layer {
			result = append(result, err)
		}
	}
	return result
}

// GetErrorSummary returns a summary of all errors
func (c *errorCollectorImpl) GetErrorSummary() string {
	if len(c.errors) == 0 {
		return "No validation errors"
	}

	summary := fmt.Sprintf("Found %d validation error(s):\n", len(c.errors))

	// Group errors by field
	fieldErrors := make(map[string][]ValidationError)
	for _, err := range c.errors {
		field := err.GetField()
		if field == "" {
			field = "<unknown>"
		}
		fieldErrors[field] = append(fieldErrors[field], err)
	}

	// Build summary
	for field, errors := range fieldErrors {
		summary += fmt.Sprintf("  Field '%s':\n", field)
		for _, err := range errors {
			constraint := err.GetConstraint()
			if constraint == "" {
				constraint = "<general>"
			}
			summary += fmt.Sprintf("    - [%s] %s\n", constraint, err.Error())
		}
	}

	return summary
}

// Clear clears all collected errors
func (c *errorCollectorImpl) Clear() {
	c.errors = make([]ValidationError, 0)
}

// ToRichError converts all errors to a single rich error
func (c *errorCollectorImpl) ToRichError() *cd.Error {
	if len(c.errors) == 0 {
		return nil
	}

	if len(c.errors) == 1 {
		return c.errors[0].ToRichError()
	}

	// Group by field for summary
	fieldSummary := make(map[string]int)
	for _, err := range c.errors {
		field := err.GetField()
		if field == "" {
			field = "<unknown>"
		}
		fieldSummary[field]++
	}

	fieldStrings := make([]string, 0, len(fieldSummary))
	for field, count := range fieldSummary {
		fieldStrings = append(fieldStrings, fmt.Sprintf("%s (%d)", field, count))
	}

	message := fmt.Sprintf("Multiple validation errors: %s", strings.Join(fieldStrings, ", "))
	return cd.NewError(cd.IllegalParam, message)
}

// ErrorBuilder helps build validation errors with fluent API
type ErrorBuilder interface {
	WithField(field string) ErrorBuilder
	WithConstraint(constraint string) ErrorBuilder
	WithValue(value any) ErrorBuilder
	WithExpected(expected any) ErrorBuilder
	WithLayer(layer ValidationLayer) ErrorBuilder
	WithScenario(scenario Scenario) ErrorBuilder
	WithMessage(message string) ErrorBuilder
	Build() ValidationError
}

// errorBuilderImpl implements ErrorBuilder
type errorBuilderImpl struct {
	field      string
	constraint string
	value      any
	expected   any
	layer      ValidationLayer
	scenario   Scenario
	message    string
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder() ErrorBuilder {
	return &errorBuilderImpl{
		layer: LayerConstraint, // Default layer
	}
}

// WithField sets the field name
func (b *errorBuilderImpl) WithField(field string) ErrorBuilder {
	b.field = field
	return b
}

// WithConstraint sets the constraint name
func (b *errorBuilderImpl) WithConstraint(constraint string) ErrorBuilder {
	b.constraint = constraint
	return b
}

// WithValue sets the actual value
func (b *errorBuilderImpl) WithValue(value any) ErrorBuilder {
	b.value = value
	return b
}

// WithExpected sets the expected value
func (b *errorBuilderImpl) WithExpected(expected any) ErrorBuilder {
	b.expected = expected
	return b
}

// WithLayer sets the validation layer
func (b *errorBuilderImpl) WithLayer(layer ValidationLayer) ErrorBuilder {
	b.layer = layer
	return b
}

// WithScenario sets the operation scenario
func (b *errorBuilderImpl) WithScenario(scenario Scenario) ErrorBuilder {
	b.scenario = scenario
	return b
}

// WithMessage sets the error message
func (b *errorBuilderImpl) WithMessage(message string) ErrorBuilder {
	b.message = message
	return b
}

// Build creates the validation error
func (b *errorBuilderImpl) Build() ValidationError {
	if b.message == "" {
		// Generate default message
		if b.constraint != "" && b.field != "" {
			b.message = fmt.Sprintf("constraint '%s' violation for field '%s'", b.constraint, b.field)
		} else if b.field != "" {
			b.message = fmt.Sprintf("validation error for field '%s'", b.field)
		} else {
			b.message = "validation error"
		}
	}

	return &validationErrorImpl{
		field:      b.field,
		constraint: b.constraint,
		value:      b.value,
		expected:   b.expected,
		layer:      b.layer,
		scenario:   b.scenario,
		message:    b.message,
	}
}
