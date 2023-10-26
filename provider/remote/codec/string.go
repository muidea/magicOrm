package codec

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"

	"github.com/muidea/magicOrm/model"
)

// encodeString get string value str
func (s *impl) encodeString(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	switch vVal.Get().(type) {
	case string:
		ret = vVal.Get().(string)
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeSting failed, illegal string value, value:%v", vVal.Get()))
	}

	return
}

// decodeString decode string from string
func (s *impl) decodeString(val interface{}, tType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = tType.Interface(val)
	return
}
