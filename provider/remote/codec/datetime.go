package codec

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
)

// encodeDateTime get datetime value str
func (s *impl) encodeDateTime(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	switch vVal.Get().(type) {
	case string:
		ret = vVal.Get().(string)
	default:
		err = fmt.Errorf("encodeDateTime failed, illegal dateTime type, value:%v", vVal.Get())
	}

	return
}

// decodeDateTime decode datetime from string
func (s *impl) decodeDateTime(val interface{}, vType model.Type) (ret model.Value, err error) {
	ret, err = vType.Interface(val)
	return
}
