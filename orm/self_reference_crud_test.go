package orm

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func selfReferenceValueModel(t *testing.T, kind string, value *schemaSelfReference) (provider.Provider, models.Model) {
	t.Helper()
	p := provider.NewLocalProvider("self-reference", nil)
	var definition any = &schemaSelfReference{}
	var input any = value
	if kind == "remote" {
		p = provider.NewRemoteProvider("self-reference", nil)
		object, err := helper.GetObject(definition)
		if err != nil {
			t.Fatal(err)
		}
		definition = object
		objectValue, err := helper.GetObjectValue(value)
		if err != nil {
			t.Fatal(err)
		}
		input = objectValue
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

// 这里只验证 ORM 执行计划和事务边界，不模拟数据库提交或树业务约束。
func TestSingleSelfReferenceWritesOnlyHostAndLinks(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		for _, operation := range []string{"insert-root", "insert-null", "insert-reference", "update-self-id", "clear", "delete"} {
			t.Run(kind+"/"+operation, func(t *testing.T) {
				value := &schemaSelfReference{ID: 1}
				if strings.HasPrefix(operation, "insert") {
					value.ID = 0
				}
				if operation == "insert-reference" {
					value.Parent = &schemaSelfReference{ID: 9}
				} else if operation == "update-self-id" {
					// 同主键的有限值不是 Go 指针环，也不证明业务树允许自指。
					value.Parent = &schemaSelfReference{ID: 1}
				}
				p, model := selfReferenceValueModel(t, kind, value)
				if operation == "clear" || operation == "insert-null" {
					if err := model.SetFieldValue("parent", nil); err != nil {
						t.Fatal(err)
					}
				}
				c := codec.New(p, "self_reference")
				relationTable, err := c.ConstructRelationTableName(model, model.GetField("parent"))
				if err != nil {
					t.Fatal(err)
				}
				executor := &fakeExecutor{insertIDs: []any{int64(41)}, responses: []fakeQueryResponse{{
					match: func(sql string, args []any) bool {
						return strings.Contains(sql, relationTable) && reflect.DeepEqual(args, []any{int64(1)})
					},
					rows: [][]any{{int64(9)}},
				}}}
				handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: c, validationMgr: newDefaultValidationManager()}
				switch operation {
				case "insert-root", "insert-null", "insert-reference":
					var result models.Model
					result, err = handler.Insert(model)
					if err == nil && result.GetPrimaryField().GetValue().Get() != int64(41) {
						t.Fatal("自引用改变了自动主键分配")
					}
				case "update-self-id", "clear":
					_, err = handler.Update(model)
				case "delete":
					_, err = handler.Delete(model)
				}
				if err != nil {
					t.Fatal(err)
				}
				if executor.beginCalls != 1 || executor.commitCalls != 1 || executor.rollbackCalls != 0 {
					t.Fatalf("关系事务边界错误: %+v", executor)
				}
				writes := []fakeExecCall{}
				for _, call := range executor.execCalls {
					if call.kind != "query" {
						writes = append(writes, call)
					}
				}
				want := 2
				if operation == "insert-root" || operation == "insert-null" || operation == "clear" {
					want = 1
				}
				if len(writes) != want {
					t.Fatalf("引用目标被额外写入或链接遗漏: %+v", writes)
				}
				for idx, call := range writes {
					link := strings.Contains(call.sql, relationTable)
					wantLink := operation == "clear" || operation == "update-self-id" || idx == 1
					if link != wantLink {
						t.Fatalf("引用写入越过宿主/链接边界: %+v", writes)
					}
				}
				if operation == "insert-reference" && !reflect.DeepEqual(writes[1].args, []any{int64(41), int64(9)}) {
					t.Fatalf("链接未使用分配后的主键: %+v", writes)
				}
				if operation == "update-self-id" && (!reflect.DeepEqual(writes[0].args, []any{int64(1), int64(9)}) || !reflect.DeepEqual(writes[1].args, []any{int64(1), int64(1)})) {
					t.Fatalf("引用更新未按主键差集写入: %+v", writes)
				}
				if operation == "clear" && !reflect.DeepEqual(writes[0].args, []any{int64(1), int64(9)}) {
					t.Fatalf("清空引用错误: %+v", writes)
				}
			})
		}
	}
}

