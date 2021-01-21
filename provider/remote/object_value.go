package remote

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// ItemValue item value
type ItemValue struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// ObjectValue Object value
type ObjectValue struct {
	Name    string       `json:"name"`
	PkgPath string       `json:"pkgPath"`
	IsPtr   bool         `json:"isPtr"`
	Items   []*ItemValue `json:"items"`
}

// SliceObjectValue slice object value
type SliceObjectValue struct {
	Name      string         `json:"name"`
	PkgPath   string         `json:"pkgPath"`
	IsPtr     bool           `json:"isPtr"`
	IsElemPtr bool           `json:"isElemPtr"`
	Values    []*ObjectValue `json:"values"`
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
		if val.Value != nil {
			ret = true
			break
		}

		/*
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
		*/
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

// IsElemPtr isPtrValue
func (s *SliceObjectValue) IsElemPtrValue() bool {
	return s.IsElemPtr
}

// IsAssigned is assigned value
func (s *SliceObjectValue) IsAssigned() (ret bool) {
	ret = len(s.Values) > 0
	return
}

func encodeDateTime(val reflect.Value) (ret string, err error) {
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Struct:
		ts, ok := val.Interface().(time.Time)
		if ok {
			ret = fmt.Sprintf("%s", ts.Format("2006-01-02 15:04:05"))
			if ret == "0001-01-01 00:00:00" {
				ret = ""
			}
		} else {
			err = fmt.Errorf("illegal dateTime value, type:%s", val.Type().String())
		}
	default:
		err = fmt.Errorf("illegal dateTime value, type:%s", val.Type().String())
	}

	return
}

func decodeDateTime(val string, tType model.Type) (ret reflect.Value, err error) {
	if util.TypeSliceField == tType.GetValue() {
		if val == "" {
			return
		}

		items := []string{}
		err = json.Unmarshal([]byte(val), &items)
		if err != nil {
			return
		}

		eType := tType.Elem()
		if eType.IsPtrType() {
			dtSliceVal := []*time.Time{}
			for _, v := range items {
				dtVal, dtErr := time.Parse("2006-01-02 15:04:05", v)
				if dtErr != nil {
					err = dtErr
					return
				}

				dtSliceVal = append(dtSliceVal, &dtVal)
			}

			ret = reflect.ValueOf(dtSliceVal)
			return
		}

		dtSliceVal := []time.Time{}
		for _, v := range items {
			dtVal, dtErr := time.Parse("2006-01-02 15:04:05", v)
			if dtErr != nil {
				err = dtErr
				return
			}

			dtSliceVal = append(dtSliceVal, dtVal)
		}
		ret = reflect.ValueOf(dtSliceVal)
		return
	}

	if val == "" {
		val = "0001-01-01 00:00:00"
	}

	dtVal, dtErr := time.Parse("2006-01-02 15:04:05", val)
	if dtErr != nil {
		err = dtErr
		return
	}

	ret = reflect.ValueOf(dtVal)
	return
}

func getFieldValue(fieldName string, itemType *TypeImpl, fieldValue reflect.Value) (ret *ItemValue, err error) {
	if util.IsNil(fieldValue) {
		ret = &ItemValue{Name: fieldName, Value: nil}
		return
	}

	fieldValue = reflect.Indirect(fieldValue)
	if itemType.IsBasic() {
		if util.TypeDateTimeField == itemType.GetValue() {
			dtVal, dtErr := encodeDateTime(fieldValue)
			if dtErr != nil {
				err = dtErr
				log.Errorf("encode dateTime failed, err:%s", err.Error())
				return
			}

			if itemType.Elem().IsPtrType() {
				ret = &ItemValue{Name: fieldName, Value: &dtVal}
				return
			}
			ret = &ItemValue{Name: fieldName, Value: dtVal}
			return
		}

		ret = &ItemValue{Name: fieldName, Value: fieldValue.Interface()}
		return
	}

	objVal, objErr := GetObjectValue(fieldValue.Interface())
	if objErr != nil {
		err = objErr
		log.Errorf("GetObjectValue failed, raw type:%s, err:%s", fieldValue.Type().String(), err.Error())
		return
	}

	ret = &ItemValue{Name: fieldName, Value: objVal}
	return
}

