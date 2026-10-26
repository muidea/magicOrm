package database

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/internal/database/codec"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider"
)

var readQueryIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var readQueryInteger = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
var readQueryDecimal = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// BuildReadQueryClause 为两种生成器共享同一有界读合同；所有值只进入参数栈。
// 该函数不执行 SQL，不取得连接，也不识别应用、租户或授权策略。
func BuildReadQueryClause(model models.Model, filter models.Filter, p provider.Provider, dialect string, offset int) (where, order string, args []any, err *cd.Error) {
	return buildReadQueryClause(model, filter, p, dialect, offset, nil, 0)
}

// BuildReadQueryClauseWithCodec adds model-resolved reference semi-joins.
func BuildReadQueryClauseWithCodec(model models.Model, filter models.Filter, p provider.Provider, dialect string, offset int, c codec.Codec) (where, order string, args []any, err *cd.Error) {
	return buildReadQueryClause(model, filter, p, dialect, offset, c, 0)
}
func buildReadQueryClause(model models.Model, filter models.Filter, p provider.Provider, dialect string, offset int, c codec.Codec, depth int) (where, order string, args []any, err *cd.Error) {
	source, ok := filter.(models.ReadQueryPlanProvider)
	if !ok {
		return
	}
	invalid := func() *cd.Error {
		return cd.NewError(cd.IllegalParam, "read query plan does not match scalar model or stable order")
	}
	plan := source.ReadQueryPlan()
	if err = models.ValidateReadQueryPlan(plan); err != nil {
		return "", "", nil, err
	}
	if model == nil || p == nil || offset < 0 || dialect != "postgres" && dialect != "mysql" {
		return "", "", nil, invalid()
	}
	// local Filter 的历史名称为空，使用其模型身份复核；不改变旧 Filter 接口。
	identity := filter.MaskModel()
	if identity == nil || identity.GetPkgKey() != model.GetPkgKey() {
		return "", "", nil, invalid()
	}
	quote := func(name string) string { return `"` + name + `"` }
	if dialect == "mysql" {
		quote = func(name string) string { return "`" + name + "`" }
	}
	field := func(name string) models.Field {
		if !readQueryIdentifier.MatchString(name) {
			return nil
		}
		f := model.GetField(name)
		if f == nil || f.GetType() == nil || f.GetSpec() == nil || !models.IsBasicField(f) || models.IsSliceField(f) {
			return nil
		}
		if constraints := f.GetSpec().GetConstraints(); constraints != nil && constraints.Has(models.KeyWriteOnly) {
			return nil
		}
		return f
	}
	var compile func(models.QueryExpression) (string, *cd.Error)
	compile = func(e models.QueryExpression) (string, *cd.Error) {
		if e.Kind == "and" || e.Kind == "or" {
			parts := make([]string, len(e.Children))
			for i, child := range e.Children {
				value, err := compile(child)
				if err != nil {
					return "", err
				}
				parts[i] = value
			}
			return "(" + strings.Join(parts, " "+strings.ToUpper(e.Kind)+" ") + ")", nil
		}
		f := field(e.Field)
		if f == nil {
			return "", invalid()
		}
		name := quote(e.Field)
		if e.Operator == "is-null" || e.Operator == "not-null" {
			if !f.GetType().IsPtrType() {
				return "", invalid()
			}
			if e.Operator == "is-null" {
				return name + " IS NULL", nil
			}
			return name + " IS NOT NULL", nil
		}
		value, valueErr := readQueryScalarValue(*e.Value, f.GetType())
		if valueErr != nil {
			return "", valueErr
		}
		if e.Value.Kind == "boolean" && e.Operator != "eq" && e.Operator != "ne" {
			return "", invalid()
		}
		encoded, encodeErr := p.EncodeValue(value, f.GetType())
		if encodeErr != nil {
			return "", encodeErr
		}
		args = append(args, encoded)
		placeholder := "?"
		if dialect == "postgres" {
			placeholder = fmt.Sprintf("$%d", offset+len(args))
		}
		op := map[string]string{"eq": "=", "ne": "<>", "lt": "<", "lte": "<=", "gt": ">", "gte": ">="}[e.Operator]
		return name + " " + op + " " + placeholder, nil
	}
	if plan.Where != nil {
		where, err = compile(*plan.Where)
		if err != nil {
			return "", "", nil, err
		}
	}
	if len(plan.References) > 0 {
		if c == nil || depth >= 3 {
			return "", "", nil, invalid()
		}
		clauses, values, referenceErr := buildReadReferences(model, plan.References, p, dialect, offset+len(args), c, depth)
		if referenceErr != nil {
			return "", "", nil, referenceErr
		}
		args = append(args, values...)
		if where != "" {
			where = "(" + where + ") AND (" + clauses + ")"
		} else {
			where = clauses
		}
	}
	if len(plan.OrderBy) != 0 {
		primary := model.GetPrimaryField()
		if filter.Sorter() != nil || primary == nil || plan.OrderBy[len(plan.OrderBy)-1].Field != primary.GetName() {
			return "", "", nil, invalid()
		}
		parts := make([]string, len(plan.OrderBy))
		for i, item := range plan.OrderBy {
			f := field(item.Field)
			if f == nil || f.GetType().IsPtrType() || f.GetType().GetValue() == models.TypeBooleanValue {
				return "", "", nil, invalid()
			}
			parts[i] = quote(item.Field) + " " + strings.ToUpper(item.Direction)
		}
		order = strings.Join(parts, ", ")
	}
	return
}

