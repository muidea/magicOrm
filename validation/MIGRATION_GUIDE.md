# MagicORM 验证系统迁移指南

## 概述

本文档提供从旧验证系统迁移到新四层验证架构的完整指南。新验证系统提供更好的性能、更丰富的错误处理和场景感知验证。

## 迁移前准备

### 1. 评估当前使用情况

检查项目中验证系统的使用模式：

```go
// 检查当前验证调用
grep -r "Validate\|validation" --include="*.go" . | grep -v test | grep -v vendor

// 检查约束定义
grep -r "constraint:" --include="*.go" .

// 检查模型定义
grep -r "orm:" --include="*.go" . | head -20
```

### 2. 备份当前代码

```bash
# 创建备份分支
git checkout -b backup-validation-system-$(date +%Y%m%d)

# 备份相关文件
cp -r validation/ validation_backup/
cp orm/orm.go orm/orm_backup.go
cp provider/local/validation_ext.go provider/local/validation_ext_backup.go
```

### 3. 运行现有测试

确保现有测试通过：

```bash
# 运行所有验证相关测试
go test ./validation/... -v
go test ./orm/... -v
go test ./provider/... -v
```

## 迁移步骤

### 步骤1：更新依赖

更新 `go.mod` 文件（如果需要）：

```bash
go mod tidy
```

### 步骤2：更新模型定义（可选）

新验证系统完全兼容现有模型定义，但可以优化：

```go
// 旧定义（仍然兼容）
type User struct {
    ID     int      `orm:"uid key auto"`
    Name   string   `orm:"name" constraint:"req,min=3,max=50"`
    Email  string   `orm:"email" constraint:"email"`
}

// 新定义（支持更多约束类型）
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`
    Name   string   `orm:"name" constraint:"req,min=3,max=50" view:"detail,lite"`
    Email  string   `orm:"email" constraint:"req,email,unique" view:"detail,lite"`
    Status *Status  `orm:"status" constraint:"in:active,inactive,suspended" view:"detail,lite"`
}
```

### 步骤3：更新验证调用

#### 3.1 基本验证迁移

```go
// 旧方式（直接调用）
err := provider.ValidateModel(model)
if err != nil {
    return err
}

// 新方式（通过验证管理器）
import "github.com/muidea/magicOrm/validation"

// 创建验证管理器
config := validation.DefaultConfig()
factory := validation.NewValidationFactory()
manager := factory.CreateValidationManager(config)

// 执行验证
ctx := validation.NewContext(
    validation.ScenarioInsert,
    validation.OperationCreate,
    model,
    "postgresql", // 或 "mysql"
)

err := manager.ValidateModel(model, ctx)
if err != nil {
    // 错误处理
    if collector, ok := err.(validation.ErrorCollector); ok {
        for _, e := range collector.GetErrors() {
            fmt.Printf("Field: %s, Error: %s\n", e.GetField(), e.Error())
        }
    }
    return err
}
```

#### 3.2 Provider层验证迁移

```go
// 旧方式（Provider直接验证）
func (p *LocalProvider) Insert(model models.Model) (models.Model, error) {
    // 验证逻辑
    if err := p.validateModel(model, "insert"); err != nil {
        return nil, err
    }
    // 插入逻辑
}

// 新方式（使用扩展方法）
import "github.com/muidea/magicOrm/validation"

func (p *LocalProvider) Insert(model models.Model) (models.Model, error) {
    // 使用场景感知验证
    err := p.ValidateModelForInsert(model)
    if err != nil {
        return nil, err
    }
    // 插入逻辑
}
```

### 步骤4：更新错误处理

```go
// 旧错误处理
if err != nil {
    log.Printf("Validation error: %v", err)
    return err
}

