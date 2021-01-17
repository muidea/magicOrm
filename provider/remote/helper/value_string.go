package helper

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// encodeStringValue get string value str
func (s *impl) encodeStringValue(vVal model.Value) (ret string, err error) {
	val, ok := vVal.Get().(string)
	if ok {
		ret = val
		return
	}

	err = fmt.Errorf("illegal string value, value:%v", vVal.Get())
	return
}

//decodeStringValue decode string from string
func (s *impl) decodeStringValue(val string, tType model.Type) (ret model.Value, err error) {
	ret, err = s.getValue(val)
	if err != nil {
		return
	}

	if tType.IsPtrType() {
		ret = ret.Addr()
	}
	return
}
