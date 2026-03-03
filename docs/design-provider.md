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
| **Remote** | `provider/remote` | 用于通过远程服务访问模型的场景。 |

测试中通过 `localProvider` / `remoteProvider` 区分（如 `test/simple_test.go`、`test/constraint_local_test.go` 等）。Orm 创建时需传入 Provider，用于获取 Model、Filter 等，见 [design-orm.md](design-orm.md) 与 [design-models.md](design-models.md)。