// 新错误处理（支持丰富错误信息）
if err != nil {
    // 检查错误类型
    switch e := err.(type) {
    case *validation.TypeError:
        log.Printf("Type validation failed: %s (field: %s)", e.Message, e.Field)
    case *validation.ConstraintError:
        log.Printf("Constraint validation failed: %s (constraint: %s)", e.Message, e.Constraint)
    case validation.ErrorCollector:
        log.Printf("Multiple validation errors:")
        for _, err := range e.GetErrors() {
            log.Printf("  - %s", err.Error())
        }
    default:
        log.Printf("Validation error: %v", err)
    }
    return err
}
```

### 步骤5：配置验证系统

```go
// 基本配置
config := validation.ValidationConfig{
    EnableTypeValidation:      true,
    EnableConstraintValidation: true,
    EnableDatabaseValidation:  true,
    EnableScenarioAdaptation:  true,
    EnableCaching:            true,
    CacheTTL:                 5 * time.Minute,
}

// 根据环境调整配置
func GetValidationConfig(env string) validation.ValidationConfig {
    switch env {
    case "development":
        return validation.ValidationConfig{
            EnableTypeValidation:      true,
            EnableConstraintValidation: true,
            EnableDatabaseValidation:  false, // 开发环境可关闭
            EnableScenarioAdaptation:  true,
            EnableCaching:            false,  // 开发环境关闭缓存
            StopOnFirstError:        true,    // 快速失败
        }
    case "production":
        return validation.ValidationConfig{
            EnableTypeValidation:      true,
            EnableConstraintValidation: true,
            EnableDatabaseValidation:  true,
            EnableScenarioAdaptation:  true,
            EnableCaching:            true,
            CacheTTL:                 10 * time.Minute,
            StopOnFirstError:        false,   // 收集所有错误
            MaxConcurrentValidations: 100,    // 限制并发
        }
    default:
        return validation.DefaultConfig()
    }
}
```

## 迁移检查清单

### 代码变更检查

- [ ] 更新所有验证调用使用新API
- [ ] 更新错误处理逻辑
- [ ] 配置验证管理器
- [ ] 更新测试用例
- [ ] 检查模型约束定义

### 功能验证检查

- [ ] 基本CRUD操作验证正常
- [ ] 约束验证正常工作
- [ ] 场景感知验证（Insert/Update/Query/Delete）
- [ ] 错误信息包含足够上下文
- [ ] 缓存功能正常工作

### 性能检查

- [ ] 验证性能无显著下降
- [ ] 缓存命中率合理
- [ ] 内存使用正常
- [ ] 并发验证正常工作

## 向后兼容性

### 完全兼容的特性

1. **模型定义**：现有`orm:`和`constraint:`标签完全兼容
2. **基本验证**：类型验证和约束验证行为一致
3. **错误代码**：错误类型和代码保持兼容

### 行为变更

1. **错误收集**：默认收集所有错误而非在第一个错误停止
2. **场景感知**：不同操作（Insert/Update）应用不同验证规则
3. **缓存行为**：验证结果可能被缓存

### 配置默认值变更

| 配置项 | 旧默认值 | 新默认值 | 影响 |
|--------|----------|----------|------|
| 停止首个错误 | 是 | 否 | 可能看到更多错误 |
| 数据库验证 | 开启 | 开启 | 无变化 |
| 缓存 | 无 | 开启 | 性能提升 |
| 场景适配 | 无 | 开启 | 更智能的验证 |

## 故障排除

### 常见迁移问题

#### 问题1：验证错误数量增加
**症状**：迁移后看到更多验证错误
**原因**：新系统默认收集所有错误而非在第一个错误停止
**解决方案**：
```go
config := validation.ValidationConfig{
    StopOnFirstError: true, // 恢复旧行为
}
```

#### 问题2：性能下降
**症状**：验证操作变慢
**原因**：可能未启用缓存或配置不当
**解决方案**：
```go
config := validation.ValidationConfig{
    EnableCaching: true,
    CacheTTL:      5 * time.Minute,
    CacheMaxSize:  1000,
}
```

#### 问题3：场景验证行为异常
**症状**：Update操作验证了只读字段
**原因**：场景适配未正确配置
**解决方案**：
```go
// 确保使用正确的场景
ctx := validation.NewContext(
    validation.ScenarioUpdate, // 使用Update场景
    validation.OperationUpdate,
    model,
    dbType,
)
```

#### 问题4：数据库验证失败
**症状**：数据库约束验证错误
**原因**：数据库验证需要数据库连接
**解决方案**：
```go
// 临时禁用数据库验证
config.EnableDatabaseValidation = false

