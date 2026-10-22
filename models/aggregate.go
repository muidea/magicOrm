package models

import cd "github.com/muidea/magicCommon/def"

const AggregatePlanAPIVersion = "magic.orm-aggregate/v1"

type AggregateMetric struct {
	Key                     string           `json:"key"`
	Operation               string           `json:"operation"`
	Field                   string           `json:"field,omitempty"`
	Where                   *QueryExpression `json:"where,omitempty"`
	IncludeValidSampleCount bool             `json:"includeValidSampleCount,omitempty"`
}

// AggregatePlan 只计算服务端指标；不承载授权、脚本或任意 SQL。
// GroupBy 的结果超过 MaxGroups 时整体失败，不能将截断结果当作完整统计。
type AggregatePlan struct {
	APIVersion string            `json:"apiVersion"`
	Metrics    []AggregateMetric `json:"metrics"`
	GroupBy    []string          `json:"groupBy"`
	MaxGroups  int               `json:"maxGroups"`
	Bucket     *AggregateBucket  `json:"bucket,omitempty"`
}

type AggregateValue struct {
	Key              string       `json:"key"`
	Value            QueryScalar  `json:"value"`
	ValidSampleCount *QueryScalar `json:"validSampleCount,omitempty"`
}

type AggregateRow struct {
	BucketStart string           `json:"bucketStart,omitempty"`
	Groups      []AggregateValue `json:"groups"`
	Metrics     []AggregateValue `json:"metrics"`
}

type AggregateResult struct {
	APIVersion string         `json:"apiVersion"`
	Rows       []AggregateRow `json:"rows"`
}

func ValidateAggregatePlan(plan AggregatePlan) *cd.Error {
	invalid := func() *cd.Error { return cd.NewError(cd.IllegalParam, "invalid or over-budget aggregate plan") }
	if plan.APIVersion != AggregatePlanAPIVersion || len(plan.Metrics) < 1 || len(plan.Metrics) > 16 || len(plan.GroupBy) > 4 || plan.MaxGroups < 1 || plan.MaxGroups > 1000 {
		return invalid()
	}
	if plan.Bucket != nil {
		starts, err := AggregateBucketStarts(plan.Bucket, len(plan.Metrics))
		if err != nil || len(plan.GroupBy) != 0 || len(starts) > plan.MaxGroups {
			return invalid()
		}
	}
	seen := map[string]bool{}
	for _, item := range plan.Metrics {
		if item.Key == "" || len(item.Key) > 128 || seen[item.Key] {
			return invalid()
		}
		seen[item.Key] = true
		switch item.Operation {
		case "count":
		case "sum", "average", "min", "max":
			if item.Field == "" {
				return invalid()
			}
		default:
			return invalid()
		}
		if len(item.Field) > 128 || item.IncludeValidSampleCount && item.Operation != "average" {
			return invalid()
		}
		if err := ValidateReadQueryPlan(ReadQueryPlan{APIVersion: ReadQueryPlanAPIVersion, Where: item.Where}); err != nil {
			return err
		}
	}
	seen = map[string]bool{}
	for _, field := range plan.GroupBy {
		if field == "" || len(field) > 128 || seen[field] {
			return invalid()
		}
		seen[field] = true
	}
	return nil
}