// ValidateReadQueryPage 仅限制新读计划的行页；旧 Filter 和 Count 的行为不变。
// 1001 行允许上层以 1000 行页面加一行探测后续页；深分页应改用 keyset。
func ValidateReadQueryPage(filter models.Filter) *cd.Error {
	if _, ok := filter.(models.ReadQueryPlanProvider); !ok {
		return nil
	}
	page := filter.Paginationer()
	if page == nil {
		return cd.NewError(cd.IllegalParam, "read query plan requires bounded pagination")
	}
	limit, offset := page.Limit(), page.Offset()
	if limit < 1 || limit > 1001 || offset < 0 || offset > 10000 {
		return cd.NewError(cd.IllegalParam, "read query page exceeds row or offset budget")
	}
	return nil
}

func readQueryScalarValue(value models.QueryScalar, typ models.Type) (any, *cd.Error) {
	invalid := func() (any, *cd.Error) {
		return nil, cd.NewError(cd.IllegalParam, "query scalar kind, format or range differs from model")
	}
	kind := typ.GetValue()
	switch kind {
	case models.TypeStringValue:
		if value.Kind != "string" {
			return invalid()
		}
		return value.Text, nil
	case models.TypeBooleanValue:
		if value.Kind != "boolean" {
			return invalid()
		}
		return value.Boolean, nil
	case models.TypeDateTimeValue:
		if value.Kind != "datetime" {
			return invalid()
		}
		t, err := time.Parse(time.RFC3339Nano, value.Text)
		if err != nil || !t.Equal(t.Truncate(time.Millisecond)) {
			return invalid()
		}
		return t.UTC(), nil
	case models.TypeFloatValue, models.TypeDoubleValue:
		if value.Kind != "decimal" || !readQueryDecimal.MatchString(value.Number) {
			return invalid()
		}
		bits := 64
		if kind == models.TypeFloatValue {
			bits = 32
		}
		n, err := strconv.ParseFloat(value.Number, bits)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return invalid()
		}
		return n, nil
	}
	if value.Kind != "integer" || !readQueryInteger.MatchString(value.Number) {
		return invalid()
	}
	bits := 64
	unsigned := false
	switch kind {
	case models.TypeByteValue:
		bits = 8
	case models.TypeSmallIntegerValue:
		bits = 16
	case models.TypeInteger32Value:
		bits = 32
	case models.TypeIntegerValue:
		bits = strconv.IntSize
	case models.TypeBigIntegerValue:
	case models.TypePositiveByteValue:
		bits, unsigned = 8, true
	case models.TypePositiveSmallIntegerValue:
		bits, unsigned = 16, true
	case models.TypePositiveInteger32Value:
		bits, unsigned = 32, true
	case models.TypePositiveIntegerValue:
		bits, unsigned = strconv.IntSize, true
	case models.TypePositiveBigIntegerValue:
		unsigned = true
	default:
		return invalid()
	}
	if unsigned {
		n, err := strconv.ParseUint(value.Number, 10, bits)
		if err != nil {
			return invalid()
		}
		return n, nil
	}
	n, err := strconv.ParseInt(value.Number, 10, bits)
	if err != nil {
		return invalid()
	}
	return n, nil
}
