package orm

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type aggregateExecutor struct {
	fakeExecutor
	readErr  *cd.Error
	scanErr  *cd.Error
	closeErr *cd.Error
}

func (s *aggregateExecutor) RowsError() *cd.Error { return s.readErr }
func (s *aggregateExecutor) GetField(dest ...any) *cd.Error {
	if s.scanErr != nil {
		return s.scanErr
	}
	return s.fakeExecutor.GetField(dest...)
}
func (s *aggregateExecutor) Finish() {
	s.fakeExecutor.Finish()
	if s.closeErr != nil {
		s.readErr = s.closeErr
	}
}

func TestAggregateORMExecution(t *testing.T) {
	for _, remoteModel := range []bool{false, true} {
		for _, scenario := range []string{"success", "overflow", "read-failure", "scan-failure", "close-failure", "canceled", "empty", "ungrouped", "samples", "zero-samples", "bad-samples", "null-with-samples"} {
			t.Run(map[bool]string{false: "local/", true: "remote/"}[remoteModel]+scenario, func(t *testing.T) {
				p := provider.NewLocalProvider("aggregate-test", nil)
				var definition any = &readPlanRow{}
				if remoteModel {
					p = provider.NewRemoteProvider("aggregate-test", nil)
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
				base, err := p.GetModelFilter(model)
				if err != nil {
					t.Fatal(err)
				}
				if err = base.Equal("name", "scope"); err != nil {
					t.Fatal(err)
				}
				filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &models.QueryExpression{Kind: "compare", Field: "rank", Operator: "gte", Value: &models.QueryScalar{Kind: "integer", Number: "0"}}})
				if err != nil {
					t.Fatal(err)
				}
				plan := models.AggregatePlan{APIVersion: models.AggregatePlanAPIVersion, MaxGroups: 1, GroupBy: []string{"name"}, Metrics: []models.AggregateMetric{{Key: "count", Operation: "count"}, {Key: "sum", Operation: "sum", Field: "rank"}, {Key: "average", Operation: "average", Field: "rank"}}}
				cell := func(value string) sql.NullString { return sql.NullString{String: value, Valid: true} }
				rows := [][]any{{cell("scope"), cell("2"), cell("18446744073709551614.000"), cell("0.5000")}}
				if scenario == "overflow" {
					rows = append(rows, rows[0])
				}
				if scenario == "empty" {
					rows = nil
				}
				if scenario == "ungrouped" {
					plan.GroupBy = nil
					rows = [][]any{{cell("0"), sql.NullString{}, sql.NullString{}}}
				}
				if strings.Contains(scenario, "samples") {
					condition := models.QueryExpression{Kind: "compare", Field: "rank", Operator: "gte", Value: &models.QueryScalar{Kind: "integer", Number: "1"}}
					plan.Metrics[2].Where = &condition
					plan.Metrics[2].IncludeValidSampleCount = true
					rows[0] = append(rows[0], cell("18446744073709551615.000"))
					switch scenario {
					case "zero-samples":
						rows[0][3], rows[0][4] = sql.NullString{}, cell("0")
					case "bad-samples":
						rows[0][4] = cell("-1")
					case "null-with-samples":
						rows[0][3] = sql.NullString{}
					}
				}
				executor := &aggregateExecutor{fakeExecutor: fakeExecutor{responses: []fakeQueryResponse{{match: func(string, []any) bool { return true }, rows: rows}}}}
				switch scenario {
				case "read-failure":
					executor.readErr = cd.NewError(cd.Unexpected, "stream interrupted")
				case "scan-failure":
					executor.scanErr = cd.NewError(cd.Unexpected, "scan failed")
				case "close-failure":
					executor.closeErr = cd.NewError(cd.Unexpected, "close failed")
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if scenario == "canceled" {
					cancel()
				}
				handler := &impl{context: ctx, executor: executor, modelProvider: p, modelCodec: codec.New(p, "aggregate")}
				result, err := Aggregate(handler, filter, plan)
				success := scenario == "success" || scenario == "empty" || scenario == "ungrouped" || scenario == "samples" || scenario == "zero-samples"
				if !success {
					if err == nil || result != nil {
						t.Fatal("失败返回部分统计", result, err)
					}
					if scenario == "canceled" && len(executor.execCalls) != 0 {
						t.Fatal("取消后仍执行")
					}
					return
				}
				if err != nil || result == nil {
					t.Fatal(result, err)
				}
				if scenario == "empty" {
					if len(result.Rows) != 0 {
						t.Fatal(result)
					}
					return
				}
				if len(result.Rows) != 1 {
					t.Fatal(result)
				}
				if scenario == "ungrouped" {
					if result.Rows[0].Metrics[0].Value.Number != "0" || result.Rows[0].Metrics[1].Value.Kind != "null" {
						t.Fatal(result)
					}
					return
				}
				if scenario == "samples" || scenario == "zero-samples" {
					metric := result.Rows[0].Metrics[2]
					want := "18446744073709551615"
					if scenario == "zero-samples" {
						want = "0"
					}
					if metric.ValidSampleCount == nil || metric.ValidSampleCount.Number != want || !strings.Contains(executor.execCalls[0].sql, "AVG(CASE WHEN") {
						t.Fatal("样本数或条件聚合丢失", metric)
					}
					return
				}
				if result.Rows[0].Metrics[1].Value.Number != "18446744073709551614" || result.Rows[0].Metrics[2].Value.Number != "0.5000" {
					t.Fatal("聚合数值丢失精度", result)
				}
				call := executor.execCalls[0]
				if !strings.Contains(call.sql, ") AND (") || len(call.args) != 3 || call.args[0] != "scope" || call.args[1] != int64(0) {
					t.Fatal("聚合丢失scope", call)
				}
			})
		}
	}
}
