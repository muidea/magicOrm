package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeBoolValue get bool value str
func (s *impl) encodeBoolValue(vVal model.Value) (ret string, err error) {
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

// decodeBoolValue decode bool from string
func (s *impl) decodeBoolValue(val string) (ret model.Value, err error) {
	bVal := false
	switch val {
	case "1":
		bVal = true
	case "0":
		bVal = false
	default:
		err = fmt.Errorf("illegal boolean value, val:%s", val)
	}

	if err != nil {
		return
	}

	ret, err = s.getValue(&bVal)
	return
}
