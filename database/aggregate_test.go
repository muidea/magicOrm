package database_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/muidea/magicOrm/database"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/database/mysql"
	"github.com/muidea/magicOrm/database/postgres"
	"github.com/muidea/magicOrm/models"
)

func TestAggregateBothProvidersAndBuilders(t *testing.T) {
	for _, remoteModel := range []bool{false, true} {
		p, model := readQueryModel(t, remoteModel)
		for dialect, legacy := range map[string]database.Builder{"postgres": postgres.NewBuilder(p, codec.New(p, "aggregate")), "mysql": mysql.NewBuilder(p, codec.New(p, "aggregate"))} {
			builder, ok := legacy.(database.AggregateBuilder)
			if !ok {
				t.Fatal("缺少聚合生成能力", dialect)
			}
			t.Run(dialect+map[bool]string{false: "/local", true: "/remote"}[remoteModel], func(t *testing.T) {
				base, err := p.GetModelFilter(model)
				if err != nil {
					t.Fatal(err)
				}
				if err = base.Equal("active", false); err != nil {
					t.Fatal(err)
				}
				where := models.QueryExpression{Kind: "and", Children: []models.QueryExpression{queryCompare("rank", "gte", "integer", "0"), queryCompare("rank", "lt", "integer", "30")}}
				filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &where})
				if err != nil {
					t.Fatal(err)
				}
				plan := models.AggregatePlan{APIVersion: models.AggregatePlanAPIVersion, MaxGroups: 20, GroupBy: []string{"name"}, Metrics: []models.AggregateMetric{{Key: "all", Operation: "count"}, {Key: "valid", Operation: "count", Field: "note"}, {Key: "total", Operation: "sum", Field: "rank"}, {Key: "mean", Operation: "average", Field: "ratio"}, {Key: "first", Operation: "min", Field: "name"}, {Key: "last", Operation: "max", Field: "rank"}}}
				result, err := builder.BuildAggregate(model, filter, plan)
				if err != nil {
					t.Fatal(err)
				}
				q := "\""
				if dialect == "mysql" {
					q = "`"
				}
				for _, part := range []string{"COUNT(*)", "COUNT(" + q + "note" + q + ")", "SUM(" + q + "rank" + q + ")", "AVG(" + q + "ratio" + q + ")", "MIN(", "MAX(", ") AND (", "GROUP BY " + q + "name" + q, "LIMIT"} {
					if !strings.Contains(result.SQL(), part) {
						t.Fatal(part, result.SQL())
					}
				}
				if !reflect.DeepEqual(result.Args(), []any{false, int64(0), int64(30), 21}) {
					t.Fatal(result.Args())
				}
				if dialect == "postgres" && !strings.Contains(result.SQL(), "LIMIT $4") {
					t.Fatal(result.SQL())
				}
				conditional := plan
				conditional.Metrics = append([]models.AggregateMetric(nil), plan.Metrics...)
				for i := range conditional.Metrics {
					condition := queryCompare("rank", "gte", "integer", "7")
					conditional.Metrics[i].Where = &condition
					conditional.Metrics[i].IncludeValidSampleCount = conditional.Metrics[i].Operation == "average"
				}
				result, err = builder.BuildAggregate(model, filter, conditional)
				if err != nil {
					t.Fatal(err)
				}
				wantArgs := []any{int64(7), int64(7), int64(7), int64(7), int64(7), int64(7), int64(7), false, int64(0), int64(30), 21}
				if !reflect.DeepEqual(result.Args(), wantArgs) || strings.Count(result.SQL(), "CASE WHEN") != 7 {
					t.Fatal("条件投影参数与范围参数顺序错误", result.SQL(), result.Args())
				}
				for _, part := range []string{"THEN 1 ELSE NULL END", "AVG(CASE WHEN", "COUNT(CASE WHEN", "THEN " + q + "ratio" + q + " ELSE NULL END"} {
					if !strings.Contains(result.SQL(), part) {
						t.Fatal(part, result.SQL())
					}
				}
				if dialect == "postgres" {
					for _, part := range []string{"$1", "$7", "$8", "$9", "$10", "LIMIT $11"} {
						if !strings.Contains(result.SQL(), part) {
							t.Fatal(part, result.SQL())
						}
					}
				}
				if _, _, err := database.AggregateProjection(model, conditional, dialect); err == nil {
					t.Fatal("无参数接口接受条件投影")
				}
				for _, field := range []string{"secret", "tags", "missing", "rank) FROM hacked"} {
					bad := plan
					bad.Metrics = []models.AggregateMetric{{Key: "bad", Operation: "sum", Field: field}}
					if _, err := builder.BuildAggregate(model, filter, bad); err == nil {
						t.Fatal("接受非法聚合字段", field)
					}
				}
				base.Pagination(1, 10)
				if _, err := builder.BuildAggregate(model, filter, plan); err == nil {
					t.Fatal("把行页当完整统计")
				}
			})
		}
	}
}
