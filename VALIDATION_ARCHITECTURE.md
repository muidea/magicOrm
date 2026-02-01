# MagicORM Validation Architecture

## Overview

This document describes the new validation architecture for MagicORM, implementing a four-layer validation responsibility separation to address the current architectural issues where:
1. Provider layer cannot perceive CURD scenarios
2. Database layer contains business logic validation
3. Validation logic is scattered across multiple layers

## Architecture Design

### Four-Layer Validation Responsibility

#### 1. Type Validation Layer (`validation/types/`)
**Responsibility**: Basic type validation and conversion
- Validates Go type compatibility with database types
- Handles type conversions (string ↔ int, time formats, etc.)
- Ensures basic data integrity at the type level

**Key Interfaces**:
```go
type TypeValidator interface {
    ValidateType(value any, fieldType reflect.Type) error
    Convert(value any, targetType reflect.Type) (any, error)
    GetSupportedTypes() []reflect.Type
}
```

#### 2. Constraint Validation Layer (`validation/constraints/`)
**Responsibility**: Business constraint validation
- Validates business rules defined in struct tags (`req`, `min`, `max`, `range`, `in`, `re`)
- Handles access behavior constraints (`ro`, `wo`)
- Supports scenario-aware validation (Insert vs Update)

**Key Interfaces**:
```go
type ConstraintValidator interface {
    ValidateConstraints(value any, constraints models.Constraints, scenario Scenario) error
    GetApplicableConstraints(scenario Scenario) []models.Key
    RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc)
}
```

#### 3. Database Constraint Layer (`validation/database/`)
**Responsibility**: Database-specific constraint validation
- Validates database-level constraints (NOT NULL, UNIQUE, FOREIGN KEY, etc.)
- Handles database type compatibility
- Provides database-specific error messages

**Key Interfaces**:
```go
type DatabaseValidator interface {
    ValidateDatabaseConstraints(value any, field models.Field, dbType string) error
    GetDatabaseConstraints(field models.Field) []string
    ConvertToDatabaseValue(value any, field models.Field) (any, error)
}
```

#### 4. Scenario Adaptation Layer (`validation/scenario/`)
**Responsibility**: Scenario-aware validation orchestration
- Orchestrates validation based on operation type (Insert, Update, Query, Delete)
- Applies different validation strategies per scenario
- Provides validation context to other layers

**Key Interfaces**:
```go
type ScenarioAdapter interface {
    GetValidationStrategy(scenario Scenario) ValidationStrategy
    ShouldValidateConstraint(constraint models.Key, scenario Scenario) bool
    GetValidationContext(scenario Scenario) ValidationContext
}

type ValidationStrategy interface {
    Validate(value any, validators []Validator) error
    GetPriorityOrder() []ValidatorType
}
```

## Core Components

### Validation Manager

Central coordinator that orchestrates validation across all layers:

```go
type ValidationManager interface {
    Validate(value any, context ValidationContext) error
    ValidateField(field models.Field, value any, context ValidationContext) error
    ValidateModel(model models.Model, context ValidationContext) error
    
    // Configuration
    EnableLayer(layer ValidationLayer) error
    DisableLayer(layer ValidationLayer) error
    SetScenario(scenario Scenario)
    
    // Statistics
    GetValidationStats() ValidationStats
    ResetStats()
}

type ValidationContext struct {
    Scenario     Scenario
    Operation    OperationType
    Model        models.Model
    Field        models.Field
    DatabaseType string
    Options      ValidationOptions
}
```

### Error Handling System

Enhanced error handling with rich context:

```go
type ValidationError interface {
    error
    GetField() string
    GetConstraint() string
    GetValue() any
    GetExpected() any
    GetLayer() ValidationLayer
    GetScenario() Scenario
    ToRichError() *cd.Error
}

type ErrorCollector interface {
    AddError(err ValidationError)
    HasErrors() bool
    GetErrors() []ValidationError
    GetErrorsByField(field string) []ValidationError
    Clear()
}

type ValidationErrorBuilder struct {
    field      string
    constraint string
    value      any
    expected   any
    layer      ValidationLayer
    scenario   Scenario
    message    string
}
```

### Configuration System

Flexible configuration for validation behavior:

