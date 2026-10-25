# 校验调用迁移

本次配置修复不改变模型身份、必填、唯一、nullability 或 owner 授权规则。迁移前先使用[默认离线测试](~/codespace/magicOrm/docs/testing-guide.md)，实际数据库与消费者分别验证。

1. 场景常量从 `github.com/muidea/magicOrm/validation/errors` 导入，例如 verrors.ScenarioInsert；OperationCreate 等仍属于 validation。
2. 将直接修改 context.Options 的调用改为 `ctx = ctx.WithOptions(options)`，传入完整选项快照，包括显式 false。NewContext 的默认值不标记显式覆盖。
3. SetScenario 设置 Manager 默认场景；空场景逐次解析，未知场景在任何快捷返回前拒绝。复用 context 不会锁定上次解析的场景。
4. CacheTTL/MaxCacheSize 现在生效，0 使用默认，负数禁用并报告。容量为约束结果条目数，TTL 为滑动期限。不要依赖已移除的自动清理协程或未被 Manager 使用的整体模型结果缓存。
5. 池化归还和显式 Reset 丢弃使用期自定义注册，包括 Factory 注入规则。需要规则时在下一次独立使用时重新注册；不要跨使用者共享可变 Manager。

完整编译样例见[quickstart](~/codespace/magicOrm/examples/quickstart/main.go)，详细架构见[验证设计](~/codespace/magicOrm/docs/design-validation.md)。
