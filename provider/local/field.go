package local

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
)

// field single field impl
type field struct {
	index int
	name  string

	typePtr  *typeImpl
	specPtr  *specImpl
	valuePtr *valueImpl
}

func (s *field) GetIndex() int {
	return s.index
}

func (s *field) GetName() string {
	return s.name
}

func (s *field) GetType() (ret model.Type) {
	if s.typePtr != nil {
		ret = s.typePtr
	}

	return
}

func (s *field) GetSpec() (ret model.Spec) {
	if s.specPtr != nil {
		ret = s.specPtr
	}

	return
}

func (s *field) GetValue() (ret model.Value) {
	if s.valuePtr != nil {
		ret = s.valuePtr
	}

	return
}

func (s *field) SetValue(val model.Value) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("SetValue failed, unexpect field:%v, err:%v", s.name, err)
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
		val.valuePtr = s.valuePtr.copy()
	}

	return val
}

func (s *field) verify() error {
	if s.typePtr == nil {
		return fmt.Errorf("illegal filed")
	}

	if s.specPtr != nil {
		val := s.typePtr.GetValue()
		if s.specPtr.IsAutoIncrement() {
			switch val {
			case util.TypeBooleanValue,
				util.TypeStringValue,
				util.TypeDateTimeValue,
				util.TypeFloatValue,
				util.TypeDoubleValue,
				util.TypeStructValue,
				util.TypeSliceValue:
				return fmt.Errorf("illegal auto_increment field type, type:%s", s.typePtr.dump())
			default:
			}
		}

		if s.specPtr.IsPrimaryKey() {
			switch val {
			case util.TypeStructValue, util.TypeSliceValue:
				return fmt.Errorf("illegal primary key field type, type:%s", s.typePtr.dump())
			default:
			}
		}
	}

	if s.valuePtr == nil || s.valuePtr.IsNil() {
		return nil
	}

	return s.valuePtr.verify()
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

	valuePtr := newValue(fieldValue)

	fieldPtr := &field{}
	fieldPtr.index = idx

	fieldPtr.name = fieldType.Name
	if specPtr.GetFieldName() != "" {
		fieldPtr.name = specPtr.GetFieldName()
	}

	fieldPtr.typePtr = typePtr
	fieldPtr.specPtr = specPtr
	fieldPtr.valuePtr = valuePtr

	//err = fieldPtr.verify()
	//if err != nil {
	//	log.Errorf("illegal fieldPtr info, err:%s", err.Error())
	//	return
	//}

	ret = fieldPtr
	return
}
