package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// ConvertValue convert interface{} to reflect.Value
func ConvertValue(rawVal reflect.Value, dstVal *reflect.Value) (err error) {
	vType := dstVal.Type()
	switch vType.Kind() {
	case reflect.Bool:
		switch rawVal.Kind() {
		case reflect.Bool:
			dstVal.SetBool(rawVal.Bool())
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			dstVal.SetBool(rawVal.Int() != 0)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			dstVal.SetBool(rawVal.Uint() != 0)
		case reflect.Float32, reflect.Float64:
			dstVal.SetBool(rawVal.Float() != 0)
		default:
			err = fmt.Errorf("illegal bool value, dstVal:%v", rawVal.Interface())
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		switch rawVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			dstVal.SetInt(rawVal.Int())
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			dstVal.SetInt(int64(rawVal.Uint()))
		case reflect.Float32, reflect.Float64:
			dstVal.SetInt(int64(rawVal.Float()))
		default:
			err = fmt.Errorf("illegal int value, dstVal:%v", rawVal.Interface())
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		switch rawVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			dstVal.SetUint(uint64(rawVal.Int()))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			dstVal.SetUint(rawVal.Uint())
		case reflect.Float32, reflect.Float64:
			dstVal.SetUint(uint64(rawVal.Float()))
		default:
			err = fmt.Errorf("illegal uint value, dstVal:%v", rawVal.Interface())
		}
	case reflect.Float32, reflect.Float64:
		switch rawVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			dstVal.SetFloat(float64(rawVal.Int()))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			dstVal.SetFloat(float64(rawVal.Uint()))
		case reflect.Float32, reflect.Float64:
			dstVal.SetFloat(rawVal.Float())
		default:
			err = fmt.Errorf("illegal float value, dstVal:%v", rawVal.Interface())
		}
	case reflect.String:
		switch rawVal.Kind() {
		case reflect.String:
			dstVal.SetString(rawVal.String())
		default:
			err = fmt.Errorf("illegal string value, dstVal:%v", rawVal.Interface())
		}
	case reflect.Struct:
		switch rawVal.Kind() {
		case reflect.String:
			dtVal, dtErr := time.ParseInLocation("2006-01-02 15:04:05", rawVal.String(), time.Local)
			if dtErr != nil {
				err = fmt.Errorf("illegal datetime value, err:%s", dtErr.Error())
			} else {
				dstVal.Set(reflect.ValueOf(dtVal))
			}
		default:
			err = fmt.Errorf("illegal datetime value, dstVal:%v", rawVal.Interface())
		}
	case reflect.Slice:
		sliceErr := ConvertSliceValue(rawVal, dstVal)
		if sliceErr != nil {
			err = sliceErr
		}
	default:
		err = fmt.Errorf("illegal field type, dstVal:%v", rawVal.Interface())
	}

	return
}

// ConvertSliceValue convert interface{} slice to reflect.Value
func ConvertSliceValue(rawVal reflect.Value, dstVal *reflect.Value) (err error) {
	vType := dstVal.Type().Elem()
	if rawVal.Kind() == reflect.String {
		array := []string{}
		err = json.Unmarshal([]byte(rawVal.String()), &array)
		if err != nil {
			return
		}

		rawVal = reflect.ValueOf(array)
	}

	for idx := 0; idx < rawVal.Len(); idx++ {
		v := rawVal.Index(idx)
		iv := reflect.New(vType).Elem()
		vErr := ConvertValue(v, &iv)
		if vErr != nil {
			err = vErr
			return
		}

		*dstVal = reflect.Append(*dstVal, iv)
	}

	return
}
