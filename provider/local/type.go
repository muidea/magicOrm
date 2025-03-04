package local

import (
	"fmt"
	"path"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
	"github.com/muidea/magicOrm/utils"
)

type TypeImpl struct {
	typeVal reflect.Type
}

func getValueType(val reflect.Value) (ret *TypeImpl, err *cd.Result) {
	if !utils.IsReallyValidTypeForReflect(val.Type()) {
		err = cd.NewResult(cd.UnExpected, "can't get nil value type")
		return
	}

	ret, err = NewType(val.Type())
	return
}

func NewType(val reflect.Type) (ret *TypeImpl, err *cd.Result) {
	log.Infof("NewType: %v", val)
	rType := val
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	tVal, tErr := util.GetTypeEnum(rType)
	if tErr != nil {
		err = tErr
		return
	}
	if tVal == model.TypeMapValue {
		err = cd.NewResult(cd.UnExpected, "unsupported map type")
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

func (s *TypeImpl) Interface(initVal any) (ret model.Value, err *cd.Result) {
	tVal := reflect.New(s.getRawType()).Elem()

	if initVal != nil {
		switch s.GetValue() {
		case model.TypeBooleanValue:
			rawVal, rawErr := util.GetBool(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetBool(rawVal.Value().(bool))
		case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
			rawVal, rawErr := util.GetInt64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetInt(rawVal.Value().(int64))
		case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
			rawVal, rawErr := util.GetUint64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetUint(rawVal.Value().(uint64))
		case model.TypeFloatValue, model.TypeDoubleValue:
			rawVal, rawErr := util.GetFloat64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetFloat(rawVal.Value().(float64))
		case model.TypeStringValue:
			rawVal, rawErr := util.GetString(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetString(rawVal.Value().(string))
		case model.TypeDateTimeValue:
			rawVal, rawErr := util.GetDateTime(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.Set(reflect.ValueOf(rawVal.Value()))
		default:
			rInitVal := reflect.Indirect(reflect.ValueOf(initVal))
			if rInitVal.Type() != tVal.Type() {
				err = cd.NewResult(cd.UnExpected, "missmatch value type")
			} else {
				tVal.Set(rInitVal)
			}
		}
	}
	if err != nil {
		return
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
