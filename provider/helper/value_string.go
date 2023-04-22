package helper

import (
	"encoding/json"
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeString get string value str
func (s *impl) encodeString(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()

	var byteVal []byte
	switch val.Kind() {
	case reflect.String:
		ret = val.String()
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		byteVal, err = json.Marshal(val.Int())
		if err == nil {
			ret = string(byteVal)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		byteVal, err = json.Marshal(val.Uint())
		if err == nil {
			ret = string(byteVal)
		}
	case reflect.Float32, reflect.Float64:
		byteVal, err = json.Marshal(val.Float())
		if err == nil {
			ret = string(byteVal)
		}
	case reflect.Bool:
		byteVal, err = json.Marshal(val.Bool())
		if err == nil {
			ret = string(byteVal)
		}
	default:
		err = fmt.Errorf("illegal string value, type:%s", val.Type().String())
	}

	return
}

// decodeString decode string from string
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

	tVal := tType.Interface()
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
