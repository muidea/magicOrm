package helper

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

// EncodeFloatValue get float value str
func EncodeFloatValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	switch rawVal.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", rawVal.Float())
	default:
		err = fmt.Errorf("illegal value, type:%s", rawVal.Type().String())
	}

	return
}

// DecodeFloatValue decode float from string
func DecodeFloatValue(val string, vType model.Type) (ret reflect.Value, err error) {
	ret = reflect.Indirect(vType.Interface())
	switch vType.GetType().Kind() {
	case reflect.Float32:
		fVal, fErr := strconv.ParseFloat(val, 32)
		if fErr != nil {
			err = fErr
			return
		}
		ret.SetFloat(fVal)
	case reflect.Float64:
		fVal, fErr := strconv.ParseFloat(val, 64)
		if fErr != nil {
			err = fErr
			return
		}
		ret.SetFloat(fVal)
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
