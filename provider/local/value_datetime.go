package local

import (
	"fmt"
	"reflect"
	"time"
)

//GetDateTimeValueStr get datetime value str
func GetDateTimeValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ts, ok := rawVal.Interface().(time.Time)
	if ok {
		ret = fmt.Sprintf("'%s'", ts.Format("2006-01-02 15:04:05"))
	} else {
		err = fmt.Errorf("no support get string value from struct, [%s]", rawVal.Type().String())
	}

	return
}
