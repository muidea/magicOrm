package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

//encodeDateTime get datetime value str
func (s *impl) encodeDateTime(vVal model.Value) (ret string, err error) {
	val := reflect.ValueOf(vVal.Get())
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	val = reflect.Indirect(val)

	switch val.Kind() {
	case reflect.String:
		ret = val.String()
	default:
		err = fmt.Errorf("illegal dateTime value, type:%s", val.Type().String())
	}
	return
}

// decodeDateTime decode datetime from string
func (s *impl) decodeDateTime(tVal reflect.Value, tType model.Type, cVal reflect.Value) (ret reflect.Value, err error) {
	var dtVal string
	switch tVal.Kind() {
	case reflect.String:
		str := tVal.String()
		if str == "" {
			str = "0001-01-01 00:00:00"
		}
		dtVal = str
	default:
		err = fmt.Errorf("illegal dateTime value, value type:%v", tVal.Type().String())
	}

	if err != nil {
		return
	}

	cVal = reflect.Indirect(cVal)
	cVal.SetString(dtVal)
	if tType.IsPtrType() {
		cVal = cVal.Addr()
	}
	ret = cVal
	return
}
