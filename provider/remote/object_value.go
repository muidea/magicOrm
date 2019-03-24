package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// ObjectValue Object Value
type ObjectValue struct {
	TypeName string                 `json:"typeName"`
	PkgPath  string                 `json:"pkgPath"`
	Items    map[string]interface{} `json:"items"`
}

// GetName get object name
func (s *ObjectValue) GetName() string {
	return s.TypeName
}

// GetPkgPath get pkgpath
func (s *ObjectValue) GetPkgPath() string {
	return s.PkgPath
}

// GetObjectValue get object value
func GetObjectValue(obj interface{}) (ret *ObjectValue, err error) {
	ret = &ObjectValue{Items: map[string]interface{}{}}
	objValue := reflect.Indirect(reflect.ValueOf(obj))

	objType := objValue.Type()
	ret.TypeName = objType.Name()
	ret.PkgPath = objType.PkgPath()

	fieldNum := objValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := objType.Field(idx)
		typeVal := fieldType.Type
		if typeVal.Kind() == reflect.Ptr {
			typeVal = typeVal.Elem()
		}
		_, tErr := util.GetTypeValueEnum(typeVal)
		if tErr != nil {
			err = tErr
			return
		}

		fieldValue := objValue.Field(idx)
		if typeVal.Kind() == reflect.Struct {
			if typeVal.String() != "time.Time" {
				dtVal, dtErr := encodeDateTimeValue(fieldValue)
				if dtErr != nil {
					err = dtErr
					return
				}

				ret.Items[fieldType.Name] = dtVal
			} else {
				fVal, fErr := GetObjectValue(fieldValue.Interface())
				if fErr != nil {
					err = fErr
					return
				}

				ret.Items[fieldType.Name] = fVal
			}
			continue
		}

		if typeVal.Kind() == reflect.Slice {
			sliceVal := []interface{}{}
			for idx := 0; idx < fieldValue.Len(); idx++ {
				itemVal := reflect.Indirect(fieldValue.Index(idx))
				tVal, tErr := util.GetTypeValueEnum(itemVal.Type())
				if tErr != nil {
					err = tErr
					return
				}

				if util.IsBasicType(tVal) {
					sliceVal = append(sliceVal, itemVal.Interface())
					continue
				}

				if util.IsStructType(tVal) {
					fVal, fErr := GetObjectValue(itemVal)
					if fErr != nil {
						err = fErr
						return
					}
					sliceVal = append(sliceVal, fVal)
					continue
				}

				err = fmt.Errorf("illegal slice item value")
				return
			}

			ret.Items[fieldType.Name] = sliceVal
			continue
		}

		ret.Items[fieldType.Name] = fieldValue.Interface()
	}

	return
}

// getValueStr get value str
func getValueStr(vType model.Type, vVal model.Value, cache Cache) (ret string, err error) {
	if vVal.IsNil() {
		return
	}

	rawType := vType.GetType()
	switch rawType.Kind() {
	case reflect.Bool:
		ret, err = encodeBoolValue(vVal.Get())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Float32, reflect.Float64:
		ret, err = encodeFloatValue(vVal.Get())
	case reflect.String:
		strRet, strErr := encodeStringValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case reflect.Slice:
		strRet, strErr := encodeSliceValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case reflect.Struct:
		if rawType.String() == "time.Time" {
			strRet, strErr := encodeDateTimeValue(vVal.Get())
			if strErr != nil {
				err = strErr
				return
			}
			ret = fmt.Sprintf("'%s'", strRet)
		} else {
			ret, err = encodeStructValue(vVal.Get(), cache)
		}
	default:
		err = fmt.Errorf("illegal value kind, kind:%v", rawType.Kind())
	}

	return
}
