# Provider 层设计

**功能块**：`provider/`  
**依据**：`provider/provider.go` 中 `Provider` 接口。

[← 返回设计文档索引](README.md)

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
| **Local** | `provider/local` | 基于反射与本地类型信息，用于单进程、直连数据库。 |
| **Remote** | `provider/remote` | 用于通过远程服务访问模型的场景。详见 [design-remote-provider.md](design-remote-provider.md)。 |

测试中通过 `localProvider` / `remoteProvider` 区分。Orm 创建时需传入 Provider，用于获取 Model、Filter 等，见 [design-orm.md](design-orm.md) 与 [design-models.md](design-models.md)。

### 2.1 Provider 选择指南

- **Local**：单进程、应用直连数据库；所有 Model/Filter 基于本地反射与类型注册，无网络开销，适合常规 CRUD 与事务。
  - 当前本地 `ValueImpl` 也支持显式赋值状态：可以区分“未赋值”“显式赋值为零值”“显式赋值为 typed nil”。
  - 这套能力在通过 `Model.Copy(models.MetaView)` + `SetFieldValue(...)` 构造 patch 时生效。
  - 直接从原始 Go struct 做 `GetEntityModel(...)` 时，框架仍然只能从反射结果推断赋值状态，因此原始 struct 本身仍不能可靠区分“字段未提供”和“字段就是 nil”。
- **Remote**：通过远程服务访问模型数据；当前仓库内的运行时模型以 `Object` / `ObjectValue` / `SliceObjectValue` 为核心，VMI JSON 为事实样例，详细语义见 [design-remote-provider.md](design-remote-provider.md)。Remote 的外部通讯协议仍未单独文档化。

### 2.2 错误处理

- Provider 层方法返回 `*cd.Error`，常见为 `IllegalParam`（如 entity/model 为 nil、类型不合法）。完整错误码见 [error-codes.md](error-codes.md)。
