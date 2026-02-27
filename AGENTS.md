# AGENTS.md - magicOrm Development Guide

This document provides essential guidelines for AI agents working on magicOrm, a Go ORM framework supporting PostgreSQL and MySQL.

## Project Overview

- **Language**: Go 1.24+
- **Database Support**: PostgreSQL and MySQL
- **Architecture**: Provider-based (local and remote providers)
- **Dependencies**: Uses `github.com/muidea/magicCommon` as local replace

## Essential Commands

### Build & Test
```bash
# Build project
go build ./...

# Run all tests (PostgreSQL default)
go test ./...

# Run MySQL tests
go test -tags=mysql ./...

# Run specific test
go test -run TestSimpleLocal ./test

# Run verbose tests
go test -v ./...

# Disable test caching
go test -count=1 ./...
```

### Test Scripts
```bash
./local_test.sh      # Local provider tests
./remote_test.sh     # Remote provider tests
./coverage.sh        # Code coverage report
```

### Database Demos
```bash
cd database/postgres/demo && go run main.go
cd database/mysql/demo && go run main.go
```

## Code Style Guidelines

### Import Order
1. Standard library
2. Third-party imports
3. Local project imports
4. `github.com/muidea/magicCommon` imports (as `cd`)

Example:
```go
import (
    "context"
    "fmt"
    "time"

    cd "github.com/muidea/magicCommon/def"
    "github.com/muidea/magicOrm/models"
)
```

### Naming Conventions
- **Packages**: lowercase single words (`local`, `remote`, `models`)
- **Interfaces**: `er` suffix when appropriate (`Provider`, `Executor`)
- **Types**: PascalCase (`TypeImpl`, `ValueImpl`)
- **Variables**: camelCase (`localProvider`, `entityModel`)
- **Constants**: UPPER_SNAKE_CASE (`SIMPLE_LOCAL_OWNER`)

### Error Handling
Use custom error type from `github.com/muidea/magicCommon/def` (imported as `cd`):
```go
func GetEntityType(entity any) (ret models.Type, err *cd.Error) {
    if entity == nil {
        err = cd.NewError(cd.IllegalParam, "entity is invalid")
        return
    }
    // implementation
}
```

### Model Definitions
- Struct tags: `orm:` prefix for fields
- View tags: `view:` for view mode declarations
- Constraint tags: `constraint:` for validation rules
- Use pointers for optional fields (nullable columns)

Example:
```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`
    Name   string   `orm:"name" constraint:"req,min=3,max=50" view:"detail,lite"`
    EMail  string   `orm:"email" view:"detail,lite"`
    Status *Status  `orm:"status" view:"detail,lite"`
}
```

### Testing Conventions
- Test files: `_test.go` suffix
- Test functions: `Test` prefix
- Use table-driven tests for multiple cases
- Setup/teardown: `orm.Initialize()` and `orm.Uninitialized()`
- Test both local and remote providers

## Development Workflow

### Adding Features
1. Understand provider architecture (local vs remote)
2. Follow existing patterns in codebase
3. Add comprehensive tests (unit + integration)
4. Test with both PostgreSQL and MySQL
5. Update documentation if needed

### Database Compatibility
- All features must work with both PostgreSQL and MySQL
- Use database-agnostic SQL when possible
- For database-specific features, use build tags
- Test both database configurations

### Constraint System
- Access constraints: `req` (required), `ro` (read-only), `wo` (write-only)
- Value constraints: `min`, `max`, `range`, `in`, `re` (regex)
- Defined in struct tags
- Validation at model and database levels

## Validation System

### Basic Usage
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

### Scenarios
- `ScenarioInsert`: Strict validation
- `ScenarioUpdate`: Skip read-only fields
- `ScenarioQuery`: Skip write-only fields
- `ScenarioDelete`: Minimal validation

## Troubleshooting

### Common Issues
1. **Database connection failures**: Check database service
2. **Test tag issues**: Use correct build tags (`-tags=mysql`)
3. **Type conversion errors**: Check pointer vs value types
4. **Constraint validation failures**: Verify struct tag definitions

### Debugging Tips
- Enable verbose tests: `go test -v`
- Check database logs for SQL errors
- Use demo applications for specific scenarios
- Review existing test cases for patterns

## References
- [README.md](./README.md) - Comprehensive documentation
- [test/](./test/) - Integration tests and examples
- [database/](./database/) - Database implementations