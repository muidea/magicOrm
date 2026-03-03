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
