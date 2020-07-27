package remote

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"

	log "github.com/cihub/seelog"

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
	TypeName  string       `json:"typeName"`
	PkgPath   string       `json:"pkgPath"`
	IsPtrFlag bool         `json:"isPtr"`
	Items     []*ItemValue `json:"items"`
}

// SliceObjectValue slice object value
type SliceObjectValue struct {
	TypeName  string        `json:"typeName"`
	PkgPath   string        `json:"pkgPath"`
	IsPtrFlag bool          `json:"isPtr"`
	Values    []ObjectValue `json:"values"`
}

// SliceObjectPtrValue slice object ptr value
type SliceObjectPtrValue struct {
	TypeName  string         `json:"typeName"`
	PkgPath   string         `json:"pkgPath"`
	IsPtrFlag bool           `json:"isPtr"`
	Values    []*ObjectValue `json:"values"`
}

// GetName get object name
func (s *ObjectValue) GetName() string {
	return s.TypeName
}

// GetPkgPath get pkg path
func (s *ObjectValue) GetPkgPath() string {
	return s.PkgPath
}

// IsPtrValue isPtrValue
func (s *ObjectValue) IsPtrValue() bool {
	return s.IsPtrFlag
}

// IsAssigned is assigned value
func (s *ObjectValue) IsAssigned() (ret bool) {
	ret = false
	for _, val := range s.Items {
		if val.Value == nil {
			continue
		}

		strVal, strOK := val.Value.(string)
		if strOK {
			ret = strVal != ""
			if ret {
				return
			}

			continue
		}

		fltVal, fltOK := val.Value.(float64)
		if fltOK {
			ret = math.Abs(fltVal-0.00000) > 0.00001
			if ret {
				return
			}

			continue
		}

		objVal, objOK := val.Value.(ObjectValue)
		if objOK {
			ret = objVal.IsAssigned()
			if ret {
				return
			}
		}

		sliceObjVal, sliceObjOK := val.Value.([]ObjectValue)
		if sliceObjOK {
			ret = len(sliceObjVal) > 0
			if ret {
				return
			}
		}
		sliceObjPtrVal, sliceObjPtrOK := val.Value.([]*ObjectValue)
		if sliceObjPtrOK {
			ret = len(sliceObjPtrVal) > 0
			if ret {
				return
			}
		}

		ptrObjVal, ptrObjOK := val.Value.(*ObjectValue)
		if ptrObjOK {
			ret = ptrObjVal.IsAssigned()
			if ret {
				return
			}
		}

		ptrSliceObjVal, ptrSliceObjOK := val.Value.(*[]ObjectValue)
		if ptrSliceObjOK {
			ret = len(*ptrSliceObjVal) > 0
			if ret {
				return
			}
		}
		ptrSliceObjPtrVal, ptrSliceObjPtrOK := val.Value.(*[]*ObjectValue)
		if ptrSliceObjPtrOK {
			ret = len(*ptrSliceObjPtrVal) > 0
			if ret {
				return
			}
		}
	}

	return
}

// GetName get object name
func (s *SliceObjectValue) GetName() string {
	return s.TypeName
}

// GetPkgPath get pkg path
func (s *SliceObjectValue) GetPkgPath() string {
	return s.PkgPath
}

// IsPtrValue isPtrValue
func (s *SliceObjectValue) IsPtrValue() bool {
	return s.IsPtrFlag
}

// GetName get object name
func (s *SliceObjectPtrValue) GetName() string {
	return s.TypeName
}

// GetPkgPath get pkg path
func (s *SliceObjectPtrValue) GetPkgPath() string {
	return s.PkgPath
}