func getSliceFieldValue(fieldName string, itemType *TypeImpl, fieldValue reflect.Value) (ret *ItemValue, err error) {
	ret = &ItemValue{Name: fieldName}
	if util.IsNil(fieldValue) {
		return
	}

	elemType := itemType.Elem()
	if elemType.IsBasic() {
		if util.TypeDateTimeField != elemType.GetValue() {
			ret.Value = fieldValue.Interface()
			return
		}

		sliceVal := []string{}
		fieldValue = reflect.Indirect(fieldValue)
		for idx := 0; idx < fieldValue.Len(); idx++ {
			itemVal := fieldValue.Index(idx)
			if util.IsNil(itemVal) {
				continue
			}
			dtVal, dtErr := encodeDateTime(itemVal)
			if dtErr != nil {
				err = dtErr
				log.Errorf("encodeDateTime failed, err:%s", err.Error())
				return
			}
			sliceVal = append(sliceVal, dtVal)
		}

		ret.Value = sliceVal
		return
	}

	sliceObjectVal := []*ObjectValue{}
	fieldValue = reflect.Indirect(fieldValue)
	for idx := 0; idx < fieldValue.Len(); idx++ {
		itemVal := fieldValue.Index(idx)
		if util.IsNil(itemVal) {
			continue
		}

		itemVal = reflect.Indirect(itemVal)
		objVal, objErr := GetObjectValue(itemVal.Interface())
		if objErr != nil {
			err = objErr
			log.Errorf("encodeDateTimeValue failed, err:%s", err.Error())
			return
		}

		sliceObjectVal = append(sliceObjectVal, objVal)
	}
	ret.Value = &SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath(), Values: sliceObjectVal}
	return
}

