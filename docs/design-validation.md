# 验证系统设计

**功能块**：`~/codespace/magicOrm/validation`
**依据**：`~/codespace/magicOrm/validation/manager.go`、`~/codespace/magicOrm/validation/scenario.go`、`~/codespace/magicOrm/validation/constraints`、`~/codespace/magicOrm/validation/types`、`~/codespace/magicOrm/validation/database`、`~/codespace/magicOrm/validation/errors`。

[← 返回设计文档索引](~/codespace/magicOrm/docs/README.md)

---

## 1. 四层架构

与 README、VALIDATION_ARCHITECTURE.md 一致：

| 层 | 包 | 职责 |
|----|-----|------|
| 类型层 | types | Go 类型与数据库类型兼容性、基础转换。 |
| 约束层 | constraints | 结构体标签中的业务规则（req, min, max, range, in, re, ro, wo）。 |
| 数据库层 | database | 内存中的非空/库类型兼容检查；不访问数据库，不执行 UNIQUE/FK 查询。 |
| 场景层 | scenario | 按 Insert/Update/Query/Delete 编排策略。 |

---

## 2. 场景与 Orm 操作对应

| 场景 | Orm 操作 | 是否调用 validateModel |
|------|----------|-------------------------|
| ScenarioInsert | Insert | 是 |
| ScenarioUpdate | Update | 是 |
| ScenarioDelete | Delete | 是 |
| ScenarioQuery | Query / BatchQuery | **否** |

入口：`ValidationManager.ValidateModel(model, context)`。Orm 通过内部 `validateModel(model, scenario)` 传入 `errors.Scenario*` 与 `validation.OperationType`（Create/Read/Update/Delete）。约束定义见 [design-models.md](~/codespace/magicOrm/docs/design-models.md)。

### 2.1 四层之间的交互与数据流

- 验证入口为 `ValidateModel(model, context)`，context 中含场景（Insert/Update/Delete 等）与数据库类型。
- 执行顺序由场景适配器编排：先**类型层**（类型兼容与转换），再**约束层**（req/min/max/range/in/re/ro/wo 等），再**数据库层**（NOT NULL、库类型兼容等）。任一层失败即返回错误，是否继续后续层可由配置（如 StopOnFirstError）决定。
- 各层使用 Model 的字段与约束元数据，不直接访问数据库连接。

---

## 3. 配置

- `DefaultConfig()` / `SimpleConfig()`：与 README 一致。
- 各层开关、EnableCaching、CacheTTL、StopOnFirstError 等见 `~/codespace/magicOrm/validation/manager.go` 与 README 验证配置小节。

### 3.1 验证失败时的错误

- 验证失败返回 `*cd.Error`，当前主要使用 **IllegalParam**；错误信息由各约束/类型校验生成，带字段名或上下文。
- 若错误来自数据库层验证，则会映射为 **DatabaseError**；多字段错误会聚合为 `Multiple validation errors: field (count), ...`。
- 错误码与格式的完整约定见 [error-codes.md](~/codespace/magicOrm/docs/error-codes.md)。

### 3.2 自定义验证扩展（评审 VAL-002）

- **自定义约束**：通过 `ValidationFactory.RegisterCustomConstraint(key models.Key, validator models.ValidatorFunc) error` 注册。`key` 与 struct 标签 `constraint:"key"` 或 `constraint:"key,arg1,arg2"` 对应；`validator(value, args)` 在约束层被调用。使用前需通过同一 Factory 创建 ValidationManager，以便 Manager 使用已注册的约束。
- **自定义类型处理**：通过 `ValidationFactory.RegisterTypeHandler(typeName string, handler types.TypeHandler) error` 注册，用于类型层的转换与校验。示例见 `~/codespace/magicOrm/validation/example/usage_example.go`、`~/codespace/magicOrm/validation/test/integration_test.go`。

### 3.3 国际化与性能（评审 VAL-003、VAL-004）

- **国际化**：当前错误信息为硬编码字符串，**未提供国际化（i18n）接口**；若需多语言，需在应用层对 `*cd.Error` 的 Message 做映射或包装。
- **性能**：配置支持 EnableCaching、CacheTTL 等；验证在 Insert/Update/Delete 路径上同步执行，对延迟有直接影响。

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
  - `ValidateModel` 会通过 `AdaptModel/AdaptField` 基于真实 `models.Model` / `models.Field` 抽取字段类型、约束与当前值；
  - 适配时会优先保留字段运行时承载类型，避免将 `datetime`、struct relation、slice relation 误降级成 `time.Time` / `interface{}` / `[]interface{}`。
- **规划中的演进方向**：
  - 明确验证错误的结构化格式（错误码 + 字段路径 + 约束 key），与 [error-codes.md](~/codespace/magicOrm/docs/error-codes.md) 中的约定保持一致；
  - 按需补充数据库层与场景层的更多策略（如更细粒度的 Query 验证、只读/只写字段的统一策略等）。

在这些演进完成前，使用方应将当前验证系统理解为「已具备真实模型适配能力与扩展点，但更细粒度的错误结构和策略编排仍可继续增强」。

### 可空文本值约束（2026-10-22）

内置 `in` 与 `re` 校验先解引用指针/接口，对非空值检查实际内容；未赋值、nil 接口与 nil 指针不参与成员/格式比较。`req`、类型校验和数据库非空约束保持独立：可空字段的 null 不等于合法的必填值，也不会把错误的非空值放行。此通用修复不识别任何平台/应用字段。local 与 remote provider 回归分别覆盖 null、合法指针、格式错误、枚举错误和 required-null；应用业务仍经 magicBase/magicOrm 执行。

全量 PostgreSQL 回归同时发现旧 TestRemoteOnlineEntity 夹具以空目标对象（包含显式主键0）过滤关联，并假定新建ID=1。使用修复前校验器可独立复现；用例现读取 owner 实际分配键、按真实关联键查询并核对返回目标内容，不改变生产查询语义。默认本机连接因认证失败的首轮及隔离库首轮失败日志均保留；最终隔离库全仓 test/vet/build 通过。

本批源码终验与夹具边界见[可靠写入与设备规则证据](~/codespace/docs/acceptance/runs/2026-10-22-reliable-entity-value-writes-source1/source-checks.json)；完整页面/五角色/初始化和正式候选仍待验证，liveReady=false。

## 有效配置与生命周期（2026-10-25）

三个公开入口在所有捷径前统一解析上下文，顺序为 Manager 配置、场景默认、调用显式完整选项。空场景使用 SetScenario 默认，否则 Insert；显式未知场景报错。使用 WithOptions 标记覆盖，包括 false；直接赋值 Options 不表示显式覆盖。原始模型与 AdaptModel 使用相同的更新赋值跳过和自动主键处理。

Manager/Factory 将 CacheTTL 与 MaxCacheSize 传入约束缓存；0 为库默认，负数禁用并报告，EnableCaching=false 优先。容量指约束结果条目，TTL 为滑动期限；读取惰性过期、写入容量淘汰，不启动每个缓存的清理协程，未启用整体模型结果缓存。注册自定义规则清缓存；池化 Manager 归还时清空所有使用期状态并恢复构造配置。Manager 不保证并发配置与调用安全。详见[校验使用](~/codespace/magicOrm/validation/README.md)。
