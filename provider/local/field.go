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
	Index int
	Name  string

	Type  *typeImpl
	Tag   *tagImpl
	value *valueImpl
}

func (s *field) GetIndex() int {
	return s.Index
}

// GetName GetName
func (s *field) GetName() string {
	return s.Name
}

// GetEntityType GetEntityType
func (s *field) GetType() (ret model.Type) {
	if s.Type != nil {
		ret = s.Type
	}

	return
}

// GetTag GetTag
func (s *field) GetTag() (ret model.Tag) {
	if s.Tag != nil {
		ret = s.Tag
	}

	return
}

// GetEntityValue GetEntityValue
func (s *field) GetValue() (ret model.Value) {
	if s.value != nil {
		ret = s.value
	}

	return
}

func (s *field) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}

func (s *field) SetValue(val model.Value) (err error) {
	err = s.value.Set(val.Get())
	if err != nil {
		log.Errorf("set field value failed, name:%s, err:%s", s.Name, err.Error())
	}

	return
}

func (s *field) UpdateValue(val model.Value) (err error) {
	err = s.value.Update(val.Get())
	if err != nil {
		log.Errorf("update field value failed, name:%s, err:%s", s.Name, err.Error())
	}

	return
}

func (s *field) copy() *field {
	return &field{
		Index: s.Index,
		Name:  s.Name,
		Type:  s.Type.copy(),
		Tag:   s.Tag.copy(),
		value: s.value.copy(),
	}
}

// verify verify
func (s *field) verify() error {
	if s.Tag.GetName() == "" {
		return fmt.Errorf("no define field tag")
	}

	val := s.Type.GetValue()
	if s.Tag.IsAutoIncrement() {
		switch val {
		case util.TypeBooleanField,
			util.TypeStringField,
			util.TypeDateTimeField,
			util.TypeFloatField,
			util.TypeDoubleField,
			util.TypeStructField,
			util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.Type.dump())
		default:
		}
	}

	if s.Tag.IsPrimaryKey() {
		switch val {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.Type.dump())
		default:
		}
	}

	return nil
}

// dump dump
func (s *field) dump() string {
	str := fmt.Sprintf("index:%d,name:%s,type:[%s],tag:[%s]", s.Index, s.Name, s.Type.dump(), s.Tag.dump())
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
	field.Index = idx
	field.Name = fieldType.Name
	field.Type = typeImpl
	field.Tag = tagImpl
	field.value = valueImpl

	err = field.verify()
	if err != nil {
		return
	}

	ret = field
	return
}
