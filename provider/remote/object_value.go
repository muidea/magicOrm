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
	TypeName  string      `json:"typeName"`
	PkgPath   string      `json:"pkgPath"`
	IsPtrFlag bool        `json:"isPtr"`
	Items     []ItemValue `json:"items"`
}

// SliceObjectValue slice object value
type SliceObjectValue struct {
	TypeName  string `json:"typeName"`
	PkgPath   string `json:"pkgPath"`
	IsPtrFlag bool   `json:"isPtr"`
}

// GetName get object name
func (s *ObjectValue) GetName() string {
	return s.TypeName
}

// GetPkgPath get pkgpath
func (s *ObjectValue) GetPkgPath() string {
	return s.PkgPath
}

// IsPtrValue isPtrValue
func (s *ObjectValue) IsPtrValue() bool {
	return s.IsPtrFlag
}

// GetName get object name
func (s *SliceObjectValue) GetName() string {
	return s.TypeName
}

// GetPkgPath get pkgpath
func (s *SliceObjectValue) GetPkgPath() string {
	return s.PkgPath
}

// IsPtrValue isPtrValue
func (s *SliceObjectValue) IsPtrValue() bool {
	return s.IsPtrFlag
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
		if itemType.IsPtrType() {
			ret = &ItemValue{Name: fieldName, Value: &dtVal}
		} else {
			ret = &ItemValue{Name: fieldName, Value: dtVal}
		}
	case util.TypeStructField:
		objVal, objErr := GetObjectValue(fieldValue.Interface())
		if objErr != nil {
			err = objErr
			return
		}
		if itemType.IsPtrType() {
			ret = &ItemValue{Name: fieldName, Value: objVal}
		} else {
			ret = &ItemValue{Name: fieldName, Value: *objVal}
		}
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

	subItemType, subItemErr := GetType(itemType.GetType().Elem())
	if subItemErr != nil {
		err = subItemErr
		log.Printf("Get subItem Type faield, err:%s", err.Error())
		return
	}

	fieldValue = reflect.Indirect(fieldValue)
	for idx := 0; idx < fieldValue.Len(); idx++ {
		itemVal := fieldValue.Index(idx)
		switch subItemType.GetValue() {
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
			if subItemType.IsPtrType() {
				sliceVal = append(sliceVal, &dtVal)
			} else {
				sliceVal = append(sliceVal, dtVal)
			}
		case util.TypeStructField:
			objVal, objErr := GetObjectValue(itemVal.Interface())
			if objErr != nil {
				err = objErr
				return
			}

			if subItemType.IsPtrType() {
				sliceVal = append(sliceVal, objVal)
			} else {
				sliceVal = append(sliceVal, *objVal)
			}
		case util.TypeSliceField:
			err = fmt.Errorf("illegal slice item type, type:%s", subItemType.GetName())
		default:
			err = fmt.Errorf("illegal slice item type, type:%s", subItemType.GetName())
		}

		if err != nil {
			return
		}
	}

	if itemType.IsPtrType() {
		ret.Value = &sliceVal
	} else {
		ret.Value = sliceVal
	}

	return
}

