package orm

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/database/mysql"
	"github.com/muidea/magicOrm/internal/database/postgres"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type inlineReferenceTarget struct {
	ID int64 `orm:"id key auto" view:"detail,lite"`
}

type inlineReferenceSample struct {
	ID     int64                  `orm:"id key auto" view:"detail,lite"`
	Stamp  int64                  `orm:"stamp" view:"detail,lite"`
	Target *inlineReferenceTarget `orm:"target" view:"detail,lite"`
}

func (*inlineReferenceSample) GetUniqueConstraints() []models.UniqueConstraint {
	return []models.UniqueConstraint{{Name: "uq_target_stamp", Fields: []string{"target", "stamp"}}}
}

func inlineReferenceFixture(t *testing.T, kind string, value *inlineReferenceSample) (provider.Provider, models.Model) {
	t.Helper()
	p := provider.NewLocalProvider("inline-reference", nil)
	var definition any = &inlineReferenceSample{}
	var target any = &inlineReferenceTarget{}
	var input any = value
	if kind == "remote" {
		p = provider.NewRemoteProvider("inline-reference", nil)
		object, err := helper.GetObject(definition)
		if err != nil {
			t.Fatal(err)
		}
		definition = object
		object, err = helper.GetObject(target)
		if err != nil {
			t.Fatal(err)
		}
		target = object
		objectValue, err := helper.GetObjectValue(value)
		if err != nil {
			t.Fatal(err)
		}
		input = objectValue
	}
	if _, err := p.RegisterModel(target); err != nil {
		t.Fatal(err)
	}
	if _, err := p.RegisterModel(definition); err != nil {
		t.Fatal(err)
	}
	model, err := p.GetEntityModel(input, true)
	if err != nil {
		t.Fatal(err)
	}
	return p, model
}

// 这里验证双数据库 SQL 和执行计划，不将夹具当作真实数据库唯一性证明。
func TestInlineReferenceBuildersAndSchema(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		t.Run(kind, func(t *testing.T) {
			p, model := inlineReferenceFixture(t, kind, &inlineReferenceSample{ID: 1, Stamp: 7, Target: &inlineReferenceTarget{ID: 9}})
			c := codec.New(p, "inline")
			if !models.IsInlineReference(model.Copy(models.MetaView), model.GetField("target")) {
				t.Fatal("复制丢失唯一声明")
			}
			constraints := models.GetUniqueConstraints(model)
			constraints[0].Fields[0] = "id"
			if models.GetUniqueConstraints(model)[0].Fields[0] != "target" {
				t.Fatal("声明不是防御复制")
			}
			graph, err := collectSchemaGraph(context.Background(), model, p, c)
			if err != nil || len(graph.tables) != 1 {
				t.Fatal("主表引用仍拥有链接表或目标表", graph, err)
			}
			for _, builder := range []database.Builder{postgres.NewBuilder(p, c), mysql.NewBuilder(p, c)} {
				declaration, err := builder.(database.SchemaBuilder).DescribeTable(model)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, column := range declaration.Columns {
					if column.Name == "target" {
						found = true
						if !column.Nullable || column.AutoIncrement || column.Default != "" || column.Type != "bigint" {
							t.Fatal("引用继承了目标主键的分配/默认值", column)
						}
					}
				}
				if !found {
					t.Fatal("引用列缺失")
				}
				ddl, err := builder.BuildCreateTable(model)
				if err != nil || !strings.Contains(ddl.SQL(), "uq_target_stamp") {
					t.Fatal("唯一约束缺失", err)
				}
				insert, err := builder.BuildInsert(model)
				if err != nil || !reflect.DeepEqual(insert.Args(), []any{int64(7), int64(9)}) {
					t.Fatal("引用未在宿主写入目标主键", err, insert)
				}
				update, err := builder.BuildUpdate(model)
				if err != nil || !reflect.DeepEqual(update.Args(), []any{int64(7), int64(9), int64(1)}) {
					t.Fatal("引用更新未使用同一主表语句", err, update)
				}
				filter, err := p.GetModelFilter(model)
				if err != nil {
					t.Fatal(err)
				}
				if err := filter.Equal("target", model.GetField("target").GetValue().Get()); err != nil {
					t.Fatal(err)
				}
				query, err := builder.BuildQuery(model, filter)
				if err != nil || strings.Contains(query.SQL(), "SELECT DISTINCT") || !reflect.DeepEqual(query.Args(), []any{int64(9)}) {
					t.Fatal("引用筛选仍依赖链接表", err, query)
				}
				cas, err := builder.BuildUpdateWithFilter(model, filter)
				if err != nil || !reflect.DeepEqual(cas.Args(), []any{int64(7), int64(9), int64(9)}) {
					t.Fatal("条件更新未使用引用列", err, cas)
				}
				relation, err := builder.BuildQueryRelation(model, model.GetField("target"))
				if err != nil || !strings.Contains(relation.SQL(), "IS NOT NULL") || !strings.Contains(relation.SQL(), c.ConstructModelTableName(model)) {
					t.Fatal("单项回读未使用主表列", err, relation)
				}
				batch, err := builder.BuildBatchQueryRelation(model, model.GetField("target"), []any{int64(1), int64(2)})
				if err != nil || !strings.Contains(batch.SQL(), "IS NOT NULL") || !strings.Contains(batch.SQL(), c.ConstructModelTableName(model)) {
					t.Fatal("批量回读未使用主表列", err, batch)
				}
				if err := model.SetFieldValue("target", nil); err != nil {
					t.Fatal(err)
				}
				update, err = builder.BuildUpdate(model)
				if err != nil || !reflect.DeepEqual(update.Args(), []any{int64(7), nil, int64(1)}) {
					t.Fatal("显式 null 未写入主表", err, update)
				}
				if err := model.SetFieldValue("target", &inlineReferenceTarget{ID: 9}); err != nil {
					// remote 值通过 provider 重新构造，不能跨 provider 注入本地结构。
					_, model = inlineReferenceFixture(t, kind, &inlineReferenceSample{ID: 1, Stamp: 7, Target: &inlineReferenceTarget{ID: 9}})
				}
			}
		})
	}
}

