package codec

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

const (
	falseVal = int8(iota)
	trueVal
)

// encodeBool encode bool to int
func (s *impl) encodeBool(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	switch vVal.Get().(type) {
	case bool:
		if vVal.Get().(bool) {
			ret = trueVal
		} else {
			ret = falseVal
		}
	default:
		err = fmt.Errorf("encodeBool failed, illegal boolean value, value:%v", vVal.Get())
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
	tVal.Set(bVal)

	ret = tVal
	return
}
