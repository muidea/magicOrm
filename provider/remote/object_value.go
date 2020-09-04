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
	Name    string       `json:"name"`
	PkgPath string       `json:"pkgPath"`
	IsPtr   bool         `json:"isPtr"`
	Items   []*ItemValue `json:"items"`
}

// SliceObjectValue slice object value
type SliceObjectValue struct {
	Name    string         `json:"name"`
	PkgPath string         `json:"pkgPath"`
	IsPtr   bool           `json:"isPtr"`
	Values  []*ObjectValue `json:"values"`
}

// GetName get object name
func (s *ObjectValue) GetName() string {
	return s.Name
}

// GetPkgPath get pkg path
func (s *ObjectValue) GetPkgPath() string {
	return s.PkgPath
}

// IsPtrValue isPtrValue
func (s *ObjectValue) IsPtrValue() bool {
	return s.IsPtr
}

// IsAssigned is assigned value
func (s *ObjectValue) IsAssigned() (ret bool) {
	ret = false
	for _, val := range s.Items {
		if val.Value == nil {
			continue
		}

		bVal, bOK := val.Value.(bool)
		if bOK {
			ret = bVal
			if ret {
				return
			}

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

		sliceObjPtrVal, sliceObjPtrOK := val.Value.(*SliceObjectValue)
		if sliceObjPtrOK {
			ret = len(sliceObjPtrVal.Values) > 0
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
	}

	return
}

// GetName get object name
func (s *SliceObjectValue) GetName() string {
	return s.Name
}

// GetPkgPath get pkg path
func (s *SliceObjectValue) GetPkgPath() string {
	return s.PkgPath
}

// IsPtrValue isPtrValue
func (s *SliceObjectValue) IsPtrValue() bool {
	return s.IsPtr
}

func getFieldValue(fieldName string, itemType *TypeImpl, fieldValue reflect.Value) (ret *ItemValue, err error) {
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

		ret = &ItemValue{Name: fieldName, Value: dtVal}
	case util.TypeStructField:
		objVal, objErr := GetObjectValue(fieldValue.Interface())
		if objErr != nil {
			err = objErr
			log.Errorf("GetObjectValue failed, raw type:%s, err:%s", fieldValue.Type().String(), err.Error())
			return
		}

		ret = &ItemValue{Name: fieldName, Value: objVal}
	default:
		err = fmt.Errorf("illegal item type, type:%s", itemType.GetName())
	}

	return
}

func getSliceFieldValue(fieldName string, itemType *TypeImpl, fieldValue reflect.Value) (ret *ItemValue, err error) {
	var sliceVal []interface{}
	var sliceObjectVal []*ObjectValue
	ret = &ItemValue{Name: fieldName}
	if util.IsNil(fieldValue) {
		return
	}

	dependType := itemType.Depend()
	fieldValue = reflect.Indirect(fieldValue)
	for idx := 0; idx < fieldValue.Len(); idx++ {
		itemVal := fieldValue.Index(idx)
		if util.IsNil(itemVal) {
			continue
		}

		itemVal = reflect.Indirect(itemVal)
		switch dependType.GetValue() {
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
			sliceVal = append(sliceVal, dtVal)
		case util.TypeStructField:
			objVal, objErr := GetObjectValue(itemVal.Interface())
			if objErr != nil {
				err = objErr
				log.Errorf("encodeDateTimeValue failed, err:%s", err.Error())
				return
			}

			sliceObjectVal = append(sliceObjectVal, objVal)
		case util.TypeSliceField:
			err = fmt.Errorf("illegal slice item type, type:%s", dependType.GetName())
		default:
			err = fmt.Errorf("illegal slice item type, type:%s", dependType.GetName())
		}

		if err != nil {
			log.Errorf("getSliceFieldValue failed, err:%s", err.Error())
			return
		}
	}

	if util.IsStructType(dependType.GetValue()) {
		ret.Value = &SliceObjectValue{Name: dependType.GetName(), PkgPath: dependType.GetPkgPath(), IsPtr: itemType.IsPtrType(), Values: sliceObjectVal}
		return
	}

	ret.Value = sliceVal
	return
}

// GetObjectValue get object value
func GetObjectValue(entity interface{}) (ret *ObjectValue, err error) {
	objectValue := reflect.ValueOf(entity)
	objectValue = reflect.Indirect(objectValue)
	objectType := objectValue.Type()

	entityType, entityErr := newType(objectType)
	if entityErr != nil {
		err = entityErr
		return
	}
	if !util.IsStructType(entityType.GetValue()) {
		err = fmt.Errorf("illegal entity value")
		return
	}

	//!! must be String, not Name
	ret = &ObjectValue{Name: entityType.GetName(), PkgPath: entityType.GetPkgPath(), IsPtr: entityType.IsPtrType(), Items: []*ItemValue{}}
	fieldNum := objectValue.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldValue := objectValue.Field(idx)
		fieldType := objectType.Field(idx)

		itemType, itemErr := newType(fieldType.Type)
		if itemErr != nil {
			err = itemErr
			log.Errorf("GetType failed, type%s, err:%s", fieldType.Type.String(), err.Error())
			return
		}

		if itemType.GetValue() != util.TypeSliceField {
			val, valErr := getFieldValue(fieldType.Name, itemType, fieldValue)
			if valErr != nil {
				err = valErr
				log.Errorf("getFieldValue failed, type%s, err:%s", fieldType.Type.String(), err.Error())
				return
			}
			ret.Items = append(ret.Items, val)
		} else {
			val, valErr := getSliceFieldValue(fieldType.Name, itemType, fieldValue)
			if valErr != nil {
				err = valErr
				log.Errorf("getSliceFieldValue failed, type%s, err:%s", fieldType.Type.String(), err.Error())
				return
			}
			ret.Items = append(ret.Items, val)
		}
	}

	return
}

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity interface{}) (ret *SliceObjectValue, err error) {
	sliceValue := reflect.ValueOf(sliceEntity)
	sliceType, sliceErr := newType(sliceValue.Type())
	if sliceErr != nil {
		err = fmt.Errorf("get slice object type failed, err:%s", err.Error())
		log.Errorf("GetType failed, type%s, err:%s", sliceType.GetName(), err.Error())
		return
	}

	if !util.IsSliceType(sliceType.GetValue()) {
		err = fmt.Errorf("illegal slice object value")
		log.Errorf("illegal slice type, type%s, err:%s", sliceType.GetName(), err.Error())
		return
	}

	elemType := sliceType.Elem()
	if !util.IsStructType(elemType.GetValue()) {
		err = fmt.Errorf("illegal slice item type")
		log.Errorf("illegal slice elem type, type%s, err:%s", elemType.GetName(), err.Error())
		return
	}

	ret = &SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath(), IsPtr: sliceType.IsPtrType(), Values: []*ObjectValue{}}
	sliceValue = reflect.Indirect(sliceValue)
	for idx := 0; idx < sliceValue.Len(); idx++ {
		val := sliceValue.Index(idx)

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

func convertStructValue(objectValue *ObjectValue, entityValue reflect.Value) (ret reflect.Value, err error) {
	entityType := entityValue.Type()
	fieldNum := entityValue.NumField()

	valueType, valueErr := newType(entityValue.Type())
	if valueErr != nil {
		err = valueErr
		log.Errorf("newType failed, type%s, err:%s", entityValue.Type().String(), err.Error())
		return
	}
	if valueType.GetName() != objectValue.GetName() || valueType.GetPkgPath() != objectValue.GetPkgPath() {
		err = fmt.Errorf("illegal object value, objectValue name:%s, entityType name:%s", objectValue.GetName(), valueType.GetName())
		log.Error(err)
		return
	}

	for idx := 0; idx < fieldNum; idx++ {
		curItem := objectValue.Items[idx]
		if curItem.Value == nil {
			continue
		}

		fieldType := entityType.Field(idx).Type
		isPtr := fieldType.Kind() == reflect.Ptr
		if isPtr {
			fieldType = fieldType.Elem()
		}
		fieldValue := reflect.New(fieldType).Elem()

		tVal, tErr := util.GetTypeValueEnum(fieldType)
		if tErr != nil {
			err = tErr
			log.Errorf("illegal struct field, err:%s", err.Error())
			return
		}

		for {
			if util.IsBasicType(tVal) {
				fieldValue, err = convertBasicItemValue(curItem.Value, fieldValue)
				if err != nil {
					log.Errorf("convertBasicItemValue failed, fieldName:%s", fieldType.Name())
					return
				}
				break
			}

			if util.IsStructType(tVal) {
				fieldValue, err = convertStructItemValue(curItem.Value, fieldValue)
				if err != nil {
					log.Errorf("convertStructItemValue failed, fieldName:%s", fieldType.Name())
					return
				}
				break
			}

			if util.IsSliceType(tVal) {
				fieldValue, err = convertSliceItemValue(curItem.Value, fieldValue)
				if err != nil {
					log.Errorf("convertSliceItemValue failed, fieldName:%s", fieldType.Name())
					return
				}
				break
			}

			err = fmt.Errorf("illegal item type, fieldName:%s, fieldType:%s", fieldType.Name(), fieldType.String())
			return
		}

		if isPtr {
			fieldValue = fieldValue.Addr()
		}

		entityValue.Field(idx).Set(fieldValue)
	}

	ret = entityValue
	return
}

func convertBasicItemValue(itemValue interface{}, fieldValue reflect.Value) (ret reflect.Value, err error) {
	fieldValue, err = helper.AssignValue(reflect.ValueOf(itemValue), fieldValue)
	if err != nil {
		log.Errorf("assignValue failed, valType:%s, err:%s", fieldValue.Type().String(), err.Error())
		return
	}

	ret = fieldValue
	return
}

func convertStructItemValue(itemValue interface{}, fieldValue reflect.Value) (ret reflect.Value, err error) {
	itemObject, ok := itemValue.(*ObjectValue)
	if !ok {
		err = fmt.Errorf("illegal itemValue")
		return
	}

	fieldValue, err = convertStructValue(itemObject, fieldValue)
	if err != nil {
		log.Errorf("convertStructValue failed, valType:%s, err:%s", fieldValue.Type().String(), err.Error())
		return
	}

	ret = fieldValue
	return
}

func convertSliceItemValue(itemValue interface{}, fieldValue reflect.Value) (ret reflect.Value, err error) {
	if fieldValue.Kind() != reflect.Slice {
		err = fmt.Errorf("illegal fieldValue, type:%s", fieldValue.Type().String())
		return
	}

	elemType, elemErr := newType(fieldValue.Type().Elem())
	if elemErr != nil {
		err = elemErr
		return
	}

	if util.IsBasicType(elemType.GetValue()) {
		fieldValue, err = helper.AssignSliceValue(reflect.ValueOf(itemValue), fieldValue)
		if err != nil {
			log.Errorf("assignSliceValue failed, valType:%s", fieldValue.Type().String())
			return
		}

		ret = fieldValue
		return
	}

	if util.IsStructType(elemType.GetValue()) {
		fieldValue, err = convertSliceStructValue(reflect.ValueOf(itemValue), fieldValue)
		if err != nil {
			log.Errorf("convertSliceStructValue failed, valType:%s", fieldValue.Type().String())
			return
		}

		ret = fieldValue
		return
	}

	err = fmt.Errorf("invalid slice element type, element type:%s", elemType.GetName())
	return
}

func convertSliceStructValue(itemValue reflect.Value, fieldValue reflect.Value) (ret reflect.Value, err error) {
	elemType := fieldValue.Type().Elem()
	fieldType, fieldErr := newType(elemType)
	if fieldErr != nil {
		err = fieldErr
		return
	}

	// SliceObjectValue{}
	itemValue = reflect.Indirect(itemValue)
	sliceName := itemValue.FieldByName("Name").String()
	slicePkgPath := itemValue.FieldByName("PkgPath").String()
	if sliceName != fieldType.GetName() || slicePkgPath != fieldType.GetPkgPath() {
		err = fmt.Errorf("illegal slice struct")
		return
	}
	sliceValue := itemValue.FieldByName("Value")
	itemSlice := reflect.MakeSlice(fieldValue.Type(), 0, 0)
	for idx := 0; idx < sliceValue.Len(); idx++ {
		sliceItem := sliceValue.Index(idx)
		if fieldType.IsPtrType() {
			elemType = elemType.Elem()
		}

		elemVal := reflect.New(elemType).Elem()
		for {
			if util.IsBasicType(fieldType.GetValue()) {
				elemVal, err = helper.AssignValue(sliceItem, elemVal)
				if err != nil {
					log.Errorf("AssignValue failed, elemType:%s, valType:%s", sliceItem.Type().String(), elemVal.Type().String())
					return
				}

				break
			}

			if util.IsStructType(fieldType.GetValue()) {
				elemVal, err = convertStructValue(sliceItem.Interface().(*ObjectValue), elemVal)
				if err != nil {
					log.Errorf("convertStructValue failed, elemType:%s, valType:%s", sliceItem.Type().String(), elemVal.Type().String())
					return
				}

				break
			}

			break
		}

		if fieldType.IsPtrType() {
			elemVal = elemVal.Addr()
		}

		itemSlice = reflect.Append(itemSlice, elemVal)
	}

	fieldValue.Set(itemSlice)
	ret = fieldValue

	return
}

// UpdateEntity update object value -> entity
func UpdateEntity(objectValue *ObjectValue, entity interface{}) (err error) {
	entityValue := reflect.ValueOf(entity)
	if entityValue.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal entity value")
		return
	}

	entityValue = reflect.Indirect(entityValue)
	if !entityValue.CanSet() {
		err = fmt.Errorf("illegal entity value, can't be set")
		return
	}

	_, err = convertStructValue(objectValue, entityValue)
	return
}

