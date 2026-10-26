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
)

func TestAggregateBucketBothProvidersAndBuilders(t *testing.T) {
	for _, remoteModel := range []bool{false, true} {
		p, model := readQueryModel(t, remoteModel)
		for dialect, legacy := range map[string]database.Builder{"postgres": postgres.NewBuilder(p, codec.New(p, "bucket")), "mysql": mysql.NewBuilder(p, codec.New(p, "bucket"))} {
			t.Run(dialect+map[bool]string{false: "/local", true: "/remote"}[remoteModel], func(t *testing.T) {
				base, err := p.GetModelFilter(model)
				if err != nil {
					t.Fatal(err)
				}
				if err := base.Equal("active", false); err != nil {
					t.Fatal(err)
				}
				where := queryCompare("rank", "gte", "integer", "0")
				filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: &where})
				if err != nil {
					t.Fatal(err)
				}
				condition := queryCompare("rank", "gte", "integer", "7")
				plan := models.AggregatePlan{APIVersion: models.AggregatePlanAPIVersion, MaxGroups: 3, Bucket: &models.AggregateBucket{Field: "stamp", Interval: "hour", Timezone: "UTC", From: "2026-10-14T00:30:00Z", To: "2026-10-14T02:30:00Z", MaxBuckets: 3, MaxPoints: 6}, Metrics: []models.AggregateMetric{{Key: "all", Operation: "count"}, {Key: "mean", Operation: "average", Field: "ratio", Where: &condition, IncludeValidSampleCount: true}}}
				result, err := legacy.(database.AggregateBuilder).BuildAggregate(model, filter, plan)
				if err != nil {
					t.Fatal(err)
				}
				encode := func(text string) any {
					instant, parseErr := time.Parse(time.RFC3339Nano, text)
					if parseErr != nil {
						t.Fatal(parseErr)
					}
					value, encodeErr := p.EncodeValue(instant, model.GetField("stamp").GetType())
					if encodeErr != nil {
						t.Fatal(encodeErr)
					}
					return value
				}
				want := []any{encode("2026-10-14T01:00:00Z"), encode("2026-10-14T02:00:00Z"), encode("2026-10-14T02:30:00Z"), int64(7), int64(7), false, int64(0), encode(plan.Bucket.From), encode(plan.Bucket.To), 4}
				if !reflect.DeepEqual(result.Args(), want) || !strings.Contains(result.SQL(), "GROUP BY 1 ORDER BY 1") || strings.Count(result.SQL(), "CASE") != 3 || strings.Contains(result.SQL(), "2026-") {
					t.Fatal("桶边界、指标参数或原 scope 丢失", result.SQL(), result.Args(), want)
				}
				if dialect == "postgres" && !strings.Contains(result.SQL(), "LIMIT $10") {
					t.Fatal(result.SQL())
				}
				for _, field := range []string{"rank", "secret", "tags", "missing", "stamp)"} {
					plan.Bucket.Field = field
					if _, err := legacy.(database.AggregateBuilder).BuildAggregate(model, filter, plan); err == nil {
						t.Fatal("接受非法分桶字段", field)
					}
				}
			})
		}
	}
}
