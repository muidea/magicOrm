package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

type typeImpl struct {
	typeImpl reflect.Type
}

func (s *typeImpl) Name() string {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem().Name()
	}
	return s.typeImpl.Name()
}

func (s *typeImpl) Value() (ret int) {
	if s.typeImpl.Kind() == reflect.Ptr {
		ret, _ = util.GetTypeValueEnum(s.typeImpl.Elem())
		return
	}

	ret, _ = util.GetTypeValueEnum(s.typeImpl)
	return
}

func (s *typeImpl) IsPtr() bool {
	return s.typeImpl.Kind() == reflect.Ptr
}

func (s *typeImpl) PkgPath() string {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem().PkgPath()
	}
	return s.typeImpl.PkgPath()
}

func (s *typeImpl) String() string {
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.Value(), s.Name(), s.PkgPath(), s.IsPtr())
}

func (s *typeImpl) Depend() reflect.Type {
	rawVal := s.typeImpl
	if rawVal.Kind() == reflect.Ptr {
		rawVal = rawVal.Elem()
	}
	if rawVal.Kind() == reflect.Struct {
		if rawVal.String() != "time.Time" {
			return s.typeImpl
		}
	}
	if rawVal.Kind() == reflect.Slice {
		rawVal = rawVal.Elem()
		if rawVal.String() != "time.Time" {
			return rawVal
		}
	}

	return nil
}

func (s *typeImpl) Copy() model.FieldType {
	return &typeImpl{
		typeImpl: s.typeImpl,
	}
}

// NewFieldType NewFieldType
func NewFieldType(val reflect.Type) (ret model.FieldType, err error) {
	rawVal := val
	if rawVal.Kind() == reflect.Ptr {
		rawVal = rawVal.Elem()
	}
	if rawVal.Kind() != reflect.Slice {
		_, tErr := util.GetTypeValueEnum(rawVal)
		if tErr != nil {
			err = tErr
			return
		}
		ret = &typeImpl{typeImpl: val}
		return
	}

	// check slice elemnt type
	sliceVal := rawVal.Elem()
	if sliceVal.Kind() == reflect.Ptr {
		sliceVal = sliceVal.Elem()
	}
	if sliceVal.Kind() == reflect.Slice {
		err = fmt.Errorf("illegal slice element type")
		return
	}

	_, tErr := util.GetTypeValueEnum(sliceVal)
	if tErr != nil {
		err = tErr
		return
	}
	ret = &typeImpl{typeImpl: val}
	return
}

// NewFieldType NewFieldType
/*
func NewFieldType(val reflect.Type) (ret model.FieldType, err error) {
	isPtr := false
	rawVal := val
	if rawVal.Kind() == reflect.Ptr {
		rawVal = rawVal.Elem()
		isPtr = true
	}

	tVal, tErr := util.GetTypeValueEnum(rawVal)
	if tErr != nil {
		err = tErr
		return
	}
	if util.IsBasicType(tVal) {
		ret, err = getBasicType(rawVal, isPtr)
		return
	}

	if util.IsStructType(tVal) {
		ret, err = getStructType(rawVal, isPtr)
		return
	}

	if util.IsSliceType(tVal) {
		ret, err = getSliceType(rawVal, isPtr)
		return
	}

	err = fmt.Errorf("illegal fieldType, type:%s", val.String())
	return
}
*/
