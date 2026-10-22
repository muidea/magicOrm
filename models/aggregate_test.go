package models

import "testing"

func TestAggregatePlanBudgets(t *testing.T) {
	for _, tc := range []struct {
		name  string
		plan  AggregatePlan
		valid bool
	}{
		{"average-samples", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1, Metrics: []AggregateMetric{{Key: "mean", Operation: "average", Field: "rank", IncludeValidSampleCount: true}}}, true},
		{"count-samples", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1, Metrics: []AggregateMetric{{Key: "count", Operation: "count", IncludeValidSampleCount: true}}}, false},
		{"invalid-condition", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1, Metrics: []AggregateMetric{{Key: "count", Operation: "count", Where: &QueryExpression{Kind: "script"}}}}, false},
		{"count", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1, Metrics: []AggregateMetric{{Key: "count", Operation: "count"}}}, true},
		{"future", AggregatePlan{APIVersion: "v2", MaxGroups: 1, Metrics: []AggregateMetric{{Key: "count", Operation: "count"}}}, false},
		{"no-metrics", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1}, false},
		{"over-groups", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1001, Metrics: []AggregateMetric{{Key: "count", Operation: "count"}}}, false},
		{"sum-star", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1, Metrics: []AggregateMetric{{Key: "sum", Operation: "sum"}}}, false},
		{"script", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1, Metrics: []AggregateMetric{{Key: "script", Operation: "script"}}}, false},
		{"duplicate-group", AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 1, GroupBy: []string{"name", "name"}, Metrics: []AggregateMetric{{Key: "count", Operation: "count"}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ValidateAggregatePlan(tc.plan); (err == nil) != tc.valid {
				t.Fatal(err)
			}
		})
	}
}
