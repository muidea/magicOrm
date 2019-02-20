package local

import (
	"fmt"
	"reflect"
)

type valueImpl struct {
	valueImpl reflect.Value
}

func newValue(val reflect.Value) (ret *valueImpl, err error) {
	ret = &valueImpl{valueImpl: val}
	return
}

func (s *valueImpl) IsNil() (ret bool) {
	if s.valueImpl.Kind() == reflect.Ptr {
		return s.valueImpl.IsNil()
	}

	ret = s.valueImpl.Kind() == reflect.Invalid

	return
}

func (s *valueImpl) Set(val reflect.Value) (err error) {
	if val.Kind() == reflect.Invalid {
		return
	}

	if s.valueImpl.Kind() == reflect.Invalid {
		s.valueImpl = val
		return
	}

	valTypeName := val.Type().String()
	expectTypeName := s.valueImpl.Type().String()
	if expectTypeName != valTypeName {
		err = fmt.Errorf("illegal value type, type:%s, expect:%s", expectTypeName, valTypeName)
		return
	}

	s.valueImpl.Set(val)
	return
}

func (s *valueImpl) Get() (ret reflect.Value) {
	ret = s.valueImpl

	return
}

func (s *valueImpl) Copy() (ret *valueImpl) {
	ret = &valueImpl{valueImpl: s.valueImpl}

	return
}
