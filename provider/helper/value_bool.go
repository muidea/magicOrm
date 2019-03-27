package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// EncodeBoolValue get bool value str
func EncodeBoolValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.Interface:
		bVal, bOK := val.Interface().(bool)
		if !bOK {
			err = fmt.Errorf("illegal bool value, val:%v", val.Interface())
		} else {
			if bVal {
				ret = "1"
			} else {
				ret = "0"
			}
		}
	default:
		err = fmt.Errorf("illegal value, type:%s", val.Type().String())
	}

	return
}

// DecodeBoolValue decode bool from string
func DecodeBoolValue(val string, vType model.Type) (ret reflect.Value, err error) {
	if vType.GetType().Kind() != reflect.Bool {
		err = fmt.Errorf("unsupport value type, type:%s", vType.GetType().String())
		return
	}

	ret = reflect.Indirect(vType.Interface())
	if val == "1" {
		ret.SetBool(true)
	} else if val == "0" {
		ret.SetBool(false)
	} else {
		err = fmt.Errorf("illegal bool value")
	}

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
