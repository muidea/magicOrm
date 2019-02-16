package local

import (
	"fmt"
	"reflect"
)

// GetStringValueStr get string value str
func GetStringValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("'%s'", rawVal.String())

	return
}
