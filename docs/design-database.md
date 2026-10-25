# 数据库层设计

**功能块**：`~/codespace/magicOrm/database`
**依据**：`~/codespace/magicOrm/database/executor.go`、`~/codespace/magicOrm/database/postgres`、`~/codespace/magicOrm/database/mysql`。

[← 返回设计文档索引](~/codespace/magicOrm/docs/README.md)

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

Orm 通过 Runner（如 InsertRunner、QueryRunner）调用 Executor，不直接暴露 `CheckTableExist`。如需对外提供“表是否存在”能力，需要在 Orm 层另外封装。

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

提供：`Server()`、`Username()`、`Password()`、`Database()`、`Schema()`、`GetDsn()`。

### 3.1 DSN 格式（GetDsn）

- **PostgreSQL**：`postgres://{user}:{password}@{server}/{dbName}?sslmode={sslmode}&options=-c%20search_path={schema}`。Database 与 Schema 必须通过独立字段提供；执行器不解析组合名称，也不会缺省创建 Schema。
- **MySQL**：`{user}:{password}@tcp({server})/{dbName}?charset={charset}`，charset 由 Config 实现（如 utf8mb4）。
- **可选参数**（如 sslmode、charset 等）：**直接使用驱动的默认值**，框架不单独提供配置项。

---

## 4. 连接池与连接管理

- **maxConnNum**：**由调用方传入**（如 `orm.AddDatabase(..., maxConnNum, owner)`），框架内无默认值；底层使用 `db.SetMaxOpenConns(maxConnNum)`。
- **连接生命周期**：连接由标准库 `database/sql` 管理；获取 Executor 时从 Pool 取连接，Release 时归还。**超时、重连、健康检查**：当前**依赖数据库驱动的默认行为**，不单独配置。

---

## 5. 实现

- **PostgreSQL**：`~/codespace/magicOrm/database/postgres`
- **MySQL**：`~/codespace/magicOrm/database/mysql`

---

## 6. Schema 元数据与演进

表结构仍由 Model 声明生成，但 dynamic remote Model 现在可以同时声明：

- `models.UniqueConstraint{Name, Fields}`：数据库强制的唯一键；
- `models.Index{Name, Fields}`：普通索引。

普通索引声明仅能引用当前模型的基础字段；组合唯一另支持内联单值引用。声明名称和字段名必须是安全标识符；它们不是运行时 SQL 片段，也不是查询过滤条件。`Create` 会创建声明中的唯一约束和索引；`Reconcile` 只补建新声明，不会修改或删除已存在的声明。

PostgreSQL 的普通索引和唯一约束创建均支持安全重试：前者使用 `IF NOT EXISTS`，后者会检查系统目录后再添加约束。这覆盖“数据库 DDL 已成功、上层 schema 元数据尚未提交”时的重试窗口。具体可演进与必须显式迁移的边界见 [design-orm.md](~/codespace/magicOrm/docs/design-orm.md)。

Schema 由调用 owner 显式执行 `orm.EnsureSchema(server, database, schema, username, password)`。AddDatabase/NewOrm 不创建 Schema，也不解析 database/schema 合并字段。普通进程仅连接部署 owner 已准备的结构。

PostgreSQL UNIQUE 的底层索引及普通索引名在 Schema 内共享。物理名采用 table_declaration 的前至多 46 字节、下划线和 SHA-256(table + NUL + declaration) 前 8 字节的十六进制值，长度至多 63 字节；两段身份边界明确，不依赖下划线分割。物理表名超过 63 字节在预检时拒绝。Create、Reconcile、Describe/Inspect、Repair 共用命名映射，MySQL 保留表内声明名。旧未限定物理名不会自动迁移；本批只部署干净候选 Schema。

`UpdateWithFilter` 是与 schema 无关的单表条件 DML builder：它只接受主表字段，包括内联单值引用，供上层完成 compare-and-set 状态转换；详细语义见 [design-orm.md](~/codespace/magicOrm/docs/design-orm.md)。
