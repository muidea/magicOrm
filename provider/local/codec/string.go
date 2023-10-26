package codec

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

// encodeString get string value str
func (s *impl) encodeString(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.String:
		ret = val.String()
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeSting failed, illegal string value, type:%s", val.Type().String()))
	}

	return
}

// decodeString decode string from string
func (s *impl) decodeString(val interface{}, tType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = tType.Interface(val)
	return
}
