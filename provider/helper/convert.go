package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/util"
)

// ConvertValue convert interface{} to reflect.Value
func ConvertValue(fromVal reflect.Value, toVal *reflect.Value) (err error) {
	if util.IsNil(fromVal) {
		return
	}

	if fromVal.Kind() == reflect.Interface {
		fromVal = fromVal.Elem()
	}

	fromVal = reflect.Indirect(fromVal)
	vType := toVal.Type()
	switch vType.Kind() {
	case reflect.Bool:
		switch fromVal.Kind() {
		case reflect.Bool:
			toVal.SetBool(fromVal.Bool())
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			toVal.SetBool(fromVal.Int() != 0)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			toVal.SetBool(fromVal.Uint() != 0)
		case reflect.Float32, reflect.Float64:
			toVal.SetBool(fromVal.Float() != 0)
		default:
			err = fmt.Errorf("illegal bool value, fromVal:%v", fromVal.Interface())
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		switch fromVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			toVal.SetInt(fromVal.Int())
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			toVal.SetInt(int64(fromVal.Uint()))
		case reflect.Float32, reflect.Float64:
			toVal.SetInt(int64(fromVal.Float()))
		default:
			err = fmt.Errorf("illegal int value, fromVal:%v", fromVal.Interface())
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		switch fromVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			toVal.SetUint(uint64(fromVal.Int()))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			toVal.SetUint(fromVal.Uint())
		case reflect.Float32, reflect.Float64:
			toVal.SetUint(uint64(fromVal.Float()))
		default:
			err = fmt.Errorf("illegal uint value, fromVal:%v", fromVal.Interface())
		}
	case reflect.Float32, reflect.Float64:
		switch fromVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			toVal.SetFloat(float64(fromVal.Int()))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			toVal.SetFloat(float64(fromVal.Uint()))
		case reflect.Float32, reflect.Float64:
			toVal.SetFloat(fromVal.Float())
		default:
			err = fmt.Errorf("illegal float value, fromVal:%v", fromVal.Interface())
		}
	case reflect.String:
		switch fromVal.Kind() {
		case reflect.Bool:
			if fromVal.Bool() {
				toVal.SetString("1")
			} else {
				toVal.SetString("0")
			}
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			strVal := fmt.Sprintf("%d", fromVal.Int())
			toVal.SetString(strVal)
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			strVal := fmt.Sprintf("%d", fromVal.Uint())
			toVal.SetString(strVal)
		case reflect.Float32, reflect.Float64:
			strVal := fmt.Sprintf("%f", fromVal.Float())
			toVal.SetString(strVal)
		case reflect.Struct:
			if fromVal.Type().String() == "time.Time" {
				strVal := fmt.Sprintf("%s", fromVal.Interface().(time.Time).Format("2006-01-02 15:04:05"))
				toVal.SetString(strVal)
			} else {
				err = fmt.Errorf("illegal struct value, fromVal:%v", fromVal.Interface())
			}
		case reflect.String:
			toVal.SetString(fromVal.String())
		case reflect.Slice:
			array := []string{}
			for idx := 0; idx < fromVal.Len(); idx++ {
				str := ""
				val := reflect.ValueOf(str)
				valErr := ConvertValue(fromVal.Index(idx), &val)
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
			toVal.SetString(string(data))
		default:
			err = fmt.Errorf("illegal string value, fromVal type:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
		}
	case reflect.Struct:
		switch fromVal.Kind() {
		case reflect.String:
			dtVal, dtErr := time.ParseInLocation("2006-01-02 15:04:05", fromVal.String(), time.Local)
			if dtErr != nil {
				err = fmt.Errorf("illegal datetime value, err:%s", dtErr.Error())
			} else {
				toVal.Set(reflect.ValueOf(dtVal))
			}
		default:
			err = fmt.Errorf("illegal datetime value, fromVal:%v", fromVal.Interface())
		}
	case reflect.Slice:
		sliceErr := ConvertSliceValue(fromVal, toVal)
		if sliceErr != nil {
			err = sliceErr
		}
	default:
		err = fmt.Errorf("illegal field type, fromVal:%v", fromVal.Interface())
	}

	return
}

// ConvertSliceValue convert interface{} slice to reflect.Value
func ConvertSliceValue(fromVal reflect.Value, toVal *reflect.Value) (err error) {
	if util.IsNil(fromVal) {
		return
	}

	if fromVal.Kind() == reflect.Interface {
		fromVal = fromVal.Elem()
	}

	fromVal = reflect.Indirect(fromVal)
	if fromVal.Kind() == reflect.String {
		array := []string{}
		err = json.Unmarshal([]byte(fromVal.String()), &array)
		if err != nil {
			err = fmt.Errorf("unmarshal to slice failed,err:%s", err.Error())
			return
		}

		fromVal = reflect.ValueOf(array)
	}

	if fromVal.Kind() != reflect.Slice || toVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal field type, val:%v", fromVal.Interface())
		return
	}

	itemSlice := reflect.MakeSlice(toVal.Type(), 0, 0)
	vType := toVal.Type().Elem()
	for idx := 0; idx < fromVal.Len(); idx++ {
		v := fromVal.Index(idx)
		iv := reflect.New(vType).Elem()
		vErr := ConvertValue(v, &iv)
		if vErr != nil {
			err = vErr
			return
		}

		itemSlice = reflect.Append(itemSlice, iv)
	}

	toVal.Set(itemSlice)

	return
}
