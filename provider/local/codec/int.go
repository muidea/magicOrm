package codec

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
	"strconv"
)

func (s *impl) encodeInt(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = val.Int()
	default:
		err = fmt.Errorf("encodeInt failed, illegal int value, type:%s", val.Type().String())
	}

	return
}

// decodeInt decode int from string
func (s *impl) decodeInt(val interface{}, vType model.Type) (ret model.Value, err error) {
	var iVal int64
	switch val.(type) {
	case int, int8, int16, int32, int64:
		iVal = val.(int64)
	case float64: // only for []int
		iVal = int64(val.(float64))
	case string: // only for []int
		iVal, err = strconv.ParseInt(val.(string), 10, 64)
	default:
		err = fmt.Errorf("decodeInt failed, illegal int value, val:%v", val)
	}
	if err != nil {
		return
	}

	tVal := vType.Interface()
	rVal := reflect.Indirect(tVal.Get().(reflect.Value))
	rVal.SetInt(iVal)
	if vType.IsPtrType() {
		err = tVal.Set(rVal.Addr())
	} else {
		err = tVal.Set(rVal)
	}
	if err != nil {
		return
	}
	ret = tVal
	return
}

// encodeUint get uint value str
func (s *impl) encodeUint(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = val.Uint()
	default:
		err = fmt.Errorf("encodeUint failed, illegal uint value, type:%s", val.Type().String())
	}

	return
}

// decodeUint decode uint from string
func (s *impl) decodeUint(val interface{}, vType model.Type) (ret model.Value, err error) {
	var uiVal uint64
	switch val.(type) {
	case uint, uint8, uint16, uint32, uint64:
		uiVal = val.(uint64)
	case float64: // only for []uint
		uiVal = uint64(val.(float64))
	case string: // only for []uint
		uiVal, err = strconv.ParseUint(val.(string), 10, 64)
	default:
		err = fmt.Errorf("decodeUint failed, illegal uint value, val:%v", val)
	}
	if err != nil {
		return
	}

	tVal := vType.Interface()
	rVal := reflect.Indirect(tVal.Get().(reflect.Value))
	rVal.SetUint(uiVal)
	if vType.IsPtrType() {
		err = tVal.Set(rVal.Addr())
	} else {
		err = tVal.Set(rVal)
	}
	if err != nil {
		return
	}
	ret = tVal
	return
}
