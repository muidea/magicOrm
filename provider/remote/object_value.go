package remote

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

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

func getItemValue(fieldName string, itemType *TypeImpl, fieldValue reflect.Value) (ret *ItemValue, err error) {
	if util.IsNil(fieldValue) {
		ret = &ItemValue{Name: fieldName}
		return
	}

	switch itemType.GetValue() {
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
		err = fmt.Errorf("illegal item type, type:%s", itemType.GetName())
	}

	return
}

func getSliceItemValue(fieldName string, itemType *TypeImpl, fieldValue reflect.Value) (ret *ItemValue, err error) {
	sliceVal := []interface{}{}
	ret = &ItemValue{Name: fieldName}
	if util.IsNil(fieldValue) {
		return
	}

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
			objVal, objErr := GetObjectValue(itemVal.Interface())
			if objErr != nil {
				err = objErr
				return
			}

			sliceVal = append(sliceVal, objVal)
		case util.TypeSliceField:
			err = fmt.Errorf("illegal slice item type, type:%s", itemType.GetName())
		default:
			err = fmt.Errorf("illegal slice item type, type:%s", itemType.GetName())
		}

		if err != nil {
			return
		}
	}

	ret.Value = sliceVal

	return
}

// GetObjectValue get object value
func GetObjectValue(obj interface{}) (ret *ObjectValue, err error) {
	ret = &ObjectValue{Items: []ItemValue{}}
	objValue := reflect.Indirect(reflect.ValueOf(obj))

	objType := objValue.Type()
	ret.TypeName = objType.String()
	ret.PkgPath = objType.PkgPath()

	fieldNum := objValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := objType.Field(idx)
		itemType, itemErr := GetType(fieldType.Type)
		if itemErr != nil {
			err = itemErr
			log.Printf("GetType faield, type%s, err:%s", fieldType.Type.String(), err.Error())
			return
		}

		fieldValue := reflect.Indirect(objValue.Field(idx))
		if itemType.GetValue() != util.TypeSliceField {
			val, valErr := getItemValue(fieldType.Name, itemType, fieldValue)
			if valErr != nil {
				err = valErr
				log.Printf("getItemValue faield, type%s, err:%s", fieldType.Type.String(), err.Error())
				return
			}
			ret.Items = append(ret.Items, *val)
		} else {
			itemType, itemErr = GetType(itemType.GetType().Elem())
			if itemErr != nil {
				err = itemErr
				log.Printf("GetType faield, err:%s", err.Error())
				return
			}
			val, valErr := getSliceItemValue(fieldType.Name, itemType, fieldValue)
			if valErr != nil {
				err = valErr
				log.Printf("getSliceItemValue faield, type%s, err:%s", fieldType.Type.String(), err.Error())
				return
			}
			ret.Items = append(ret.Items, *val)
		}
	}

	return
}

func convertStructValue(structObj reflect.Value, structVal *reflect.Value) (err error) {
	if util.IsNil(structObj) || util.IsNil(*structVal) {
		return
	}

	structObj = reflect.Indirect(structObj)
	objVal, objOK := structObj.Interface().(ObjectValue)
	if !objOK {
		err = fmt.Errorf("illegal struct value, value type:%s", structObj.Type().String())
		return
	}

	structType := structVal.Type()
	fieldNum := structVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		itemType, itemErr := GetType(structType.Field(idx).Type)
		if itemErr != nil {
			err = itemErr
			return
		}

		fieldValue := reflect.Indirect(structVal.Field(idx))
		itemValue := reflect.ValueOf(objVal.Items[idx].Value)

		dependType := itemType.Depend()
		if dependType == nil || util.IsBasicType(dependType.GetValue()) {
			if itemType.GetValue() != util.TypeSliceField {
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
			if itemType.GetValue() != util.TypeSliceField {
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

func convertSliceValue(sliceObj reflect.Value, sliceVal *reflect.Value) (err error) {
	if util.IsNil(sliceObj) || util.IsNil(*sliceVal) {
		return
	}

	rawType := sliceVal.Type().Elem()
	itemType, itemErr := GetType(rawType)
	if itemErr != nil {
		err = itemErr
		return
	}

	if itemType.GetValue() == util.TypeSliceField {
		err = fmt.Errorf("illegal slice element type")
		return
	}

	if itemType.IsPtrType() {
		rawType = rawType.Elem()
	}

	for idx := 0; idx < sliceObj.Len(); idx++ {
		itemObj := sliceObj.Index(idx)
		itemVal := reflect.New(rawType).Elem()

		dependType := itemType.Depend()
		if dependType != nil && !util.IsBasicType(dependType.GetValue()) {
			valErr := convertStructValue(itemObj, &itemVal)
			if valErr != nil {
				err = valErr
				return
			}
		} else {
			valErr := helper.ConvertValue(itemObj, &itemVal)
			if valErr != nil {
				err = valErr
				return
			}
		}

		if itemType.IsPtrType() {
			itemVal = itemVal.Addr()
		}

		*sliceVal = reflect.Append(*sliceVal, itemVal)
	}

	return
}

// UpdateObject update object value -> obj
func UpdateObject(objectValue *ObjectValue, obj interface{}) (err error) {
	objValue := reflect.Indirect(reflect.ValueOf(obj))
	objType := objValue.Type()

	itemType, itemErr := GetType(objType)
	if itemErr != nil {
		err = itemErr
		return
	}

	if itemType.GetName() != objectValue.GetName() || itemType.GetPkgPath() != objectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, objectValue name:%s, objType name:%s", objectValue.GetName(), itemType.GetName())
		return
	}

	fieldNum := objValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		itemType, itemErr := GetType(objType.Field(idx).Type)
		if itemErr != nil {
			err = itemErr
			return
		}

		fieldValue := reflect.Indirect(objValue.Field(idx))
		itemValue := reflect.ValueOf(objectValue.Items[idx].Value)

		dependType := itemType.Depend()
		if dependType == nil || util.IsBasicType(dependType.GetValue()) {
			if itemType.GetValue() != util.TypeSliceField {
				valErr := helper.ConvertValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					log.Printf("convert basic field value failed, name:%s, err:%s", itemType.GetName(), valErr.Error())
					return
				}
			} else {
				valErr := helper.ConvertSliceValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					log.Printf("convert basic slice field value failed, name:%s, err:%s", itemType.GetName(), valErr.Error())
					return
				}
			}
		} else {
			if itemType.GetValue() != util.TypeSliceField {
				valErr := convertStructValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					log.Printf("convert struct field value failed, name:%s, err:%s", itemType.GetName(), valErr.Error())
					return
				}
			} else {
				valErr := convertSliceValue(itemValue, &fieldValue)
				if valErr != nil {
					err = valErr
					log.Printf("convert struct slice field value failed, name:%s, err:%s", itemType.GetName(), valErr.Error())
					return
				}

				if !util.IsNil(fieldValue) {
					log.Print(fieldValue.Interface())
				}
			}
		}
	}

	log.Print(objValue.Interface())
	log.Print(obj)

	return
}

