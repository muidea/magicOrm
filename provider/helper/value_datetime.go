package helper

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
)

//encodeDateTime get datetime value str
func (s *impl) encodeDateTime(vVal model.Value) (ret string, err error) {
	val := vVal.Get().(reflect.Value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Struct:
		ts, ok := val.Interface().(time.Time)
		if ok {
			ret = fmt.Sprintf("%s", ts.Format("2006-01-02 15:04:05"))
			if ret == "0001-01-01 00:00:00" {
				ret = ""
			}
		} else {
			err = fmt.Errorf("illegal dateTime value, type:%s", val.Type().String())
		}
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

	var dtVal time.Time
	switch rVal.Kind() {
	case reflect.String:
		str := rVal.String()
		if str == "" {
			str = "0001-01-01 00:00:00"
		}
		dtVal, err = time.Parse("2006-01-02 15:04:05", str)
	case reflect.Struct:
		if rVal.Type().String() == "time.Time" {
			dtVal = rVal.Interface().(time.Time)
		} else {
			err = fmt.Errorf("illegal dateTime value, val:%v", val)
		}
	default:
		err = fmt.Errorf("illegal dateTime value, val:%v", val)
	}

	if err != nil {
		return
	}

	ret, err = s.getValue(&dtVal)
	return
}
