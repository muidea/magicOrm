package database

import (
	"fmt"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

// AggregateBuilder 是独立读取能力，不改变旧 Builder 或业务写合同。
type AggregateBuilder interface {
	BuildAggregate(models.Model, models.Filter, models.AggregatePlan) (Result, *cd.Error)
}

// CheckedRowsExecutor 防止将中途断流误认为完整统计；未配套执行器明确拒绝。
type CheckedRowsExecutor interface {
	RowsError() *cd.Error
}

// AggregateProjection 只从模型字段和封闭操作构造投影；键仅用于结果，不拼入 SQL。
func AggregateProjection(model models.Model, plan models.AggregatePlan, dialect string) (projection, groups string, err *cd.Error) {
	if plan.Bucket != nil {
		return "", "", cd.NewError(cd.IllegalParam, "bucket aggregate requires parameterized projection")
	}
	for _, metric := range plan.Metrics {
		if metric.Where != nil {
			return "", "", cd.NewError(cd.IllegalParam, "conditional aggregate requires parameterized projection")
		}
	}
	projection, groups, _, err = BuildAggregateProjection(model, plan, dialect, nil, 0)
	return
}

// BuildAggregateProjection 在 SELECT 中参数化指标条件；外层 scope 仍由原 Filter 限制。
func BuildAggregateProjection(model models.Model, plan models.AggregatePlan, dialect string, p provider.Provider, offset int) (projection, groups string, args []any, err *cd.Error) {
	if err = models.ValidateAggregatePlan(plan); err != nil {
		return
	}
	if model == nil || offset < 0 || dialect != "postgres" && dialect != "mysql" {
		return "", "", nil, cd.NewError(cd.IllegalParam, "aggregate model or dialect is invalid")
	}
	quote := func(name string) string { return `"` + name + `"` }
	if dialect == "mysql" {
		quote = func(name string) string { return "`" + name + "`" }
	}
	var columns, groupColumns []string
	if plan.Bucket != nil {
		bucketColumn, values, bucketErr := buildBucketProjection(model, plan, p, dialect, offset)
		if bucketErr != nil {
			return "", "", nil, bucketErr
		}
		columns = append(columns, bucketColumn)
		args = append(args, values...)
		// 两数据库均按首列分组，避免重复 CASE 参数或依赖会话时区函数。
		groupColumns = append(groupColumns, "1")
	}
	for _, name := range plan.GroupBy {
		field, fieldErr := AggregateField(model, name)
		if fieldErr != nil {
			return "", "", nil, fieldErr
		}
		if field.GetType().GetValue() == models.TypeDateTimeValue {
			return "", "", nil, cd.NewError(cd.IllegalParam, "datetime grouping requires an explicit bucket contract")
		}
		groupColumns = append(groupColumns, quote(name))
		columns = append(columns, quote(name))
	}
	for _, metric := range plan.Metrics {
		name := "*"
		if metric.Field != "" {
			field, fieldErr := AggregateField(model, metric.Field)
			if fieldErr != nil {
				return "", "", nil, fieldErr
			}
			kind := field.GetType().GetValue()
			if metric.Operation != "count" {
				numeric := kind >= models.TypeByteValue && kind <= models.TypeDoubleValue
				if !numeric && !(kind == models.TypeStringValue && (metric.Operation == "min" || metric.Operation == "max")) {
					return "", "", nil, cd.NewError(cd.IllegalParam, "aggregate operation requires a compatible numeric or string field")
				}
			}
			name = quote(metric.Field)
		}
		condition := func() (string, *cd.Error) {
			if metric.Where == nil {
				return "", nil
			}
			if p == nil {
				return "", cd.NewError(cd.IllegalParam, "conditional aggregate requires provider")
			}
			base, filterErr := p.GetModelFilter(model)
			if filterErr != nil {
				return "", filterErr
			}
			filter, filterErr := models.WithReadQueryPlan(base, models.ReadQueryPlan{APIVersion: models.ReadQueryPlanAPIVersion, Where: metric.Where})
			if filterErr != nil {
				return "", filterErr
			}
			where, _, values, filterErr := BuildReadQueryClause(model, filter, p, dialect, offset+len(args))
			if filterErr != nil {
				return "", filterErr
			}
			args = append(args, values...)
			return where, nil
		}
		operand := func() (string, *cd.Error) {
			where, whereErr := condition()
			if whereErr != nil || where == "" {
				return name, whereErr
			}
			value := name
			if value == "*" {
				value = "1"
			}
			return fmt.Sprintf("CASE WHEN %s THEN %s ELSE NULL END", where, value), nil
		}
		value, valueErr := operand()
		if valueErr != nil {
			return "", "", nil, valueErr
		}
		op := strings.ToUpper(metric.Operation)
		if metric.Operation == "average" {
			op = "AVG"
		}
		columns = append(columns, op+"("+value+")")
		if metric.IncludeValidSampleCount {
			value, valueErr := operand()
			if valueErr != nil {
				return "", "", nil, valueErr
			}
			columns = append(columns, "COUNT("+value+")")
		}
	}
	return strings.Join(columns, ", "), strings.Join(groupColumns, ", "), args, nil
}

// ValidateAggregateFilter 拒绝行页/排序，避免把当前页误当完整统计。
func ValidateAggregateFilter(filter models.Filter) *cd.Error {
	if filter == nil || filter.MaskModel() == nil {
		return cd.NewError(cd.IllegalParam, "aggregate requires a model filter")
	}
	if filter.Paginationer() != nil || filter.Sorter() != nil {
		return cd.NewError(cd.IllegalParam, "aggregate cannot use row pagination or sorting")
	}
	if planned, ok := filter.(models.ReadQueryPlanProvider); ok && len(planned.ReadQueryPlan().OrderBy) != 0 {
		return cd.NewError(cd.IllegalParam, "aggregate cannot use row ordering or keyset pagination")
	}
	return nil
}

func AggregateField(model models.Model, name string) (models.Field, *cd.Error) {
	if model == nil || !readQueryIdentifier.MatchString(name) {
		return nil, cd.NewError(cd.IllegalParam, "aggregate field is invalid")
	}
	field := model.GetField(name)
	if field == nil || field.GetType() == nil || field.GetSpec() == nil || !models.IsBasicField(field) || models.IsSliceField(field) {
		return nil, cd.NewError(cd.IllegalParam, "aggregate field must be a model scalar")
	}
	if constraints := field.GetSpec().GetConstraints(); constraints != nil && constraints.Has(models.KeyWriteOnly) {
		return nil, cd.NewError(cd.IllegalParam, "aggregate cannot read a write-only field")
	}
	return field, nil
}
