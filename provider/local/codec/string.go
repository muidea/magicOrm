package codec

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeString get string value str
func (s *impl) encodeString(vVal model.Value, tType model.Type) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.String:
		ret = val.String()
	default:
		err = fmt.Errorf("encodeSting failed, illegal string value, type:%s", val.Type().String())
	}

	return
}

// decodeString decode string from string
func (s *impl) decodeString(val interface{}, tType model.Type) (ret model.Value, err error) {
	ret, err = tType.Interface(val)
	return
}
