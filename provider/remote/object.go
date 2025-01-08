package remote

import (
	"encoding/json"
	"fmt"
	"path"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
)

const (
	NameTag    = "name"
	PkgPathTag = "pkgPath"
	FieldsTag  = "fields"
	ValuesTag  = "values"
	ValueTag   = "value"
)

type Object struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	NickName    string   `json:"nickName"`
	Icon        string   `json:"icon"`
	PkgPath     string   `json:"pkgPath"`
	Description string   `json:"description"`
	Fields      []*Field `json:"fields"`
}

// ObjectValue Object value
type ObjectValue struct {
	ID      string        `json:"id"`
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

func (s *Object) SetFieldValue(name string, val model.Value) {
	for _, item := range s.Fields {
		if item.Name == name {
			item.SetValue(val)
			return
		}
	}
}

func (s *Object) SetPrimaryFieldValue(val model.Value) {
	for _, sf := range s.Fields {
		if sf.IsPrimaryKey() {
			sf.SetValue(val)
			return
		}
	}
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
func (s *Object) Interface(_ bool, viewSpec model.ViewDeclare) (ret any) {
	objVal := &ObjectValue{Name: s.Name, PkgPath: s.PkgPath, Fields: []*FieldValue{}}

	for _, sf := range s.Fields {
		var initVal any
		if sf.Spec != nil && sf.Spec.DefaultValue != nil {
			initVal = sf.Spec.DefaultValue
		}

		if viewSpec > 0 {
			if sf.Spec != nil && sf.Spec.EnableView(viewSpec) {
				if sf.value == nil || !sf.value.IsValid() {
					vVal, vErr := sf.Type.Interface(initVal)
					if vErr == nil {
						objVal.Fields = append(objVal.Fields, &FieldValue{Name: sf.Name, Value: vVal.Get()})
					} else {
						log.Errorf("Interface failed, pkgKey:%s, fieldName:%s, error:%s", s.GetPkgKey(), sf.GetName(), vErr.Error())
					}
					continue
				}

				objVal.Fields = append(objVal.Fields, &FieldValue{Name: sf.Name, Value: sf.value.Get()})
			}

			continue
		}

		if sf.value == nil || !sf.value.IsValid() {
			objVal.Fields = append(objVal.Fields, &FieldValue{Name: sf.Name, Value: initVal})
			continue
		}

		objVal.Fields = append(objVal.Fields, &FieldValue{Name: sf.Name, Value: sf.value.Get()})
	}

	pkValue := s.GetPrimaryField().GetValue()
	if pkValue.IsValid() {
		objVal.ID = fmt.Sprintf("%v", pkValue.Interface().Value())
	}

	ret = objVal
	return
}

func (s *Object) Copy(reset bool) (ret model.Model) {
	obj := &Object{
		ID:          s.ID,
		Name:        s.Name,
		NickName:    s.NickName,
		Icon:        s.Icon,
		PkgPath:     s.PkgPath,
		Description: s.Description,
		Fields:      []*Field{},
	}
	for _, val := range s.Fields {
		valPtr, valErr := val.copy(reset)
		if valErr != nil {
			log.Errorf("Copy field failed, name:%s, err:%s", val.GetName(), valErr.Error())
			panic(valErr)
		}

		obj.Fields = append(obj.Fields, valPtr)
	}

	ret = obj
	return
}

func (s *Object) Dump() (ret string) {
	ret = "\nmodelImpl:\n"
	ret = fmt.Sprintf("%s\tname:%s, pkgPath:%s\n", ret, s.GetName(), s.GetPkgPath())

	ret = fmt.Sprintf("%sfields:\n", ret)
	for _, field := range s.Fields {
		ret = fmt.Sprintf("%s\t%s\n", ret, field.dump())
	}

	return
}

func (s *Object) Verify() (err *cd.Result) {
	if s.Name == "" {
		err = cd.NewError(cd.UnExpected, "illegal object declare informain")
		return
	}

	for _, val := range s.Fields {
		err = val.verify()
		if err != nil {
			log.Errorf("Verify field failed, name:%s, err:%s", val.Name, err.Error())
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

func (s *ObjectValue) GetFieldValue(name string) any {
	for _, val := range s.Fields {
		if val.GetName() != name {
			continue
		}

		return val.Get()
	}

	return nil
}

func (s *ObjectValue) SetFieldValue(name string, value any) {
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

func encodeValue[T any](valPtr *T) (ret []byte, err *cd.Result) {
	byteVal, byteErr := json.Marshal(valPtr)
	if byteErr != nil {
		err = cd.NewError(cd.UnExpected, byteErr.Error())
		return
	}

	ret = byteVal
	return
}

// EncodeObjectValue encode objectValue to []byte
func EncodeObjectValue(objVal *ObjectValue) (ret []byte, err *cd.Result) {
	ret, err = encodeValue(objVal)
	return
}

// EncodeSliceObjectValue encode slice objectValue to []byte
func EncodeSliceObjectValue(objVal *SliceObjectValue) (ret []byte, err *cd.Result) {
	ret, err = encodeValue(objVal)
	return
}
func decodeObjectValueFromMap(mapVal map[string]any) (ret *ObjectValue, err *cd.Result) {
	nameVal, nameOK := mapVal[NameTag]
	pkgPathVal, pkgPathOK := mapVal[PkgPathTag]
	itemsVal, itemsOK := mapVal[FieldsTag]
	if !nameOK || !pkgPathOK || !itemsOK {
		err = cd.NewError(cd.UnExpected, "illegal ObjectValue")
		return
	}

	if itemsVal == nil {
		return
	}

	items := itemsVal.([]any)
	objVal := &ObjectValue{
		Name:    nameVal.(string),
		PkgPath: pkgPathVal.(string),
		Fields:  make([]*FieldValue, 0, len(items)),
	}

	for _, val := range items {
		item, itemOK := val.(map[string]any)
		if !itemOK {
			err = cd.NewError(cd.UnExpected, "illegal object field item value")
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

func decodeSliceObjectValueFromMap(mapVal map[string]any) (ret *SliceObjectValue, err *cd.Result) {
	nameVal, nameOK := mapVal[NameTag]
	pkgPathVal, pkgPathOK := mapVal[PkgPathTag]
	valuesVal, valuesOK := mapVal[ValuesTag]
	if !nameOK || !pkgPathOK || !valuesOK {
		err = cd.NewError(cd.UnExpected, "illegal SliceObjectValue")
		return
	}

	if valuesVal == nil {
		return
	}

	values := valuesVal.([]any)
	objVal := &SliceObjectValue{
		Name:    nameVal.(string),
		PkgPath: pkgPathVal.(string),
		Values:  make([]*ObjectValue, 0, len(values)),
	}

	for _, val := range values {
		item, itemOK := val.(map[string]any)
		if !itemOK {
			err = cd.NewError(cd.UnExpected, "illegal slice object field item value")
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

func decodeItemValue(itemVal map[string]any) (ret *FieldValue, err *cd.Result) {
	if itemVal == nil {
		err = cd.NewError(cd.UnExpected, "itemVal is nil")
		return
	}

	nameVal, nameOK := itemVal[NameTag]
	if !nameOK {
		err = cd.NewError(cd.UnExpected, "illegal item value")
		return
	}

	nameStr, ok := nameVal.(string)
	if !ok {
		err = cd.NewError(cd.UnExpected, "nameVal is not a string")
		return
	}

	valVal := itemVal[ValueTag]
	ret = &FieldValue{Name: nameStr, Value: valVal}
	ret, err = ConvertItem(ret)
	return
}

func ConvertItem(val *FieldValue) (ret *FieldValue, err *cd.Result) {
	objVal, objOK := val.Value.(map[string]any)
	// for struct or slice struct
	if objOK {
		_, itemsOK := objVal[FieldsTag]
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

		_, valuesOK := objVal[ValuesTag]
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

		err = cd.NewError(cd.UnExpected, "illegal itemValue")
		return
	}

	// for basic slice
	sliceVal, sliceOK := val.Value.([]any)
	if sliceOK {
		ret = &FieldValue{Name: val.Name, Value: sliceVal}
		return
	}

	// for basic
	ret = val
	return
}
func decodeValue[T any](data []byte, val *T, decodeFunc func(*T) (*T, *cd.Result)) (*T, *cd.Result) {
	byteErr := json.Unmarshal(data, val)
	if byteErr != nil {
		return nil, cd.NewError(cd.UnExpected, byteErr.Error())
	}

	ret, err := decodeFunc(val)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func DecodeObjectValue(data []byte) (ret *ObjectValue, err *cd.Result) {
	return decodeValue(data, &ObjectValue{}, ConvertObjectValue)
}

func DecodeSliceObjectValue(data []byte) (ret *SliceObjectValue, err *cd.Result) {
	return decodeValue(data, &SliceObjectValue{}, ConvertSliceObjectValue)
}

// ConvertObjectValue convert object value
func ConvertObjectValue(objVal *ObjectValue) (ret *ObjectValue, err *cd.Result) {
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

func ConvertSliceObjectValue(objVal *SliceObjectValue) (ret *SliceObjectValue, err *cd.Result) {
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
	return l.Name == r.Name && l.IsNil() == r.IsNil()
}

func CompareObjectValue(l, r *ObjectValue) bool {
	if l.Name != r.Name || l.PkgPath != r.PkgPath || len(l.Fields) != len(r.Fields) {
		return false
	}

	for idx := 0; idx < len(l.Fields); idx++ {
		if !compareItemValue(l.Fields[idx], r.Fields[idx]) {
			return false
		}
	}

	return true
}

func CompareSliceObjectValue(l, r *SliceObjectValue) bool {
	if l.Name != r.Name || l.PkgPath != r.PkgPath || len(l.Values) != len(r.Values) {
		return false
	}

	for idx := 0; idx < len(l.Values); idx++ {
		if !CompareObjectValue(l.Values[idx], r.Values[idx]) {
			return false
		}
	}

	return true
}
