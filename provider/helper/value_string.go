package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeStringValue get string value str
func (s *impl) encodeStringValue(vVal model.Value) (ret string, err error) {
	val := vVal.Get().(reflect.Value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.String:
		ret = val.String()
	default:
		err = fmt.Errorf("illegal string value, type:%s", val.Type().String())
	}

	return
}

//decodeStringValue decode string from string
func (s *impl) decodeStringValue(val string) (ret model.Value, err error) {
	ret = s.getValue(reflect.ValueOf(&val).Elem())
	return
}
