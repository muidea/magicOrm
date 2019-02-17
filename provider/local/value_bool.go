package local

import (
	"reflect"
)

// getBoolValueStr get bool value str
func getBoolValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	if rawVal.Bool() {
		ret = "1"
	} else {
		ret = "0"
	}

	return
}
