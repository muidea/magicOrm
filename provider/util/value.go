package util

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
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
	ret = IsNil(s.value)
	return
}

func (s *ValueImpl) IsZero() (ret bool) {
	ret = IsNil(s.value) || IsZero(s.value)
	return
}

func (s *ValueImpl) Set(val reflect.Value) (err error) {
	//rVal := reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
		//rVal = reflect.Indirect(rVal)
	}

	if !val.IsValid() {
		s.value = val
		return
	}

	if !s.value.IsValid() || s.value.IsZero() {
		s.value = reflect.New(val.Type()).Elem()
	}
	//if !s.value.IsValid() || util.IsNil(s.value) || util.IsNil(val) {
	//	s.value = val
	//	return
	//}

	// special for ptr value
	if s.value.Kind() == reflect.Ptr && val.Kind() != reflect.Ptr {
		val = val.Addr()
	}
	// special for struct value
	if s.value.Kind() == reflect.Struct {
		val = reflect.Indirect(val)
	}

	s.value.Set(val)
	return
}

func (s *ValueImpl) Get() reflect.Value {
	return s.value
}

func (s *ValueImpl) Addr() model.Value {
	if !s.value.CanAddr() {
		panic("illegal value")
	}

	impl := &ValueImpl{value: s.value.Addr()}
	return impl
}

func (s *ValueImpl) Interface() any {
	if IsNil(s.value) {
		return nil
	}

	return s.value.Interface()
}

func (s *ValueImpl) IsBasic() bool {
	if IsNil(s.value) {
		return false
	}

	rType := s.value.Type()
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	if s.value.Kind() == reflect.Interface {
		rType = s.value.Elem().Type()
	}
	if IsSlice(rType) {
		rType = rType.Elem()
	}

	return !IsStruct(rType)
}

func (s *ValueImpl) Verify() error {
	if s.IsNil() {
		return nil
	}

	if !s.value.CanAddr() || !s.value.CanSet() {
		return fmt.Errorf("illegal vlaue")
	}

	return nil
}

func (s *ValueImpl) Copy() (ret *ValueImpl) {
	if !IsNil(s.value) {
		ret = &ValueImpl{value: reflect.New(s.value.Type()).Elem()}
		ret.value.Set(s.value)
		return
	}

	ret = &ValueImpl{}
	return
}
