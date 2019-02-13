package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

// NewFieldValue NewFieldValue
func NewFieldValue(val reflect.Value) (ret model.FieldValue, err error) {
	if val.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal val, must be a ptr")
		return
	}

	rawVal := reflect.Indirect(val)
	switch rawVal.Kind() {
	case reflect.Bool:
		ret = &boolImpl{value: rawVal}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		ret = &intImpl{value: rawVal}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		ret = &uintImpl{value: rawVal}
	case reflect.Float32, reflect.Float64:
		ret = &floatImpl{value: rawVal}
	case reflect.String:
		ret = &stringImpl{value: rawVal}
	case reflect.Struct:
		if rawVal.Type().String() == "time.Time" {
			ret = &datetimeImpl{value: rawVal}
		} else {
			ret = &structImpl{value: rawVal}
		}
	case reflect.Slice:
		ret = &sliceImpl{value: rawVal}
	case reflect.Ptr:
		if rawVal.IsNil() {
			ret = &nilImpl{value: rawVal}
			return
		}
		rawRawVal := reflect.Indirect(rawVal)
		switch rawRawVal.Kind() {
		case reflect.Bool:
			ret = &boolImpl{value: rawVal}
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			ret = &intImpl{value: rawVal}
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			ret = &uintImpl{value: rawVal}
		case reflect.Float32, reflect.Float64:
			ret = &floatImpl{value: rawVal}
		case reflect.String:
			ret = &stringImpl{value: rawVal}
		case reflect.Struct:
			if rawVal.Type().String() == "time.Time" {
				ret = &datetimeImpl{value: rawVal}
			} else {
				ret = &structImpl{value: rawVal}
			}
		case reflect.Slice:
			ret = &sliceImpl{value: rawVal}
		default:
			err = fmt.Errorf("no support value ptr type, type:%s", val.Type().String())
		}
	default:
		err = fmt.Errorf("no support value type, kind:%s, type:%s", val.Kind().String(), val.Type().String())
	}

	return
}

// GetModelValueStr GetModelValueStr
func GetModelValueStr(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	if val.Kind() != reflect.Struct {
		err = fmt.Errorf("illegal model type, type:%s", val.Type().String())
		return
	}

	pkField, pkErr := getStructPrimaryKey(val)
	if pkErr != nil {
		err = pkErr
		return
	}

	ret, err = pkField.GetValue().GetValueStr()
	return
}