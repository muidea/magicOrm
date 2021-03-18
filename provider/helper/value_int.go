package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

//encodeInt get int value str
func (s *impl) encodeInt(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = val.Int()
	case reflect.Float32, reflect.Float64:
		ret = int64(val.Float())
	default:
		err = fmt.Errorf("illegal int value, type:%s", val.Type().String())
	}

	return
}

// decodeInt decode int from string
func (s *impl) decodeInt(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var iVal int64
	switch rVal.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		iVal = rVal.Int()
	case reflect.Float64: // only for json unmarshal
		iVal = int64(rVal.Float())
	default:
		err = fmt.Errorf("illegal int value, val:%v", val)
	}
	if err != nil {
		return
	}

	tVal, _ := tType.Interface()
	switch tVal.Get().Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		tVal.Get().SetInt(iVal)
	default:
		err = fmt.Errorf("illegal int value, type:%s", tVal.Get().Type().String())
	}
	if err != nil {
		return
	}

	ret = tVal
	return
}

//encodeUint get uint value str
func (s *impl) encodeUint(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = val.Uint()
	default:
		err = fmt.Errorf("illegal uint value, type:%s", val.Type().String())
	}

	return
}

// decodeUint decode uint from string
func (s *impl) decodeUint(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var uVal uint64
	switch rVal.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		uVal = rVal.Uint()
	case reflect.Float64: // only for json unmarshal
		uVal = uint64(rVal.Float())
	default:
		err = fmt.Errorf("illegal uint value, val:%v", val)
	}
	if err != nil {
		return
	}

	tVal, _ := tType.Interface()
	switch tVal.Get().Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		tVal.Get().SetUint(uVal)
	default:
		err = fmt.Errorf("illegal uint value, type:%s", tVal.Get().Type().String())
	}
	if err != nil {
		return
	}

	ret = tVal
	return
}
