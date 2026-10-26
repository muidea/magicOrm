package orm

import (
	"database/sql"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database"
	"github.com/muidea/magicOrm/models"
)

// Aggregate 使用独立读取能力，旧 Orm 实现无需扩展稳定接口；不支持时明确失败。
func Aggregate(handler Orm, filter models.Filter, plan models.AggregatePlan) (*models.AggregateResult, *cd.Error) {
	capability, ok := handler.(interface {
		Aggregate(models.Filter, models.AggregatePlan) (*models.AggregateResult, *cd.Error)
	})
	if !ok || capability == nil {
		return nil, cd.NewError(cd.InvalidOperation, "ORM does not support aggregate queries")
	}
	return capability.Aggregate(filter, plan)
}

func (s *impl) Aggregate(filter models.Filter, plan models.AggregatePlan) (ret *models.AggregateResult, err *cd.Error) {
	if err = s.checkContext(); err != nil {
		return nil, err
	}
	if err = database.ValidateAggregateFilter(filter); err != nil {
		return nil, err
	}
	model := filter.MaskModel()
	runner := newBaseRunner(s.context, model, s.executor, s.modelProvider, s.modelCodec, false, 0)
	builder, ok := runner.sqlBuilder.(database.AggregateBuilder)
	if !ok {
		return nil, cd.NewError(cd.InvalidOperation, "builder does not support aggregate queries")
	}
	checked, ok := s.executor.(database.CheckedRowsExecutor)
	if !ok {
		return nil, cd.NewError(cd.InvalidOperation, "aggregate requires checked row iteration")
	}
	query, err := builder.BuildAggregate(model, filter, plan)
	if err != nil {
		return nil, err
	}
	if _, err = s.executor.Query(query.SQL(), false, query.Args()...); err != nil {
		return nil, err
	}
	defer func() {
		s.executor.Finish()
		if err == nil {
			err = checked.RowsError()
		}
		if err != nil {
			ret = nil
		}
	}()
	result := &models.AggregateResult{APIVersion: models.AggregatePlanAPIVersion, Rows: []models.AggregateRow{}}
	var bucketStarts []time.Time
	lastBucket := -1
	if plan.Bucket != nil {
		bucketStarts, err = models.AggregateBucketStarts(plan.Bucket, len(plan.Metrics))
		if err != nil {
			return nil, err
		}
	}
	for s.executor.Next() {
		if err = s.checkContext(); err != nil {
			return nil, err
		}
		if len(result.Rows) >= plan.MaxGroups {
			return nil, cd.NewError(cd.IllegalParam, "aggregate exceeds group budget")
		}
		columnCount := len(plan.GroupBy) + len(plan.Metrics)
		if plan.Bucket != nil {
			columnCount++
		}
		for _, metric := range plan.Metrics {
			if metric.IncludeValidSampleCount {
				columnCount++
			}
		}
		values := make([]sql.NullString, columnCount)
		destinations := make([]any, len(values))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err = s.executor.GetField(destinations...); err != nil {
			return nil, err
		}
		row := models.AggregateRow{Groups: []models.AggregateValue{}, Metrics: []models.AggregateValue{}}
		for i, name := range plan.GroupBy {
			groupField, groupErr := database.AggregateGroupField(model, name, s.modelProvider)
			if groupErr != nil {
				return nil, groupErr
			}
			value, valueErr := aggregateScalar(values[i], groupField.GetType().GetValue(), false)
			if valueErr != nil {
				return nil, valueErr
			}
			row.Groups = append(row.Groups, models.AggregateValue{Key: name, Value: value})
		}
		column := len(plan.GroupBy)
		if plan.Bucket != nil {
			index, indexErr := strconv.Atoi(values[0].String)
			if !values[0].Valid || indexErr != nil || strconv.Itoa(index) != values[0].String || index <= lastBucket || index >= len(bucketStarts) {
				return nil, cd.NewError(cd.Unexpected, "aggregate bucket identity or order is invalid")
			}
			for missing := lastBucket + 1; missing < index; missing++ {
				result.Rows = append(result.Rows, models.EmptyAggregateBucket(plan, bucketStarts[missing]))
			}
			lastBucket = index
			row.BucketStart = bucketStarts[index].Format(time.RFC3339Nano)
			column++
		}
		for _, metric := range plan.Metrics {
			kind := models.TypeBigIntegerValue
			if metric.Operation != "count" {
				kind = model.GetField(metric.Field).GetType().GetValue()
			}
			value, valueErr := aggregateScalar(values[column], kind, metric.Operation == "average")
			column++
			if valueErr != nil {
				return nil, valueErr
			}
			if metric.Operation == "count" && (value.Kind != "integer" || value.Number[0] == '-') {
				return nil, cd.NewError(cd.Unexpected, "aggregate count is invalid")
			}
			item := models.AggregateValue{Key: metric.Key, Value: value}
			if metric.IncludeValidSampleCount {
				count, countErr := aggregateScalar(values[column], models.TypeBigIntegerValue, false)
				column++
				if countErr != nil || count.Kind != "integer" || strings.HasPrefix(count.Number, "-") || (count.Number == "0") != (value.Kind == "null") {
					return nil, cd.NewError(cd.Unexpected, "aggregate average sample count is invalid")
				}
				item.ValidSampleCount = &count
			}
			row.Metrics = append(row.Metrics, item)
		}
		result.Rows = append(result.Rows, row)
	}
	if err = checked.RowsError(); err != nil {
		return nil, err
	}
	if err = s.checkContext(); err != nil {
		return nil, err
	}
	if plan.Bucket != nil {
		for missing := lastBucket + 1; missing < len(bucketStarts); missing++ {
			result.Rows = append(result.Rows, models.EmptyAggregateBucket(plan, bucketStarts[missing]))
		}
	} else if len(plan.GroupBy) == 0 && len(result.Rows) != 1 {
		return nil, cd.NewError(cd.Unexpected, "ungrouped aggregate must return exactly one row")
	}
	return result, nil
}

var aggregateNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

func aggregateScalar(raw sql.NullString, kind models.TypeDeclare, average bool) (models.QueryScalar, *cd.Error) {
	invalid := func() (models.QueryScalar, *cd.Error) {
		return models.QueryScalar{}, cd.NewError(cd.Unexpected, "aggregate returned an invalid scalar")
	}
	if !raw.Valid {
		return models.QueryScalar{Kind: "null"}, nil
	}
	if len(raw.String) > 4096 {
		return invalid()
	}
	if !average && kind == models.TypeStringValue {
		return models.QueryScalar{Kind: "string", Text: raw.String}, nil
	}
	if !average && kind == models.TypeBooleanValue {
		value, err := strconv.ParseBool(raw.String)
		if err != nil {
			return invalid()
		}
		return models.QueryScalar{Kind: "boolean", Boolean: value}, nil
	}
	if !aggregateNumber.MatchString(raw.String) {
		return invalid()
	}
	if index := strings.IndexAny(raw.String, "eE"); index >= 0 {
		exponent, err := strconv.ParseInt(raw.String[index+1:], 10, 32)
		if err != nil || exponent < -1024 || exponent > 1024 {
			return invalid()
		}
	}
	if !average && kind >= models.TypeByteValue && kind <= models.TypePositiveBigIntegerValue {
		value, ok := new(big.Rat).SetString(raw.String)
		if !ok || !value.IsInt() || len(value.Num().String()) > 4096 {
			return invalid()
		}
		return models.QueryScalar{Kind: "integer", Number: value.Num().String()}, nil
	}
	return models.QueryScalar{Kind: "decimal", Number: raw.String}, nil
}
