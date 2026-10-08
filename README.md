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
- **条件更新**: 单 SQL compare-and-set 更新并返回影响行数
- **Schema 演进**: 可安全补建字段、关系、唯一约束和索引
- **Schema 核验**: 只读检查实际主表、关系表及 owned 模型，报告结构缺失或漂移
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

仅用于内存组合的字段声明 `orm:"-"`，不会生成列或关系，也不要求注册其类型。`json:"-"` 只隐藏 JSON；内部组合通常同时声明两者。详见 [字段排除规则](docs/tags-reference.md#15-忽略字段)。

### 2. 初始化ORM

```go
import (
    "context"
    "github.com/muidea/magicOrm/orm"
    "github.com/muidea/magicOrm/provider"
)

// 初始化全局 ORM（连接池管理）
orm.Initialize()
defer orm.Uninitialized()

// 注册数据库（示例：PostgreSQL）
// dbServer 形如 "localhost:5432"，dbName 为数据库名，"default" 为自定义 owner 标识
if err := orm.AddDatabase("localhost:5432", "mydb", "user", "password", 25, "default"); err != nil {
    log.Fatal(err)
}

// 创建 Provider（与数据库 owner 对应）
localProvider := provider.NewLocalProvider("default", nil)

// 从连接池获取 ORM 实例
ctx := context.Background()
o1, err := orm.GetOrm(ctx, localProvider, "schema_prefix")
if err != nil {
    log.Fatal(err)
}
defer o1.Release()
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

> 基于当前实现整理的设计文档见 [docs/README.md](./docs/README.md)。

当前仅 **Insert、Update、UpdateWithFilter、Delete** 会执行模型验证；**Query、BatchQuery** 不执行验证。详见 [docs/design-validation.md](./docs/design-validation.md) 与 [docs/design-orm.md](./docs/design-orm.md)。

### CRUD操作

- **Insert** - 插入单个对象
- **Update** - 更新指定对象  
- **UpdateWithFilter** - 单表条件更新，返回受影响行数
- **Delete** - 删除指定对象
- **Query** - 针对模型对象的单条查询
- **BatchQuery** - 按条件批量查询多个对象

当前稳定的对外查询约定：
- `Query(model)` 是单条查询正式入口，适合业务侧基于模型对象做单查
- `BatchQuery(filter)` 是多条查询正式入口，适合按过滤条件获取对象集合
- ORM 层不再提供额外的“按 filter 单查”入口，避免与 `Query(model)` 形成重叠语义
- 查询返回裁剪遵循固定优先级：
  - `Query(model)`：不处理 `ValueMask`，顶层结果固定按 `DetailView` 返回
  - `BatchQuery(filter)`：顶层对象按 `ValueMask > view` 裁剪
  - 子对象：统一按 `lite` 返回，不接受父对象 `detail` 或嵌套 `ValueMask` 放大
  - 主键字段始终保留
- 因此业务若需要子对象详情，应拿到子对象主键后单独查询

批量关系展开在单次 QueryRunner 内复用已读取的目标和已确认不存在的目标；不同关系字段指向相同目标时，仅查询尚未读取的 ID。关系边仍从存储读取，子对象仍按 lite view 投影。缓存随该次查询结束，后续查询重新读取，不提供跨请求、跨事务或跨数据库的结果缓存。

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

### 条件状态更新

当状态变更必须同时验证当前状态、版本或租户时，使用 `UpdateWithFilter`，不要拆成“先查询再更新”：

```go
change, _ := localProvider.GetEntityModel(&Session{State: "consumed"}, true)
filter, _ := localProvider.GetModelFilter(change)
filter.Equal("namespace", "tenant-a")
filter.Equal("state", "issued")

rows, err := o1.UpdateWithFilter(change, filter)
if err != nil {
    return err
}
if rows == 0 {
    return ErrStateChanged
}
```

该操作仅更新基础字段，过滤条件也只能使用基础字段。完整约定见 [docs/design-orm.md](./docs/design-orm.md)。

### 模型管理

```go
// 创建表
err := o1.Create(model)

// 删除表
err := o1.Drop(model)

// 仅补建安全的 schema 差异；不兼容变更会返回迁移错误
err = o1.Reconcile(previousModel, currentModel)
```

`InspectSchema(model, true)` 只读核验 ORM 生成的物理结构；`InspectSchema(model, false)` 核验 `Drop` 管理的全部表已不存在。报告包括 `matches`、已检查的表名和不含默认值内容/数据库凭据的差异项；查询失败、读取中断、取消或不支持检查的执行器返回错误且不返回部分成功报告。

`Create`、`Drop`、`Reconcile` 及其公开 Runner 在执行任何 SQL 之前检查完整声明图：主键、owned 环、图内物理表名冲突和最多 1024 张表，并预先构造全部 DDL。后续字段/关系/索引构造失败或新增必填字段被拒绝时，不执行前面已生成的 SQL。共享 owned 模型只生成一次建/删表步骤；增量迁移新增 owned 字段时同时建立其组件表和关系表，不建立引用模型的主表。

不打开数据库的预检可使用 `PreflightSchemaChange(ctx, previous, desired, provider, prefix)`：previous 为 nil 表示建表，desired 为 nil 表示删表，二者非空表示增量迁移；provider 必须包含本次操作的关联声明。`ValidateSchema` 仅检查声明图。预检不查询实际结构、不保证 SQL 执行成功，也不替代 DDL 后的物理核验或提供事务回滚。

检查覆盖主表、关系表、递归 owned 模型（不包含引用模型的主表）、列类型/精度、可空性、默认值、自增、主键、命名唯一约束与有序索引。多余列/索引、视图及无效/部分/表达式等非声明索引均报告漂移。检查使用数据库真实目录并绑定 schema/table，不用 `IF NOT EXISTS` 的执行成功替代核验。PostgreSQL 使用显式配置的 Schema；MySQL 要求已选择数据库及 MySQL 8.0.13+ 的目录字段。未知结构不猜测兼容，也不执行 catalog 中的表达式。

`InspectSchema` 不是自动迁移/恢复接口：调用者负责串行 Schema 写入，检查期间不得并发执行外部 DDL；结果不是分布式锁或持久化完成回执。检查限于 ORM 所生成的结构，不审计额外的触发器、权限、RLS 策略、CHECK/外键策略或存储参数。`Reconcile` 的安全边界见 [docs/design-orm.md](./docs/design-orm.md)。

显式受控补执行可使用 `RepairSchema(previous, desired)`，nil 的含义与预检一致。调用者必须提供已验证操作身份的冻结声明，并保证所有 Schema 写入串行。ORM 先校验完整计划并读取目标图的物理结构，再构造全部缺失步骤：补建新表/owned/关系表、原增量计划中新加的可空列及命名唯一约束/索引，或删除原 drop 计划中仍完整匹配的表；已完成步骤不重放，执行后再次完整检查物理结果。增量迁移中原有表/列/索引丢失、主键缺失、既有关系表结构不完整、类型/约束漂移或额外对象均拒绝修复，不猜测数据恢复、不补填必填列、不接受任意 SQL。中途失败可由同一冻结计划重新检查后续跑，但不承诺 DDL 事务回滚、跨进程互斥或应用元数据/运行态恢复。

物理目录集成测试默认跳过，必须显式提供隔离测试库：

```bash
MAGICORM_SCHEMA_TEST_POSTGRES_DSN='<disposable PostgreSQL DSN>' go test ./database/postgres -run TestPostgresPhysicalSchemaCatalog -count=1
MAGICORM_SCHEMA_TEST_MYSQL_DSN='<disposable MySQL DSN>' go test ./database/mysql -run TestMySQLPhysicalSchemaCatalog -count=1
```

测试创建随机 Schema/数据库，仅清理本次测试创建的对象；不要指向生产实例。

### 事务支持

事务在同一 Orm 实例上通过 `BeginTransaction` / `CommitTransaction` / `RollbackTransaction` 完成，无返回值 tx 对象：

```go
// 开始事务
err := o1.BeginTransaction()
if err != nil {
    return err
}
defer o1.RollbackTransaction()

// 在事务中执行操作（同一 Orm 实例）
userModel, err = o1.Insert(userModel)
if err != nil {
    return err
}

// 提交事务
err = o1.CommitTransaction()
if err != nil {
    return err
}
```

## 数据类型和标签

### 支持的数据类型

#### 基础数据类型
- 整数: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- 浮点数: `float32`, `float64`
- 布尔值: `boolean`
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

字段内容约束失败统一返回 `IllegalParam`，local/remote provider 与 ORM 场景校验
保留字段名、失败指令和原因，例如 `field 'price': constraint 'min': too small/short`。
多个字段失败也保留各自诊断。内置包装不追加输入值；自定义 validator 应返回适合公开的原因，
不要把密码、令牌或完整对象写入错误。可通过 `errors.As` 获取 `models.ConstraintViolation`，
其 `Unwrap` 保留原 validator 的错误。

可选指针关系在创建时省略或显式置空均不插入关联；`req` 关系仍要求有效目标。
显式 null 的赋值标记保留，用于更新时清空已有引用。没有持久关系的查询返回 null，
不会返回查询掩码中的空对象。普通 Insert 出错或 panic 会回滚当前事务；panic 仍传回调用者，
调用者不能把异常当作成功或盲目重放业务写入。
PostgreSQL 无普通列输入时使用 `DEFAULT VALUES`，支持仅由数据库生成主键的主对象插入。

ORM 默认关闭校验缓存。显式开启后，缓存键包含标量值、类型、规则参数和场景；
不能安全编码的复杂值不缓存。注册自定义规则会清空该 validator 的缓存。

隔离数据库回归（所有建表、写入和回读均经过 ORM；服务器须指向一次性测试库）：

```bash
MAGICORM_RELATION_TEST_SERVER=127.0.0.1:<port> MAGICORM_RELATION_TEST_USER=postgres \
  go test ./test -run '^TestNullReferenceDatabaseContract$' -count=1
MAGICORM_RELATION_TEST_SERVER=127.0.0.1:<port> MAGICORM_RELATION_TEST_USER=root \
  go test -tags=mysql ./test -run '^TestNullReferenceDatabaseContract$' -count=1
```

该回归使用 `testdb` 和仓库测试密码，覆盖两种 provider 的省略/null/绑定、必需引用、
清空及关系转换失败或 panic 后的独立回读。其他集成测试可通过
`MAGICORM_POSTGRES_*` / `MAGICORM_MYSQL_*` 配置服务器、数据库、Schema、用户名和密码。

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

MagicORM 的监控通过 **`metrics` 包**与 **magicCommon/monitoring** 集成：ORM 操作在内部自动上报到全局 collector，并在存在 `monitoring.GlobalManager` 时注册为 Provider（名称 `magicorm_orm`）。无独立 `monitoring` 包或 `MonitoredOrm` 包装。

**当前实现**：`orm.Initialize()` 会创建 ORM / DB / Validation 三类 collector；当 `monitoring.GlobalManager` 已存在时会自动注册 `magicorm_orm`、`magicorm_database` 与 `magicorm_validation`。推荐顺序是先执行 `monitoring.InitializeGlobalManager()`，再执行 `orm.Initialize()`；如果监控系统晚于 ORM 初始化，可后续调用 `orm.EnsureORMMetricProviderRegistered()`、`metricsdb.EnsureDatabaseMetricProviderRegistered()` 与 `metricsvalidation.EnsureValidationMetricProviderRegistered()` 做幂等注册。详细说明见 [docs/README.md](./docs/README.md)、[docs/design-metrics.md](./docs/design-metrics.md) 与 [METRICS_TODO.md](./METRICS_TODO.md)（当前用于记录本轮改造结果）。

### 架构设计

**核心原则**：MagicORM 只负责数据收集，不负责导出和管理。监控数据由外部系统（magicCommon/monitoring）处理。

**文件结构**（当前实现）：
```
metrics/
├── metrics.go          # 操作类型、错误类型等常量
├── orm/                 # ORM 监控 collector 与 provider
│   ├── collector.go
│   └── provider.go
├── metricsdb/           # 数据库监控（已接入执行层）
└── validation/          # 验证监控（已接入验证链路）
```

### 监控数据类型

#### ORM 操作监控（已接入）
- **操作类型**: Insert, Update, Delete, Query, BatchQuery, Create, Drop, Count
- **指标**: 成功率、延迟、错误类型
- **标签**: 模型名称、操作类型等（见 `metrics` 包）
- **成熟度**: 已接入 ORM 主路径

#### 验证系统监控（已接入）
- **场景**: Insert, Update, Query, Delete
- **指标**: 校验次数、耗时、错误类型、缓存命中率、约束检查结果
- **接入点**: `validation.Manager`、`validation/cache`、约束校验链路
- **成熟度**: 已接入验证主路径

#### 数据库执行监控（已接入）
- **查询类型**: Select, Insert, Update, Delete, Create, Drop, Alter, Truncate 等 SQL 首关键字
- **事务类型**: Begin, Commit, Rollback
- **连接池状态**: Active, Idle, Open, Max
- **成熟度**: 已接入 MySQL/PostgreSQL 执行层

指标导出与外部系统集成由 magicCommon/monitoring 提供；测试可运行 `go test ./metrics/... -v`。更多说明见 [docs/design-metrics.md](./docs/design-metrics.md)。

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

// Query 固定返回顶层 DetailView
queryModel := userModel.Copy(models.LiteView) // 输入 view 只影响本地模型形状，不影响 Query 返回层级
result, err := o1.Query(queryModel)

// BatchQuery 才使用 view / ValueMask 控制顶层返回
filter, _ := localProvider.GetEntityFilter(&User{}, models.LiteView)
resultList, err := o1.BatchQuery(filter)
```

当前稳定约定：

- `Query(model)` 不处理 `ValueMask`，顶层固定按 `DetailView` 返回；
- `BatchQuery(filter)` 的顶层返回遵循 `ValueMask > view`；
- 包含/引用的子对象默认只返回 `lite` 字段，不会随着父对象 `detail` 自动扩成子对象 `detail`；
- 如果业务确实需要子对象详情，应拿到子对象主键后单独查询，不要把一次查询扩成多层 detail 读取。

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
   - 不支持基础类型指针切片，如 `[]*boolean`, `[]*int`
   - 支持切片指针，如 `*[]boolean`, `*[]int`

## 配置选项

### 数据库配置

```go
// 应用自定义的数据库配置结构体（示例），并非 magicOrm 内置类型
type DBOptions struct {
    Driver         string
    DSN            string
    MaxOpenConns   int
    MaxIdleConns   int
    ConnMaxLifetime time.Duration
}

// PostgreSQL配置（示例）
config := &DBOptions{
    Driver: "postgres",
    DSN:    "postgres://user:password@localhost:5432/dbname?sslmode=disable",
    MaxOpenConns: 25,
    MaxIdleConns: 5,
    ConnMaxLifetime: time.Hour,
}

// MySQL配置（示例）
config = &DBOptions{
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

- `test/simple_local_test.go` - 基础 CRUD 操作
- `test/reference_local_test.go` - 模型关系示例
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
详细的架构设计文档（以本列表为准，保持单一入口）：

- [docs/README.md](./docs/README.md) — 设计文档总览（整理设计 + 按功能块拆分的独立文档索引）
- [VALIDATION_ARCHITECTURE.md](./VALIDATION_ARCHITECTURE.md) — 四层验证架构设计
- [VALIDATION_IMPLEMENTATION_PLAN.md](./VALIDATION_IMPLEMENTATION_PLAN.md) — 验证系统实施计划
- [AGENTS.md](./AGENTS.md) — 开发指南和命令参考
- [docs/design-metrics.md](./docs/design-metrics.md) — 监控与指标设计
- [docs/testing-guide.md](./docs/testing-guide.md) — 测试分层与隔离测试库配置

## 许可证

MIT License

## 关系读取诊断

非指针单值关系没有关联 ID 时，记录 `query relation has no association`，包含所属 `model`、`field`、目标 `relation_model` 和原因 `non_pointer_relation_ids_empty`。整数主键记录为 `owner_id`；字符串主键仅记录 `owner_id_sha256`，避免主键本身含敏感信息。日志不输出完整对象或业务字段。

该提醒只描述当前关系读取结果，不证明引用目标损坏或业务必需关系缺失。指针关系为空仍不提醒；原有查询返回值、字段约束、关系与事务语义保持不变。业务 owner 应通过自身受控查询核对具体对象，不通过直接数据库访问绕过 ORM。

## SQL 成本窗口

启动时设置 `MAGIC_PROFILE_WINDOW=60s` 可独立于监控 collector 统计 MySQL/PostgreSQL Query、Execute 和逻辑事务调用。标签只使用引擎及 SQL 操作类别，不记录 SQL 或参数；查询耗时不包含完整行扫描与对象映射，事务次数不等于实际数据库提交数。完整采集边界及 SQL 指纹差分见工作区 [QPS 分析指南](../magicRunner/docs/guide-qps-analysis.md)。


## 连接池复用策略

PostgreSQL 与 MySQL 的 `Pool.Initialize(maxConnNum, config)` 使用相同策略：正数 `maxConnNum` 同时限制打开连接和空闲连接数量，按需建连，不预热全部连接。并发查询归还连接后可供下一批请求复用，避免默认仅保留两条空闲连接造成反复建连与认证。连接连续闲置一分钟后由 `database/sql` 清理；清理不是精确计时器，实际释放可能稍晚。

`maxConnNum <= 0` 保留既有不限打开连接的语义，空闲连接最多保留两条。各 owner 的连接预算仍独立；部署时需将多个进程、多个 owner 的上限合计纳入数据库容量规划。此次不改变事务、查询视图、租户/owner 隔离或业务授权，不缓存查询结果。连接池初始化探活失败时关闭已创建的数据库句柄。


## 查询模型的请求内复用

Count 与 BatchQuery 的指标记录复用该次查询已经解析的模型身份，避免为计数标签重新构造完整模型；发生模型解析前错误时仍保留原有标签回退。remote Object 的字段列表、复制结果和批量查询结果按已知长度分配切片，继续返回独立切片/模型，不缓存查询值或共享可变的结果模型。字段视图、显式响应 mask、私有字段读取与 NULL 处理保持既有合同。

`go test ./orm -run '^$' -bench '^BenchmarkFilterReadModels$' -benchmem` 测量带指标的一次 Count 与 BatchQuery（16 字段、单行、内存假执行器），用于跟踪模型解析/分配开销；它不测量数据库或 HTTP 吞吐。

### 数据库唯一约束错误

PostgreSQL 的 `23505` 和 MySQL 的 `1062` 在数据库执行层映射为通用 `def.Duplicated`，包括包装后的驱动错误。调用方可以针对明确冲突重新读取或重试；其他数据库错误及结果不明的提交保持失败。该分类不包含计费、配额或应用治理逻辑。

### Auto-increment validation

On insert, an omitted or zero-valued `auto` field requests database generation and is not checked as a caller-provided value before INSERT. Explicit nonzero values retain validation. Ordinary required fields and manually assigned primary keys retain their constraints; update/query/delete behavior is unchanged. This rule applies to local and remote models without application-specific logic.
