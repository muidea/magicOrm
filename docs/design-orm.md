# ORM 核心接口设计

**功能块**：`~/codespace/magicOrm/orm`
**依据**：`~/codespace/magicOrm/orm/orm.go` 中 `Orm` 接口及 `impl` 实现。

[← 返回设计文档索引](~/codespace/magicOrm/docs/README.md)

---

## 1. 接口清单

| 方法 | 签名 | 说明 |
|------|------|------|
| Create | `Create(entity models.Model) *cd.Error` | 创建表（含关联表） |
| Drop | `Drop(entity models.Model) *cd.Error` | 删除表 |
| InspectSchema | `InspectSchema(model, expectedExist) (*SchemaReport, *cd.Error)` | 核验完整管理图 |
| RepairSchema | `RepairSchema(previous, desired) (*SchemaReport, *cd.Error)` | 按冻结声明恢复并核验 |
| Reconcile | `Reconcile(previous, current models.Model) *cd.Error` | 仅执行安全的增量 schema 演进 |
| Insert | `Insert(entity models.Model) (models.Model, *cd.Error)` | 插入单条，返回带主键的 Model |
| Update | `Update(entity models.Model) (models.Model, *cd.Error)` | 更新单条 |
| UpdateWithFilter | `UpdateWithFilter(entity, filter) (int64, *cd.Error)` | 单表条件更新，返回受影响行数 |
| Delete | `Delete(entity models.Model) (models.Model, *cd.Error)` | 删除单条 |
| Query | `Query(entity models.Model) (models.Model, *cd.Error)` | 按模型已赋值字段查询单条 |
| Count | `Count(filter models.Filter) (int64, *cd.Error)` | 按条件计数 |
| BatchQuery | `BatchQuery(filter models.Filter) ([]models.Model, *cd.Error)` | 按条件批量查询 |
| BeginTransaction | `BeginTransaction() *cd.Error` | 开启事务（当前 Orm 实例） |
| CommitTransaction | `CommitTransaction() *cd.Error` | 提交事务 |
| RollbackTransaction | `RollbackTransaction() *cd.Error` | 回滚事务 |
| Release | `Release()` | 释放资源 |

---

## 2. 设计说明

### 2.1 事务

- 无 `Begin() (Tx, error)` 形式返回值。
- 在**同一 Orm 实例**上先 `BeginTransaction()`，再执行若干 Insert/Update/Delete/Query 等，最后 `CommitTransaction()` 或 `RollbackTransaction()`。
- Insert/Delete 与关系 Update 配对开启、结束逻辑事务；纯主表 Update 可直接执行单条 SQL。显式事务将多次操作包进同一物理事务。
- 嵌套 Begin 共享物理事务，不创建 savepoint。内层 Rollback 标记 rollback-only；拒绝继续 Begin/Query/Execute，外层 Commit 明确失败并在最外层回滚。空闲 Commit 返回 InvalidOperation，空闲 Rollback 幂等。
- 每个成功 Begin 必须恰好结束一次。只有最外层可无条件 defer Rollback，延迟函数执行前不能在同一 handler 开启下一笔事务；内层 defer 使用结束标记，物理提交错误同样已结束作用域。Release 幂等，释放后不能再使用 handler。
- Commit 连接错误保留未知结果，ORM 不重发业务写入。 owner 在新事务/连接按原身份回查；失败事务中不得继续业务查询。`BeginSerializableTransaction` 只接受空闲 handler。
- 注册已有 owner 时先校验配置再增加引用；同 owner 注册/注销串行，不同 owner 使用分段锁。`OwnerPoolStats` 为只读快照，不授予连接或写权限。

### 2.2 验证

- **Insert、Update、Delete**：会调用内部 `validateModel(model, scenario)`，对应场景见 [design-validation.md](~/codespace/magicOrm/docs/design-validation.md)。
- **Query 与 BatchQuery**：不调用验证。

### 2.3 未在 Orm 接口暴露的能力

- **表是否存在**：底层 `database.Executor` 提供 `CheckTableExist(tableName string) (bool, *cd.Error)`，但 Orm 接口**无** `Exist(model)`。

以下可选能力通过包级函数调用，基础 Orm 接口不要求第三方实现同名方法：`orm.BeginSerializableTransaction(handler)`、`orm.LockWithFilter(handler, model, filter)`、`orm.Aggregate(handler, filter, plan)`。不具备能力时明确返回错误，不静默降级。`ValidateSchema(ctx, model, provider, prefix)` 和 `SchemaTables(...)` 是离线声明工具，不证明物理存在或授予写入权限。