// IsPtrValue isPtrValue
func (s *SliceObjectPtrValue) IsPtrValue() bool {
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
			log.Errorf("encode dateTimeValue failed, raw type:%s, err:%s", fieldValue.Type().String(), err.Error())
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
			log.Errorf("GetObjectValue failed, raw type:%s, err:%s", fieldValue.Type().String(), err.Error())
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
	var sliceVal []interface{}
	ret = &ItemValue{Name: fieldName}
	if util.IsNil(fieldValue) {
		return
	}

	subItemType, subItemErr := GetType(itemType.GetType().Elem())
	if subItemErr != nil {
		err = subItemErr
		log.Errorf("Get subItem Type failed, err:%s", err.Error())
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
				log.Errorf("encodeDateTimeValue failed, err:%s", err.Error())
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
				log.Errorf("encodeDateTimeValue failed, err:%s", err.Error())
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
			log.Errorf("getSliceItemValue failed, err:%s", err.Error())
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

	//!! must be String, not Name
	ret = &ObjectValue{TypeName: entityType.String(), PkgPath: entityType.PkgPath(), IsPtrFlag: isPtr, Items: []*ItemValue{}}
	fieldNum := entityValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		itemType, itemErr := GetType(fieldType.Type)
		if itemErr != nil {
			err = itemErr
			log.Errorf("GetType failed, type%s, err:%s", fieldType.Type.String(), err.Error())
			return
		}

		fieldValue := entityValue.Field(idx)
		if itemType.GetValue() != util.TypeSliceField {
			val, valErr := getItemValue(fieldType.Name, itemType, fieldValue)
			if valErr != nil {
				err = valErr
				log.Errorf("getItemValue failed, type%s, err:%s", fieldType.Type.String(), err.Error())
				return
			}
			ret.Items = append(ret.Items, val)
		} else {
			val, valErr := getSliceItemValue(fieldType.Name, itemType, fieldValue)
			if valErr != nil {
				err = valErr
				log.Errorf("getSliceItemValue failed, type%s, err:%s", fieldType.Type.String(), err.Error())
				return
			}
			ret.Items = append(ret.Items, val)
		}
	}

	return
}

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity interface{}) (ret *SliceObjectValue, err error) {
	entityValue := reflect.ValueOf(sliceEntity)
	typeImpl, typeErr := GetType(entityValue.Type())
	if typeErr != nil {
		err = fmt.Errorf("get slice object type failed, err:%s", err.Error())

		log.Errorf("GetType failed, type%s, err:%s", entityValue.Type().String(), err.Error())
		return
	}

	if !util.IsSliceType(typeImpl.GetValue()) {
		err = fmt.Errorf("illegal slice object value")
		log.Errorf("illegal slice type, type%s, err:%s", entityValue.Type().String(), err.Error())
		return
	}

	subType := typeImpl.Elem()
	if !util.IsStructType(subType.GetValue()) {
		err = fmt.Errorf("illegal slice item type")
		log.Errorf("illegal slice elem type, type%s, err:%s", subType.GetName(), err.Error())
		return
	}

	ret = &SliceObjectValue{TypeName: subType.GetName(), PkgPath: subType.GetPkgPath(), IsPtrFlag: subType.IsPtrType(), Values: []ObjectValue{}}
	entityValue = reflect.Indirect(entityValue)
	for idx := 0; idx < entityValue.Len(); idx++ {
		val := entityValue.Index(idx)

		objVal, objErr := GetObjectValue(val.Interface())
		if objErr != nil {
			err = objErr
			log.Errorf("GetObjectValue failed, type%s, err:%s", val.Type().String(), err.Error())
			return
		}

		ret.Values = append(ret.Values, *objVal)
	}

	return
}

