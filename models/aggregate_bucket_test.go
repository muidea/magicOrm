package models

import "testing"

func TestAggregateBucketBudgets(t *testing.T) {
	for _, scenario := range []string{"partial", "aligned", "timezone", "precision", "backwards", "buckets", "points", "long", "noncanonical", "group"} {
		t.Run(scenario, func(t *testing.T) {
			bucket := &AggregateBucket{Field: "stamp", Interval: "hour", Timezone: "UTC", From: "2026-10-14T00:30:00Z", To: "2026-10-14T02:30:00Z", MaxBuckets: 3, MaxPoints: 3}
			plan := AggregatePlan{APIVersion: AggregatePlanAPIVersion, MaxGroups: 3, Bucket: bucket, Metrics: []AggregateMetric{{Key: "all", Operation: "count"}}}
			switch scenario {
			case "aligned":
				bucket.From = "2026-10-14T00:00:00Z"
				bucket.To = "2026-10-14T03:00:00Z"
			case "timezone":
				bucket.Timezone = "Asia/Shanghai"
			case "precision":
				bucket.From = "2026-10-14T00:30:00.000001Z"
			case "backwards":
				bucket.To = bucket.From
			case "buckets":
				bucket.MaxBuckets = 2
			case "points":
				bucket.MaxPoints = 2
			case "long":
				bucket.To = "2026-11-15T00:00:00Z"
			case "noncanonical":
				bucket.From = "2026-10-14T00:30:00+00:00"
			case "group":
				plan.GroupBy = []string{"id"}
			}
			if err := ValidateAggregatePlan(plan); (err == nil) != (scenario == "partial" || scenario == "aligned") {
				t.Fatal(scenario, err)
			}
			if scenario == "partial" || scenario == "aligned" {
				starts, err := AggregateBucketStarts(bucket, 1)
				if err != nil || len(starts) != 3 || starts[0].Hour() != 0 || starts[2].Hour() != 2 {
					t.Fatal(starts, err)
				}
			}
		})
	}
}
