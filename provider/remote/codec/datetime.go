package codec

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

// encodeDateTime get datetime value str
func (s *impl) encodeDateTime(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
	switch vVal.Get().(type) {
	case string:
		ret = vVal.Get().(string)
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeDateTime failed, illegal dateTime type, value:%v", vVal.Get()))
	}

	return
}

// decodeDateTime decode datetime from string
func (s *impl) decodeDateTime(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}
