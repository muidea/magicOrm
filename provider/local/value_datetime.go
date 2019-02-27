package local

import (
	"fmt"
	"reflect"
	"time"
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
