# AGENTS.md - magicOrm Development Guide

This document provides essential guidelines for AI agents working on magicOrm, a Go ORM framework supporting PostgreSQL and MySQL.

## Model and Workflow Boundaries

For model, DTO, lifecycle, recovery or cleanup changes, read [Workspace Boundary Rules](~/dataspace/magicPlatform/docs/technical/model-state-boundary-rules.md) and use its change-review template.

- Preserve generic identity, field, relation, validation and transaction semantics for all providers. Do not special-case platform names, tenant names, lifecycle stages, Schema journals or compensation policy.
- Business owners define independent workflow models and safety fences. Do not weaken required/default/unique/nullability rules, change ID allocation, or turn an update into an insert to bypass a platform failure; generic changes need independent provider regressions.
- Change source owners first; never edit vendor manually. Sync direct dependents through the workspace `upgrade_vendor.sh`, update affected contracts and product materials, and record source regression separately from Live acceptance.

## Project Overview

- **Language**: Go 1.26+
- **Database Support**: PostgreSQL and MySQL
- **Architecture**: Provider-based (local and remote providers)
- **Dependencies**: Uses `github.com/muidea/magicCommon` as local replace

## Essential Commands

### Build & Test
```bash
# Build project
go build ./...

# Run all tests (PostgreSQL default)
go test ./... # offline by default; integration tests skip

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
cd internal/database/postgres/demo && go run main.go
cd internal/database/mysql/demo && go run main.go
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

### Public API Compatibility

- Prefer `orm` / `models` / root `provider` for application code. Dynamic declarations belong to `provider/remote`; helper, validation and metrics packages are optional integrations.
- Local and Remote are peer public adapters: retain both package paths, constructors, model/value/filter/codec capabilities, concrete type identities and JSON contracts. The reviewed support-package migration moves validators, caches and collectors to internal; their contracts must remain usable through public models/validation/metrics APIs. Do not expose internal types in public signatures.
- Contracts and implementations may be organized into separate files in the same package. Add no compatibility forwarding layer unless an actual boundary requires one.
- Run `go test ./internal/apicheck ./provider/...` after public API changes. Review the compatibility impact before updating the snapshot with `go test ./internal/apicheck -update-api`.
- Do not export production setters for test state. Exercise normal registration through its public contract and unregistered startup in a separate test process.

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
import (
 "github.com/muidea/magicOrm/validation"
 verrors "github.com/muidea/magicOrm/validation/errors"
)

config := validation.DefaultConfig()
manager := validation.NewValidationManager(config)

ctx := validation.NewContext(
    verrors.ScenarioInsert,
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

## Metrics Integration

- Preferred initialization order: call `monitoring.InitializeGlobalManager()` first, then `orm.Initialize()`. Standalone validation/metrics users may call `metrics.Initialize()`; both initialization paths reuse existing collectors.
- If monitoring is initialized later, call public `metrics.EnsureRegistered()` after the global manager becomes available. The legacy `orm.EnsureORMMetricProviderRegistered()` still registers the ORM provider only.
- Current maturity: ORM metrics are wired to the ORM main path, DB metrics are wired to MySQL/PostgreSQL executors and pools, Validation metrics are wired to `validation.Manager`, validation cache, and constraint checks.
- Internal metric definitions should use `internal/metrics/shared.DefaultLabels()` merged with component-specific labels so dashboards can group all MagicORM providers consistently.

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
- [README.md](~/codespace/magicOrm/README.md) - Comprehensive documentation
- [test/](~/codespace/magicOrm/test) - Integration tests and examples
- [internal/database/](~/codespace/magicOrm/internal/database) - Database implementations

Real database tests require explicit disposable magicorm_test_ targets; run `~/codespace/magicOrm/scripts/integration_postgres.sh` for an owned PostgreSQL fixture. Never use shared/default databases or replication/GRANT tests.
