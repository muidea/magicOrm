package remote

import (
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

type SpecImpl struct {
	FieldName    string               `json:"fieldName"`
	PrimaryKey   bool                 `json:"primaryKey"`
	ValueDeclare models.ValueDeclare  `json:"valueDeclare"`
	ViewDeclare  []models.ViewDeclare `json:"viewDeclare"`

	Constraints  string `json:"constraints"`
	DefaultValue any    `json:"defaultValue"`

	// 这里是为了避免在使用时多次解析
	constraintsVal models.Constraints
}

var emptySpec = SpecImpl{PrimaryKey: false, ValueDeclare: models.Customer}

func (m SpecImpl) GetFieldName() string {
	return m.FieldName
}

func (m SpecImpl) IsPrimaryKey() bool {
	return m.PrimaryKey
}

func (m SpecImpl) GetValueDeclare() models.ValueDeclare {
	return m.ValueDeclare
}

func (m SpecImpl) GetConstraints() models.Constraints {
	if m.constraintsVal == nil {
		if m.Constraints != "" {
			m.constraintsVal = utils.ParseConstraints(m.Constraints)
		}
	}

	return m.constraintsVal
}

func (m SpecImpl) EnableView(viewSpec models.ViewDeclare) bool {
	if viewSpec == models.MetaView {
		return true
	}

	for _, val := range m.ViewDeclare {
		if val == viewSpec {
			return true
		}
	}

	return false
}

// GetDefaultValue
// 这里只允许是基本数值,不允许是表达式，不允许是[]any
func (m SpecImpl) GetDefaultValue() any {
	return m.DefaultValue
}

func (m SpecImpl) Copy() *SpecImpl {
	ret := SpecImpl{
		FieldName:    m.FieldName,
		PrimaryKey:   m.PrimaryKey,
		ValueDeclare: m.ValueDeclare,
		ViewDeclare:  m.ViewDeclare,
		Constraints:  m.Constraints,
		DefaultValue: m.DefaultValue,
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
