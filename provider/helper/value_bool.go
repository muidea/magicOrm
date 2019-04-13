package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// EncodeBoolValue get bool value str
func EncodeBoolValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	default:
		err = fmt.Errorf("illegal value, type:%s", val.Type().String())
	}

	return
}

// DecodeBoolValue decode bool from string
func DecodeBoolValue(val string, vType model.Type) (ret reflect.Value, err error) {
	tVal := vType.GetValue()
	switch tVal {
	case util.TypeBooleanField:
	default:
		err = fmt.Errorf("illegal bool value type")
		return
	}

	ret = reflect.Indirect(vType.Interface())
	err = ConvertValue(reflect.ValueOf(val), &ret)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
