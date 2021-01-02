package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
	"strconv"
)

// encodeFloatValue get float value str
func (s *impl) encodeFloatValue(vVal model.Value) (ret string, err error) {
	val := vVal.Get().(reflect.Value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", val.Float())
	default:
		err = fmt.Errorf("illegal float value, type:%s", val.Type().String())
	}

	return
}

// decodeFloatValue decode float from string
func (s *impl) decodeFloatValue(val string) (ret model.Value, err error) {
	fVal, fErr := strconv.ParseFloat(val, 64)
	if fErr != nil {
		err = fErr
		return
	}

	ret = s.getValue(reflect.ValueOf(&fVal).Elem())
	return
}
