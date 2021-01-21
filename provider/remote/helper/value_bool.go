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
func (s *impl) decodeBool(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	var bVal bool
	switch tVal.Kind() {
	case reflect.String:
		if tVal.String() == "1" {
			bVal = true
		} else {
			bVal = false
		}
	case reflect.Bool:
		bVal = tVal.Bool()
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		bVal = tVal.Uint() > 0
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		bVal = tVal.Int() > 0
	default:
		err = fmt.Errorf("illegal boolean value, value type:%v", tVal.Type().String())
	}

	if err != nil {
		return
	}

	cVal = reflect.Indirect(cVal)
	cVal.SetBool(bVal)
	if tType.IsPtrType() {
		cVal = cVal.Addr()
	}
	ret = cVal
	return
}
