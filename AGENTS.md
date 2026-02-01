# AGENTS.md - magicOrm Development Guide

This document provides guidelines for AI agents working on the magicOrm project, a Go ORM framework supporting PostgreSQL and MySQL databases.

## Project Overview

magicOrm is a Go ORM framework with the following characteristics:
- **Language**: Go 1.24+
- **Database Support**: PostgreSQL and MySQL
- **Architecture**: Provider-based (local and remote providers)
- **Testing**: Comprehensive test suite with both local and remote test modes
- **Dependencies**: Uses `github.com/muidea/magicCommon` as a local replace dependency

## Build and Development Commands

### Basic Go Commands
```bash
# Build the project
go build ./...

# Run all tests (default uses PostgreSQL)
go test ./...

# Run tests with MySQL database
go test -tags=mysql ./...

# Run specific test package
go test ./provider/local

# Run single test
go test -run TestSimpleLocal ./test

# Run tests with verbose output
go test -v ./...

# Run tests once (disable test caching)
go test -count=1 ./...
```

### Test Scripts
The project includes convenience scripts for testing:
```bash
# Run local provider tests
./local_test.sh

# Run remote provider tests  
./remote_test.sh

# Generate code coverage report
./coverage.sh
```

### Database-Specific Testing
```bash
# PostgreSQL tests (default)
go test ./...

# MySQL tests
go test -tags=mysql ./...

# Both database tests with build tags
go test -tags="local mysql" ./test/...
```

### Demo Applications
```bash
# Run PostgreSQL demo
cd database/postgres/demo && go run main.go

# Run MySQL demo  
cd database/mysql/demo && go run main.go
```

## Code Style Guidelines

### Package Organization
- **Root package**: `github.com/muidea/magicOrm`
- **Provider packages**: `provider/local`, `provider/remote`
- **Model packages**: `models`, `orm`
- **Utility packages**: `utils`, `database`
- **Test packages**: `test` (integration tests)

### Import Order
Follow standard Go import grouping:
1. Standard library imports
2. Third-party imports  
3. Local project imports
4. `github.com/muidea/magicCommon` imports (local replace)

Example:
```go
import (
    "context"
    "fmt"
    "reflect"
    "time"

    cd "github.com/muidea/magicCommon/def"
    
    "github.com/muidea/magicOrm/models"
    "github.com/muidea/magicOrm/orm"
)
```

### Naming Conventions
- **Packages**: Use lowercase, single-word names (e.g., `local`, `remote`, `models`)
- **Interfaces**: Use `er` suffix when appropriate (e.g., `Provider`, `Executor`)
- **Types**: Use PascalCase (e.g., `TypeImpl`, `ValueImpl`)
- **Variables**: Use camelCase (e.g., `localProvider`, `entityModel`)
- **Constants**: Use UPPER_SNAKE_CASE (e.g., `SIMPLE_LOCAL_OWNER`)

### Error Handling
- Use custom error type from `github.com/muidea/magicCommon/def` (imported as `cd`)
- Always check and return errors immediately
- Use descriptive error messages with context
- Log errors using the project's logging system

Example:
```go
func GetEntityType(entity any) (ret models.Type, err *cd.Error) {
    if entity == nil {
        err = cd.NewError(cd.IllegalParam, "entity is invalid")
        return
    }
    // ... implementation
}
```

### Type Definitions and Models
- Struct tags use `orm:` prefix for field definitions
- Support `view:` tag for view mode declarations
- Support `constraint:` tag for validation rules
- Use pointer types for optional fields (nullable database columns)

Example model definition:
```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`
    Name   string   `orm:"name" constraint:"req,min=3,max=50" view:"detail,lite"`
    EMail  string   `orm:"email" view:"detail,lite"`
    Status *Status  `orm:"status" view:"detail,lite"`
    Group  []*Group `orm:"group" view:"detail,lite"`
}
```

### Testing Conventions
- Test files use `_test.go` suffix
- Test functions start with `Test` prefix
- Use table-driven tests for multiple test cases
- Setup and teardown using `orm.Initialize()` and `orm.Uninitialized()`
- Test both local and remote providers

Example test structure:
```go
func TestSimpleLocal(t *testing.T) {
    orm.Initialize()
    defer orm.Uninitialized()
    
    localProvider := provider.NewLocalProvider("testOwner", nil)
    // ... test implementation
}
```

## Development Workflow

### Adding New Features
1. **Understand the provider architecture**: Local vs remote providers
2. **Follow existing patterns**: Check similar implementations in the codebase
3. **Add comprehensive tests**: Include both unit and integration tests
4. **Test with both databases**: Ensure compatibility with PostgreSQL and MySQL
5. **Update documentation**: Add examples to README.md if applicable

### Database Compatibility
- All features must work with both PostgreSQL and MySQL
- Use database-agnostic SQL when possible
- For database-specific features, use build tags or provider-specific implementations
- Test both database configurations

### Provider Implementation
- **Local provider**: Direct database access
- **Remote provider**: Network-based access
- Maintain interface compatibility between providers
- Use helper functions in `provider/helper` for common functionality

### Constraint System
The project includes a sophisticated constraint system:
- Access behavior constraints: `req` (required), `ro` (read-only), `wo` (write-only)
- Content value constraints: `min`, `max`, `range`, `in`, `re` (regex)
- Constraints are defined in struct tags
- Validation happens at both model and database levels

## CI/CD Pipeline

### GitHub Actions
The project uses GitHub Actions for CI:
- Automated testing with PostgreSQL and MySQL
- Docker build and release workflows
- Tests run on push and manual trigger

### Running Tests Locally
Before committing changes:
1. Run local provider tests: `./local_test.sh`
2. Run remote provider tests: `./remote_test.sh`
3. Run database-specific tests if applicable
4. Ensure all tests pass with both PostgreSQL and MySQL

## Common Patterns

### Model Registration
```go
entityList := []any{&User{}, &Status{}, &Group{}}
modelList, err := registerLocalModel(localProvider, entityList)
```

### ORM Operations
```go
// Create ORM instance
o1, err := orm.NewOrm(localProvider, config, "schema_prefix")
defer o1.Release()

