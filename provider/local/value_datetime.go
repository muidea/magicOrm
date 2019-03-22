package local

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
)

//encodeDateTimeValue get datetime value str
func encodeDateTimeValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	if rawVal.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal datetime value type. type kind:%v", rawVal.Kind())
		return
	}

	ts, ok := rawVal.Interface().(time.Time)
	if ok {
		ret = fmt.Sprintf("%s", ts.Format("2006-01-02 15:04:05"))
	} else {
		err = fmt.Errorf("no support get string value from struct, [%s]", rawVal.Type().String())
	}

	return
}

func decodeDateTimeValue(val string, vType model.Type) (ret reflect.Value, err error) {
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
