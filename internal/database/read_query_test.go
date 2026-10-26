package database_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/internal/database/mysql"
	"github.com/muidea/magicOrm/internal/database/postgres"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

type readQueryRecord struct {
	ID     int64     `orm:"id key" view:"detail,lite"`
	Rank   int64     `orm:"rank" view:"detail,lite"`
	Small  int8      `orm:"small" view:"detail,lite"`
	Name   string    `orm:"name" view:"detail,lite"`
	Active bool      `orm:"active" view:"detail,lite"`
	Note   *string   `orm:"note" view:"detail,lite"`
	Secret string    `orm:"secret" constraint:"wo"`
	Stamp  time.Time `orm:"stamp" view:"detail,lite"`
	Ratio  float64   `orm:"ratio" view:"detail,lite"`
	Tags   []string  `orm:"tags" view:"detail,lite"`
}

func readQueryModel(t *testing.T, remoteModel bool) (provider.Provider, models.Model) {
	t.Helper()
	if !remoteModel {
		p := provider.NewLocalProvider("query-test", nil)
		model, err := p.RegisterModel(&readQueryRecord{})
		if err != nil {
			t.Fatal(err)
		}
		return p, model
	}
	p := provider.NewRemoteProvider("query-test", nil)
	field := func(name, kind string) *remote.Field {
		return &remote.Field{Name: name, Type: &remote.TypeImpl{Name: kind}, Spec: &remote.SpecImpl{ViewDeclare: []models.ViewDeclare{models.DetailView, models.LiteView}}}
	}
	id, rank, small := field("id", "int64"), field("rank", "int64"), field("small", "int8")
	name, active, note, secret := field("name", "string"), field("active", "boolean"), field("note", "string"), field("secret", "string")
	id.Spec.PrimaryKey, note.Type.IsPtr, secret.Spec.Constraint = true, true, "wo"
	stamp, ratio := field("stamp", "datetime"), field("ratio", "float64")
	tags := field("tags", "array")
	tags.Type.Value, tags.Type.ElemType = models.TypeSliceValue, &remote.TypeImpl{Name: "string", Value: models.TypeStringValue}
	model, err := p.RegisterModel(&remote.Object{Name: "record", PkgPath: "/query", Fields: []*remote.Field{id, rank, small, name, active, note, secret, stamp, ratio, tags}})
	if err != nil {
		t.Fatal(err)
	}
	return p, model
}

func queryCompare(field, op, kind, value string) models.QueryExpression {
	scalar := &models.QueryScalar{Kind: kind}
	if kind == "integer" || kind == "decimal" {
		scalar.Number = value
	} else if kind == "string" {
		scalar.Text = value
	}
	return models.QueryExpression{Kind: "compare", Field: field, Operator: op, Value: scalar}
}

