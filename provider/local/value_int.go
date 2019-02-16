package local

import (
	"fmt"
	"reflect"
)

//GetIntValueStr get int value str
func GetIntValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%d", rawVal.Int())

	return
}

//GetUintValueStr get uint value str
func GetUintValueStr(val reflect.Value) (ret string, err error) {
	rawVal := reflect.Indirect(val)
	ret = fmt.Sprintf("%d", rawVal.Uint())

	return
}
