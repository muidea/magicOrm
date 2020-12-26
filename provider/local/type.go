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
	_, err = util.GetTypeEnum(rawType)
	if err != nil {
		return
	}

	ret = &typeImpl{typeImpl: val}
	return
}

func (s *typeImpl) GetName() string {
	tType := s.getType()
	return tType.String()
}

func (s *typeImpl) GetValue() (ret int) {
	tType := s.getType()
	ret, _ = util.GetTypeEnum(tType)
	return
}

func (s *typeImpl) GetPkgPath() string {
	tType := s.getType()
	return tType.PkgPath()
}

func (s *typeImpl) IsPtrType() bool {
	return s.typeImpl.Kind() == reflect.Ptr
}

func (s *typeImpl) Interface() reflect.Value {
	tType := s.getType()
	val := reflect.New(tType)
	if !s.IsPtrType() {
		val = val.Elem()
	}

	return val
}

func (s *typeImpl) Elem() model.Type {
	tType := s.getType()
	if tType.Kind() == reflect.Slice {
		return &typeImpl{typeImpl: tType.Elem()}
	}

	return &typeImpl{typeImpl: s.typeImpl}
}

func (s *typeImpl) IsBasic() bool {
	elemType := s.Elem()

	return util.IsBasicType(elemType.GetValue())
}

func (s *typeImpl) getType() reflect.Type {
	if s.typeImpl.Kind() == reflect.Ptr {
		return s.typeImpl.Elem()
	}

	return s.typeImpl
}

func (s *typeImpl) copy() (ret *typeImpl) {
	ret = &typeImpl{
		typeImpl: s.typeImpl,
	}

	return
}

func (s *typeImpl) dump() string {
	val := s.GetValue()
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", val, s.GetName(), s.GetPkgPath(), s.IsPtrType())
}
