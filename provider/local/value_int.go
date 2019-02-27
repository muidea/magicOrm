package local

import (
	"fmt"
	"reflect"
)

//encodeIntValue get int value str
func encodeIntValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%d", rawVal.Int())

	return
}

//encodeUintValue get uint value str
func encodeUintValue(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%d", rawVal.Uint())

	return
}
