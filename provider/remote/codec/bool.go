package codec

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
)

const (
	falseVal = int8(iota)
	trueVal
)

// encodeBool encode bool to int
func (s *impl) encodeBool(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	switch vVal.Get().(type) {
	case bool:
		if vVal.Get().(bool) {
			ret = trueVal
		} else {
			ret = falseVal
		}
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeBool failed, illegal boolean value, value:%v", vVal.Get()))
	}

	return
}

// decodeBool decode bool from string
func (s *impl) decodeBool(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}
