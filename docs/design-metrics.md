# 监控与指标设计

**功能块**：`metrics/`、`metrics/orm`、`metrics/metricsdb`、`metrics/validation`  
**依据**：`metrics/metrics.go`、`metrics/durationutil.go`、各 collector/provider 实现

[← 返回设计文档索引](README.md)

---

## 1. 当前实现概览

MagicORM 当前已经有三套可直接接入 `magicCommon/monitoring` 的指标提供者：

- `magicorm_orm`
- `magicorm_database`
- `magicorm_validation`

其中：

- `metrics/orm` 面向 ORM 操作、事务、缓存命中率、活跃连接数。
- `metrics/metricsdb` 面向数据库查询、执行、事务、连接池、错误分类。
- `metrics/validation` 面向校验操作、耗时、错误、缓存访问、约束检查。

各 provider 都通过内存 collector 聚合数据，再在 `Collect()` 时导出为监控指标。

---

## 2. 通用约定

### 2.1 指标键

- collector 内部统一通过 `metrics.BuildKey(...)` 组织多维键。
- provider 导出时通过 `metrics.ParseKey(...)` 还原标签。

### 2.2 默认标签

- `metrics.DefaultLabels()` 返回 `component=magicorm`、`version=1.0.0`。
- provider 还会额外打自身标签，例如 `orm`、`database`、`validation`。

### 2.3 耗时采样策略

- 耗时样本统一通过 `metrics.RecordDurationSample(...)` 写入。
- `metrics.AverageDurationSeconds(...)` 负责计算平均耗时秒数。
- 当前默认限制：
  - 最多保留 `1000` 个耗时键
  - 每个键最多保留 `1000` 个样本
- 超出上限时按 LRU 淘汰旧键，并对单键样本做滑动窗口截断，避免 collector 内存无界增长。

---

## 3. ORM 指标

### 3.1 Collector

`metrics/orm/collector.go` 当前维护：

- 操作计数：`operation + model + status`
- 操作耗时：`operation + model + status`
- 错误计数：`operation + model + error_type`
- 事务计数：`type + status`
- 缓存命中/未命中
- 活跃连接数

### 3.2 Provider

`metrics/orm/provider.go` 当前导出：

- `magicorm_orm_operations_total`
- `magicorm_orm_operation_duration_seconds`
- `magicorm_orm_errors_total`
- `magicorm_orm_transactions_total`
- `magicorm_orm_cache_hit_ratio`
- `magicorm_orm_active_connections`

说明：

- ORM 缓存命中率目前按 `default` 单一 cache type 导出。
- 平均耗时由 collector 中的采样窗口计算，不导出原始样本。

---

## 4. Database 指标

### 4.1 Collector

`metrics/metricsdb/collector.go` 当前维护：

- 查询计数：`database + query_type + status`
- 查询耗时：`database + query_type + status`
- 错误计数：`database + operation + error_type`
- 事务计数：`database + type + status`
- 执行计数：`database + operation + status`
- 连接池状态：`database + state`

错误类型当前按字符串做轻量分类，主要覆盖：

- `connection`
- `timeout`
- `database`
- `constraint`
- `unknown`

### 4.2 Provider

`metrics/metricsdb/provider.go` 当前导出：

- `magicorm_database_queries_total`
- `magicorm_database_query_duration_seconds`
- `magicorm_database_errors_total`
- `magicorm_database_transactions_total`
- `magicorm_database_executions_total`
- `magicorm_database_connections`

---

## 5. Validation 指标

### 5.1 Collector

`metrics/validation/collector.go` 当前维护：

- 校验计数：`operation + model + scenario + status`
- 校验耗时：`operation + model + scenario + status`
- 校验错误计数：`operation + model + scenario + error_type`
- 缓存访问计数：`cache_type + hit_miss`
- 约束检查计数：`constraint_type + field + status`

### 5.2 Provider

`metrics/validation/provider.go` 当前导出：

- `magicorm_validation_operations_total`
- `magicorm_validation_duration_seconds`
- `magicorm_validation_errors_total`
- `magicorm_validation_cache_access_total`
- `magicorm_validation_constraint_checks_total`
- `magicorm_validation_cache_hit_ratio`

说明：

- cache hit ratio 不再只导出 `default`，而是按 collector 中实际出现的 `cache_type` 动态导出。
- 当前常见 `cache_type` 包括 `type`、`constraint`，但实现并不限制具体取值。

---

## 6. 注册与使用

当前 metrics 会按能力分层初始化：

- ORM 层会在 `orm.Initialize()` 中创建 ORM collector，并在 `monitoring.GlobalManager` 可用时注册 `magicorm_orm` provider。
- Database 层也会在 `orm.Initialize()` 中创建 DB collector；MySQL/PostgreSQL executor 与 pool 当前会把 query / execute / transaction / connection stats 写入该 collector，并在 `monitoring.GlobalManager` 可用时注册 `magicorm_database` provider。
- Validation 层也会在 `orm.Initialize()` 中创建 validation collector；`validation.Manager`、`validation/cache` 与约束校验链路当前会把 validation / cache / constraint metrics 写入该 collector，并在 `monitoring.GlobalManager` 可用时注册 `magicorm_validation` provider。

如果调用方没有注册 monitoring manager：

- collector 仍可在进程内累计指标
- provider 不会自动对外暴露
- 后续可显式调用 `orm.EnsureORMMetricProviderRegistered()`、`metricsdb.EnsureDatabaseMetricProviderRegistered()` 或 `metricsvalidation.EnsureValidationMetricProviderRegistered()` 做幂等注册

---

## 7. 当前边界

- 当前导出的是平均耗时，不是分位数或直方图。
- 内存 collector 重启即丢失，不承担持久化职责。
- 错误分类是轻量字符串分类，不等同于数据库方言级错误码体系。
- ORM 缓存命中率目前仍按单一 `default` 维度导出；Validation 已支持多 cache type。

---

## 8. 维护建议

- 新增 metrics 维度时，先评估 key cardinality，避免高基数字段直接进入 `BuildKey(...)`。
- 新增 duration 指标时，优先复用 `RecordDurationSample(...)` 和 `AverageDurationSeconds(...)`。
- 新增 provider 指标时，优先补 collector 与 provider 的成对测试，确保标签和平均值语义一致。
