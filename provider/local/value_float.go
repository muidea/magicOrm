package local

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

// encodeFloatValue get float value str
func encodeFloatValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%f", rawVal.Float())

	return
}

func decodeFloatValue(val string, vType model.Type) (ret reflect.Value, err error) {
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
