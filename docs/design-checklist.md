# 与现有文档差异及实现核对清单

以下为与 README/现有文档的差异及建议核对项，便于发现缺陷与遗漏。根据最新评审完成的设计文档补充见各 design-*.md、[需澄清信息.md](需澄清信息.md) 与 [待确认项清单.md](待确认项清单.md)（评审 CONS-001）。

[← 返回设计文档索引](README.md)

---

## 1. Exist 未暴露

| 项目 | 说明 |
|------|------|
| **现象** | README 写「检查表是否存在 exists, err := o1.Exist(model)」；Orm 接口**无** Exist 方法。 |
| **实现** | 仅 `database.Executor` 有 `CheckTableExist(tableName string) (bool, *cd.Error)`。 |
| **建议** | 设计文档明确：表存在检查当前未在 Orm 暴露；若需此能力，需在 Orm 层封装 Executor.CheckTableExist（根据 Model 解析表名再调用）。 |
| **核对** | [ ] 产品是否需要 Orm.Exist(model)；若需要则补实现并更新 README。 |

---

## 2. 事务 API 与 README 不一致

| 项目 | 说明 |
|------|------|
| **现象** | 早期 README 示例使用 `tx, err := o1.Begin()` 与 `tx.Insert(...)`；当前文档与实现均已统一为无返回值的 `BeginTransaction()/CommitTransaction()/RollbackTransaction()`，操作均在**同一 Orm 实例**上。 |
| **建议** | 设计文档与 README 均以当前 API 为准；旧示例仅在归档文档中保留，用于对比历史设计。 |
| **核对** | [x] README 事务示例已按当前 API 更新。 |

---

## 3. 监控文档与实现不符

| 项目 | 说明 |
|------|------|
| **现象** | 早期 README「监控系统」章节曾使用自有 `monitoring` 包（如 `NewCollector()`、`MonitoredOrm` 等）；当前实现与文档已统一为 `metrics` 包 + 内置 `ORMMetricsCollector` + magicCommon/monitoring 的 Provider 注册，无独立 `monitoring` 包与 MonitoredOrm 包装。 |
| **建议** | 以 `docs/design-metrics.md` 与当前 README 中的说明为准；旧版 monitoring 示例仅保留在归档/部署指南中，并在文首注明为历史示例。 |
| **核对** | [x] README 监控一节已改为描述 metrics + magicCommon 注册，并引用本目录设计文档与 `METRICS_TODO.md`。 |

---

## 4. Query / BatchQuery 无验证

| 项目 | 说明 |
|------|------|
| **现象** | Insert、Update、Delete 均调用 `validateModel(model, scenario)`；Query、BatchQuery **未**调用验证。 |
| **建议** | 在设计文档中明确「当前仅 Insert/Update/Delete 做模型验证」；若产品上希望查询前也做轻量校验，可列为待办或单独设计。 |
| **核对** | [ ] 已确认 Query/BatchQuery 不需验证，或已补充验证策略并实现。 |

---

## 5. METRICS_TODO 未闭环

| 项目 | 说明 |
|------|------|
| **现象** | METRICS_TODO.md 中 DB Metrics、Validation Metrics 生产接入、默认标签与文档等仍为待办。 |
| **建议** | 设计文档「监控」一节引用 METRICS_TODO，标明 ORM 已接入，DB/Validation 待完善。 |
| **核对** | [ ] 按 METRICS_TODO 推进 DB/Validation 接入与文档更新。 |
