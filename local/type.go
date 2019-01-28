package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

// NewFieldType NewFieldType
func NewFieldType(val reflect.Type) (ret model.FieldType, err error) {
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
