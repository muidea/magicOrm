# magicOrm

Golang对象的ORM框架，支持PostgreSQL和MySQL数据库。一个所见即所得的ORM框架。

## 特性

- **多数据库支持**: PostgreSQL和MySQL
- **类型安全**: 完整的Go类型映射支持
- **灵活的关系映射**: 支持一对一、一对多、多对多关系
- **强大的约束系统**: 内置数据验证和业务规则
- **四层验证架构**: 类型、约束、数据库、场景分层验证
- **场景感知验证**: 支持Insert/Update/Query/Delete不同策略
- **视图模式**: 支持detail/lite视图控制字段输出
- **事务支持**: 完整的ACID事务处理
- **高性能**: 连接池、批量操作和验证缓存优化

## 安装

```bash
go get github.com/muidea/magicOrm
```

## 快速开始

### 1. 定义模型

```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`
    Name   string   `orm:"name" view:"detail,lite"`
    EMail  string   `orm:"email" view:"detail,lite"`
    Status *Status  `orm:"status" view:"detail,lite"`
    Group  []*Group `orm:"group" view:"detail,lite"`
}

type Status struct {
    ID    int `orm:"id key auto" view:"detail,lite"`
    Value int `orm:"value" view:"detail,lite"`
}

type Group struct {
    ID     int      `orm:"gid key auto" view:"detail,lite"`
    Name   string   `orm:"name" view:"detail,lite"`
    Users  *[]*User `orm:"users" view:"detail,lite"`
    Parent *Group   `orm:"parent" view:"detail,lite"`
}
```

### 2. 初始化ORM

```go
import (
    "github.com/muidea/magicOrm/orm"
    "github.com/muidea/magicOrm/provider"
)

// 数据库配置
config := &orm.Options{
    Driver: "postgres", // 或 "mysql"
    DSN:    "数据库连接字符串",
}

// 初始化
orm.Initialize()
defer orm.Uninitialized()

// 创建Provider
localProvider := provider.NewLocalProvider("default", nil)

// 创建ORM实例
o1, err := orm.NewOrm(localProvider, config, "schema_prefix")
defer o1.Release()
if err != nil {
    log.Fatal(err)
}
```

### 3. 注册模型

```go
// 注册所有模型
entityList := []any{&User{}, &Status{}, &Group{}}
modelList, err := registerLocalModel(localProvider, entityList)
if err != nil {
    log.Fatal(err)
}

// 创建数据表
err = createModel(o1, modelList)
if err != nil {
    log.Fatal(err)
}
```

### 4. 基本CRUD操作

#### 插入数据
```go
user := &User{
    Name:  "demo", 
    EMail: "123@demo.com", 
    Group: []*Group{},
}

userModel, err := localProvider.GetEntityModel(user, true)
if err != nil {
    log.Fatal(err)
}

userModel, err = o1.Insert(userModel)
if err != nil {
    log.Fatal(err)
}

user = userModel.Interface(true).(*User)
fmt.Printf("插入成功，ID: %d\n", user.ID)
```

#### 查询数据
```go
// 单个查询
queryUser := &User{ID: user.ID}
queryModel, err := localProvider.GetEntityModel(queryUser, true)
if err != nil {
    log.Fatal(err)
}

queryModel, err = o1.Query(queryModel)
if err != nil {
    log.Fatal(err)
}

result := queryModel.Interface(true).(*User)
fmt.Printf("查询结果: %+v\n", result)
```

#### 更新数据
```go
user.Name = "updated name"
userModel, err = localProvider.GetEntityModel(user, true)
if err != nil {
    log.Fatal(err)
}

userModel, err = o1.Update(userModel)
if err != nil {
    log.Fatal(err)
}
```

#### 删除数据
```go
_, err = o1.Delete(userModel)
if err != nil {
    log.Fatal(err)
}
```

## 核心功能

### CRUD操作

- **Insert** - 插入单个对象
- **Update** - 更新指定对象  
- **Delete** - 删除指定对象
- **Query** - 查询指定对象（根据主键）
- **BatchQuery** - 按条件批量查询对象

### 查询过滤器

```go
filter, _ := localProvider.GetModelFilter(model)