func TestInlineReferenceWriteRunnersNeverWriteLinks(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		for _, operation := range []string{"insert", "update", "clear", "delete", "invalid-insert", "invalid-update"} {
			t.Run(kind+"/"+operation, func(t *testing.T) {
				value := &inlineReferenceSample{ID: 1, Stamp: 7, Target: &inlineReferenceTarget{ID: 9}}
				if strings.HasPrefix(operation, "invalid") {
					value.Target.ID = 0
				}
				p, model := inlineReferenceFixture(t, kind, value)
				if operation == "clear" {
					if err := model.SetFieldValue("target", nil); err != nil {
						t.Fatal(err)
					}
				}
				executor := &fakeExecutor{insertIDs: []any{int64(41)}}
				handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: codec.New(p, "inline"), validationMgr: newDefaultValidationManager()}
				var err error
				switch operation {
				case "insert", "invalid-insert":
					_, e := handler.Insert(model)
					if e != nil {
						err = e
					}
				case "update", "clear", "invalid-update":
					_, e := handler.Update(model)
					if e != nil {
						err = e
					}
				case "delete":
					_, e := handler.Delete(model)
					if e != nil {
						err = e
					}
				}
				if strings.HasPrefix(operation, "invalid") {
					if err == nil || len(executor.execCalls) != 0 {
						t.Fatal("无效引用已写入", err, executor.execCalls)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				writes := 0
				for _, call := range executor.execCalls {
					if call.kind != "query" {
						writes++
						if !strings.Contains(call.sql, handler.modelCodec.ConstructModelTableName(model)) {
							t.Fatal("引用仍写链接或目标", call)
						}
					}
				}
				if writes != 1 {
					t.Fatal("引用操作不是单宿主写入", executor.execCalls)
				}
			})
		}
	}
}

func TestInlineReferenceConditionalUpdateAndReferenceFilter(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		p, model := inlineReferenceFixture(t, kind, &inlineReferenceSample{ID: 1, Stamp: 7, Target: &inlineReferenceTarget{ID: 9}})
		filter, err := p.GetModelFilter(model)
		if err != nil {
			t.Fatal(err)
		}
		if err := filter.Equal("target", model.GetField("target").GetValue().Get()); err != nil {
			t.Fatal(err)
		}
		c := codec.New(p, "inline")
		for _, builder := range []database.Builder{postgres.NewBuilder(p, c), mysql.NewBuilder(p, c)} {
			query, err := builder.BuildQuery(model, filter)
			if err != nil || strings.Contains(query.SQL(), "SELECT DISTINCT") || !reflect.DeepEqual(query.Args(), []any{int64(9)}) {
				t.Fatal("条件引用筛选错误", err, query)
			}
			undeclared := model.Copy(models.MetaView)
			if kind == "remote" {
				undeclared.(*remote.Object).UniqueConstraints = nil
				if result, err := builder.BuildCreateUniqueConstraint(undeclared, models.UniqueConstraint{Name: "uq_undeclared", Fields: []string{"target"}}); err == nil || result != nil {
					t.Fatal("独立补建约束引用了链接字段")
				}
			}
		}
		for _, affected := range []int64{0, 1} {
			executor := &fakeExecutor{execRowsAffected: affected}
			handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: c, validationMgr: newDefaultValidationManager()}
			n, err := handler.UpdateWithFilter(model, filter)
			if err != nil || n != affected || len(executor.execCalls) != 1 {
				t.Fatal("条件引用更新没有保留精确行数", n, err, executor.execCalls)
			}
		}
	}
}

