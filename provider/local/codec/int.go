package codec

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

func (s *impl) encodeInt(vVal model.Value, _ model.Type) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = val.Int()
	case reflect.Float32, reflect.Float64:
		ret = int64(val.Float())
	default:
		err = fmt.Errorf("encodeInt failed, illegal int value, type:%s", val.Type().String())
	}

	return
}

// decodeInt decode int from string
func (s *impl) decodeInt(val interface{}, vType model.Type) (ret model.Value, err error) {
	ret, err = vType.Interface(val)
	return
}

// encodeUint get uint value str
func (s *impl) encodeUint(vVal model.Value, _ model.Type) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = val.Uint()
	case reflect.Float32, reflect.Float64:
		ret = uint64(val.Float())
	default:
		err = fmt.Errorf("encodeUint failed, illegal uint value, type:%s", val.Type().String())
	}

	return
}

// decodeUint decode uint from string
func (s *impl) decodeUint(val interface{}, vType model.Type) (ret model.Value, err error) {
	ret, err = vType.Interface(val)
	return
}
