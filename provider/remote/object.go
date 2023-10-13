package remote

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

const (
	nameTag    = "name"
	pkgPathTag = "pkgPath"
	fieldsTag  = "fields"
	valuesTag  = "values"
	valueTag   = "value"
)

type Object struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	PkgPath     string   `json:"pkgPath"`
	Description string   `json:"description"`
	Fields      []*Field `json:"fields"`
}

// ObjectValue Object value
type ObjectValue struct {
	Name    string        `json:"name"`
	PkgPath string        `json:"pkgPath"`
	Fields  []*FieldValue `json:"fields"`
}

// SliceObjectValue slice object value
type SliceObjectValue struct {
	Name    string         `json:"name"`
	PkgPath string         `json:"pkgPath"`
	Values  []*ObjectValue `json:"values"`
}

func (s *Object) GetName() (ret string) {
	ret = s.Name
	return
}

func (s *Object) GetPkgPath() (ret string) {
	ret = s.PkgPath
	return
}

func (s *Object) GetDescription() (ret string) {
	ret = s.Description
	return
}

func (s *Object) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
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
			item.SetValue(val)
			return
		}
	}

	err = fmt.Errorf("invalid field name:%s", name)
	return
}

func (s *Object) GetPrimaryField() (ret model.Field) {
	for _, v := range s.Fields {
		if v.IsPrimaryKey() {
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

// Interface object value
func (s *Object) Interface(_ bool) (ret any) {
	objVal := &ObjectValue{Name: s.Name, PkgPath: s.PkgPath, Fields: []*FieldValue{}}

	for _, v := range s.Fields {
		if v.value == nil || v.value.IsNil() {
			objVal.Fields = append(objVal.Fields, &FieldValue{Name: v.Name})
			continue
		}

		objVal.Fields = append(objVal.Fields, &FieldValue{Name: v.Name, Value: v.value.Get()})
	}

	ret = objVal
	return
}

func (s *Object) Copy(reset bool) (ret model.Model) {
	obj := &Object{
		Name:        s.Name,
		PkgPath:     s.PkgPath,
		Description: s.Description,
		Fields:      []*Field{},
	}
	for _, val := range s.Fields {
		obj.Fields = append(obj.Fields, val.copy(reset))
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
	if s.Name == "" {
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

func CompareObject(l, r *Object) bool {
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

func (s *ObjectValue) GetName() string {
	return s.Name
}

func (s *ObjectValue) GetPkgPath() string {
	return s.PkgPath
}

func (s *ObjectValue) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *ObjectValue) GetValue() []*FieldValue {
	return s.Fields
}

func (s *ObjectValue) SetFieldValue(name string, value interface{}) {
	found := false
	for _, val := range s.Fields {
		if val.GetName() != name {
			continue
		}

		found = true
		val.Set(value)
	}

	if !found {
		s.Fields = append(s.Fields, &FieldValue{Name: name, Value: value})
	}
}

func (s *ObjectValue) isFieldAssigned(val *FieldValue) (ret bool) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			log.Errorf("check isFieldAssigned unexpected, name:%s, err:%v", val.GetName(), errInfo)
		}
	}()

	valPtr := NewValue(val.Value)
	ret = !valPtr.IsZero()
	return
}

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

func (s *ObjectValue) Copy() (ret *ObjectValue) {
	ptr := &ObjectValue{
		Name:    s.Name,
		PkgPath: s.PkgPath,
	}

	for idx := 0; idx < len(s.Fields); idx++ {
		ptr.Fields = append(ptr.Fields, s.Fields[idx].copy())
	}

	ret = ptr
	return
}

func (s *SliceObjectValue) GetName() string {
	return s.Name
}

func (s *SliceObjectValue) GetPkgPath() string {
	return s.PkgPath
}

func (s *SliceObjectValue) GetPkgKey() string {
	return path.Join(s.GetPkgPath(), s.GetName())
}

func (s *SliceObjectValue) GetValue() []*ObjectValue {
	return s.Values
}

func (s *SliceObjectValue) IsAssigned() (ret bool) {
	ret = len(s.Values) > 0
	return
}

func (s *SliceObjectValue) Copy() (ret *SliceObjectValue) {
	ptr := &SliceObjectValue{
		Name:    s.Name,
		PkgPath: s.PkgPath,
	}

	for idx := 0; idx < len(s.Values); idx++ {
		sv := s.Values[idx].Copy()
		sv.Name = s.Name
		sv.PkgPath = s.PkgPath

		ptr.Values = append(ptr.Values, sv)
	}

	ret = ptr
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
	nameVal, nameOK := mapVal[nameTag]
	pkgPathVal, pkgPathOK := mapVal[pkgPathTag]
	itemsVal, itemsOK := mapVal[fieldsTag]
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
	nameVal, nameOK := mapVal[nameTag]
	pkgPathVal, pkgPathOK := mapVal[pkgPathTag]
	valuesVal, valuesOK := mapVal[valuesTag]
	if !nameOK || !pkgPathOK || !valuesOK {
		err = fmt.Errorf("illegal SliceObjectValue")
		return
	}

	if valuesVal == nil {
		return
	}

	objVal := &SliceObjectValue{Name: nameVal.(string), PkgPath: pkgPathVal.(string), Values: []*ObjectValue{}}
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
	nameVal, nameOK := itemVal[nameTag]
	valVal, _ := itemVal[valueTag]
	if !nameOK {
		err = fmt.Errorf("illegal item value")
	}

	ret = &FieldValue{Name: nameVal.(string), Value: valVal}
	ret, err = ConvertItem(ret)
	return
}

func ConvertItem(val *FieldValue) (ret *FieldValue, err error) {
	objVal, objOK := val.Value.(map[string]interface{})
	// for struct or slice struct
	if objOK {
		_, itemsOK := objVal[fieldsTag]
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

		_, valuesOK := objVal[valuesTag]
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

func CompareSliceObjectValue(l, r *SliceObjectValue) bool {
	if l.Name != r.Name {
		return false
	}
	if l.PkgPath != r.PkgPath {
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
