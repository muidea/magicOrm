package codec

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// encodeFloat get float value str
func (s *impl) encodeFloat(vVal model.Value, _ model.Type) (ret interface{}, err error) {
	switch vVal.Get().(type) {
	case float32:
		ret = vVal.Get().(float32)
	case float64:
		ret = vVal.Get().(float64)
	case int8, int16, int32, int, int64:
		ret = reflect.ValueOf(vVal.Get()).Int()
	case uint8, uint16, uint32, uint, uint64:
		ret = reflect.ValueOf(vVal.Get()).Uint()
	default:
		err = fmt.Errorf("encodeFloat failed, illegal float value, value:%v", vVal.Get())
	}

	return
}

// decodeFloat decode float from string
func (s *impl) decodeFloat(val interface{}, vType model.Type) (ret model.Value, err error) {
	ret, err = vType.Interface(val)
	return
}
