package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/util"
)

// FieldType FieldType
type FieldType interface {
	Name() string
	Value() int
	IsPtr() bool
	PkgPath() string
	String() string
	Depend() (dependType reflect.Type, isTypePtr bool)
	Copy() FieldType
}

// NewFieldType NewFieldType
func NewFieldType(val reflect.Type) (ret FieldType, err error) {
	isPtr := false
	rawVal := val
	if rawVal.Kind() == reflect.Ptr {
		rawVal = rawVal.Elem()
		isPtr = true
	}

	tVal, tErr := util.GetTypeValueEnum(rawVal)
	if tErr != nil {
		err = tErr
		return
	}
	if util.IsBasicType(tVal) {
		ret, err = getBasicType(rawVal, isPtr)
		return
	}

	if util.IsStructType(tVal) {
		ret, err = getStructType(rawVal, isPtr)
		return
	}

	if util.IsSliceType(tVal) {
		ret, err = getSliceType(rawVal, isPtr)
		return
	}

	err = fmt.Errorf("illegal fieldType, type:%s", val.String())
	return
}