func TestReadQueryPlanBothProvidersAndBuilders(t *testing.T) {
	for _, remoteModel := range []bool{false, true} {
		p, model := readQueryModel(t, remoteModel)
		builders := map[string]database.Builder{"postgres": postgres.NewBuilder(p, codec.New(p, "query")), "mysql": mysql.NewBuilder(p, codec.New(p, "query"))}
		for name, builder := range builders {
			t.Run(name+map[bool]string{false: "/local", true: "/remote"}[remoteModel], func(t *testing.T) {
				base, err := p.GetModelFilter(model)
				if err != nil {
					t.Fatal(err)
				}
				if err := base.Equal("active", false); err != nil {
					t.Fatal(err)
				}
				base.Pagination(1, 21)
				order := []models.QueryOrder{{Field: "rank", Direction: "desc"}, {Field: "id", Direction: "asc"}}
				after, err := models.QueryKeyset(order, []models.QueryScalar{{Kind: "integer", Number: "20"}, {Kind: "integer", Number: "9223372036854775001"}})
				if err != nil {
					t.Fatal(err)
				}
				where := models.QueryExpression{Kind: "and", Children: []models.QueryExpression{queryCompare("rank", "gte", "integer", "0"), queryCompare("rank", "lt", "integer", "30"), queryCompare("name", "eq", "string", "x' OR 1=1 --"), *after}}
				filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &where, OrderBy: order})
				if err != nil {
					t.Fatal(err)
				}
				// 追加计划和读取副本均不能共享可变条件。
				where.Children[0].Value.Number = "999"
				copy := filter.(models.ReadQueryPlanProvider).ReadQueryPlan()
				copy.OrderBy[0].Field = "secret"
				result, err := builder.BuildQuery(model, filter)
				if err != nil {
					t.Fatal(err)
				}
				quote := "\""
				if name == "mysql" {
					quote = "`"
				}
				q := func(field string) string { return quote + field + quote }
				for _, part := range []string{q("active"), q("rank") + " >=", q("rank") + " <", " OR ", q("id") + " >", "ORDER BY " + q("rank") + " DESC, " + q("id") + " ASC", "LIMIT"} {
					if !strings.Contains(result.SQL(), part) {
						t.Fatal("组合条件或稳定排序丢失", part, result.SQL())
					}
				}
				if strings.Contains(result.SQL(), "OR 1=1") || strings.Contains(result.SQL(), "9223372036854775001") {
					t.Fatal("查询值直接进入 SQL")
				}
				args := result.Args()
				if len(args) != 9 || args[1] != int64(0) || args[2] != int64(30) || args[3] != "x' OR 1=1 --" || args[6] != int64(9223372036854775001) {
					t.Fatal("双边界、零值、原始整数或参数顺序变化", args)
				}
				if name == "postgres" && (!strings.Contains(result.SQL(), "$7") || !strings.Contains(result.SQL(), "LIMIT $8 OFFSET $9")) {
					t.Fatal("追加计划覆盖旧参数编号", result.SQL())
				}
				count, err := builder.BuildCount(model, filter)
				if err != nil || !reflect.DeepEqual(count.Args(), args[:7]) || strings.Contains(count.SQL(), "ORDER BY") || strings.Contains(count.SQL(), "LIMIT") {
					t.Fatal("统计未复用同一完整过滤条件", count, err)
				}
				if _, err := builder.BuildUpdateWithFilter(model, filter); err == nil {
					t.Fatal("读计划成为条件写入许可")
				}
			})
		}
	}
}

func TestReadQueryScalarTimeDecimalAndPageBudgets(t *testing.T) {
	for _, remoteModel := range []bool{false, true} {
		p, model := readQueryModel(t, remoteModel)
		for dialect, builder := range map[string]database.Builder{"postgres": postgres.NewBuilder(p, codec.New(p, "query")), "mysql": mysql.NewBuilder(p, codec.New(p, "query"))} {
			for _, tc := range []struct {
				name, field, kind, token string
				valid                    bool
			}{
				{"utc", "stamp", "datetime", "2026-10-19T12:34:56.123Z", true},
				{"offset", "stamp", "datetime", "2026-10-19T20:34:56.123+08:00", true},
				{"sub-millisecond", "stamp", "datetime", "2026-10-19T12:34:56.123456Z", false},
				{"bad-time", "stamp", "datetime", "2026-02-30T00:00:00Z", false},
				{"negative-decimal", "ratio", "decimal", "-1.25e2", true},
				{"nan", "ratio", "decimal", "NaN", false},
				{"infinity", "ratio", "decimal", "1e9999", false},
				{"array", "tags", "string", "not-a-scalar", false},
			} {
				t.Run(dialect+map[bool]string{false: "/local/", true: "/remote/"}[remoteModel]+tc.name, func(t *testing.T) {
					base, err := p.GetModelFilter(model)
					if err != nil {
						t.Fatal(err)
					}
					e := queryCompare(tc.field, "gte", tc.kind, tc.token)
					if tc.kind == "datetime" {
						e.Value.Text = tc.token
					}
					filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &e})
					if err != nil {
						t.Fatal(err)
					}
					_, _, args, err := database.BuildReadQueryClause(model, filter, p, dialect, 0)
					if (err == nil) != tc.valid {
						t.Fatal("标量格式或精度校验错误", args, err)
					}
					if tc.valid && tc.kind == "datetime" && (len(args) != 1 || args[0] != "2026-10-19 12:34:56.123") {
						t.Fatal("时间未规范化为相同 UTC 毫秒", args)
					}
					if tc.valid && tc.kind == "decimal" && (len(args) != 1 || args[0] != float64(-125)) {
						t.Fatal("负小数编码变化", args)
					}
				})
			}
			for _, size := range []int{0, 1001, 1002} {
				base, err := p.GetModelFilter(model)
				if err != nil {
					t.Fatal(err)
				}
				if size != 0 {
					base.Pagination(1, size)
				}
				filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := builder.BuildQuery(model, filter); (err == nil) != (size == 1001) {
					t.Fatal("新增读计划行预算未执行", size, err)
				}
				if _, err := builder.BuildCount(model, filter); err != nil {
					t.Fatal("Count错误复用了返回行页限制", err)
				}
				if _, err := builder.BuildQuery(model, base); err != nil {
					t.Fatal("旧Filter行为变化", err)
				}
			}
			base, err := p.GetModelFilter(model)
			if err != nil {
				t.Fatal(err)
			}
			base.Pagination(12, 1000)
			filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := builder.BuildQuery(model, filter); err == nil {
				t.Fatal("新增读计划深分页未拒绝")
			}
		}
	}
}

