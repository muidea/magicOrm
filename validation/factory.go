package validation

import (
	"reflect"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/constraints"
	"github.com/muidea/magicOrm/validation/database"
	"github.com/muidea/magicOrm/validation/errors"
	"github.com/muidea/magicOrm/validation/types"
)

// ValidationFactory creates and manages validation components
type ValidationFactory interface {
	// CreateValidationManager creates a new validation manager with given configuration
	CreateValidationManager(config ValidationConfig) ValidationManager

	// CreateTypeValidator creates a type validator
	CreateTypeValidator() TypeValidator

	// CreateConstraintValidator creates a constraint validator
	CreateConstraintValidator(enableCaching bool) ConstraintValidator

	// CreateDatabaseValidator creates a database validator
	CreateDatabaseValidator() DatabaseValidator

	// CreateScenarioAdapter creates a scenario adapter
	CreateScenarioAdapter() ScenarioAdapter

	// GetDefaultConfig returns the default validation configuration
	GetDefaultConfig() ValidationConfig

	// GetSimpleConfig returns a simple configuration for basic validation
	GetSimpleConfig() ValidationConfig

	// RegisterCustomConstraint registers a custom constraint globally
	RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error

	// RegisterTypeHandler registers a custom type handler globally
	RegisterTypeHandler(typeName string, handler types.TypeHandler) error
}

// TypeValidator validates basic type compatibility
type TypeValidator interface {
	ValidateType(value any, fieldType reflect.Type) error
	Convert(value any, targetType reflect.Type) (any, error)
	GetSupportedTypes() []reflect.Type
	RegisterTypeHandler(typeName string, handler types.TypeHandler) error
	GetZeroValue(fieldType reflect.Type) any
}

// ConstraintValidator validates business constraints
type ConstraintValidator interface {
	ValidateConstraints(value any, constraints models.Constraints, scenario errors.Scenario) error
	GetApplicableConstraints(scenario errors.Scenario) []models.Key
	RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error
	ClearCache()
}

// DatabaseValidator validates database-specific constraints
type DatabaseValidator interface {
	ValidateDatabaseConstraints(value any, fieldName string, constraints models.Constraints, dbType string) error
	GetDatabaseConstraints(constraints models.Constraints) []string
	ConvertToDatabaseValue(value any) (any, error)
}

// validationFactoryImpl implements ValidationFactory
type validationFactoryImpl struct {
	customConstraints map[models.Key]models.ValidatorFunc
	typeHandlers      map[string]types.TypeHandler
}

// NewValidationFactory creates a new validation factory
func NewValidationFactory() ValidationFactory {
	return &validationFactoryImpl{
		customConstraints: make(map[models.Key]models.ValidatorFunc),
		typeHandlers:      make(map[string]types.TypeHandler),
	}
}

// CreateValidationManager creates a new validation manager with given configuration
func (f *validationFactoryImpl) CreateValidationManager(config ValidationConfig) ValidationManager {
	// Create validators
	typeValidator := f.CreateTypeValidator()
	constraintValidator := f.CreateConstraintValidator(config.EnableCaching)
	databaseValidator := f.CreateDatabaseValidator()
	scenarioAdapter := f.CreateScenarioAdapter()

	// Create manager
	manager := &validationManagerImpl{
		typeValidator:       typeValidator,
		constraintValidator: constraintValidator,
		databaseValidator:   databaseValidator,
		scenarioAdapter:     scenarioAdapter,
		enabledLayers:       make(map[ValidationLayer]bool),
		config:              config,
		stats:               newValidationStats(),
	}

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

// CreateTypeValidator creates a type validator
func (f *validationFactoryImpl) CreateTypeValidator() TypeValidator {
	// Import types package
	validator := types.NewTypeValidator()

	// Register custom type handlers
	for typeName, handler := range f.typeHandlers {
		validator.RegisterTypeHandler(typeName, handler)
	}

	return validator
}

// CreateConstraintValidator creates a constraint validator
func (f *validationFactoryImpl) CreateConstraintValidator(enableCaching bool) ConstraintValidator {
	// Import constraints package
	validator := constraints.NewConstraintValidator(enableCaching)

	// Register custom constraints
	for key, handler := range f.customConstraints {
		validator.RegisterCustomConstraint(key, handler)
	}

	return validator
}

// CreateDatabaseValidator creates a database validator
func (f *validationFactoryImpl) CreateDatabaseValidator() DatabaseValidator {
	// Import database package
	return database.NewDatabaseValidator()
}

// CreateScenarioAdapter creates a scenario adapter
func (f *validationFactoryImpl) CreateScenarioAdapter() ScenarioAdapter {
	return NewScenarioAdapter()
}

// GetDefaultConfig returns the default validation configuration
func (f *validationFactoryImpl) GetDefaultConfig() ValidationConfig {
	return DefaultConfig()
}

// GetSimpleConfig returns a simple configuration for basic validation
func (f *validationFactoryImpl) GetSimpleConfig() ValidationConfig {
	return SimpleConfig()
}

// RegisterCustomConstraint registers a custom constraint globally
func (f *validationFactoryImpl) RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error {
	f.customConstraints[key] = validator
	return nil
}

// RegisterTypeHandler registers a custom type handler globally
func (f *validationFactoryImpl) RegisterTypeHandler(typeName string, handler types.TypeHandler) error {
	f.typeHandlers[typeName] = handler
	return nil
}