func getObjectValue(entityVal reflect.Value) (ret *ObjectValue, err error) {
	entityVal = reflect.Indirect(entityVal)
	entityType := entityVal.Type()
	objType, objErr := newType(entityType)
	if objErr != nil {
		err = objErr
		return
	}
	if !util.IsStructType(objType.GetValue()) {
		err = fmt.Errorf("illegal entity, entity type:%s", entityType.String())
		return
	}

	//!! must be String, not Name
	ret = &ObjectValue{Name: objType.GetName(), PkgPath: objType.GetPkgPath(), IsPtr: objType.IsPtrType(), Items: []*ItemValue{}}
	fieldNum := entityVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldValue := reflect.Indirect(entityVal.Field(idx))
		fieldType := entityType.Field(idx)

		itemType, itemErr := newType(fieldType.Type)
		if itemErr != nil {
			err = itemErr
			log.Errorf("get entity field type failed, type%s, err:%s", fieldType.Type.String(), err.Error())
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

// GetObjectValue get object value
func GetObjectValue(entity interface{}) (ret *ObjectValue, err error) {
	entityVal := reflect.ValueOf(entity)
	ret, err = getObjectValue(entityVal)
	return
}

// GetSliceObjectValue get slice object value
func GetSliceObjectValue(sliceEntity interface{}) (ret *SliceObjectValue, err error) {
	sliceValue := reflect.ValueOf(sliceEntity)
	sliceType, sliceErr := newType(sliceValue.Type())
	if sliceErr != nil {
		err = fmt.Errorf("get slice object type failed, err:%s", err.Error())
		log.Errorf("GetSliceObjectValue failed, slice type name:%s", sliceType.GetName())
		return
	}

	if !util.IsSliceType(sliceType.GetValue()) {
		err = fmt.Errorf("illegal slice object value")
		log.Errorf("illegal slice type, slice type name:%s", sliceType.GetName())
		return
	}

	elemType := sliceType.Elem()
	if !util.IsStructType(elemType.GetValue()) {
		err = fmt.Errorf("illegal slice item type")
		log.Errorf("illegal slice elem type, type%s", elemType.GetName())
		return
	}

	ret = &SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath(), IsPtr: sliceType.IsPtrType(), IsElemPtr: elemType.IsPtrType(), Values: []*ObjectValue{}}
	sliceValue = reflect.Indirect(sliceValue)
	for idx := 0; idx < sliceValue.Len(); idx++ {
		val := sliceValue.Index(idx)
		objVal, objErr := getObjectValue(val)
		if objErr != nil {
			err = objErr
			log.Errorf("GetObjectValue failed, type%s, err:%s", val.Type().String(), err.Error())
			return
		}

		ret.Values = append(ret.Values, objVal)
	}

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
func decodeObjectValueFromMap(mapVal map[string]interface{}) (ret *ObjectValue, err error) {
	nameVal, nameOK := mapVal["name"]
	pkgPathVal, pkgPathOK := mapVal["pkgPath"]
	isPtrVal, isPtrOK := mapVal["isPtr"]
	itemsVal, itemsOK := mapVal["items"]
	if !nameOK || !pkgPathOK || !itemsOK || !isPtrOK {
		err = fmt.Errorf("illegal ObjectValue")
		return
	}

	objVal := &ObjectValue{Name: nameVal.(string), PkgPath: pkgPathVal.(string), IsPtr: isPtrVal.(bool), Items: []*ItemValue{}}
	for _, val := range itemsVal.([]interface{}) {
		item, itemOK := val.(map[string]interface{})
		if !itemOK {
			err = fmt.Errorf("illegal object field item value")
			return
		}

		itemVal, itemErr := decodeItemValue(item)
		if itemErr != nil {
			err = itemErr
			return
		}

		objVal.Items = append(objVal.Items, itemVal)
	}

	ret = objVal
	return
}

// decodeSliceObjectValueFromMap decode slice object value from map
func decodeSliceObjectValueFromMap(mapVal map[string]interface{}) (ret *SliceObjectValue, err error) {
	nameVal, nameOK := mapVal["name"]
	pkgPathVal, pkgPathOK := mapVal["pkgPath"]
	isPtrVal, isPtrOK := mapVal["isPtr"]
	isElemPtrVal, isElemPtrOK := mapVal["isElemPtr"]
	valuesVal, valuesOK := mapVal["values"]
	if !nameOK || !pkgPathOK || !valuesOK || !isPtrOK || !isElemPtrOK {
		err = fmt.Errorf("illegal SliceObjectValue")
		return
	}

	objVal := &SliceObjectValue{Name: nameVal.(string), PkgPath: pkgPathVal.(string), IsPtr: isPtrVal.(bool), IsElemPtr: isElemPtrVal.(bool), Values: []*ObjectValue{}}
	if valuesVal != nil {
		for _, val := range valuesVal.([]interface{}) {
			item, itemOK := val.(map[string]interface{})
			if !itemOK {
				err = fmt.Errorf("illegal slice object field item value")
				return
			}

			itemVal, itemErr := decodeObjectValueFromMap(item)
			if itemErr != nil {
				err = itemErr
				return
			}

			objVal.Values = append(objVal.Values, itemVal)
		}
	}

	ret = objVal
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
	// for struct or slice struct
	if objOK {
		_, itemsOK := objVal["items"]
		if itemsOK {
			ret = &ItemValue{Name: val.Name}

			oVal, oErr := decodeObjectValueFromMap(objVal)
			if oErr != nil {
				err = oErr
				return
			}

			ret.Value = oVal
			return
		}

		_, valuesOK := objVal["values"]
		if valuesOK {
			ret = &ItemValue{Name: val.Name}

			oVal, oErr := decodeSliceObjectValueFromMap(objVal)
			if oErr != nil {
				err = oErr
				return
			}

			ret.Value = oVal
			return
		}

		err = fmt.Errorf("illegal itemValue")
		return
	}

	// for basic slice
	sliceVal, sliceOK := val.Value.([]interface{})
	if sliceOK {
		ret = &ItemValue{Name: val.Name, Value: sliceVal}
		return
	}

	// for basic
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
	if l.IsElemPtr != r.IsElemPtr {
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
