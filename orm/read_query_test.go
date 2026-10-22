package orm

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

type readPlanRow struct {
	ID          int64  `orm:"id key" view:"detail,lite"`
	Rank        int64  `orm:"rank" view:"detail,lite"`
	Name        string `orm:"name" view:"detail,lite"`
	Description string `orm:"description" view:"detail"`
}

// 真实 ORM 入口与执行替身共同证明计划传播；替身不证明数据库排序或事务。
func TestReadQueryPlanExplicitValueMask(t *testing.T) {
	p := provider.NewRemoteProvider("read-mask", nil)
	registerQueryMaskViewRemoteModel(t, p)
	base, err := p.GetEntityFilter(&remote.ObjectValue{Name: "queryMaskViewModel", PkgPath: "github.com/muidea/magicOrm/orm"}, models.LiteView)
	if err != nil {
		t.Fatal(err)
	}
	if err := base.Equal("id", int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := base.ValueMask(&remote.ObjectValue{Name: "queryMaskViewModel", PkgPath: "github.com/muidea/magicOrm/orm", Fields: []*remote.FieldValue{{Name: "description", Value: "", Assigned: true}}}); err != nil {
		t.Fatal(err)
	}
	base.Pagination(1, 2)
	filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion})
	if err != nil {
		t.Fatal(err)
	}
	exec := &fakeExecutor{responses: []fakeQueryResponse{{match: func(string, []any) bool { return true }, rows: [][]any{{int64(1), "detail-description"}}}}}
	handler := &impl{context: context.Background(), executor: exec, modelProvider: p, modelCodec: codec.New(p, "read_mask")}
	rows, err := handler.BatchQuery(filter)
	if err != nil || len(rows) != 1 {
		t.Fatal(rows, err)
	}
	value := rows[0].Interface(true).(*remote.ObjectValue)
	if value.GetFieldValue("name") != nil || value.GetFieldValue("description") != "detail-description" || value.GetFieldValue("id") != int64(1) {
		t.Fatal("读计划丢失显式ValueMask", value)
	}
}

func TestReadQueryPlanORMExecutionAndProjection(t *testing.T) {
	for _, kind := range []string{"local", "remote"} {
		for _, canceled := range []bool{false, true} {
			t.Run(kind+map[bool]string{false: "/read", true: "/canceled"}[canceled], func(t *testing.T) {
				p := provider.NewLocalProvider("read-plan", nil)
				var definition any = &readPlanRow{}
				if kind == "remote" {
					p = provider.NewRemoteProvider("read-plan", nil)
					object, err := helper.GetObject(definition)
					if err != nil {
						t.Fatal(err)
					}
					definition = object
				}
				model, err := p.RegisterModel(definition)
				if err != nil {
					t.Fatal(err)
				}
				filter, err := p.GetEntityFilter(definition, models.LiteView)
				if err != nil {
					t.Fatal(err)
				}
				if err := filter.Equal("name", "scope"); err != nil {
					t.Fatal(err)
				}
				filter.Pagination(1, 21)
				where := models.QueryExpression{Kind: "or", Children: []models.QueryExpression{
					{Kind: "compare", Field: "rank", Operator: "gte", Value: &models.QueryScalar{Kind: "integer", Number: "0"}},
					{Kind: "compare", Field: "id", Operator: "gt", Value: &models.QueryScalar{Kind: "integer", Number: "9223372036854775001"}},
				}}
				wrapped, err := models.WithReadQueryPlan(filter, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &where, OrderBy: []models.QueryOrder{{Field: "rank", Direction: "desc"}, {Field: "id", Direction: "asc"}}})
				if err != nil {
					t.Fatal(err)
				}
				exec := &fakeExecutor{responses: []fakeQueryResponse{
					{match: func(q string, _ []any) bool { return strings.Contains(q, "COUNT(*)") }, rows: [][]any{{sql.NullInt64{Int64: 1, Valid: true}}}},
					{match: func(q string, _ []any) bool { return strings.Contains(q, "SELECT") }, rows: [][]any{{int64(8), int64(2), "scope", "private-detail"}}},
				}}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if canceled {
					cancel()
				}
				handler := &impl{context: ctx, executor: exec, modelProvider: p, modelCodec: codec.New(p, "read_plan")}
				count, err := handler.Count(wrapped)
				if canceled {
					if err == nil {
						t.Fatal("取消后仍统计")
					}
				} else if err != nil || count != 1 {
					t.Fatal("真实Count入口失败", count, err)
				}
				rows, err := handler.BatchQuery(wrapped)
				if canceled {
					if err == nil || len(exec.execCalls) != 0 {
						t.Fatal("取消后仍进入执行器", err, exec.execCalls)
					}
					return
				}
				if err != nil || len(rows) != 1 {
					t.Fatal("真实BatchQuery入口失败", len(rows), err)
				}
				if len(exec.execCalls) != 2 {
					t.Fatal("重复执行或遗漏查询", exec.execCalls)
				}
				for i, call := range exec.execCalls {
					if !strings.Contains(call.sql, ") AND (") || !strings.Contains(call.sql, " OR ") {
						t.Fatal("scope或OR表达式丢失", call.sql)
					}
					want := []any{"scope", int64(0), int64(9223372036854775001)}
					if i == 1 {
						want = append(want, int64(21), int64(0))
						if !strings.Contains(call.sql, "ORDER BY") {
							t.Fatal("执行丢失排序")
						}
					}
					if !reflect.DeepEqual(call.args, want) {
						t.Fatal("执行参数丢失或大整数变位", call.args, want)
					}
				}
				response, byMask, err := buildQueryResponseModel(nil, filter)
				if err != nil {
					t.Fatal(err)
				}
				wrappedResponse, wrappedByMask, err := buildQueryResponseModel(nil, wrapped)
				if err != nil || byMask != wrappedByMask || !reflect.DeepEqual(response.Interface(true), wrappedResponse.Interface(true)) {
					t.Fatal("追加计划改变原投影", err)
				}
				// 与未包装 Filter 的实际返回逐字段比对，不借此扩展改变旧视图语义。
				legacyRows, err := handler.BatchQuery(filter)
				if err != nil || len(legacyRows) != 1 || !reflect.DeepEqual(rows[0].Interface(true), legacyRows[0].Interface(true)) {
					t.Fatal("追加计划改变旧Filter返回投影", err)
				}
				if kind == "remote" {
					if value := rows[0].GetField("description").GetValue().Get(); value != nil && value != "" {
						t.Fatal("远程lite投影泄漏detail字段", value)
					}
				}
				if err := model.SetFieldValue("name", "changed"); err != nil {
					t.Fatal(err)
				}
				if _, err := handler.UpdateWithFilter(model, wrapped); err == nil || len(exec.execCalls) != 3 {
					t.Fatal("读计划触发了业务写", err, exec.execCalls)
				}
			})
		}
	}
}
