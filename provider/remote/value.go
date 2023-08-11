package remote

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type FieldValue struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

func (s *FieldValue) IsNil() bool {
	return s.Value == nil
}

func (s *FieldValue) Set(val reflect.Value) error {
	s.Value = val.Interface()
	return nil
}

func (s *FieldValue) Get() reflect.Value {
	return reflect.ValueOf(s.Value)
}

func (s *FieldValue) Addr() model.Value {
	impl := &FieldValue{Value: &s.Value}
	return impl
}

func (s *FieldValue) Interface() any {
	return s.Value
}

func (s *FieldValue) IsBasic() bool {
	if s.Value == nil {
		return false
	}

	rValue := reflect.ValueOf(s.Value)
	if rValue.Kind() == reflect.Interface {
		rValue = rValue.Elem()
	}
	rType := rValue.Type()
	if util.IsSlice(rType) {
		rType = rType.Elem()
	}

	return !util.IsStruct(rType)
}

func (s *FieldValue) copy() (ret *FieldValue) {
	if s.Value == nil {
		ret = &FieldValue{}
		return
	}

	ret = &FieldValue{Value: s.Value}
	return
}
