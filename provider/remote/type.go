package remote

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
	"log"
	"reflect"
)

var _declareObjectValue ObjectValue
var _declareObjectSliceValue SliceObjectValue

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
	tType := s.getType()
	tVal := reflect.New(tType).Elem()
	if s.IsBasic() {
		if s.IsPtrType() {
			tVal = tVal.Addr()
		}

		ret = newValue(tVal.Interface())
		return
	}

	if util.IsStructType(s.Value) {
		tVal.FieldByName("Name").SetString(s.Name)
		tVal.FieldByName("PkgPath").SetString(s.PkgPath)
		tVal.FieldByName("IsPtr").SetBool(s.IsPtr)

		ret = newValue(tVal.Addr().Interface())
		return
	}

	tVal.FieldByName("Name").SetString(s.ElemType.Name)
	tVal.FieldByName("PkgPath").SetString(s.ElemType.PkgPath)
	tVal.FieldByName("IsPtr").SetBool(s.ElemType.IsPtr)

	ret = newValue(tVal.Addr().Interface())
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

var declareSliceObject SliceObjectValue
var declareObject ObjectValue

func (s *TypeImpl) getSliceType() (ret reflect.Type) {
	switch s.ElemType.GetValue() {
	case util.TypeBooleanField:
		var val []bool
		ret = reflect.TypeOf(val)
	case util.TypeStringField,
		util.TypeDateTimeField:
		var val []string
		ret = reflect.TypeOf(val)
	case util.TypeBitField:
		var val []int8
		ret = reflect.TypeOf(val)
	case util.TypeSmallIntegerField:
		var val []int16
		ret = reflect.TypeOf(val)
	case util.TypeInteger32Field:
		var val []int32
		ret = reflect.TypeOf(val)
	case util.TypeIntegerField:
		var val []int
		ret = reflect.TypeOf(val)
	case util.TypeBigIntegerField:
		var val []int64
		ret = reflect.TypeOf(val)
	case util.TypePositiveBitField:
		var val []uint8
		ret = reflect.TypeOf(val)
	case util.TypePositiveSmallIntegerField:
		var val []uint16
		ret = reflect.TypeOf(val)
	case util.TypePositiveInteger32Field:
		var val []uint32
		ret = reflect.TypeOf(val)
	case util.TypePositiveIntegerField:
		var val []uint
		ret = reflect.TypeOf(val)
	case util.TypePositiveBigIntegerField:
		var val []uint64
		ret = reflect.TypeOf(val)
	case util.TypeFloatField:
		var val []float32
		ret = reflect.TypeOf(val)
	case util.TypeDoubleField:
		var val []float64
		ret = reflect.TypeOf(val)
	case util.TypeStructField:
		ret = reflect.TypeOf(_declareObjectSliceValue)
	default:
		log.Fatalf("unexpect slice item type, name:%s, pkgPath:%s, type:%d", s.ElemType.GetName(), s.ElemType.GetPkgPath(), s.ElemType.GetValue())
	}

	return
}

// getType GetEntityType
func (s *TypeImpl) getType() (ret reflect.Type) {
	switch s.GetValue() {
	case util.TypeBooleanField:
		var val bool
		ret = reflect.TypeOf(val)
	case util.TypeStringField,
		util.TypeDateTimeField:
		var val string
		ret = reflect.TypeOf(val)
	case util.TypeBitField:
		var val int8
		ret = reflect.TypeOf(val)
	case util.TypeSmallIntegerField:
		var val int16
		ret = reflect.TypeOf(val)
	case util.TypeInteger32Field:
		var val int32
		ret = reflect.TypeOf(val)
	case util.TypeIntegerField:
		var val int
		ret = reflect.TypeOf(val)
	case util.TypeBigIntegerField:
		var val int64
		ret = reflect.TypeOf(val)
	case util.TypePositiveBitField:
		var val uint8
		ret = reflect.TypeOf(val)
	case util.TypePositiveSmallIntegerField:
		var val uint16
		ret = reflect.TypeOf(val)
	case util.TypePositiveInteger32Field:
		var val uint32
		ret = reflect.TypeOf(val)
	case util.TypePositiveIntegerField:
		var val uint
		ret = reflect.TypeOf(val)
	case util.TypePositiveBigIntegerField:
		var val uint64
		ret = reflect.TypeOf(val)
	case util.TypeFloatField:
		var val float32
		ret = reflect.TypeOf(val)
	case util.TypeDoubleField:
		var val float64
		ret = reflect.TypeOf(val)
	case util.TypeStructField:
		ret = reflect.TypeOf(_declareObjectValue)
	case util.TypeSliceField:
		ret = s.getSliceType()
	default:
		log.Fatalf("unexpect item type, name:%s, type:%d", s.Name, s.Value)
	}

	return
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
