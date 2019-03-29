package helper

import (
	"fmt"
	"reflect"
	"time"
)

// ConvertValue convert interface{} to reflect.Value
func ConvertValue(itemVal interface{}, val *reflect.Value) (err error) {
	vType := val.Type()
	switch vType.Kind() {
	case reflect.Bool:
		bVal, bOK := itemVal.(bool)
		if !bOK {
			err = fmt.Errorf("illegal bool value")
		} else {
			val.SetBool(bVal)
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		intVal, intOK := itemVal.(int64)
		if !intOK {
			err = fmt.Errorf("illegal int value")
		} else {
			val.SetInt(intVal)
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		uintVal, uintOK := itemVal.(uint64)
		if !uintOK {
			err = fmt.Errorf("illegal uint value")
		} else {
			val.SetUint(uintVal)
		}
	case reflect.Float32, reflect.Float64:
		fltVal, fltOK := itemVal.(float64)
		if !fltOK {
			err = fmt.Errorf("illegal float value")
		} else {
			val.SetFloat(fltVal)
		}
	case reflect.String:
		strVal, strOK := itemVal.(string)
		if !strOK {
			err = fmt.Errorf("illegal string value")
		} else {
			val.SetString(strVal)
		}
	case reflect.Struct:
		if vType.String() == "time.Time" {
			strVal, strOK := itemVal.(string)
			if !strOK {
				err = fmt.Errorf("illegal datetime value")
			} else {
				dtVal, dtErr := time.ParseInLocation("2006-01-02 15:04:05", strVal, time.Local)
				if dtErr != nil {
					err = fmt.Errorf("illegal datetime value, err:%s", dtErr.Error())
				} else {
					val.Set(reflect.ValueOf(dtVal))
				}
			}
		} else {
			err = fmt.Errorf("illegal field value")
		}
	case reflect.Slice:
		sliceErr := ConvertSliceValue(itemVal, val)
		if sliceErr != nil {
			err = sliceErr
		}
	default:
		err = fmt.Errorf("illegal field type")
	}

	return
}

// ConvertSliceValue convert interface{} slice to reflect.Value
func ConvertSliceValue(itemVal interface{}, val *reflect.Value) (err error) {
	vType := val.Type().Elem()
	sliceVal := itemVal.([]interface{})
	for _, v := range sliceVal {
		iv := reflect.New(vType).Elem()
		vErr := ConvertValue(v, &iv)
		if vErr != nil {
			err = vErr
			return
		}

		*val = reflect.Append(*val, iv)
	}

	return
}
