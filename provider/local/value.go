package local

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
)

type ValueImpl struct {
	value reflect.Value
}

func NewValue(val reflect.Value) (ret *ValueImpl) {
	ret = &ValueImpl{value: val}
	return
}

func (s *ValueImpl) IsValid() bool {
	if !s.value.IsValid() {
		return false
	}
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return false
		}
	}
	if s.value.Kind() == reflect.Slice || s.value.Kind() == reflect.Map {
		if s.value.IsNil() {
			return false
		}
	}

	return true
}

func (s *ValueImpl) IsZero() bool {
	if !s.value.IsValid() {
		return true
	}
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return true
		}
		return s.value.Elem().IsZero()
	}
	if s.value.Kind() == reflect.Slice || s.value.Kind() == reflect.Map {
		if s.value.IsNil() {
			return true
		}
		return s.value.Len() == 0
	}

	return s.value.IsZero()
}

func (s *ValueImpl) Set(val any) (err *cd.Error) {
	if !s.value.CanSet() {
		err = cd.NewError(cd.UnExpected, "Set failed, value is not settable")
		log.Warnf("Set failed, value is not settable")
		return
	}
	if !s.value.IsValid() {
		log.Errorf("Set failed, value is not valid, s.value canSet:%+v", s.value.CanSet())
		return
	}

	rVal := reflect.ValueOf(val)
	isPtr := s.value.Kind() == reflect.Ptr
	if !isPtr {
		if rVal.Type() != s.value.Type() {
			err = cd.NewError(cd.UnExpected, "Set failed, value type is not match")
			log.Warnf("Set failed, value type is not match, data type:%+v, value type:%+v", rVal.Type(), s.value.Type())
			return
		}

		s.value.Set(rVal)
		return
	}

	rVal = reflect.Indirect(rVal)
	if rVal.Type() != s.value.Type().Elem() {
		err = cd.NewError(cd.UnExpected, "Set failed, value type is not match")
		log.Warnf("Set failed, value type is not match")
		return
	}

	reallyValPtr := reflect.New(s.value.Type().Elem())
	reallyValPtr.Elem().Set(rVal)
	s.value.Set(reallyValPtr)
	return
}

func (s *ValueImpl) Get() any {
	return s.value.Interface()
}

// UnpackValue expands the contained value into individual elements.
// For slices, it returns each element as separate reflect.Value entries.
// For non-slice values, returns a single-element slice containing the value.
func (s *ValueImpl) UnpackValue() (ret []model.Value) {
	ret = []model.Value{}
	realVal := reflect.Indirect(s.value)
	if realVal.Kind() == reflect.Slice {
		for idx := 0; idx < realVal.Len(); idx++ {
			ret = append(ret, NewValue(realVal.Index(idx)))
		}
	} else {
		ret = append(ret, NewValue(s.value))
	}
	return
}

// Append appends the given value to the slice value.
func (s *ValueImpl) Append(val reflect.Value) (err *cd.Error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("%v", errInfo))
		}
	}()

	isPtr := s.value.Kind() == reflect.Ptr
	if !isPtr {
		//if !s.IsValid() {
		//	initSlice := reflect.MakeSlice(s.value.Type(), 0, 0)
		//	s.value.Set(initSlice)
		//}

		s.value.Set(reflect.Append(s.value, val))
		return
	}

	if !s.value.IsValid() || s.value.IsZero() {
		initSlice := reflect.MakeSlice(s.value.Type().Elem(), 0, 0)
		initSlicePtr := reflect.New(s.value.Type().Elem())
		initSlicePtr.Elem().Set(initSlice)
		s.value.Set(initSlicePtr)
	}

	reallySlice := reflect.Indirect(s.value)
	reallySlice = reflect.Append(reallySlice, val)
	reallySlicePtr := reflect.New(reallySlice.Type())
	reallySlicePtr.Elem().Set(reallySlice)
	s.value.Set(reallySlicePtr)
	return
}

func (s *ValueImpl) reset(needInit bool) {
	if !s.value.IsValid() {
		return
	}

	if s.value.Kind() == reflect.Ptr {
		if needInit {
			s.value.Set(reflect.New(s.value.Type().Elem()))
		} else {
			s.value.Set(reflect.Zero(s.value.Type()))
		}
		return
	}

	if s.value.Kind() == reflect.Slice || s.value.Kind() == reflect.Map {
		if needInit {
			if s.value.Kind() == reflect.Slice {
				s.value.Set(reflect.MakeSlice(s.value.Type(), 0, 0))
			} else {
				s.value.Set(reflect.MakeMap(s.value.Type()))
			}
		} else {
			s.value.Set(reflect.Zero(s.value.Type()))
		}
		return
	}

	s.value.SetZero()
}
