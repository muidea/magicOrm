package local

import (
	"fmt"
	"reflect"
)

// encodeStringValue get string value str
func encodeStringValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%s", rawVal.String())

	return
}
