package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// EncodeSliceValue get slice value str
func EncodeSliceValue(val reflect.Value) (ret string, err error) {
	var valSlice []string

	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	if !util.IsNil(val) {
		val = reflect.Indirect(val)
		pos := val.Len()
		for idx := 0; idx < pos; {
			sv := reflect.Indirect(val.Index(idx))
			switch sv.Kind() {
			case reflect.Bool:
				strVal, strErr := EncodeBoolValue(sv)
				if strErr != nil {
					err = strErr
					return
				}
				valSlice = append(valSlice, strVal)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				strVal, strErr := EncodeIntValue(sv)
				if strErr != nil {
					err = strErr
					return
				}
				valSlice = append(valSlice, strVal)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				strVal, strErr := EncodeUintValue(sv)
				if strErr != nil {
					err = strErr
					return
				}
				valSlice = append(valSlice, strVal)
			case reflect.Float32, reflect.Float64:
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
			case reflect.Interface:
				sVal := sv.Interface()
				for {
					_, bOK := sVal.(bool)
					if bOK {
						strVal, strErr := EncodeBoolValue(sv)
						if strErr != nil {
							err = strErr
							return
						}
						valSlice = append(valSlice, strVal)
						break
					}
					_, intOK := sVal.(int64)
					if intOK {
						strVal, strErr := EncodeIntValue(sv)
						if strErr != nil {
							err = strErr
							return
						}
						valSlice = append(valSlice, strVal)
						break
					}
					_, fltOK := sVal.(float64)
					if fltOK {
						strVal, strErr := EncodeFloatValue(sv)
						if strErr != nil {
							err = strErr
							return
						}
						valSlice = append(valSlice, strVal)
						break
					}
					_, dtOK := sVal.(time.Time)
					if dtOK {
						strVal, strErr := EncodeDateTimeValue(sv)
						if strErr != nil {
							err = strErr
							return
						}
						valSlice = append(valSlice, strVal)
						break
					}
					_, strOK := sVal.(string)
					if strOK {
						strVal, strErr := EncodeStringValue(sv)
						if strErr != nil {
							err = strErr
							return
						}
						valSlice = append(valSlice, strVal)
						break
					}

					err = fmt.Errorf("no support slice element val, [%v]", sVal)
					break
				}
			default:
				err = fmt.Errorf("no support slice element type, [%s]", sv.Type().String())
			}

			idx++
		}
	}

	data, dataErr := json.Marshal(valSlice)
	if dataErr != nil {
		err = dataErr
	}
	ret = fmt.Sprintf("%s", string(data))

	return
}

// DecodeSliceValue decode slice from string
func DecodeSliceValue(val string, vType model.Type) (ret reflect.Value, err error) {
	tVal := vType.GetValue()
	switch tVal {
	case util.TypeSliceField:
	default:
		err = fmt.Errorf("illegal slice value type")
		return
	}

	ret = reflect.Indirect(vType.Interface())
	ret, err = ConvertSliceValue(reflect.ValueOf(val), ret)

	if err != nil {
		if vType.IsPtrType() {
			ret = ret.Addr()
		}
	}

	return
}
