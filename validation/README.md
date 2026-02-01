# MagicORM Validation System

## Overview

This is the new validation system for MagicORM, implementing a four-layer validation architecture to address current architectural limitations.

## Architecture

The validation system is organized into four layers:

1. **Type Validation** (`validation/types/`): Basic type compatibility and conversion
2. **Constraint Validation** (`validation/constraints/`): Business rule validation from struct tags
3. **Database Validation** (`validation/database/`): Database-specific constraint validation
4. **Scenario Adaptation** (`validation/scenario/`): Scenario-aware validation orchestration

## Core Components

### Validation Manager

The central coordinator that orchestrates validation across all layers:

```go
import "github.com/muidea/magicOrm/validation"

// Create validation manager
config := validation.DefaultConfig()
manager := validation.NewValidationManager(config)

// Validate a value
ctx := validation.NewContext(
    validation.ScenarioInsert,
    validation.OperationCreate,
    model,
    "postgresql",
)
err := manager.Validate(value, ctx)
```

### Error Handling

Enhanced error handling with rich context:

```go
import "github.com/muidea/magicOrm/validation/errors"

// Create error collector
collector := errors.NewErrorCollector()

// Create context with collector
ctx := validation.NewContextWithCollector(
    validation.ScenarioInsert,
    collector,
)

// Validate and collect errors
err := manager.Validate(value, ctx)
if collector.HasErrors() {
    for _, err := range collector.GetErrors() {
        fmt.Printf("Field: %s, Constraint: %s\n", 
            err.GetField(), err.GetConstraint())
    }
}
```

## Configuration

### Default Configuration

```go
config := validation.DefaultConfig()
```

### Simple Configuration

```go
config := validation.SimpleConfig()
```

### Custom Configuration

```go
config := validation.ValidationConfig{
    EnableTypeValidation:      true,
    EnableConstraintValidation: true,
    EnableDatabaseValidation:  false,
    EnableScenarioAdaptation:  true,
    EnableCaching:            true,
    CacheTTL:                 5 * time.Minute,
    MaxCacheSize:             1000,
    DefaultOptions: validation.ValidationOptions{
        StopOnFirstError:        false,
        IncludeFieldPathInError: true,
        ValidateReadOnlyFields:  true,
        ValidateWriteOnlyFields: true,
    },
}
```

## Usage Examples

### Basic Validation

```go
// Create validation manager
config := validation.DefaultConfig()
manager := validation.NewValidationManager(config)

// Validate a model
ctx := validation.NewContext(
    validation.ScenarioInsert,
    validation.OperationCreate,
    model,
    "postgresql",
)

err := manager.ValidateModel(model, ctx)
if err != nil {
    // Handle validation error
    fmt.Printf("Validation failed: %v\n", err)
}
```

### Field-Level Validation

```go
// Get field from model
field := model.GetField("email")

// Validate field value
ctx := validation.NewContext(
    validation.ScenarioUpdate,
    validation.OperationUpdate,
    model,
    "mysql",
)
ctx.Field = field

err := manager.Validate("user@example.com", ctx)
if err != nil {
    // Handle field validation error
}
```

### Scenario-Aware Validation

```go
// Different validation for different scenarios
scenarios := []validation.Scenario{
    validation.ScenarioInsert,
    validation.ScenarioUpdate,
    validation.ScenarioQuery,
}

for _, scenario := range scenarios {
    ctx := validation.NewContext(
        scenario,
        validation.OperationCreate,
        model,
        "postgresql",
    )
    
    err := manager.ValidateModel(model, ctx)
    fmt.Printf("Scenario %s: %v\n", scenario, err)
}
```

## Customization

### Register Custom Constraints

```go
// Register custom constraint
manager.RegisterCustomConstraint("custom", func(value any, args []string) error {
    // Custom validation logic
    return nil
})
```

### Register Custom Type Handlers

```go
// Register custom type handler
typeHandler := &MyTypeHandler{}
manager.RegisterTypeHandler("MyType", typeHandler)
```

### Custom Validation Strategies

```go
import "github.com/muidea/magicOrm/validation/scenario"

// Create custom strategy
strategy := scenario.CreateCustomStrategy(
    []scenario.Validator{...},
    []scenario.ValidatorType{...},
    func(constraint models.Key) bool {
        // Custom constraint filter
        return true
    },
    scenario.ErrorHandlingStrategy{...},
    false, // skipReadOnly
    false, // skipWriteOnly
    true,  // strict
)

// Register custom strategy
adapter := scenario.NewScenarioAdapter()
adapter.RegisterCustomStrategy(validation.ScenarioInsert, strategy)
```

## Integration with Existing Code

The validation system is designed to integrate seamlessly with existing MagicORM code:

1. **Backward Compatibility**: Existing validation logic continues to work
2. **Gradual Migration**: Can be enabled/disabled via configuration
3. **Performance Optimized**: Caching and lazy loading for minimal overhead
4. **Consistent Behavior**: Same validation rules across local and remote providers

## Testing

Run validation tests:

```bash
cd validation
go test ./test/... -v
```

## Next Steps

1. **Integration**: Update provider layer to use validation manager
2. **Optimization**: Implement caching and performance improvements
3. **Documentation**: Add comprehensive usage examples
4. **Migration**: Create migration guide for existing code

## References

- [Validation Architecture](./VALIDATION_ARCHITECTURE.md)
- [Implementation Plan](../VALIDATION_IMPLEMENTATION_PLAN.md)
- [AGENTS.md](../AGENTS.md) - Development guidelines