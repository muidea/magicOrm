package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeString get string value str
func (s *impl) encodeString(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
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

	tVal, _ := tType.Interface()
	switch tVal.Get().Kind() {
	case reflect.String:
		tVal.Get().SetString(strVal)
	default:
		err = fmt.Errorf("illegal string value, type:%s", tVal.Get().Type().String())
	}
	if err != nil {
		return
	}

	ret = tVal
	return
}
