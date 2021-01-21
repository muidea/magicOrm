package helper

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

// encodeFloat get float value str
func (s *impl) encodeFloat(vVal model.Value) (ret string, err error) {
	val := reflect.ValueOf(vVal.Get())
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	val = reflect.Indirect(val)

	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", val.Float())
	default:
		err = fmt.Errorf("illegal float value, type:%s", val.Type().String())
	}

	return
}

// decodeFloat decode float from string
func (s *impl) decodeFloat(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	var fVal float64
	switch tVal.Kind() {
	case reflect.String:
		fVal, err = strconv.ParseFloat(tVal.String(), 64)
	case reflect.Float32, reflect.Float64:
		fVal = tVal.Float()
	default:
		err = fmt.Errorf("illegal float value, value type:%v", tVal.Type().String())
	}
	if err != nil {
		return
	}

	cVal.SetFloat(fVal)
	ret = cVal
	return
}
