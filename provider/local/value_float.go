package local

import (
	"fmt"
	"reflect"
)

// GetFloatValueStr get float value str
func GetFloatValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%f", rawVal.Float())

	return
}
