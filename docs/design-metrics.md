# 监控与指标设计

**功能块**：`metrics/` + magicCommon/monitoring  
**依据**：`metrics/metrics.go`、`metrics/orm/collector.go`、`metrics/orm/provider.go`、`orm/orm.go` 中 `registerORMMetrics`。

[← 返回设计文档索引](README.md)

---

## 1. 当前实现

### 1.1 metrics 包

- 定义 `OperationType`：insert、update、query、delete、create、drop、count、batch。
- 定义 `QueryType`、`ErrorType`、`DefaultLabels()` 等。

### 1.2 metrics/orm

- **ORMMetricsCollector**：提供 `RecordOperation(operation, model, duration, err)`、`RecordTransaction(txType, success)` 等。
- Orm 各操作在实现内部上报到全局 `ormMetricCollector`。

### 1.3 注册

- `orm.Initialize()` 时创建 `ORMMetricsCollector`；若存在 `monitoring.GetGlobalManager()` 则注册 `metricsorm.NewORMMetricProvider(collector)` 为全局 Provider（名称 `magicorm_orm`）。
- 可后续调用 `orm.EnsureORMMetricProviderRegistered()` 做幂等注册。

---

## 2. 与 README 的差异

README「监控系统」一节中描述的 `monitoring` 包（如 `monitoring.NewCollector()`、`RecordORMOperation`、`MonitoredOrm`）与当前代码**不一致**。当前无独立 `monitoring` 包，实际为 **metrics 包 + 内置 collector + magicCommon/monitoring 的 Provider 注册**。核对建议见 [design-checklist.md](design-checklist.md)。

---

## 3. METRICS_TODO

见项目根目录 [METRICS_TODO.md](../METRICS_TODO.md)：ORM 指标已接入；DB Metrics、Validation Metrics 生产接入及默认标签/文档为待办。
