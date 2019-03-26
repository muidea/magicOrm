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

// IsPrimary IsPrimary
func (s *Item) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}

// SetValue SetValue
func (s *Item) SetValue(val reflect.Value) (err error) {
	val = reflect.Indirect(val)
	typeVal, typeErr := util.GetTypeValueEnum(val.Type())
	if typeErr != nil {
		err = typeErr
		return
	}

	switch s.Type.GetValue() {
	case util.TypeBooleanField:
		if typeVal != util.TypeBooleanField {
			err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
		}
	case util.TypeStringField:
		if typeVal != util.TypeStringField {
			err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
		}
	case util.TypeDateTimeField:
		if typeVal == util.TypeStringField {
			_, tmErr := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
			if tmErr != nil {
				err = tmErr
			}
		} else {
			err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
		}
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		if typeVal != util.TypeDoubleField {
			err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
		}
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		if typeVal != util.TypeDoubleField {
			err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
		}
	case util.TypeFloatField, util.TypeDoubleField:
		if typeVal != util.TypeDoubleField {
			err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
		}
	case util.TypeStructField:
		if typeVal == util.TypeStructField {
			objVal, objOK := val.Interface().(ObjectValue)
			if !objOK {
				err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
			} else {
				if objVal.GetName() != s.Type.GetName() || objVal.GetPkgPath() != objVal.GetPkgPath() {
					err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
				}
			}
		} else {
			err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
		}
	case util.TypeSliceField:
		// TODO slice element miss match
		if typeVal == util.TypeSliceField {
			if val.Type().String() != s.Type.GetType().String() {
				err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%s, expect type:%s", s.Name, val.Type().String(), s.Type.GetType().String())
			}
		} else {
			err = fmt.Errorf("SetValue failed, fieldName:%s,illegal value type,current type:%d, expect type:%d", s.Name, typeVal, s.Type.GetValue())
		}
	}

	if err == nil {
		err = s.value.Set(val)
	}

	return
}

// UpdateValue UpdateValue
func (s *Item) UpdateValue(val reflect.Value) (err error) {
	val = reflect.Indirect(val)
	_, typeErr := util.GetTypeValueEnum(val.Type())
	if typeErr != nil {
		err = typeErr
		return
	}

	err = s.value.Update(val)

	return
}

// Copy Copy
func (s *Item) Copy() (ret model.Field) {
	return &Item{Index: s.Index, Name: s.Name, Tag: *(s.Tag.Copy()), Type: *(s.Type.Copy()), value: *(s.value.Copy())}
}

// Dump Dump
func (s *Item) Dump() (ret string) {
	return
}
