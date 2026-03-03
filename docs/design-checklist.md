# 与现有文档差异及实现核对清单

以下为与 README/现有文档的差异及建议核对项，便于发现缺陷与遗漏。

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
| **现象** | README 示例为 `tx, err := o1.Begin()` 与 `tx.Insert(...)`；实际 API 无返回值 tx，为 `BeginTransaction()/CommitTransaction()/RollbackTransaction()`，操作均在**同一 Orm 实例**上。 |
| **建议** | 设计文档以当前 API 为准；在「与 README 差异」中注明，并建议更新 README 事务示例为上述顺序调用形式。 |
| **核对** | [ ] 已按当前 API 修改 README 事务小节。 |

---

## 3. 监控文档与实现不符

| 项目 | 说明 |
|------|------|
| **现象** | README「监控系统」使用 `monitoring` 包、`NewCollector()`、`RecordORMOperation`、`MonitoredOrm` 等；实际为 `metrics` 包 + 内置 `ORMMetricsCollector` + magicCommon/monitoring 的 Provider 注册，无独立 `monitoring` 包与 MonitoredOrm 包装。 |
| **建议** | 设计文档描述当前 metrics 与注册方式；标注 README 该节已过时或为旧设计，需同步或删除/重写。 |
| **核对** | [ ] README 监控一节已改为描述 metrics + magicCommon 注册，或已注明「见本目录 [README.md](README.md)（设计总览）与 [METRICS_TODO.md](../METRICS_TODO.md)」。 |

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
