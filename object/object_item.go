package object

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"muidea.com/magicOrm/util"
)

// Item Item
type Item struct {
	Name       string `json:"name"`
	Tag        string `json:"tag"`
	Type       int    `json:"type"`
	IsPtr      bool   `json:"isPtr"`
	DependInfo *Info  `json:"dependInfo"`
	value      interface{}
}

// GetVal get value
func (s *Item) GetVal() interface{} {
	return s.value
}

// SetVal set value
func (s *Item) SetVal(val interface{}) (err error) {
	rawVal := reflect.ValueOf(val)
	rawVal = reflect.Indirect(rawVal)
	switch rawVal.Kind() {
	case reflect.Bool:
		switch s.Type {
		case util.TypeBooleanField:
			s.value = rawVal.Interface().(bool)
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type, rawVal.Type().String())
			return
		}
	case reflect.Float64:
		switch s.Type {
		case util.TypeBitField:
			s.value = int8(rawVal.Interface().(float64))
		case util.TypePositiveBitField:
			s.value = uint8(rawVal.Interface().(float64))
		case util.TypeSmallIntegerField:
			s.value = int16(rawVal.Interface().(float64))
		case util.TypePositiveSmallIntegerField:
			s.value = uint16(rawVal.Interface().(float64))
		case util.TypeInteger32Field:
			s.value = int32(rawVal.Interface().(float64))
		case util.TypePositiveInteger32Field:
			s.value = uint32(rawVal.Interface().(float64))
		case util.TypeBigIntegerField:
			s.value = int64(rawVal.Interface().(float64))
		case util.TypePositiveBigIntegerField:
			s.value = uint64(rawVal.Interface().(float64))
		case util.TypeIntegerField:
			s.value = int(rawVal.Interface().(float64))
		case util.TypePositiveIntegerField:
			s.value = uint(rawVal.Interface().(float64))
		case util.TypeFloatField:
			s.value = float32(rawVal.Interface().(float64))
		case util.TypeDoubleField:
			s.value = rawVal.Interface().(float64)
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type, rawVal.Type().String())
			return
		}
	case reflect.String:
		switch s.Type {
		case util.TypeStringField:
			s.value = rawVal.Interface().(string)
		case util.TypeDateTimeField:
			ts, tsErr := time.ParseInLocation(time.RFC3339, val.(string), time.Local)
			if tsErr != nil {
				err = tsErr
				return
			}
			s.value = ts
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type, rawVal.Type().String())
			return
		}
	case reflect.Map:
		switch s.Type {
		case util.TypeStructField:
			s.value = val
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type, rawVal.Type().String())
			return
		}
	case reflect.Slice:
		switch s.Type {
		case util.TypeSliceField:
			s.value = val
		default:
			err = fmt.Errorf("illegal value type. item type:%d, value type:%s", s.Type, rawVal.Type().String())
			return
		}
	default:
		err = fmt.Errorf("illegal value type, type:%s", rawVal.Kind())
	}
	return
}

// GetFieldName GetFieldName
func (s *Item) GetFieldName() string {
	items := strings.Split(s.Tag, " ")
	if len(items) < 1 {
		return ""
	}

	return items[0]
}

// GetDepend GetDepend
func (s *Item) GetDepend() *Info {
	return s.DependInfo
}

// IsPrimary is primary key
func (s *Item) IsPrimary() bool {
	items := strings.Split(s.Tag, " ")
	if len(items) < 1 {
		return false
	}

	isPrimaryKey := false
	if len(items) >= 2 {
		switch items[1] {
		case "key":
			isPrimaryKey = true
		}
	}
	if len(items) >= 3 {
		switch items[2] {
		case "key":
			isPrimaryKey = true
		}
	}
	return isPrimaryKey
}

// IsAutoIncrement is autoincrement
func (s *Item) IsAutoIncrement() bool {
	items := strings.Split(s.Tag, " ")
	if len(items) < 1 {
		return false
	}

	isAutoIncrement := false
	if len(items) >= 2 {
		switch items[1] {
		case "auto":
			isAutoIncrement = true
		}
	}
	if len(items) >= 3 {
		switch items[2] {
		case "auto":
			isAutoIncrement = true
		}
	}
	return isAutoIncrement
}
