# 验证系统设计

**功能块**：`validation/`  
**依据**：`validation/manager.go`、`validation/scenario.go`、`validation/constraints/`、`validation/types/`、`validation/database/`、`validation/errors/`。

[← 返回设计文档索引](README.md)

---

## 1. 四层架构

与 README、VALIDATION_ARCHITECTURE.md 一致：

| 层 | 包 | 职责 |
|----|-----|------|
| 类型层 | types | Go 类型与数据库类型兼容性、基础转换。 |
| 约束层 | constraints | 结构体标签中的业务规则（req, min, max, range, in, re, ro, wo）。 |
| 数据库层 | database | 库级约束（NOT NULL、UNIQUE、FOREIGN KEY 等）及库类型兼容。 |
| 场景层 | scenario | 按 Insert/Update/Query/Delete 编排策略。 |

---

## 2. 场景与 Orm 操作对应

| 场景 | Orm 操作 | 是否调用 validateModel |
|------|----------|-------------------------|
| ScenarioInsert | Insert | 是 |
| ScenarioUpdate | Update | 是 |
| ScenarioDelete | Delete | 是 |
| ScenarioQuery | Query / BatchQuery | **否** |

入口：`ValidationManager.ValidateModel(model, context)`。Orm 通过内部 `validateModel(model, scenario)` 传入 `errors.Scenario*` 与 `validation.OperationType`（Create/Read/Update/Delete）。约束定义见 [design-models.md](design-models.md)。

---

## 3. 配置

- `DefaultConfig()` / `SimpleConfig()`：与 README 一致。
- 各层开关、EnableCaching、CacheTTL、StopOnFirstError 等见 `validation/manager.go` 与 README 验证配置小节。