// 基础比较
filter.Equal("name", "value")        // 等于
filter.NotEqual("name", "value")     // 不等于
filter.Below("age", 18)              // 小于
filter.Above("age", 18)              // 大于
filter.In("id", []int{1, 2, 3})      // 在指定集合内
filter.NotIn("id", []int{1, 2, 3})   // 在指定集合外
filter.Like("name", "%demo%")        // 模糊匹配

// 组合查询
filter.Equal("status", "active")
filter.Above("created_at", startTime)
```

### 模型管理

```go
// 创建表
err := o1.Create(model)

// 删除表
err := o1.Drop(model)

// 检查表是否存在
exists, err := o1.Exist(model)
```

### 事务支持

```go
// 开始事务
tx, err := o1.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// 在事务中执行操作
userModel, err := tx.Insert(userModel)
if err != nil {
    return err
}

// 提交事务
err = tx.Commit()
if err != nil {
    return err
}
```

## 数据类型和标签

### 支持的数据类型

#### 基础数据类型
- 整数: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- 浮点数: `float32`, `float64`
- 布尔值: `bool`
- 字符串: `string`
- 时间: `time.Time`
- UUID: `string` (配合 `uuid` 标签使用)
- 以及对应的指针类型

#### 复合数据类型
- 结构体: `struct`
- 切片: `slice`
- 指针: `pointer`

### ORM标签说明

```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`                // 主键，自增
    Name   string   `orm:"name" constraint:"req,min=3,max=50" view:"detail,lite"` // 必填，长度3-50
    EMail  string   `orm:"email" constraint:"re=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,64}$" view:"detail,lite"` // 邮箱格式
    UU_ID  string   `orm:"uuid key uuid" view:"detail,lite"`                // UUID主键
    Status *Status  `orm:"status" view:"detail,lite"`                       // 可选关联
    Group  []*Group `orm:"group" view:"detail,lite"`                        // 多对多关联
    Time   time.Time `orm:"created_at datetime" constraint:"ro" view:"detail"` // 不可变时间
}
```

**ORM标签说明:**
- `key`: 主键标识
- `auto`: 自增标识 (PostgreSQL使用SERIAL/BIGSERIAL)
- `uuid`: UUID类型主键 (PostgreSQL VARCHAR(32))
- `datetime`: 时间类型
- `view`: 视图声明，支持 `detail` 和 `lite` 模式

### 约束系统

约束标签用于定义字段的业务规则和数据验证，支持访问行为约束和内容值约束。

#### 语法规范
```
constraint:"指令1,指令2=参数1,指令3=参数1:参数2"
```

- **指令分隔符**: `,` (英文逗号)
- **键值分隔符**: `=` (等号)
- **参数分隔符**: `:` (冒号)

#### 访问行为约束

| 约束 | 参数 | 描述 | 使用场景 |
| :--- | :--- | :--- | :--- |
| **`req`** | 无 | **Required**: 必填/必传 | 校验值不能为零值（0, "", nil） |
| **`ro`** | 无 | **Read-Only**: 只读 | 输出接口展示，更新接口忽略此字段 |
| **`wo`** | 无 | **Write-Only**: 只写 | 敏感字段（如密码），禁止在展示接口输出 |

#### 内容值约束

| 约束 | 参数示例 | 描述 | 适用类型 |
| :--- | :--- | :--- | :--- |
| **`min`** | `min=1` | **最小值/最小长度** | 数字、字符串、数组 |
| **`max`** | `max=100` | **最大值/最大长度** | 数字、字符串、数组 |
| **`range`** | `range=1:100` | **区间约束**: 定义数值的闭区间 `[min, max]` | 数字、浮点数 |
| **`in`** | `in=active:inactive:pending` | **枚举约束**: 字段值必须在定义的参数集合内 | 字符串、数字 |
| **`re`** | `re=^[a-z]+$` | **正则约束**: 字段值必须匹配指定的正则表达式 | 字符串 |

#### 约束示例

