package codec

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
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
	ret, err = vType.Interface(val)
	return
}