### 2.4 对外查询接口稳定约定

- `Query(model)` 是 `magicOrm` 对外的正式单查接口，面向“针对模型对象查询单条结果”的业务语义。
- `BatchQuery(filter)` 是 `magicOrm` 对外的正式多查接口，面向“按过滤条件查询多条结果”的业务语义。
- ORM 层不再暴露额外的“按 filter 单查”入口，避免与 `Query(model)` 形成重叠的单查语义。
- 若上层框架需要“显式 filter 的单查封装”，应在 helper/业务适配层完成，而不是扩展 `orm.Orm` 公共接口。

### 2.5 Create / Drop 与关联表

- **Create/Drop/Inspect/Repair**：统一处理宿主、需要链接表的关系与递归 owned 表；单值内联引用不建链接表。普通引用目标主表不进入管理图。先完整校验图再构建 DDL；同图共享 owned 去重，owned 环和物理表重名拒绝。
- **Drop**：删除整个 owned 管理图。直接调用者须证明物理命名空间独占、检查全部图外使用者并串行 Schema 写入；Base 单根删除在跨根 owned 共享、独立实体入口或入向引用存在时拒绝。ORM 不承担平台目录与协调日志。`SchemaTables` 离线返回实际管理图表名，不授予删除权限。

### 2.6 Schema 演进

`Reconcile(previous, current)` 只允许不会丢失既有数据的增量变更：新增可空基础字段、新增关系表、新增唯一约束和新增普通索引。它不会猜测字段重命名，也不会执行回填或删除数据。

基础 `Reconcile` 的重复 DDL 行为按数据库实现区分：PostgreSQL 对新增命名唯一约束使用目录检查，重复执行同一约束 DDL 可跳过；MySQL 不提供该保护，重复新增约束会返回数据库错误，不能将重复错误直接视为恢复成功。DDL 结果未知时，应由调用者验证操作身份、冻结声明并串行执行 `RepairSchema`，通过物理检查确认已完成步骤和补执行范围；MySQL 物理检查要求 8.0.13+。此边界不承诺通用迁移幂等性，也不改变唯一约束或字段存储类型。

以下变更会在执行 SQL 前返回 `InvalidOperation`（消息以 `schema migration required:` 开头）：删除或重命名字段、修改字段类型/主键/数据库存储默认值、增加非空字段、修改或删除既有唯一约束/索引。调用方必须显式处理这类迁移；若上层采用“按声明重建 schema”的策略，应由上层完成重建和元数据更新。

默认声明变化先使用当前数据库 builder 的列描述核对物理结构。基础字段的类型、可空性、主键合同不变，且生成的完整列描述一致时，允许只更新运行期默认声明，不执行 DDL。生成的 SQL DEFAULT 或其他列属性不同仍要求显式迁移；不能通过忽略所有默认值变化绕过检查。此规则不解析业务身份表达式，也不放宽实体值校验。

### 2.7 条件更新

`UpdateWithFilter` 用于 compare-and-set 状态迁移，例如一次性消费授权码或轮换 refresh token。它把已赋值的可写主表字段（基础字段及内联单值引用）与 `Filter` 一起编译为一条 `UPDATE ... WHERE ...`，并返回数据库实际影响的行数。

- 受影响行数为 `0` 表示条件未满足，调用方应按业务语义拒绝或重试；
- 更新字段和过滤字段只接受主表字段，包括参与组合唯一而内联的单值引用；其余关系字段拒绝；
- `Filter` 必须绑定同一模型且至少包含一个条件，避免无条件批量更新；
- 显式事务开启时，该更新参与当前 Orm 实例的事务。

### 2.8 运行路径

- **Insert**：`validateModel` -> `InsertRunner` -> 先写 host，再写 relation，必要时回填主键和默认值声明。
- **Update**：`validateModel` -> `UpdateRunner` -> 先更新 host，再按关系类型刷新 relation。
  - 引用关系：普通链接刷新差集；内联单值引用直接更新宿主列，均不更新目标实体。
  - 包含关系：先比较数据库当前值与本次输入；未变化直接跳过。
  - 单值包含关系：若新旧子对象主键相同，则对子对象走原地 `Update`；否则删除旧子对象并重建关系。
  - 集合包含关系：优先按子对象主键做增删改；无法稳定识别主键时，回退到整组替换。
  - Local patch model 若通过 `Model.Copy(models.MetaView)` 后再 `SetFieldValue(...)` 构造，可以表达“显式 zero”和“显式 typed nil”；未赋值字段仍会跳过。
  - 但直接从原始 Go struct 构造本地更新模型时，`nil` 与“未提供字段”仍无法彻底区分，这是当前 local provider 的已知边界。
  - Remote 字段只有在“显式赋值”或“非零值”时才参与更新；helper 导出的默认零值会被跳过。
  - Remote 单值引用若显式赋值为 `nil`，表示清空关系；协议上要求字段为 `FieldValue{Assigned:true, Value:nil}`。若只是未赋值 `nil`，则跳过更新。
