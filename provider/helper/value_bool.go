package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

const (
	falseVal = iota
	trueVal
)

// encodeBool get bool value str
func (s *impl) encodeBool(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			ret = trueVal
		} else {
			ret = falseVal
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		if val.Int() > 0 {
			ret = trueVal
		} else {
			ret = falseVal
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

	var bVal int64
	switch rVal.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		bVal = rVal.Int()
	case reflect.Float64:
		bVal = int64(rVal.Float())
	case reflect.Bool:
		if rVal.Bool() {
			bVal = 1
		}
	case reflect.String:
		if rVal.String() == "1" {
			bVal = 1
		}
	default:
		err = fmt.Errorf("illegal boolean value, val:%v", val)
	}

	if err != nil {
		return
	}

	tVal, _ := tType.Interface()
	switch tVal.Get().Kind() {
	case reflect.Bool:
		tVal.Get().SetBool(bVal > 0)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		tVal.Get().SetInt(bVal)
	default:
		err = fmt.Errorf("illegal boolean value, type:%s", tVal.Get().Type().String())
	}
	if err != nil {
		return
	}

	ret = tVal
	return
}
