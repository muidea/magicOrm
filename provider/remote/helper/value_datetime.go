package helper

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
)

//encodeDateTimeValue get datetime value str
func (s *impl) encodeDateTimeValue(vVal model.Value) (ret string, err error) {
	val, ok := vVal.Get().(string)
	if ok {
		ret = val
		return
	}

	err = fmt.Errorf("illegal dateTime value, val:%v", vVal.Get())

	return
}

// decodeDateTimeValue decode datetime from string
func (s *impl) decodeDateTimeValue(val string, tType model.Type) (ret model.Value, err error) {
	if val == "" {
		val = "0001-01-01 00:00:00"
	}

	ret, err = s.getValue(val)
	if err != nil {
		return
	}

	if tType.IsPtrType() {
		ret = ret.Addr()
	}
	return
}
