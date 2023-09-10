package codec

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
)

// encodeDateTime get datetime value str
func (s *impl) encodeDateTime(vVal model.Value, vType model.Type) (ret interface{}, err error) {
	val := reflect.Indirect(vVal.Get().(reflect.Value))
	switch val.Kind() {
	case reflect.Struct:
		ts, ok := val.Interface().(time.Time)
		if ok {
			ret = fmt.Sprintf("%s", ts.Format(util.CSTLayout))
			if ret == "0001-01-01 00:00:00" {
				ret = ""
			}
		} else {
			err = fmt.Errorf("encodeDateTime failed, illegal dateTime value, type:%s", val.Type().String())
		}
	default:
		err = fmt.Errorf("encodeDateTime failed, illegal dateTime type, type:%s", val.Type().String())
	}

	return
}

// decodeDateTime decode datetime from string
func (s *impl) decodeDateTime(val interface{}, vType model.Type) (ret model.Value, err error) {
	strVal := ""
	switch val.(type) {
	case string:
		strVal = val.(string)
	default:
		err = fmt.Errorf("decodeDateTime failed, illegal dateTime value, val:%v", val)
	}

	if err != nil {
		return
	}

	dtVal, dtErr := time.Parse(util.CSTLayout, strVal)
	if dtErr != nil {
		err = fmt.Errorf("decodeDateTime failed, illegal dateTime value, val:%v", strVal)
	}

	tVal := vType.Interface()
	if vType.IsPtrType() {
		err = tVal.Set(reflect.ValueOf(&dtVal))
	} else {
		err = tVal.Set(reflect.ValueOf(dtVal))
	}

	if err != nil {
		return
	}
	ret = tVal
	return
}