```go
type ValidationConfig struct {
    // Layer configuration
    EnableTypeValidation      bool
    EnableConstraintValidation bool
    EnableDatabaseValidation  bool
    EnableScenarioAdaptation  bool
    
    // Scenario-specific settings
    Scenarios map[Scenario]ScenarioConfig
    
    // Performance settings
    EnableCaching            bool
    CacheTTL                 time.Duration
    MaxCacheSize             int
    
    // Error handling
    StopOnFirstError         bool
    CollectAllErrors         bool
    IncludeFieldPathInError  bool
    
    // Logging
    EnableValidationLogging  bool
    LogLevel                 LogLevel
}

type ScenarioConfig struct {
    StrictMode               bool
    SkipReadOnlyFields       bool
    SkipWriteOnlyFields      bool
    RequiredFieldsOnly       bool
    CustomConstraints        []models.Key
}
```

## Implementation Details

### Type Validation Implementation

The type validator handles:
- Basic type checking (int, string, bool, time.Time, etc.)
- Pointer vs value type handling
- Slice and map type validation
- Custom type registration

```go
type typeValidatorImpl struct {
    supportedTypes map[reflect.Type]TypeHandler
    typeConverters map[reflect.Type]map[reflect.Type]TypeConverter
    cache          *lru.Cache
}

type TypeHandler interface {
    Validate(value any) error
    Convert(value any, targetType reflect.Type) (any, error)
    GetZeroValue() any
}

type TypeConverter interface {
    Convert(value any) (any, error)
    CanConvert(from, to reflect.Type) bool
}
```

### Constraint Validation Implementation

Builds upon existing `models.ValueValidator` with enhancements:
- Scenario-aware constraint application
- Constraint dependency resolution
- Custom constraint registration
- Constraint caching

```go
type constraintValidatorImpl struct {
    baseValidator   models.ValueValidator
    scenarioRules   map[Scenario][]models.Key
    constraintCache map[string]cachedConstraint
    customHandlers  map[models.Key]models.ValidatorFunc
}

type cachedConstraint struct {
    directives []models.Directive
    expiresAt  time.Time
    scenario   Scenario
}
```

### Database Validation Implementation

Database-specific validation that:
- Validates against database schema
- Handles database type conversions
- Provides database-specific error messages
- Supports multiple database backends

```go
type databaseValidatorImpl struct {
    dbValidators map[string]DatabaseBackendValidator
    schemaCache  map[string]DatabaseSchema
}

type DatabaseBackendValidator interface {
    ValidateField(field models.Field, value any) error
    ConvertValue(value any, field models.Field) (any, error)
    GetErrorMessage(err error, field models.Field) string
}

type DatabaseSchema struct {
    Tables    map[string]TableSchema
    Relations map[string][]Relation
}
```

### Scenario Adaptation Implementation

Orchestrates validation based on operation context:

```go
type scenarioAdapterImpl struct {
    strategies    map[Scenario]ValidationStrategy
    contextCache  map[string]ValidationContext
    scenarioRules map[Scenario]ScenarioRule
}

type ValidationStrategy interface {
    GetValidators() []Validator
    GetValidatorOrder() []ValidatorType
    ShouldValidate(constraint models.Key) bool
    GetErrorHandling() ErrorHandlingStrategy
}

type ScenarioRule struct {
    RequiredConstraints []models.Key
    OptionalConstraints []models.Key
    SkippedConstraints  []models.Key
    StrictMode          bool
}
```

## Integration Points

### Provider Layer Integration

Both local and remote providers will integrate with the validation system:

```go
// Local provider integration
type localProviderWithValidation struct {
    *localProvider
    validationManager ValidationManager
    config            ValidationConfig
}

// Remote provider integration  
type remoteProviderWithValidation struct {
    *remoteProvider
    validationManager ValidationManager
    config            ValidationConfig
}
```

### ORM Layer Integration

ORM operations will use the validation manager:

```go
func (o *ormImpl) Insert(model models.Model) (models.Model, error) {
    // Create validation context
    ctx := validation.NewContext(
        validation.ScenarioInsert,
        validation.OperationCreate,
        model,
        o.provider.GetDatabaseType(),
    )
    
    // Validate before insertion
    if err := o.validationManager.ValidateModel(model, ctx); err != nil {
        return nil, err
    }
    
    // Proceed with insertion
    return o.provider.Insert(model)
}
```

### Model Layer Updates

Models will be enhanced to support the new validation system:

```go
type Model interface {
    // Existing methods...
    
    // New validation methods
    GetValidationConstraints() models.Constraints
    GetFieldValidationConstraints(fieldName string) models.Constraints
    ShouldValidateField(fieldName string, scenario Scenario) bool
}
```

