package remote

import (
	"fmt"
	"path"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
)

type TypeImpl struct {
	Name        string            `json:"name"`
	PkgPath     string            `json:"pkgPath"`
	Description string            `json:"description"`
	Value       model.TypeDeclare `json:"value"`
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

func (s *TypeImpl) GetDescription() (ret string) {
	ret = s.Description
	return
}

func (s *TypeImpl) GetValue() (ret model.TypeDeclare) {
	ret = s.Value
	return
}

func (s *TypeImpl) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *TypeImpl) IsPtrType() (ret bool) {
	ret = s.IsPtr
	return
}

func (s *TypeImpl) Interface(initVal any) (ret model.Value, err *cd.Result) {
	if initVal != nil {
		switch s.GetValue() {
		case model.TypeBooleanValue:
			rawVal, rawErr := util.GetBool(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetBool initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeBitValue:
			rawVal, rawErr := util.GetInt8(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt8 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeSmallIntegerValue:
			rawVal, rawErr := util.GetInt16(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt16 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeInteger32Value:
			rawVal, rawErr := util.GetInt32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeIntegerValue:
			rawVal, rawErr := util.GetInt(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeBigIntegerValue:
			rawVal, rawErr := util.GetInt64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetInt64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypePositiveBitValue:
			rawVal, rawErr := util.GetUint8(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetUint8 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypePositiveSmallIntegerValue:
			rawVal, rawErr := util.GetUint16(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetUint16 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypePositiveInteger32Value:
			rawVal, rawErr := util.GetUint32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetUint32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypePositiveIntegerValue:
			rawVal, rawErr := util.GetUint(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetUint initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypePositiveBigIntegerValue:
			rawVal, rawErr := util.GetUint64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetUint64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeFloatValue:
			rawVal, rawErr := util.GetFloat32(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetFloat32 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeDoubleValue:
			rawVal, rawErr := util.GetFloat64(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetFloat64 initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeDateTimeValue, model.TypeStringValue:
			rawVal, rawErr := util.GetString(initVal)
			if rawErr != nil {
				err = rawErr
				log.Errorf("Interface failed, util.GetString initVal:%+v, error:%s", initVal, err.Error())
				return
			}
			initVal = rawVal.Value()
		case model.TypeSliceValue:
			if !s.Elem().IsBasic() {
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

func (s *TypeImpl) IsBasic() bool {
	if s.ElemType != nil {
		return model.IsBasicType(s.ElemType.Value)
	}

	return model.IsBasicType(s.Value)
}

func (s *TypeImpl) IsStruct() bool {
	if s.ElemType != nil {
		return model.IsStructType(s.ElemType.Value)
	}

	return model.IsStructType(s.Value)
}

func (s *TypeImpl) IsSlice() bool {
	return model.IsSliceType(s.Value)
}

func (s *TypeImpl) copy() (ret *TypeImpl) {
	ret = &TypeImpl{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: s.Description,
		Value:       s.Value,
		IsPtr:       s.IsPtr,
	}
	if s.ElemType != nil {
		ret.ElemType = s.ElemType.copy()
	}

	return
}

func (s *TypeImpl) dump() string {
	val := s.GetValue()
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", val, s.GetName(), s.GetPkgPath(), s.IsPtrType())
}

func compareType(l, r *TypeImpl) bool {
	if l.Name != r.Name {
		return false
	}
	if l.Value != r.Value {
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
