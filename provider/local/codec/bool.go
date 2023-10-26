package codec

import (
	"fmt"
	"reflect"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

const (
	falseVal = int8(iota)
	trueVal
)

// encodeBool encode bool to int
func (s *impl) encodeBool(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			ret = trueVal
		} else {
			ret = falseVal
		}
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeBool failed, illegal boolean value, type:%s", val.Type().String()))
	}

	return
}

// decodeBool decode bool from string
func (s *impl) decodeBool(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}
