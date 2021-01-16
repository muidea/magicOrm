package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/util"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

//encodeIntValue get int value str
func (s *impl) encodeIntValue(vVal model.Value) (ret string, err error) {
	val := vVal.Get().(reflect.Value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = fmt.Sprintf("%d", val.Int())
	default:
		err = fmt.Errorf("illegal int value, type:%s", val.Type().String())
	}

	return
}

// decodeIntValue decode int from string
func (s *impl) decodeIntValue(val string, tType model.Type) (ret model.Value, err error) {
	iVal, iErr := strconv.ParseInt(val, 0, 64)
	if iErr != nil {
		err = iErr
		return
	}

	switch tType.GetValue() {
	case util.TypeBitField:
		i8Val := int8(iVal)
		ret, err = s.getValue(&i8Val)
	case util.TypeSmallIntegerField:
		i16Val := int16(iVal)
		ret, err = s.getValue(&i16Val)
	case util.TypeInteger32Field:
		i32Val := int32(iVal)
		ret, err = s.getValue(&i32Val)
	case util.TypeIntegerField:
		i32Val := int(iVal)
		ret, err = s.getValue(&i32Val)
	case util.TypeBigIntegerField:
		ret, err = s.getValue(&iVal)
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
	val := vVal.Get().(reflect.Value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = fmt.Sprintf("%d", val.Uint())
	default:
		err = fmt.Errorf("illegal uint value, type:%s", val.Type().String())
	}

	return
}

// decodeUintValue decode uint from string
func (s *impl) decodeUintValue(val string, tType model.Type) (ret model.Value, err error) {
	uVal, uErr := strconv.ParseUint(val, 0, 64)
	if uErr != nil {
		err = uErr
		return
	}
	switch tType.GetValue() {
	case util.TypePositiveBitField:
		u8Val := uint8(uVal)
		ret, err = s.getValue(&u8Val)
	case util.TypePositiveSmallIntegerField:
		u16Val := uint16(uVal)
		ret, err = s.getValue(&u16Val)
	case util.TypePositiveInteger32Field:
		u32Val := uint32(uVal)
		ret, err = s.getValue(&u32Val)
	case util.TypePositiveIntegerField:
		u32Val := uint(uVal)
		ret, err = s.getValue(&u32Val)
	case util.TypePositiveBigIntegerField:
		ret, err = s.getValue(&uVal)
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