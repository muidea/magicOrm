package remote

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// EncodeSliceValue get slice value str
func EncodeSliceValue(val reflect.Value) (ret string, err error) {
	valSlice := []string{}

	rawVal := reflect.Indirect(val)
	pos := rawVal.Len()
	for idx := 0; idx < pos; {
		sv := rawVal.Index(idx)
		sv = reflect.Indirect(sv)
		switch sv.Kind() {
		case reflect.Bool:
			strVal, strErr := EncodeBoolValue(sv)
			if strErr != nil {
				err = strErr
				return
				
			}
			valSlice = append(valSlice, strVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			strVal, strErr := EncodeFloatValue(sv)
			if strErr != nil {
				err = strErr
				return
			}
			valSlice = append(valSlice, strVal)
		case reflect.String:
			strVal, strErr := EncodeStringValue(sv)
			if strErr != nil {
				err = strErr
				return
			}
			valSlice = append(valSlice, strVal)
		case reflect.Struct:
			if sv.Type().String() == "time.Time" {
				strVal, strErr := EncodeDateTimeValue(sv)
				if strErr != nil {
					err = strErr
					return
				}
				valSlice = append(valSlice, strVal)
			} else {
				err = fmt.Errorf("no support slice element type, [%s]", sv.Type().String())
			}
		default:
			err = fmt.Errorf("no support slice element type, [%s]", sv.Type().String())
		}

		idx++
	}

	data, dataErr := json.Marshal(valSlice)
	if dataErr != nil {
		err = dataErr
	}
	ret = fmt.Sprintf("%s", string(data))

	return
}

func DecodeSliceValue(val string, vType model.Type) (ret reflect.Value, err error) {
	log.Print(val)

	if vType.GetType().Kind() != reflect.Slice {
		err = fmt.Errorf("illegal value type")
		return
	}

	array := []string{}
	err = json.Unmarshal([]byte(val), &array)
	if err != nil {
		return
	}

	ret = reflect.Indirect(vType.Interface())
	iType := vType.Elem()
	for idx := 0; idx < len(array); idx++ {
		val := array[idx]
		switch iType.GetType().Kind() {
		case reflect.Bool:
			itemVal, itemErr := DecodeBoolValue(val, iType)
			if itemErr != nil {
				err = itemErr
				return
			}
			ret = reflect.Append(ret, itemVal)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			itemVal, itemErr := DecodeFloatValue(val, iType)
			if itemErr != nil {
				err = itemErr
				return
			}
			ret = reflect.Append(ret, itemVal)
		case reflect.String:
			itemVal, itemErr := DecodeStringValue(val, iType)
			if itemErr != nil {
				err = itemErr
				return
			}
			ret = reflect.Append(ret, itemVal)
		case reflect.Struct:
			if iType.GetType().String() == "time.Time" {
				itemVal, itemErr := DecodeDateTimeValue(val, iType)
				if itemErr != nil {
					err = itemErr
					return
				}
				ret = reflect.Append(ret, itemVal)
			} else {
				err = fmt.Errorf("illegal value type, type:%s, expect time.Time", iType.GetType().String())
			}
		default:
			err = fmt.Errorf("illegal value type, unexpect type:%s", iType.GetType().String())
		}
	}

	if vType.IsPtrType() {
		ret = ret.Addr()
	}

	return
}
