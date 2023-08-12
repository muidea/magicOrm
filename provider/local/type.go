package local

import (
	"fmt"
	"path"
	"reflect"

	"github.com/muidea/magicOrm/model"
	pu "github.com/muidea/magicOrm/provider/util"
)

type typeImpl struct {
	typeVal reflect.Type
}

func getValueType(val reflect.Value) (ret *typeImpl, err error) {
	if pu.IsNil(val) {
		err = fmt.Errorf("can't get nil value type")
		return
	}

	ret, err = newType(val.Type())
	return
}

func newType(val reflect.Type) (ret *typeImpl, err error) {
	rType := val
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	_, err = pu.GetTypeEnum(rType)
	if err != nil {
		return
	}

	ret = &typeImpl{typeVal: val}
	return
}

func (s *typeImpl) GetName() string {
	rType := s.getElemType()
	return rType.Name()
}

func (s *typeImpl) GetValue() (ret model.TypeDeclare) {
	rType := s.getRawType()
	ret, _ = pu.GetTypeEnum(rType)
	return
}

func (s *typeImpl) GetPkgPath() string {
	rType := s.getElemType()
	return rType.PkgPath()
}

func (s *typeImpl) GetPkgKey() string {
	rType := s.getElemType()
	return path.Join(rType.PkgPath(), rType.Name())
}

func (s *typeImpl) IsPtrType() bool {
	return s.typeVal.Kind() == reflect.Ptr
}

func (s *typeImpl) Interface() (ret model.Value) {
	tVal := reflect.New(s.typeVal).Elem()
	if s.IsPtrType() {
		rVal := reflect.New(s.getRawType())
		tVal.Set(rVal)
	}

	ret = pu.NewValue(tVal)
	return
}

func (s *typeImpl) Elem() model.Type {
	tType := s.getRawType()
	if tType.Kind() == reflect.Slice {
		return &typeImpl{typeVal: tType.Elem()}
	}

	return &typeImpl{typeVal: s.typeVal}
}

func (s *typeImpl) IsBasic() bool {
	elemType := s.Elem()

	return model.IsBasicType(elemType.GetValue())
}

func (s *typeImpl) getRawType() reflect.Type {
	if s.typeVal.Kind() == reflect.Ptr {
		return s.typeVal.Elem()
	}

	return s.typeVal
}

func (s *typeImpl) getElemType() reflect.Type {
	rType := s.getRawType()
	if rType.Kind() == reflect.Slice {
		rType = rType.Elem()
	}
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	return rType
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
