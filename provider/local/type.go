package local

import (
	"path"
	"reflect"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
)

type TypeImpl struct {
	typeVal reflect.Type
}

func NewType(val reflect.Type) (ret *TypeImpl, err *cd.Result) {
	rType := val
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}
	_, tErr := utils.GetTypeEnum(rType)
	if tErr != nil {
		err = tErr
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

func (s *TypeImpl) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *TypeImpl) GetDescription() string {
	return ""
}

func (s *TypeImpl) GetValue() (ret model.TypeDeclare) {
	rType := s.getRawType()
	ret, _ = utils.GetTypeEnum(rType)
	return
}

func (s *TypeImpl) IsPtrType() bool {
	return s.typeVal.Kind() == reflect.Ptr
}

func (s *TypeImpl) Interface(initVal any) (ret model.Value, err *cd.Result) {
	tVal := reflect.New(s.getRawType()).Elem()
	if initVal != nil {
		switch s.GetValue() {
		case model.TypeBooleanValue:
			rawVal, rawErr := utils.ConvertRawToBool(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetBool(rawVal)
		case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
			rawVal, rawErr := utils.ConvertRawToInt64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetInt(rawVal)
		case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
			rawVal, rawErr := utils.ConvertRawToUint64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetUint(rawVal)
		case model.TypeFloatValue, model.TypeDoubleValue:
			rawVal, rawErr := utils.ConvertRawToFloat64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetFloat(rawVal)
		case model.TypeStringValue:
			rawVal, rawErr := utils.ConvertRawToString(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetString(rawVal)
		case model.TypeDateTimeValue:
			rawVal, rawErr := utils.ConvertRawToDateTime(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.Set(reflect.ValueOf(rawVal))
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