// UpdateSliceEntity update object value list -> entitySlice
func UpdateSliceEntity(sliceObjectValue *SliceObjectValue, entitySlice interface{}) (err error) {
	entitySliceVal := reflect.ValueOf(entitySlice)
	if entitySliceVal.Kind() != reflect.Ptr {
		err = fmt.Errorf("illegal slice entity value")
		return
	}
	entitySliceVal = reflect.Indirect(entitySliceVal)
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
	entityType, entityErr := newType(itemType)
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

		entityVal, err = convertStructValue(objEntityVal, entityVal)
		if err != nil {
			err = fmt.Errorf("convertStructValue failed, err:%s", err.Error())
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

// decodeObjectValueFromMap decode object value from map
func decodeObjectValueFromMap(objVal map[string]interface{}) (ret *ObjectValue, err error) {
	nameVal, nameOK := objVal["name"]
	pkgPathVal, pkgPathOK := objVal["pkgPath"]
	itemsVal, itemsOK := objVal["items"]
	if !nameOK || !pkgPathOK || !itemsOK {
		err = fmt.Errorf("illegal ObjectValue")
		return
	}

	ret = &ObjectValue{Name: nameVal.(string), PkgPath: pkgPathVal.(string), Items: []*ItemValue{}}

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
			item, itemErr := decodeObjectValueFromMap(itemVal)
			if itemErr != nil {
				err = itemErr
				log.Errorf("decodeObjectValueFromMap failed, itemVal:%v", itemVal)
				return
			}

			ret = append(ret, item)

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
	ret, err = convertItem(ret)
	return
}

// convertItem convert ItemValue
func convertItem(val *ItemValue) (ret *ItemValue, err error) {
	objVal, objOK := val.Value.(map[string]interface{})
	if objOK {
		ret = &ItemValue{Name: val.Name}

		oVal, oErr := decodeObjectValueFromMap(objVal)
		if oErr != nil {
			err = oErr
			return
		}

		ret.Value = oVal
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

		item, itemErr := convertItem(cur)
		if itemErr != nil {
			err = itemErr
			return
		}

		cur.Value = item.Value
	}

	ret = val

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
		cur := sliceVal.Values[idx]
		val, valErr := convertObjectValue(cur)
		if valErr != nil {
			err = valErr
			return
		}

		sliceVal.Values[idx] = val
	}

	ret = sliceVal
	return
}

// convertObjectValue convert object value
func convertObjectValue(objVal *ObjectValue) (ret *ObjectValue, err error) {
	for idx := range objVal.Items {
		cur := objVal.Items[idx]

		item, itemErr := convertItem(cur)
		if itemErr != nil {
			err = itemErr
			return
		}

		cur.Value = item.Value
	}

	ret = objVal

	return
}

// convertSliceObjectValue convert slice object value
func convertSliceObjectValue(sliceVal *SliceObjectValue) (ret *SliceObjectValue, err error) {
	for idx := range sliceVal.Values {
		cur := sliceVal.Values[idx]
		val, valErr := convertObjectValue(cur)
		if valErr != nil {
			err = valErr
			return
		}

		sliceVal.Values[idx] = val
	}

	ret = sliceVal
	return
}

func compareItemValue(l, r *ItemValue) bool {
	if l.Name != r.Name {
		return false
	}

	return true
}

func compareObjectValue(l, r *ObjectValue) bool {
	if l.Name != r.Name {
		return false
	}

	if l.PkgPath != r.PkgPath {
		return false
	}

	if l.IsPtr != r.IsPtr {
		return false
	}

	if len(l.Items) != len(r.Items) {
		return false
	}

	for idx := 0; idx < len(l.Items); idx++ {
		lVal := l.Items[idx]
		rVal := r.Items[idx]
		if !compareItemValue(lVal, rVal) {
			return false
		}
	}

	return true
}

func compareSliceObjectValue(l, r *SliceObjectValue) bool {
	if l.Name != r.Name {
		return false
	}
	if l.PkgPath != r.PkgPath {
		return false
	}
	if l.IsPtr != r.IsPtr {
		return false
	}
	if len(l.Values) != len(r.Values) {
		return false
	}

	for idx := 0; idx < len(l.Values); idx++ {
		lVal := l.Values[idx]
		rVal := r.Values[idx]
		if !compareObjectValue(lVal, rVal) {
			return false
		}
	}

	return true
}
