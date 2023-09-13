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
	ret, err = vType.Interface(val)
	return
}
