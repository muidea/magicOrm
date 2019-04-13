package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// EncodeStringValue get string value str
func EncodeStringValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		ret = fmt.Sprintf("%s", val.String())
	default:
		err = fmt.Errorf("illegal value, type:%s", val.Type().String())
	}

	return
}

//DecodeStringValue decode string from string
func DecodeStringValue(val string, vType model.Type) (ret reflect.Value, err error) {
	tVal := vType.GetValue()
	switch tVal {
	case util.TypeStringField:
	default:
		err = fmt.Errorf("illegal string value type")
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
