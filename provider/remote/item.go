package remote

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/helper"
)

// Item Item
type Item struct {
	Index int    `json:"index"`
	Name  string `json:"name"`

	Tag   TagImpl  `json:"tag"`
	Type  TypeImpl `json:"type"`
	value ValueImpl
}

// GetIndex GetIndex
func (s *Item) GetIndex() (ret int) {
	return s.Index
}

// GetName GetName
func (s *Item) GetName() string {
	return s.Name
}

// GetType GetType
func (s *Item) GetType() (ret model.Type) {
	ret = &s.Type
	return
}

// GetTag GetTag
func (s *Item) GetTag() (ret model.Tag) {
	ret = &s.Tag
	return
}

// GetValue GetValue
func (s *Item) GetValue() (ret model.Value) {
	ret = &s.value
	return
}

// IsPrimary IsPrimary
func (s *Item) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}

// SetValue SetValue
func (s *Item) SetValue(val reflect.Value) (err error) {
	err = s.value.Set(val)
	return
}

// UpdateValue UpdateValue
func (s *Item) UpdateValue(val reflect.Value) (err error) {
	val = reflect.Indirect(val)
	fieldVal := reflect.Indirect(s.Type.Interface())
	valErr := helper.ConvertValue(val, &fieldVal)
	if valErr != nil {
		err = valErr
		return
	}

	if s.Type.IsPtrType() {
		fieldVal = fieldVal.Addr()
	}

	err = s.value.Update(fieldVal)

	return
}

// Copy Copy
func (s *Item) Copy() (ret model.Field) {
	return &Item{Index: s.Index, Name: s.Name, Tag: *(s.Tag.Copy()), Type: *(s.Type.Copy())}
}

// Dump Dump
func (s *Item) Dump() (ret string) {
	return
}
