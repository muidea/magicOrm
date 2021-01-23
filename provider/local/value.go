package local

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

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
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
		val = reflect.Indirect(val)
	}

	if util.IsNil(s.value) {
		s.value = val
		return
	}

	s.value.Set(val)
	return
}

func (s *valueImpl) Get() (ret reflect.Value) {
	ret = s.value
	return
}

func (s *valueImpl) Addr() model.Value {
	impl := &valueImpl{value: s.value.Addr()}
	return impl
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
