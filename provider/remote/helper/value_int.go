package helper

import (
	"fmt"
	"strconv"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

//encodeIntValue get int value str
func (s *impl) encodeIntValue(vVal model.Value) (ret string, err error) {
	i8Val, ok := vVal.Get().(int8)
	if ok {
		ret = fmt.Sprintf("%d", i8Val)
		return
	}
	i16Val, ok := vVal.Get().(int16)
	if ok {
		ret = fmt.Sprintf("%d", i16Val)
		return
	}
	i32Val, ok := vVal.Get().(int32)
	if ok {
		ret = fmt.Sprintf("%d", i32Val)
		return
	}
	iVal, ok := vVal.Get().(int)
	if ok {
		ret = fmt.Sprintf("%d", iVal)
		return
	}
	i64Val, ok := vVal.Get().(int64)
	if ok {
		ret = fmt.Sprintf("%d", i64Val)
		return
	}
	fVal, ok := vVal.Get().(float64)
	if ok {
		ret = fmt.Sprintf("%d", int64(fVal))
		return
	}

	err = fmt.Errorf("illegal int value, val:%v", vVal.Get())
	return
}

// decodeIntValue decode int from string
func (s *impl) decodeIntValue(val string, tType model.Type) (ret model.Value, err error) {
	fVal, fErr := strconv.ParseFloat(val, 64)
	if fErr != nil {
		err = fErr
		return
	}

	switch tType.GetValue() {
	case util.TypeBitField:
		i8Val := int8(fVal)
		ret, err = s.getValue(i8Val)
	case util.TypeSmallIntegerField:
		i16Val := int16(fVal)
		ret, err = s.getValue(i16Val)
	case util.TypeInteger32Field:
		i32Val := int32(fVal)
		ret, err = s.getValue(i32Val)
	case util.TypeIntegerField:
		i32Val := int(fVal)
		ret, err = s.getValue(i32Val)
	case util.TypeBigIntegerField:
		ret, err = s.getValue(fVal)
	default:
		err = fmt.Errorf("illegal integer type, type:%s", tType.GetName())
	}

	if err != nil {
		return
	}

	if tType.IsPtrType() {
		ret = ret.Addr()
	}
	return
}

//encodeUintValue get uint value str
func (s *impl) encodeUintValue(vVal model.Value) (ret string, err error) {
	u8Val, ok := vVal.Get().(uint8)
	if ok {
		ret = fmt.Sprintf("%d", u8Val)
		return
	}
	u16Val, ok := vVal.Get().(uint16)
	if ok {
		ret = fmt.Sprintf("%d", u16Val)
		return
	}
	u32Val, ok := vVal.Get().(uint32)
	if ok {
		ret = fmt.Sprintf("%d", u32Val)
		return
	}
	uVal, ok := vVal.Get().(uint)
	if ok {
		ret = fmt.Sprintf("%d", uVal)
		return
	}
	u64Val, ok := vVal.Get().(uint64)
	if ok {
		ret = fmt.Sprintf("%d", u64Val)
		return
	}
	fVal, ok := vVal.Get().(float64)
	if ok {
		ret = fmt.Sprintf("%d", uint64(fVal))
		return
	}

	err = fmt.Errorf("illegal uint value, value:%v", vVal.Get())

	return
}

// decodeUintValue decode uint from string
func (s *impl) decodeUintValue(val string, tType model.Type) (ret model.Value, err error) {
	fVal, fErr := strconv.ParseFloat(val, 64)
	if fErr != nil {
		err = fErr
		return
	}
	switch tType.GetValue() {
	case util.TypePositiveBitField:
		u8Val := uint8(fVal)
		ret, err = s.getValue(u8Val)
	case util.TypePositiveSmallIntegerField:
		u16Val := uint16(fVal)
		ret, err = s.getValue(u16Val)
	case util.TypePositiveInteger32Field:
		u32Val := uint32(fVal)
		ret, err = s.getValue(u32Val)
	case util.TypePositiveIntegerField:
		u32Val := uint(fVal)
		ret, err = s.getValue(u32Val)
	case util.TypePositiveBigIntegerField:
		ret, err = s.getValue(fVal)
	default:
		err = fmt.Errorf("illegal unsigned integer type, type:%s", tType.GetName())
	}

	if err != nil {
		return
	}

	if tType.IsPtrType() {
		ret = ret.Addr()
	}
	return
}
