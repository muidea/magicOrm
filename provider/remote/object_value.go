package remote

import (
	"fmt"
	"reflect"

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
			fVal, fErr := GetObjectValue(fieldValue.Interface())
			if fErr != nil {
				err = fErr
				return
			}

			ret.Items[fieldType.Name] = fVal
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
