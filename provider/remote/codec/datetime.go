package codec

import (
	"fmt"
	"time"

	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
)

// encodeDateTime get datetime value str
func (s *impl) encodeDateTime(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	switch vVal.Get().(type) {
	case string:
		ret = vVal.Get().(string)
	default:
		err = fmt.Errorf("encodeDateTime failed, illegal dateTime type, value:%v", vVal.Get())
	}

	return
}

// decodeDateTime decode datetime from string
func (s *impl) decodeDateTime(val interface{}, vType model.Type) (ret model.Value, err error) {
	strVal := ""
	switch val.(type) {
	case string:
		strVal = val.(string)
	default:
		err = fmt.Errorf("decodeDateTime failed, illegal dateTime value, val:%v", val)
	}

	if err != nil {
		return
	}

	_, dtErr := time.Parse(util.CSTLayout, strVal)
	if dtErr != nil {
		err = fmt.Errorf("decodeDateTime failed, illegal dateTime value, val:%v", strVal)
	}

	tVal, _ := vType.Interface(nil)
	err = tVal.Set(strVal)
	if err != nil {
		return
	}

	ret = tVal
	return
}
