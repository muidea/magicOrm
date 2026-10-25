# 校验运行配置

以[可编译最小样例](~/codespace/magicOrm/examples/quickstart/main.go)和[配置合同](~/codespace/magicOrm/validation/README.md)为准。先从 DefaultConfig 构造，按实际负载调整 EnableCaching、CacheTTL、MaxCacheSize 及 DefaultOptions；不存在 CacheMaxSize、ValidationTimeout 或 MaxConcurrentValidations 配置字段。

Manager 不是并发任务调度器；owner 决定请求并发与取消。自定义 validator 若依赖可变外部事实，应关闭结果缓存或自行提供相应版本隔离；缓存不代替数据库约束、授权或引用存在性检查。MaxCacheSize 是条目数，不是总内存预算。

ORM 默认池化管理器关闭缓存；独立 Manager/Factory 可启用约束缓存。注册自定义规则后旧结果失效；Manager Reset/池化归还会丢弃使用期注册、场景、统计和层开关。

指标初始化顺序与[指标设计](~/codespace/magicOrm/docs/design-metrics.md)一致；缓存启用并不证明应用吞吐或 Live 效果。测试入口见[测试指南](~/codespace/magicOrm/docs/testing-guide.md)。
