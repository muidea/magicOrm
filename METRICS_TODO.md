## MagicORM Metrics 改造跟踪

### 已完成（本轮）

- **统一错误类型与分类逻辑**
  - `metrics/orm/collector.go`、`metrics/metricsdb/collector.go`、`metrics/validation/collector.go`：
    - `classifyError` 统一使用 `metrics.ErrorTypeXXX` 常量作为返回值。
    - 对 `err.Error()` 做 `strings.ToLower` 再匹配关键字，提升鲁棒性。
    - 对 `nil` 或 `Error()` 失败场景统一归类为 `ErrorTypeUnknown`。
  - 单测同步调整并通过：
    - `metrics/metricsdb/collector_test.go`
    - `metrics/validation/collector_test.go`

- **统一 ORM 操作类型常量**
  - 在 `orm` 包中，所有 ORM 操作埋点统一使用 `metrics.OperationType` 常量：
    - `Create` 使用 `OperationCreate`
    - `Drop` 使用 `OperationDrop`
    - `Insert` 使用 `OperationInsert`
    - `Update` 使用 `OperationUpdate`
    - `Delete` 使用 `OperationDelete`
    - `Query` 使用 `OperationQuery`
    - `BatchQuery` 使用 `OperationBatch`
    - `Count` 使用 `OperationCount`
  - 涉及文件：`orm/query.go`、`orm/insert.go`、`orm/update.go`、`orm/delete.go`、`orm/count.go`、`orm/filter.go`、`orm/create.go`、`orm/drop.go`。

- **ORM metrics 注册增强**
  - `orm/orm.go`：
    - 导入 `metrics/orm` 使用别名 `metricsorm`，避免与包名冲突。
    - `registerORMMetrics` 在注册失败时输出 `slog.Warn("Failed to register ORM metrics provider", "error", err.Error())` 日志。
    - 新增幂等辅助函数 `EnsureORMMetricProviderRegistered()`：
      - 当 `ormMetricCollector` 已存在、`ormMetricProvider` 为空且 `monitoring.GetGlobalManager() != nil` 时，创建并注册 provider。
      - 失败输出 Warn 日志，成功输出 Info 日志。
  - 已通过 `go test ./metrics/... ./orm/...` 验证。

### 待办（后续再处理）

- **DB Metrics（`metrics/metricsdb`）生产接入**
  - 在数据库执行层（`database` 包）统一封装：
    - 查询类调用 `RecordQuery(database, queryType, duration, err)`。
    - DML/DDL 调用 `RecordExecution(database, operation, success)`。
    - 事务 Begin/Commit/Rollback 调用 `RecordTransaction(database, txType, success)`。
    - 定期从连接池统计中调用 `UpdateConnectionStats(database, state, count)`。
  - 在对外初始化路径中增加一次性调用 `metricsdb.RegisterDatabaseMetrics()`，并在文档中约束初始化顺序（需在 monitoring 初始化完成后调用，或配合单独的 Ensure 方法）。

- **Validation Metrics（`metrics/validation`）生产接入**
  - 在 `validation.Manager` 内部统一接入：
    - 在 `ValidateModel` / `Validate` 周围记录 `RecordValidation(operation, model, scenario, duration, err)`。
    - 在缓存模块中记录 `RecordCacheAccess(cacheType, hit)`，并确保 `GetCacheHitRatio` 反映真实数据。
    - 在各类约束检查实现处记录 `RecordConstraintCheck(constraintType, field, passed)`。
  - 在验证系统初始化入口增加 `RegisterValidationMetrics()` 调用（或类似 Ensure 机制），与 monitoring 初始化时序对齐。

- **统一默认标签与文档**
  - 根据需要，将各 Provider 的指标标签补充使用 `metrics.DefaultLabels()`（如 `component="magicorm"`, `version` 等），统一监控视图。
  - 在 README/AGENTS 文档中补充：
    - metrics 初始化与 `monitoring.InitializeGlobalManager()` 的推荐顺序。
    - ORM / DB / Validation 三类 metrics 的启用方式与当前成熟度（ORM 已接入，DB/Validation 待完善）。

