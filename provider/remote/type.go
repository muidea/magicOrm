package remote

import (
	"path"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

type TypeImpl struct {
	Name        string             `json:"name"`
	PkgPath     string             `json:"pkgPath"`
	Description string             `json:"description"`
	Value       models.TypeDeclare `json:"-"`
	IsPtr       bool               `json:"isPtr"`
	ElemType    *TypeImpl          `json:"elemType"`
}

func (s *TypeImpl) GetName() (ret string) {
	ret = s.Name
	return
}

func (s *TypeImpl) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

func (s *TypeImpl) GetPkgKey() (ret string) {
	ret = path.Join(s.PkgPath, s.Name)
	return
}

func (s *TypeImpl) GetDescription() (ret string) {
	ret = s.Description
	return
}

func (s *TypeImpl) GetValue() (ret models.TypeDeclare) {
	if s.Value == 0 {
		// 由于Value字段是会序列化，如果当前值为0，则需要重新根据name，pkgPath及ElemType重新计算
		s.Value = s.validateValue()
	}

	ret = s.Value
	return
}

func (s *TypeImpl) validateValue() (ret models.TypeDeclare) {
	tVal := models.GetTypeValue(s.Name)
	if s.ElemType == nil {
		ret = tVal
		return
	}

	ret = models.TypeSliceValue
	return
}

func (s *TypeImpl) IsPtrType() (ret bool) {
	ret = s.IsPtr
	return
}

func (s *TypeImpl) Interface(initVal any) (ret models.Value, err *cd.Error) {
	if initVal != nil {
		switch s.GetValue() {
		case models.TypeBooleanValue:
			rawVal, rawErr := utils.ConvertRawToBool(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetBool initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeByteValue:
			rawVal, rawErr := utils.ConvertRawToInt8(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt8 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeSmallIntegerValue:
			rawVal, rawErr := utils.ConvertRawToInt16(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt16 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeInteger32Value:
			rawVal, rawErr := utils.ConvertRawToInt32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeIntegerValue:
			rawVal, rawErr := utils.ConvertRawToInt(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeBigIntegerValue:
			rawVal, rawErr := utils.ConvertRawToInt64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypePositiveByteValue:
			rawVal, rawErr := utils.ConvertRawToUint8(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint8 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypePositiveSmallIntegerValue:
			rawVal, rawErr := utils.ConvertRawToUint16(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint16 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypePositiveInteger32Value:
			rawVal, rawErr := utils.ConvertRawToUint32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypePositiveIntegerValue:
			rawVal, rawErr := utils.ConvertRawToUint(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypePositiveBigIntegerValue:
			rawVal, rawErr := utils.ConvertRawToUint64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeFloatValue:
			rawVal, rawErr := utils.ConvertRawToFloat32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetFloat32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeDoubleValue:
			rawVal, rawErr := utils.ConvertRawToFloat64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetFloat64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeDateTimeValue, models.TypeStringValue:
			rawVal, rawErr := utils.ConvertRawToString(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetString initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case models.TypeSliceValue:
			if s.Elem().GetValue().IsBasicType() {
				rawVal, rawErr := s.convertRawBasicToSlice(initVal)
				if rawErr != nil {
					err = rawErr
					log.Errorf("Interface failed, convertRawToSlice initVal:%+v, error:%s", initVal, err.Error())
					return
				}
				initVal = rawVal
			} else {
				rawVal, rawErr := s.convertRawStructToSlice(initVal)
				if rawErr != nil {
					err = rawErr
					log.Errorf("Interface failed, convertRawStructToSlice initVal:%+v, error:%s", initVal, err.Error())
					return
				}
				initVal = rawVal
			}
		case models.TypeStructValue:
			rawVal, rawErr := s.convertRawStruct(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, convertRawStruct initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		default:
			initVal = nil
		}
	}
	if initVal != nil {
		ret = NewValue(initVal)
		return
	}

	ret = NewValue(getInitializeValue(s))
	return
}

func (s *TypeImpl) convertRawBasicToSlice(initVal any) (ret any, err *cd.Error) {
	rVal := reflect.ValueOf(initVal)
	rVal = reflect.Indirect(rVal)
	if rVal.Kind() != reflect.Slice {
		err = cd.NewError(cd.Unexpected, "value is not slice")
		log.Warnf("convertRawSlice failed, value is not slice")
		return
	}

	sliceVal := getSliceInitValue(s)
	rSliceVal := reflect.ValueOf(sliceVal)
	for idx := 0; idx < rVal.Len(); idx++ {
		val := rVal.Index(idx)
		vVal, vErr := s.Elem().Interface(val.Interface())
		if vErr != nil {
			err = vErr
			return
		}
		rSliceVal = reflect.Append(rSliceVal, reflect.ValueOf(vVal.Get()))
	}

	ret = rSliceVal.Interface()
	return
}

func (s *TypeImpl) convertRawStructToSlice(initVal any) (ret *SliceObjectValue, err *cd.Error) {
	rVal := reflect.ValueOf(initVal)
	rVal = reflect.Indirect(rVal)
	sliceVal := getSliceStructInitValue(s)
	switch rVal.Kind() {
	case reflect.Slice, reflect.Array:
		for idx := 0; idx < rVal.Len(); idx++ {
			val := rVal.Index(idx)
			vVal, vErr := s.ElemType.convertRawStruct(val.Interface())
			if vErr != nil {
				err = vErr
				return
			}
			sliceVal.Values = append(sliceVal.Values, vVal)
		}
	case reflect.Struct:
		initSliceObjectValPtr, initSliceObjectOK := initVal.(*SliceObjectValue)
		if initSliceObjectOK {
			sliceVal.Values = append(sliceVal.Values, initSliceObjectValPtr.Values...)
		} else {
			initSliceObjectVal, initSliceObjectOK := initVal.(SliceObjectValue)
			if initSliceObjectOK {
				sliceVal.Values = append(sliceVal.Values, initSliceObjectVal.Values...)
			} else {
				err = cd.NewError(cd.Unexpected, "value is not slice")
			}
		}
	default:
		err = cd.NewError(cd.Unexpected, "value is not slice")
	}

	if err == nil {
		ret = sliceVal
	}
	return
}

func (s *TypeImpl) convertRawStruct(initVal any) (ret *ObjectValue, err *cd.Error) {
	rVal := reflect.ValueOf(initVal)
	rVal = reflect.Indirect(rVal)
	if rVal.Kind() != reflect.Struct && rVal.Kind() != reflect.Map {
		log.Warnf("convertRawStruct failed, value is not struct or map")
		return
	}
	objectStructVal := getStructInitValue(s)
	objectVal := objectStructVal

	switch rVal.Kind() {
	case reflect.Struct:
		initObjectValPtr, initObjectOK := initVal.(*ObjectValue)
		if initObjectOK {
			// 将initObjectValue的所有FieldValue复制到objectStructPtr
			for _, field := range initObjectValPtr.Fields {
				objectVal.SetFieldValue(field.Name, field.Value)
			}
		} else {
			initObjectVal, initObjectOK := initVal.(ObjectValue)
			if initObjectOK {
				// 将initObjectValue的所有FieldValue复制到objectStructPtr
				for _, field := range initObjectVal.Fields {
					objectVal.SetFieldValue(field.Name, field.Value)
				}
			} else {
				err = cd.NewError(cd.IllegalParam, "init value is not a struct")
			}
		}
	case reflect.Map:
		// 遍历Map, 将转换成ObjectValue的FieldValue
		for _, key := range rVal.MapKeys() {
			fieldValue := rVal.MapIndex(key)
			objectVal.SetFieldValue(key.String(), fieldValue.Interface())
		}
	default:
		err = cd.NewError(cd.Unexpected, "value is not struct or map")
		log.Warnf("convertRawStruct failed, value is not struct or map")
		return
	}

	ret = objectVal
	return
}

// Elem get element type
func (s *TypeImpl) Elem() models.Type {
	var eType TypeImpl
	if s.ElemType == nil {
		eType = *s
	} else {
		eType = *s.ElemType
	}

	return &eType
}

func (s *TypeImpl) Copy() (ret *TypeImpl) {
	ret = &TypeImpl{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: s.Description,
		Value:       s.Value,
		IsPtr:       s.IsPtr,
	}
	if s.ElemType != nil {
		ret.ElemType = s.ElemType.Copy()
	}

	return
}

func compareType(l, r *TypeImpl) bool {
	if l.Name != r.Name {
		return false
	}
	if l.GetValue() != r.GetValue() {
		return false
	}
	if l.PkgPath != r.PkgPath {
		return false
	}
	if l.IsPtr != r.IsPtr {
		return false
	}

	if l.ElemType != nil && r.ElemType == nil {
		return false
	}

	if l.ElemType == nil && r.ElemType != nil {
		return false
	}

	if l.ElemType == nil && r.ElemType == nil {
		return true
	}

	return compareType(l.ElemType, r.ElemType)
}
