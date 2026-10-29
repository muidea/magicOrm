# magicOrm

2026-10-29 Live6 剩余项源码候选：整集合 Schema 生命周期通过可选 `models.SchemaSetLifecycle` 提供 `InspectSchemaSet`／`DropSchemaSet`，离线表集合使用 `orm.SchemaSetTables`；基础 ORM 接口和单根安全约定保持。图声明先整体核验，共享 owned 表去重，目标删除后独立观察确认；平台所有权、冻结计划和恢复身份由调用 owner 维护。 见[实施与验证边界](~/dataspace/magicPlatform/docs/technical/history/platform-live6-final-code-closeout-2026-10-29.md)。

## 项目定位与开发边界

本仓 `AGENTS.md` 直接列出项目定位、公开能力、职责边界及开发验证要求，单独检出即可执行；跨项目配合使用项目名和 Git 地址识别。先修改事实 owner，再同步消费者和交付材料。平台架构长文作为补充设计依据，不是读取本仓规则的前置依赖。下文日期化记录保留各批范围，不代表当前部署全部验收通过。

magicOrm 是面向 Go 的模型 ORM，默认 PostgreSQL，使用 `-tags=mysql` 构建 MySQL。Local Provider 处理 Go 模型，Remote Provider 处理动态模型声明与值；Remote 不内置网络传输。同一 ORM handler 由调用方串行使用。

## 能力与边界

- CRUD、Query 单查、BatchQuery 列表、视图与掩码、显式赋值、条件更新与精确主键行锁。Query 主键优先，单查要求唯一命中；读计划不能授权写入。
- 独立引用与递归 owned 关系；组合唯一支持基础字段和单值指针引用。参与组合唯一的引用以目标标量主键内联主表，不双写链接表。写引用检查非零主键，不证明目标存在。目标删除保护、租户隔离、授权和业务防环属于 owner。
- Local/Remote 唯一声明；Remote 普通索引。Local 标准入口没有 `GetIndexes`；多值关系和 owned 字段不能参与组合唯一。声明准入不保证任意字段均可在两类数据库物理索引。
- 普通嵌套事务共享一个物理事务；内层回滚使整体只能回滚。Serializable 必须在空闲时开启，不降级、不自动重试。PostgreSQL 唯一冲突映射 Duplicated，序列化/死锁映射 VersionConflict，连接丢失及提交未知保持 Unexpected，由 owner 回查原身份。
- Create/Drop/Inspect/Repair 管理宿主、关系表及完整递归 owned 图，保留独立引用目标主表。业务 owner 证明物理命名空间独占、图外使用者保护并串行 Schema 写入；ORM 不维护平台所有权。
- Reconcile 只处理安全增量；必填新字段、删字段、存储类型切换等要求显式迁移。`EnsureSchema` 显式创建 Schema，`AddDatabase` 不创建 Schema，不接受 database/schema 合并名称；PostgreSQL Schema 不存在或当前账号无 `USAGE` 时拒绝连接注册，不修复授权。
- UTC 毫秒持久化；无时区旧文本按 UTC，不自动转换历史本地时间。精确 JSON 数值不经浮点中转。
- 受限 AND/OR/null、多字段排序、keyset、注册单值引用子查询；独立聚合支持至多 16 指标、4 分组维、1000 组，UTC 小时桶至多 744 桶、5000 点。不是任意 SQL、授权游标或跨查询快照。
- 类型、约束、数据库兼容和场景校验；校验层不查询 UNIQUE/FK，物理唯一性由数据库执行。缓存参数与显式选项见[校验合同](~/codespace/magicOrm/docs/design-validation.md)。

PostgreSQL 约束和索引的物理名由表名与声明名两段身份派生，统一附加稳定摘要并限制长度，避免同 Schema 不同实例发生索引名碰撞。现有无表限定的旧物理名须显式迁移；本批使用干净候选 Schema，不自动重命名旧索引。

## 公共接口与内部边界

