package remote

import (
	"fmt"
	"path"

	cd "github.com/muidea/magicCommon/def"

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
			initVal, err = util.GetBool(initVal)
		case model.TypeBitValue:
			initVal, err = util.GetInt8(initVal)
		case model.TypeSmallIntegerValue:
			initVal, err = util.GetInt16(initVal)
		case model.TypeInteger32Value:
			initVal, err = util.GetInt32(initVal)
		case model.TypeIntegerValue:
			initVal, err = util.GetInt(initVal)
		case model.TypeBigIntegerValue:
			initVal, err = util.GetInt64(initVal)
		case model.TypePositiveBitValue:
			initVal, err = util.GetUint8(initVal)
		case model.TypePositiveSmallIntegerValue:
			initVal, err = util.GetUint16(initVal)
		case model.TypePositiveInteger32Value:
			initVal, err = util.GetUint32(initVal)
		case model.TypePositiveIntegerValue:
			initVal, err = util.GetUint(initVal)
		case model.TypePositiveBigIntegerValue:
			initVal, err = util.GetUint64(initVal)
		case model.TypeFloatValue:
			initVal, err = util.GetFloat32(initVal)
		case model.TypeDoubleValue:
			initVal, err = util.GetFloat64(initVal)
		case model.TypeStringValue:
			initVal, err = util.GetString(initVal)
		case model.TypeDateTimeValue:
			initVal, err = util.GetString(initVal)
		default:
			initVal = nil
		}
	}
	if err != nil {
		return
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
