package local

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
	"github.com/muidea/magicOrm/utils"
)

type ValueImpl struct {
	value reflect.Value
}

var NilValue = ValueImpl{}

func NewValue(val reflect.Value) (ret *ValueImpl) {
	ret = &ValueImpl{value: val}
	return
}

func (s *ValueImpl) IsValid() (ret bool) {
	/*
		if !s.value.IsValid() {
			return false
		}

		rawVal := reflect.Indirect(s.value)
		switch rawVal.Kind() {
		case reflect.Slice, reflect.Map:
			return !rawVal.IsNil()
		default:
		}
		ret = rawVal.IsValid()*/
	ret = utils.IsReallyValidForReflect(s.value)
	return
}

func (s *ValueImpl) IsZero() (ret bool) {
	ret = utils.IsReallyZeroForReflect(s.value)
	return
}

func (s *ValueImpl) Set(val any) {
	s.value = val.(reflect.Value)
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

func (s *ValueImpl) Interface() model.RawVal {
	if !utils.IsReallyValidForReflect(s.value) {
		return nil
	}

	return model.NewRawVal(s.value.Interface())
}

func (s *ValueImpl) IsBasic() bool {
	if !utils.IsReallyValidForReflect(s.value) {
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
	if utils.IsReallyValidForReflect(s.value) {
		ret = &ValueImpl{value: reflect.New(s.value.Type()).Elem()}
		ret.value.Set(s.value)
		return
	}

	ret = &ValueImpl{}
	return
}
