package local

import (
	"reflect"
)

// encodeBoolValue get bool value str
func encodeBoolValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	if rawVal.Bool() {
		ret = "1"
	} else {
		ret = "0"
	}

	return
}
