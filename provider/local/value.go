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
	ret = &valueImpl{value: val}
	return
}

func (s *valueImpl) IsNil() (ret bool) {
	ret = util.IsNil(s.value)
	return
}

func (s *valueImpl) Set(val reflect.Value) (err error) {
	//rVal := reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
		//rVal = reflect.Indirect(rVal)
	}

	if util.IsNil(s.value) || util.IsNil(val) {
		s.value = val
		return
	}

	s.value.Set(val)
	return
}

func (s *valueImpl) Get() reflect.Value {
	//return reflect.Indirect(s.value)
	return s.value
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
