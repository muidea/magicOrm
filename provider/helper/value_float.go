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
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", val.Float())
	case reflect.Interface:
		flt64Val, flt64OK := val.Interface().(float64)
		if flt64OK {
			ret = fmt.Sprintf("%f", flt64Val)
		} else {
			flt32Val, flt32OK := val.Interface().(float32)
			if flt32OK {
				ret = fmt.Sprintf("%f", flt32Val)
			} else {
				err = fmt.Errorf("illegal float value, val:%v", val.Interface())
			}
		}
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
	err = ConvertValue(reflect.ValueOf(val), &ret)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
