package helper

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
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
func (s *impl) decodeFloat(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var fVal float64
	switch rVal.Kind() {
	case reflect.String:
		fVal, err = strconv.ParseFloat(rVal.String(), 64)
	case reflect.Float32, reflect.Float64:
		fVal = rVal.Float()
	default:
		err = fmt.Errorf("illegal float value, val:%v", val)
	}
	if err != nil {
		return
	}

	if tType.GetValue() == util.TypeFloatField {
		f32Val := float32(fVal)
		ret, err = s.getValue(f32Val)
		return
	}

	ret, err = s.getValue(fVal)
	return
}
