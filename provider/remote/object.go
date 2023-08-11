package remote

import (
	"encoding/json"
	"fmt"
	"path"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	pu "github.com/muidea/magicOrm/provider/util"
)

type Object struct {
	Name    string   `json:"name"`
	PkgPath string   `json:"pkgPath"`
	IsPtr   bool     `json:"isPtr"`
	Fields  []*Field `json:"fields"`
}

// ObjectValue Object value
type ObjectValue struct {
	Name    string        `json:"name"`
	PkgPath string        `json:"pkgPath"`
	Fields  []*FieldValue `json:"fields"`
}

// SliceObjectValue slice object value
type SliceObjectValue struct {
	Name      string         `json:"name"`
	PkgPath   string         `json:"pkgPath"`
	IsElemPtr bool           `json:"isElemPtr"`
	Values    []*ObjectValue `json:"values"`
}

func (s *Object) GetName() (ret string) {
	ret = s.Name
	return
}

func (s *Object) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

func (s *Object) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *Object) IsPtrValue() bool {
	return s.IsPtr
}

func (s *Object) GetFields() (ret model.Fields) {
	for _, val := range s.Fields {
		ret = append(ret, val)
	}

	return
}

func (s *Object) SetFieldValue(name string, val model.Value) (err error) {
	for _, item := range s.Fields {
		if item.Name == name {
			err = item.SetValue(val)
			if err != nil {
				log.Errorf("set field value failed, object name:%s, err:%s", s.Name, err.Error())
			}

			return
		}
	}

	err = fmt.Errorf("invalid field name:%s", name)
	return
}

func (s *Object) GetPrimaryField() (ret model.Field) {
	for _, v := range s.Fields {
		if v.IsPrimary() {
			ret = v
			return
		}
	}

	return
}

func (s *Object) GetField(name string) (ret model.Field) {
	for _, v := range s.Fields {
		if v.GetName() == name {
			ret = v
			return
		}
	}

	return
}

func (s *Object) itemInterface(valPtr *Field) (ret interface{}) {
	rVal := valPtr.value.Get()
	if !valPtr.Type.IsBasic() {
		rVal = rVal.Addr()
		if model.IsStructType(valPtr.Type.GetValue()) {
			objectVal := rVal.Interface().(*ObjectValue)
			if len(objectVal.Fields) > 0 {
				ret = objectVal
			}
		}
		if model.IsSliceType(valPtr.Type.GetValue()) {
			sliceObjectVal := rVal.Interface().(*SliceObjectValue)
			if len(sliceObjectVal.Values) > 0 {
				ret = sliceObjectVal
			}
		}

		return
	}

	ret = rVal.Interface()
	return
}

// Interface object value
func (s *Object) Interface(ptrValue bool) (ret interface{}) {
	objVal := &ObjectValue{Name: s.Name, PkgPath: s.PkgPath, Fields: []*FieldValue{}}

	for _, v := range s.Fields {
		if v.value.IsNil() {
			objVal.Fields = append(objVal.Fields, &FieldValue{Name: v.Name})
			continue
		}

		interfaceVal := s.itemInterface(v)
		objVal.Fields = append(objVal.Fields, &FieldValue{Name: v.Name, Value: interfaceVal})
	}

	if ptrValue {
		ret = objVal
		return
	}

	ret = *objVal
	return
}

func (s *Object) Copy() (ret model.Model) {
	obj := &Object{Name: s.Name, PkgPath: s.PkgPath, Fields: []*Field{}}
	for _, val := range s.Fields {
		item := &Field{Index: val.Index, Name: val.Name, Type: val.Type.copy()}
		if val.Spec != nil {
			item.Spec = val.Spec.copy()
		}
		if val.value != nil {
			item.value = val.value.Copy()
		} else {
			initVal := val.Type.Interface()
			item.value = pu.NewValue(initVal.Get())
		}

		obj.Fields = append(obj.Fields, item)
	}

	ret = obj
	return
}

func (s *Object) Dump() (ret string) {
	ret = fmt.Sprintf("\nmodelImpl:\n")
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.Fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.dump())
	}

	return
}

func (s *Object) Verify() (err error) {
	if s.Name == "" || s.PkgPath == "" {
		err = fmt.Errorf("illegal object declare informain")
		return
	}

	for _, val := range s.Fields {
		err = val.verify()
		if err != nil {
			log.Errorf("Verify field failed, idx:%d, name:%s, err:%s", val.Index, val.Name, err.Error())
			return
		}
	}
	return
}