业务读写继续通过 `orm.Orm`，它是 `models.ORM` 的别名，16项方法保持；可选能力包括 Serializable、精确行锁、Aggregate 及整集合 Schema 生命周期。`orm.SchemaReport` 保留 `Matches`／`Tables`／`Issues` 字段和 JSON；结果数据、连接描述接口及池快照定义在 `models`。`NewConfig` 的既有方法包含 `GetDsn`，不返回执行器或连接池。

数据库实现位于 `~/codespace/magicOrm/internal/database/`，Go 限制外部导入；Runner及 `NewExecutor`／`NewBuilder`／`NewPool` 不再是公共门面。已有直接依赖已同步迁移。需要物理宿主表身份时使用 `orm.SchemaTableName`，完整 owned 图仍用 `SchemaTables`；两者都不授予写入权。Provider 的既有构造入口、模型声明与值 DTO 保持，业务调用方不依赖数据库驱动实现。详见[接口合同](~/codespace/magicOrm/docs/design-orm.md#公共边界整理2026-10-26)。

按实际需求选择依赖，不必导入所有支撑包：

| 使用场景 | 推荐包与入口 |
| --- | --- |
| Go struct 模型读写 | `orm`、`models`、`provider.NewLocalProviderWithOptions` |
| 动态模型声明与值 | 增加 `provider/remote`，使用 `provider.NewRemoteProviderWithOptions` |
| Local／Remote 表示转换 | 按需增加 `provider/helper` |
| 自定义校验或监控集成 | 按需使用 `validation`／`validation/errors`、`metrics` |

旧 Provider 构造器、函数注入选项、具体返回类型和 Remote JSON 继续兼容。Local 与 Remote 同为公开适配层，其具体类型保留原定义包；分页、排序 DTO 的类型身份和 JSON 同样保持。Local 入口集中于 `~/codespace/magicOrm/provider/local/api.go`，Remote 数据声明集中于 `~/codespace/magicOrm/provider/remote/declarations.go`，校验扩展合同保留在 `validation`／`models`，缓存及内置校验器迁入 `internal/validation`；指标采集迁入 `internal/metrics`。

两个 `Set*MetricsCollectorForTest` 不再作为生产接口导出；测试通过正式注册入口及独立进程验证未注册状态。业务 API 签名、公开字段／标签、常量和公开类型方法由[接口快照检查](~/codespace/magicOrm/internal/apicheck/public_api_test.go)保护，修改前需审查兼容性；复核后才可执行 `go test ./internal/apicheck -update-api` 更新基线。该检查涵盖默认与 MySQL 源文件，不代替行为测试。

当前公开面为10个功能包、168个包级函数。保留 `orm`、`models`、`provider`、`provider/local`、`provider/remote`、`provider/helper`、`validation`、`validation/errors`、`metrics`、`utils`。Local 与 Remote 地位一致，都支持类型、模型、值、过滤、编解码和场景校验扩展。

本轮有意收回七个支撑包的旧导入路径，不提供同名转发包。原 `validation/types.TypeHandler` 改用 `validation.TypeHandler`（合同定义在 `models.TypeHandler`）；内置校验器由 `validation.NewValidationFactory()` 创建，缓存通过 `ValidationConfig` 配置。`utils` 保留29个模型值／约束／时间／ID等应用函数及分页、排序 DTO，反射和深拷贝等36个辅助函数仅供内部使用。

`orm.Initialize()` 继续初始化三类指标，独立使用校验／指标时可调用 `metrics.Initialize()`；两者复用已有采集器，重复初始化不清空数据。如果监控系统稍后启动，调用 `metrics.EnsureRegistered()` 幂等注册全部指标，重复调用不会更换采集器或清空已有数据。旧 `orm.EnsureORMMetricProviderRegistered()` 保留。详见[迁移边界](~/codespace/magicOrm/docs/design-provider.md#支撑包迁移与替代入口)。业务Schema和Remote协议不变，使用被收回路径的调用方须重新编译迁移。

## 完整最小示例

以下源码与[可编译入口](~/codespace/magicOrm/examples/quickstart/main.go)保持一致。运行时必须提供独立的 `ORM_SERVER`、`ORM_DATABASE`、`ORM_SCHEMA`、`ORM_USER`、`ORM_PASSWORD`；示例创建结构和一条记录，部署 owner 负责资源生命周期。密码通过私密环境提供。

<!-- quickstart-source -->
```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/validation"
	verrors "github.com/muidea/magicOrm/validation/errors"
)

type Device struct {
	ID   int64  `orm:"id key auto" view:"detail,lite"`
	Name string `orm:"name" constraint:"req,min=2" view:"detail,lite"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// Return errors before main exits so all acquired resources are released.
func run() error {
	server, name, schema := os.Getenv("ORM_SERVER"), os.Getenv("ORM_DATABASE"), os.Getenv("ORM_SCHEMA")
	user, password := os.Getenv("ORM_USER"), os.Getenv("ORM_PASSWORD")
	if server == "" || name == "" || schema == "" || user == "" || password == "" {
		return fmt.Errorf("explicit ORM database configuration required")
	}
	// Schema provisioning is explicit and belongs to the deployment owner.
	if err := orm.EnsureSchema(server, name, schema, user, password); err != nil {
		return err
	}
	orm.Initialize()
	defer orm.Uninitialized()
	if err := orm.AddDatabase(server, name, schema, user, password, 4, "quickstart"); err != nil {
		return err
	}
	defer orm.DelDatabase("quickstart")
	p := provider.NewLocalProvider("quickstart", nil)
	model, err := p.RegisterModel(&Device{})
	if err != nil {
		return err
	}
	h, err := orm.GetOrm(context.Background(), p, "example")
	if err != nil {
		return err
	}
	defer h.Release()
	if err := h.Create(model); err != nil {
		return err
	}
	value, err := p.GetEntityModel(&Device{Name: "pump"}, true)
	if err != nil {
		return err
	}
	manager := validation.NewValidationManager(validation.DefaultConfig())
	ctx := validation.NewContext(verrors.ScenarioInsert, validation.OperationCreate, nil, "postgresql")
	if err := manager.ValidateModel(value, ctx); err != nil {
		return err
	}
	created, err := h.Insert(value)
	if err != nil {
		return err
	}
	log.Printf("created device %v", created.GetPrimaryField().GetValue().Get())
	return nil
}
```
<!-- /quickstart-source -->

## 验证入口

```bash
cd ~/codespace/magicOrm
go test ./... -count=1
go test -tags=mysql ./... -count=1
go vet ./...
go build ./...
./scripts/integration_postgres.sh
```

默认测试不连接数据库，真实数据库用例清晰 skip。隔离脚本创建专属 PostgreSQL 实例、随机口令与环回端口，退出时按本次精确容器 ID 清理并回查。可选 MySQL 验证要求专属 `magicorm_test_` 数据库和显式配置，不运行 GRANT/复制管理测试。详见[测试指南](~/codespace/magicOrm/docs/testing-guide.md)。源码、隔离 PostgreSQL、可选 MySQL、消费者与 Live 状态分别记录；不得互相替代。

## 使用资料

- [部署与监控接入](~/codespace/magicOrm/DEPLOYMENT_GUIDE.md)
- [完整使用样例与执行层级](~/codespace/magicOrm/docs/testing-guide.md#可编译使用样例与专项入口)
- [设计索引](~/codespace/magicOrm/docs/README.md)
- [ORM 与事务](~/codespace/magicOrm/docs/design-orm.md)
- [模型与约束](~/codespace/magicOrm/docs/design-models.md)
- [关系与删除](~/codespace/magicOrm/docs/design-relation.md)
- [类型映射](~/codespace/magicOrm/docs/type-mapping.md)
- [Remote Provider](~/codespace/magicOrm/docs/design-remote-provider.md)
- [平台查询合同](~/dataspace/magicPlatform/docs/technical/orm-query-processing-spec.md)
- [本批收口方案与状态](~/dataspace/magicPlatform/docs/technical/orm-capability-closeout-plan.md)
