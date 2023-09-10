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
	bVal := false
	switch val.(type) {
	case int8:
		bVal = val.(int8) > 0
	case string: // only for []bool decode
		bVal = val.(string) == "1"
	case float64: // only for []bool decode
		bVal = val.(float64) == 1
	default:
		err = fmt.Errorf("decodeBool failed, illegal source boolean value, val:%v", val)
	}
	if err != nil {
		return
	}

	tVal := vType.Interface()
	rVal := reflect.Indirect(tVal.Get().(reflect.Value))
	rVal.SetBool(bVal)
	if vType.IsPtrType() {
		err = tVal.Set(rVal.Addr())
	} else {
		err = tVal.Set(rVal)
	}

	if err != nil {
		return
	}
	ret = tVal
	return
}
