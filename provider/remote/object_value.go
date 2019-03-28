package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/util"
)

// ItemValue item value
type ItemValue struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// ObjectValue Object Value
type ObjectValue struct {
	TypeName string      `json:"typeName"`
	PkgPath  string      `json:"pkgPath"`
	Items    []ItemValue `json:"items"`
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
	ret = &ObjectValue{Items: []ItemValue{}}
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
			if typeVal.String() == "time.Time" {
				dtVal, dtErr := helper.EncodeDateTimeValue(fieldValue)
				if dtErr != nil {
					err = dtErr
					return
				}

				item := ItemValue{Name: fieldType.Name, Value: dtVal}
				ret.Items = append(ret.Items, item)
			} else {
				fVal, fErr := GetObjectValue(fieldValue.Interface())
				if fErr != nil {
					err = fErr
					return
				}

				item := ItemValue{Name: fieldType.Name, Value: fVal}
				ret.Items = append(ret.Items, item)
			}
			continue
		}

		if typeVal.Kind() == reflect.Slice {
			sliceVal := []interface{}{}
			for idx := 0; idx < fieldValue.Len(); idx++ {
				itemVal := reflect.Indirect(fieldValue.Index(idx))
				itemType := itemVal.Type()
				_, tErr := util.GetTypeValueEnum(itemType)
				if tErr != nil {
					err = tErr
					return
				}

				if itemType.Kind() == reflect.Slice {
					err = fmt.Errorf("illegal slice item value")
					return
				}

				if itemType.Kind() == reflect.Struct {
					if itemType.String() == "time.Time" {
						dtVal, dtErr := helper.EncodeDateTimeValue(itemVal)
						if dtErr != nil {
							err = dtErr
							return
						}

						sliceVal = append(sliceVal, dtVal)
					} else {
						fVal, fErr := GetObjectValue(itemVal)
						if fErr != nil {
							err = fErr
							return
						}

						sliceVal = append(sliceVal, fVal)
					}

					continue
				}

				sliceVal = append(sliceVal, itemVal.Interface())
			}

			item := ItemValue{Name: fieldType.Name, Value: sliceVal}
			ret.Items = append(ret.Items, item)
			continue
		}

		item := ItemValue{Name: fieldType.Name, Value: fieldValue.Interface()}
		ret.Items = append(ret.Items, item)
	}

	return
}

// getValueStr get value str
func getValueStr(vType model.Type, vVal model.Value, cache Cache) (ret string, err error) {
	if vVal.IsNil() {
		return
	}

	switch vType.GetValue() {
	case util.TypeBooleanField:
		ret, err = helper.EncodeBoolValue(vVal.Get())
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		ret, err = helper.EncodeIntValue(vVal.Get())
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		ret, err = helper.EncodeUintValue(vVal.Get())
	case util.TypeFloatField, util.TypeDoubleField:
		ret, err = helper.EncodeFloatValue(vVal.Get())
	case util.TypeStringField:
		strRet, strErr := helper.EncodeStringValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case util.TypeSliceField:
		strRet, strErr := helper.EncodeSliceValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case util.TypeDateTimeField:
		strRet, strErr := helper.EncodeStringValue(vVal.Get())
		if strErr != nil {
			err = strErr
			return
		}
		ret = fmt.Sprintf("'%s'", strRet)
	case util.TypeStructField:
		//ret, err = encodeStructValue(vVal.Get(), cache)
	default:
		err = fmt.Errorf("illegal value type, type:%v", vType.GetValue())
	}

	return
}
