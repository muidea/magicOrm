package validation

import (
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/constraints"
	"github.com/muidea/magicOrm/validation/database"
	"github.com/muidea/magicOrm/validation/errors"
	"github.com/muidea/magicOrm/validation/types"
)

// ValidationLayer represents a validation layer
type ValidationLayer string

const (
	LayerType       ValidationLayer = "type"
	LayerConstraint ValidationLayer = "constraint"
	LayerDatabase   ValidationLayer = "database"
	LayerScenario   ValidationLayer = "scenario"
)

// OperationType represents the type of database operation
type OperationType string

const (
	OperationCreate OperationType = "create"
	OperationRead   OperationType = "read"
	OperationUpdate OperationType = "update"
	OperationDelete OperationType = "delete"
)

// ValidationContext contains context for validation
type ValidationContext struct {
	Scenario     errors.Scenario
	Operation    OperationType
	Model        ModelAdapter
	Field        FieldAdapter
	DatabaseType string
	Options      ValidationOptions
	Collector    errors.ErrorCollector
}

// ValidationOptions contains validation configuration options
type ValidationOptions struct {
	StopOnFirstError        bool
	IncludeFieldPathInError bool
	ValidateReadOnlyFields  bool
	ValidateWriteOnlyFields bool
}

// ValidationStats contains validation statistics
type ValidationStats struct {
	TotalValidations      int64
	SuccessfulValidations int64
	FailedValidations     int64
	TypeValidations       int64
	ConstraintValidations int64
	DatabaseValidations   int64
	AverageTime           time.Duration
	CacheHits             int64
	CacheMisses           int64
}

// ValidationManager orchestrates validation across all layers
type ValidationManager interface {
	// Validate validates a value with the given context
	Validate(value any, context ValidationContext) error

	// ValidateField validates a single field
	ValidateField(field models.Field, value any, context ValidationContext) error

	// ValidateModel validates an entire model
	ValidateModel(model models.Model, context ValidationContext) error

	// Configuration
	EnableLayer(layer ValidationLayer) error
	DisableLayer(layer ValidationLayer) error
	SetScenario(scenario errors.Scenario)

	// Statistics
	GetValidationStats() ValidationStats
	ResetStats()

	// Utility methods
	RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error
	RegisterTypeHandler(typeName string, handler types.TypeHandler) error
}

// validationManagerImpl implements ValidationManager
type validationManagerImpl struct {
	typeValidator       types.TypeValidator
	constraintValidator constraints.ConstraintValidator
	databaseValidator   database.DatabaseValidator
	scenarioAdapter     ScenarioAdapter

	enabledLayers map[ValidationLayer]bool
	config        ValidationConfig
	stats         validationStatsImpl
}

// ValidationConfig contains configuration for the validation manager
type ValidationConfig struct {
	EnableTypeValidation       bool
	EnableConstraintValidation bool
	EnableDatabaseValidation   bool
	EnableScenarioAdaptation   bool
	EnableCaching              bool
	CacheTTL                   time.Duration
	MaxCacheSize               int
	DefaultOptions             ValidationOptions
}

// NewValidationManager creates a new validation manager
func NewValidationManager(config ValidationConfig) ValidationManager {
	manager := &validationManagerImpl{
		enabledLayers: make(map[ValidationLayer]bool),
		config:        config,
		stats:         newValidationStats(),
	}

	// Initialize validators
	manager.typeValidator = types.NewTypeValidator()
	manager.constraintValidator = constraints.NewConstraintValidator(config.EnableCaching)
	manager.databaseValidator = database.NewDatabaseValidator()
	manager.scenarioAdapter = NewScenarioAdapter()

	// Enable layers based on config
	if config.EnableTypeValidation {
		manager.enabledLayers[LayerType] = true
	}
	if config.EnableConstraintValidation {
		manager.enabledLayers[LayerConstraint] = true
	}
	if config.EnableDatabaseValidation {
		manager.enabledLayers[LayerDatabase] = true
	}
	if config.EnableScenarioAdaptation {
		manager.enabledLayers[LayerScenario] = true
	}

	return manager
}

