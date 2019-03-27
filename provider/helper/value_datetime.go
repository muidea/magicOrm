package helper

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
)

//EncodeDateTimeValue get datetime value str
func EncodeDateTimeValue(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
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
	case reflect.Interface:
		dtVal, dtOK := val.Interface().(time.Time)
		if dtOK {
			ret = fmt.Sprintf("%s", dtVal.Format("2006-01-02 15:04:05"))
		} else {
			strVal, strOK := val.Interface().(string)
			if strOK {
				_, tmErr := time.ParseInLocation("2006-01-02 15:04:05", strVal, time.Local)
				if tmErr != nil {
					err = fmt.Errorf("illegal datetime value, val:%v", strVal)
				} else {
					ret = val.String()
				}
			} else {
				err = fmt.Errorf("illegal datetime value, val:%v", val.Interface())
			}
		}
	default:
		err = fmt.Errorf("illegal value, type:%s", val.Type().String())
	}

	return
}

// DecodeDateTimeValue decode datetime from string
func DecodeDateTimeValue(val string, vType model.Type) (ret reflect.Value, err error) {
	ret = reflect.Indirect(vType.Interface())

	tmVal, tmErr := time.ParseInLocation("2006-01-02 15:04:05", val, time.Local)
	if tmErr != nil {
		err = tmErr
		return
	}

	switch vType.GetType().Kind() {
	case reflect.Struct:
		if vType.GetType().String() != "time.Time" {
			err = fmt.Errorf("unsupport value type, type:%s", vType.GetType().String())
			return
		}

		ret.Set(reflect.ValueOf(tmVal))
	case reflect.String:
		ret.SetString(val)
	default:
		err = fmt.Errorf("unsupport value type, type:%s", vType.GetType().String())
	}

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
