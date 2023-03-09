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
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		ret = fmt.Sprintf("%d", val.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		ret = fmt.Sprintf("%d", val.Uint())
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%g", val.Float())
	case reflect.Bool:
		ret = fmt.Sprintf("%t", val.Bool())
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
