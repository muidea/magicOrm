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

	if util.IsNil(s.value) {
		s.value = v
		return
	}

	s.value.Set(v)
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

func (s *valueImpl) copy() (ret *valueImpl) {
	ret = &valueImpl{value: s.value}

	return
}
