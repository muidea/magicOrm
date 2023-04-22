package helper

import (
	"encoding/json"
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeFloat encode float value to float64
// json value to database value float64 -> float64
func (s *impl) encodeFloat(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = val.Float()
	default:
		err = fmt.Errorf("illegal float value, type:%s", val.Type().String())
	}

	return
}

// decodeFloat decode float from float64
// database value to json value float64 -> float64
// json value to json value float64 -> float64 Or string -> float64
func (s *impl) decodeFloat(val interface{}, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface()
	switch val.(type) {
	case float32, float64:
		rVal := reflect.Indirect(reflect.ValueOf(val))
		tVal.Get().SetFloat(rVal.Float())
	case string: // only for slice element
		var fVal float64
		err = json.Unmarshal([]byte(val.(string)), &fVal)
		if err != nil {
			return
		}
		tVal.Get().SetFloat(fVal)
	default:
		err = fmt.Errorf("illegal float value, type:%s", tVal.Get().Type().String())
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
