package helper

import (
	"fmt"
	"strconv"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// encodeFloatValue get float value str
func (s *impl) encodeFloatValue(vVal model.Value) (ret string, err error) {
	val, ok := vVal.Get().(float64)
	if ok {
		ret = fmt.Sprintf("%f", val)
		return
	}

	err = fmt.Errorf("illegal float value, val:%v", vVal.Get())
	return
}

// decodeFloatValue decode float from string
func (s *impl) decodeFloatValue(val string, tType model.Type) (ret model.Value, err error) {
	if tType.GetValue() == util.TypeFloatField {
		fVal, fErr := strconv.ParseFloat(val, 32)
		if fErr != nil {
			err = fErr
			return
		}

		f32Val := float32(fVal)
		ret, err = s.getValue(f32Val)
		if err != nil {
			return
		}

		if tType.IsPtrType() {
			ret = ret.Addr()
		}
		return
	}

	fVal, fErr := strconv.ParseFloat(val, 64)
	if fErr != nil {
		err = fErr
		return
	}

	ret, err = s.getValue(fVal)
	if err != nil {
		return
	}

	if tType.IsPtrType() {
		ret = ret.Addr()
	}
	return
}
