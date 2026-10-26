package validation

import (
	"reflect"
	"time"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/validation/errors"
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
	RegisterTypeHandler(typeName string, handler TypeHandler) error
}

// TypeValidator validates basic type compatibility
type TypeValidator interface {
	ValidateType(value any, fieldType reflect.Type) error
	Convert(value any, targetType reflect.Type) (any, error)
	GetSupportedTypes() []reflect.Type
	RegisterTypeHandler(typeName string, handler TypeHandler) error
	GetZeroValue(fieldType reflect.Type) any
}

// ConstraintValidator validates business constraints
type ConstraintValidator interface {
	ValidateConstraints(value any, constraints models.Constraints, scenario errors.Scenario) error
	GetApplicableDirectives(constraints models.Constraints, scenario errors.Scenario) []models.Directive
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

// ValidationLayer represents a validation layer
type ValidationLayer string

// OperationType represents the type of database operation
type OperationType string

// ValidationContext contains context for validation
type ValidationContext struct {
	Scenario        errors.Scenario
	Operation       OperationType
	Model           ModelAdapter
	Field           FieldAdapter
	DatabaseType    string
	Options         ValidationOptions
	Collector       errors.ErrorCollector
	optionsExplicit bool
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
	RegisterTypeHandler(typeName string, handler TypeHandler) error
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

// TypeHandler is the application extension contract for custom value conversion.
type TypeHandler = models.TypeHandler
