# 数据库层设计

**功能块**：`database/`  
**依据**：`database/executor.go`、`database/postgres/`、`database/mysql/`。

[← 返回设计文档索引](README.md)

---

## 1. Executor 接口

| 方法 | 说明 |
|------|------|
| Query(sql, needCols, args...) | 执行查询，返回列名与行迭代 |
| Next / GetField(value...) / Finish | 行迭代与取字段 |
| Execute(sql, args...) | 执行 DML/DDL，返回影响行数 |
| ExecuteInsert(sql, pkValOut, args...) | 插入并返回主键 |
| CheckTableExist(tableName) | 检查表是否存在（**未在 Orm 暴露**） |
| BeginTransaction / CommitTransaction / RollbackTransaction | 事务 |
| Release | 释放连接 |

Orm 通过 Runner（如 InsertRunner、QueryRunner）调用 Executor，不直接暴露 `CheckTableExist`。若需在 Orm 层提供「表是否存在」能力，需在此层之上封装，见 [design-checklist.md](design-checklist.md)。

---

## 2. Pool 接口

| 方法 | 说明 |
|------|------|
| Initialize(maxConnNum, config) | 初始化连接池 |
| GetExecutor(ctx) | 获取 Executor |
| CheckConfig(config) | 校验配置 |
| IncReference / DecReference | 引用计数 |
| Uninitialized | 反初始化 |

---

## 3. Config 接口

提供：`Server()`、`Username()`、`Password()`、`Database()`、`GetDsn()`。

---

## 4. 实现

- **PostgreSQL**：`database/postgres/`
- **MySQL**：`database/mysql/`
