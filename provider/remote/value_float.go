package remote

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// EncodeFloatValue get float value str
func EncodeFloatValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%f", rawVal.Float())

	return
}

func DecodeFloatValue(val string, vType model.Type) (ret reflect.Value, err error) {
	ret = reflect.Indirect(vType.Interface())
	switch vType.GetValue() {
	case util.TypeFloatField:
		fVal, fErr := strconv.ParseFloat(val, 32)
		if fErr != nil {
			err = fErr
			return
		}
		ret.SetFloat(fVal)
	case util.TypeDoubleField:
		fVal, fErr := strconv.ParseFloat(val, 64)
		if fErr != nil {
			err = fErr
			return
		}
		ret.SetFloat(fVal)
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		iVal, iErr := strconv.ParseInt(val, 10, 64)
		if iErr != nil {
			err = iErr
			return
		}
		ret.SetInt(iVal)
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		uiVal, uiErr := strconv.ParseUint(val, 10, 64)
		if uiErr != nil {
			err = uiErr
			return
		}
		ret.SetUint(uiVal)
	default:
		err = fmt.Errorf("illegal value type")
		return
	}

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
