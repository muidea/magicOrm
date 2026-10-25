# 测试指南

默认 `go test ./... -count=1` 与 `~/codespace/magicOrm/unit_test.sh` 不连接数据库，所有数据库测试保留可见 skip。MySQL 编译路径使用 `go test -tags=mysql ./... -count=1`。`go vet ./...` 和 `go build ./...` 亦纳入 CI。完整 README 样例与可编译源文件一致性由测试校验。

## PostgreSQL 隔离全量

```bash
cd ~/codespace/magicOrm
./scripts/integration_postgres.sh
```

脚本仅新建本次专属容器、随机口令与环回随机端口，数据库固定为本次专属实例中的 magicorm_test_closeout，非 public 的 integration Schema。退出/失败/取消按精确 ID 删除容器及其匿名卷并回查，不扫描前缀、不访问平台数据库、不保留凭据。Docker/KILL 或宿主中断时由调用者根据已记录的精确 ID 核对，不能把清理未执行当作成功。可传入命令，在同一隔离环境验证消费者；超时由各用例控制。

手动环境须显式设置 MAGICORM_INTEGRATION_ALLOW=1 与 MAGICORM_POSTGRES_SERVER、DATABASE、SCHEMA、USER、PASSWORD（各项带 MAGICORM_POSTGRES_ 前缀）。Database 必须以 magicorm_test_ 开始，Schema 不可为 public，所有凭据显式提供。启用但配置/连接错误会失败，不能降级 skip。整个旧集成包操作可销毁数据库，勿将现有平台库配置给测试。

Schema 物理专项另读 MAGICORM_SCHEMA_TEST_POSTGRES_DSN，必须为显式账号/口令及 magicorm_test_ 数据库的 PostgreSQL URL，并同时设置 MAGICORM_INTEGRATION_ALLOW=1。专项用例创建随机 Schema 并清理本次身份。datetime 与 inline-reference 专项使用其各自 MAGICORM_*_TEST_ALLOW=isolated 及 SERVER/DATABASE/SCHEMA/USER/PASSWORD；也要求 magicorm_test_ 数据库和非 public Schema，脚本统一提供。

## 可选 MySQL

手动提供专属 magicorm_test_ 数据库，以 MAGICORM_INTEGRATION_ALLOW=1 及 MAGICORM_MYSQL_SERVER/DATABASE/SCHEMA/USER/PASSWORD 配置后执行 `go test -tags=mysql ./... -count=1`。Schema 专项使用 MAGICORM_SCHEMA_TEST_MYSQL_DSN，物理目录核验要求 MySQL 8.0.13+。不授予复制、GRANT 等管理权限；不把 MySQL 离线通过记为真实数据库通过。

## 回归分类

- 双库四种 executor 的普通事务生命周期、嵌套与错误恢复使用内存 driver，核对物理事务调用次数。
- PostgreSQL 实际提交响应丢失由协议代理转发 COMMIT 并确认服务端完成后断开回包；ORM 必须报告 Unexpected，另一分支在 COMMIT 转发前断连，独立连接按原 ID 分别确认一条和零条记录，不自动重试。它证明 ORM 原语，不能代替业务 owner 的完整恢复验收。
- owned 图、内联组合唯一、跨实例物理名隔离、目录检查、UTC 使用隔离 PostgreSQL；并发写唯一成功/冲突分别核对。
- 旧 TestNewPool/TestMasterSlaver 固定数据库压力与复制授权代码已移除，其事务风险由有断言的内存生命周期测试替代；禁止用按名称排除失败测试充当默认安全入口。

源码、隔离库、消费者默认 vendor 与 Live 分别记录。本轮方案和状态见[收口方案](~/dataspace/magicPlatform/docs/technical/orm-capability-closeout-plan.md)。

## 可编译使用样例与专项入口

初次接入先读[完整 quickstart](~/codespace/magicOrm/examples/quickstart/main.go)，它包含 imports、显式配置、Schema 准备、数据库绑定、模型注册、校验、写入和释放。其余主题直接引用同仓编译的完整用例，避免在文档重复手抄 API；内存 executor 用例只说明接口与失败路径，物理结果以隔离测试为准。

| 主题 | 完整源码 | 执行层级 |
| --- | --- | --- |
| Local 注册与 CRUD | [Local 用例](~/codespace/magicOrm/test/simple_local_test.go)及[环境门禁](~/codespace/magicOrm/test/integration_test.go) | 隔离 PostgreSQL |
| Remote 声明、值与 CRUD | [Remote 用例](~/codespace/magicOrm/test/simple_remote_test.go)及[结构定义](~/codespace/magicOrm/test/define_local_test.go) | 隔离 PostgreSQL |
| 显式 patch、nil 清空与 owned 差异 | [关系更新用例](~/codespace/magicOrm/test/update_relation_diff_local_test.go) | 隔离 PostgreSQL |
| 普通事务提交/回滚 | [事务用例](~/codespace/magicOrm/test/transaction_local_test.go) | 隔离 PostgreSQL |
| Serializable 可选能力 | [能力与错误传播](~/codespace/magicOrm/orm/serializable_test.go) | 离线；不等于实库并发冲突验收 |
| 单值引用组合唯一 | [双 Provider 定义](~/codespace/magicOrm/orm/inline_reference_test.go)与[并发持久验证](~/codespace/magicOrm/orm/inline_reference_database_test.go) | 离线＋隔离 PostgreSQL |
| 条件写与精确行锁 | [条件写和引用筛选](~/codespace/magicOrm/orm/inline_reference_test.go)与[事务行锁](~/codespace/magicOrm/orm/conditional_lock_test.go) | 离线；Base 行锁专项另见实施证据 |
| Schema 预检、检查与恢复 | [物理 DDL 用例](~/codespace/magicOrm/orm/schema_ddl_postgres_test.go)与[恢复失败路径](~/codespace/magicOrm/orm/schema_repair_test.go) | 隔离 PostgreSQL＋离线 |
| 独立 ValidationManager | [完整调用和配置回归](~/codespace/magicOrm/validation/effective_config_test.go) | 离线 |

在同一专属夹具中只执行某个数据库样例，可用：

```bash
cd ~/codespace/magicOrm
./scripts/integration_postgres.sh go test ./test -run '^TestSimpleRemote$' -count=1 -v
```

测试包包含共享模型与辅助函数，应通过上述包级命令运行，不单独 go run 某个测试文件。测试负责可销毁夹具中的结构与数据，普通生产接入不可照搬 Drop 初始化步骤。