// GetObjectValue get object value
func GetObjectValue(entity interface{}) (ret *ObjectValue, err error) {
	entityValue := reflect.ValueOf(entity)
	isPtr := entityValue.Kind() == reflect.Ptr
	entityValue = reflect.Indirect(entityValue)
	entityType := entityValue.Type()

	ret = &ObjectValue{TypeName: entityType.String(), PkgPath: entityType.PkgPath(), IsPtrFlag: isPtr, Items: []ItemValue{}}
	fieldNum := entityValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		itemType, itemErr := GetType(fieldType.Type)
		if itemErr != nil {
			err = itemErr
			log.Printf("GetType faield, type%s, err:%s", fieldType.Type.String(), err.Error())
			return
		}

		fieldValue := entityValue.Field(idx)
		if itemType.GetValue() != util.TypeSliceField {
			val, valErr := getItemValue(fieldType.Name, itemType, fieldValue)
			if valErr != nil {
				err = valErr
				log.Printf("getItemValue faield, type%s, err:%s", fieldType.Type.String(), err.Error())
				return
			}
			ret.Items = append(ret.Items, *val)
		} else {
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

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity interface{}) (ret *SliceObjectValue, err error) {
	entityType := reflect.TypeOf(sliceEntity)
	typeImpl, typeErr := GetType(entityType)
	if typeErr != nil {
		err = fmt.Errorf("get slice object type failed, err:%s", err.Error())
		return
	}
	if !util.IsSliceType(typeImpl.GetValue()) {
		err = fmt.Errorf("illegal slice object value")
		return
	}
	subType := typeImpl.Elem()
	if !util.IsStructType(subType.GetValue()) {
		err = fmt.Errorf("illegal slice item type")
		return
	}

	ret = &SliceObjectValue{TypeName: subType.GetName(), PkgPath: subType.GetPkgPath(), IsPtrFlag: subType.IsPtrType()}

	return
}

func convertStructValue(objectValue reflect.Value, entityValue *reflect.Value) (err error) {
	if util.IsNil(objectValue) || util.IsNil(*entityValue) {
		return
	}

	if objectValue.Kind() == reflect.Interface {
		objectValue = objectValue.Elem()
	}

	objectValue = reflect.Indirect(objectValue)
	objVal, objOK := objectValue.Interface().(ObjectValue)
	if !objOK {
		err = fmt.Errorf("illegal struct value, value type:%s", objectValue.Type().String())
		return
	}

	entityType := entityValue.Type()
	fieldNum := entityValue.NumField()
	items := objVal.Items
	for idx := 0; idx < fieldNum; idx++ {
		if items[idx].Value == nil {
			continue
		}

		itemType, itemErr := GetType(entityType.Field(idx).Type)
		if itemErr != nil {
			err = itemErr
			return
		}

		fieldValue := reflect.Indirect(entityValue.Field(idx))
		itemValue := reflect.ValueOf(items[idx].Value)

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

	if sliceObj.Kind() == reflect.Interface {
		sliceObj = sliceObj.Elem()
	}

	sliceObj = reflect.Indirect(sliceObj)
	itemSlice := reflect.MakeSlice(sliceVal.Type(), 0, 0)
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

		itemSlice = reflect.Append(itemSlice, itemVal)
	}

	sliceVal.Set(itemSlice)

	return
}

// UpdateEntity update object value -> entity
func UpdateEntity(objectValue *ObjectValue, entity interface{}) (err error) {
	entityValue := reflect.ValueOf(entity)
	return updateEntity(objectValue, entityValue)
}

func updateEntity(objectValue *ObjectValue, entityValue reflect.Value) (err error) {
	entityValue = reflect.Indirect(entityValue)
	entityType := entityValue.Type()
	itemType, itemErr := GetType(entityType)
	if itemErr != nil {
		err = itemErr
		return
	}

	if itemType.GetName() != objectValue.GetName() || itemType.GetPkgPath() != objectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, objectValue name:%s, entityType name:%s", objectValue.GetName(), itemType.GetName())
		return
	}

	fieldNum := entityValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		itemType, itemErr := GetType(entityType.Field(idx).Type)
		if itemErr != nil {
			err = itemErr
			return
		}

		fieldValue := reflect.Indirect(entityValue.Field(idx))
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
			}
		}
	}

	return
}

// UpdateSliceEntity update object value list -> entitySlice
func UpdateSliceEntity(objectValueSlice interface{}, entitySlice interface{}) (err error) {
	objectValueSliceVal := reflect.ValueOf(objectValueSlice)
	objectValueSliceVal = reflect.Indirect(objectValueSliceVal)

	entitySliceVal := reflect.Indirect(reflect.ValueOf(entitySlice))
	if objectValueSliceVal.Kind() != reflect.Slice || entitySliceVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal objectValueSlice")
		return
	}

	sliceType := entitySliceVal.Type()
	itemType := sliceType.Elem()
	entityType, entityErr := GetType(itemType)
	if entityErr != nil || !util.IsStructType(entityType.GetValue()) {
		err = fmt.Errorf("illegal entity slice value")
		return
	}

	sliceVal := reflect.MakeSlice(sliceType, 0, 0)
	for idx := 0; idx < objectValueSliceVal.Len(); idx++ {
		objVal := reflect.Indirect(objectValueSliceVal.Index(idx))
		if objVal.Kind() == reflect.Interface {
			objVal = objVal.Elem()
		}
		objVal = reflect.Indirect(objVal)
		objEntityVal, objEntityOK := objVal.Interface().(ObjectValue)
		if !objEntityOK {
			err = fmt.Errorf("illegal object slice value")
			return
		}

		entityVal := reflect.New(itemType)
		if !entityType.IsPtrType() {
			entityVal = reflect.Indirect(entityVal)
		}

		err = updateEntity(&objEntityVal, entityVal)
		if err != nil {
			err = fmt.Errorf("updateEntity failed, err:%s", err.Error())
			return
		}

		sliceVal = reflect.Append(sliceVal, entityVal)
	}

	entitySliceVal.Set(sliceVal)

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
	isPtrVal, isPtrOK := objVal["isPtr"]
	itemsVal, itemsOK := objVal["items"]
	if !nameOK || !pkgPathOK || !isPtrOK || !itemsOK {
		err = fmt.Errorf("illegal ObjectValue")
		return
	}

	ret = &ObjectValue{TypeName: nameVal.(string), PkgPath: pkgPathVal.(string), IsPtrFlag: isPtrVal.(bool), Items: []ItemValue{}}

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

			if item.IsPtrFlag {
				ret = append(ret, item)
			} else {
				ret = append(ret, *item)
			}

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

		if oVal.IsPtrFlag {
			ret.Value = oVal
		} else {
			ret.Value = *oVal
		}
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
