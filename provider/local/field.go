package local

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// field single field impl
type field struct {
	index int
	name  string

	typePtr  *TypeImpl
	specPtr  *SpecImpl
	valuePtr *ValueImpl
}

func (s *field) GetIndex() int {
	return s.index
}

func (s *field) GetName() string {
	return s.name
}

func (s *field) GetDescription() string {
	return ""
}

func (s *field) GetType() (ret model.Type) {
	ret = s.typePtr
	return
}

func (s *field) GetSpec() (ret model.Spec) {
	if s.specPtr != nil {
		ret = s.specPtr
		return
	}

	ret = &emptySpec
	return
}

func (s *field) GetValue() (ret model.Value) {
	if s.valuePtr != nil {
		ret = s.valuePtr
		return
	}

	ret = &NilValue
	return
}

func (s *field) SetValue(val model.Value) (err error) {
	s.valuePtr = val.(*ValueImpl)
	return
}

func (s *field) IsPrimaryKey() bool {
	if s.specPtr == nil {
		return false
	}

	return s.specPtr.IsPrimaryKey()
}

func (s *field) copy() *field {
	val := &field{
		index: s.index,
		name:  s.name,
	}

	if s.typePtr != nil {
		val.typePtr = s.typePtr.copy()
	}
	if s.specPtr != nil {
		val.specPtr = s.specPtr.copy()
	}
	if s.valuePtr != nil {
		val.valuePtr = s.valuePtr.Copy()
	}
	return val
}

func (s *field) verifyAutoIncrement(typeVal model.TypeDeclare) error {
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

func (s *field) verifyUUID(typeVal model.TypeDeclare) error {
	if typeVal != model.TypeStringValue {
		return fmt.Errorf("illegal uuid field type, type:%v", typeVal)
	}

	return nil
}

func (s *field) verifySnowFlake(typeVal model.TypeDeclare) error {
	if typeVal != model.TypeBigIntegerValue {
		return fmt.Errorf("illegal snowflake field type, type:%v", typeVal)
	}

	return nil
}

func (s *field) verifyDateTime(typeVal model.TypeDeclare) error {
	if typeVal != model.TypeDateTimeValue {
		return fmt.Errorf("illegal dateTime field type, type:%v", typeVal)
	}

	return nil
}

func (s *field) verifyPK(typeVal model.TypeDeclare) error {
	switch typeVal {
	case model.TypeStructValue, model.TypeSliceValue:
		return fmt.Errorf("illegal primary key field type, type:%v", typeVal)
	default:
	}

	return nil
}

func (s *field) verify() (err error) {
	if s.typePtr == nil {
		err = fmt.Errorf("illegal filed, field type is null, index:%d, name:%v", s.index, s.name)
		return
	}

	if s.specPtr == nil {
		return
	}
	val := s.typePtr.GetValue()
	if s.specPtr.GetValueDeclare() == model.AutoIncrement {
		err = s.verifyAutoIncrement(val)
		if err != nil {
			return
		}
	}
	if s.specPtr.GetValueDeclare() == model.UUID {
		err = s.verifyUUID(val)
		if err != nil {
			return
		}
	}

	if s.specPtr.GetValueDeclare() == model.SnowFlake {
		err = s.verifySnowFlake(val)
		if err != nil {
			return
		}
	}

	if s.specPtr.GetValueDeclare() == model.DateTime {
		err = s.verifyDateTime(val)
		if err != nil {
			return
		}
	}

	if s.specPtr.IsPrimaryKey() {
		err = s.verifyPK(val)
	}

	return
}

func (s *field) dump() string {
	str := fmt.Sprintf("index:%d,name:%s,type:[%s],spec:[%s]", s.index, s.name, s.typePtr.dump(), s.specPtr.dump())
	return str
}

func getFieldName(fieldType reflect.StructField) (ret string, err error) {
	fieldName := fieldType.Name
	specPtr, specErr := NewSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	if specPtr.GetFieldName() != "" {
		fieldName = specPtr.GetFieldName()
	}

	ret = fieldName
	return
}

func getFieldInfo(idx int, fieldType reflect.StructField, fieldValue reflect.Value) (ret *field, err error) {
	typePtr, typeErr := NewType(fieldType.Type)
	if typeErr != nil {
		err = typeErr
		return
	}

	specPtr, specErr := NewSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	valuePtr := NewValue(fieldValue)

	fieldPtr := &field{}
	fieldPtr.index = idx

	fieldPtr.name = fieldType.Name
	if specPtr.GetFieldName() != "" {
		fieldPtr.name = specPtr.GetFieldName()
	}

	fieldPtr.typePtr = typePtr
	fieldPtr.specPtr = specPtr
	fieldPtr.valuePtr = valuePtr

	ret = fieldPtr
	return
}