// CRUD operations
userModel, err := o1.Insert(userModel)
userModel, err = o1.Query(userModel)
userModel, err = o1.Update(userModel)
_, err = o1.Delete(userModel)
```

### Transaction Management
```go
tx, err := o1.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// Perform operations within transaction
userModel, err := tx.Insert(userModel)
if err != nil {
    return err
}

err = tx.Commit()
```

## Validation System

### Overview
MagicORM includes a comprehensive validation system with four-layer architecture:

1. **Type Validation**: Basic type compatibility and conversion
2. **Constraint Validation**: Business rule validation from struct tags
3. **Database Validation**: Database-specific constraint validation  
4. **Scenario Adaptation**: Scenario-aware validation orchestration

### Key Components
- **Validation Manager**: Central coordinator (`validation.NewValidationManager()`)
- **Error Handling**: Enhanced error collection (`validation/errors/`)
- **Configuration**: Flexible validation configuration (`validation.ValidationConfig`)

### Usage Examples

#### Basic Validation
```go
import "github.com/muidea/magicOrm/validation"

config := validation.DefaultConfig()
manager := validation.NewValidationManager(config)

ctx := validation.NewContext(
    validation.ScenarioInsert,
    validation.OperationCreate,
    model,
    "postgresql",
)

err := manager.ValidateModel(model, ctx)
```

#### Error Collection
```go
import "github.com/muidea/magicOrm/validation/errors"

collector := errors.NewErrorCollector()
ctx := validation.NewContextWithCollector(
    validation.ScenarioInsert,
    collector,
)

// Validate and collect errors
if collector.HasErrors() {
    for _, err := range collector.GetErrors() {
        fmt.Printf("Field: %s, Error: %s\n", err.GetField(), err.Error())
    }
}
```

#### Scenario-Aware Validation
```go
// Different validation for different operations
scenarios := []validation.Scenario{
    validation.ScenarioInsert,  // Strict validation
    validation.ScenarioUpdate,  // Skip read-only fields  
    validation.ScenarioQuery,   // Skip write-only fields
    validation.ScenarioDelete,  // Minimal validation
}

for _, scenario := range scenarios {
    ctx := validation.NewContext(scenario, ...)
    err := manager.ValidateModel(model, ctx)
}
```

### Configuration Options
```go
config := validation.ValidationConfig{
    EnableTypeValidation:      true,
    EnableConstraintValidation: true,
    EnableDatabaseValidation:  true,
    EnableScenarioAdaptation:  true,
    EnableCaching:            true,
    CacheTTL:                 5 * time.Minute,
    DefaultOptions: validation.ValidationOptions{
        StopOnFirstError:        false,
        ValidateReadOnlyFields:  true,
        ValidateWriteOnlyFields: true,
    },
}
```

### Customization
```go
// Register custom constraint
manager.RegisterCustomConstraint("custom", func(value any, args []string) error {
    // Custom validation logic
    return nil
})

// Register custom type handler  
manager.RegisterTypeHandler("MyType", myTypeHandler)
```

### Testing Validation
```bash
# Run validation tests
cd validation && go test ./test/... -v

# Run example
cd validation/example && go run simple_example.go
```

### References
- [Validation Architecture](./VALIDATION_ARCHITECTURE.md)
- [Validation Implementation Plan](./VALIDATION_IMPLEMENTATION_PLAN.md)
- [Validation README](./validation/README.md)

## Troubleshooting

### Common Issues
1. **Database connection failures**: Check database service is running
2. **Test tag issues**: Use correct build tags for database-specific tests
3. **Type conversion errors**: Ensure proper use of pointer vs value types
4. **Constraint validation failures**: Check struct tag definitions
5. **Validation system errors**: Check validation configuration and error messages

### Debugging Tips
- Enable verbose test output: `go test -v`
- Check database logs for SQL errors
- Use the demo applications to test specific scenarios
- Review existing test cases for implementation patterns

## References

- [README.md](./README.md) - Comprehensive documentation and examples
- [test/](./test/) - Integration tests and usage examples
- [database/](./database/) - Database-specific implementations and demos
- [.github/workflows/](./.github/workflows/) - CI/CD configuration