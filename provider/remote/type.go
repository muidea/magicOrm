package remote

import (
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// TypeImpl TypeImpl
type TypeImpl struct {
	Name       string    `json:"name"`
	Value      int       `json:"value"`
	PkgPath    string    `json:"pkgPath"`
	IsPtr      bool      `json:"isPtr"`
	DependType *TypeImpl `json:"depend"`
}

// GetName GetName
func (s *TypeImpl) GetName() (ret string) {
	ret = s.Name
	return
}

// GetValue GetValue
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
func (s *TypeImpl) Interface() reflect.Value {
	if util.IsBasicType(s.Value) {
		rawType := s.getType()
		val := reflect.New(rawType)
		if !s.IsPtrType() {
			val = val.Elem()
		}

		return val
	}

	if util.IsStructType(s.Value) {
		val := &ObjectValue{Name: s.Name, PkgPath: s.PkgPath, Items: []*ItemValue{}}
		return reflect.ValueOf(val)
	}

	elemVal := s.DependType.GetValue()
	if util.IsBasicType(elemVal) {
		rawType := s.getType()
		val := reflect.New(rawType)
		if !s.IsPtrType() {
			val = val.Elem()
		}

		return val
	}

	var val []*ObjectValue
	return reflect.ValueOf(val)
}

// Depend get depend type
func (s *TypeImpl) Depend() model.Type {
	if s.DependType != nil {
		return s.DependType
	}

	return nil
}

// Elem get slice element type
func (s *TypeImpl) Elem() model.Type {
	if s.Value == util.TypeSliceField && s.DependType != nil {
		return s.DependType
	}

	return nil
}

// Copy Copy
func (s *TypeImpl) Copy() (ret *TypeImpl) {
	ret = &TypeImpl{Name: s.Name, Value: s.Value, PkgPath: s.PkgPath, IsPtr: s.IsPtr}
	if s.DependType != nil {
		ret.DependType = s.DependType.Copy()
	}

	return
}

func (s *TypeImpl) getSliceType() (ret reflect.Type) {
	vType := s.DependType
	switch vType.GetValue() {
	case util.TypeBooleanField:
		var val []bool
		ret = reflect.TypeOf(val)
	case util.TypeStringField,
		util.TypeDateTimeField:
		var val []string
		ret = reflect.TypeOf(val)
	case util.TypeBitField,
		util.TypeSmallIntegerField,
		util.TypeInteger32Field,
		util.TypeIntegerField,
		util.TypeBigIntegerField,
		util.TypePositiveBitField,
		util.TypePositiveSmallIntegerField,
		util.TypePositiveInteger32Field,
		util.TypePositiveIntegerField,
		util.TypePositiveBigIntegerField,
		util.TypeFloatField,
		util.TypeDoubleField:
		var val []float64
		ret = reflect.TypeOf(val)
	case util.TypeStructField:
		var val []*ObjectValue
		ret = reflect.TypeOf(val)
	default:
		log.Fatalf("unexpect slice item type, name:%s, pkgPath:%s, type:%d", vType.GetName(), vType.GetPkgPath(), vType.GetValue())
	}

	return
}

// getType GetType
func (s *TypeImpl) getType() (ret reflect.Type) {
	switch s.GetValue() {
	case util.TypeBooleanField:
		var val bool
		ret = reflect.TypeOf(val)
	case util.TypeStringField,
		util.TypeDateTimeField:
		var val string
		ret = reflect.TypeOf(val)
	case util.TypeBitField,
		util.TypeSmallIntegerField,
		util.TypeInteger32Field,
		util.TypeIntegerField,
		util.TypeBigIntegerField,
		util.TypePositiveBitField,
		util.TypePositiveSmallIntegerField,
		util.TypePositiveInteger32Field,
		util.TypePositiveIntegerField,
		util.TypePositiveBigIntegerField,
		util.TypeFloatField,
		util.TypeDoubleField:
		var val float64
		ret = reflect.TypeOf(val)
	case util.TypeStructField:
		var val *ObjectValue
		ret = reflect.TypeOf(val)
	case util.TypeSliceField:
		ret = s.getSliceType()
	default:
		log.Fatalf("unexpect item type, name:%s, type:%d", s.Name, s.Value)
	}

	return
}

// newType new typeImpl
func newType(itemType reflect.Type) (ret *TypeImpl, err error) {
	isPtr := false
	if itemType.Kind() == reflect.Ptr {
		isPtr = true
		itemType = itemType.Elem()
	}

	typeVal, typeErr := util.GetTypeValueEnum(itemType)
	if typeErr != nil {
		err = typeErr
		return
	}

	ret = &TypeImpl{Name: itemType.String(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}

	if util.IsStructType(typeVal) {
		ret.DependType = &TypeImpl{Name: itemType.String(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
		return
	}

	if util.IsSliceType(typeVal) {
		slicePtr := false
		sliceType := itemType.Elem()
		if sliceType.Kind() == reflect.Ptr {
			sliceType = sliceType.Elem()
			slicePtr = true
		}
		sliceVal, sliceErr := util.GetTypeValueEnum(sliceType)
		if sliceErr != nil {
			err = sliceErr
			return
		}
		if util.IsSliceType(sliceVal) {
			err = fmt.Errorf("illegal slice type, type:%s", sliceType.String())
			return
		}

		ret.DependType = &TypeImpl{Name: sliceType.String(), Value: sliceVal, PkgPath: sliceType.PkgPath(), IsPtr: slicePtr}
		return
	}

	return
}