// Validate validates a value with the given context
func (m *validationManagerImpl) Validate(value any, context ValidationContext) error {
	startTime := time.Now()
	m.stats.TotalValidations++

	var err error

	// Apply scenario adaptation if enabled
	if m.enabledLayers[LayerScenario] {
		strategy := m.scenarioAdapter.GetValidationStrategy(context.Scenario)
		context.Options = m.applyScenarioOptions(context.Options, strategy)
	}

	// Create error collector if not provided
	if context.Collector == nil {
		context.Collector = errors.NewErrorCollector()
	}

	// Perform validation based on enabled layers
	if m.enabledLayers[LayerType] && context.Field != nil {
		err = m.validateType(value, context)
		if err != nil && context.Options.StopOnFirstError {
			m.recordValidationResult(startTime, err == nil)
			return context.Collector.ToRichError()
		}
	}

	if m.enabledLayers[LayerConstraint] && context.Field != nil {
		err = m.validateConstraints(value, context)
		if err != nil && context.Options.StopOnFirstError {
			m.recordValidationResult(startTime, err == nil)
			return context.Collector.ToRichError()
		}
	}

	if m.enabledLayers[LayerDatabase] && context.Field != nil && context.DatabaseType != "" {
		err = m.validateDatabase(value, context)
		if err != nil && context.Options.StopOnFirstError {
			m.recordValidationResult(startTime, err == nil)
			return context.Collector.ToRichError()
		}
	}

	m.recordValidationResult(startTime, !context.Collector.HasErrors())

	if context.Collector.HasErrors() {
		return context.Collector.ToRichError()
	}

	return nil
}

// ValidateField validates a single field
func (m *validationManagerImpl) ValidateField(field models.Field, value any, context ValidationContext) error {
	// For now, create a simple field adapter
	// In production, we would extract actual constraints from the field
	fieldAdapter := NewFieldAdapter(
		field.GetName(),
		nil, // field type
		nil, // constraints
		value,
	)
	context.Field = fieldAdapter
	return m.Validate(value, context)
}

// ValidateModel validates an entire model
func (m *validationManagerImpl) ValidateModel(model models.Model, context ValidationContext) error {
	// For now, create an empty model adapter
	// In production, we would extract fields from the model
	context.Model = NewModelAdapter([]FieldAdapter{})
	context.Collector = errors.NewErrorCollector()

	// Get all fields from the model adapter
	fields := context.Model.GetFields()

	// Validate each field
	for _, field := range fields {
		fieldValue := field.GetValue()

		fieldContext := context
		fieldContext.Field = field

		err := m.Validate(fieldValue, fieldContext)
		if err != nil && context.Options.StopOnFirstError {
			return err
		}
	}

	if context.Collector.HasErrors() {
		return context.Collector.ToRichError()
	}

	return nil
}

// EnableLayer enables a validation layer
func (m *validationManagerImpl) EnableLayer(layer ValidationLayer) error {
	m.enabledLayers[layer] = true
	return nil
}

// DisableLayer disables a validation layer
func (m *validationManagerImpl) DisableLayer(layer ValidationLayer) error {
	delete(m.enabledLayers, layer)
	return nil
}

// SetScenario sets the current scenario
func (m *validationManagerImpl) SetScenario(scenario errors.Scenario) {
	// Scenario is handled through context, but we can update default behavior
}

// GetValidationStats returns validation statistics
func (m *validationManagerImpl) GetValidationStats() ValidationStats {
	return m.stats.toValidationStats()
}

// ResetStats resets validation statistics
func (m *validationManagerImpl) ResetStats() {
	m.stats = newValidationStats()
}

// RegisterCustomConstraint registers a custom constraint
func (m *validationManagerImpl) RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error {
	return m.constraintValidator.RegisterCustomConstraint(key, validator)
}

// RegisterTypeHandler registers a custom type handler
func (m *validationManagerImpl) RegisterTypeHandler(typeName string, handler types.TypeHandler) error {
	return m.typeValidator.RegisterTypeHandler(typeName, handler)
}

// validateType performs type validation
func (m *validationManagerImpl) validateType(value any, context ValidationContext) error {
	m.stats.TypeValidations++

	if context.Field == nil {
		return nil
	}

	fieldType := context.Field.GetType()
	err := m.typeValidator.ValidateType(value, fieldType)
	if err != nil {
		validationErr := errors.NewTypeError(
			context.Field.GetName(),
			value,
			fieldType.String(),
		).WithScenario(context.Scenario)

		context.Collector.AddError(validationErr)
		return err
	}

	return nil
}

// validateConstraints performs constraint validation
func (m *validationManagerImpl) validateConstraints(value any, context ValidationContext) error {
	m.stats.ConstraintValidations++

	if context.Field == nil {
		return nil
	}

	// Skip read-only fields in update scenarios if configured
	if !context.Options.ValidateReadOnlyFields &&
		context.Scenario == errors.ScenarioUpdate &&
		context.Field.HasConstraint(models.KeyReadOnly) {
		return nil
	}

	// Skip write-only fields in query scenarios if configured
	if !context.Options.ValidateWriteOnlyFields &&
		context.Scenario == errors.ScenarioQuery &&
		context.Field.HasConstraint(models.KeyWriteOnly) {
		return nil
	}

	constraints := context.Field.GetConstraints()
	err := m.constraintValidator.ValidateConstraints(value, constraints, context.Scenario)
	if err != nil {
		// Extract constraint information from error if possible
		validationErr := errors.NewConstraintError(
			context.Field.GetName(),
			"", // Will be filled by constraint validator
			value,
			nil,
		).WithScenario(context.Scenario)

		context.Collector.AddError(validationErr)
		return err
	}

	return nil
}

