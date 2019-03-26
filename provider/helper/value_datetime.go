package helper

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
)

//EncodeDateTimeValue get datetime value str
func EncodeDateTimeValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	switch rawVal.Kind() {
	case reflect.Struct:
		ts, ok := rawVal.Interface().(time.Time)
		if ok {
			ret = fmt.Sprintf("%s", ts.Format("2006-01-02 15:04:05"))
		} else {
			err = fmt.Errorf("no support get string value from struct, [%s]", rawVal.Type().String())
		}
	case reflect.String:
		_, tmErr := time.ParseInLocation("2006-01-02 15:04:05", rawVal.String(), time.Local)
		if tmErr != nil {
			err = tmErr
		} else {
			ret = rawVal.String()
		}
	default:
		err = fmt.Errorf("illegal value type, type:%s", rawVal.Type().String())
	}

	return
}

// DecodeDateTimeValue decode datetime from string
func DecodeDateTimeValue(val string, vType model.Type) (ret reflect.Value, err error) {
	if vType.GetType().String() != "time.Time" {
		err = fmt.Errorf("illegal value type")
		return
	}

	tmVal, tmErr := time.ParseInLocation("2006-01-02 15:04:05", val, time.Local)
	if tmErr != nil {
		err = tmErr
		return
	}

	ret = reflect.Indirect(vType.Interface())
	ret.Set(reflect.ValueOf(tmVal))
	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
