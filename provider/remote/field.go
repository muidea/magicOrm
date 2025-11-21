package remote

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/utils"
)

type Field struct {
	Name        string    `json:"name"`
	ShowName    string    `json:"showName"`
	Description string    `json:"description"`
	Type        *TypeImpl `json:"type"`
	Spec        *SpecImpl `json:"spec"`
	value       *ValueImpl
}

type FieldValue struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

func (s *FieldValue) String() string {
	return fmt.Sprintf("name:%s,value:%+v", s.Name, s.Value)
}

func (s *Field) GetName() string {
	return s.Name
}

func (s *Field) GetShowName() string {
	return s.ShowName
}

func (s *Field) GetDescription() string {
	return s.Description
}

func (s *Field) GetType() (ret models.Type) {
	ret = s.Type
	return
}

func (s *Field) GetSpec() (ret models.Spec) {
	if s.Spec != nil {
		ret = s.Spec
		return
	}

	ret = &emptySpec
	return
}

func (s *Field) GetValue() (ret models.Value) {
	if s.value != nil {
		ret = s.value
		return
	}

	ret = &ValueImpl{}
	return
}

func (s *Field) SetValue(val any) *cd.Error {
	if s.value == nil {
		s.value = &ValueImpl{}
	}

	return s.value.Set(val)
}

func (s *Field) GetSliceValue() (ret []models.Value) {
	if !models.IsSlice(s.Type) || !s.value.IsValid() {
		return
	}

	ret = s.value.UnpackValue()
	return
}

func (s *Field) AppendSliceValue(val any) (err *cd.Error) {
	if val == nil {
		err = cd.NewError(cd.Unexpected, "field append slice value is nil")
		return
	}
	if !models.IsSlice(s.Type) {
		err = cd.NewError(cd.Unexpected, "field is not slice")
		return
	}

	err = s.value.Append(val)
	return
}

func (s *Field) Reset() {
	if s.value != nil && s.value.IsValid() {
		s.value = NewValue(getInitializeValue(s.Type))
		return
	}

	s.value = &ValueImpl{}
}

func (s *Field) copy(viewSpec models.ViewDeclare) (ret *Field, err error) {
	ret = &Field{
		Name:        s.Name,
		ShowName:    s.ShowName,
		Description: s.Description,
		Type:        s.Type.Copy(),
		//Spec:        s.Spec.Copy(),
	}
	if s.Spec == nil {
		ret.Spec = &emptySpec
	} else {
		ret.Spec = s.Spec.Copy()
	}

	switch viewSpec {
	case models.MetaView:
		if !s.Type.IsPtrType() {
			ret.value = NewValue(getInitializeValue(s.Type))
		} else {
			ret.value = &ValueImpl{}
		}
	case models.DetailView, models.LiteView:
		if !ret.Spec.EnableView(viewSpec) {
			ret.value = &ValueImpl{}
		} else {
			ret.value = NewValue(getInitializeValue(s.Type))
		}
	case models.OriginView:
		if s.value != nil {
			ret.value, _ = s.value.copy()
		} else {
			if !s.Type.IsPtrType() {
				ret.value = NewValue(getInitializeValue(s.Type))
			}
		}
	default:
		log.Warnf("fieldName:%s,unknown view spec:%v", s.Name, viewSpec)
	}

	return
}

func (s *Field) verifyAutoIncrement(typeVal models.TypeDeclare) *cd.Error {
	switch typeVal {
	case models.TypeBooleanValue,
		models.TypeStringValue,
		models.TypeDateTimeValue,
		models.TypeFloatValue,
		models.TypeDoubleValue,
		models.TypeStructValue,
		models.TypeSliceValue:
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal auto_increment field type, type:%v", typeVal))
	default:
	}

	return nil
}

func (s *Field) verifyUUID(typeVal models.TypeDeclare) *cd.Error {
	if typeVal != models.TypeStringValue {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal uuid field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifySnowflake(typeVal models.TypeDeclare) *cd.Error {
	if typeVal != models.TypeBigIntegerValue {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal snowflake field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifyDateTime(typeVal models.TypeDeclare) *cd.Error {
	if typeVal != models.TypeDateTimeValue {
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal datetime field type, type:%v", typeVal))
	}

	return nil
}

func (s *Field) verifyPK(typeVal models.TypeDeclare) *cd.Error {
	switch typeVal {
	case models.TypeStructValue, models.TypeSliceValue:
		return cd.NewError(cd.Unexpected, fmt.Sprintf("illegal primary key field type, type:%v", typeVal))
	default:
	}

	return nil
}

func (s *Field) verify() (err *cd.Error) {
	if s.Type == nil {
		err = cd.NewError(cd.Unexpected, fmt.Sprintf("illegal filed, field type is null, name:%v", s.Name))
		return
	}

	if s.Spec == nil {
		return
	}

	val := s.Type.GetValue()
	if s.Spec.GetValueDeclare() == models.AutoIncrement {
		err = s.verifyAutoIncrement(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == models.UUID {
		err = s.verifyUUID(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == models.Snowflake {
		err = s.verifySnowflake(val)
		if err != nil {
			return
		}
	}

	if s.Spec.GetValueDeclare() == models.DateTime {
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

func compareItem(l, r *Field) bool {
	if l.Name != r.Name {
		return false
	}
	if l.ShowName != r.ShowName {
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

func (s *FieldValue) IsValid() bool {
	if s.Value == nil {
		return false
	}

	switch v := s.Value.(type) {
	case *ObjectValue:
		return v != nil
	case *SliceObjectValue:
		return v != nil
	default:
		return utils.IsReallyValidValue(s.Value)
	}
}

func (s *FieldValue) IsZero() bool {
	if s.Value == nil {
		return true
	}

	switch v := s.Value.(type) {
	case *ObjectValue:
		return v == nil || len(v.Fields) == 0
	case *SliceObjectValue:
		return v == nil || len(v.Values) == 0
	default:
		return utils.IsReallyZeroValue(s.Value)
	}
}

func (s *FieldValue) Set(val any) {
	if val == nil {
		s.Value = nil
		return
	}

	switch val.(type) {
	case *ObjectValue, *SliceObjectValue:
		s.Value = val
	case ObjectValue, SliceObjectValue:
		s.Value = &val
	default:
		if !utils.IsReallyValidValue(val) {
			panic(fmt.Sprintf("illegal value:%+v", val))
		}

		s.Value = val
	}
}

func (s *FieldValue) Get() any {
	return s.Value
}

func (s *FieldValue) GetName() string {
	return s.Name
}

func (s *FieldValue) GetValue() models.Value {
	return &ValueImpl{value: s.Value}
}

func (s *FieldValue) copy() (ret *FieldValue) {
	if s.Value == nil {
		ret = &FieldValue{
			Name:  s.Name,
			Value: nil,
		}
		return
	}

	switch v := s.Value.(type) {
	case *ObjectValue:
		ret = &FieldValue{
			Name:  s.Name,
			Value: v.Copy(),
		}
		return
	case *SliceObjectValue:
		ret = &FieldValue{
			Name:  s.Name,
			Value: v.Copy(),
		}
		return ret
	default:
		copiedVal, copiedErr := utils.DeepCopy(s.Value)
		if copiedErr != nil {
			panic(copiedErr)
		}
		ret = &FieldValue{
			Name:  s.Name,
			Value: copiedVal,
		}
		return
	}
}
