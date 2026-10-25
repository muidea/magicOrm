# magicOrm

magicOrm 是面向 Go 的模型 ORM，默认 PostgreSQL，使用 `-tags=mysql` 构建 MySQL。Local Provider 处理 Go 模型，Remote Provider 处理动态模型声明与值；Remote 不内置网络传输。同一 ORM handler 由调用方串行使用。

## 能力与边界

- CRUD、Query 单查、BatchQuery 列表、视图与掩码、显式赋值、条件更新与精确主键行锁。Query 主键优先，单查要求唯一命中；读计划不能授权写入。
- 独立引用与递归 owned 关系；组合唯一支持基础字段和单值指针引用。参与组合唯一的引用以目标标量主键内联主表，不双写链接表。写引用检查非零主键，不证明目标存在。目标删除保护、租户隔离、授权和业务防环属于 owner。
- Local/Remote 唯一声明；Remote 普通索引。Local 标准入口没有 `GetIndexes`；多值关系和 owned 字段不能参与组合唯一。声明准入不保证任意字段均可在两类数据库物理索引。
- 普通嵌套事务共享一个物理事务；内层回滚使整体只能回滚。Serializable 必须在空闲时开启，不降级、不自动重试。PostgreSQL 唯一冲突映射 Duplicated，序列化/死锁映射 VersionConflict，连接丢失及提交未知保持 Unexpected，由 owner 回查原身份。
- Create/Drop/Inspect/Repair 管理宿主、关系表及完整递归 owned 图，保留独立引用目标主表。业务 owner 证明物理命名空间独占、图外使用者保护并串行 Schema 写入；ORM 不维护平台所有权。
- Reconcile 只处理安全增量；必填新字段、删字段、存储类型切换等要求显式迁移。`EnsureSchema` 显式创建 Schema，`AddDatabase` 不创建 Schema，不接受 database/schema 合并名称。
- UTC 毫秒持久化；无时区旧文本按 UTC，不自动转换历史本地时间。精确 JSON 数值不经浮点中转。
- 受限 AND/OR/null、多字段排序、keyset、注册单值引用子查询；独立聚合支持至多 16 指标、4 分组维、1000 组，UTC 小时桶至多 744 桶、5000 点。不是任意 SQL、授权游标或跨查询快照。
- 类型、约束、数据库兼容和场景校验；校验层不查询 UNIQUE/FK，物理唯一性由数据库执行。缓存参数与显式选项见[校验合同](~/codespace/magicOrm/docs/design-validation.md)。

PostgreSQL 约束和索引的物理名由表名与声明名两段身份派生，统一附加稳定摘要并限制长度，避免同 Schema 不同实例发生索引名碰撞。现有无表限定的旧物理名须显式迁移；本批使用干净候选 Schema，不自动重命名旧索引。

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
