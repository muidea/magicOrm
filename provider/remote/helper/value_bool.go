package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeBool get bool value str
func (s *impl) encodeBool(vVal model.Value) (ret string, err error) {
	val := reflect.ValueOf(vVal.Get())
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		if val.Uint() > 0 {
			ret = "1"
		} else {
			ret = "0"
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		if val.Int() > 0 {
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
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		bVal = rVal.Uint() > 0
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		bVal = rVal.Int() > 0
	default:
		err = fmt.Errorf("illegal boolean value, val:%v", val)
	}

	if err != nil {
		return
	}

	ret, err = s.getValue(bVal)
	return
}
