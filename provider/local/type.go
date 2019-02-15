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

// newFieldType newFieldType
func newFieldType(val reflect.Type) (ret *typeImpl, err error) {
	ret = &typeImpl{typeImpl: val}
	return
}

func (s *typeImpl) GetName() string {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem().Name()
	}

	return s.typeImpl.Name()
}

func (s *typeImpl) GetValue() (ret int, err error) {
	if s.typeImpl.Kind() == reflect.Ptr {
		ret, err = util.GetTypeValueEnum(s.typeImpl.Elem())
		return
	}

	ret, err = util.GetTypeValueEnum(s.typeImpl)
	return
}

func (s *typeImpl) GetPkgPath() string {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem().PkgPath()
	}

	return s.typeImpl.PkgPath()
}

func (s *typeImpl) GetType() reflect.Type {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem()
	}

	return s.typeImpl
}

func (s *typeImpl) GetDepend() (ret model.Model, err error) {
	return
}

func (s *typeImpl) IsPtrType() bool {
	return s.typeImpl.Kind() == reflect.Ptr
}

func (s *typeImpl) Copy() (ret *typeImpl) {
	ret = &typeImpl{
		typeImpl: s.typeImpl,
	}
	return
}

func (s *typeImpl) Dump() string {
	val, _ := s.GetValue()
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", val, s.GetName(), s.GetPkgPath(), s.IsPtrType())
}
