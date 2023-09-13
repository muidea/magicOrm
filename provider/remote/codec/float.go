package codec

import (
	"fmt"
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
	ret, err = vType.Interface(val)
	return
}
