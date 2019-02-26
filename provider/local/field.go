package local

import (
	"fmt"
	"reflect"
	"time"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

// fieldImpl single field impl
type fieldImpl struct {
	fieldIndex int
	fieldName  string

	fieldType  typeImpl
	fieldTag   tagImpl
	fieldValue valueImpl
}

func (s *fieldImpl) GetIndex() int {
	return s.fieldIndex
}

// GetName GetName
func (s *fieldImpl) GetName() string {
	return s.fieldName
}

// GetType GetType
func (s *fieldImpl) GetType() model.Type {
	return &s.fieldType
}

// GetTag GetTag
func (s *fieldImpl) GetTag() model.Tag {
	return &s.fieldTag
}

// GetValue GetValue
func (s *fieldImpl) GetValue() model.Value {
	return &s.fieldValue
}

func (s *fieldImpl) IsPrimary() bool {
	return s.fieldTag.IsPrimaryKey()
}

func (s *fieldImpl) UpdateValue(val reflect.Value) (err error) {
	val = reflect.Indirect(val)
	valType, valErr := util.GetTypeValueEnum(val.Type())
	if valErr != nil {
		err = valErr
		return
	}

	fieldVal := reflect.New(s.fieldType.GetType()).Elem()
	switch s.fieldType.GetValue() {
	case util.TypeBooleanField:
		switch valType {
		case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
			bVal := val.Int() > 0
			fieldVal.SetBool(bVal)
		case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
			bVal := val.Uint() > 0
			fieldVal.SetBool(bVal)
		case util.TypeBooleanField:
			fieldVal.Set(val)
		default:
			err = fmt.Errorf("illegal value type,current type:%d, expect type:%d", valType, s.fieldType.GetValue())
		}
	case util.TypeStringField:
		switch valType {
		case util.TypeStringField:
			fieldVal.Set(val)
		default:
			err = fmt.Errorf("illegal value type,current type:%d, expect type:%d", valType, s.fieldType.GetValue())
		}
	case util.TypeDateTimeField:
		switch valType {
		case util.TypeStringField:
			tmVal, tmErr := time.ParseInLocation("2006-01-02 15:04:05", val.String(), time.Local)
			if tmErr != nil {
				err = tmErr
			} else {
				fieldVal.Set(reflect.ValueOf(tmVal))
			}
		case util.TypeDateTimeField:
			fieldVal.Set(val)
		default:
			err = fmt.Errorf("illegal value type,current type:%d, expect type:%d", valType, s.fieldType.GetValue())
		}
	case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
		switch valType {
		case util.TypeBitField, util.TypeSmallIntegerField, util.TypeInteger32Field, util.TypeIntegerField, util.TypeBigIntegerField:
			fieldVal.SetInt(val.Int())
		default:
			err = fmt.Errorf("illegal value type,current type:%d, expect type:%d", valType, s.fieldType.GetValue())
		}
	case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
		switch valType {
		case util.TypePositiveBitField, util.TypePositiveSmallIntegerField, util.TypePositiveInteger32Field, util.TypePositiveIntegerField, util.TypePositiveBigIntegerField:
			fieldVal.SetUint(val.Uint())
		default:
			err = fmt.Errorf("illegal value type,current type:%d, expect type:%d", valType, s.fieldType.GetValue())
		}
	case util.TypeFloatField, util.TypeDoubleField:
		switch valType {
		case util.TypeFloatField, util.TypeDoubleField:
			fieldVal.SetFloat(val.Float())
		default:
			err = fmt.Errorf("illegal value type,current type:%d, expect type:%d", valType, s.fieldType.GetValue())
		}
	case util.TypeStructField, util.TypeSliceField:
		if val.Type().String() == s.fieldType.GetType().String() {
			fieldVal.Set(val)
		} else {
			err = fmt.Errorf("illegal value type,current type:%d, expect type:%d", valType, s.fieldType.GetValue())
		}
	}

	if err == nil {
		err = s.fieldValue.Set(fieldVal)
	}

	return
}

// Verify Verify
func (s *fieldImpl) Verify() error {
	if s.fieldTag.GetName() == "" {
		return fmt.Errorf("no define field tag")
	}

	val := s.fieldType.GetValue()
	if s.fieldTag.IsAutoIncrement() {
		switch val {
		case util.TypeBooleanField, util.TypeStringField, util.TypeDateTimeField, util.TypeFloatField, util.TypeDoubleField, util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.fieldType.Dump())
		default:
		}
	}

	if s.fieldTag.IsPrimaryKey() {
		switch val {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.fieldType.Dump())
		default:
		}
	}

	return nil
}

func (s *fieldImpl) Copy() *fieldImpl {
	return &fieldImpl{
		fieldIndex: s.fieldIndex,
		fieldName:  s.fieldName,
		fieldType:  s.fieldType,
		fieldTag:   s.fieldTag,
		fieldValue: s.fieldValue,
	}
}

// Dump Dump
func (s *fieldImpl) Dump(cache Cache) string {
	str, _ := getValueStr(&s.fieldType, &s.fieldValue, cache)
	str = fmt.Sprintf("index:[%d],name:[%s],type:[%s],tag:[%s],value:[%s]", s.fieldIndex, s.fieldName, s.fieldType.Dump(), s.fieldTag.Dump(), str)

	return str
}
func getFieldInfo(idx int, fieldType reflect.StructField) (ret *fieldImpl, err error) {
	typeImpl, err := newType(fieldType.Type)
	if err != nil {
		return
	}

	tagImpl, err := newTag(fieldType.Tag.Get("orm"))
	if err != nil {
		return
	}

	field := &fieldImpl{}
	field.fieldIndex = idx
	field.fieldName = fieldType.Name
	field.fieldType = *typeImpl
	field.fieldTag = *tagImpl

	err = field.Verify()
	if err != nil {
		return
	}

	ret = field
	return
}
