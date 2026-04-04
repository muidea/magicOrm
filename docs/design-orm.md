# ORM 核心接口设计

**功能块**：`orm/`  
**依据**：`orm/orm.go` 中 `Orm` 接口及 `impl` 实现。

[← 返回设计文档索引](README.md)

---

## 1. 接口清单

| 方法 | 签名 | 说明 |
|------|------|------|
| Create | `Create(entity models.Model) *cd.Error` | 创建表（含关联表） |
| Drop | `Drop(entity models.Model) *cd.Error` | 删除表 |
| Insert | `Insert(entity models.Model) (models.Model, *cd.Error)` | 插入单条，返回带主键的 Model |
| Update | `Update(entity models.Model) (models.Model, *cd.Error)` | 更新单条 |
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
- 单次 CRUD 在内部会开启并结束自己的事务；显式事务用于把多次操作包进同一 executor transaction。

### 2.2 验证

- **Insert、Update、Delete**：会调用内部 `validateModel(model, scenario)`，对应场景见 [design-validation.md](design-validation.md)。
- **Query 与 BatchQuery**：不调用验证。

### 2.3 未在 Orm 接口暴露的能力

- **表是否存在**：底层 `database.Executor` 提供 `CheckTableExist(tableName string) (bool, *cd.Error)`，但 Orm 接口**无** `Exist(model)`。

### 2.4 对外查询接口稳定约定

- `Query(model)` 是 `magicOrm` 对外的正式单查接口，面向“针对模型对象查询单条结果”的业务语义。
- `BatchQuery(filter)` 是 `magicOrm` 对外的正式多查接口，面向“按过滤条件查询多条结果”的业务语义。
- ORM 层不再暴露额外的“按 filter 单查”入口，避免与 `Query(model)` 形成重叠的单查语义。
- 若上层框架需要“显式 filter 的单查封装”，应在 helper/业务适配层完成，而不是扩展 `orm.Orm` 公共接口。

### 2.5 Create / Drop 与关联表

- **Create**：会递归创建该实体表及其**所有**关系字段对应的关系表；先创建依赖的包含关系实体表，再创建 host 表，再创建各关系表。关联表创建顺序与依赖解析见 [design-relation.md](design-relation.md)。
- **Drop**：仅处理**当前 Model 对应的数据表**（即该实体表 + 以该实体为 host 的关系表）；不级联删除其它实体表或对端实体表中的数据，由调用方按需自行删除。

### 2.6 运行路径

- **Insert**：`validateModel` -> `InsertRunner` -> 先写 host，再写 relation，必要时回填主键和默认值声明。
- **Update**：`validateModel` -> `UpdateRunner` -> 先更新 host，再按关系类型刷新 relation。
  - 引用关系：只刷新关系表差集，不更新对端实体。
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

### 2.7 事务与资源

- **事务**：在同一 Orm 实例上顺序调用 `BeginTransaction()` → 若干 Insert/Update/Delete/Query → `CommitTransaction()` 或 `RollbackTransaction()`；无返回 `Tx` 的 API。事务隔离级别、超时、死锁等按**数据库与 context 默认值**处理，当前不提供单独配置项。
- **单次 CRUD 与事务**：每次 Insert/Update/Delete 等操作在实现上均在同一事务内完成，并在当次操作结束时自动提交或回滚（成功则提交，失败则回滚），无需调用方在单次 CRUD 后显式 Commit/Rollback。
- **并发**：同一 Orm 实例的**并发安全由外部调用方保证**（如单 goroutine 使用或由调用方加锁）；框架不在此层做并发保护。
- **Release**：释放 Orm 占用的资源（如连接池引用）；应在使用完毕后调用。因每次 CRUD 都会在当次操作内完成提交或回滚，正常情况下不存在「未提交事务」；若在已调用 `BeginTransaction()` 且未 `CommitTransaction()`/`RollbackTransaction()` 的情况下调用 Release，行为以实现为准，建议调用方保证事务在 Release 前已结束。

### 2.8 错误处理

- 各方法返回 `*cd.Error`，错误码与含义见 [error-codes.md](error-codes.md)。常见为 `IllegalParam`（参数非法）、`NotFound`（Query 无匹配）。

---

## 3. 完整接口定义（附录）

```go
type Orm interface {
    Create(entity models.Model) *cd.Error
    Drop(entity models.Model) *cd.Error
    Insert(entity models.Model) (models.Model, *cd.Error)
    Update(entity models.Model) (models.Model, *cd.Error)
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
