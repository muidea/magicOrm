package remote

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	pu "github.com/muidea/magicOrm/provider/util"
)

type Field struct {
	Index int    `json:"index"`
	Name  string `json:"name"`

	Type  *TypeImpl `json:"type"`
	Spec  *SpecImpl `json:"spec"`
	value *pu.ValueImpl
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

	ret = &pu.NilValue
	return
}

func (s *Field) IsPrimary() bool {
	if s.Spec == nil {
		return false
	}

	return s.Spec.IsPrimaryKey()
}

func (s *Field) SetValue(val model.Value) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("SetValue failed, unexpected field:%v, err:%v", s.Name, err)
		}
	}()

	if s.value != nil {
		err = s.value.Set(val.Get())
		if err != nil {
			log.Errorf("set field value failed, name:%s, err:%s", s.Name, err.Error())
		}
		return
	}

	initVal := s.Type.Interface()
	initVal.Set(val.Get())
	s.value = pu.NewValue(initVal.Get())
	return
}

func (s *Field) copy() (ret model.Field) {
	return &Field{Index: s.Index, Name: s.Name, Spec: s.Spec, Type: s.Type, value: s.value}
}

func (s *Field) verify() (err error) {
	if s.Type == nil {
		return fmt.Errorf("illegal filed, field type is null, index:%d, name:%v", s.Index, s.Name)
	}

	if s.Spec != nil {
		val := s.Type.GetValue()
		if s.Spec.GetValueDeclare() == model.AutoIncrement {
			switch val {
			case model.TypeBooleanValue,
				model.TypeStringValue,
				model.TypeDateTimeValue,
				model.TypeFloatValue,
				model.TypeDoubleValue,
				model.TypeStructValue,
				model.TypeSliceValue:
				return fmt.Errorf("illegal auto_increment field type, type:%s", s.Type.dump())
			default:
			}
		}
		if s.Spec.GetValueDeclare() == model.UUID && val != model.TypeStringValue {
			return fmt.Errorf("illegal uuid field type, type:%s", s.Type.dump())
		}

		if s.Spec.GetValueDeclare() == model.SnowFlake && val != model.TypeBigIntegerValue {
			return fmt.Errorf("illegal snowflake field type, type:%s", s.Type.dump())
		}

		if s.Spec.GetValueDeclare() == model.DateTime && val != model.TypeDateTimeValue {
			return fmt.Errorf("illegal dateTime field type, type:%s", s.Type.dump())
		}

		if s.Spec.IsPrimaryKey() {
			switch val {
			case model.TypeStructValue, model.TypeSliceValue:
				return fmt.Errorf("illegal primary key field type, type:%s", s.Type.dump())
			default:
			}
		}
	}

	if s.value == nil || s.value.IsNil() {
		return nil
	}

	return s.value.Verify()
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

func getFieldName(fieldType reflect.StructField) (ret string, err error) {
	specPtr, specErr := newSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	fieldName := fieldType.Name
	if specPtr.GetFieldName() != "" {
		fieldName = specPtr.GetFieldName()
	}

	ret = fieldName
	return
}

func getItemInfo(idx int, fieldType reflect.StructField) (ret *Field, err error) {
	typeImpl, typeErr := newType(fieldType.Type)
	if typeErr != nil {
		err = typeErr
		return
	}

	specImpl, specErr := newSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	initVal := typeImpl.Interface()

	item := &Field{}
	item.Index = idx
	item.Name = fieldType.Name
	if specImpl.GetFieldName() != "" {
		item.Name = specImpl.GetFieldName()
	}
	item.Type = typeImpl
	item.Spec = specImpl
	item.value = pu.NewValue(initVal.Get())

	ret = item
	return
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

func (s *FieldValue) Set(val reflect.Value) error {
	s.Value = val.Interface()
	return nil
}

func (s *FieldValue) Get() reflect.Value {
	return reflect.ValueOf(s.Value)
}

func (s *FieldValue) Addr() model.Value {
	impl := &FieldValue{Value: &s.Value}
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
	if pu.IsSlice(rType) {
		rType = rType.Elem()
	}

	return !pu.IsStruct(rType)
}

func (s *FieldValue) copy() (ret *FieldValue) {
	if s.Value == nil {
		ret = &FieldValue{}
		return
	}

	ret = &FieldValue{Value: s.Value}
	return
}
