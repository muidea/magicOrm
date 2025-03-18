package remote

import (
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
)

type SpecImpl struct {
	FieldName    string              `json:"fieldName"`
	PrimaryKey   bool                `json:"primaryKey"`
	ValueDeclare model.ValueDeclare  `json:"valueDeclare"`
	ViewDeclare  []model.ViewDeclare `json:"viewDeclare"`
	DefaultValue any                 `json:"defaultValue"`
}

var emptySpec = SpecImpl{PrimaryKey: false, ValueDeclare: model.Customer}

func (s SpecImpl) GetFieldName() string {
	return s.FieldName
}

func (s SpecImpl) IsPrimaryKey() bool {
	return s.PrimaryKey
}

func (s SpecImpl) GetValueDeclare() model.ValueDeclare {
	return s.ValueDeclare
}

func (s SpecImpl) EnableView(viewSpec model.ViewDeclare) bool {
	if viewSpec == model.MetaView {
		return true
	}

	for _, val := range s.ViewDeclare {
		if val == viewSpec {
			return true
		}
	}

	return false
}

// GetDefaultValue
// 这里只允许是基本数值,不允许是表达式，不允许是[]any
func (s SpecImpl) GetDefaultValue() any {
	if s.DefaultValue == nil {
		return nil
	}

	if !utils.IsReallyValidValue(s.DefaultValue) {
		return nil
	}

	switch val := s.DefaultValue.(type) {
	case string:
		if strings.Contains(val, "$referenceValue.") {
			return nil
		}

		return val
	case []any:
		return nil
	default:
		return s.DefaultValue
	}
}

func (s SpecImpl) Copy() *SpecImpl {
	ret := SpecImpl{
		FieldName:    s.FieldName,
		PrimaryKey:   s.PrimaryKey,
		ValueDeclare: s.ValueDeclare,
		ViewDeclare:  s.ViewDeclare,
		DefaultValue: s.DefaultValue,
	}

	return &ret
}

func compareSpec(l, r *SpecImpl) bool {
	if l == nil && r == nil {
		return true
	}

	if l != nil && r != nil {
		return l.FieldName == r.FieldName &&
			l.PrimaryKey == r.PrimaryKey &&
			l.ValueDeclare == r.ValueDeclare
	}

	return false
}