```go
type UserAccount struct {
    // 访问行为约束示例
    ID         int    `orm:"id key auto" constraint:"ro"`           // 自增主键，只读
    Name       string `orm:"name" constraint:"req"`                // 必填
    Password   string `orm:"password" constraint:"wo"`              // 只写（敏感字段）
    CreateTime int64  `orm:"create_time" constraint:"ro"`         // 不可变
    Email      string `orm:"email" constraint:"req,re=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,64}$"`
    UpdateTime int64  `orm:"update_time"`                          // 普通字段
    Status     int    `orm:"status" constraint:"req,ro"`           // 必填且只读
}

type Product struct {
    // 内容值约束示例
    ID          int     `orm:"id key auto" constraint:"ro"`          // 只读主键
    Name        string  `orm:"name" constraint:"req,min=3,max=50"`   // 必填，长度3-50
    Age         int     `orm:"age" constraint:"min=0,max=150"`      // 年龄0-150
    Score       float64 `orm:"score" constraint:"range=0.0:100.0"`  // 分数0.0-100.0
    Status      string  `orm:"status" constraint:"in=active:inactive:pending"` // 枚举值
    Description string  `orm:"description" constraint:"max=500"`    // 最大长度500
    Price       float64 `orm:"price" constraint:"range=0.01:9999.99"` // 价格范围
    Category    string  `orm:"category" constraint:"in=A:B:C:D"`     // 分类枚举
    Code        string  `orm:"code" constraint:"re=^[A-Z]{3}-\\d{3}$"` // 格式：ABC-123
}
```

## 验证系统

MagicORM 实现了先进的四层验证架构，提供场景感知的验证策略和丰富的错误处理。

### 四层验证架构

#### 1. 类型验证层 (`validation/types/`)
- **职责**: 基础类型验证和转换
- 验证Go类型与数据库类型的兼容性
- 处理类型转换（字符串↔整数、时间格式等）
- 确保基本的数据完整性

#### 2. 约束验证层 (`validation/constraints/`)
- **职责**: 业务约束验证
- 验证结构体标签中定义的业务规则（`req`, `min`, `max`, `range`, `in`, `re`）
- 处理访问行为约束（`ro`, `wo`）
- 支持场景感知验证（Insert vs Update）

#### 3. 数据库验证层 (`validation/database/`)
- **职责**: 数据库特定约束验证
- 验证数据库级约束（NOT NULL, UNIQUE, FOREIGN KEY等）
- 处理数据库类型兼容性
- 提供数据库特定的错误消息

#### 4. 场景适配层 (`validation/scenario/`)
- **职责**: 场景感知验证编排
- 基于操作类型编排验证（Insert, Update, Query, Delete）
- 为不同场景应用不同的验证策略
- 为其他层提供验证上下文

### 验证使用示例

#### 基本验证配置

```go
import (
    "github.com/muidea/magicOrm/validation"
    "github.com/muidea/magicOrm/validation/errors"
)

// 使用默认配置
config := validation.DefaultConfig()
manager := validation.NewValidationManager(config)

// 创建验证上下文
ctx := validation.NewContext(
    errors.ScenarioInsert,      // 插入场景
    validation.OperationCreate, // 创建操作
    nil,                        // 模型适配器
    "postgresql",               // 数据库类型
)

// 验证值
err := manager.Validate("test value", ctx)
if err != nil {
    // 处理验证错误
    fmt.Printf("验证失败: %v\n", err)
}
```

#### 场景感知验证

```go
// 定义带约束的模型
type User struct {
    ID       int    `orm:"id key auto" constraint:"ro"`
    Username string `orm:"username" constraint:"req,min=3,max=20"`
    Email    string `orm:"email" constraint:"req,re=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
    Age      int    `orm:"age" constraint:"min=18,max=120"`
    Status   string `orm:"status" constraint:"in=active:inactive:suspended"`
}

// 不同场景的验证策略
scenarios := []errors.Scenario{
    errors.ScenarioInsert,  // 插入：严格验证
    errors.ScenarioUpdate,  // 更新：跳过只读字段
    errors.ScenarioQuery,   // 查询：跳过只写字段
    errors.ScenarioDelete,  // 删除：最小验证
}

