package helper

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

// encodeBoolValue get bool value str
func (s *impl) encodeBoolValue(vVal model.Value) (ret string, err error) {
	val, ok := vVal.Get().(bool)
	if ok {
		if val {
			ret = "1"
			return
		}
		ret = "0"
		return
	}

	err = fmt.Errorf("illegal boolean value, val:%v", vVal.Get())

	return
}

// decodeBoolValue decode bool from string
func (s *impl) decodeBoolValue(val string, tType model.Type) (ret model.Value, err error) {
	bVal := false
	switch val {
	case "1":
		bVal = true
	case "0":
		bVal = false
	default:
		err = fmt.Errorf("illegal boolean value, val:%s", val)
	}

	if err != nil {
		return
	}

	ret, err = s.getValue(bVal)
	if err != nil {
		return
	}

	if tType.IsPtrType() {
		ret = ret.Addr()
	}
	return
}
