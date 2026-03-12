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

### 3.1 DSN 格式（GetDsn）

- **PostgreSQL**：`postgres://{user}:{password}@{server}/{dbName}?sslmode={sslmode}&options=-c%20search_path={schema}`。其中 `Database()` 可为 `databaseName` 或 `databaseName/schemaName`，后者指定非默认 schema（默认 `public`）。
- **MySQL**：`{user}:{password}@tcp({server})/{dbName}?charset={charset}`，charset 由 Config 实现（如 utf8mb4）。
- **可选参数**（如 sslmode、charset 等）：**直接使用驱动的默认值**，框架不单独提供配置项。

---

## 4. 连接池与连接管理

- **maxConnNum**：**由调用方传入**（如 `orm.AddDatabase(..., maxConnNum, owner)`），框架内无默认值；底层使用 `db.SetMaxOpenConns(maxConnNum)`。
- **连接生命周期**：连接由标准库 `database/sql` 管理；获取 Executor 时从 Pool 取连接，Release 时归还。**超时、重连、健康检查**：当前**依赖数据库驱动的默认行为**，不单独配置。

---

## 5. 实现

- **PostgreSQL**：`database/postgres/`
- **MySQL**：`database/mysql/`

---

## 6. 索引与其它（评审 FUNC-002）

- **索引**：当前设计文档与实现中，表结构由 Model 元数据生成，**未体现显式索引定义/创建 API**（如唯一索引、复合索引）。若需支持，属设计扩展项。**需澄清**：是否有计划支持在模型或 DDL 中声明索引，见 [需澄清信息.md](需澄清信息.md)。