for _, scenario := range scenarios {
    ctx := validation.NewContext(
        scenario,
        validation.OperationCreate,
        nil,
        "postgresql",
    )
    
    // 执行场景特定的验证
    err := manager.ValidateModel(model, ctx)
    if err != nil {
        fmt.Printf("%s 场景验证失败: %v\n", scenario, err)
    }
}
```

#### 错误处理

```go
// 创建错误收集器
collector := errors.NewErrorCollector()

// 创建带错误收集器的上下文
ctx := validation.NewContextWithCollector(
    errors.ScenarioInsert,
    collector,
)

// 执行验证（收集所有错误）
err := manager.ValidateModel(model, ctx)
if collector.HasErrors() {
    // 获取所有错误
    allErrors := collector.GetErrors()
    
    // 按字段获取错误
    fieldErrors := collector.GetErrorsByField("username")
    
    // 按验证层获取错误
    typeErrors := collector.GetErrorsByLayer(errors.LayerType)
    constraintErrors := collector.GetErrorsByLayer(errors.LayerConstraint)
    
    // 获取错误摘要
    summary := collector.GetErrorSummary()
    fmt.Printf("验证错误摘要:\n%s\n", summary)
}
```

### 验证配置

#### 默认配置

```go
// 默认配置（推荐用于大多数场景）
config := validation.DefaultConfig()
// 启用所有验证层
// 启用缓存（5分钟TTL）
// 收集所有错误（不提前停止）
```

#### 简单配置

```go
// 简单配置（基本验证需求）
config := validation.SimpleConfig()
// 启用类型和约束验证
// 禁用数据库验证和场景适配
// 禁用缓存
// 遇到第一个错误即停止
```

#### 性能优化配置

```go
// 性能优化配置
config := validation.ValidationConfig{
    EnableTypeValidation:       true,
    EnableConstraintValidation: true,
    EnableDatabaseValidation:   false, // 跳过数据库验证以提高性能
    EnableScenarioAdaptation:   true,
    EnableCaching:              true,
    CacheTTL:                   10 * time.Minute, // 更长TTL
    MaxCacheSize:               2000,             // 更大缓存
    DefaultOptions: validation.ValidationOptions{
        StopOnFirstError:        true, // 提前停止以提高性能
        IncludeFieldPathInError: false,
        ValidateReadOnlyFields:  true,
        ValidateWriteOnlyFields: true,
    },
}
```

#### 严格验证配置

```go
// 严格验证配置
config := validation.ValidationConfig{
    EnableTypeValidation:       true,
    EnableConstraintValidation: true,
    EnableDatabaseValidation:   true,
    EnableScenarioAdaptation:   true,
    EnableCaching:              false, // 禁用缓存以确保严格验证
    DefaultOptions: validation.ValidationOptions{
        StopOnFirstError:        false, // 收集所有错误
        IncludeFieldPathInError: true,  // 包含字段路径
        ValidateReadOnlyFields:  true,
        ValidateWriteOnlyFields: true,
    },
}
```

### 验证缓存

验证系统支持多层缓存以提高性能：

```go
// 启用缓存
config := validation.DefaultConfig()
config.EnableCaching = true
config.CacheTTL = 5 * time.Minute
config.MaxCacheSize = 1000

manager := validation.NewValidationManager(config)

// 缓存统计
stats := manager.GetValidationStats()
fmt.Printf("缓存命中率: %.2f%%\n", stats.CacheHitRate*100)
fmt.Printf("类型验证次数: %d\n", stats.TypeValidations)
fmt.Printf("约束验证次数: %d\n", stats.ConstraintValidations)
```

### 自定义验证

#### 注册自定义约束

```go
// 创建约束验证器
validator := constraints.NewConstraintValidator(true)

