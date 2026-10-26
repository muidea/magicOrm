package database

import (
	"fmt"
	"strings"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

// BuildAggregateBucketClause 复用模型时间编码；范围条件独立追加，不能依赖调用者记得提供时间窗。
func BuildAggregateBucketClause(model models.Model, bucket *models.AggregateBucket, p provider.Provider, dialect string, offset int) (string, []any, *cd.Error) {
	if bucket == nil {
		return "", nil, nil
	}
	where := &models.QueryExpression{Kind: "and", Children: []models.QueryExpression{
		{Kind: "compare", Field: bucket.Field, Operator: "gte", Value: &models.QueryScalar{Kind: "datetime", Text: bucket.From}},
		{Kind: "compare", Field: bucket.Field, Operator: "lt", Value: &models.QueryScalar{Kind: "datetime", Text: bucket.To}},
	}}
	return buildBucketCondition(model, where, p, dialect, offset)
}

func buildBucketCondition(model models.Model, where *models.QueryExpression, p provider.Provider, dialect string, offset int) (string, []any, *cd.Error) {
	if p == nil {
		return "", nil, cd.NewError(cd.IllegalParam, "bucket requires a parameterized provider")
	}
	base, err := p.GetModelFilter(model)
	if err != nil {
		return "", nil, err
	}
	filter, err := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: where})
	if err != nil {
		return "", nil, err
	}
	clause, _, args, err := BuildReadQueryClause(model, filter, p, dialect, offset)
	return clause, args, err
}

func buildBucketProjection(model models.Model, plan models.AggregatePlan, p provider.Provider, dialect string, offset int) (string, []any, *cd.Error) {
	field, err := AggregateField(model, plan.Bucket.Field)
	if err != nil {
		return "", nil, err
	}
	if field.GetType().GetValue() != models.TypeDateTimeValue {
		return "", nil, cd.NewError(cd.IllegalParam, "bucket requires a datetime field")
	}
	starts, err := models.AggregateBucketStarts(plan.Bucket, len(plan.Metrics))
	if err != nil {
		return "", nil, err
	}
	parts := []string{"CASE"}
	args := []any{}
	to, _ := time.Parse(time.RFC3339Nano, plan.Bucket.To)
	for i, start := range starts {
		end := start.Add(time.Hour)
		if end.After(to) {
			end = to
		}
		where := &models.QueryExpression{Kind: "compare", Field: plan.Bucket.Field, Operator: "lt", Value: &models.QueryScalar{Kind: "datetime", Text: end.Format(time.RFC3339Nano)}}
		clause, values, err := buildBucketCondition(model, where, p, dialect, offset+len(args))
		if err != nil {
			return "", nil, err
		}
		args = append(args, values...)
		parts = append(parts, fmt.Sprintf("WHEN %s THEN %d", clause, i))
	}
	return strings.Join(parts, " ") + " ELSE NULL END", args, nil
}
