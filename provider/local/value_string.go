package local

import (
	"fmt"
	"reflect"
)

// getStringValueStr get string value str
func getStringValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%s", rawVal.String())

	return
}
