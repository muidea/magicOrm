# MagicORM 校验架构

ValidationManager 协调类型、约束、数据库兼容与场景四层。公开接口、签名与有效配置以[验证设计](~/codespace/magicOrm/docs/design-validation.md)及[源码](~/codespace/magicOrm/validation/manager.go)为准。

类型层检查运行时类型，约束层执行 req/min/max/range/in/re/ro/wo，数据库兼容层检查内存值与库类型规则，场景层适配 Insert/Update/Query/Delete。各层不持数据库连接，不查询 UNIQUE/FK；数据库物理约束与业务 owner 的身份、授权、引用存在性检查独立执行。

ORM 的 Insert/Update/Delete 在写入前调用校验；Query/BatchQuery 不自动调用。三个公开入口先解析有效上下文，再处理整体模型遍历、未赋值字段、自动主键等捷径。AdaptModel 保留真实字段的赋值与主键元数据，手工适配器没有这些元数据时按完整提供值处理。

Manager/Factory 共享构造逻辑与缓存配置；缓存只服务约束结果，采用滑动 TTL、容量淘汰与惰性过期，不启动每个缓存的后台协程。自定义规则注册清缓存，池化归还恢复构造配置并丢弃使用期状态。配置/调用本身不承诺并发安全。完整示例见[quickstart](~/codespace/magicOrm/examples/quickstart/main.go)。
