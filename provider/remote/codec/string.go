package codec

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// encodeString get string value str
func (s *impl) encodeString(vVal model.Value, tType model.Type) (ret interface{}, err error) {
	switch vVal.Get().(type) {
	case string:
		ret = vVal.Get().(string)
	default:
		err = fmt.Errorf("encodeSting failed, illegal string value, value:%s", vVal.Get())
	}

	return
}

// decodeString decode string from string
func (s *impl) decodeString(val interface{}, tType model.Type) (ret model.Value, err error) {
	ret, err = tType.Interface(val)
	return
}
