package codec

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

func (s *impl) encodeInt(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	switch vVal.Get().(type) {
	case int8, int16, int32, int, int64:
		ret = reflect.ValueOf(vVal.Get()).Int()
	case float32, float64:
		ret = int64(reflect.ValueOf(vVal.Get()).Float())
	default:
		err = fmt.Errorf("encodeInt failed, illegal int value, value:%v", vVal.Get())
	}

	return
}

// decodeInt decode int from string
func (s *impl) decodeInt(val interface{}, vType model.Type) (ret model.Value, err error) {
	var iVal int64
	switch val.(type) {
	case int, int8, int16, int32, int64:
		iVal = reflect.ValueOf(val).Int()
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
	tVal.Set(iVal)
	ret = tVal
	return
}

// encodeUint get uint value str
func (s *impl) encodeUint(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	switch vVal.Get().(type) {
	case uint8, uint16, uint32, uint, uint64:
		ret = reflect.ValueOf(vVal.Get()).Uint()
	case float32, float64:
		ret = uint64(reflect.ValueOf(vVal.Get()).Float())
	default:
		err = fmt.Errorf("encodeInt failed, illegal uint value, value:%v", vVal.Get())
	}

	return
}

// decodeUint decode uint from string
func (s *impl) decodeUint(val interface{}, vType model.Type) (ret model.Value, err error) {
	var uiVal uint64
	switch val.(type) {
	case uint, uint8, uint16, uint32, uint64:
		uiVal = reflect.ValueOf(val).Uint()
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
	tVal.Set(uiVal)
	ret = tVal
	return
}