// GetSliceObjectPtrValue get slice object ptr value
func GetSliceObjectPtrValue(sliceEntity interface{}) (ret *SliceObjectPtrValue, err error) {
	entityValue := reflect.ValueOf(sliceEntity)
	typeImpl, typeErr := GetType(entityValue.Type())
	if typeErr != nil {
		err = fmt.Errorf("get slice object type failed, err:%s", err.Error())
		log.Errorf("GetType failed, type%s, err:%s", entityValue.Type().String(), err.Error())
		return
	}
	if !util.IsSliceType(typeImpl.GetValue()) {
		err = fmt.Errorf("illegal slice object value")
		log.Errorf("illegal slice type, type%s, err:%s", entityValue.Type().String(), err.Error())
		return
	}
	subType := typeImpl.Elem()
	if !util.IsStructType(subType.GetValue()) {
		err = fmt.Errorf("illegal slice item type")
		log.Errorf("illegal slice elem type, type%s, err:%s", subType.GetName(), err.Error())
		return
	}

	ret = &SliceObjectPtrValue{TypeName: subType.GetName(), PkgPath: subType.GetPkgPath(), IsPtrFlag: subType.IsPtrType(), Values: []*ObjectValue{}}
	entityValue = reflect.Indirect(entityValue)
	for idx := 0; idx < entityValue.Len(); idx++ {
		val := entityValue.Index(idx)

		objVal, objErr := GetObjectValue(val.Interface())
		if objErr != nil {
			err = objErr
			log.Errorf("GetObjectValue failed, type%s, err:%s", val.Type().String(), err.Error())
			return
		}

		ret.Values = append(ret.Values, objVal)
	}

	return
}

func convertStructValue(objectValue reflect.Value, entityValue reflect.Value) (ret reflect.Value, err error) {
	if objectValue.Kind() == reflect.Interface {
		objectValue = objectValue.Elem()
	}

	objectValue = reflect.Indirect(objectValue)
	objVal, objOK := objectValue.Interface().(ObjectValue)
	if !objOK {
		err = fmt.Errorf("illegal struct value, value type:%s", objectValue.Type().String())
		log.Errorf("illegal objectValue, err:%s", err.Error())
		return
	}

	entityType := entityValue.Type()
	fieldNum := entityValue.NumField()
	items := objVal.Items
	for idx := 0; idx < fieldNum; idx++ {
		if items[idx].Value == nil {
			continue
		}

		fieldType := entityType.Field(idx).Type
		itemType, itemErr := GetType(fieldType)
		if itemErr != nil {
			err = itemErr
			log.Errorf("GetType failed, type%s, err:%s", fieldType.String(), err.Error())
			return
		}
		if itemType.IsPtrType() {
			fieldType = fieldType.Elem()
		}

		fieldValue := reflect.New(fieldType).Elem()
		itemValue := reflect.ValueOf(items[idx].Value)

		dependType := itemType.Depend()
		if dependType == nil || util.IsBasicType(dependType.GetValue()) {
			if itemType.GetValue() != util.TypeSliceField {
				fieldValue, err = helper.AssignValue(itemValue, fieldValue)
				if err != nil {
					log.Errorf("assignValue failed, rawType:%s, valType:%s", itemValue.Type().String(), fieldValue.Type().String())
					return
				}
			} else {
				fieldValue, err = helper.AssignSliceValue(itemValue, fieldValue)
				if err != nil {
					log.Errorf("assignSliceValue failed, rawType:%s, valType:%s", itemValue.Type().String(), fieldValue.Type().String())
					return
				}
			}
		} else {
			if itemType.GetValue() != util.TypeSliceField {
				fieldValue, err = convertStructValue(itemValue, fieldValue)
				if err != nil {
					log.Errorf("convertStructValue failed, rawType:%s, valType:%s", itemValue.Type().String(), fieldValue.Type().String())
					return
				}
			} else {
				fieldValue, err = convertSliceValue(itemValue, fieldValue)
				if err != nil {
					log.Errorf("convertSliceValue failed, rawType:%s, valType:%s", itemValue.Type().String(), fieldValue.Type().String())
					return
				}
			}
		}

		if itemType.IsPtrType() {
			fieldValue = fieldValue.Addr()
		}

		entityValue.Field(idx).Set(fieldValue)
	}

	ret = entityValue

	return
}