- **Delete**：`validateModel` -> `DeleteRunner` -> 先删 relation，再删 host。
- **Query**：`QueryRunner` 先按查询模型生成过滤条件，再使用 query mask 拉取 host 行，随后按字段加载 relation，最后回填 `models.Model`。
  - `Query(model)` 的输入模型用于“过滤”；
  - 若主键字段已赋值，则仅按主键过滤；
  - 若主键未赋值，则会把其它“已赋值字段”隐式转成 `AND` 条件；
  - relation 字段参与过滤时，会先压缩成关联对象主键；若 relation 已赋值但其主键未赋值，会直接返回 `IllegalParam`，不会再静默编码成零值；
  - 单条 `Query(model)` 命中多条记录时返回 `Unexpected`，不会默认取第一条；
  - 查询执行时仍会补齐足够的 basic 字段与 relation key，以保证回填完整性；
  - 最终返回给调用方的模型按以下稳定规则裁剪：
    - `Query(model)` 不处理 `ValueMask`，顶层对象固定按 `DetailView` 返回；
    - `BatchQuery(filter)` 若指定了 `filter.ValueMask(...)`，则顶层对象由 `ValueMask` 决定返回字段，`view` 不再参与顶层字段裁剪；
    - `BatchQuery(filter)` 未指定 `ValueMask(...)` 时，顶层对象按过滤器绑定模型的当前 `view` 决定返回字段；
    - `BatchQuery(filter)` 的顶层对象稳定优先级为：`ValueMask > view`；
    - 主键字段始终保留；
  - 子对象（包含/引用 relation）按以下稳定规则处理：
    - 只要子对象字段被纳入顶层响应，子对象本身统一按其自身 `lite` 视图返回；
    - 父对象为 `detail` 不会放大子对象为 `detail`；
    - 顶层 `ValueMask` 里的嵌套子对象结构只用于表达“是否包含该 relation 字段”，不会放大子对象层级；
  - 因此业务侧若需要子对象的详细信息，应先拿到子对象 `id` / lite 信息，再发起单独查询；不应依赖一次查询把多层对象同时放大到 detail；
  - 当前实现优先修正“返回字段语义”，尚未把 SQL `SELECT` 列彻底缩减到与 `ValueMask/View` 完全一致。
- **BatchQuery / Count**：基于 `models.Filter` 走 builder 生成 SQL。

### 2.9 事务与资源

- **事务**：在同一 Orm 实例上顺序调用 `BeginTransaction()` → 若干 Insert/Update/Delete/Query → `CommitTransaction()` 或 `RollbackTransaction()`；无返回 `Tx` 的 API。事务隔离级别、超时、死锁等按**数据库与 context 默认值**处理，当前不提供单独配置项。
- **单次 CRUD 与事务**：Insert/Delete 与关系 Update 自行配对事务作用域；纯主表 Update 可直接执行单条 SQL。已有显式事务时只结束自己的嵌套作用域，不提前提交调用者的事务；独立调用无需额外 Commit/Rollback。
- **并发**：同一 Orm 实例的**并发安全由外部调用方保证**（如单 goroutine 使用或由调用方加锁）；框架不在此层做并发保护。
- **Release**：释放 handler 的执行器连接及校验管理器，幂等；未结束的物理事务回滚。Release 不是成功提交证明，调用者须先确认事务结果；释放后不可继续使用 handler。连接池绑定通过对应 DelDatabase 释放。

### 2.10 错误处理

- 各方法返回 `*cd.Error`，错误码与含义见 [error-codes.md](~/codespace/magicOrm/docs/error-codes.md)。常见为 `IllegalParam`（参数非法）、`NotFound`（Query 无匹配）。

---

## 3. 完整接口定义（附录）

