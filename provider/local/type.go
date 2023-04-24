package local

import (
	"fmt"
	"path"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type typeImpl struct {
	typeVal reflect.Type
}

func getValueType(val reflect.Value) (ret *typeImpl, err error) {
	if util.IsNil(val) {
		err = fmt.Errorf("can't get nil value type")
		return
	}

	ret, err = newType(val.Type())
	return
}

func newType(val reflect.Type) (ret *typeImpl, err error) {
	rawType := val
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}
	_, err = util.GetTypeEnum(rawType)
	if err != nil {
		return
	}

	ret = &typeImpl{typeVal: val}
	return
}

func (s *typeImpl) GetName() string {
	tType := s.getType()
	if tType.Kind() == reflect.Slice {
		tType = tType.Elem()
	}
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	return tType.Name()
}

func (s *typeImpl) GetValue() (ret int) {
	tType := s.getType()
	ret, _ = util.GetTypeEnum(tType)
	return
}

func (s *typeImpl) GetPkgPath() string {
	tType := s.getType()
	if tType.Kind() == reflect.Slice {
		tType = tType.Elem()
	}
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	return tType.PkgPath()
}

func (s *typeImpl) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *typeImpl) IsPtrType() bool {
	return s.typeVal.Kind() == reflect.Ptr
}

func (s *typeImpl) Interface() (ret model.Value) {
	tType := s.getType()
	tVal := reflect.New(tType)
	if !s.IsPtrType() {
		tVal = tVal.Elem()
	}

	ret = newValue(tVal)
	return
}

func (s *typeImpl) Elem() model.Type {
	tType := s.getType()
	if tType.Kind() == reflect.Slice {
		return &typeImpl{typeVal: tType.Elem()}
	}

	return &typeImpl{typeVal: s.typeVal}
}

func (s *typeImpl) IsBasic() bool {
	elemType := s.Elem()

	return util.IsBasicType(elemType.GetValue())
}

func (s *typeImpl) getType() reflect.Type {
	if s.typeVal.Kind() == reflect.Ptr {
		return s.typeVal.Elem()
	}

	return s.typeVal
}

func (s *typeImpl) copy() (ret *typeImpl) {
	ret = &typeImpl{
		typeVal: s.typeVal,
	}

	return
}

func (s *typeImpl) dump() string {
	val := s.GetValue()
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", val, s.GetName(), s.GetPkgPath(), s.IsPtrType())
}
