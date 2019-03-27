package helper

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

//EncodeIntValue get int value str
func EncodeIntValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Float64:
		ret = fmt.Sprintf("%d", int64(val.Float()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = fmt.Sprintf("%d", val.Int())
	case reflect.Interface:
		fltVal, fltOK := val.Interface().(float64)
		if fltOK {
			ret = fmt.Sprintf("%d", int64(fltVal))
		} else {
			intVal, intOK := val.Interface().(int64)
			if intOK {
				ret = fmt.Sprintf("%d", intVal)
			} else {
				err = fmt.Errorf("illegal int value, val:%v", val.Interface())
			}
		}
	default:
		err = fmt.Errorf("illegal int value, type:%s", val.Type().String())
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
	case reflect.Float64:
		fltVal, fltErr := strconv.ParseFloat(val, 64)
		if fltErr != nil {
			err = fltErr
			return
		}
		ret.SetFloat(fltVal)
	default:
		err = fmt.Errorf("unsupport value type, type:%s", vType.GetType().String())
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
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Float64:
		ret = fmt.Sprintf("%d", uint64(val.Float()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = fmt.Sprintf("%d", val.Uint())
	case reflect.Interface:
		fltVal, fltOK := val.Interface().(float64)
		if fltOK {
			ret = fmt.Sprintf("%d", uint64(fltVal))
		} else {
			uintVal, uintOK := val.Interface().(uint64)
			if uintOK {
				ret = fmt.Sprintf("%d", uintVal)
			} else {
				err = fmt.Errorf("illegal uint value, val:%v", val.Interface())
			}
		}
	default:
		err = fmt.Errorf("illegal uint value, type:%s", val.Type().String())
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
	case reflect.Float64:
		fltVal, fltErr := strconv.ParseFloat(val, 64)
		if fltErr != nil {
			err = fltErr
			return
		}
		ret.SetFloat(fltVal)
	default:
		err = fmt.Errorf("unsupport value type, type:%s", vType.GetType().String())
		return
	}

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
