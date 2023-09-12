package codec

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
	"strconv"
)

// encodeFloat get float value str
func (s *impl) encodeFloat(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = val.Float()
	default:
		err = fmt.Errorf("encodeFloat failed, illegal float value, type:%s", val.Type().String())
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
	rVal := reflect.Indirect(tVal.Get().(reflect.Value))
	rVal.SetFloat(fVal)
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
