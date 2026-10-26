# Provider 层设计

**功能块**：`~/codespace/magicOrm/provider`
**依据**：`~/codespace/magicOrm/provider/provider.go` 中 `Provider` 接口。

[← 返回设计文档索引](~/codespace/magicOrm/docs/README.md)

---

## 1. 接口清单

| 方法 | 说明 |
|------|------|
| RegisterModel(entity) | 注册实体类型，返回 Model |
| UnregisterModel(entity) | 注销实体类型 |
| GetEntityType(entity) | 获取实体 Type |
| GetEntityValue(entity) | 获取实体 Value |
| GetEntityModel(entity, disableValidator) | 获取实体 Model（可选禁用校验） |
| GetEntityFilter(entity, viewSpec) | 根据实体与视图声明获取 Filter |
| GetTypeModel(vType) | 根据 Type 获取 Model |
| GetModelFilter(vModel) | 根据 Model 获取 Filter |
| GetTypeFilter(vType, viewSpec) | 根据 Type 与视图声明获取 Filter |
| SetModelValue(vModel, vVal) | 用 Value 设置 Model 字段值 |
| EncodeValue(vVal, vType) | 值编码 |
| DecodeValue(vVal, vType) | 值解码 |
| Owner() | 返回 Provider 所属 owner |
| Reset() | 重置内部状态 |

---

## 2. 实现与使用场景

| 实现 | 包路径 | 说明 |
|------|--------|------|
| **Local** | `~/codespace/magicOrm/provider/local` | 基于反射与本地类型信息，用于单进程、直连数据库。 |
| **Remote** | `~/codespace/magicOrm/provider/remote` | 动态 Object/ObjectValue 模型载体；不内置 RPC/HTTP 客户端。详见 [design-remote-provider.md](~/codespace/magicOrm/docs/design-remote-provider.md)。 |

测试中通过 `localProvider` / `remoteProvider` 区分。Orm 创建时需传入 Provider，用于获取 Model、Filter 等，见 [design-orm.md](~/codespace/magicOrm/docs/design-orm.md) 与 [design-models.md](~/codespace/magicOrm/docs/design-models.md)。

### 应用入口与兼容边界

普通调用通过 `provider.NewLocalProviderWithOptions`／`provider.NewRemoteProviderWithOptions` 取得 `Provider`，不需要构造底层 `TypeImpl`／`ValueImpl`。旧构造器和七项函数注入选项继续支持，不能仅因工作区当前没有消费者就删除。

目录按声明和实现组织，保持 Go 包身份：

- [Provider 合同](~/codespace/magicOrm/provider/provider.go)保留接口和旧构造器；[私有实现](~/codespace/magicOrm/provider/provider_impl.go)维护模型缓存和读写组合，配置仍在 `~/codespace/magicOrm/provider/options.go`。
- [Local 入口](~/codespace/magicOrm/provider/local/api.go)集中构造和模型／值操作；反射类型、字段、模型、过滤及编解码细节由同包相应实现文件维护。
- [Remote 声明](~/codespace/magicOrm/provider/remote/declarations.go)集中 Object、Field、Type／Spec、值与过滤 DTO；现有文件维护其运行方法，声明仍含原有不序列化的私有状态。
- [校验合同](~/codespace/magicOrm/validation/contracts.go)与配置从 Manager／Factory 实现中分离；缓存配置和观察数据集中于 `~/codespace/magicOrm/internal/validation/cache/contracts.go`。各内置校验包保留合同文件和私有实现。

Local 与 Remote 的公开具体类型保留原包定义，以保持具体返回类型、反射包路径及 JSON；二者是同级适配层，不能因工作区直接使用量不同而隐藏 Local。支撑包按下表迁入 internal；已有应用使用 Local／Remote 的入口保持，使用旧支撑包导入的调用方须迁移。新增或删除公开项须通过[接口快照审查](~/codespace/magicOrm/internal/apicheck/public_api_test.go)，具体类型归属由[外部调用兼容回归](~/codespace/magicOrm/provider/public_contract_test.go)保护。

### 2.1 Provider 选择指南

- **Local**：单进程、应用直连数据库；所有 Model/Filter 基于本地反射与类型注册，无网络开销，适合常规 CRUD 与事务。
  - 当前本地 `ValueImpl` 也支持显式赋值状态：可以区分“未赋值”“显式赋值为零值”“显式赋值为 typed nil”。
  - 这套能力在通过 `Model.Copy(models.MetaView)` + `SetFieldValue(...)` 构造 patch 时生效。
  - 直接从原始 Go struct 做 `GetEntityModel(...)` 时，框架仍然只能从反射结果推断赋值状态，因此原始 struct 本身仍不能可靠区分“字段未提供”和“字段就是 nil”。
- **Remote**：由动态声明管理模型与值，网络传输由上层完成；当前仓库内的运行时模型以 `Object` / `ObjectValue` / `SliceObjectValue` 为核心，VMI JSON 为事实样例，详细语义见 [design-remote-provider.md](~/codespace/magicOrm/docs/design-remote-provider.md)。Remote Provider 本身不定义外部通讯协议。

### 2.2 错误处理

- Provider 层方法返回 `*cd.Error`，常见为 `IllegalParam`（如 entity/model 为 nil、类型不合法）。完整错误码见 [error-codes.md](~/codespace/magicOrm/docs/error-codes.md)。

### 2.3 本地字段排除

`orm:"-"` 在反射类型检查和字段映射之前生效，规则对任何应用均相同。排除字段不进入 Model、视图、DDL 图或 CRUD；其关联类型不需要注册。`json:"-"` 独立控制 JSON，不能代替 ORM 排除。完整语义见 [标签参考](~/codespace/magicOrm/docs/tags-reference.md#15-忽略字段)。

回归必须检查实际 Model 中不存在名为 `-` 的字段，并通过未注册被排除关联类型的真实 Create 图验证。仅检查字段的原 Go 名称或改变模型注册顺序，不能证明已排除持久化关系。


### 支撑包迁移与替代入口

| 原公开包／函数 | 当前归属或应用入口 |
| --- | --- |
| `provider/local`、`provider/remote` | 同级公开适配层，两者均保留，不提供网络传输 |
| `validation/cache` | `internal/validation/cache`；应用通过 `validation.ValidationConfig` 配置缓存 |
| `validation/constraints`、`validation/database` | `internal/validation` 对应实现；应用通过 ValidationFactory 的 CreateConstraintValidator／CreateDatabaseValidator 使用接口 |
| `validation/types` | 内置转换实现位于 `internal/validation/types`；应用通过 CreateTypeValidator 使用，自定义处理器实现 `validation.TypeHandler`／`models.TypeHandler` |
| 三个 `metrics/*` 子包 | 采集和注册实现位于 `internal/metrics`；ORM自动初始化仍由 `orm.Initialize()` 负责，独立采集调用 `metrics.Initialize()`，晚启动统一调用 `metrics.EnsureRegistered()` |
| 原 `metrics` 键、标签和采样辅助函数 | `internal/metrics/shared`；应用通过 magicCommon/monitoring 读取指标 |
| `utils` 反射、深拷贝等辅助函数 | `internal/utils`；应用保留公开的29项模型值／约束／时间／ID操作和分页、排序 DTO |

不为已收回的七个支撑包提供原路径转发。公开签名只能引用公开合同，检查同时记录导入路径，不能用别名把 internal 类型重新暴露。ORM／Provider 的现有功能与Local／Remote声明不变；本次是Go源码导入边界调整，不涉及数据迁移或业务Schema更新。
