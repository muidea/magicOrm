package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/muidea/magicOrm/util"
)

// AssignValue assign interface{} to reflect.Value
// fromVal -> toVal
// fromVal -> *toVal
func AssignValue(fromVal reflect.Value, toVal reflect.Value) (ret reflect.Value, err error) {
	if fromVal.Kind() == reflect.Interface {
		fromVal = fromVal.Elem()
	}

	if util.IsNil(toVal) {
		err = fmt.Errorf("unexpected! toVal is nil ")
	}

	if util.IsNil(fromVal) {
		ret = toVal
		return
	}

	isPtr := toVal.Kind() == reflect.Ptr
	toVal = reflect.Indirect(toVal)

	toType := toVal.Type()
	switch toType.Kind() {
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
		case reflect.String:
			strVal := fromVal.String()
			if strVal == "1" {
				toVal.SetBool(true)
			} else if strVal == "0" {
				toVal.SetBool(false)
			} else {
				err = fmt.Errorf("illegal bool value, fromVal:%v", fromVal.Interface())
			}
		default:
			err = fmt.Errorf("illegal bool value, fromType:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		switch fromVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			toVal.SetInt(fromVal.Int())
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			toVal.SetInt(int64(fromVal.Uint()))
		case reflect.Float32, reflect.Float64:
			toVal.SetInt(int64(fromVal.Float()))
		case reflect.String:
			iVal, iErr := strconv.Atoi(fromVal.String())
			if iErr == nil {
				toVal.SetInt(int64(iVal))
			} else {
				err = fmt.Errorf("illegal int value, fromVal:%v, err:%s", fromVal.String(), iErr.Error())
			}
		default:
			err = fmt.Errorf("illegal int value, fromVal type:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		switch fromVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			toVal.SetUint(uint64(fromVal.Int()))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			toVal.SetUint(fromVal.Uint())
		case reflect.Float32, reflect.Float64:
			toVal.SetUint(uint64(fromVal.Float()))
		case reflect.String:
			iVal, iErr := strconv.Atoi(fromVal.String())
			if iErr == nil {
				toVal.SetUint(uint64(iVal))
			} else {
				err = fmt.Errorf("illegal uint value, fromVal:%v, err:%s", fromVal.String(), iErr.Error())
			}
		default:
			err = fmt.Errorf("illegal uint value, fromVal type:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
		}
	case reflect.Float32, reflect.Float64:
		switch fromVal.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			toVal.SetFloat(float64(fromVal.Int()))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
			toVal.SetFloat(float64(fromVal.Uint()))
		case reflect.Float32, reflect.Float64:
			toVal.SetFloat(fromVal.Float())
		case reflect.String:
			fVal, fErr := strconv.ParseFloat(fromVal.String(), 64)
			if fErr == nil {
				toVal.SetFloat(fVal)
			} else {
				err = fmt.Errorf("illegal float value, fromVal:%v, err:%s", fromVal.String(), fErr.Error())
			}
		default:
			err = fmt.Errorf("illegal float value, fromVal type:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
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
				err = fmt.Errorf("illegal string value, fromVal type:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
			}
		case reflect.String:
			toVal.SetString(fromVal.String())
		case reflect.Slice:
			var array []string
			for idx := 0; idx < fromVal.Len(); idx++ {
				str := ""
				val := reflect.ValueOf(str)
				val, err = AssignValue(fromVal.Index(idx), val)
				if err != nil {
					err = fmt.Errorf("convert to string failed, err:%s", err.Error())
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
			if fromVal.String() != "" {
				dtVal, dtErr := time.ParseInLocation("2006-01-02 15:04:05", fromVal.String(), time.Local)
				if dtErr != nil {
					err = fmt.Errorf("illegal datetime value, err:%s", dtErr.Error())
				} else {
					toVal.Set(reflect.ValueOf(dtVal))
				}
			}
		case reflect.Struct:
			toVal.Set(fromVal)
		default:
			err = fmt.Errorf("illegal datetime value, fromVal type:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
		}
	case reflect.Slice:
		toVal, err = AssignSliceValue(fromVal, toVal)
	case reflect.Ptr:
		if fromVal.Type().String() == toVal.Type().String() {
			toVal.Set(fromVal)
		} else {
			err = fmt.Errorf("illegal ptr value, fromVal type:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
		}
	default:
		err = fmt.Errorf("illegal field type, fromVal type:%s, fromVal:%v", fromVal.Type().String(), fromVal.Interface())
	}

	if err != nil {
		return
	}

	if isPtr {
		ret = toVal.Addr()
		return
	}

	ret = toVal
	return
}

// AssignSliceValue assign interface{} slice to reflect.Value
// fromSliceVal -> toSliceVal
// fromVal -> *toSliceVal
func AssignSliceValue(fromVal reflect.Value, toVal reflect.Value) (ret reflect.Value, err error) {
	if util.IsNil(toVal) {
		err = fmt.Errorf("unexpected! toVal is nil ")
		return
	}

	if util.IsNil(fromVal) {
		ret = toVal
		return
	}

	isPtr := toVal.Kind() == reflect.Ptr
	toVal = reflect.Indirect(toVal)

	if fromVal.Kind() == reflect.Interface {
		fromVal = fromVal.Elem()
	}

	fromVal = reflect.Indirect(fromVal)
	if fromVal.Kind() == reflect.String {
		var array []string
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
		iv, err = AssignValue(v, iv)
		if err != nil {
			return
		}

		itemSlice = reflect.Append(itemSlice, iv)
	}

	toVal.Set(itemSlice)

	if isPtr {
		ret = toVal.Addr()
		return
	}

	ret = toVal
	return
}
