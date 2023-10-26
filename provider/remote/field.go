package remote

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

type Field struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        *TypeImpl `json:"type"`
	Spec        *SpecImpl `json:"spec"`
	value       *ValueImpl
}

type FieldValue struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

func (s *Field) GetName() string {
	return s.Name
}

func (s *Field) GetDescription() string {
	return s.Description
}

func (s *Field) GetType() (ret model.Type) {
	ret = s.Type
	return
}

func (s *Field) GetSpec() (ret model.Spec) {
	if s.Spec != nil {
		ret = s.Spec
		return
	}

	ret = &emptySpec
	return
}

func (s *Field) GetValue() (ret model.Value) {
	if s.value != nil {
		ret = s.value
		return
	}

	ret = &NilValue
	return
}

func (s *Field) SetValue(val model.Value) {
	s.value = val.(*ValueImpl)
	return
}

func (s *Field) IsPrimaryKey() bool {
	if s.Spec == nil {
		return false
	}

	return s.Spec.IsPrimaryKey()
}

func (s *Field) IsBasic() bool {
	return s.Type.IsBasic()
}

func (s *Field) IsStruct() bool {
	return s.Type.IsStruct()
}

func (s *Field) IsSlice() bool {
	return s.Type.IsSlice()
}

func (s *Field) IsPtrType() bool {
	return s.Type.IsPtrType()
}

func (s *Field) copy() (ret *Field) {
	val := &Field{
		Name:        s.Name,
		Description: s.Description,
	}

	if s.Spec != nil {
		val.Spec = s.Spec.copy()
	}
	if s.Type != nil {
		val.Type = s.Type.copy()
	}

	if s.value != nil {
		val.value = s.value.Copy()
	}

	ret = val
	return
}

func (s *Field) verifyAutoIncrement(typeVal model.TypeDeclare) *cd.Result {
	switch typeVal {
	case model.TypeBooleanValue,
		model.TypeStringValue,
		model.TypeDateTimeValue,
		model.TypeFloatValue,
		model.TypeDoubleValue,
		model.TypeStructValue,
		model.TypeSliceValue:
		return cd.NewError(cd.UnExpected, fmt.Sprintf("illegal auto_increment field type, type:%v", typeVal))
	default:
	}

	return nil
}

func (s *Field) verifyUUID(typeVal model.TypeDeclare) *cd.Result {
	if typeVal != model.TypeStringValue {
		return cd.NewError(cd.UnExpected, fmt.Sprintf("illegal uuid field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifySnowFlake(typeVal model.TypeDeclare) *cd.Result {
	if typeVal != model.TypeBigIntegerValue {
		return cd.NewError(cd.UnExpected, fmt.Sprintf("illegal snowflake field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifyDateTime(typeVal model.TypeDeclare) *cd.Result {
	if typeVal != model.TypeDateTimeValue {
		return cd.NewError(cd.UnExpected, fmt.Sprintf("illegal dateTime field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifyPK(typeVal model.TypeDeclare) *cd.Result {
	switch typeVal {
	case model.TypeStructValue, model.TypeSliceValue:
		return cd.NewError(cd.UnExpected, fmt.Sprintf("illegal primary key field type, type:%v", typeVal))
	default:
	}

	return nil
}

func (s *Field) verify() (err *cd.Result) {
	if s.Type == nil {
		err = cd.NewError(cd.UnExpected, fmt.Sprintf("illegal filed, field type is null, name:%v", s.Name))
		return
	}

	if s.Spec == nil {
		return
	}

	val := s.Type.GetValue()
	if s.Spec.GetValueDeclare() == model.AutoIncrement {
		err = s.verifyAutoIncrement(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == model.UUID {
		err = s.verifyUUID(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == model.SnowFlake {
		err = s.verifySnowFlake(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == model.DateTime {
		err = s.verifyDateTime(val)
		if err != nil {
			return
		}
	}

	if s.Spec.IsPrimaryKey() {
		err = s.verifyPK(val)
	}

	return
}

func (s *Field) dump() string {
	str := fmt.Sprintf("name:%s,type:[%s]", s.Name, s.Type.dump())
	if s.Spec != nil {
		str = fmt.Sprintf("%s,spec:[%s]", str, s.Spec.dump())
	}
	if s.value != nil {
		str = fmt.Sprintf("%s,value:%v", str, s.value.Interface())
	}

	return str
}

func compareItem(l, r *Field) bool {
	if l.Name != r.Name {
		return false
	}

	if !compareType(l.Type, r.Type) {
		return false
	}
	if !compareSpec(l.Spec, r.Spec) {
		return false
	}

	return true
}

func (s *FieldValue) IsNil() bool {
	return s.Value == nil
}

func (s *FieldValue) IsZero() bool {
	return s.Value == nil
}

func (s *FieldValue) Set(val any) {
	s.Value = val
}

func (s *FieldValue) Get() any {
	return s.Value
}

func (s *FieldValue) GetName() string {
	return s.Name
}

func (s *FieldValue) GetValue() model.Value {
	return &ValueImpl{value: s.Value}
}

func (s *FieldValue) copy() (ret *FieldValue) {
	if s.Value == nil {
		ret = &FieldValue{}
		return
	}

	ret = &FieldValue{Value: s.Value}
	return
}
