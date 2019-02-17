package local

import (
	"fmt"
	"reflect"
)

// getFloatValueStr get float value str
func getFloatValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%f", rawVal.Float())

	return
}
