package codec

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

const (
	falseVal = int8(iota)
	trueVal
)

// encodeBool encode bool to int
func (s *impl) encodeBool(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			ret = trueVal
		} else {
			ret = falseVal
		}
	default:
		err = fmt.Errorf("encodeBool failed, illegal boolean value, type:%s", val.Type().String())
	}

	return
}

// decodeBool decode bool from string
func (s *impl) decodeBool(val interface{}, vType model.Type) (ret model.Value, err error) {
	ret, err = vType.Interface(val)
	return
}
