package remote

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

type SpecImpl struct {
	FieldName    string             `json:"fieldName"`
	PrimaryKey   bool               `json:"primaryKey"`
	ValueDeclare model.ValueDeclare `json:"valueDeclare"`
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

func (s SpecImpl) copy() *SpecImpl {
	ret := SpecImpl{
		FieldName:    s.FieldName,
		PrimaryKey:   s.PrimaryKey,
		ValueDeclare: s.ValueDeclare,
	}

	return &ret
}

func (s SpecImpl) dump() (ret string) {
	return fmt.Sprintf("name=%s key=%v value=%v", s.GetFieldName(), s.IsPrimaryKey(), s.GetValueDeclare())
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
