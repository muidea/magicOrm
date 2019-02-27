package local

import (
	"fmt"
	"reflect"
	"strconv"

	"muidea.com/magicOrm/model"
)

//encodeIntValue get int value str
func encodeIntValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%d", rawVal.Int())

	return
}

func decodeIntValue(val string, vType model.Type) (ret reflect.Value, err error) {
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

//encodeUintValue get uint value str
func encodeUintValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%d", rawVal.Uint())

	return
}

func decodeUintValue(val string, vType model.Type) (ret reflect.Value, err error) {
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
