package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
	"strconv"
)

// encodeFloat get float value str
func (s *impl) encodeFloat(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = val.Float()
	default:
		err = fmt.Errorf("illegal float value, type:%s", val.Type().String())
	}

	return
}

// decodeFloat decode float from string
func (s *impl) decodeFloat(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var fVal float64
	switch rVal.Kind() {
	case reflect.Float32, reflect.Float64:
		fVal = rVal.Float()
	case reflect.String:
		fVal, err = strconv.ParseFloat(rVal.String(), 64)
	default:
		err = fmt.Errorf("illegal float value, val:%v", val)
	}
	if err != nil {
		return
	}

	tVal, _ := tType.Interface()
	switch tVal.Get().Kind() {
	case reflect.Float32, reflect.Float64:
		tVal.Get().SetFloat(fVal)
	default:
		err = fmt.Errorf("illegal float value, type:%s", tVal.Get().Type().String())
	}
	if err != nil {
		return
	}

	ret = tVal
	return
}
