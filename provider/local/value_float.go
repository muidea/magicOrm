package local

import (
	"fmt"
	"reflect"
)

// encodeFloatValue get float value str
func encodeFloatValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%f", rawVal.Float())

	return
}
