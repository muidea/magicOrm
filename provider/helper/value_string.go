package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeString get string value str
func (s *impl) encodeString(vVal model.Value) (ret string, err error) {
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

//decodeString decode string from string
func (s *impl) decodeString(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var strVal string
	switch rVal.Kind() {
	case reflect.String:
		strVal = rVal.String()
	default:
		err = fmt.Errorf("illegal string value, val:%v", val)
	}
	if err != nil {
		return
	}

	ret, err = s.getValue(&strVal)
	return
}