// 注册自定义约束
validator.RegisterCustomConstraint("custom", func(value any, args []string) error {
    // 自定义验证逻辑
    strValue, ok := value.(string)
    if !ok {
        return fmt.Errorf("值必须是字符串类型")
    }
    
    // 检查自定义规则
    if len(strValue) < 5 {
        return fmt.Errorf("值长度必须至少为5个字符")
    }
    
    return nil
})

// 使用自定义约束
type CustomModel struct {
    ID   int    `orm:"id key auto"`
    Code string `orm:"code" constraint:"custom"` // 使用自定义约束
}
```

#### 注册自定义类型处理器

```go
// 创建类型验证器
typeValidator := types.NewTypeValidator()

// 注册自定义类型处理器
typeValidator.RegisterTypeHandler("MyCustomType", &myTypeHandler{})

// 自定义类型处理器实现
type myTypeHandler struct{}

func (h *myTypeHandler) Validate(value any) error {
    // 验证自定义类型
    return nil
}

func (h *myTypeHandler) Convert(value any) (any, error) {
    // 转换到自定义类型
    return value, nil
}

func (h *myTypeHandler) GetZeroValue() any {
    return MyCustomType{}
}

func (h *myTypeHandler) GetType() reflect.Type {
    return reflect.TypeOf(MyCustomType{})
}
```

## 监控系统

MagicORM 提供了简洁的监控系统，专注于数据收集，支持 ORM 操作、验证系统和数据库执行的监控。

### 架构设计

**核心原则**：MagicORM 只负责数据收集，不负责导出和管理。监控数据由外部系统处理。

**文件结构**：
```
monitoring/
├── collector.go                    # 顶层collector接口和类型定义
├── init.go                         # 初始化集成
├── e2e_test.go                     # 端到端测试
├── core/                           # 核心类型定义
├── orm/                            # ORM监控
├── validation/                     # 验证监控
├── database/                       # 数据库监控
└── example/                        # 使用示例
```

### 快速开始

#### 基本使用

```go
import "github.com/muidea/magicOrm/monitoring"

// 创建collector
collector := monitoring.NewCollector()

// 记录ORM操作
collector.RecordORMOperation(
    monitoring.OperationInsert,
    "User",
    true,
    150*time.Millisecond,
    nil,
    map[string]string{"database": "postgresql"},
)

// 记录验证操作
collector.RecordValidationOperation(
    "validate_user",
    "User",
    monitoring.ScenarioInsert,
    50*time.Millisecond,
    nil,
    map[string]string{"field_count": "5"},
)

// 记录数据库操作
collector.RecordDatabaseOperation(
    "postgresql",
    monitoring.QueryTypeSelect,
    true,
    200*time.Millisecond,
    10,
    nil,
    map[string]string{"table": "users"},
)
```

#### 集成到现有代码

```go
import (
    "github.com/muidea/magicOrm/monitoring"
    "github.com/muidea/magicOrm/orm"
)

// 创建带监控的ORM
func createMonitoredORM(provider orm.Provider, config *orm.Options) (*orm.Orm, monitoring.Collector) {
    collector := monitoring.NewCollector()
    
    // 创建ORM实例
    o, err := orm.NewOrm(provider, config, "schema_prefix")
    if err != nil {
        return nil, nil
    }
    
    // 包装为带监控的ORM
    monitoredOrm := monitoring.NewMonitoredOrm(o, collector)
    return monitoredOrm, collector
}
```

### 监控数据类型

#### ORM 操作监控
- **操作类型**: Insert, Update, Delete, Query, BatchQuery
- **指标**: 成功率、延迟、错误类型
- **标签**: 模型名称、数据库类型、操作类型

#### 验证系统监控
- **场景**: Insert, Update, Query, Delete
- **指标**: 验证延迟、缓存命中率、错误统计
- **标签**: 验证器名称、模型名称、场景类型

#### 数据库执行监控
- **查询类型**: Select, Insert, Update, Delete, Transaction
- **指标**: 查询延迟、返回行数、连接状态
- **标签**: 数据库类型、表名、操作类型

### 标签系统

支持灵活的标签系统，用于分类和过滤监控数据：

```go
// 基本标签
labels := map[string]string{
    "database": "postgresql",
    "table":    "users",
    "operation": "insert",
}

