package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeBool get bool value str
func (s *impl) encodeBool(vVal model.Value) (ret string, err error) {
	val := vVal.Get().(reflect.Value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	default:
		err = fmt.Errorf("illegal boolean value, type:%s", val.Type().String())
	}

	return
}

// decodeBool decode bool from string
func (s *impl) decodeBool(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var bVal bool
	switch rVal.Kind() {
	case reflect.String:
		if rVal.String() == "1" {
			bVal = true
		} else {
			bVal = false
		}
	case reflect.Bool:
		bVal = rVal.Bool()
	default:
		err = fmt.Errorf("illegal boolean value, val:%v", val)
	}

	if err != nil {
		return
	}

	ret, err = s.getValue(&bVal)
	return
}
