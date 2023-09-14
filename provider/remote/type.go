package remote

import (
	"fmt"
	"github.com/muidea/magicOrm/provider/util"
	"path"

	"github.com/muidea/magicOrm/model"
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

func (s *TypeImpl) Interface(initVal any) (ret model.Value, err error) {
	if initVal != nil {
		switch s.GetValue() {
		case model.TypeBooleanValue:
			initVal, err = util.GetBool(initVal)
		case model.TypeBitValue, model.TypeSmallIntegerValue, model.TypeInteger32Value, model.TypeIntegerValue, model.TypeBigIntegerValue:
			initVal, err = util.GetInt(initVal)
		case model.TypePositiveBitValue, model.TypePositiveSmallIntegerValue, model.TypePositiveInteger32Value, model.TypePositiveIntegerValue, model.TypePositiveBigIntegerValue:
			initVal, err = util.GetUint(initVal)
		case model.TypeFloatValue, model.TypeDoubleValue:
			initVal, err = util.GetFloat(initVal)
		case model.TypeStringValue:
			initVal, err = util.GetString(initVal)
		case model.TypeDateTimeValue:
			initVal, err = util.GetDateTimeStr(initVal)
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
