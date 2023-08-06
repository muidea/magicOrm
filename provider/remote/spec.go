package remote

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/muidea/magicOrm/model"
)

const (
	fieldName      = "fieldName"
	fieldType      = "fieldType"
	valueDeclare   = "valueDeclare"
	fieldKey       = "key"
	fieldAuto      = "auto"
	fieldUUID      = "uuid"
	fieldSnowFlake = "snowflake"
)

type SpecImpl map[string]string

func (s SpecImpl) GetFieldName() string {
	val, ok := s[fieldName]
	if ok {
		return val
	}

	return ""
}

func (s SpecImpl) IsPrimaryKey() (ret bool) {
	val, ok := s[fieldType]
	if ok {
		return val == fieldKey
	}

	return false
}

func (s SpecImpl) GetValueDeclare() model.ValueDeclare {
	val, ok := s[valueDeclare]
	if !ok {
		return model.Customer
	}

	switch val {
	case model.AutoIncrement.String():
		return model.AutoIncrement
	case model.UUID.String():
		return model.UUID
	case model.SnowFlake.String():
		return model.SnowFlake
	}

	return model.Customer
}

func (s SpecImpl) IsAutoIncrement() (ret bool) {
	val, ok := s[valueDeclare]
	if !ok {
		return false
	}

	return val == model.AutoIncrement.String()
}

func (s SpecImpl) copy() *SpecImpl {
	ret := SpecImpl{}

	for k, v := range s {
		ret[k] = v
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

	ret = SpecImpl{}

	ret[fieldName] = items[0]
	for idx := 1; idx < len(items); idx++ {
		switch items[idx] {
		case fieldAuto:
			ret[valueDeclare] = model.AutoIncrement.String()
		case fieldUUID:
			ret[valueDeclare] = model.UUID.String()
		case fieldSnowFlake:
			ret[valueDeclare] = model.SnowFlake.String()
		case fieldKey:
			ret[fieldType] = fieldKey
		}
	}

	return
}

func compareSpec(l, r *SpecImpl) bool {
	if l == nil && r == nil {
		return true
	}

	if l != nil && r != nil {
		for k, v := range *l {
			rv, rk := (*r)[k]
			if !rk || rv != v {
				return false
			}
		}

		return true
	}

	return false
}
