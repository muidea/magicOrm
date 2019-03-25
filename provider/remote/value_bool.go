package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// EncodeBoolValue get bool value str
func EncodeBoolValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	if rawVal.Bool() {
		ret = "1"
	} else {
		ret = "0"
	}

	return
}

func DecodeBoolValue(val string, vType model.Type) (ret reflect.Value, err error) {
	if vType.GetType().Kind() != reflect.Bool {
		err = fmt.Errorf("illegal value type")
		return
	}

	ret = reflect.Indirect(vType.Interface())
	if val == "1" {
		ret.SetBool(true)
	} else if val == "0" {
		ret.SetBool(false)
	} else {
		err = fmt.Errorf("illegal bool value")
	}

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
