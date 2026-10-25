# 部署与监控接入

magicOrm 是嵌入 Go 服务的库，随业务服务构建和发布；不提供独立 ORM 服务、HTTP 端口、健康检查路由或内置 Prometheus/Datadog 推送器。服务部署清单和监控导出由宿主应用维护。

## 数据库与资源生命周期

完整入口见[quickstart 源码](~/codespace/magicOrm/examples/quickstart/main.go)及[项目 README](~/codespace/magicOrm/README.md)。配置顺序为显式 `EnsureSchema(server, dbName, schemaName, user, password)`（仅部署 owner 需要创建 Schema 时调用）、`Initialize()`、七参数 `AddDatabase(server, dbName, schemaName, user, password, maxConnNum, owner)`，再通过 Provider 与 `GetOrm(ctx, provider, prefix)` 使用业务表。释放顺序为 handler.Release、成功绑定对应的 DelDatabase、进程停止时 Uninitialized。

AddDatabase 不创建 Schema；database 与 schema 独立，不能传组合名称。凭据从私密运行环境获取，示例的 ORM_* 环境变量属于示例入口，不是库自动读取的全局配置。默认构建 PostgreSQL，MySQL 用 `-tags=mysql`；物理 Schema 检查的 MySQL 路径要求 8.0.13+。同一 handler 串行使用，context 决定取消和超时；连接预算由 owner 明确配置。

普通事务每个成功 Begin 恰好结束一次，嵌套无 savepoint；提交错误可能意味着结果未知，业务 owner 应在新连接按原身份回查，禁止自动重发。详细事务与 Schema 合同见[ORM 设计](~/codespace/magicOrm/docs/design-orm.md)。

## Schema 交付

Create/Drop/Inspect/Repair 管理完整 owned 图，保留独立引用目标主表。业务 owner 提供命名空间独占证明、图外使用者保护、持久意图和管理写入串行；不能把 Inspect 当作锁或把 IF NOT EXISTS 成功当作物理一致。

PostgreSQL 约束和索引物理名包含表/声明身份的稳定摘要；历史未限定名称不会自动重命名。表名超过 63 字节在执行前拒绝。此候选采用干净 Schema；保留历史数据时必须另行设计并验证显式迁移。新必填字段、存储类型切换及破坏性变更不能用 Reconcile 自动完成。

平台 Base 还需 foundation 初始化独立 SchemaManagementFence 协调表；所有管理者使用同一权威控制数据库与目录前缀并同步升级，停止旧管理写者。控制连接池至少两个连接。旧删除计划缺少准入证明时保留 pending。平台部署限制及真实验收门槛见[收口方案](~/dataspace/magicPlatform/docs/technical/orm-capability-closeout-plan.md)。

## 指标接入

建议宿主先初始化 magicCommon 的 monitoring 全局管理器，再调用 orm.Initialize。ORM 自动初始化三个内存 collector，并在全局管理器存在时注册 magicorm_orm、magicorm_database、magicorm_validation provider。

若全局管理器晚于 ORM 初始化，宿主在监控初始化完成后分别调用 `orm.EnsureORMMetricProviderRegistered()`、`metricsdb.EnsureDatabaseMetricProviderRegistered()`、`metricsvalidation.EnsureValidationMetricProviderRegistered()`。导入位置与指标清单见[指标设计](~/codespace/magicOrm/docs/design-metrics.md)。宿主负责监控生命周期、采集出口和鉴权；库没有独立 magicOrm/monitoring 包。

当前耗时指标是有限样本的平均值，内存 collector 重启丢失，不提供分位数或持久队列保证。不要把业务 ID、SQL、DSN 或凭据放入指标标签。可选有界查询成本窗口由 MAGIC_PROFILE_WINDOW 控制，详细边界同指标设计。

## 发布前验证

运行[测试指南](~/codespace/magicOrm/docs/testing-guide.md)中的默认离线、mysql-tag、build/vet 和独立 PostgreSQL 入口；owner 变化后通过 `~/codespace/upgrade_vendor.sh` 同步实际消费者，并用默认 vendor 模式验证。源码、隔离数据库、消费者、正式接口与 Live 各自记录；构建成功不能代替部署验收。本文不提供未经验证的部署资源配额、通用监控开关或采样配置。
