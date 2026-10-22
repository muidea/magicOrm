package models

import (
	"time"

	cd "github.com/muidea/magicCommon/def"
)

// AggregateBucket 是 UTC 日历小时桶；From/To 限制实际采样区间，边缘桶可以不完整。
type AggregateBucket struct {
	Field      string `json:"field"`
	Interval   string `json:"interval"`
	Timezone   string `json:"timezone"`
	From       string `json:"from"`
	To         string `json:"to"`
	MaxBuckets int    `json:"maxBuckets"`
	MaxPoints  int    `json:"maxPoints"`
}

// AggregateBucketStarts 生成唯一 UTC 桶身份，不使用数据库会话时区或客户端标签分组。
func AggregateBucketStarts(bucket *AggregateBucket, metrics int) ([]time.Time, *cd.Error) {
	invalid := func() ([]time.Time, *cd.Error) {
		return nil, cd.NewError(cd.IllegalParam, "invalid or over-budget UTC hourly aggregate bucket")
	}
	if bucket == nil || bucket.Field == "" || len(bucket.Field) > 128 || bucket.Interval != "hour" || bucket.Timezone != "UTC" || bucket.MaxBuckets < 1 || bucket.MaxBuckets > 744 || bucket.MaxPoints < 1 || bucket.MaxPoints > 5000 || metrics < 1 || metrics > 16 {
		return invalid()
	}
	from, err := time.Parse(time.RFC3339Nano, bucket.From)
	if err != nil {
		return invalid()
	}
	to, err := time.Parse(time.RFC3339Nano, bucket.To)
	if err != nil || !from.Before(to) || to.Sub(from) > 744*time.Hour || from.Year() < 1 || to.Year() > 9999 || from.UTC().Format(time.RFC3339Nano) != bucket.From || to.UTC().Format(time.RFC3339Nano) != bucket.To || !from.Equal(from.Truncate(time.Millisecond)) || !to.Equal(to.Truncate(time.Millisecond)) {
		return invalid()
	}
	starts := []time.Time{}
	for start := from.Truncate(time.Hour); start.Before(to); start = start.Add(time.Hour) {
		if len(starts) >= bucket.MaxBuckets || (len(starts)+1)*metrics > bucket.MaxPoints {
			return invalid()
		}
		starts = append(starts, start)
	}
	return starts, nil
}

// EmptyAggregateBucket 只填补无数据库分组的桶；count=0，其余指标为 null。
func EmptyAggregateBucket(plan AggregatePlan, start time.Time) AggregateRow {
	row := AggregateRow{BucketStart: start.UTC().Format(time.RFC3339Nano), Groups: []AggregateValue{}, Metrics: []AggregateValue{}}
	for _, metric := range plan.Metrics {
		value := AggregateValue{Key: metric.Key, Value: QueryScalar{Kind: "null"}}
		if metric.Operation == "count" {
			value.Value = QueryScalar{Kind: "integer", Number: "0"}
		}
		if metric.IncludeValidSampleCount {
			count := QueryScalar{Kind: "integer", Number: "0"}
			value.ValidSampleCount = &count
		}
		row.Metrics = append(row.Metrics, value)
	}
	return row
}
