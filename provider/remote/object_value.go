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

func getItemValue(fieldName string, fieldType *TypeImpl, fieldValue reflect.Value) (ret *ItemValue, err error) {
	switch fieldType.GetValue() {
	case util.TypeBooleanField,
		util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField,
		util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField,
		util.TypeFloatField, util.TypeDoubleField,
		util.TypeStringField:
		ret = &ItemValue{Name: fieldName, Value: fieldValue.Interface()}
	case util.TypeDateTimeField:
		dtVal, dtErr := helper.EncodeDateTimeValue(fieldValue)
		if dtErr != nil {
			err = dtErr
			return
		}
		ret = &ItemValue{Name: fieldName, Value: dtVal}
	case util.TypeStructField:
		objVal, objErr := GetObjectValue(fieldValue.Interface())
		if objErr != nil {
			err = objErr
			return
		}
		ret = &ItemValue{Name: fieldName, Value: objVal}
	default:
		err = fmt.Errorf("illegal item value")
	}

	return
}

func getSliceItemValue(fieldName string, fieldType, itemType *TypeImpl, fieldValue reflect.Value) (ret *ItemValue, err error) {
	switch fieldType.GetValue() {
	case util.TypeSliceField:
		sliceVal := []interface{}{}
		for idx := 0; idx < fieldValue.Len(); idx++ {
			itemVal := reflect.Indirect(fieldValue.Index(idx))
			switch itemType.GetValue() {
			case util.TypeBooleanField,
				util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField,
				util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField,
				util.TypeFloatField, util.TypeDoubleField,
				util.TypeStringField:
				sliceVal = append(sliceVal, itemVal.Interface())
			case util.TypeDateTimeField:
				dtVal, dtErr := helper.EncodeDateTimeValue(itemVal)
				if dtErr != nil {
					err = dtErr
					return
				}

				sliceVal = append(sliceVal, dtVal)
			case util.TypeStructField:
				fVal, fErr := GetObjectValue(itemVal)
				if fErr != nil {
					err = fErr
					return
				}

				sliceVal = append(sliceVal, fVal)
			case util.TypeSliceField:
				err = fmt.Errorf("illegal slice item type")
			default:
				err = fmt.Errorf("illegal slice item type")
			}

			if err != nil {
				return
			}
		}

		ret = &ItemValue{Name: fieldName, Value: sliceVal}
	default:
		err = fmt.Errorf("illegal slice item value")
	}

	return
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
		field := objType.Field(idx)
		fieldType, fieldErr := GetType(field.Type)
		if fieldErr != nil {
			err = fieldErr
			return
		}

		fieldValue := reflect.Indirect(objValue.Field(idx))
		if fieldType.GetValue() != util.TypeSliceField {
			val, valErr := getItemValue(field.Name, fieldType, fieldValue)
			if valErr != nil {
				err = valErr
				return
			}
			ret.Items = append(ret.Items, *val)
		} else {
			itemType, itemErr := GetType(field.Type.Elem())
			if itemErr != nil {
				err = itemErr
				return
			}
			val, valErr := getSliceItemValue(field.Name, fieldType, itemType, fieldValue)
			if valErr != nil {
				err = valErr
				return
			}
			ret.Items = append(ret.Items, *val)
		}
	}

	return
}

func convertStructValue(structObj interface{}, val *reflect.Value) (err error) {
	objVal, objOK := structObj.(ObjectValue)
	if !objOK {
		err = fmt.Errorf("illegal struct value")
		return
	}

	valType := val.Type()
	fieldNum := val.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType, fieldErr := GetType(valType.Field(idx).Type)
		if fieldErr != nil {
			err = fieldErr
			return
		}

		fieldValue := reflect.Indirect(val.Field(idx))
		itemValue := objVal.Items[idx].Value
		dependType := fieldType.Elem()
		if dependType == nil {
			if fieldType.GetValue() != util.TypeSliceField {
				valErr := helper.ConvertValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					return
				}
			} else {
				valErr := helper.ConvertSliceValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					return
				}
			}
		} else {
			if fieldType.GetValue() != util.TypeSliceField {
				valErr := convertStructValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					return
				}
			} else {
				valErr := convertSliceValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					return
				}
			}
		}
	}

	return
}

func convertSliceValue(sliceObj interface{}, val *reflect.Value) (err error) {
	vType := val.Type().Elem()
	itemType, itemErr := GetType(vType)
	if itemErr != nil {
		err = itemErr
		return
	}

	if itemType.GetValue() == util.TypeSliceField {
		err = fmt.Errorf("illegal slice element type")
		return
	}

	sliceVal := sliceObj.([]interface{})
	for _, v := range sliceVal {
		iv := reflect.New(vType).Elem()

		dependType := itemType.Elem()
		if dependType != nil {
			valErr := convertStructValue(v, &iv)
			if valErr != nil {
				err = valErr
				return
			}
		} else {
			valErr := helper.ConvertValue(v, &iv)
			if valErr != nil {
				err = valErr
				return
			}
		}

		if itemType.IsPtrType() {
			iv = iv.Addr()
		}

		*val = reflect.Append(*val, iv)
	}

	return
}

// UpdateObject update object value
func UpdateObject(objectVal *ObjectValue, obj interface{}) (err error) {
	objValue := reflect.Indirect(reflect.ValueOf(obj))

	objType := objValue.Type()
	if objType.Name() != objectVal.GetName() || objType.PkgPath() != objectVal.GetPkgPath() {
		err = fmt.Errorf("illegal object value")
		return
	}

	fieldNum := objValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType, fieldErr := GetType(objType.Field(idx).Type)
		if fieldErr != nil {
			err = fieldErr
			return
		}

		fieldValue := reflect.Indirect(objValue.Field(idx))
		itemValue := objectVal.Items[idx].Value
		dependType := fieldType.Elem()
		if dependType == nil {
			if fieldType.GetValue() != util.TypeSliceField {
				valErr := helper.ConvertValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					return
				}
			} else {
				valErr := helper.ConvertSliceValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					return
				}
			}
		} else {
			if fieldType.GetValue() != util.TypeSliceField {
				valErr := convertStructValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					return
				}
			} else {
				valErr := convertSliceValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					return
				}
			}
		}
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
