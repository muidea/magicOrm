package remote

import (
	"path"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
)

type TypeImpl struct {
	Name        string            `json:"name"`
	PkgPath     string            `json:"pkgPath"`
	Description string            `json:"description"`
	Value       model.TypeDeclare `json:"-"`
	IsPtr       bool              `json:"isPtr"`
	ElemType    *TypeImpl         `json:"elemType"`
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

func (s *TypeImpl) GetValue() (ret model.TypeDeclare) {
	if s.Value == 0 {
		// 由于Value字段是会序列化，如果当前值为0，则需要重新根据name，pkgPath及ElemType重新计算
		s.Value = s.validateValue()
	}

	ret = s.Value
	return
}

func (s *TypeImpl) validateValue() (ret model.TypeDeclare) {
	tVal := model.GetTypeValue(s.Name)
	if s.ElemType == nil {
		ret = tVal
		return
	}

	ret = model.TypeSliceValue
	return
}

func (s *TypeImpl) IsPtrType() (ret bool) {
	ret = s.IsPtr
	return
}

func (s *TypeImpl) Interface(initVal any) (ret model.Value, err *cd.Error) {
	if initVal != nil {
		switch s.GetValue() {
		case model.TypeBooleanValue:
			rawVal, rawErr := utils.ConvertRawToBool(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetBool initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeByteValue:
			rawVal, rawErr := utils.ConvertRawToInt8(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt8 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeSmallIntegerValue:
			rawVal, rawErr := utils.ConvertRawToInt16(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt16 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeInteger32Value:
			rawVal, rawErr := utils.ConvertRawToInt32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeIntegerValue:
			rawVal, rawErr := utils.ConvertRawToInt(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeBigIntegerValue:
			rawVal, rawErr := utils.ConvertRawToInt64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypePositiveByteValue:
			rawVal, rawErr := utils.ConvertRawToUint8(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint8 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypePositiveSmallIntegerValue:
			rawVal, rawErr := utils.ConvertRawToUint16(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint16 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypePositiveInteger32Value:
			rawVal, rawErr := utils.ConvertRawToUint32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypePositiveIntegerValue:
			rawVal, rawErr := utils.ConvertRawToUint(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypePositiveBigIntegerValue:
			rawVal, rawErr := utils.ConvertRawToUint64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetUint64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeFloatValue:
			rawVal, rawErr := utils.ConvertRawToFloat32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetFloat32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeDoubleValue:
			rawVal, rawErr := utils.ConvertRawToFloat64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetFloat64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeDateTimeValue, model.TypeStringValue:
			rawVal, rawErr := utils.ConvertRawToString(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, utils.GetString initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal
		case model.TypeSliceValue:
			if !model.IsBasic(s.Elem()) {
				initVal = nil
			}
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

// Elem get element type
func (s *TypeImpl) Elem() model.Type {
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
