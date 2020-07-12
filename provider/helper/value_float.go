package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// EncodeFloatValue get float value str
func EncodeFloatValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", val.Float())
	default:
		err = fmt.Errorf("illegal value, type:%s", val.Type().String())
	}

	return
}

// DecodeFloatValue decode float from string
func DecodeFloatValue(val string, vType model.Type) (ret reflect.Value, err error) {
	tVal := vType.GetValue()
	switch tVal {
	case util.TypeFloatField, util.TypeDoubleField:
	default:
		err = fmt.Errorf("illegal float value type")
		return
	}

	ret = reflect.Indirect(vType.Interface())
	ret, err = AssignValue(reflect.ValueOf(val), ret)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
