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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = fmt.Sprintf("%f", float64(rawVal.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = fmt.Sprintf("%f", float64(rawVal.Uint()))
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", rawVal.Float())
	default:
		err = fmt.Errorf("illegal value type, type:%s", rawVal.Type().String())
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
