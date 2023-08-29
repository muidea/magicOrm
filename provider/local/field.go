package local

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	pu "github.com/muidea/magicOrm/provider/util"
)

// field single field impl
type field struct {
	index int
	name  string

	typePtr  *typeImpl
	specPtr  *specImpl
	valuePtr *pu.ValueImpl
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

	ret = &pu.NilValue
	return
}

func (s *field) SetValue(val model.Value) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("SetValue failed, unexpected field:%v, err:%v", s.name, err)
		}
	}()

	err = s.valuePtr.Set(val.Get())
	if err != nil {
		log.Errorf("set field valuePtr failed, name:%s, err:%s", s.name, err.Error())
	}

	return
}

func (s *field) IsPrimary() bool {
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

func (s *field) verify() error {
	if s.typePtr == nil {
		return fmt.Errorf("illegal filed, field type is null, index:%d, name:%v", s.index, s.name)
	}

	if s.specPtr != nil {
		val := s.typePtr.GetValue()
		if s.specPtr.GetValueDeclare() == model.AutoIncrement {
			switch val {
			case model.TypeBooleanValue,
				model.TypeStringValue,
				model.TypeDateTimeValue,
				model.TypeFloatValue,
				model.TypeDoubleValue,
				model.TypeStructValue,
				model.TypeSliceValue:
				return fmt.Errorf("illegal auto_increment field type, type:%s", s.typePtr.dump())
			default:
			}
		}
		if s.specPtr.GetValueDeclare() == model.UUID && val != model.TypeStringValue {
			return fmt.Errorf("illegal uuid field type, type:%s", s.typePtr.dump())
		}

		if s.specPtr.GetValueDeclare() == model.SnowFlake && val != model.TypeBigIntegerValue {
			return fmt.Errorf("illegal snowflake field type, type:%s", s.typePtr.dump())
		}

		if s.specPtr.GetValueDeclare() == model.DateTime && val != model.TypeDateTimeValue {
			return fmt.Errorf("illegal dateTime field type, type:%s", s.typePtr.dump())
		}

		if s.specPtr.IsPrimaryKey() {
			switch val {
			case model.TypeStructValue, model.TypeSliceValue:
				return fmt.Errorf("illegal primary key field type, type:%s", s.typePtr.dump())
			default:
			}
		}
	}

	if s.valuePtr == nil || s.valuePtr.IsNil() {
		return nil
	}

	return s.valuePtr.Verify()
}

func (s *field) dump() string {
	str := fmt.Sprintf("index:%d,name:%s,type:[%s],spec:[%s]", s.index, s.name, s.typePtr.dump(), s.specPtr.dump())
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

func getFieldInfo(idx int, fieldType reflect.StructField, fieldValue reflect.Value) (ret *field, err error) {
	typePtr, typeErr := newType(fieldType.Type)
	if typeErr != nil {
		err = typeErr
		return
	}

	specPtr, specErr := newSpec(fieldType.Tag)
	if specErr != nil {
		err = specErr
		return
	}

	valuePtr := pu.NewValue(fieldValue)

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