// 或提供数据库连接
ctx.SetDatabaseConnection(dbConn)
```

### 调试步骤

1. **启用详细日志**：
```go
logger := validation.NewLogger("debug")
config.Logger = logger
```

2. **检查验证配置**：
```go
fmt.Printf("Config: %+v\n", manager.GetConfig())
```

3. **验证单个字段**：
```go
// 隔离问题字段
field := model.GetField("email")
err := manager.ValidateField(field, ctx)
```

4. **检查缓存状态**：
```go
stats := manager.GetCacheStats()
fmt.Printf("Cache hit rate: %.2f%%\n", stats.HitRate*100)
```

## 回滚指南

### 紧急回滚步骤

如果迁移后遇到严重问题，可以快速回滚：

```bash
# 1. 停止应用
# 2. 恢复备份文件
cp validation_backup/* validation/
cp orm/orm_backup.go orm/orm.go
cp provider/local/validation_ext_backup.go provider/local/validation_ext.go

# 3. 清理新文件
rm -rf validation/monitoring/
rm validation/PRODUCTION_GUIDE.md
rm validation/MIGRATION_GUIDE.md

# 4. 恢复依赖
go mod tidy

# 5. 重启应用
```

### 部分回滚

如果只有部分功能有问题，可以部分回滚：

```go
// 临时禁用新功能
config := validation.ValidationConfig{
    EnableScenarioAdaptation: false, // 禁用场景适配
    EnableCaching:           false,  // 禁用缓存
    StopOnFirstError:       true,    // 恢复旧错误行为
}
```

## 最佳实践

### 迁移策略

1. **分阶段迁移**：
   - 阶段1：更新测试环境
   - 阶段2：更新预生产环境
   - 阶段3：更新生产环境

2. **A/B测试**：
   ```go
   // 可以同时运行新旧验证系统比较结果
   oldResult := oldValidate(model)
   newResult := newValidate(model)
   compareResults(oldResult, newResult)
   ```

3. **监控迁移**：
   - 监控错误率变化
   - 监控性能指标
   - 监控内存使用

### 配置管理

1. **环境特定配置**：
   ```go
   // 使用环境变量
   config.EnableCaching = os.Getenv("VALIDATION_CACHE") == "true"
   config.CacheTTL = getEnvDuration("VALIDATION_CACHE_TTL", 5*time.Minute)
   ```

2. **动态配置**：
   ```go
   // 支持运行时配置更新
   manager.UpdateConfig(newConfig)
   ```

3. **配置验证**：
   ```go
   // 验证配置有效性
   if err := config.Validate(); err != nil {
       return fmt.Errorf("invalid validation config: %w", err)
   }
   ```

### 测试策略

1. **迁移测试**：
   ```go
   func TestMigrationCompatibility(t *testing.T) {
       // 测试新旧系统行为一致
       testCases := []struct{
           model models.Model
           scenario string
       }{
           // 测试用例
       }
       
       for _, tc := range testCases {
           t.Run(tc.scenario, func(t *testing.T) {
               oldErr := oldValidate(tc.model, tc.scenario)
               newErr := newValidate(tc.model, tc.scenario)
               
               // 验证错误一致
               assertErrorsEqual(t, oldErr, newErr)
           })
       }
   }
   ```

2. **性能测试**：
   ```go
   func BenchmarkValidationMigration(b *testing.B) {
       // 基准测试新旧系统性能
       b.Run("OldSystem", func(b *testing.B) {
           for i := 0; i < b.N; i++ {
               oldValidate(benchmarkModel)
           }
       })
       
       b.Run("NewSystem", func(b *testing.B) {
           for i := 0; i < b.N; i++ {
               newValidate(benchmarkModel)
           }
       })
   }
   ```

## 高级迁移场景

### 自定义约束迁移

```go
// 旧自定义约束
type CustomConstraint struct {
    // 旧实现
}

// 新自定义约束
func RegisterCustomConstraints(manager validation.ValidationManager) {
    manager.RegisterCustomConstraint("custom", func(value any, args []string) error {
        // 新实现
        return nil
    })
}
```

### 批量操作迁移

```go
// 旧批量验证
func ValidateBatch(models []models.Model) []error {
    errors := make([]error, len(models))
    for i, model := range models {
        errors[i] = oldValidate(model)
    }
    return errors
}

// 新批量验证（支持并发）
func ValidateBatch(models []models.Model, manager validation.ValidationManager) []error {
    errors := make([]error, len(models))
    var wg sync.WaitGroup
    
    ctx := validation.NewContext(...)
    
    for i, model := range models {
        wg.Add(1)
        go func(idx int, m models.Model) {
            defer wg.Done()
            errors[idx] = manager.ValidateModel(m, ctx)
        }(i, model)
    }
    
    wg.Wait()
    return errors
}
```

### 异步验证迁移

```go
// 旧同步验证
err := validate(model)
if err != nil {
    return err
}

// 新异步验证
func ValidateAsync(model models.Model, manager validation.ValidationManager) <-chan error {
    result := make(chan error, 1)
    
    go func() {
        err := manager.ValidateModel(model, ctx)
        result <- err
        close(result)
    }()
    
    return result
}

// 使用
validationResult := ValidateAsync(model, manager)
// 继续其它工作
err := <-validationResult
```

## 监控和告警

### 迁移期间监控

1. **关键指标**：
   - 验证错误率
   - 平均验证时间
   - 缓存命中率
   - 内存使用

2. **告警规则**：
   ```yaml
   # 迁移期间特殊告警
   - alert: MigrationValidationErrorSpike
     expr: increase(validation_errors_total[5m]) > 100
     for: 2m
     labels:
       severity: critical
     annotations:
       summary: "Validation errors spiked during migration"
   ```

### 迁移后验证

迁移完成后，运行完整验证：

```bash
# 1. 运行所有测试
go test ./... -v

# 2. 性能基准测试
go test ./validation/... -bench=. -benchtime=5s

# 3. 内存泄漏测试
go test ./validation/... -race -count=100

# 4. 集成测试
./local_test.sh
./remote_test.sh
```

## 支持资源

### 文档
- [生产环境指南](./PRODUCTION_GUIDE.md)
- [监控模块文档](./monitoring/README.md)
- [验证系统架构](./VALIDATION_ARCHITECTURE.md)

### 示例代码
- [基本使用示例](./example/usage_example.go)
- [配置示例](./example/configuration_example.go)
- [监控示例](./monitoring/example.go)

### 工具
- 迁移检查脚本
- 配置验证工具
- 性能比较工具

## 获取帮助

如遇迁移问题，请：

1. **检查日志**：启用调试日志查看详细错误
2. **简化测试**：创建最小复现示例
3. **联系支持**：提供以下信息：
   - 错误信息和堆栈跟踪
   - 相关配置
   - 复现步骤
   - 环境信息

### 紧急联系方式

- **问题跟踪**：GitHub Issues
- **文档**：本项目README和指南
- **社区**：项目讨论区

## 迁移成功标准

完成迁移后，应满足以下标准：

### 功能标准
- [ ] 所有现有功能正常工作
- [ ] 错误处理符合预期
- [ ] 性能达到或超过旧系统
- [ ] 新功能（场景感知、缓存）正常工作

### 质量标准
- [ ] 所有测试通过
- [ ] 代码覆盖率保持或提高
- [ ] 文档完整且准确
- [ ] 监控指标正常

### 运营标准
- [ ] 生产环境运行稳定
- [ ] 监控告警配置正确
- [ ] 回滚计划就绪
- [ ] 团队培训完成

---

**迁移完成时间估计**：
- 小型项目：1-2天
- 中型项目：3-5天  
- 大型项目：1-2周

**建议迁移窗口**：低流量时段，有完整回滚计划。