func convertSliceValue(sliceObj reflect.Value, sliceVal reflect.Value) (ret reflect.Value, err error) {
	rawType := sliceVal.Type().Elem()
	itemType, itemErr := GetType(rawType)
	if itemErr != nil {
		err = itemErr
		log.Errorf("GetType failed, type%s, err:%s", rawType.String(), err.Error())
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
			itemVal, err = convertStructValue(itemObj, itemVal)
			if err != nil {
				log.Errorf("convertStructValue failed, rawType:%s, valType:%s", itemObj.Type().String(), itemVal.Type().String())
				return
			}
		} else {
			itemVal, err = helper.AssignValue(itemObj, itemVal)
			if err != nil {
				log.Errorf("AssignValue failed, rawType:%s, valType:%s", itemObj.Type().String(), itemVal.Type().String())
				return
			}
		}

		if itemType.IsPtrType() {
			itemVal = itemVal.Addr()
		}

		itemSlice = reflect.Append(itemSlice, itemVal)
	}

	sliceVal.Set(itemSlice)
	ret = sliceVal

	return
}

// UpdateEntity update object value -> entity
func UpdateEntity(objectValue *ObjectValue, entity interface{}) (err error) {
	entityValue := reflect.ValueOf(entity)
	return updateEntity(objectValue, entityValue)
}

func updateEntity(objectValue *ObjectValue, entityValue reflect.Value) (err error) {
	entityValue = reflect.Indirect(entityValue)
	entityType, entityErr := GetType(entityValue.Type())
	if entityErr != nil {
		err = entityErr
		log.Errorf("GetType failed, type%s, err:%s", entityValue.Type().String(), err.Error())
		return
	}

	if entityType.GetName() != objectValue.GetName() || entityType.GetPkgPath() != objectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, objectValue name:%s, entityType name:%s", objectValue.GetName(), entityType.GetName())
		log.Error(err)
		return
	}

	for idx := 0; idx < entityValue.NumField(); idx++ {
		itemName := reflect.ValueOf(objectValue.Items[idx].Name)
		itemValue := reflect.ValueOf(objectValue.Items[idx].Value)
		fieldType, fieldErr := GetType(entityValue.Field(idx).Type())
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("GetType failed, type%s, err:%s", entityValue.Type().String(), err.Error())
			return
		}

		fieldValue := fieldType.Interface()

		log.Infof("name:%s,itemValue type:%s, value:%v, fieldValue type:%s", itemName, itemValue.Type().String(), itemValue.Interface(), fieldValue.Type().String())

		dependType := fieldType.Depend()
		if dependType == nil {
			if fieldType.GetValue() != util.TypeSliceField {
				fieldValue, err = helper.AssignValue(itemValue, fieldValue)
				if err != nil {
					log.Errorf("convert basic field value failed, name:%s, err:%s", fieldType.GetName(), err.Error())
					return
				}
			} else {
				fieldValue, err = helper.AssignSliceValue(itemValue, fieldValue)
				if err != nil {
					log.Errorf("convert basic slice field value failed, name:%s, err:%s", fieldType.GetName(), err.Error())
					return
				}
			}
		} else {
			if fieldType.GetValue() != util.TypeSliceField {
				fieldValue, err = convertStructValue(itemValue, fieldValue)
				if err != nil {
					log.Errorf("convert struct field value failed, name:%s, err:%s", fieldType.GetName(), err.Error())
					return
				}
			} else {
				fieldValue, err = convertSliceValue(itemValue, fieldValue)
				if err != nil {
					log.Errorf("convert struct slice field value failed, name:%s, err:%s", fieldType.GetName(), err.Error())
					return
				}
			}
		}

		log.Info(fieldValue.Interface())

		entityValue.Field(idx).Set(fieldValue)
	}

	return
}

