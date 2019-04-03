package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/util"
)

// ConvertValue convert interface{} to reflect.Value
func ConvertValue(rawVal reflect.Value, dstVal *reflect.Value) (err error) {
	if util.IsNil(rawVal) {
		return
	}

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
			err = fmt.Errorf("illegal bool value, rawVal:%v", rawVal.Interface())
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
			err = fmt.Errorf("illegal int value, rawVal:%v", rawVal.Interface())
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
			err = fmt.Errorf("illegal uint value, rawVal:%v", rawVal.Interface())
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
			err = fmt.Errorf("illegal float value, rawVal:%v", rawVal.Interface())
		}
	case reflect.String:
		switch rawVal.Kind() {
		case reflect.Bool:
			if rawVal.Bool() {
				dstVal.SetString("1")
			} else {
				dstVal.SetString("0")
			}
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			strVal := fmt.Sprintf("%d", rawVal.Int())
			dstVal.SetString(strVal)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			strVal := fmt.Sprintf("%d", rawVal.Uint())
			dstVal.SetString(strVal)
		case reflect.Float32, reflect.Float64:
			strVal := fmt.Sprintf("%f", rawVal.Float())
			dstVal.SetString(strVal)
		case reflect.Struct:
			if rawVal.Type().String() == "time.Time" {
				strVal := fmt.Sprintf("%s", rawVal.Interface().(time.Time).Format("2006-01-02 15:04:05"))
				dstVal.SetString(strVal)
			} else {
				err = fmt.Errorf("illegal string value, rawVal:%v", rawVal.Interface())
			}
		case reflect.String:
			dstVal.SetString(rawVal.String())
		case reflect.Slice:
			array := []string{}
			for idx := 0; idx < rawVal.Len(); idx++ {
				str := ""
				val := reflect.ValueOf(str)
				valErr := ConvertValue(rawVal.Index(idx), &val)
				if valErr != nil {
					err = fmt.Errorf("convert to string failed, err:%s", valErr.Error())
					return
				}

				array = append(array, str)
			}
			data, dataErr := json.Marshal(&array)
			if dataErr != nil {
				err = fmt.Errorf("marshal slice to string failed, err:%s", dataErr.Error())
				return
			}
			dstVal.SetString(string(data))
		default:
			err = fmt.Errorf("illegal string value, rawVal:%v", rawVal.Interface())
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
			err = fmt.Errorf("illegal datetime value, rawVal:%v", rawVal.Interface())
		}
	case reflect.Slice:
		sliceErr := ConvertSliceValue(rawVal, dstVal)
		if sliceErr != nil {
			err = sliceErr
		}
	default:
		err = fmt.Errorf("illegal field type, rawVal:%v", rawVal.Interface())
	}

	return
}

// ConvertSliceValue convert interface{} slice to reflect.Value
func ConvertSliceValue(rawVal reflect.Value, dstVal *reflect.Value) (err error) {
	if util.IsNil(rawVal) {
		return
	}

	if rawVal.Kind() == reflect.String {
		array := []string{}
		err = json.Unmarshal([]byte(rawVal.String()), &array)
		if err != nil {
			err = fmt.Errorf("unmarshal to slice failed,err:%s", err.Error())
			return
		}

		rawVal = reflect.ValueOf(array)
	}

	if rawVal.Kind() != reflect.Slice || dstVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal field type, val:%v", rawVal.Interface())
		return
	}

	vType := dstVal.Type().Elem()
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
