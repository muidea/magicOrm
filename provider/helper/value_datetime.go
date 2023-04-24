package helper

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
)

// encodeDateTime get datetime value str
func (s *impl) encodeDateTime(vVal model.Value) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get())
	switch val.Kind() {
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
		dtVal = rVal.String()
	default:
		err = fmt.Errorf("illegal dateTime value, val:%v", val)
	}

	if err != nil {
		return
	}

	tVal := tType.Interface()
	switch tVal.Get().Kind() {
	case reflect.Struct:
		if dtVal == "" {
			dtVal = "0001-01-01 00:00:00"
		}
		dtV, dtErr := time.Parse(util.CSTLayout, dtVal)
		if dtErr != nil {
			err = dtErr
		} else {
			tVal.Get().Set(reflect.ValueOf(dtV))
		}
	case reflect.String:
		tVal.Get().SetString(dtVal)
	default:
		err = fmt.Errorf("illegal dateTime value, type:%s", tVal.Get().Type().String())
	}
	if err != nil {
		return
	}

	ret = tVal
	return

}
