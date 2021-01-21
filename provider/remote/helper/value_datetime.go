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
func (s *impl) decodeDateTime(val interface{}, tType model.Type) (ret model.Value, err error) {
	rVal := reflect.ValueOf(val)
	if rVal.Kind() == reflect.Interface {
		rVal = rVal.Elem()
	}
	rVal = reflect.Indirect(rVal)

	var dtVal string
	switch rVal.Kind() {
	case reflect.String:
		str := rVal.String()
		if str == "" {
			str = "0001-01-01 00:00:00"
		}
		dtVal = str
	default:
		err = fmt.Errorf("illegal dateTime value, val:%v", val)
	}

	if err != nil {
		return
	}

	ret, err = s.getValue(dtVal)
	return
}
