package remote

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
)

type Field struct {
	Index int    `json:"index"`
	Name  string `json:"name"`

	Type  *TypeImpl `json:"type"`
	Tag   *TagImpl  `json:"tag"`
	value *valueImpl
}

func (s *Field) GetIndex() (ret int) {
	return s.Index
}

func (s *Field) GetName() string {
	return s.Name
}

func (s *Field) GetType() (ret model.Type) {
	if s.Type != nil {
		ret = s.Type
	}
	return
}

func (s *Field) GetTag() (ret model.Tag) {
	if s.Tag != nil {
		ret = s.Tag
	}

	return
}

func (s *Field) GetValue() (ret model.Value) {
	if s.value != nil {
		ret = s.value
		return
	}

	ret = s.Type.Interface()
	return
}

func (s *Field) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}

func (s *Field) SetValue(val model.Value) (err error) {
	if s.value != nil {
		err = s.value.Set(val.Get())
		if err != nil {
			log.Errorf("set field value failed, name:%s, err:%s", s.Name, err.Error())
		}
		return
	}

	initVal := s.Type.Interface()
	initVal.Set(val.Get())
	s.value = &valueImpl{value: initVal.Get()}
	return
}

func (s *Field) copy() (ret model.Field) {
	return &Field{Index: s.Index, Name: s.Name, Tag: s.Tag, Type: s.Type, value: s.value}
}

func (s *Field) dump() string {
	str := fmt.Sprintf("index:%d,name:%s,type:[%s],tag:[%s]", s.Index, s.Name, s.Type.dump(), s.Tag.dump())
	if s.value != nil {
		str = fmt.Sprintf("%s,value:%v", str, s.value.Interface())
	}

	return str
}

func getItemInfo(idx int, fieldType reflect.StructField) (ret *Field, err error) {
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

	initVal := typeImpl.Interface()

	item := &Field{}
	item.Index = idx
	item.Name = fieldType.Name
	item.Type = typeImpl
	item.Tag = tagImpl
	item.value = newValue(initVal.Get())

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
	if !compareTag(l.Tag, r.Tag) {
		return false
	}

	return true
}
