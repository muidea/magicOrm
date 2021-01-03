package helper

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
)

//encodeDateTimeValue get datetime value str
func (s *impl) encodeDateTimeValue(vVal model.Value, tType model.Type) (ret string, err error) {
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
	default:
		err = fmt.Errorf("illegal dateTime value, type:%s", val.Type().String())
	}

	return
}

// decodeDateTimeValue decode datetime from string
func (s *impl) decodeDateTimeValue(val string, tType model.Type) (ret model.Value, err error) {
	if val == "" {
		val = "0001-01-01 00:00:00"
	}

	dtVal, dtErr := time.Parse("2006-01-02 15:04:05", val)
	if dtErr != nil {
		err = dtErr
		return
	}

	ret, err = s.getValue(&dtVal)
	return
}
