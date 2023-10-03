package local

import (
	"fmt"
	"path"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
)

type TypeImpl struct {
	typeVal reflect.Type
}

func getValueType(val reflect.Value) (ret *TypeImpl, err error) {
	if util.IsNil(val) {
		err = fmt.Errorf("can't get nil value type")
		return
	}

	ret, err = NewType(val.Type())
	return
}

func NewType(val reflect.Type) (ret *TypeImpl, err error) {
	rType := val
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	_, err = util.GetTypeEnum(rType)
	if err != nil {
		return
	}

	ret = &TypeImpl{typeVal: val}
	return
}

func (s *TypeImpl) GetName() string {
	rType := s.getElemType()
	return rType.Name()
}

func (s *TypeImpl) GetPkgPath() string {
	rType := s.getElemType()
	return rType.PkgPath()
}

func (s *TypeImpl) GetDescription() string {
	return ""
}

func (s *TypeImpl) GetValue() (ret model.TypeDeclare) {
	rType := s.getRawType()
	ret, _ = util.GetTypeEnum(rType)
	return
}

func (s *TypeImpl) GetPkgKey() string {
	rType := s.getElemType()
	return path.Join(rType.PkgPath(), rType.Name())
}

func (s *TypeImpl) IsPtrType() bool {
	return s.typeVal.Kind() == reflect.Ptr
}

func (s *TypeImpl) Interface(initVal any) (ret model.Value, err error) {
	tVal := reflect.New(s.getRawType()).Elem()

	if initVal != nil {
		switch s.GetValue() {
		case model.TypeBooleanValue:
			initVal, err = util.GetBool(initVal)
			tVal.SetBool(initVal.(bool))
		case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
			initVal, err = util.GetInt64(initVal)
			tVal.SetInt(initVal.(int64))
		case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
			initVal, err = util.GetUint64(initVal)
			tVal.SetUint(initVal.(uint64))
		case model.TypeFloatValue, model.TypeDoubleValue:
			initVal, err = util.GetFloat64(initVal)
			tVal.SetFloat(initVal.(float64))
		case model.TypeStringValue:
			initVal, err = util.GetString(initVal)
			tVal.SetString(initVal.(string))
		case model.TypeDateTimeValue:
			initVal, err = util.GetDateTimeDt(initVal)
			tVal.Set(reflect.ValueOf(initVal))
		default:
			initVal = nil
		}
	}

	if s.IsPtrType() {
		tVal = tVal.Addr()
	}

	ret = NewValue(tVal)
	return
}

func (s *TypeImpl) Elem() model.Type {
	tType := s.getRawType()
	if tType.Kind() == reflect.Slice {
		return &TypeImpl{typeVal: tType.Elem()}
	}

	return &TypeImpl{typeVal: s.typeVal}
}

func (s *TypeImpl) IsBasic() bool {
	elemType := s.Elem()

	return model.IsBasicType(elemType.GetValue())
}

func (s *TypeImpl) IsStruct() bool {
	elemType := s.Elem()
	return model.IsStructType(elemType.GetValue())
}

func (s *TypeImpl) IsSlice() bool {
	return model.IsSliceType(s.GetValue())
}

func (s *TypeImpl) getRawType() reflect.Type {
	if s.typeVal.Kind() == reflect.Ptr {
		return s.typeVal.Elem()
	}

	return s.typeVal
}

func (s *TypeImpl) getElemType() reflect.Type {
	rType := s.getRawType()
	if rType.Kind() == reflect.Slice {
		rType = rType.Elem()
	}
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	return rType
}

func (s *TypeImpl) copy() (ret *TypeImpl) {
	ret = &TypeImpl{
		typeVal: s.typeVal,
	}

	return
}

func (s *TypeImpl) dump() string {
	val := s.GetValue()
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", val, s.GetName(), s.GetPkgPath(), s.IsPtrType())
}
