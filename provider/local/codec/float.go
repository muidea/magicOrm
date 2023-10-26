package codec

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

// encodeFloat get float value str
func (s *impl) encodeFloat(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = val.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = val.Uint()
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeFloat failed, illegal float value, type:%s", val.Type().String()))
	}

	return
}

// decodeFloat decode float from string
func (s *impl) decodeFloat(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}