// EncodeObjectValue encode objectValue
func EncodeObjectValue(objVal *ObjectValue) (ret []byte, err error) {
	ret, err = json.Marshal(objVal)
	return
}

func decodeObjectValue(objVal map[string]interface{}) (ret *ObjectValue, err error) {
	nameVal, nameOK := objVal["typeName"]
	pkgPathVal, pkgPathOK := objVal["pkgPath"]
	itemsVal, itemsOK := objVal["items"]
	if !nameOK || !pkgPathOK || !itemsOK {
		err = fmt.Errorf("illegal ObjectValue")
		return
	}

	ret = &ObjectValue{TypeName: nameVal.(string), PkgPath: pkgPathVal.(string), Items: []ItemValue{}}

	for _, val := range itemsVal.([]interface{}) {
		item, itemOK := val.(map[string]interface{})
		if !itemOK {
			err = fmt.Errorf("illegal object field item value")
			ret = nil
			return
		}

		itemVal, itemErr := decodeItemValue(item)
		if itemErr != nil {
			err = itemErr
			ret = nil
			return
		}

		ret.Items = append(ret.Items, *itemVal)
	}

	return
}

func decodeSliceValue(sliceVal []interface{}) (ret []interface{}, err error) {
	for _, val := range sliceVal {
		itemVal, itemOK := val.(map[string]interface{})
		if itemOK {
			item, itemErr := decodeObjectValue(itemVal)
			if itemErr != nil {
				err = itemErr
				log.Printf("decodeObjectValue failed, itemVal:%v", itemVal)
				return
			}

			ret = append(ret, *item)
			continue
		}

		_, sliceOK := val.([]interface{})
		if sliceOK {
			err = fmt.Errorf("illegal slice item value")
			return
		}

		ret = append(ret, val)
	}

	return
}

func decodeItemValue(itemVal map[string]interface{}) (ret *ItemValue, err error) {
	nameVal, nameOK := itemVal["name"]
	valVal, valOK := itemVal["value"]
	if !nameOK || !valOK {
		err = fmt.Errorf("illegal item value")
	}

	ret = &ItemValue{Name: nameVal.(string), Value: valVal}
	ret, err = decodeItem(ret)
	return
}

func decodeItem(val *ItemValue) (ret *ItemValue, err error) {
	objVal, objOK := val.Value.(map[string]interface{})
	if objOK {
		ret = &ItemValue{Name: val.Name}

		oVal, oErr := decodeObjectValue(objVal)
		if oErr != nil {
			err = oErr
			return
		}

		ret.Value = *oVal
		return
	}

	sliceVal, sliceOK := val.Value.([]interface{})
	if sliceOK {
		ret = &ItemValue{Name: val.Name}
		sVal, sErr := decodeSliceValue(sliceVal)
		if sErr != nil {
			err = sErr
			return
		}

		ret.Value = sVal
		return
	}

	ret = val
	return
}

// DecodeObjectValue decode objectValue
func DecodeObjectValue(data []byte) (ret *ObjectValue, err error) {
	val := &ObjectValue{}
	err = json.Unmarshal(data, val)
	if err != nil {
		return
	}

	for idx := range val.Items {
		cur := &val.Items[idx]

		item, itemErr := decodeItem(cur)
		if itemErr != nil {
			err = itemErr
			return
		}

		cur.Value = item.Value
	}

	ret = val

	return
}
