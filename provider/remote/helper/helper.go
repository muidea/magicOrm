package helper

import (
	"fmt"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

type ElemDependValueFunc func(model.Value) ([]model.Value, error)

type Helper interface {
	Encode(vVal model.Value, tType model.Type) (ret interface{}, err error)
	Decode(val interface{}, tType model.Type) (ret model.Value, err error)
}

type impl struct {
	elemDependValue ElemDependValueFunc
}

func New(elemDependValue ElemDependValueFunc) Helper {
	return &impl{elemDependValue: elemDependValue}
}

/*
local struct -> db field value

remote object value -> db field value
*/
func (s *impl) Encode(vVal model.Value, tType model.Type) (ret interface{}, err error) {
	if !tType.IsBasic() {
		err = fmt.Errorf("illegal value type, type:%s", tType.GetName())
		return
	}

	switch tType.GetValue() {
	case util.TypeBooleanValue:
		ret, err = s.encodeBool(vVal)
	case util.TypeDateTimeValue:
		ret, err = s.encodeDateTime(vVal)
	case util.TypeFloatValue, util.TypeDoubleValue:
		ret, err = s.encodeFloat(vVal)
	case util.TypeBitValue, util.TypeSmallIntegerValue, util.TypeInteger32Value, util.TypeIntegerValue, util.TypeBigIntegerValue:
		ret, err = s.encodeInt(vVal)
	case util.TypePositiveBitValue, util.TypePositiveSmallIntegerValue, util.TypePositiveInteger32Value, util.TypePositiveIntegerValue, util.TypePositiveBigIntegerValue:
		ret, err = s.encodeUint(vVal)
	case util.TypeSliceValue:
		ret, err = s.encodeSlice(vVal, tType)
	case util.TypeStringValue:
		ret, err = s.encodeString(vVal)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	return
}

/*
db field value -> local struct

db field value -> remote object value
*/
func (s *impl) Decode(val interface{}, tType model.Type) (ret model.Value, err error) {
	if !tType.IsBasic() {
		err = fmt.Errorf("illegal value type, type:%s", tType.GetName())
		return
	}

	switch tType.GetValue() {
	case util.TypeBooleanValue:
		ret, err = s.decodeBool(val, tType)
	case util.TypeDateTimeValue:
		ret, err = s.decodeDateTime(val, tType)
	case util.TypeFloatValue, util.TypeDoubleValue:
		ret, err = s.decodeFloat(val, tType)
	case util.TypeBitValue, util.TypeSmallIntegerValue, util.TypeInteger32Value, util.TypeIntegerValue, util.TypeBigIntegerValue:
		ret, err = s.decodeInt(val, tType)
	case util.TypePositiveBitValue, util.TypePositiveSmallIntegerValue, util.TypePositiveInteger32Value, util.TypePositiveIntegerValue, util.TypePositiveBigIntegerValue:
		ret, err = s.decodeUint(val, tType)
	case util.TypeSliceValue:
		ret, err = s.decodeSlice(val, tType)
	case util.TypeStringValue:
		ret, err = s.decodeString(val, tType)
	default:
		err = fmt.Errorf("illegal type, type:%s", tType.GetName())
	}

	if err != nil {
		return
	}

	if tType.IsPtrType() && !ret.IsNil() {
		ret = ret.Addr()
	}

	return
}
