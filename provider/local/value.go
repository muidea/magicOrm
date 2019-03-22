package local

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/util"
)

type valueImpl struct {
	valueImpl reflect.Value
}

func newValue(val reflect.Value) (ret *valueImpl, err error) {
	val = reflect.Indirect(val)
	_, err = util.GetTypeValueEnum(val.Type())
	if err != nil {
		return
	}

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
		err = fmt.Errorf("invalid set value")
		return
	}

	s.valueImpl = val
	return
}

func (s *valueImpl) Update(val reflect.Value) (err error) {
	if s.valueImpl.Kind() == reflect.Invalid {
		err = fmt.Errorf("invalid current value")
		return
	}

	if val.Kind() == reflect.Invalid {
		err = fmt.Errorf("invalid update value")
		return
	}

	valTypeName := val.Type().String()
	expectTypeName := s.valueImpl.Type().String()
	if expectTypeName != valTypeName {
		err = fmt.Errorf("illegal value type, type:%s, expect:%s", valTypeName, expectTypeName)
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