// UpdateSliceEntity update object value list -> entitySlice
func UpdateSliceEntity(sliceObjectValue *SliceObjectValue, entitySlice interface{}) (err error) {
	entitySliceVal := reflect.Indirect(reflect.ValueOf(entitySlice))
	if entitySliceVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal objectValueSlice")
		return
	}
	if !entitySliceVal.CanSet() {
		err = fmt.Errorf("illegal entitySlice value, can't be set")
		return
	}

	sliceType := entitySliceVal.Type()
	itemType := sliceType.Elem()
	entityType, entityErr := GetType(itemType)
	if entityErr != nil || !util.IsStructType(entityType.GetValue()) || entityType.IsPtrType() {
		err = fmt.Errorf("illegal entity slice value")
		return
	}

	if entityType.GetName() != sliceObjectValue.GetName() || entityType.GetPkgPath() != sliceObjectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, objectValue name:%s, entityType name:%s", sliceObjectValue.GetName(), entityType.GetName())
		return
	}

	sliceVal := reflect.MakeSlice(sliceType, 0, 0)
	for idx := 0; idx < len(sliceObjectValue.Values); idx++ {
		objEntityVal := sliceObjectValue.Values[idx]
		entityVal := reflect.New(itemType).Elem()

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

// UpdateSlicePtrEntity update object value list -> ptrEntitySlice
func UpdateSlicePtrEntity(sliceObjectValue *SliceObjectPtrValue, entitySlice interface{}) (err error) {
	entitySliceVal := reflect.Indirect(reflect.ValueOf(entitySlice))
	if entitySliceVal.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal objectValueSlice")
		return
	}

	if !entitySliceVal.CanSet() {
		err = fmt.Errorf("illegal entitySlice value, can't be set")
		return
	}

	sliceType := entitySliceVal.Type()
	itemType := sliceType.Elem()
	entityType, entityErr := GetType(itemType)
	if entityErr != nil || !util.IsStructType(entityType.GetValue()) || !entityType.IsPtrType() {
		err = fmt.Errorf("illegal entity slice value")
		return
	}

	if entityType.GetName() != sliceObjectValue.GetName() || entityType.GetPkgPath() != sliceObjectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, objectValue name:%s, entityType name:%s", sliceObjectValue.GetName(), entityType.GetName())
		return
	}

	itemType = itemType.Elem()
	sliceVal := reflect.MakeSlice(sliceType, 0, 0)
	for idx := 0; idx < len(sliceObjectValue.Values); idx++ {
		objEntityVal := sliceObjectValue.Values[idx]
		entityVal := reflect.New(itemType)

		err = updateEntity(objEntityVal, entityVal)
		if err != nil {
			err = fmt.Errorf("updateEntity failed, err:%s", err.Error())
			return
		}

		sliceVal = reflect.Append(sliceVal, entityVal)
	}

	entitySliceVal.Set(sliceVal)

	return
}

// EncodeObjectValue encode objectValue to []byte
func EncodeObjectValue(objVal *ObjectValue) (ret []byte, err error) {
	ret, err = json.Marshal(objVal)
	return
}

// EncodeSliceObjectValue encode slice objectValue to []byte
func EncodeSliceObjectValue(objVal *SliceObjectValue) (ret []byte, err error) {
	ret, err = json.Marshal(objVal)
	return
}

// EncodeSliceObjectPtrValue encode slice objectPtrValue to []byte
func EncodeSliceObjectPtrValue(objVal *SliceObjectPtrValue) (ret []byte, err error) {
	ret, err = json.Marshal(objVal)
	return
}

// DecodeObjectValueFromMap decode object value from map
func DecodeObjectValueFromMap(objVal map[string]interface{}) (ret *ObjectValue, err error) {
	nameVal, nameOK := objVal["typeName"]
	pkgPathVal, pkgPathOK := objVal["pkgPath"]
	isPtrVal, isPtrOK := objVal["isPtr"]
	itemsVal, itemsOK := objVal["items"]
	if !nameOK || !pkgPathOK || !isPtrOK || !itemsOK {
		err = fmt.Errorf("illegal ObjectValue")
		return
	}

	ret = &ObjectValue{TypeName: nameVal.(string), PkgPath: pkgPathVal.(string), IsPtrFlag: isPtrVal.(bool), Items: []*ItemValue{}}

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

		ret.Items = append(ret.Items, itemVal)
	}

	return
}

