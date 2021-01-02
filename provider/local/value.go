package local

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"

	"github.com/muidea/magicOrm/util"
)

type valueImpl struct {
	value reflect.Value
}

func GetValue(val reflect.Value) model.Value {
	return newValue(val)
}

func newValue(val reflect.Value) (ret *valueImpl) {
	ret = &valueImpl{value: val}
	return
}

func (s *valueImpl) IsNil() (ret bool) {
	ret = util.IsNil(s.value)
	return
}

func (s *valueImpl) Set(val interface{}) (err error) {
	v, ok := val.(reflect.Value)
	if !ok {
		err = fmt.Errorf("illegal set value")
		return
	}

	s.value = v
	return
}

func (s *valueImpl) Update(val interface{}) (err error) {
	v, ok := val.(reflect.Value)
	if !ok {
		err = fmt.Errorf("illegal update value")
		return
	}
	if !util.IsNil(s.value) {
		if v.Type().String() != s.value.Type().String() {
			err = fmt.Errorf("illegal update value")
			return
		}

		s.value.Set(v)
		return
	}

	s.value = v
	return
}

func (s *valueImpl) Get() (ret interface{}) {
	ret = s.value

	return
}

func (s *valueImpl) Addr() model.Value {
	impl := &valueImpl{value: s.value.Addr()}
	return impl
}

func (s *valueImpl) Type() (model.Type, error) {
	return newType(s.value.Type())
}

func (s *valueImpl) copy() (ret *valueImpl) {
	ret = &valueImpl{value: s.value}

	return
}
