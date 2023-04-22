package helper

import (
	"encoding/json"
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeInt encode int value to int64
// json value to database value float64 -> int64
func (s *impl) encodeInt(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = int64(val.Float())
	default:
		err = fmt.Errorf("illegal int value, type:%s", val.Type().String())
	}

	return
}

// decodeInt decode int from int or float
// database value to json value int -> float
// json value to json value float -> float
func (s *impl) decodeInt(val interface{}, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface()
	switch val.(type) {
	case int8, int16, int, int32, int64:
		rVal := reflect.Indirect(reflect.ValueOf(val))
		tVal.Get().SetFloat(float64(rVal.Int()))
	case float64:
		tVal.Get().SetFloat(val.(float64))
	case string: // only for slice element
		var fVal float64
		err = json.Unmarshal([]byte(val.(string)), &fVal)
		if err != nil {
			return
		}
		tVal.Get().SetFloat(fVal)
	default:
		err = fmt.Errorf("illegal int value, type:%s", tVal.Get().Type().String())
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

// encodeUint encode uint value to uint64
// json value to database value float64 -> uint64
func (s *impl) encodeUint(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = uint64(val.Float())
	default:
		err = fmt.Errorf("illegal int value, type:%s", val.Type().String())
	}

	return
}

// decodeUint decode uint from uint or float
// database value to json value uint -> float
// json value to json value float -> float
func (s *impl) decodeUint(val interface{}, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface()
	switch val.(type) {
	case uint8, uint16, uint, uint32, uint64:
		rVal := reflect.Indirect(reflect.ValueOf(val))
		tVal.Get().SetFloat(float64(rVal.Uint()))
	case float64:
		tVal.Get().SetFloat(val.(float64))
	case string: // only for slice element
		var fVal float64
		err = json.Unmarshal([]byte(val.(string)), &fVal)
		if err != nil {
			return
		}
		tVal.Get().SetFloat(fVal)
	default:
		err = fmt.Errorf("illegal int value, type:%s", tVal.Get().Type().String())
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
