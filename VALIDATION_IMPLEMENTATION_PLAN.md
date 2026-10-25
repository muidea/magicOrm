# 校验收口实施记录

更新：2026-10-25。本文件只记录当前实施落点；使用合同以[校验设计](~/codespace/magicOrm/docs/design-validation.md)和[校验使用](~/codespace/magicOrm/validation/README.md)为准，跨仓验收状态以[ORM 收口方案](~/dataspace/magicPlatform/docs/technical/orm-capability-closeout-plan.md)为准。

| 工作项 | 当前实施 | 核对入口 |
| --- | --- | --- |
| 分层职责 | 类型、值约束、数据库兼容、场景校验；不执行数据库 UNIQUE/FK 查询 | [架构](~/codespace/magicOrm/VALIDATION_ARCHITECTURE.md) |
| 有效上下文 | 三个公开入口前置解析，Manager 默认→场景→WithOptions 完整显式快照 | [Manager](~/codespace/magicOrm/validation/manager.go) |
| 字段赋值与适配 | ModelAdapter 保留原字段元数据，Update 遗漏字段与自增主键遵循相同合同 | [适配器](~/codespace/magicOrm/validation/adapters.go) |
| 缓存 | Manager/Factory 配置传入约束结果缓存，滑动 TTL、容量上限、访问时过期，无自动后台任务 | [缓存](~/codespace/magicOrm/validation/cache/validation_cache.go) |
| 生命周期 | 自定义规则注册清缓存；Reset/池化归还恢复构造配置，清空使用期状态 | [回归](~/codespace/magicOrm/validation/effective_config_test.go) |
| 使用迁移 | 场景常量导入、WithOptions、默认场景、缓存参数与 Reset 影响 | [迁移指南](~/codespace/magicOrm/validation/MIGRATION_GUIDE.md) |

默认 ORM 池化管理器关闭缓存；独立 Manager 可以按需开启。Manager 不承诺并发配置安全，不承担业务授权、引用存在性或数据库唯一性责任。实现测试通过不代表正式接口、性能容量或 Live 验收。
