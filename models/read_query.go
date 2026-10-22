package models

import cd "github.com/muidea/magicCommon/def"

const ReadQueryPlanAPIVersion = "magic.orm-read-query/v1"

// QueryScalar 是封闭标量；数值以十进制 token 表达，避免协议中转丢失整数位值。
// null 仅用于显式空值判断，不与缺省值或零值混用。
type QueryScalar struct {
	Kind    string `json:"kind"`
	Text    string `json:"text,omitempty"`
	Number  string `json:"number,omitempty"`
	Boolean bool   `json:"boolean,omitempty"`
}

type QueryExpression struct {
	Kind     string            `json:"kind"`
	Field    string            `json:"field,omitempty"`
	Operator string            `json:"operator,omitempty"`
	Value    *QueryScalar      `json:"value,omitempty"`
	Children []QueryExpression `json:"children,omitempty"`
}

type QueryOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// ReadQueryPlan 只追加读取条件；旧 Filter 的条件仍以 AND 组合，不能被此计划覆盖。
// 当前只提供标量条件、组合排序和 keyset 表达式，不宣称聚合、关系或游标授权已配套。
type ReadQueryPlan struct {
	APIVersion string           `json:"apiVersion"`
	Where      *QueryExpression `json:"where,omitempty"`
	OrderBy    []QueryOrder     `json:"orderBy,omitempty"`
}

type ReadQueryPlanProvider interface {
	ReadQueryPlan() ReadQueryPlan
}

type readQueryFilter struct {
	Filter
	plan ReadQueryPlan
}

// WithReadQueryPlan 隔离调用方材料；local/remote provider 共用同一读取扩展。
// 不改变 Filter 接口及其历史单字段覆盖行为；写入生成器必须显式拒绝本扩展。
func WithReadQueryPlan(base Filter, plan ReadQueryPlan) (Filter, *cd.Error) {
	if base == nil {
		return nil, cd.NewError(cd.IllegalParam, "read query filter is nil")
	}
	if _, exists := base.(ReadQueryPlanProvider); exists {
		return nil, cd.NewError(cd.IllegalParam, "read query plans cannot be nested")
	}
	if err := ValidateReadQueryPlan(plan); err != nil {
		return nil, err
	}
	return &readQueryFilter{Filter: base, plan: cloneReadQueryPlan(plan)}, nil
}

func (f *readQueryFilter) ReadQueryPlan() ReadQueryPlan { return cloneReadQueryPlan(f.plan) }

func (f *readQueryFilter) HasValueMask() bool {
	if source, ok := f.Filter.(interface{ HasValueMask() bool }); ok {
		return source.HasValueMask()
	}
	return false
}

func (f *readQueryFilter) ResponseModel() Model {
	if source, ok := f.Filter.(interface{ ResponseModel() Model }); ok {
		return source.ResponseModel()
	}
	return f.Filter.MaskModel()
}

func (f *readQueryFilter) ExplicitResponseModel() Model {
	if source, ok := f.Filter.(interface{ ExplicitResponseModel() Model }); ok {
		return source.ExplicitResponseModel()
	}
	return f.Filter.MaskModel()
}

func cloneReadQueryPlan(plan ReadQueryPlan) ReadQueryPlan {
	plan.OrderBy = append([]QueryOrder(nil), plan.OrderBy...)
	var clone func(QueryExpression) QueryExpression
	clone = func(e QueryExpression) QueryExpression {
		if e.Value != nil {
			value := *e.Value
			e.Value = &value
		}
		if e.Children != nil {
			children := make([]QueryExpression, len(e.Children))
			for i := range e.Children {
				children[i] = clone(e.Children[i])
			}
			e.Children = children
		}
		return e
	}
	if plan.Where != nil {
		e := clone(*plan.Where)
		plan.Where = &e
	}
	return plan
}

func ValidateReadQueryPlan(plan ReadQueryPlan) *cd.Error {
	invalid := func() *cd.Error { return cd.NewError(cd.IllegalParam, "invalid or over-budget read query plan") }
	if plan.APIVersion != ReadQueryPlanAPIVersion || len(plan.OrderBy) > 8 {
		return invalid()
	}
	orders := map[string]bool{}
	for _, order := range plan.OrderBy {
		if order.Field == "" || len(order.Field) > 128 || orders[order.Field] || order.Direction != "asc" && order.Direction != "desc" {
			return invalid()
		}
		orders[order.Field] = true
	}
	nodes := 0
	var check func(QueryExpression, int) bool
	check = func(e QueryExpression, depth int) bool {
		nodes++
		if depth > 8 || nodes > 64 {
			return false
		}
		if e.Kind == "and" || e.Kind == "or" {
			if e.Field != "" || e.Operator != "" || e.Value != nil || len(e.Children) < 2 || len(e.Children) > 16 {
				return false
			}
			for _, child := range e.Children {
				if !check(child, depth+1) {
					return false
				}
			}
			return true
		}
		if e.Kind != "compare" || e.Field == "" || len(e.Field) > 128 || len(e.Children) != 0 || e.Value == nil {
			return false
		}
		switch e.Operator {
		case "eq", "ne", "lt", "lte", "gt", "gte":
			if e.Value.Kind == "null" {
				return false
			}
		case "is-null", "not-null":
			if e.Value.Kind != "null" {
				return false
			}
		default:
			return false
		}
		v := e.Value
		if len(v.Text)+len(v.Number) > 4096 {
			return false
		}
		switch v.Kind {
		case "null":
			return v.Text == "" && v.Number == "" && !v.Boolean
		case "string", "datetime":
			return v.Number == "" && !v.Boolean
		case "integer", "decimal":
			return v.Text == "" && v.Number != "" && !v.Boolean
		case "boolean":
			return v.Text == "" && v.Number == ""
		}
		return false
	}
	if plan.Where != nil && !check(*plan.Where, 1) {
		return invalid()
	}
	return nil
}

// QueryKeyset 生成混合升降序的词典序尾部条件；唯一主键末项及模型类型由生成器复核。
// 游标的 scope、授权和查询摘要绑定仍由上层 owner 维护，不在通用 ORM 中推断。
func QueryKeyset(order []QueryOrder, after []QueryScalar) (*QueryExpression, *cd.Error) {
	if len(order) == 0 || len(order) > 8 || len(order) != len(after) {
		return nil, cd.NewError(cd.IllegalParam, "keyset order and values differ")
	}
	terms := make([]QueryExpression, 0, len(order))
	for i, item := range order {
		if after[i].Kind == "null" {
			return nil, cd.NewError(cd.IllegalParam, "keyset null ordering is unsupported")
		}
		children := make([]QueryExpression, 0, i+1)
		for j := 0; j <= i; j++ {
			op := "eq"
			if j == i {
				op = "gt"
				if item.Direction == "desc" {
					op = "lt"
				}
			}
			value := after[j]
			children = append(children, QueryExpression{Kind: "compare", Field: order[j].Field, Operator: op, Value: &value})
		}
		term := children[0]
		if len(children) > 1 {
			term = QueryExpression{Kind: "and", Children: children}
		}
		terms = append(terms, term)
	}
	result := terms[0]
	if len(terms) > 1 {
		result = QueryExpression{Kind: "or", Children: terms}
	}
	if err := ValidateReadQueryPlan(ReadQueryPlan{APIVersion: ReadQueryPlanAPIVersion, Where: &result, OrderBy: order}); err != nil {
		return nil, err
	}
	return &result, nil
}