func TestReadQueryPlanRejectsInvalidAndKeepsLegacy(t *testing.T) {
	for _, remoteModel := range []bool{false, true} {
		p, model := readQueryModel(t, remoteModel)
		for _, dialect := range []string{"mysql", "postgres"} {
			for _, fault := range []string{"unknown", "injection-field", "write-only", "overflow", "fraction-integer", "wrong-kind", "null-comparison", "bad-operator", "mixed-value", "nullable-order", "no-primary-order", "duplicate-order", "legacy-sort", "too-deep", "too-many", "future", "null-ok", "zero-false"} {
				t.Run(dialect+"/"+map[bool]string{false: "local/", true: "remote/"}[remoteModel]+fault, func(t *testing.T) {
					base, err := p.GetModelFilter(model)
					if err != nil {
						t.Fatal(err)
					}
					e := queryCompare("rank", "gte", "integer", "0")
					plan := models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &e, OrderBy: []models.QueryOrder{{Field: "id", Direction: "asc"}}}
					switch fault {
					case "unknown":
						e.Field = "missing"
					case "injection-field":
						e.Field = "rank\" OR 1=1 --"
					case "write-only":
						e = queryCompare("secret", "eq", "string", "private")
					case "overflow":
						e = queryCompare("small", "eq", "integer", "128")
					case "fraction-integer":
						e.Value.Number = "1.5"
					case "wrong-kind":
						e.Value = &models.QueryScalar{Kind: "string", Text: "0"}
					case "null-comparison":
						e.Value = &models.QueryScalar{Kind: "null"}
					case "bad-operator":
						e.Operator = "sql"
					case "mixed-value":
						e.Value.Text = "untrusted"
					case "nullable-order":
						plan.OrderBy = []models.QueryOrder{{Field: "note", Direction: "asc"}, {Field: "id", Direction: "asc"}}
					case "no-primary-order":
						plan.OrderBy[0].Field = "rank"
					case "duplicate-order":
						plan.OrderBy = append(plan.OrderBy, plan.OrderBy[0])
					case "legacy-sort":
						base.Sort("id", true)
					case "too-deep":
						for i := 0; i < 9; i++ {
							e = models.QueryExpression{Kind: "and", Children: []models.QueryExpression{e, queryCompare("id", "gt", "integer", "0")}}
						}
					case "too-many":
						e = models.QueryExpression{Kind: "or", Children: make([]models.QueryExpression, 17)}
					case "future":
						plan.APIVersion += "-future"
					case "null-ok":
						e = models.QueryExpression{Kind: "compare", Field: "note", Operator: "is-null", Value: &models.QueryScalar{Kind: "null"}}
					case "zero-false":
						e = models.QueryExpression{Kind: "compare", Field: "active", Operator: "eq", Value: &models.QueryScalar{Kind: "boolean"}}
					}
					filter, err := models.WithReadQueryPlan(base, plan)
					if err == nil {
						_, _, _, err = database.BuildReadQueryClause(model, filter, p, dialect, 0)
					}
					if (err == nil) != (fault == "null-ok" || fault == "zero-false") {
						t.Fatal("无效读计划未拒绝或零值被拒绝", err)
					}
				})
			}
		}
	}
}
