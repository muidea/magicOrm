package helper

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

//encodeInt get int value str
func (s *impl) encodeInt(vVal model.Value) (ret string, err error) {
	val := reflect.ValueOf(vVal.Get())
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = fmt.Sprintf("%d", val.Int())
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%d", int64(val.Float()))
	default:
		err = fmt.Errorf("illegal int value, type:%s", val.Type().String())
	}

	return
}

// decodeInt decode int from string
func (s *impl) decodeInt(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	var iVal int64
	switch tVal.Kind() {
	case reflect.String:
		iVal, err = strconv.ParseInt(tVal.String(), 0, 64)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		iVal = tVal.Int()
	case reflect.Float32, reflect.Float64:
		iVal = int64(tVal.Float())
	default:
		err = fmt.Errorf("illegal int value, value type:%v", tVal.Type().String())
	}
	if err != nil {
		return
	}

	cVal = reflect.Indirect(cVal)
	cVal.SetInt(iVal)
	if tType.IsPtrType() {
		cVal = cVal.Addr()
	}
	ret = cVal

	return
}

//encodeUint get uint value str
func (s *impl) encodeUint(vVal model.Value) (ret string, err error) {
	val := reflect.ValueOf(vVal.Get())
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = fmt.Sprintf("%d", val.Uint())
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%d", uint64(val.Float()))
	default:
		err = fmt.Errorf("illegal uint value, type:%s", val.Type().String())
	}

	return
}

// decodeUint decode uint from string
func (s *impl) decodeUint(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	var uVal uint64
	switch tVal.Kind() {
	case reflect.String:
		uVal, err = strconv.ParseUint(tVal.String(), 0, 64)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		uVal = tVal.Uint()
	case reflect.Float32, reflect.Float64:
		uVal = uint64(tVal.Float())
	default:
		err = fmt.Errorf("illegal uint value, value type:%v", tVal.Type().String())
	}
	if err != nil {
		return
	}

	cVal = reflect.Indirect(cVal)
	cVal.SetUint(uVal)
	if tType.IsPtrType() {
		cVal = cVal.Addr()
	}
	ret = cVal
	return
}