func TestSingleSelfReferenceInsertRejectsMissingTargetKey(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		t.Run(kind, func(t *testing.T) {
			p, model := selfReferenceValueModel(t, kind, &schemaSelfReference{Parent: &schemaSelfReference{}})
			executor := &fakeExecutor{insertIDs: []any{int64(41)}}
			handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: codec.New(p, "self_reference"), validationMgr: newDefaultValidationManager()}
			result, err := handler.Insert(model)
			if err == nil || err.Code != cd.IllegalParam || result != nil {
				t.Fatal("无主键的引用目标被当作合法关系", err)
			}
			if len(executor.execCalls) != 0 || executor.commitCalls != 0 {
				t.Fatalf("无效引用已写入宿主或链接: %+v", executor)
			}
		})
	}
}

func TestSingleSelfReferenceUpdateRejectsMissingKeyBeforeLinkChanges(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		t.Run(kind, func(t *testing.T) {
			p, model := selfReferenceValueModel(t, kind, &schemaSelfReference{ID: 1, Parent: &schemaSelfReference{}})
			executor := &fakeExecutor{}
			handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: codec.New(p, "self_reference"), validationMgr: newDefaultValidationManager()}
			result, err := handler.Update(model)
			if err == nil || err.Code != cd.IllegalParam || result != nil || len(executor.execCalls) != 0 || executor.commitCalls != 0 {
				t.Fatal("无主键的目标导致旧链接被修改", err, executor.execCalls)
			}
		})
	}
}

func TestInsertReferenceCollectionsValidateEveryTargetBeforeHost(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		for _, valid := range []bool{false, true} {
			name := kind + "/missing-key"
			secondID := 0
			if valid {
				name, secondID = kind+"/valid", 3
			}
			t.Run(name, func(t *testing.T) {
				p := provider.NewLocalProvider("reference-collection", nil)
				var definition any = &responseCycleNode{}
				var input any = &responseCycleNode{ID: 1, Children: []*responseCycleNode{{ID: 2}, {ID: secondID}}}
				if kind == "remote" {
					p = provider.NewRemoteProvider("reference-collection", nil)
					object, err := helper.GetObject(definition)
					if err != nil {
						t.Fatal(err)
					}
					definition = object
					value, err := helper.GetObjectValue(input)
					if err != nil {
						t.Fatal(err)
					}
					input = value
				}
				if _, err := p.RegisterModel(definition); err != nil {
					t.Fatal(err)
				}
				model, err := p.GetEntityModel(input, true)
				if err != nil {
					t.Fatal(err)
				}
				executor := &fakeExecutor{}
				handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: codec.New(p, "self_reference"), validationMgr: newDefaultValidationManager()}
				result, err := handler.Insert(model)
				if valid {
					if err != nil || result == nil || len(executor.execCalls) != 3 || executor.commitCalls != 1 {
						t.Fatal("合法引用集合被拒绝", err, executor.execCalls)
					}
				} else if err == nil || err.Code != cd.IllegalParam || result != nil || len(executor.execCalls) != 0 || executor.commitCalls != 0 {
					t.Fatal("无效引用集合已写入宿主或部分链接", err, executor.execCalls)
				}
			})
		}
	}
}

func TestSingleSelfReferenceRelationFailureRollsBackHost(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		t.Run(kind, func(t *testing.T) {
			p, model := selfReferenceValueModel(t, kind, &schemaSelfReference{Parent: &schemaSelfReference{ID: 9}})
			executor := &failingRelationExecutor{fakeExecutor: fakeExecutor{insertIDs: []any{int64(41)}}}
			handler := &impl{context: context.Background(), executor: executor, modelProvider: p, modelCodec: codec.New(p, "self_reference"), validationMgr: newDefaultValidationManager()}
			result, err := handler.Insert(model)
			if err == nil || result != nil || executor.beginCalls != 1 || executor.commitCalls != 0 || executor.rollbackCalls != 1 {
				t.Fatal("链接失败未回滚宿主事务", err)
			}
		})
	}
}

