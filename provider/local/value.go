package local

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
)

type ValueImpl struct {
	value reflect.Value
}

var NilValue = ValueImpl{}

func NewValue(val reflect.Value) (ret *ValueImpl) {
	ret = &ValueImpl{value: val}
	return
}

func (s *ValueImpl) IsNil() (ret bool) {
	ret = util.IsNil(s.value)
	return
}

func (s *ValueImpl) IsZero() (ret bool) {
	ret = util.IsZero(s.value)
	return
}

func (s *ValueImpl) Set(val any) (err error) {
	s.value = val.(reflect.Value)
	return
}

func (s *ValueImpl) Get() any {
	return s.value
}

func (s *ValueImpl) Addr() model.Value {
	if !s.value.CanAddr() {
		panic("illegal value, can't addr")
	}

	impl := &ValueImpl{value: s.value.Addr()}
	return impl
}

func (s *ValueImpl) Interface() any {
	if util.IsNil(s.value) {
		return nil
	}

	return s.value.Interface()
}

func (s *ValueImpl) IsBasic() bool {
	if util.IsNil(s.value) {
		return false
	}

	rType := s.value.Type()
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	if s.value.Kind() == reflect.Interface {
		rType = s.value.Elem().Type()
	}
	if util.IsSlice(rType) {
		rType = rType.Elem()
	}

	return !util.IsStruct(rType)
}

func (s *ValueImpl) Copy() (ret *ValueImpl) {
	if !util.IsNil(s.value) {
		ret = &ValueImpl{value: reflect.New(s.value.Type()).Elem()}
		ret.value.Set(s.value)
		return
	}

	ret = &ValueImpl{}
	return
}
