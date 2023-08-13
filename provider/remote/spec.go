package remote

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
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

func newSpec(tag reflect.StructTag) (ret *SpecImpl, err error) {
	spec := tag.Get("orm")
	val, vErr := getSpec(spec)
	if vErr != nil {
		err = vErr
		return
	}

	ret = &val
	return
}

func getSpec(spec string) (ret SpecImpl, err error) {
	items := strings.Split(spec, " ")
	if len(items) < 1 {
		err = fmt.Errorf("illegal spec value, val:%s", spec)
		return
	}

	ret = SpecImpl{PrimaryKey: false, ValueDeclare: model.Customer}
	ret.FieldName = items[0]
	for idx := 1; idx < len(items); idx++ {
		switch items[idx] {
		case util.Auto:
			ret.ValueDeclare = model.AutoIncrement
		case util.UUID:
			ret.ValueDeclare = model.UUID
		case util.SnowFlake:
			ret.ValueDeclare = model.SnowFlake
		case util.DateTime:
			ret.ValueDeclare = model.DateTime
		case util.Key:
			ret.PrimaryKey = true
		}
	}

	return
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
