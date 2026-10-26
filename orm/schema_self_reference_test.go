package orm

import (
	"context"
	"testing"

	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type schemaSelfReference struct {
	ID     int64                `orm:"id key auto" view:"detail,lite"`
	Parent *schemaSelfReference `orm:"parent" view:"detail,lite"`
}

// 只使用 Schema 描述和执行夹具；不连接数据库，也不证明真实 DDL 已完成。
func TestSingleSelfReferenceRegistersAndBuildsBoundedSchema(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		t.Run(kind, func(t *testing.T) {
			var p provider.Provider
			var declaration any = &schemaSelfReference{}
			if kind == "remote" {
				p = provider.NewRemoteProvider("schema-self-reference", nil)
				object, err := helper.GetObject(declaration)
				if err != nil {
					t.Fatal(err)
				}
				declaration = object
			} else {
				p = provider.NewLocalProvider("schema-self-reference", nil)
			}
			model, err := p.RegisterModel(declaration)
			if err != nil {
				t.Fatal("自引用注册失败", err)
			}
			field := model.GetField("parent")
			if field == nil || models.IsSliceField(field) || !field.GetType().IsPtrType() {
				t.Fatal("单值引用所有权或基数改变")
			}
			target, err := p.GetTypeModel(field.GetType())
			if err != nil || target.GetPkgKey() != model.GetPkgKey() {
				t.Fatal("注册后无法解析自身类型", err)
			}
			c := codec.New(p, "self_reference")
			graph, err := collectSchemaGraph(context.Background(), model, p, c)
			if err != nil || graph == nil || len(graph.tables) != 2 || len(graph.order) != 2 {
				t.Fatal("自引用 Schema 不是单主表加单关系表", graph, err)
			}
			if graph.order[0].field != nil || graph.order[1].field == nil || graph.order[1].field.GetName() != "parent" {
				t.Fatal("Schema 计划重复拥有或遗漏自引用关系")
			}
			for _, create := range []bool{true, false} {
				previous, desired := model, models.Model(nil)
				if create {
					previous, desired = nil, model
				}
				if err := PreflightSchemaChange(context.Background(), previous, desired, p, "self_reference"); err != nil {
					t.Fatal("自引用 DDL 预检失败", err)
				}
				executor := &fakeExecutor{}
				if create {
					err = newCreateRunner(context.Background(), model, executor, p, c).Create()
				} else {
					err = newDropRunner(context.Background(), model, executor, p, c).Drop()
				}
				if err != nil || len(executor.execCalls) != 2 {
					t.Fatal("自引用建表/退役执行计划有重复或遗漏", err, executor.execCalls)
				}
				if executor.execCalls[0].sql == executor.execCalls[1].sql {
					t.Fatal("自引用重复执行同一表语句")
				}
			}
			executor := &schemaExecutor{tables: graph.tables}
			handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: c}
			report, err := handler.InspectSchema(model, true)
			if err != nil || report == nil || !report.Matches || len(report.Tables) != 2 || len(executor.calls) != 2 {
				t.Fatal("自引用结构核对重复或不完整", report, err)
			}
		})
	}
}
