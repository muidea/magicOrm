package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// TypeImpl TypeImpl
type TypeImpl struct {
	Name     string    `json:"name"`
	Value    int       `json:"value"`
	PkgPath  string    `json:"pkgPath"`
	IsPtr    bool      `json:"isPtr"`
	ElemType *TypeImpl `json:"elemType"`
}

// newType new typeImpl
func newType(itemType reflect.Type) (ret *TypeImpl, err error) {
	isPtr := false
	if itemType.Kind() == reflect.Ptr {
		isPtr = true
		itemType = itemType.Elem()
	}

	typeVal, typeErr := util.GetTypeEnum(itemType)
	if typeErr != nil {
		err = typeErr
		return
	}

	if util.IsSliceType(typeVal) {
		sliceType := itemType.Elem()
		slicePtr := false
		if sliceType.Kind() == reflect.Ptr {
			sliceType = sliceType.Elem()
			slicePtr = true
		}
		ret = &TypeImpl{Name: sliceType.String(), Value: typeVal, PkgPath: sliceType.PkgPath(), IsPtr: isPtr}

		sliceVal, sliceErr := util.GetTypeEnum(sliceType)
		if sliceErr != nil {
			err = sliceErr
			return
		}
		if util.IsSliceType(sliceVal) {
			err = fmt.Errorf("illegal slice type, type:%s", sliceType.String())
			return
		}

		ret.ElemType = &TypeImpl{Name: sliceType.String(), Value: sliceVal, PkgPath: sliceType.PkgPath(), IsPtr: slicePtr}
		return
	}

	ret = &TypeImpl{Name: itemType.String(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
	ret.ElemType = &TypeImpl{Name: itemType.String(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
	return
}

// GetName GetName
func (s *TypeImpl) GetName() (ret string) {
	ret = s.Name
	return
}

// GetEntityValue GetEntityValue
func (s *TypeImpl) GetValue() (ret int) {
	ret = s.Value
	return
}

// GetPkgPath GetPkgPath
func (s *TypeImpl) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

// IsPtrType IsPtrType
func (s *TypeImpl) IsPtrType() (ret bool) {
	ret = s.IsPtr
	return
}

// Interface Interface
func (s *TypeImpl) Interface() (ret model.Value, err error) {
	tVal, tErr := getInitializeValue(s)
	if tErr != nil {
		err = tErr
		return
	}

	ret = newValue(tVal)
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
		return util.IsBasicType(s.ElemType.Value)
	}

	return util.IsBasicType(s.Value)
}

func (s *TypeImpl) copy() (ret *TypeImpl) {
	ret = &TypeImpl{Name: s.Name, Value: s.Value, PkgPath: s.PkgPath, IsPtr: s.IsPtr}
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
