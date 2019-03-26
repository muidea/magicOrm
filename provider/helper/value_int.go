package helper

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

//EncodeIntValue get int value str
func EncodeIntValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	switch rawVal.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%d", int64(rawVal.Float()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = fmt.Sprintf("%d", rawVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = fmt.Sprintf("%d", int64(rawVal.Uint()))
	default:
		err = fmt.Errorf("illegal value type, type:%s", rawVal.Type().String())
	}

	return
}

// DecodeIntValue decode int from string
func DecodeIntValue(val string, vType model.Type) (ret reflect.Value, err error) {
	ret = reflect.Indirect(vType.Interface())
	switch vType.GetType().Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		intVal, intErr := strconv.ParseInt(val, 10, 64)
		if intErr != nil {
			err = intErr
			return
		}
		ret.SetInt(intVal)
	default:
		err = fmt.Errorf("illegal value type")
		return
	}

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}

//EncodeUintValue get uint value str
func EncodeUintValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	switch rawVal.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%d", uint64(rawVal.Float()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = fmt.Sprintf("%d", uint64(rawVal.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = fmt.Sprintf("%d", rawVal.Uint())
	default:
		err = fmt.Errorf("illegal value type, type:%s", rawVal.Type().String())
	}

	return
}

// DecodeUintValue decode uint from string
func DecodeUintValue(val string, vType model.Type) (ret reflect.Value, err error) {
	ret = reflect.Indirect(vType.Interface())
	switch vType.GetType().Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		uintVal, uintErr := strconv.ParseUint(val, 10, 64)
		if uintErr != nil {
			err = uintErr
			return
		}
		ret.SetUint(uintVal)
	default:
		err = fmt.Errorf("illegal value type")
		return
	}

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