// GetObject GetObject
func GetObject(entity interface{}) (ret *Object, err error) {
	entityType := reflect.ValueOf(entity).Type()
	ret, err = type2Object(entityType)
	if err != nil {
		log.Errorf("type2Object failed, raw type:%s, err:%s", entityType.String(), err.Error())
	}

	return
}

// type2Object type2Object
func type2Object(entityType reflect.Type) (ret *Object, err error) {
	isPtr := false
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
		isPtr = true
	}
	if entityType.Kind() == reflect.Interface {
		entityType = entityType.Elem()
	}
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
		isPtr = true
	}

	typeImpl, typeErr := newType(entityType)
	if typeErr != nil {
		err = fmt.Errorf("illegal entity type, must be a struct obj, type:%s", entityType.String())
		return
	}
	if !model.IsStructType(typeImpl.GetValue()) {
		err = fmt.Errorf("illegal obj type, must be a struct obj, type:%s", entityType.String())
		return
	}

	impl := &Object{}
	impl.Name = entityType.Name()
	impl.PkgPath = entityType.PkgPath()
	impl.IsPtr = isPtr
	impl.Fields = []*Field{}

	hasPrimaryKey := false
	fieldNum := entityType.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		fItem, fErr := getItemInfo(idx, fieldType)
		if fErr != nil {
			err = fErr
			return
		}
		if fItem.IsPrimary() {
			if hasPrimaryKey {
				err = fmt.Errorf("duplicate primary key field, field idx:%d,field name:%s, struct name:%s", idx, fieldType.Name, impl.GetName())
				return
			}

			hasPrimaryKey = true
		}

		impl.Fields = append(impl.Fields, fItem)
	}

	if len(impl.Fields) == 0 {
		err = fmt.Errorf("no define orm field, struct name:%s", impl.GetName())
		return
	}

	if !hasPrimaryKey {
		err = fmt.Errorf("no define primary key field, struct name:%s", impl.GetName())
		return
	}

	ret = impl
	return
}

func EncodeObject(objPtr *Object) (ret []byte, err error) {
	ret, err = json.Marshal(objPtr)
	return
}

func DecodeObject(data []byte) (ret *Object, err error) {
	objPtr := &Object{}
	err = json.Unmarshal(data, objPtr)
	if err != nil {
		return
	}

	ret = objPtr
	return
}

func compareObject(l, r *Object) bool {
	if l.Name != r.Name {
		return false
	}

	if l.PkgPath != r.PkgPath {
		return false
	}

	if len(l.Fields) != len(r.Fields) {
		return false
	}

	for idx := 0; idx < len(l.Fields); idx++ {
		lVal := l.Fields[idx]
		rVal := r.Fields[idx]
		if !compareItem(lVal, rVal) {
			return false
		}
	}

	return true
}

// GetName get object name
func (s *ObjectValue) GetName() string {
	return s.Name
}

// GetPkgPath get pkg path
func (s *ObjectValue) GetPkgPath() string {
	return s.PkgPath
}

func (s *ObjectValue) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *ObjectValue) isFieldAssigned(val *FieldValue) (ret bool) {
	if val.Value == nil {
		return
	}

	bVal, bOK := val.Value.(bool)
	if bOK {
		ret = bVal
		return
	}

	strVal, strOK := val.Value.(string)
	if strOK {
		ret = strVal != ""
		return
	}

	i64Val, iOK := val.Value.(int64)
	if iOK {
		ret = i64Val != 0
		return
	}

	iVal, iOK := val.Value.(int)
	if iOK {
		ret = iVal != 0
		return
	}

	fltVal, fltOK := val.Value.(float64)
	if fltOK {
		ret = fltVal != 0
		return
	}

	sliceObjPtrVal, sliceObjPtrOK := val.Value.(*SliceObjectValue)
	if sliceObjPtrOK {
		ret = len(sliceObjPtrVal.Values) > 0
		return
	}

	ptrObjVal, ptrObjOK := val.Value.(*ObjectValue)
	if ptrObjOK {
		ret = ptrObjVal.IsAssigned()
	}
	return
}

