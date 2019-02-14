package remote

import (
	"fmt"
	"reflect"
	"time"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
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
func (s *Item) GetType() (ret model.FieldType) {
	ret = &s.Type
	return
}

// GetTag GetTag
func (s *Item) GetTag() (ret model.FieldTag) {
	ret = &s.Tag
	return
}

// GetValue GetValue
func (s *Item) GetValue() (ret model.FieldValue) {
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

// GetVal get value
func (s *Item) GetVal() interface{} {
	return s.value.Value
}

// SetVal set value
func (s *Item) SetVal(val interface{}) (err error) {
	rawVal := reflect.ValueOf(val)
	rawVal = reflect.Indirect(rawVal)
	switch rawVal.Kind() {
	case reflect.Bool:
		switch s.Type.Value {
		case util.TypeBooleanField:
			s.value.Value = rawVal.Interface().(bool)
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type.Value, rawVal.Type().String())
			return
		}
	case reflect.Float64:
		switch s.Type.Value {
		case util.TypeBitField:
			s.value.Value = int8(rawVal.Interface().(float64))
		case util.TypePositiveBitField:
			s.value.Value = uint8(rawVal.Interface().(float64))
		case util.TypeSmallIntegerField:
			s.value.Value = int16(rawVal.Interface().(float64))
		case util.TypePositiveSmallIntegerField:
			s.value.Value = uint16(rawVal.Interface().(float64))
		case util.TypeInteger32Field:
			s.value.Value = int32(rawVal.Interface().(float64))
		case util.TypePositiveInteger32Field:
			s.value.Value = uint32(rawVal.Interface().(float64))
		case util.TypeBigIntegerField:
			s.value.Value = int64(rawVal.Interface().(float64))
		case util.TypePositiveBigIntegerField:
			s.value.Value = uint64(rawVal.Interface().(float64))
		case util.TypeIntegerField:
			s.value.Value = int(rawVal.Interface().(float64))
		case util.TypePositiveIntegerField:
			s.value.Value = uint(rawVal.Interface().(float64))
		case util.TypeFloatField:
			s.value.Value = float32(rawVal.Interface().(float64))
		case util.TypeDoubleField:
			s.value.Value = rawVal.Interface().(float64)
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type.Value, rawVal.Type().String())
			return
		}
	case reflect.String:
		switch s.Type.Value {
		case util.TypeStringField:
			s.value.Value = rawVal.Interface().(string)
		case util.TypeDateTimeField:
			ts, tsErr := time.ParseInLocation(time.RFC3339, val.(string), time.Local)
			if tsErr != nil {
				err = tsErr
				return
			}
			s.value.Value = ts
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type.Value, rawVal.Type().String())
			return
		}
	case reflect.Map:
		switch s.Type.Value {
		case util.TypeStructField:
			s.value.Value = val
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type.Value, rawVal.Type().String())
			return
		}
	case reflect.Slice:
		switch s.Type.Value {
		case util.TypeSliceField:
			s.value.Value = val
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type.Value, rawVal.Type().String())
			return
		}
	default:
		err = fmt.Errorf("illegal value type, type:%s", rawVal.Kind())
	}
	return
}

// GetDepend GetDepend
func (s *Item) GetDepend() *Object {
	return s.Type.Depend
}

// IsPrimary is primary key
func (s *Item) IsPrimary() bool {
	return s.Tag.IsPrimaryKey()
}
