package remote

import (
	"fmt"
	"reflect"
	"time"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
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
	switch s.Type.GetValue() {
	case util.TypeBooleanField:
		fieldVal.Set(val)
	case util.TypeStringField:
		fieldVal.SetString(val.String())
	case util.TypeDateTimeField:
		_, tmErr := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
		if tmErr != nil {
			err = tmErr
		} else {
			fieldVal.SetString(val.String())
		}
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		fieldVal.SetInt(val.Int())
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		fieldVal.SetUint(val.Uint())
	case util.TypeFloatField, util.TypeDoubleField:
		fieldVal.SetFloat(val.Float())
	case util.TypeStructField:
		if val.Type().String() == s.Type.GetType().String() {
			fieldVal.Set(val)
		} else {
			err = fmt.Errorf("UpdateValue failed, fieldName:%s,illegal value type,current type:%s, expect type:%s", s.Name, val.Type().String(), s.Type.GetType().String())
		}
	case util.TypeSliceField:
		if val.Type().String() == s.Type.GetType().String() {
			fieldVal.Set(val)
		} else {
			err = fmt.Errorf("UpdateValue failed, fieldName:%s,illegal value type,current type:%s, expect type:%s", s.Name, val.Type().String(), s.Type.GetType().String())
		}
	}

	if err == nil {
		if s.Type.IsPtrType() {
			fieldVal = fieldVal.Addr()
		}

		err = s.value.Update(fieldVal)
	}

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
