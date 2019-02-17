package remote

import (
	"reflect"

	"muidea.com/magicOrm/model"
)

// Item Item
type Item struct {
	Index int    `json:"index"`
	Name  string `json:"name"`

	Tag   ItemTag  `json:"tag"`
	Type  ItemType `json:"type"`
	value ItemValue
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

// SetValue SetValue
func (s *Item) SetValue(val reflect.Value) (err error) {
	err = s.value.Set(val)
	return
}

// Copy Copy
func (s *Item) Copy() (ret model.Field) {
	return
}

// Dump Dump
func (s *Item) Dump() (ret string) {
	return
}

// IsPrimary IsPrimary
func (s *Item) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}
