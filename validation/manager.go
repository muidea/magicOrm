package validation

import (
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/internal/utils"
	"github.com/muidea/magicOrm/internal/validation/cache"
	"github.com/muidea/magicOrm/internal/validation/constraints"
	"github.com/muidea/magicOrm/internal/validation/database"
	"github.com/muidea/magicOrm/internal/validation/types"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/errors"
)

const (
	LayerType       ValidationLayer = "type"
	LayerConstraint ValidationLayer = "constraint"
	LayerDatabase   ValidationLayer = "database"
	LayerScenario   ValidationLayer = "scenario"
)

const (
	OperationCreate OperationType = "create"
	OperationRead   OperationType = "read"
	OperationUpdate OperationType = "update"
	OperationDelete OperationType = "delete"
)

// validationManagerImpl implements ValidationManager
type validationManagerImpl struct {
	typeValidator       types.TypeValidator
	constraintValidator constraints.ConstraintValidator
	databaseValidator   database.DatabaseValidator
	scenarioAdapter     ScenarioAdapter

	enabledLayers   map[ValidationLayer]bool
	config          ValidationConfig
	stats           validationStatsImpl
	defaultScenario errors.Scenario
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
	manager.constraintValidator = constraints.NewConstraintValidatorWithConfig(validationCacheConfig(config))
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

// WithOptions marks a complete call-level option snapshot, including explicit false.
func (context ValidationContext) WithOptions(options ValidationOptions) ValidationContext {
	context.Options = options
	context.optionsExplicit = true
	return context
}

func validationCacheConfig(config ValidationConfig) cache.CacheConfig {
	result := cache.DefaultCacheConfig()
	result.Enabled = config.EnableCaching
	if config.CacheTTL < 0 || config.MaxCacheSize < 0 {
		result.Enabled = false
		slog.Warn("negative validation cache configuration; caching disabled")
	}
	if config.CacheTTL > 0 {
		result.DefaultTTL = config.CacheTTL
	}
	if config.MaxCacheSize > 0 {
		result.MaxConstraintEntries = config.MaxCacheSize
	}
	return result
}

// resolveContext runs once at each public boundary, before any early return.
// Internal validation receives only this call's resolved snapshot.
func (m *validationManagerImpl) resolveContext(context ValidationContext) (ValidationContext, error) {
	if context.Scenario == "" {
		context.Scenario = m.defaultScenario
	}
	if context.Scenario == "" {
		context.Scenario = errors.ScenarioInsert
	}
	switch context.Scenario {
	case errors.ScenarioInsert, errors.ScenarioUpdate, errors.ScenarioQuery, errors.ScenarioDelete:
	default:
		return context, fmt.Errorf("unknown validation scenario: %s", context.Scenario)
	}
	options := m.config.DefaultOptions
	if m.enabledLayers[LayerScenario] {
		options = m.applyScenarioOptions(options, m.scenarioAdapter.GetValidationStrategy(context.Scenario))
	}
	if context.optionsExplicit {
		options = context.Options
	}
	context.Options = options
	if context.Collector == nil {
		context.Collector = errors.NewErrorCollector()
	}
	return context, nil
}

func (m *validationManagerImpl) Validate(value any, context ValidationContext) error {
	context, err := m.resolveContext(context)
	if err != nil {
		return err
	}
	return m.validateResolved(value, context)
}

// Validate validates a value with the given context
func (m *validationManagerImpl) validateResolved(value any, context ValidationContext) error {
	startTime := time.Now()
	m.stats.TotalValidations++
	modelName := getValidationModelName(context)

	defer func() {
		var recordErr error
		if context.Collector != nil && context.Collector.HasErrors() {
			recordErr = context.Collector.ToRichError()
		}
		recordValidationMetric(context.Operation, modelName, context.Scenario, time.Since(startTime), recordErr)
	}()

	var err error

	// Perform validation based on enabled layers
	if m.enabledLayers[LayerType] && context.Field != nil {
		err = m.validateType(value, context)
		if err != nil && context.Options.StopOnFirstError {
			m.recordValidationResult(startTime, false)
			return context.Collector.ToRichError()
		}
	}

	if m.enabledLayers[LayerConstraint] && context.Field != nil {
		err = m.validateConstraints(value, context)
		if err != nil && context.Options.StopOnFirstError {
			m.recordValidationResult(startTime, false)
			return context.Collector.ToRichError()
		}
	}

	if m.enabledLayers[LayerDatabase] && context.Field != nil && context.DatabaseType != "" {
		err = m.validateDatabase(value, context)
		if err != nil && context.Options.StopOnFirstError {
			m.recordValidationResult(startTime, false)
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
	context, err := m.resolveContext(context)
	if err != nil {
		return err
	}
	return m.validateFieldResolved(field, value, context)
}

func (m *validationManagerImpl) validateFieldResolved(field models.Field, value any, context ValidationContext) error {
	// A zero auto-increment value requests database generation. It is not a
	// supplied value and cannot satisfy required/range constraints before INSERT.
	// Keep validation for explicit values and for every non-insert scenario.
	if context.Scenario == errors.ScenarioInsert && field != nil && field.GetSpec() != nil &&
		field.GetSpec().GetValueDeclare() == models.AutoIncrement &&
		(value == nil || reflect.ValueOf(value).IsZero()) {
		return nil
	}
	context.Field = AdaptField(field, value)
	return m.validateResolved(value, context)
}

// ValidateModel validates an entire model
func (m *validationManagerImpl) ValidateModel(model models.Model, context ValidationContext) error {
	context, resolveErr := m.resolveContext(context)
	if resolveErr != nil {
		return resolveErr
	}
	if context.Collector == nil {
		context.Collector = errors.NewErrorCollector()
	}

	if context.Model != nil {
		for _, field := range context.Model.GetFields() {
			if adapted, ok := field.(*fieldAdapterImpl); ok && adapted.source != nil {
				if context.Scenario == errors.ScenarioUpdate && !models.IsPrimaryField(adapted.source) && !models.IsAssignedField(adapted.source) {
					continue
				}
				value := adapted.value
				fieldContext := context
				err := m.validateFieldResolved(adapted.source, value, fieldContext)
				if err != nil && context.Options.StopOnFirstError {
					return err
				}
				continue
			}
			fieldValue := field.GetValue()
			var value any
			if fieldValue != nil {
				value = fieldValue
				if rawValue, ok := fieldValue.(interface{ Get() any }); ok {
					value = rawValue.Get()
				}
			}

			fieldContext := context
			fieldContext.Field = field

			err := m.validateResolved(value, fieldContext)
			if err != nil && context.Options.StopOnFirstError {
				return err
			}
		}

		if context.Collector.HasErrors() {
			return context.Collector.ToRichError()
		}

		return nil
	}

	if model != nil {
		modelAdapter := AdaptModel(model)
		for _, field := range model.GetFields() {
			if context.Scenario == errors.ScenarioUpdate &&
				!models.IsPrimaryField(field) &&
				!models.IsAssignedField(field) {
				continue
			}

			fieldValue := field.GetValue()
			var value any
			if fieldValue != nil {
				value = fieldValue.Get()
			}

			fieldContext := context
			fieldContext.Model = modelAdapter
			err := m.validateFieldResolved(field, value, fieldContext)
			if err != nil && context.Options.StopOnFirstError {
				return err
			}
		}

		if context.Collector.HasErrors() {
			return context.Collector.ToRichError()
		}

		return nil
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
	m.defaultScenario = scenario
}

// Reset releases per-use rules and caches before a manager is returned to a pool.
func (m *validationManagerImpl) Reset() {
	m.constraintValidator.ClearCache()
	fresh := NewValidationManager(m.config).(*validationManagerImpl)
	*m = *fresh
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
	if fieldType == nil {
		return nil
	}
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
	applicableDirectives := m.constraintValidator.GetApplicableDirectives(constraints, context.Scenario)
	err := m.constraintValidator.ValidateConstraints(value, constraints, context.Scenario)
	recordConstraintChecks(context.Field.GetName(), directiveKeys(applicableDirectives), err == nil)
	if err != nil {
		key, message := utils.ConstraintErrorDetails(err)
		validationErr := errors.NewValidationError(message).
			WithField(context.Field.GetName()).WithConstraint(string(key)).
			WithScenario(context.Scenario)

		context.Collector.AddError(validationErr)
		return err
	}

	return nil
}

func getValidationModelName(context ValidationContext) string {
	if context.Model != nil {
		return context.Model.GetName()
	}
	return "unknown"
}

func directiveKeys(directives []models.Directive) []string {
	if len(directives) == 0 {
		return nil
	}

	keys := make([]string, 0, len(directives))
	for _, directive := range directives {
		keys = append(keys, string(directive.Key()))
	}
	return keys
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
