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

## 3. 指标与标签

- **DefaultLabels()**：返回 `component: "magicorm"`、`version: "1.0.0"`，用于与 magicCommon/monitoring 集成时打标。
- **自定义标签**：通过 `MergeLabels` 或 Provider 实现与监控系统对接时扩展。**需澄清**：时间指标精度（毫秒/微秒）、存储策略及监控对性能的影响是否在文档中约定，见 [需澄清信息.md](需澄清信息.md)。

---

## 4. METRICS_TODO

见项目根目录 [METRICS_TODO.md](../METRICS_TODO.md)：ORM 指标已接入；DB Metrics、Validation Metrics 生产接入及默认标签/文档为待办（评审 CONS-004）。

---

## 5. 当前可用范围说明

- **已在生产可用的部分**：
  - ORM 操作指标（Insert/Update/Delete/Query/BatchQuery/Create/Drop/Count），通过 `metrics/orm` + `orm.Initialize()` 在存在 GlobalManager 时自动注册为 Provider（`magicorm_orm`）；
  - 默认标签与基础延迟/错误统计已在现有实现中使用，具体字段以 `metrics` 包定义为准。
- **仍在规划/待办的部分**（以 [METRICS_TODO.md](../METRICS_TODO.md) 为准）：
  - 数据库层指标（Query/Execute/事务级别的 DB Metrics）；
  - 验证层指标（按场景与验证层拆分的 Validation Metrics）；
  - 更细粒度的标签约定与对接不同监控系统的最佳实践。

在这些待办完成前，可以将本设计文档理解为：**ORM Metrics 为当前稳定能力，DB/Validation Metrics 为明确规划的扩展能力**。
