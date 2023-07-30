package remote

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	fieldName = "fieldName"
	fieldAuto = "auto"
	fieldKey  = "key"
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
	_, ok := s[fieldKey]
	if ok {
		return true
	}

	return false
}

func (s SpecImpl) IsAutoIncrement() (ret bool) {
	_, ok := s[fieldAuto]
	if ok {
		return true
	}

	return false
}

func (s SpecImpl) copy() *SpecImpl {
	ret := SpecImpl{}

	for k, v := range s {
		ret[k] = v
	}
	return &ret
}

func (s SpecImpl) dump() (ret string) {
	return fmt.Sprintf("name=%s key=%v auto=%v", s.GetFieldName(), s.IsPrimaryKey(), s.IsAutoIncrement())
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
			ret[fieldAuto] = "true"
		case fieldKey:
			ret[fieldKey] = "true"
		}
	}

	return
}

func compareSpec(l, r *SpecImpl) bool {
	for k, v := range *l {
		rv, rk := (*r)[k]
		if !rk || rv != v {
			return false
		}
	}

	return true
}