## Performance Considerations

### Caching Strategy

1. **Constraint Cache**: Cache parsed constraints by field and scenario
2. **Type Conversion Cache**: Cache type conversion results
3. **Validation Result Cache**: Cache validation results for immutable data
4. **Schema Cache**: Cache database schema information

### Optimization Techniques

1. **Lazy Loading**: Load validators only when needed
2. **Parallel Validation**: Validate independent fields in parallel
3. **Early Exit**: Stop validation on first error when configured
4. **Batch Processing**: Validate multiple fields/models in batches

## Migration Strategy

### Phase 1: Foundation (Current)
- Create validation directory structure
- Implement core interfaces
- Create basic type validator
- Implement error handling system

### Phase 2: Integration
- Update provider layer to use validation manager
- Integrate with ORM operations
- Add configuration system
- Implement caching layer

### Phase 3: Enhancement
- Add scenario-aware validation
- Implement database-specific validation
- Add performance optimizations
- Create comprehensive test suite

### Phase 4: Migration
- Update existing code to use new validation system
- Create migration guide
- Update documentation and examples
- Deprecate old validation methods

## Testing Strategy

### Unit Tests
- Test each validation layer independently
- Test error handling and edge cases
- Test performance and caching

### Integration Tests
- Test validation with actual database operations
- Test scenario-aware validation
- Test provider integration

### Compatibility Tests
- Ensure backward compatibility
- Test with existing models and constraints
- Verify behavior matches old system

## Configuration Examples

### Basic Configuration
```go
config := validation.NewConfig()
config.EnableTypeValidation = true
config.EnableConstraintValidation = true
config.EnableDatabaseValidation = false
config.EnableScenarioAdaptation = true

// Scenario-specific configuration
config.Scenarios[validation.ScenarioInsert] = validation.ScenarioConfig{
    StrictMode: true,
    SkipReadOnlyFields: false,
    RequiredFieldsOnly: false,
}

config.Scenarios[validation.ScenarioUpdate] = validation.ScenarioConfig{
    StrictMode: false,
    SkipReadOnlyFields: true,
    RequiredFieldsOnly: true,
}
```

### Custom Constraint Registration
```go
validator := validation.NewValidationManager(config)

// Register custom constraint
validator.RegisterConstraint("custom", func(value any, args []string) error {
    // Custom validation logic
    return nil
})

// Register custom type handler
validator.RegisterTypeHandler(reflect.TypeOf(MyCustomType{}), myTypeHandler)
```

## Error Handling Examples

### Basic Error Handling
```go
err := validator.ValidateModel(model, ctx)
if err != nil {
    if validationErr, ok := err.(validation.ValidationError); ok {
        fmt.Printf("Field: %s, Constraint: %s, Value: %v\n",
            validationErr.GetField(),
            validationErr.GetConstraint(),
            validationErr.GetValue())
    }
    return err
}
```

### Error Collection
```go
collector := validation.NewErrorCollector()
ctx := validation.NewContextWithCollector(scenario, collector)

err := validator.ValidateModel(model, ctx)
if collector.HasErrors() {
    for _, err := range collector.GetErrors() {
        // Process each error
    }
}
```

## Performance Benchmarks

Target performance improvements:
- 20% reduction in validation overhead
- 50% reduction in memory allocations
- 30% faster constraint parsing
- 40% faster type conversions

## Monitoring and Metrics

Built-in monitoring capabilities:
- Validation success/failure rates
- Average validation time per layer
- Cache hit/miss ratios
- Most frequent validation errors
- Performance bottlenecks identification

## Future Extensions

### Planned Enhancements
1. **Internationalization**: Support for localized error messages
2. **Async Validation**: Support for asynchronous validation operations
3. **Distributed Validation**: Support for validation in distributed systems
4. **AI-Powered Validation**: Machine learning for constraint optimization
5. **Visual Validation Rules**: GUI for defining and managing validation rules

### Plugin System
Allow third-party validation plugins:
- Custom constraint validators
- Database-specific validators
- Scenario adapters
- Error formatters

## Conclusion

The new validation architecture provides:
1. **Clear separation of concerns** with four distinct layers
2. **Scenario-aware validation** for different operation types
3. **Enhanced error handling** with rich context
4. **Performance optimizations** through caching and lazy loading
5. **Backward compatibility** with existing code
6. **Extensibility** for future enhancements

This architecture addresses the current limitations while providing a solid foundation for future development.