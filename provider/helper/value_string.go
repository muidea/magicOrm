package helper

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// EncodeStringValue get string value str
func EncodeStringValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)

	switch rawVal.Kind() {
	case reflect.String:
		ret = fmt.Sprintf("%s", rawVal.String())
	default:
		err = fmt.Errorf("illegal value, type:%s", rawVal.Type().String())
	}

	return
}

//DecodeStringValue decode string from string
func DecodeStringValue(val string, vType model.Type) (ret reflect.Value, err error) {
	if vType.GetType().Kind() != reflect.String {
		err = fmt.Errorf("unsupport value type, type:%s", vType.GetType().String())
		return
	}

	ret = reflect.Indirect(vType.Interface())
	ret.SetString(val)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
