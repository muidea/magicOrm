package remote

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/util"
)

type Field struct {
	Index       int       `json:"index"`
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

func (s *Field) GetIndex() (ret int) {
	return s.Index
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

func (s *Field) SetValue(val model.Value) (err error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			log.Errorf("SetValue failed, unexpected field, name:%v, err:%v", s.Name, errInfo)
			err = fmt.Errorf("illegal value")
		}
	}()

	valPtr, valErr := s.Type.Interface(val.Interface())
	if valErr != nil {
		err = valErr
		return
	}

	s.value = valPtr.(*ValueImpl)
	return
}

func (s *Field) IsPrimaryKey() bool {
	if s.Spec == nil {
		return false
	}

	return s.Spec.IsPrimaryKey()
}

func (s *Field) copy() (ret model.Field) {
	return &Field{Index: s.Index, Name: s.Name, Spec: s.Spec, Type: s.Type, value: s.value}
}

func (s *Field) verifyAutoIncrement(typeVal model.TypeDeclare) error {
	switch typeVal {
	case model.TypeBooleanValue,
		model.TypeStringValue,
		model.TypeDateTimeValue,
		model.TypeFloatValue,
		model.TypeDoubleValue,
		model.TypeStructValue,
		model.TypeSliceValue:
		return fmt.Errorf("illegal auto_increment field type, type:%v", typeVal)
	default:
	}

	return nil
}

func (s *Field) verifyUUID(typeVal model.TypeDeclare) error {
	if typeVal != model.TypeStringValue {
		return fmt.Errorf("illegal uuid field type, type:%v", typeVal)
	}

	return nil
}

func (s *Field) verifySnowFlake(typeVal model.TypeDeclare) error {
	if typeVal != model.TypeBigIntegerValue {
		return fmt.Errorf("illegal snowflake field type, type:%v", typeVal)
	}

	return nil
}

func (s *Field) verifyDateTime(typeVal model.TypeDeclare) error {
	if typeVal != model.TypeDateTimeValue {
		return fmt.Errorf("illegal dateTime field type, type:%v", typeVal)
	}

	return nil
}

func (s *Field) verifyPK(typeVal model.TypeDeclare) error {
	switch typeVal {
	case model.TypeStructValue, model.TypeSliceValue:
		return fmt.Errorf("illegal primary key field type, type:%v", typeVal)
	default:
	}

	return nil
}

func (s *Field) verify() (err error) {
	if s.Type == nil {
		err = fmt.Errorf("illegal filed, field type is null, index:%d, name:%v", s.Index, s.Name)
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
	str := fmt.Sprintf("index:%d,name:%s,type:[%s]", s.Index, s.Name, s.Type.dump())
	if s.Spec != nil {
		str = fmt.Sprintf("%s,spec:[%s]", str, s.Spec.dump())
	}
	if s.value != nil {
		str = fmt.Sprintf("%s,value:%v", str, s.value.Interface())
	}

	return str
}

func compareItem(l, r *Field) bool {
	if l.Index != r.Index {
		return false
	}
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

func (s *FieldValue) Set(val any) error {
	s.Value = val
	return nil
}

func (s *FieldValue) Get() any {
	return s.Value
}

func (s *FieldValue) Addr() model.Value {
	impl := &ValueImpl{value: &s.Value}
	return impl
}

func (s *FieldValue) Interface() any {
	return s.Value
}

func (s *FieldValue) IsBasic() bool {
	if s.Value == nil {
		return false
	}

	rValue := reflect.ValueOf(s.Value)
	if rValue.Kind() == reflect.Interface {
		rValue = rValue.Elem()
	}
	rType := rValue.Type()
	if util.IsSlice(rType) {
		rType = rType.Elem()
	}

	return !util.IsStruct(rType)
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
