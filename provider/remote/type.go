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
	Name    string    `json:"name"`
	Value   int       `json:"value"`
	PkgPath string    `json:"pkgPath"`
	IsPtr   bool      `json:"isPtr"`
	Depend  *TypeImpl `json:"depend"`
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

// GetType GetType
func (s *TypeImpl) GetType() (ret reflect.Type) {
	switch s.Value {
	case util.TypeBooleanField:
		var val bool
		ret = reflect.TypeOf(val)
	case util.TypeStringField:
		var val string
		ret = reflect.TypeOf(val)
	case util.TypeDateTimeField:
		var val string
		ret = reflect.TypeOf(val)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		var val int64
		ret = reflect.TypeOf(val)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		var val uint64
		ret = reflect.TypeOf(val)
	case util.TypeFloatField, util.TypeDoubleField:
		var val float64
		ret = reflect.TypeOf(val)
	case util.TypeStructField:
		var val ObjectValue
		ret = reflect.TypeOf(val)
	case util.TypeSliceField:
		if s.Depend == nil {
			var val []interface{}
			ret = reflect.TypeOf(val)
		} else {
			var val []ObjectValue
			ret = reflect.TypeOf(val)
		}
	default:
		log.Fatalf("unexpect item type, name:%s, type:%d", s.Name, s.Value)
	}

	return
}

// IsPtrType IsPtrType
func (s *TypeImpl) IsPtrType() (ret bool) {
	ret = s.IsPtr
	return
}

// Interface Interface
func (s *TypeImpl) Interface() reflect.Value {
	rawType := s.GetType()
	val := reflect.New(rawType)
	if !s.IsPtrType() {
		val = val.Elem()
	}

	return val
}

// Elem GetDepend
func (s *TypeImpl) Elem() model.Type {
	if s.Depend != nil {
		return s.Depend
	}

	return nil
}

func (s *TypeImpl) String() (ret string) {
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.GetValue(), s.GetName(), s.GetPkgPath(), s.IsPtrType())
}

// Copy Copy
func (s *TypeImpl) Copy() (ret *TypeImpl) {
	ret = &TypeImpl{Name: s.Name, Value: s.Value, PkgPath: s.PkgPath, IsPtr: s.IsPtr, Depend: s.Depend}
	return
}

// GetType GetType
func GetType(itemType reflect.Type) (ret *TypeImpl, err error) {
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
		ret.Depend = &TypeImpl{Name: itemType.String(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
		return
	}

	if util.IsSliceType(typeVal) {
		slicePtr := false
		sliceType := itemType.Elem()
		if sliceType.Kind() == reflect.Ptr {
			sliceType = sliceType.Elem()
			slicePtr = true
		}
		typeVal, typeErr = util.GetTypeValueEnum(sliceType)
		if typeErr != nil {
			err = typeErr
			return
		}
		if util.IsSliceType(typeVal) {
			err = fmt.Errorf("illegal slice type, type:%s", sliceType.String())
			return
		}

		if util.IsStructType(typeVal) {
			ret.Depend = &TypeImpl{Name: sliceType.String(), Value: typeVal, PkgPath: sliceType.PkgPath(), IsPtr: slicePtr}
		}

		return
	}

	return
}