func TestInlineReferenceQueryHydratesFromMainTable(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		for _, batch := range []bool{false, true} {
			for _, hasTarget := range []bool{false, true} {
				p, model := inlineReferenceFixture(t, kind, &inlineReferenceSample{ID: 1})
				c := codec.New(p, "inline")
				host := c.ConstructModelTableName(model)
				targetModel, err := p.GetTypeModel(model.GetField("target").GetType())
				if err != nil {
					t.Fatal(err)
				}
				target := c.ConstructModelTableName(targetModel)
				rows := [][]any{}
				if hasTarget {
					rows = [][]any{{int64(1), int64(9)}}
				}
				executor := &fakeExecutor{responses: []fakeQueryResponse{
					{match: func(sql string, args []any) bool {
						return strings.Contains(sql, host) && strings.Contains(sql, "IS NOT NULL")
					}, rows: rows},
					{match: func(sql string, args []any) bool { return strings.Contains(sql, host) }, rows: [][]any{{int64(1), int64(7)}}},
					{match: func(sql string, args []any) bool { return strings.Contains(sql, target) }, rows: [][]any{{int64(9)}}},
				}}
				response := model.Copy(models.DetailView)
				mask, err := buildFullQueryMaskModel(response)
				if err != nil {
					t.Fatal(err)
				}
				runner := newQueryRunner(context.Background(), mask, response, false, executor, p, c, batch, 0)
				results, err := runner.Query(nil)
				if err != nil || len(results) != 1 {
					t.Fatal("主表引用回读失败", kind, batch, hasTarget, err)
				}
				value, err := helper.GetObjectValue(results[0].Interface(true))
				if err != nil {
					t.Fatal(err)
				}
				found := value.GetFieldValue("target")
				if hasTarget {
					object, ok := found.(*remote.ObjectValue)
					if !ok || object.GetFieldValue("id") != int64(9) {
						t.Fatal("引用目标主键错误", found)
					}
				} else if found != nil {
					t.Fatal("空引用未清除查询种子", found)
				}
			}
		}
	}
}

func TestInlineReferenceAdditiveReconcileAndRepair(t *testing.T) {
	p, current := inlineReferenceFixture(t, "remote", &inlineReferenceSample{})
	previous := current.Copy(models.MetaView).(*remote.Object)
	previous.UniqueConstraints = nil
	previous.Fields = previous.Fields[:2]
	c := codec.New(p, "inline")
	plan, err := newReconcileRunner(context.Background(), previous, current, nil, p, c).planReconcile()
	if err != nil || len(plan) != 2 || !strings.Contains(plan[0].SQL(), "ADD COLUMN") || !strings.Contains(plan[1].SQL(), "ADD CONSTRAINT") {
		t.Fatal("新增引用未先补列再建唯一约束", err, plan)
	}
	builder := newBuilder(p, c).(database.SchemaBuilder)
	before, err := builder.DescribeTable(previous)
	if err != nil {
		t.Fatal(err)
	}
	desired, err := builder.DescribeTable(current)
	if err != nil {
		t.Fatal(err)
	}
	executor := &repairExecutor{schemaExecutor: &schemaExecutor{tables: map[string]database.TableSchema{before.Name: before}}}
	executor.apply = func(sql string) {
		if len(executor.execCalls) == 2 {
			executor.tables[desired.Name] = desired
		}
	}
	handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: c}
	report, err := handler.RepairSchema(previous, current)
	if err != nil || report == nil || !report.Matches || len(executor.execCalls) != 2 {
		t.Fatal("新增引用恢复不一致", report, err, executor.execCalls)
	}
}

func TestInlineReferenceStorageChangeRequiresMigration(t *testing.T) {
	p, current := inlineReferenceFixture(t, "remote", &inlineReferenceSample{})
	previous := current.Copy(models.MetaView).(*remote.Object)
	previous.UniqueConstraints = nil
	executor := &fakeExecutor{}
	err := newReconcileRunner(context.Background(), previous, current, executor, p, codec.New(p, "inline")).Reconcile()
	if err == nil || !strings.Contains(err.Message, "reference storage") || len(executor.execCalls) != 0 {
		t.Fatal("已有链接字段被静默改为主表列", err, executor.execCalls)
	}
}
