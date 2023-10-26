package codec

import (
	"fmt"
	"reflect"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/util"

	"github.com/muidea/magicOrm/model"
)

// encodeDateTime get datetime value str
func (s *impl) encodeDateTime(vVal model.Value, _ model.Type) (ret interface{}, err *cd.Result) {
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
			err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeDateTime failed, illegal dateTime value, type:%s", val.Type().String()))
		}
	default:
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("encodeDateTime failed, illegal dateTime type, type:%s", val.Type().String()))
	}

	return
}

// decodeDateTime decode datetime from string
func (s *impl) decodeDateTime(val interface{}, vType model.Type) (ret model.Value, err *cd.Result) {
	ret, err = vType.Interface(val)
	return
}
