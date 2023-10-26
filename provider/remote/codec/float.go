package codec

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
)

// encodeFloat get float value str
func (s *impl) encodeFloat(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
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
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeFloat failed, illegal float value, value:%v", vVal.Get()))
	}

	return
}

// decodeFloat decode float from string
func (s *impl) decodeFloat(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}