// IsAssigned is assigned value
func (s *ObjectValue) IsAssigned() (ret bool) {
	ret = false
	for _, val := range s.Fields {
		ret = s.isFieldAssigned(val)
		if ret {
			return
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

func (s *SliceObjectValue) IsElemPtrValue() bool {
	return s.IsElemPtr
}

// IsAssigned is assigned value
func (s *SliceObjectValue) IsAssigned() (ret bool) {
	ret = len(s.Values) > 0
	return
}

func getFieldValue(fieldName string, itemType *TypeImpl, itemValue *pu.ValueImpl) (ret *FieldValue, err error) {
	if itemValue.IsNil() {
		ret = &FieldValue{Name: fieldName, Value: nil}
		return
	}

	if itemType.IsBasic() {
		encodeVal, encodeErr := _codec.Encode(itemValue, itemType)
		if encodeErr != nil {
			err = encodeErr
			return
		}
		ret = &FieldValue{Name: fieldName, Value: encodeVal}
		return
	}

	objVal, objErr := getObjectValue(itemValue.Get())
	if objErr != nil {
		err = objErr
		log.Errorf("GetObjectValue failed, raw type:%s, err:%s", itemType.GetName(), err.Error())
		return
	}

	ret = &FieldValue{Name: fieldName, Value: objVal}
	return
}

func getSliceFieldValue(fieldName string, itemType *TypeImpl, itemValue *pu.ValueImpl) (ret *FieldValue, err error) {
	ret = &FieldValue{Name: fieldName}
	if itemValue.IsNil() {
		ret = &FieldValue{Name: fieldName, Value: nil}
		return
	}

	elemType := itemType.Elem()
	if elemType.IsBasic() {
		encodeVal, encodeErr := _codec.Encode(itemValue, itemType)
		if encodeErr != nil {
			err = encodeErr
			return
		}
		ret = &FieldValue{Name: fieldName, Value: encodeVal}
		return
	}

	sliceObjectVal := []*ObjectValue{}
	rawVal := reflect.Indirect(itemValue.Get())
	for idx := 0; idx < rawVal.Len(); idx++ {
		itemVal := rawVal.Index(idx)
		objVal, objErr := getObjectValue(itemVal)
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
	if !model.IsStructType(objType.GetValue()) {
		err = fmt.Errorf("illegal entity, entity type:%s", entityType.String())
		return
	}

	//!! must be String, not Name
	ret = &ObjectValue{Name: objType.GetName(), PkgPath: objType.GetPkgPath(), Fields: []*FieldValue{}}
	fieldNum := entityVal.NumField()
	for idx := 0; idx < fieldNum; idx++ {
		fieldType := entityType.Field(idx)
		fieldName, fieldErr := getFieldName(fieldType)
		if fieldErr != nil {
			err = fieldErr
			log.Errorf("get entity failed, field name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		valuePtr := pu.NewValue(entityVal.Field(idx))
		typePtr, typeErr := newType(fieldType.Type)
		if typeErr != nil {
			err = typeErr
			log.Errorf("get entity type failed, field name:%s, err:%s", fieldType.Name, err.Error())
			return
		}

		if typePtr.GetValue() != model.TypeSliceValue {
			val, valErr := getFieldValue(fieldName, typePtr, valuePtr)
			if valErr != nil {
				err = valErr
				log.Errorf("getFieldValue failed, field name:%s, err:%s", fieldType.Name, err.Error())
				return
			}
			ret.Fields = append(ret.Fields, val)
		} else {
			val, valErr := getSliceFieldValue(fieldName, typePtr, valuePtr)
			if valErr != nil {
				err = valErr
				log.Errorf("getSliceFieldValue failed, field name:%s, err:%s", fieldType.Name, err.Error())
				return
			}
			ret.Fields = append(ret.Fields, val)
		}
	}

	return
}

func GetMapValue(entity interface{}) (ret interface{}, err error) {
	mVal, mOK := entity.(map[string]interface{})
	if !mOK {
		err = fmt.Errorf("illegal map value")
		return
	}
	ret, mOK = mVal["id"]
	if !mOK {
		err = fmt.Errorf("illegal map value, miss id")
		return
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

	if !model.IsSliceType(sliceType.GetValue()) {
		err = fmt.Errorf("illegal slice object value")
		log.Errorf("illegal slice type, slice type name:%s", sliceType.GetName())
		return
	}

	elemType := sliceType.Elem()
	if !model.IsStructType(elemType.GetValue()) {
		err = fmt.Errorf("illegal slice item type")
		log.Errorf("illegal slice elem type, type%s", elemType.GetName())
		return
	}

	ret = &SliceObjectValue{Name: elemType.GetName(), PkgPath: elemType.GetPkgPath(), IsElemPtr: elemType.IsPtrType(), Values: []*ObjectValue{}}
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

func TransferObjectValue(name, pkgPath string, vals []*ObjectValue) (ret *SliceObjectValue) {
	ret = &SliceObjectValue{
		Name:    name,
		PkgPath: pkgPath,
		Values:  vals,
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
	itemsVal, itemsOK := mapVal["fields"]
	if !nameOK || !pkgPathOK || !itemsOK {
		err = fmt.Errorf("illegal ObjectValue")
		return
	}

	if itemsVal == nil {
		return
	}

	objVal := &ObjectValue{Name: nameVal.(string), PkgPath: pkgPathVal.(string), Fields: []*FieldValue{}}
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

		objVal.Fields = append(objVal.Fields, itemVal)
	}

	ret = objVal
	return
}

// decodeSliceObjectValueFromMap decode slice object value from map
func decodeSliceObjectValueFromMap(mapVal map[string]interface{}) (ret *SliceObjectValue, err error) {
	nameVal, nameOK := mapVal["name"]
	pkgPathVal, pkgPathOK := mapVal["pkgPath"]
	isElemPtrVal, isElemPtrOK := mapVal["isElemPtr"]
	valuesVal, valuesOK := mapVal["values"]
	if !nameOK || !pkgPathOK || !valuesOK || !isElemPtrOK {
		err = fmt.Errorf("illegal SliceObjectValue")
		return
	}

	if valuesVal == nil {
		return
	}

	objVal := &SliceObjectValue{Name: nameVal.(string), PkgPath: pkgPathVal.(string), IsElemPtr: isElemPtrVal.(bool), Values: []*ObjectValue{}}
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

	ret = objVal
	return
}

func decodeItemValue(itemVal map[string]interface{}) (ret *FieldValue, err error) {
	nameVal, nameOK := itemVal["name"]
	valVal, valOK := itemVal["value"]
	if !nameOK || !valOK {
		err = fmt.Errorf("illegal item value")
	}

	ret = &FieldValue{Name: nameVal.(string), Value: valVal}
	ret, err = ConvertItem(ret)
	return
}

// ConvertItem convert FieldValue
func ConvertItem(val *FieldValue) (ret *FieldValue, err error) {
	objVal, objOK := val.Value.(map[string]interface{})
	// for struct or slice struct
	if objOK {
		_, itemsOK := objVal["fields"]
		if itemsOK {
			ret = &FieldValue{Name: val.Name}

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
			ret = &FieldValue{Name: val.Name}

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
		ret = &FieldValue{Name: val.Name, Value: sliceVal}
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

	for idx := range val.Fields {
		cur := val.Fields[idx]

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

// DecodeSliceObjectValue decode objectValue
func DecodeSliceObjectValue(data []byte) (ret *SliceObjectValue, err error) {
	sliceVal := &SliceObjectValue{}
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

// ConvertObjectValue convert object value
func ConvertObjectValue(objVal *ObjectValue) (ret *ObjectValue, err error) {
	for idx := range objVal.Fields {
		cur := objVal.Fields[idx]

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

func ConvertSliceObjectValue(objVal *SliceObjectValue) (ret *SliceObjectValue, err error) {
	for idx := range objVal.Values {
		cur := objVal.Values[idx]

		valPtr, valErr := ConvertObjectValue(cur)
		if valErr != nil {
			err = valErr
			return
		}

		objVal.Values[idx] = valPtr
	}

	ret = objVal
	return
}

func compareItemValue(l, r *FieldValue) bool {
	if l.Name != r.Name {
		return false
	}
	if l.IsNil() != r.IsNil() {
		return false
	}

	return true
}

func CompareObjectValue(l, r *ObjectValue) bool {
	if l.Name != r.Name {
		return false
	}

	if l.PkgPath != r.PkgPath {
		return false
	}

	if len(l.Fields) != len(r.Fields) {
		return false
	}

	for idx := 0; idx < len(l.Fields); idx++ {
		lVal := l.Fields[idx]
		rVal := r.Fields[idx]
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
	if l.IsElemPtr != r.IsElemPtr {
		return false
	}
	if len(l.Values) != len(r.Values) {
		return false
	}

	for idx := 0; idx < len(l.Values); idx++ {
		lVal := l.Values[idx]
		rVal := r.Values[idx]
		if !CompareObjectValue(lVal, rVal) {
			return false
		}
	}

	return true
}
