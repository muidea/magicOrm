package local

import (
	"path"
	"reflect"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

type TypeImpl struct {
	typeVal reflect.Type
}

func NewType(val reflect.Type) (ret *TypeImpl, err *cd.Error) {
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

func (s *TypeImpl) GetValue() (ret models.TypeDeclare) {
	rType := s.getRawType()
	ret, _ = utils.GetTypeEnum(rType)
	return
}

func (s *TypeImpl) IsPtrType() bool {
	return s.typeVal.Kind() == reflect.Ptr
}

func (s *TypeImpl) Interface(initVal any) (ret models.Value, err *cd.Error) {
	tVal := reflect.New(s.getRawType()).Elem()
	if initVal != nil {
		switch s.GetValue() {
		case models.TypeBooleanValue:
			rawVal, rawErr := utils.ConvertRawToBool(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetBool(rawVal)
		case models.TypeByteValue, models.TypeSmallIntegerValue, models.TypeInteger32Value, models.TypeIntegerValue, models.TypeBigIntegerValue:
			rawVal, rawErr := utils.ConvertRawToInt64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetInt(rawVal)
		case models.TypePositiveByteValue, models.TypePositiveSmallIntegerValue, models.TypePositiveInteger32Value, models.TypePositiveIntegerValue, models.TypePositiveBigIntegerValue:
			rawVal, rawErr := utils.ConvertRawToUint64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetUint(rawVal)
		case models.TypeFloatValue, models.TypeDoubleValue:
			rawVal, rawErr := utils.ConvertRawToFloat64(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetFloat(rawVal)
		case models.TypeStringValue:
			rawVal, rawErr := utils.ConvertRawToString(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.SetString(rawVal)
		case models.TypeDateTimeValue:
			rawVal, rawErr := utils.ConvertRawToDateTime(initVal)
			if rawErr != nil {
				err = rawErr
				return
			}
			tVal.Set(reflect.ValueOf(rawVal))
		default:
			rInitVal := reflect.Indirect(reflect.ValueOf(initVal))
			if rInitVal.Type() != tVal.Type() {
				err = cd.NewError(cd.Unexpected, "missmatch value type")
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

func (s *TypeImpl) Elem() models.Type {
	tType := s.getRawType()
	if tType.Kind() == reflect.Slice {
		return &TypeImpl{typeVal: tType.Elem()}
	}

	return &TypeImpl{typeVal: s.typeVal}
}

func (s *TypeImpl) IsBasic() bool {
	elemType := s.Elem()

	return models.IsBasicType(elemType.GetValue())
}

func (s *TypeImpl) IsStruct() bool {
	elemType := s.Elem()
	return models.IsStructType(elemType.GetValue())
}

func (s *TypeImpl) IsSlice() bool {
	return models.IsSliceType(s.GetValue())
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
