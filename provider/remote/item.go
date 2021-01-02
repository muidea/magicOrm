package remote

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicOrm/model"
)

// Item Item
type Item struct {
	Index int    `json:"index"`
	Name  string `json:"name"`

	Tag   *TagImpl  `json:"tag"`
	Type  *TypeImpl `json:"type"`
	value *ValueImpl
}

// GetIndex GetIndex
func (s *Item) GetIndex() (ret int) {
	return s.Index
}

// GetName GetName
func (s *Item) GetName() string {
	return s.Name
}

// GetEntityType GetEntityType
func (s *Item) GetType() (ret model.Type) {
	if s.Type != nil {
		ret = s.Type
	}
	return
}

// GetTag GetTag
func (s *Item) GetTag() (ret model.Tag) {
	if s.Tag != nil {
		ret = s.Tag
	}

	return
}

// GetEntityValue GetEntityValue
func (s *Item) GetValue() (ret model.Value) {
	if s.value != nil {
		ret = s.value
	}

	return
}

// IsPrimary IsPrimary
func (s *Item) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}

// SetValue SetValue
func (s *Item) SetValue(val model.Value) (err error) {
	if s.value != nil {
		err = s.value.Set(val.Get())
		if err != nil {
			log.Errorf("set field value failed, name:%s, err:%s", s.Name, err.Error())
		}
		return
	}

	s.value = &ValueImpl{value: val.Get()}
	return
}

// UpdateValue UpdateValue
func (s *Item) UpdateValue(val model.Value) (err error) {
	err = s.value.Update(val.Get())
	if err != nil {
		log.Errorf("update field value failed, name:%s, err:%s", s.Name, err.Error())
	}

	return
}

// copy copy
func (s *Item) copy() (ret model.Field) {
	return &Item{Index: s.Index, Name: s.Name, Tag: s.Tag, Type: s.Type, value: s.value}
}

func (s *Item) dump() string {
	str := fmt.Sprintf("index:%d,name:%s,type:[%s],tag:[%s]", s.Index, s.Name, s.Type.dump(), s.Tag.dump())
	return str
}

func getItemInfo(idx int, fieldType reflect.StructField) (ret *Item, err error) {
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

	item := &Item{}
	item.Index = idx
	item.Name = fieldType.Name
	item.Type = typeImpl
	item.Tag = tagImpl

	ret = item
	return
}

func compareItem(l, r *Item) bool {
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
