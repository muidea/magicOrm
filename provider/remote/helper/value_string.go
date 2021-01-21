package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeString get string value str
func (s *impl) encodeString(vVal model.Value) (ret string, err error) {
	val := reflect.ValueOf(vVal.Get())
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
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
func (s *impl) decodeString(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	var strVal string
	switch tVal.Kind() {
	case reflect.String:
		strVal = tVal.String()
	default:
		err = fmt.Errorf("illegal string value, value type:%v", tVal.Type().String())
	}
	if err != nil {
		return
	}

	cVal = reflect.Indirect(cVal)
	cVal.SetString(strVal)
	if tType.IsPtrType() {
		cVal = cVal.Addr()
	}

	ret = cVal
	return
}
