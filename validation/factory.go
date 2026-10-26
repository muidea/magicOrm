package validation

import (
	"github.com/muidea/magicOrm/internal/validation/constraints"
	"github.com/muidea/magicOrm/internal/validation/database"
	"github.com/muidea/magicOrm/internal/validation/types"
	"github.com/muidea/magicOrm/models"
)

// validationFactoryImpl implements ValidationFactory
type validationFactoryImpl struct {
	customConstraints map[models.Key]models.ValidatorFunc
	typeHandlers      map[string]TypeHandler
}

// NewValidationFactory creates a new validation factory
func NewValidationFactory() ValidationFactory {
	return &validationFactoryImpl{
		customConstraints: make(map[models.Key]models.ValidatorFunc),
		typeHandlers:      make(map[string]TypeHandler),
	}
}

// CreateValidationManager creates a new validation manager with given configuration
func (f *validationFactoryImpl) CreateValidationManager(config ValidationConfig) ValidationManager {
	manager := NewValidationManager(config).(*validationManagerImpl)
	manager.typeValidator = f.CreateTypeValidator()
	for key, handler := range f.customConstraints {
		_ = manager.RegisterCustomConstraint(key, handler)
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
func (f *validationFactoryImpl) RegisterTypeHandler(typeName string, handler TypeHandler) error {
	f.typeHandlers[typeName] = handler
	return nil
}
