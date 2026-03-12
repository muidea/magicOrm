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

### 2.1 四层之间的交互与数据流

- 验证入口为 `ValidateModel(model, context)`，context 中含场景（Insert/Update/Delete 等）与数据库类型。
- 执行顺序由场景适配器编排：先**类型层**（类型兼容与转换），再**约束层**（req/min/max/range/in/re/ro/wo 等），再**数据库层**（NOT NULL、库类型兼容等）。任一层失败即返回错误，是否继续后续层可由配置（如 StopOnFirstError）决定。
- 各层使用 Model 的字段与约束元数据，不直接访问数据库连接。

---

## 3. 配置

- `DefaultConfig()` / `SimpleConfig()`：与 README 一致。
- 各层开关、EnableCaching、CacheTTL、StopOnFirstError 等见 `validation/manager.go` 与 README 验证配置小节。

### 3.1 验证失败时的错误

- 验证失败返回 `*cd.Error`，当前主要使用 **IllegalParam**；错误信息由各约束/类型校验生成，带字段名或上下文。
- 错误码与格式的完整约定见 [error-codes.md](error-codes.md)。**待确认**：是否需在文档中单独列出「验证错误信息格式与字段路径」规范（见 [待确认项清单.md](待确认项清单.md) ERR-C3）。

### 3.2 自定义验证扩展（评审 VAL-002）

- **自定义约束**：通过 `ValidationFactory.RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error` 注册。`key` 与 struct 标签 `constraint:"key"` 或 `constraint:"key,arg1,arg2"` 对应；`validator(value, args)` 在约束层被调用。使用前需通过同一 Factory 创建 ValidationManager，以便 Manager 使用已注册的约束。
- **自定义类型处理**：通过 `ValidationFactory.RegisterTypeHandler(typeName string, handler types.TypeHandler) error` 注册，用于类型层的转换与校验。示例见 `validation/example/usage_example.go`、`validation/test/integration_test.go`。

### 3.3 国际化与性能（评审 VAL-003、VAL-004）

- **国际化**：当前错误信息为硬编码字符串，**未提供国际化（i18n）接口**；若需多语言，需在应用层对 `*cd.Error` 的 Message 做映射或包装。**需澄清**：是否计划在验证层提供 i18n 支持，见 [需澄清信息.md](需澄清信息.md)。
- **性能**：配置支持 EnableCaching、CacheTTL 等；验证在 Insert/Update/Delete 路径上同步执行，对延迟有直接影响。**需澄清**：是否在文档中约定性能目标或给出基准参考，见 [需澄清信息.md](需澄清信息.md)。

---

## 4. 当前实现状态与演进方向

为避免误解，当前验证系统在「框架能力」与「模型集成程度」上存在如下边界：

- **已实现并稳定的能力**：
  - 四层架构与 `ValidationManager` 接口本身（类型层、约束层、数据库层、场景层）；
  - 基于 `ValidationConfig` 的开关与缓存配置，以及 `DefaultConfig` / `SimpleConfig` 等预设；
  - 自定义约束（`RegisterCustomConstraint`）与自定义类型处理（`RegisterTypeHandler`）扩展点；
  - 错误收集器与统计信息（`ValidationStats`）。
- **与 Model 集成的当前状态**：
  - Orm 在 Insert/Update/Delete 路径上会调用内部 `validateModel(model, scenario)`，并通过 `validation.NewContext` 传入场景与操作类型；
  - 目前 `ValidateModel` 使用的是**简化版的 ModelAdapter/FieldAdapter**：尚未完全从 `models.Model` 中自动抽取所有字段的类型信息与 `constraint` 标签列表，字段级验证覆盖度以实际实现为准。
- **规划中的演进方向**：
  - 基于 Model 元数据自动构建完整的 FieldAdapter 列表（含类型、约束、视图等信息），使「struct 标签 → 验证管线」成为默认路径；
  - 明确验证错误的结构化格式（错误码 + 字段路径 + 约束 key），与 [error-codes.md](error-codes.md) 中的约定保持一致；
  - 按需补充数据库层与场景层的更多策略（如更细粒度的 Query 验证、只读/只写字段的统一策略等）。

在这些演进完成前，使用方应将当前验证系统理解为「已具备完整骨架与扩展点，但对模型标签的自动集成仍在逐步增强」。
