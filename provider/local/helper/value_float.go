package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/util"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

// encodeFloatValue get float value str
func (s *impl) encodeFloatValue(vVal model.Value) (ret string, err error) {
	val := vVal.Get().(reflect.Value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		ret = fmt.Sprintf("%f", val.Float())
	default:
		err = fmt.Errorf("illegal float value, type:%s", val.Type().String())
	}

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
		ret, err = s.getValue(&f32Val)
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

	ret, err = s.getValue(&fVal)
	if err != nil {
		return
	}

	if tType.IsPtrType() {
		ret = ret.Addr()
	}
	return
}
