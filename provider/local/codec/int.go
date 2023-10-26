package codec

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

func (s *impl) encodeInt(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = val.Int()
	case reflect.Float32, reflect.Float64:
		ret = int64(val.Float())
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeInt failed, illegal int value, type:%s", val.Type().String()))
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
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = val.Uint()
	case reflect.Float32, reflect.Float64:
		ret = uint64(val.Float())
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeUint failed, illegal uint value, type:%s", val.Type().String()))
	}

	return
}

// decodeUint decode uint from string
func (s *impl) decodeUint(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}
