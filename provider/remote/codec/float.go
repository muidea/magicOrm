package codec

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

// encodeFloat get float value str
func (s *impl) encodeFloat(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	switch vVal.Get().(type) {
	case float32:
		ret = vVal.Get().(float32)
	case float64:
		ret = vVal.Get().(float64)
	default:
		err = fmt.Errorf("encodeFloat failed, illegal float value, value:%v", vVal.Get())
	}

	return
}

// decodeFloat decode float from string
func (s *impl) decodeFloat(val interface{}, vType model.Type) (ret model.Value, err error) {
	var fVal float64
	switch val.(type) {
	case float32, float64:
		fVal = reflect.ValueOf(val).Float()
	case string: // only for []float32/[]float64
		fVal, err = strconv.ParseFloat(val.(string), 64)
	default:
		err = fmt.Errorf("decodeFloat failed, illegal float value, val:%v", val)
	}
	if err != nil {
		return
	}

	tVal := vType.Interface()
	tVal.Set(fVal)
	ret = tVal
	return
}