// 合并默认标签
collector.WithDefaultLabels(map[string]string{
    "environment": "production",
    "service":     "user-service",
})

// 记录带标签的操作
collector.RecordORMOperation(
    monitoring.OperationInsert,
    "User",
    true,
    150*time.Millisecond,
    nil,
    labels,
)
```

### 错误处理

监控系统支持详细的错误分类：

```go
// 错误类型定义
type ErrorType string

const (
    ErrorTypeDatabase   ErrorType = "database"
    ErrorTypeValidation ErrorType = "validation"
    ErrorTypeConstraint ErrorType = "constraint"
    ErrorTypeType       ErrorType = "type"
    ErrorTypeSystem     ErrorType = "system"
)

// 记录带错误信息的操作
collector.RecordORMOperation(
    monitoring.OperationInsert,
    "User",
    false, // 操作失败
    150*time.Millisecond,
    &monitoring.ErrorInfo{
        Type:    monitoring.ErrorTypeDatabase,
        Message: "duplicate key value violates unique constraint",
        Code:    "23505",
    },
    labels,
)
```

### 性能优化

监控系统设计为低开销：

1. **异步收集**：默认启用异步收集，减少对业务逻辑的影响
2. **采样率控制**：支持配置采样率，控制监控数据量
3. **内存优化**：使用高效的数据结构，避免内存泄漏
4. **零分配设计**：关键路径避免内存分配

### 测试和验证

```bash
# 运行监控系统测试
go test ./monitoring/... -v

# 运行端到端测试
go test ./monitoring/e2e_test.go -v

# 运行示例程序
cd monitoring/example && go run example.go
```

### 与外部系统集成

监控数据可以通过多种方式导出：

```go
// 获取原始监控数据
data := collector.GetMetrics()

// 转换为JSON格式
jsonData, _ := json.Marshal(data)

// 转换为Prometheus格式
prometheusData := collector.ToPrometheusFormat()

// 自定义导出处理器
collector.SetExportHandler(func(metrics []monitoring.Metric) {
    // 发送到外部监控系统
    sendToExternalSystem(metrics)
})
```

### 最佳实践

1. **合理使用标签**：使用有意义的标签便于数据分析和过滤
2. **控制数据量**：根据需求调整采样率和数据保留策略
3. **错误分类**：使用详细的错误类型便于问题排查
4. **性能监控**：监控监控系统本身的性能
5. **集成测试**：在生产环境前充分测试监控集成

## 高级特性

### 视图模式

```go
type User struct {
    ID     int      `orm:"uid key auto" view:"detail,lite"`
    Name   string   `orm:"name" view:"detail,lite"`
    EMail  string   `orm:"email" view:"detail"`           // 仅在detail视图
    Status *Status  `orm:"status" view:"detail"`          // 仅在detail视图
    Group  []*Group `orm:"group" view:"detail"`           // 仅在detail视图
}

// 使用lite视图查询（只包含基础字段）
liteModel := userModel.Copy(models.LiteView)
result, err := o1.Query(liteModel)

// 使用detail视图查询（包含所有字段）
detailModel := userModel.Copy(models.DetailView)
result, err = o1.Query(detailModel)
```

### 关联关系

```go
// 一对一关系
type User struct {
    ID     int     `orm:"uid key auto"`
    Profile *Profile `orm:"profile"`
}

// 一对多关系
type Group struct {
    ID    int      `orm:"gid key auto"`
    Name  string   `orm:"name"`
    Users *[]*User `orm:"users"`
}

// 多对多关系
type User struct {
    ID    int       `orm:"uid key auto"`
    Name  string    `orm:"name"`
    Groups []*Group `orm:"groups"`
}

