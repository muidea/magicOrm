package remote

import (
	"encoding/json"
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/utils"
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
	ShowName    string   `json:"showName"`
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

func (s *Object) GetShowName() (ret string) {
	ret = s.ShowName
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

func (s *Object) GetFields() (ret model.Fields) {
	for _, val := range s.Fields {
		ret = append(ret, val)
	}

	return
}

func (s *Object) SetFieldValue(name string, val any) (err *cd.Result) {
	for _, item := range s.Fields {
		if item.Name == name {
			err = item.SetValue(val)
			return
		}
	}

	log.Warnf("SetFieldValue failed, field:%s not found", name)
	return
}

func (s *Object) SetPrimaryFieldValue(val any) (err *cd.Result) {
	for _, sf := range s.Fields {
		if model.IsPrimaryField(sf) {
			err = sf.SetValue(val)
			return
		}
	}

	return
}

func (s *Object) GetPrimaryField() (ret model.Field) {
	for _, sf := range s.Fields {
		if model.IsPrimaryField(sf) {
			ret = sf
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

	for _, sf := range s.Fields {
		if sf.value == nil || !sf.value.IsValid() {
			continue
		}

		objVal.Fields = append(objVal.Fields, &FieldValue{Name: sf.Name, Value: sf.value.Get()})
	}

	pkValue := s.GetPrimaryField().GetValue()
	if pkValue.IsValid() {
		objVal.ID = fmt.Sprintf("%v", pkValue.Get())
	}

	ret = objVal
	return
}

func (s *Object) Copy(viewSpec model.ViewDeclare) (ret model.Model) {
	obj := &Object{
		ID:          s.ID,
		Name:        s.Name,
		ShowName:    s.ShowName,
		Icon:        s.Icon,
		PkgPath:     s.PkgPath,
		Description: s.Description,
		Fields:      []*Field{},
	}
	for _, val := range s.Fields {
		valPtr, valErr := val.copy(viewSpec)
		if valErr != nil {
			log.Errorf("copy field failed, name:%s, err:%s", val.GetName(), valErr.Error())
			panic(valErr)
		}

		obj.Fields = append(obj.Fields, valPtr)
	}

	ret = obj
	return
}

func (s *Object) Reset() {
	for _, val := range s.Fields {
		val.Reset()
	}
}

func (s *Object) Verify() (err *cd.Result) {
	if s.Name == "" {
		err = cd.NewResult(cd.UnExpected, "illegal object declare informain")
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

	ret = !utils.IsReallyZeroValue(val.Value)
	//valPtr := NewValue(val.Value)
	//ret = !valPtr.IsZero()
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
		ID:      s.ID,
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

func marshalHelper[T any](valPtr *T) (ret []byte, err *cd.Result) {
	byteVal, byteErr := json.Marshal(valPtr)
	if byteErr != nil {
		err = cd.NewResult(cd.UnExpected, byteErr.Error())
		return
	}

	ret = byteVal
	return
}

// EncodeObjectValue encode objectValue to []byte
func EncodeObjectValue(objVal *ObjectValue) (ret []byte, err *cd.Result) {
	ret, err = marshalHelper(objVal)
	return
}

// EncodeSliceObjectValue encode slice objectValue to []byte
func EncodeSliceObjectValue(objVal *SliceObjectValue) (ret []byte, err *cd.Result) {
	ret, err = marshalHelper(objVal)
	return
}
func decodeObjectValueFromMap(mapVal map[string]any) (ret *ObjectValue, err *cd.Result) {
	nameVal, nameOK := mapVal[NameTag]
	pkgPathVal, pkgPathOK := mapVal[PkgPathTag]
	itemsVal, itemsOK := mapVal[FieldsTag]
	if !nameOK || !pkgPathOK || !itemsOK {
		err = cd.NewResult(cd.UnExpected, "illegal ObjectValue")
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
			err = cd.NewResult(cd.UnExpected, "illegal object field item value")
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
		err = cd.NewResult(cd.UnExpected, "illegal SliceObjectValue")
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
			err = cd.NewResult(cd.UnExpected, "illegal slice object field item value")
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
		err = cd.NewResult(cd.UnExpected, "itemVal is nil")
		return
	}

	nameVal, nameOK := itemVal[NameTag]
	if !nameOK {
		err = cd.NewResult(cd.UnExpected, "illegal item value")
		return
	}

	nameStr, ok := nameVal.(string)
	if !ok {
		err = cd.NewResult(cd.UnExpected, "nameVal is not a string")
		return
	}

	valVal := itemVal[ValueTag]
	ret = &FieldValue{Name: nameStr, Value: valVal}
	ret, err = ConvertItem(ret)
	return
}

func convertAnySlice(slieceVal []interface{}) interface{} {
	if len(slieceVal) == 0 {
		return slieceVal
	}

	switch tVal := slieceVal[0].(type) {
	case float64:
		result := make([]float64, len(slieceVal))
		for i, v := range slieceVal {
			val, ok := v.(float64)
			if !ok {
				return []float64{}
			}

			result[i] = val
		}
		return result
	case string:
		result := make([]string, len(slieceVal))
		for i, v := range slieceVal {
			val, ok := v.(string)
			if !ok {
				return []string{}
			}
			result[i] = val
		}
		return result
	case bool:
		result := make([]bool, len(slieceVal))
		for i, v := range slieceVal {
			val, ok := v.(bool)
			if !ok {
				return []bool{}
			}
			result[i] = val
		}
		return result
	default:
		log.Errorf("convertInterfaceArrayToSlice unexpected type:%v", tVal)
		return nil
	}
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

		err = cd.NewResult(cd.UnExpected, "illegal itemValue")
		return
	}

	// for basic slice
	sliceVal, sliceOK := val.Value.([]any)
	if sliceOK {
		ret = &FieldValue{Name: val.Name, Value: convertAnySlice(sliceVal)}
		return
	}

	// for basic
	ret = val
	return
}
func unmarshalHelper[T any](data []byte, val *T, decodeFunc func(*T) (*T, *cd.Result)) (*T, *cd.Result) {
	byteErr := json.Unmarshal(data, val)
	if byteErr != nil {
		return nil, cd.NewResult(cd.UnExpected, byteErr.Error())
	}

	ret, err := decodeFunc(val)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func DecodeObjectValue(data []byte) (ret *ObjectValue, err *cd.Result) {
	return unmarshalHelper(data, &ObjectValue{}, ConvertObjectValue)
}

func DecodeSliceObjectValue(data []byte) (ret *SliceObjectValue, err *cd.Result) {
	return unmarshalHelper(data, &SliceObjectValue{}, ConvertSliceObjectValue)
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
	if l.Name != r.Name {
		return false
	}

	return true
	/*
		equal, diff := utils.CompareWithNumericConversion(l.Value, r.Value)
		if !equal {
			log.Errorf("compareItemValue failed, l:%v, r:%v, diff:%s", l, r, diff)
		}

		return equal
	*/
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
