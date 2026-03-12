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
| Query | `Query(entity models.Model) (models.Model, *cd.Error)` | 按主键查询单条 |
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
- 详细用法见 [design-data-flow.md](design-data-flow.md)。

### 2.2 验证

- **Insert、Update、Delete**：会调用内部 `validateModel(model, scenario)`，对应场景见 [design-validation.md](design-validation.md)。
- **Query 与 BatchQuery**：不调用验证；若需核对或扩展见 [design-checklist.md](design-checklist.md)。

### 2.3 未在 Orm 接口暴露的能力

- **表是否存在**：底层 `database.Executor` 提供 `CheckTableExist(tableName string) (bool, *cd.Error)`，但 Orm 接口**无** `Exist(model)`。README 中曾描述该能力，现状与建议见 [design-checklist.md](design-checklist.md)。

### 2.4 Create / Drop 与关联表

- **Create**：会递归创建该实体表及其**所有**关系字段对应的关系表；先创建依赖的包含关系实体表，再创建 host 表，再创建各关系表。关联表创建顺序与依赖解析见 [design-relation.md](design-relation.md)。
- **Drop**：仅处理**当前 Model 对应的数据表**（即该实体表 + 以该实体为 host 的关系表）；不级联删除其它实体表或对端实体表中的数据，由调用方按需自行删除。

### 2.5 事务与资源

- **事务**：在同一 Orm 实例上顺序调用 `BeginTransaction()` → 若干 Insert/Update/Delete/Query → `CommitTransaction()` 或 `RollbackTransaction()`；无返回 `Tx` 的 API。事务隔离级别、超时、死锁等按**数据库与 context 默认值**处理，当前不提供单独配置项。
- **单次 CRUD 与事务**：每次 Insert/Update/Delete 等操作在实现上均在同一事务内完成，并在当次操作结束时自动提交或回滚（成功则提交，失败则回滚），无需调用方在单次 CRUD 后显式 Commit/Rollback。
- **并发**：同一 Orm 实例的**并发安全由外部调用方保证**（如单 goroutine 使用或由调用方加锁）；框架不在此层做并发保护。
- **Release**：释放 Orm 占用的资源（如连接池引用）；应在使用完毕后调用。因每次 CRUD 都会在当次操作内完成提交或回滚，正常情况下不存在「未提交事务」；若在已调用 `BeginTransaction()` 且未 `CommitTransaction()`/`RollbackTransaction()` 的情况下调用 Release，行为以实现为准，建议调用方保证事务在 Release 前已结束。

### 2.6 错误处理

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