func decodeSliceValue(sliceVal []interface{}) (ret []interface{}, err error) {
	for _, val := range sliceVal {
		itemVal, itemOK := val.(map[string]interface{})
		if itemOK {
			item, itemErr := DecodeObjectValueFromMap(itemVal)
			if itemErr != nil {
				err = itemErr
				log.Errorf("DecodeObjectValueFromMap failed, itemVal:%v", itemVal)
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
	ret, err = ConvertItem(ret)
	return
}

// ConvertItem convert ItemValue
func ConvertItem(val *ItemValue) (ret *ItemValue, err error) {
	objVal, objOK := val.Value.(map[string]interface{})
	if objOK {
		ret = &ItemValue{Name: val.Name}

		oVal, oErr := DecodeObjectValueFromMap(objVal)
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
		cur := val.Items[idx]

		item, itemErr := ConvertItem(cur)
		if itemErr != nil {
			err = itemErr
			return
		}

		cur.Value = item.Value
	}

	ret = val

	return
}

// ConvertObjectValue convert object value
func ConvertObjectValue(objVal *ObjectValue) (ret *ObjectValue, err error) {
	for idx := range objVal.Items {
		cur := objVal.Items[idx]

		item, itemErr := ConvertItem(cur)
		if itemErr != nil {
			err = itemErr
			return
		}

		cur.Value = item.Value
	}

	ret = objVal

	return
}

// DecodeSliceObjectValue decode objectValue
func DecodeSliceObjectValue(data []byte) (ret *SliceObjectValue, err error) {
	sliceVal := &SliceObjectValue{}
	err = json.Unmarshal(data, sliceVal)
	if err != nil {
		return
	}

	for idx := range sliceVal.Values {
		cur := &sliceVal.Values[idx]
		val, valErr := ConvertObjectValue(cur)
		if valErr != nil {
			err = valErr
			return
		}

		sliceVal.Values[idx] = *val
	}

	ret = sliceVal
	return
}

// DecodeSliceObjectPtrValue decode objectValue
func DecodeSliceObjectPtrValue(data []byte) (ret *SliceObjectPtrValue, err error) {
	sliceVal := &SliceObjectPtrValue{}
	err = json.Unmarshal(data, sliceVal)
	if err != nil {
		return
	}

	for idx := range sliceVal.Values {
		cur := sliceVal.Values[idx]
		val, valErr := ConvertObjectValue(cur)
		if valErr != nil {
			err = valErr
			return
		}

		sliceVal.Values[idx] = val
	}

	ret = sliceVal
	return
}

// ConvertSliceObjectValue convert slice object value
func ConvertSliceObjectValue(sliceVal *SliceObjectValue) (ret *SliceObjectValue, err error) {
	for idx := range sliceVal.Values {
		cur := &sliceVal.Values[idx]
		val, valErr := ConvertObjectValue(cur)
		if valErr != nil {
			err = valErr
			return
		}

		sliceVal.Values[idx] = *val
	}

	ret = sliceVal
	return
}

// ConvertSliceObjectPtrValue convert slice object ptr value
func ConvertSliceObjectPtrValue(sliceVal *SliceObjectPtrValue) (ret *SliceObjectPtrValue, err error) {
	for idx := range sliceVal.Values {
		cur := sliceVal.Values[idx]
		val, valErr := ConvertObjectValue(cur)
		if valErr != nil {
			err = valErr
			return
		}

		sliceVal.Values[idx] = val
	}

	ret = sliceVal
	return
}
