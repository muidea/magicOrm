package helper

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

//EncodeDateTimeValue get datetime value str
func EncodeDateTimeValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		ts, ok := val.Interface().(time.Time)
		if ok {
			ret = fmt.Sprintf("%s", ts.Format("2006-01-02 15:04:05"))
		} else {
			err = fmt.Errorf("no support get datetime value from struct, [%s]", val.Type().String())
		}
	case reflect.String:
		_, tmErr := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
		if tmErr != nil {
			err = fmt.Errorf("illegal datetime value, val:%v", val.Interface())
		} else {
			ret = val.String()
		}
	default:
		err = fmt.Errorf("illegal value, type:%s", val.Type().String())
	}

	return
}

// DecodeDateTimeValue decode datetime from string
func DecodeDateTimeValue(val string, vType model.Type) (ret reflect.Value, err error) {
	tVal := vType.GetValue()
	switch tVal {
	case util.TypeDateTimeField:
	default:
		err = fmt.Errorf("illegal dateTime value type")
		return
	}

	ret = reflect.Indirect(vType.Interface())
	err = ConvertValue(reflect.ValueOf(val), &ret)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
