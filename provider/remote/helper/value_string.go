package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeString encode string value to string
// json value to database value string -> string
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

// decodeString decode string from string
// database value to json value string -> string
// json value to json value string -> string
func (s *impl) decodeString(val interface{}, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface()
	switch val.(type) {
	case string:
		tVal.Get().SetString(val.(string))
	default:
		err = fmt.Errorf("illegal string value, type:%s", tVal.Get().Type().String())
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
