package helper

import (
	"fmt"
	"github.com/muidea/magicOrm/model"
	"reflect"
)

// encodeDateTime get datetime value to string
// json value to database value string -> string
func (s *impl) encodeDateTime(vVal model.Value) (ret interface{}, err error) {
	val := vVal.Get()
	switch val.Kind() {
	/*
		case reflect.Struct:
			ts, ok := val.Interface().(time.Time)
			if ok {
				ret = fmt.Sprintf("%s", ts.Format(util.CSTLayout))
				if ret == "0001-01-01 00:00:00" {
					ret = ""
				}
			} else {
				err = fmt.Errorf("illegal dateTime value, type:%s", val.Type().String())
			}
	*/
	case reflect.String:
		ret = val.String()
	default:
		err = fmt.Errorf("encodeDateTime failed,illegal dateTime value, type:%s", val.Type().String())
	}

	return
}

// decodeDateTime decode datetime from string
// database value to json value string -> string
// json value to json value string -> string
func (s *impl) decodeDateTime(val interface{}, tType model.Type) (ret model.Value, err error) {
	tVal := tType.Interface()
	switch val.(type) {
	case string:
		tVal.Get().SetString(val.(string))
	default:
		err = fmt.Errorf("decodeDateTime failed,illegal dateTime value, type:%s", tVal.Get().Type().String())
	}
	if err != nil {
		return
	}
	if tType.IsPtrType() {
		tVal = tVal.Addr()
	}

	ret = tVal
	return

}
