package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

const (
	falseVal = int8(0)
	trueVal  = int8(1)
)

// encodeBool encode bool value to int8
// json value to database value bool -> int8
func (s *impl) encodeBool(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	// from json value
	case reflect.Bool:
		if val.Bool() {
			ret = trueVal
		} else {
			ret = falseVal
		}
		/*
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
				if val.Int() > 0 {
					ret = trueVal
				} else {
					ret = falseVal
				}
		*/
	default:
		err = fmt.Errorf("illegal boolean value, type:%s", val.Type().String())
	}

	return
}

// decodeBool decode bool from int8 or bool
// database value to json value int8 -> bool
// json value to json value bool -> bool
func (s *impl) decodeBool(val interface{}, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface()
	switch val.(type) {
	case int8:
		tVal.Get().SetBool(val.(int8) > 0)
	case bool:
		tVal.Get().SetBool(val.(bool))
	case string: // only for slice element
		tVal.Get().SetBool(val.(string) == "1")
	default:
		err = fmt.Errorf("illegal boolean value, type:%s", tVal.Get().Type().String())
	}
	if err != nil {
		return
	}
	if tType.IsPtrType() {
		tVal = tVal.Addr()
	}

	ret = tVal
	return
}
