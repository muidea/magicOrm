package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

type valueImpl struct {
	valueImpl reflect.Value
}

func newValue(val reflect.Value) (ret *valueImpl, err error) {
	ret = &valueImpl{valueImpl: val}
	return
}

func (s *valueImpl) IsNil() (ret bool) {
	ret = s.valueImpl.Kind() == reflect.Invalid

	return
}

func (s *valueImpl) Set(val reflect.Value) (err error) {
	if s.valueImpl.Kind() == reflect.Invalid {
		s.valueImpl = val
		return
	}

	valTypeName := val.Type().String()
	expectTypeName := s.valueImpl.Type().String()
	if expectTypeName != valTypeName {
		err = fmt.Errorf("illegal value type, type:%s, expect:%s", expectTypeName, valTypeName)
		return
	}

	s.valueImpl.Set(val)
	return
}

func (s *valueImpl) Get() (ret reflect.Value) {
	ret = s.valueImpl

	return
}

func (s *valueImpl) Copy() (ret *valueImpl) {
	ret = &valueImpl{valueImpl: s.valueImpl}

	return
}

// GetValueStr get value str
func GetValueStr(vType model.Type, vVal model.Value, cache Cache) (ret string, err error) {
	rawType := vType.GetType()
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
	}

	switch rawType.Kind() {
	case reflect.Bool:
		ret, err = getBoolValueStr(vVal.Get())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret, err = getIntValueStr(vVal.Get())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret, err = getUintValueStr(vVal.Get())
	case reflect.Float32, reflect.Float64:
		ret, err = getFloatValueStr(vVal.Get())
	case reflect.String:
		ret, err = getStringValueStr(vVal.Get())
	case reflect.Slice:
		ret, err = getSliceValueStr(vVal.Get())
	case reflect.Struct:
		ret, err = getDateTimeValueStr(vVal.Get())
	default:
		err = fmt.Errorf("illegal value kind, kind:%v", rawType.Kind())
	}
	return
}
