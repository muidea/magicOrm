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

type valueImpl struct {
	value reflect.Value
}

func newValue(val reflect.Value) (ret *valueImpl) {
	ret = &valueImpl{value: reflect.Indirect(val)}
	return
}

func (s *valueImpl) IsNil() (ret bool) {
	ret = util.IsNil(s.value)
	return
}

func (s *valueImpl) Set(val reflect.Value) (err error) {
	rVal := reflect.Indirect(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
		rVal = reflect.Indirect(rVal)
	}

	if util.IsNil(s.value) || util.IsNil(rVal) {
		s.value = rVal
		return
	}

	s.value.Set(rVal)
	return
}

func (s *valueImpl) Get() reflect.Value {
	return reflect.Indirect(s.value)
}

func (s *valueImpl) Addr() model.Value {
	if !s.value.CanAddr() {
		panic("illegal value")
	}

	impl := &valueImpl{value: s.value.Addr()}
	return impl
}

func (s *valueImpl) Interface() any {
	if util.IsNil(s.value) {
		return nil
	}

	return s.value.Interface()
}

func (s *valueImpl) IsBasic() bool {
	if util.IsNil(s.value) {
		return false
	}

	rType := s.value.Type()
	if s.value.Kind() == reflect.Interface {
		rType = s.value.Elem().Type()
	}
	if util.IsSlice(rType) {
		rType = rType.Elem()
	}

	return !util.IsStruct(rType)
}

func (s *valueImpl) copy() (ret *valueImpl) {
	if !util.IsNil(s.value) {
		ret = &valueImpl{value: reflect.New(s.value.Type()).Elem()}
		ret.value.Set(s.value)
		return
	}

	ret = &valueImpl{}
	return
}
