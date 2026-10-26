package orm

import (
	"context"
	"database/sql"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
)

type aggregateBucketRecord struct {
	ID    int64     `orm:"id key"`
	Stamp time.Time `orm:"stamp"`
	Ratio *float64  `orm:"ratio"`
}

func TestAggregateBucketORMExecution(t *testing.T) {
	for _, remoteModel := range []bool{false, true} {
		for _, scenario := range []string{"sparse", "empty", "duplicate", "outside", "null-index", "reversed", "read-failure", "close-failure", "canceled"} {
			t.Run(map[bool]string{false: "local/", true: "remote/"}[remoteModel]+scenario, func(t *testing.T) {
				p := provider.NewLocalProvider("bucket-test", nil)
				var definition any = &aggregateBucketRecord{}
				if remoteModel {
					p = provider.NewRemoteProvider("bucket-test", nil)
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
				filter, err := p.GetModelFilter(model)
				if err != nil {
					t.Fatal(err)
				}
				plan := models.AggregatePlan{APIVersion: models.AggregatePlanAPIVersion, MaxGroups: 3, Bucket: &models.AggregateBucket{Field: "stamp", Interval: "hour", Timezone: "UTC", From: "2026-10-14T00:30:00Z", To: "2026-10-14T02:30:00Z", MaxBuckets: 3, MaxPoints: 6}, Metrics: []models.AggregateMetric{{Key: "all", Operation: "count"}, {Key: "mean", Operation: "average", Field: "ratio", IncludeValidSampleCount: true}}}
				cell := func(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }
				rows := [][]any{{cell("1"), cell("2"), cell("0.0000"), cell("2")}}
				switch scenario {
				case "empty":
					rows = nil
				case "duplicate":
					rows = append(rows, rows[0])
				case "outside":
					rows[0][0] = cell("3")
				case "null-index":
					rows[0][0] = sql.NullString{}
				case "reversed":
					rows = append(rows, []any{cell("0"), cell("0"), sql.NullString{}, cell("0")})
				}
				executor := &aggregateExecutor{fakeExecutor: fakeExecutor{responses: []fakeQueryResponse{{match: func(string, []any) bool { return true }, rows: rows}}}}
				if scenario == "read-failure" {
					executor.readErr = cd.NewError(cd.Unexpected, "stream failed")
				}
				if scenario == "close-failure" {
					executor.closeErr = cd.NewError(cd.Unexpected, "close failed")
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if scenario == "canceled" {
					cancel()
				}
				handler := &impl{context: ctx, executor: executor, modelProvider: p, modelCodec: codec.New(p, "bucket")}
				result, err := Aggregate(handler, filter, plan)
				if scenario != "sparse" && scenario != "empty" {
					if err == nil || result != nil {
						t.Fatal("错误返回部分桶", result, err)
					}
					return
				}
				if err != nil || result == nil || len(result.Rows) != 3 {
					t.Fatal(result, err)
				}
				starts, _ := models.AggregateBucketStarts(plan.Bucket, len(plan.Metrics))
				for i, row := range result.Rows {
					if row.BucketStart != starts[i].Format(time.RFC3339Nano) || len(row.Groups) != 0 {
						t.Fatal(row)
					}
					if scenario == "sparse" && i == 1 {
						if row.Metrics[1].Value.Number != "0.0000" || row.Metrics[1].ValidSampleCount.Number != "2" {
							t.Fatal(row)
						}
					} else if row.Metrics[0].Value.Number != "0" || row.Metrics[1].Value.Kind != "null" || row.Metrics[1].ValidSampleCount.Number != "0" {
						t.Fatal("缺桶混淆零与空", row)
					}
				}
			})
		}
	}
}
