package local

import (
	"reflect"
)

// GetBoolValueStr get bool value str
func GetBoolValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	if rawVal.Bool() {
		ret = "1"
	} else {
		ret = "0"
	}

	return
}
