# 数据流与关键场景

**说明**：本文档描述 Orm 核心操作的数据流与事务使用方式。

[← 返回设计文档索引](README.md)

---

## 1. Insert 流程

```mermaid
sequenceDiagram
    participant App
    participant Orm
    participant Validation
    participant InsertRunner
    participant Executor
    App->>Orm: Insert(model)
    Orm->>Orm: validateModel(model, ScenarioInsert)
    Orm->>Validation: ValidateModel(model, ctx)
    Validation-->>Orm: nil / error
    Orm->>Orm: BeginTransaction()
    Orm->>InsertRunner: Insert()
    InsertRunner->>Executor: ExecuteInsert (host) / Query+Execute (relations)
    Executor-->>InsertRunner: result
    InsertRunner-->>Orm: model with PK
    Orm->>Orm: CommitTransaction() or RollbackTransaction()
    Orm->>Orm: RecordOperation(insert, ...)
    Orm-->>App: model, err
```

---

## 2. Query 流程（单条，按主键）

```mermaid
sequenceDiagram
    participant App
    participant Orm
    participant QueryRunner
    participant Executor
    App->>Orm: Query(model)
    Note over Orm: 无 validateModel
    Orm->>Orm: Copy(OriginView), getModelFilter
    Orm->>QueryRunner: Query(filter)
    QueryRunner->>Executor: Query(sql,...)
    Executor-->>QueryRunner: rows
    QueryRunner->>QueryRunner: assignModelField / queryRelation
    QueryRunner-->>Orm: single model
    Orm->>Orm: RecordOperation(query, ...)
    Orm-->>App: model, err
```

---

## 3. 事务使用方式（当前 API）

Orm 接口定义见 [design-orm.md](design-orm.md)。在同一 Orm 实例上顺序调用，**无 tx 返回值**：

1. `o.BeginTransaction()`
2. `o.Insert(...)` / `o.Update(...)` / `o.Delete(...)` / `o.Query(...)` 等
3. `o.CommitTransaction()` 或 `o.RollbackTransaction()`

与 README 中曾出现的 `tx, err := o1.Begin()`、`tx.Insert(...)` 写法不一致，以当前 API 为准。详见 [design-checklist.md](design-checklist.md)。
