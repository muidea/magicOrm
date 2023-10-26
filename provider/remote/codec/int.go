package codec

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) encodeInt(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	switch vVal.Get().(type) {
	case int8, int16, int32, int, int64:
		ret = reflect.ValueOf(vVal.Get()).Int()
	case uint8, uint16, uint32, uint, uint64:
		ret = reflect.ValueOf(vVal.Get()).Uint()
	case float32, float64:
		ret = int64(reflect.ValueOf(vVal.Get()).Float())
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeInt failed, illegal int value, value:%v", vVal.Get()))
	}

	return
}

// decodeInt decode int from string
func (s *impl) decodeInt(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}

// encodeUint get uint value str
func (s *impl) encodeUint(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	switch vVal.Get().(type) {
	case uint8, uint16, uint32, uint, uint64:
		ret = reflect.ValueOf(vVal.Get()).Uint()
	case int8, int16, int32, int, int64:
		ret = reflect.ValueOf(vVal.Get()).Int()
	case float32, float64:
		ret = uint64(reflect.ValueOf(vVal.Get()).Float())
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeInt failed, illegal uint value, value:%v", vVal.Get()))
	}

	return
}

// decodeUint decode uint from string
func (s *impl) decodeUint(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}