// validateDatabase performs database validation
func (m *validationManagerImpl) validateDatabase(value any, context ValidationContext) error {
	m.stats.DatabaseValidations++

	if context.Field == nil {
		return nil
	}

	err := m.databaseValidator.ValidateDatabaseConstraints(
		value,
		context.Field.GetName(),
		context.Field.GetConstraints(),
		context.DatabaseType,
	)
	if err != nil {
		validationErr := errors.NewDatabaseError(
			context.Field.GetName(),
			value,
			"NOT NULL", // Simplified for now
		).WithScenario(context.Scenario)

		context.Collector.AddError(validationErr)
		return err
	}

	return nil
}

// applyScenarioOptions applies scenario-specific options
func (m *validationManagerImpl) applyScenarioOptions(options ValidationOptions, strategy ValidationStrategy) ValidationOptions {
	// Apply strategy-specific options
	// This is a simplified implementation - actual implementation would use strategy
	if options.ValidateReadOnlyFields && strategy.ShouldSkipReadOnlyFields() {
		options.ValidateReadOnlyFields = false
	}

	if options.ValidateWriteOnlyFields && strategy.ShouldSkipWriteOnlyFields() {
		options.ValidateWriteOnlyFields = false
	}

	return options
}

// recordValidationResult records validation statistics
func (m *validationManagerImpl) recordValidationResult(startTime time.Time, success bool) {
	duration := time.Since(startTime)

	if success {
		m.stats.SuccessfulValidations++
	} else {
		m.stats.FailedValidations++
	}

	// Update average time (simplified moving average)
	totalTime := m.stats.AverageTime * time.Duration(m.stats.TotalValidations-1)
	m.stats.AverageTime = (totalTime + duration) / time.Duration(m.stats.TotalValidations)
}

// validationStatsImpl implements statistics tracking
type validationStatsImpl struct {
	TotalValidations      int64
	SuccessfulValidations int64
	FailedValidations     int64
	TypeValidations       int64
	ConstraintValidations int64
	DatabaseValidations   int64
	AverageTime           time.Duration
	CacheHits             int64
	CacheMisses           int64
}

func newValidationStats() validationStatsImpl {
	return validationStatsImpl{}
}

func (s *validationStatsImpl) toValidationStats() ValidationStats {
	return ValidationStats{
		TotalValidations:      s.TotalValidations,
		SuccessfulValidations: s.SuccessfulValidations,
		FailedValidations:     s.FailedValidations,
		TypeValidations:       s.TypeValidations,
		ConstraintValidations: s.ConstraintValidations,
		DatabaseValidations:   s.DatabaseValidations,
		AverageTime:           s.AverageTime,
		CacheHits:             s.CacheHits,
		CacheMisses:           s.CacheMisses,
	}
}

// Helper functions

// NewContext creates a new validation context
func NewContext(scenario errors.Scenario, operation OperationType, model ModelAdapter, dbType string) ValidationContext {
	return ValidationContext{
		Scenario:     scenario,
		Operation:    operation,
		Model:        model,
		DatabaseType: dbType,
		Options: ValidationOptions{
			StopOnFirstError:        false,
			IncludeFieldPathInError: true,
			ValidateReadOnlyFields:  true,
			ValidateWriteOnlyFields: true,
		},
	}
}

// NewContextWithCollector creates a new validation context with error collector
func NewContextWithCollector(scenario errors.Scenario, collector errors.ErrorCollector) ValidationContext {
	ctx := NewContext(scenario, OperationCreate, nil, "")
	ctx.Collector = collector
	return ctx
}

// DefaultConfig returns the default validation configuration
func DefaultConfig() ValidationConfig {
	return ValidationConfig{
		EnableTypeValidation:       true,
		EnableConstraintValidation: true,
		EnableDatabaseValidation:   true,
		EnableScenarioAdaptation:   true,
		EnableCaching:              true,
		CacheTTL:                   5 * time.Minute,
		MaxCacheSize:               1000,
		DefaultOptions: ValidationOptions{
			StopOnFirstError:        false,
			IncludeFieldPathInError: true,
			ValidateReadOnlyFields:  true,
			ValidateWriteOnlyFields: true,
		},
	}
}

// SimpleConfig returns a simple configuration for basic validation
func SimpleConfig() ValidationConfig {
	return ValidationConfig{
		EnableTypeValidation:       true,
		EnableConstraintValidation: true,
		EnableDatabaseValidation:   false,
		EnableScenarioAdaptation:   false,
		EnableCaching:              false,
		DefaultOptions: ValidationOptions{
			StopOnFirstError:        true,
			IncludeFieldPathInError: false,
			ValidateReadOnlyFields:  false,
			ValidateWriteOnlyFields: false,
		},
	}
}
