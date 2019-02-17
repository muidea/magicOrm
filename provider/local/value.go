package local

import (
	"fmt"
	"reflect"
)

type valueImpl struct {
	valueImpl reflect.Value
}

func newFieldValue(val reflect.Value) (ret *valueImpl, err error) {
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

func (s *valueImpl) Str() (ret string, err error) {
	rawVal := reflect.Indirect(s.valueImpl)
	switch rawVal.Kind() {
	case reflect.Bool:
		ret, err = GetBoolValueStr(rawVal)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret, err = GetIntValueStr(rawVal)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret, err = GetUintValueStr(rawVal)
	case reflect.Float32, reflect.Float64:
		ret, err = GetFloatValueStr(rawVal)
	case reflect.String:
		ret, err = GetStringValueStr(rawVal)
	case reflect.Slice:
		ret, err = GetSliceValueStr(rawVal)
	case reflect.Struct:
		ret, err = GetDateTimeValueStr(rawVal)
	default:
		err = fmt.Errorf("illegal value kind, kind:%v", rawVal.Kind())
	}
	return
}

func (s *valueImpl) Copy() (ret *valueImpl) {
	ret = &valueImpl{valueImpl: s.valueImpl}

	return
}

func (s *valueImpl) Dump() (ret string) {
	str, err := s.Str()
	if err != nil {
		ret = "invalid"
	} else {
		ret = str
	}

	return
}
