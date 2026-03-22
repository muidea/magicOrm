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

当前无剩余待办项。