type selfReferenceQueryExecutor struct {
	fakeExecutor
	relationTable string
	edges         map[int64]int64
	roots         []int64
}

func (e *selfReferenceQueryExecutor) Query(sql string, _ bool, args ...any) ([]string, *cd.Error) {
	if len(e.execCalls) >= 32 {
		return nil, cd.NewError(cd.Unexpected, "自引用查询未有界终止")
	}
	e.execCalls = append(e.execCalls, fakeExecCall{kind: "query", sql: sql, args: append([]any(nil), args...)})
	e.currentRows, e.index = nil, -1
	ids := e.roots
	if len(args) > 0 {
		ids = nil
		for _, arg := range args {
			id, ok := arg.(int64)
			if !ok {
				return nil, cd.NewError(cd.Unexpected, "查询主键类型改变")
			}
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		if strings.Contains(sql, e.relationTable) {
			if target, ok := e.edges[id]; ok {
				if strings.Contains(sql, " IN (") {
					e.currentRows = append(e.currentRows, []any{id, target})
				} else {
					e.currentRows = append(e.currentRows, []any{target})
				}
			}
		} else {
			e.currentRows = append(e.currentRows, []any{id})
		}
	}
	return nil, nil
}

func TestSingleSelfReferenceQueryProducesFiniteValues(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		for _, mode := range []string{"root", "chain", "deep-chain", "self-id", "two-ids"} {
			for _, batch := range []bool{false, true} {
				name := kind + "/" + mode + "/single"
				if batch {
					name = kind + "/" + mode + "/batch"
				}
				t.Run(name, func(t *testing.T) {
					p, model := selfReferenceValueModel(t, kind, &schemaSelfReference{ID: 1})
					c := codec.New(p, "self_reference")
					table, err := c.ConstructRelationTableName(model, model.GetField("parent"))
					if err != nil {
						t.Fatal(err)
					}
					edges := map[int64]int64{}
					switch mode {
					case "chain":
						edges[1], edges[2] = 2, 3
					case "deep-chain":
						for id := int64(1); id < 20; id++ {
							edges[id] = id + 1
						}
					case "self-id":
						edges[1] = 1
					case "two-ids":
						edges[1], edges[2] = 2, 1
					}
					executor := &selfReferenceQueryExecutor{relationTable: table, edges: edges, roots: []int64{1}}
					if batch {
						executor.roots = []int64{1, 2}
					}
					response := model.Copy(models.DetailView)
					mask, err := buildFullQueryMaskModel(response)
					if err != nil {
						t.Fatal(err)
					}
					runner := NewQueryRunner(context.Background(), mask, response, false, executor, p, c, batch, 0)
					results, err := runner.Query(nil)
					if err != nil || len(results) != len(executor.roots) {
						t.Fatal("自引用查询失败", err)
					}
					for _, result := range results {
						value, err := helper.GetObjectValue(result.Interface(true))
						if err != nil {
							t.Fatal("查询构造了实际值图环", err)
						}
						if _, err := json.Marshal(value); err != nil {
							t.Fatal("查询结果不可序列化", err)
						}
						if mode == "root" && value.GetFieldValue("parent") != nil {
							t.Fatal("未持久化的引用被查询种子替代")
						}
						id := result.GetPrimaryField().GetValue().Get().(int64)
						if target, exists := edges[id]; exists {
							parent, ok := value.GetFieldValue("parent").(*remote.ObjectValue)
							if !ok || parent.GetFieldValue("id") != target {
								t.Fatal("查询未保留精确目标主键")
							}
							if mode == "chain" && id == 1 {
								ancestor, ok := parent.GetFieldValue("parent").(*remote.ObjectValue)
								if !ok || ancestor.GetFieldValue("id") != int64(3) || ancestor.GetFieldValue("parent") != nil {
									t.Fatal("有限链祖先丢失或根引用不为 null")
								}
							}
						}
					}
				})
			}
		}
	}
}
