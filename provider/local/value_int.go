package local

import (
	"fmt"
	"reflect"
)

//getIntValueStr get int value str
func getIntValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%d", rawVal.Int())

	return
}

//getUintValueStr get uint value str
func getUintValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%d", rawVal.Uint())

	return
}
