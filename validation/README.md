# 校验使用

公开入口为 `Validate`、`ValidateField`、`ValidateModel`；场景常量属于 `~/codespace/magicOrm/validation/errors`，操作常量属于 `validation`。完整编译样例见[quickstart](~/codespace/magicOrm/examples/quickstart/main.go)，详细合同见[验证系统设计](~/codespace/magicOrm/docs/design-validation.md)。

```go
manager := validation.NewValidationManager(validation.DefaultConfig())
ctx := validation.NewContext(verrors.ScenarioInsert, validation.OperationCreate, nil, "postgresql")
err := manager.ValidateModel(model, ctx)
```

其中 `verrors` 为 `github.com/muidea/magicOrm/validation/errors` 的导入别名。`model` 为已注册的 `models.Model`。

Manager 默认选项、场景默认、调用显式选项依次覆盖。使用 `ctx = ctx.WithOptions(options)` 提交完整调用选项，包含 false；直接修改 Options 字段不标记覆盖。`SetScenario` 设置 Manager 默认场景；NewContext 的空场景在每个公开入口解析，不修改可复用调用上下文；未知场景在所有捷径前报错。

EnableCaching=false 优先禁用；CacheTTL/MaxCacheSize 为 0 使用默认，负数禁用并报告配置问题，正数实际生效。MaxCacheSize 是约束结果条目上限；TTL 为滑动有效期。过期在读取时淘汰，写入保持容量上限，不启动每个 Manager 的后台清理协程。Manager 不使用整体模型结果缓存。

自定义规则注册会清空已有结果。Manager 本身不承诺并发配置/校验安全；池化 Manager 归还时恢复构造配置，清空缓存、自定义注册、层开关、默认场景和统计。显式 Reset 同样丢弃使用期注册，包括 Factory 注入的规则。原始 Go struct 无法可靠区分遗漏与显式零值，patch 使用 MetaView + SetFieldValue。

校验层没有数据库连接，不发 UNIQUE/FK 或目标存在性查询；数据库约束和 owner 授权独立执行。
