package local

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type typeImpl struct {
	typeImpl reflect.Type
}

// newType newType
func newType(val reflect.Type) (ret *typeImpl, err error) {
	rawType := val
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}
	_, err = util.GetTypeValueEnum(rawType)
	if err != nil {
		return
	}

	ret = &typeImpl{typeImpl: val}
	return
}

func (s *typeImpl) GetName() string {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem().Name()
	}

	return s.typeImpl.Name()
}

func (s *typeImpl) GetValue() (ret int) {
	if s.typeImpl.Kind() == reflect.Ptr {
		ret, _ = util.GetTypeValueEnum(s.typeImpl.Elem())
		return
	}

	ret, _ = util.GetTypeValueEnum(s.typeImpl)
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

func (s *typeImpl) IsPtrType() bool {
	return s.typeImpl.Kind() == reflect.Ptr
}

func (s *typeImpl) Interface() reflect.Value {
	rawType := s.GetType()
	val := reflect.New(rawType)
	if !s.IsPtrType() {
		val = val.Elem()
	}

	return val
}

func (s *typeImpl) Elem() model.Type {
	tVal := s.GetType()
	if tVal.Kind() == reflect.Slice {
		return &typeImpl{typeImpl: tVal.Elem()}
	}

	return nil
}

func (s *typeImpl) Copy() (ret *typeImpl) {
	ret = &typeImpl{
		typeImpl: s.typeImpl,
	}

	return
}

func (s *typeImpl) Dump() string {
	val := s.GetValue()
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", val, s.GetName(), s.GetPkgPath(), s.IsPtrType())
}
