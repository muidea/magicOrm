# magicOrm 开发规则

本文件直接规定本仓的职责、边界和开发要求；单独检出本仓即可阅读执行。仓库内位置均相对于仓库根目录，协作项目通过项目名和 Git 地址识别，不依赖固定工作区或外部规范文件才能理解规则。

## 功能定位与边界

- 提供通用模型、Provider、校验、视图、CRUD／查询、关联、条件写／事务、约束／索引、聚合及模型表结构操作；不持有平台业务策略或工作流。
- 稳定接口在 models，orm.Orm 是 models.ORM 别名；应用优先使用 orm／models／provider。Local 映射 Go 类型，Remote 映射动态模型声明与值，二者是平级公开适配层；Remote 不提供网络传输，Local 不迁入 internal。
- 保留 Local／Remote 的正式构造器、类型身份、模型／值／过滤／codec 和 JSON 合同。数据库执行器／连接池、验证缓存和采集器实现放在 internal，公共签名不泄露内部类型，不因方便开放任意 SQL 或测试 setter。
- Panel／部署开通数据库、物理 Schema 容器、账号和权限；Base 经本仓管理业务模型表和数据。ORM 不推断租户归属、不补授权、不决定部署、代际或未完成 DDL 的业务恢复。
- 保持 ID 分配、必填／默认／null／唯一、关联和事务语义。不得为平台错误放宽约束、把 update 改成 insert 或识别特定应用／租户名。
- 普通嵌套事务共享物理事务，内层回滚不能伪装为可提交；Serializable 不自动降级或重试。提交未知由 owner 回查原操作，ORM 不猜测成功。业务 datetime 按 UTC 合同编码，0、null、空集合和精确大数不能混同。

- 完整模型集合的 Schema 生命周期是可选合同；共享 owned 去重和物理漂移检查由 ORM 执行，完整根集合及图外保护由调用 owner 证明，不把平台退役计划下沉。

## 协作项目

- magicCommon — https://github.com/muidea/magicCommon.git：通用基础工具。
- magicBase — https://github.com/muidea/magicBase.git：模型与数据执行 owner，维护 Schema 日志和平台恢复围栏。
- magicCas — https://github.com/muidea/magicCas.git：身份授权 owner，不由 ORM 授权。
- magicFile — https://github.com/muidea/magicFile.git、magicRunner — https://github.com/muidea/magicRunner.git：消费受支持的公开合同，不导入数据库内部实现。

## 结构、编码与验证

- Go 工具链以 go.mod 为准；使用 gofmt／goimports，沿用 `*cd.Error`。orm 门面、models 接口及 Provider 的合同与实现可同包分文件，不增加无用途的兼容转发。
- 改公开 API 后运行 `go test ./internal/apicheck ./provider/...`；先评估兼容影响，再用 `go test ./internal/apicheck -update-api` 更新快照，不能用改快照掩盖意外暴露。
- 先运行受影响包，再运行 `go test -count=1 ./...`、`go test -tags=mysql -count=1 ./...`、`go vet ./...`、`go build ./...`。离线跳过数据库不能记为数据库通过；Provider 变化覆盖 Local／Remote 和受影响 builder。
- 真实数据库只使用显式的一次性 magicorm_test_ 目标；PostgreSQL 夹具入口为 `scripts/integration_postgres.sh`，禁止共享数据库和 replication／GRANT 测试。平台当前 PostgreSQL 验收与 ORM 的 MySQL 合同验证分开。
- 文档／源码移动后运行 `python3 scripts/check_docs.py`；指标保持初始化复用，支持 monitoring 先初始化及 metrics.EnsureRegistered 的显式补注册，不导出测试开关。

## 依赖与交付

- 先修改语义 owner，再更新公开合同及直接消费者。工作区内使用 `upgrade_vendor.sh <目标项目...>` 同步依赖；重建前检查工作树和暂存区，保护 vendor 独有修改、原字节及补丁，保护失败即停止。禁止手改 vendor；同步后审查差异，并以默认 vendor 模式验证消费者。本地 replace 模式不能替代此验证；独立检出缺少同步环境时明确记录未验范围。
- 先运行最小受影响测试，再验证直接消费者。仅文档修改不要求格式化未改动的源码或执行部署；代码变化使用 gofmt／goimports、相关测试、go vet 和构建，按风险选择 race／真实数据库验证。源码、vendor、发行包、镜像和 Live 结果分别记录，夹具成功不能替代持久 owner 回读或现场验收。
- 合同或行为变化同时更新本仓使用说明、设计和对应平台材料；本文件的定位和边界直接同步更新，不把必要规则改成外部文档链接。提交按仓库分开使用 Conventional Commits，不提交凭据、密钥、运行数据或构建产物。
