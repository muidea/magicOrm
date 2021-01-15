package helper

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/muidea/magicOrm/model"
)

//encodeIntValue get int value str
func (s *impl) encodeIntValue(vVal model.Value, tType model.Type) (ret string, err error) {
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

	ret, err = s.getValue(&iVal)
	return
}

//encodeUintValue get uint value str
func (s *impl) encodeUintValue(vVal model.Value, tType model.Type) (ret string, err error) {
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

	ret, err = s.getValue(&uVal)
	return
}
