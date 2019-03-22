package remote

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// ItemType ItemType
type ItemType struct {
	Name    string    `json:"name"`
	Value   int       `json:"value"`
	PkgPath string    `json:"pkgPath"`
	IsPtr   bool      `json:"isPtr"`
	Depend  *ItemType `json:"depend"`
}

// GetName GetName
func (s *ItemType) GetName() (ret string) {
	ret = s.Name
	return
}

// GetValue GetValue
func (s *ItemType) GetValue() (ret int) {
	ret = s.Value
	return
}

// GetPkgPath GetPkgPath
func (s *ItemType) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

// GetType GetType
func (s *ItemType) GetType() (ret reflect.Type) {
	switch s.Value {
	case util.TypeBooleanField:
		var val bool
		ret = reflect.TypeOf(val)
	case util.TypeStringField:
		var val string
		ret = reflect.TypeOf(val)
	case util.TypeDateTimeField:
		var val time.Time
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
		var val map[string]interface{}
		ret = reflect.TypeOf(val)
	case util.TypeSliceField:
		var val []interface{}
		ret = reflect.TypeOf(val)
	default:
		log.Fatalf("unexpect item type, name:%s, type:%d", s.Name, s.Value)
	}

	return
}

// IsPtrType IsPtrType
func (s *ItemType) IsPtrType() (ret bool) {
	ret = s.IsPtr
	return
}

// Interface Interface
func (s *ItemType) Interface() reflect.Value {
	rawType := s.GetType()
	val := reflect.New(rawType)
	if !s.IsPtrType() {
		val = val.Elem()
	}

	return val
}

// Elem GetDepend
func (s *ItemType) Elem() model.Type {
	if s.Depend != nil {
		return s.Depend
	}

	return nil
}

func (s *ItemType) String() (ret string) {
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s,isPtr:%v", s.GetValue(), s.GetName(), s.GetPkgPath(), s.IsPtrType())
}

// Copy Copy
func (s *ItemType) Copy() (ret *ItemType) {
	ret = &ItemType{Name: s.Name, Value: s.Value, PkgPath: s.PkgPath, IsPtr: s.IsPtr, Depend: s.Depend}
	return
}

// GetType GetType
func GetType(itemType reflect.Type, cache Cache) (ret *ItemType, err error) {
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

	ret = &ItemType{Name: itemType.Name(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
	if isPtr {
		return
	}

	if util.IsStructType(typeVal) {
		ret.Depend = &ItemType{Name: itemType.Name(), Value: typeVal, PkgPath: itemType.PkgPath(), IsPtr: isPtr}
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
			ret.Depend = &ItemType{Name: sliceType.Name(), Value: typeVal, PkgPath: sliceType.PkgPath(), IsPtr: slicePtr}
		}

		return
	}

	return
}
