package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/util"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

//encodeInt get int value str
func (s *impl) encodeInt(vVal model.Value) (ret string, err error) {
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

// decodeInt decode int from string
func (s *impl) decodeInt(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var iVal int64
	switch rVal.Kind() {
	case reflect.String:
		iVal, err = strconv.ParseInt(rVal.String(), 0, 64)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		iVal = rVal.Int()
	default:
		err = fmt.Errorf("illegal int value, val:%v", val)
	}
	if err != nil {
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

	return
}

//encodeUint get uint value str
func (s *impl) encodeUint(vVal model.Value) (ret string, err error) {
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

// decodeUint decode uint from string
func (s *impl) decodeUint(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var uVal uint64
	switch rVal.Kind() {
	case reflect.String:
		uVal, err = strconv.ParseUint(rVal.String(), 0, 64)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		uVal = rVal.Uint()
	default:
		err = fmt.Errorf("illegal int value, val:%v", val)
	}
	if err != nil {
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
	return
}