```go
type Orm interface {
    Create(entity models.Model) *cd.Error
    Drop(entity models.Model) *cd.Error
    Reconcile(previous models.Model, current models.Model) *cd.Error
    InspectSchema(entity models.Model, expectedExist bool) (*SchemaReport, *cd.Error)
    RepairSchema(previous models.Model, desired models.Model) (*SchemaReport, *cd.Error)
    Insert(entity models.Model) (models.Model, *cd.Error)
    Update(entity models.Model) (models.Model, *cd.Error)
    UpdateWithFilter(entity models.Model, filter models.Filter) (int64, *cd.Error)
    Delete(entity models.Model) (models.Model, *cd.Error)
    Query(entity models.Model) (models.Model, *cd.Error)
    Count(filter models.Filter) (int64, *cd.Error)
    BatchQuery(filter models.Filter) ([]models.Model, *cd.Error)
    BeginTransaction() *cd.Error
    CommitTransaction() *cd.Error
    RollbackTransaction() *cd.Error
    Release()
}
```


## 公共边界整理（2026-10-26）

本次整理以保持已有业务调用方式和特性为约束。`models.ORM` 是操作合同的唯一声明，`orm.Orm` 为类型别名；不新增必选方法。`orm.SchemaReport` 是 `models.SchemaReport` 的别名，保留结果字段与 JSON。`models.SchemaIssue`／`models.PoolStats` 保留原观察数据，`models.DatabaseConfig` 保留 Server／Username／Password／Database／Schema／GetDsn 方法。诊断字段继续直接读取，不要求调用方改成 getter。连接配置不包含 SQL 句柄。

公开包函数18项：GetOrm、NewOrm、BeginSerializableTransaction、LockWithFilter、Aggregate、ValidateSchema、PreflightSchemaChange、SchemaTables、SchemaSetTables、SchemaTableName、Initialize、Uninitialized、AddDatabase、DelDatabase、EnsureSchema、NewConfig、OwnerPoolStats、EnsureORMMetricProviderRegistered。

`SchemaTableName(model, prefix)` 只提供宿主表身份，供管理 owner 判断图外引用，不能用完整 owned 图或排序后首项代替。`SchemaTables` 继续返回经校验的完整管理图。两者不提供 SQL，也不是所有权授权。

数据库、Codec、Builder、Executor、Pool 全部位于 `~/codespace/magicOrm/internal/database/`；无旧包转发或公共执行器注入。八类 Runner 及三个底层工厂改为包内实现，不能绕过 Orm 的统一校验／事务入口。`impl` 不再导出内部 context 检查或可变 ValidationManager 配置方法，三项正式可选能力保留。独立的 Provider 校验扩展及现有 Provider 构造／声明 DTO 合同保持，不因内部整理更改远端协议。

Base 的初始化和模型测试验证正式模型合同及 owned 图；Runner 的页面测试验证冻结计划。参数化 SQL、投影、指标样本数及桶编译继续在 ORM owner 内部测试。Base 的真实 PostgreSQL 初始化、日志及派生恢复回归保留。测试基础设施 helper 只创建并清理本次独占的新 Schema，要求 magicorm_test_ 数据库；不向业务测试返回通用 SQL 句柄，业务表读写仍走 ORM。

这是源码 API 边界整理，无业务表 Schema、模型包摘要、安装身份、授权目录、写入错误分类或恢复策略变更。旧底层源码导入的迁移属于本次确认范围；已有业务门面保持使用方式，不能宣称所有旧底层导入仍兼容。外部消费者须同步构建新的依赖，不需要重建业务数据库。

## 完整模型集合的 Schema 生命周期（2026-10-29 源码候选）

第三方实现的必选 `models.ORM` 保持不变。可选接口 `models.SchemaSetLifecycle` 声明 `InspectSchemaSet(roots []Model, expectedExist bool)` 和 `DropSchemaSet(roots []Model)`，返回既有 `SchemaReport`；默认 ORM 实现支持此接口。调用方使用类型断言核对能力，不支持时明确拒绝，不循环调用单根 Drop 作为降级。

`orm.SchemaSetTables(ctx, roots, provider, prefix)` 离线返回经核验的集合物理表并集；共享 owned 模型及其独立根去重，相同模型的冲突声明、owned 环、重复根和物理冲突拒绝。独立引用目标未作为根传入时，不属于删除集合。完整声明、DDL 构建及现存表结构预检在首个删除前完成；关系表先于宿主删除。重试允许目标表已缺失，但不能接受现存表漂移；最后独立核对全部目标表缺失。

调用者仍须证明命名空间独占、完整根集合、图外使用者保护和 Schema 写入串行；ORM 不接受平台安装身份或推断租户权限。此能力不删除数据库、物理 Schema 容器或账号。源码默认／MySQL-tag 回归及隔离 PostgreSQL 物理测试与平台 Live 分开记账。
