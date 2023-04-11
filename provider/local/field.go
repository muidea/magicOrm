package local

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"

	log "github.com/cihub/seelog"
)

// field single field impl
type field struct {
	index int
	name  string

	typePtr  *typeImpl
	tagPtr   *tagImpl
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

func (s *field) GetTag() (ret model.Tag) {
	if s.tagPtr != nil {
		ret = s.tagPtr
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
	err = s.valuePtr.Set(val.Get())
	if err != nil {
		log.Errorf("set field valuePtr failed, name:%s, err:%s", s.name, err.Error())
	}

	return
}

func (s *field) IsPrimary() bool {
	return s.tagPtr.IsPrimaryKey()
}

func (s *field) copy() *field {
	return &field{
		index:    s.index,
		name:     s.name,
		typePtr:  s.typePtr.copy(),
		tagPtr:   s.tagPtr.copy(),
		valuePtr: s.valuePtr.copy(),
	}
}

func (s *field) verify() error {
	if s.tagPtr.GetName() == "" {
		return fmt.Errorf("no define field tag")
	}

	val := s.typePtr.GetValue()
	if s.tagPtr.IsAutoIncrement() {
		switch val {
		case util.TypeBooleanField,
			util.TypeStringField,
			util.TypeDateTimeField,
			util.TypeFloatField,
			util.TypeDoubleField,
			util.TypeStructField,
			util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.typePtr.dump())
		default:
		}
	}

	if s.tagPtr.IsPrimaryKey() {
		switch val {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.typePtr.dump())
		default:
		}
	}

	return nil
}

func (s *field) dump() string {
	str := fmt.Sprintf("index:%d,name:%s,type:[%s],tag:[%s]", s.index, s.name, s.typePtr.dump(), s.tagPtr.dump())
	return str
}

func getFieldInfo(idx int, fieldType reflect.StructField, fieldValue reflect.Value) (ret *field, err error) {
	typeImpl, typeErr := newType(fieldType.Type)
	if typeErr != nil {
		err = typeErr
		return
	}

	tagImpl, tagErr := newTag(fieldType.Tag.Get("orm"))
	if tagErr != nil {
		err = tagErr
		return
	}

	valueImpl := newValue(fieldValue)

	field := &field{}
	field.index = idx
	field.name = fieldType.Name
	field.typePtr = typeImpl
	field.tagPtr = tagImpl
	field.valuePtr = valueImpl

	err = field.verify()
	if err != nil {
		return
	}

	ret = field
	return
}