type Group struct {
    ID    int      `orm:"gid key auto"`
    Name  string   `orm:"name"`
    Users []*User  `orm:"users"`
}
```

### 字段约束规则

1. **可选字段**: 如果字段类型为指针，则表示该字段为可选类型（可为NULL）
2. **复合类型指针**: 对于指针类型的复合类型成员，ORM只处理对象与复合类型成员之间的关系
3. **复合类型**: 对于普通复合类型成员，ORM会同步处理对象与复合类型成员之间的关系
4. **切片类型**: 
   - 不支持基础类型指针切片，如 `[]*bool`, `[]*int`
   - 支持切片指针，如 `*[]bool`, `*[]int`

## 配置选项

### 数据库配置

```go
// PostgreSQL配置
config := &orm.Options{
    Driver: "postgres",
    DSN:    "postgres://user:password@localhost:5432/dbname?sslmode=disable",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    ConnMaxLifetime: time.Hour,
}

// MySQL配置
config := &orm.Options{
    Driver: "mysql",
    DSN:    "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    ConnMaxLifetime: time.Hour,
}
```

## 最佳实践

### 数据库和ORM
1. **合理选择数据库**: 根据项目需求选择合适的数据库
2. **合理使用视图**: 对于大对象，使用 `lite` 视图减少数据传输
3. **事务处理**: 对于复杂操作，使用事务确保数据一致性
4. **连接池配置**: 根据应用负载合理配置连接池参数
5. **错误处理**: 始终检查和处理ORM操作返回的错误
6. **批量操作**: 对于大量数据操作，使用批量查询和插入提高性能

### 验证系统
7. **场景感知验证**: 根据操作类型使用不同的验证策略
   - **Insert**: 严格验证所有约束
   - **Update**: 跳过只读字段验证
   - **Query**: 跳过只写字段验证
   - **Delete**: 最小化验证

8. **性能优化配置**:
   - 生产环境：启用缓存，设置合理的TTL
   - 开发环境：禁用缓存以便调试
   - 测试环境：启用所有错误收集

9. **错误处理策略**:
   - 用户输入验证：收集所有错误，提供完整反馈
   - 内部数据处理：遇到第一个错误即停止，快速失败
   - 日志记录：记录详细的验证错误信息

10. **约束设计原则**:
    - 必填字段使用 `req` 约束
    - 敏感字段使用 `wo`（只写）约束
    - 不可变字段使用 `ro`（只读）约束
    - 使用 `min`/`max` 约束确保数据范围
    - 使用 `in` 约束限制枚举值
    - 使用 `re` 约束验证格式

11. **缓存策略**:
    - 静态数据：使用较长TTL（10分钟+）
    - 动态数据：使用较短TTL（1-5分钟）
    - 根据内存限制调整缓存大小
    - 监控缓存命中率和性能

12. **环境特定配置**:
    - **开发环境**: 禁用缓存，启用详细错误
    - **测试环境**: 启用所有验证层，收集所有错误
    - **预发布环境**: 启用缓存，监控性能
    - **生产环境**: 优化配置，确保稳定性和性能

## 示例和参考

### 基础示例
完整的使用示例请参考 `test/` 目录下的测试文件：

- `test/simple_test.go` - 基础CRUD操作
- `test/model_local_test.go` - 模型关系示例
- `test/batch_operation_local_test.go` - 批量操作示例
- `orm/builder_postgres_test.go` - PostgreSQL构建器测试

### 验证系统示例
验证系统的完整示例请参考 `validation/` 目录：

- `validation/example/usage_example.go` - 验证系统使用示例
- `validation/example/configuration_example.go` - 验证配置示例
- `validation/test/simple_test.go` - 基础验证测试
- `validation/test/integration_test.go` - 集成测试

### 约束测试示例
约束系统的测试示例请参考 `test/` 目录：

- `test/constraint_local_test.go` - Local Provider约束测试
- `test/constraint_remote_test.go` - Remote Provider约束测试
- `test/constraint.go` - 约束测试模型定义

### 架构文档
详细的架构设计文档：

- `VALIDATION_ARCHITECTURE.md` - 四层验证架构设计
- `VALIDATION_IMPLEMENTATION_PLAN.md` - 验证系统实施计划
- `AGENTS.md` - 开发指南和命令参考
- `monitoring/README.md` - 监控系统详细文档

## 许可证

MIT License
