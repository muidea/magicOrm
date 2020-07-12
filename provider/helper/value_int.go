package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

//EncodeIntValue get int value str
func EncodeIntValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Float64:
		ret = fmt.Sprintf("%d", int64(val.Float()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ret = fmt.Sprintf("%d", val.Int())
	default:
		err = fmt.Errorf("illegal int value, type:%s", val.Type().String())
	}

	return
}

// DecodeIntValue decode int from string
func DecodeIntValue(val string, vType model.Type) (ret reflect.Value, err error) {
	tVal := vType.GetValue()
	switch tVal {
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
	default:
		err = fmt.Errorf("illegal int value type")
		return
	}

	ret = reflect.Indirect(vType.Interface())
	ret, err = AssignValue(reflect.ValueOf(val), ret)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}

//EncodeUintValue get uint value str
func EncodeUintValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Float64:
		ret = fmt.Sprintf("%d", uint64(val.Float()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ret = fmt.Sprintf("%d", val.Uint())
	default:
		err = fmt.Errorf("illegal uint value, type:%s", val.Type().String())
	}

	return
}

// DecodeUintValue decode uint from string
func DecodeUintValue(val string, vType model.Type) (ret reflect.Value, err error) {
	tVal := vType.GetValue()
	switch tVal {
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
	default:
		err = fmt.Errorf("illegal uint value type")
		return
	}

	ret = reflect.Indirect(vType.Interface())
	ret, err = AssignValue(reflect.ValueOf(val), ret)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
