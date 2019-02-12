package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
	"muidea.com/magicOrm/util"
)

// fieldInfo single field info
type fieldInfo struct {
	fieldIndex int
	fieldName  string

	fieldType  model.FieldType
	fieldTag   model.FieldTag
	fieldValue model.FieldValue
}

func (s *fieldInfo) GetIndex() int {
	return s.fieldIndex
}

// GetName GetName
func (s *fieldInfo) GetName() string {
	return s.fieldName
}

// GetType GetType
func (s *fieldInfo) GetType() model.FieldType {
	return s.fieldType
}

// GetTag GetTag
func (s *fieldInfo) GetTag() model.FieldTag {
	return s.fieldTag
}

// GetValue GetValue
func (s *fieldInfo) GetValue() model.FieldValue {
	return s.fieldValue
}

// SetValue SetValue
func (s *fieldInfo) SetValue(val reflect.Value) (err error) {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
	}

	if s.fieldValue != nil {
		err = s.fieldValue.Set(val)
	} else {
		s.fieldValue, err = NewFieldValue(val.Addr())
	}

	return
}

// Verify Verify
func (s *fieldInfo) Verify() error {
	if s.fieldTag.GetName() == "" {
		return fmt.Errorf("no define field tag")
	}

	if s.fieldTag.IsAutoIncrement() {
		switch s.fieldType.GetValue() {
		case util.TypeBooleanField, util.TypeStringField, util.TypeDateTimeField, util.TypeFloatField, util.TypeDoubleField, util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.fieldType)
		default:
		}
	}

	if s.fieldTag.IsPrimaryKey() {
		switch s.fieldType.GetValue() {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.fieldType)
		default:
		}
	}

	return nil
}

func (s *fieldInfo) Copy() model.Field {
	var fieldValue model.FieldValue
	if s.fieldValue != nil {
		fieldValue = s.fieldValue.Copy()
	}

	return &fieldInfo{
		fieldIndex: s.fieldIndex,
		fieldName:  s.fieldName,
		fieldType:  s.fieldType.Copy(),
		fieldTag:   s.fieldTag.Copy(),
		fieldValue: fieldValue,
	}
}

// Dump Dump
func (s *fieldInfo) Dump() string {
	str := fmt.Sprintf("index:[%d],name:[%s],type:[%s],tag:[%s]", s.fieldIndex, s.fieldName, s.fieldType, s.fieldTag)
	if s.fieldValue != nil {
		valStr, _ := s.fieldValue.GetValueStr()

		str = fmt.Sprintf("%s,value:[%s]", str, valStr)
	}

	return str
}

// GetFieldInfo GetFieldInfo
func GetFieldInfo(idx int, fieldType reflect.StructField, fieldVal *reflect.Value) (ret model.Field, err error) {
	ormStr := fieldType.Tag.Get("orm")
	if ormStr == "" {
		return
	}

	info := &fieldInfo{}
	info.fieldIndex = idx
	info.fieldName = fieldType.Name

	info.fieldType, err = NewFieldType(fieldType.Type)
	if err != nil {
		return
	}

	info.fieldTag, err = NewFieldTag(ormStr)
	if err != nil {
		return
	}

	if fieldVal != nil {
		info.fieldValue, err = NewFieldValue(fieldVal.Addr())
		if err != nil {
			return
		}
	}

	err = info.Verify()
	if err != nil {
		return
	}

	ret = info
	return
}
