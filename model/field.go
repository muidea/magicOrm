package model

import (
	"fmt"
	"log"
	"reflect"

	"muidea.com/magicOrm/util"
)

// Field Field
type Field interface {
	GetIndex() int
	GetName() string
	GetType() FieldType
	GetTag() FieldTag
	GetValue() FieldValue
	SetValue(val reflect.Value) error
	Copy() Field
	Dump() string
}

// fieldInfo single field info
type fieldInfo struct {
	fieldIndex int
	fieldName  string

	fieldType  FieldType
	fieldTag   FieldTag
	fieldValue FieldValue
}

// Fields field info collection
type Fields []Field

func (s *fieldInfo) GetIndex() int {
	return s.fieldIndex
}

// GetName GetName
func (s *fieldInfo) GetName() string {
	return s.fieldName
}

// GetType GetType
func (s *fieldInfo) GetType() FieldType {
	return s.fieldType
}

// GetTag GetTag
func (s *fieldInfo) GetTag() FieldTag {
	return s.fieldTag
}

// GetValue GetValue
func (s *fieldInfo) GetValue() FieldValue {
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
		err = s.fieldValue.SetValue(val)
	} else {
		s.fieldValue, err = NewFieldValue(val.Addr())
	}

	return
}

// Verify Verify
func (s *fieldInfo) Verify() error {
	if s.fieldTag.Name() == "" {
		return fmt.Errorf("no define field tag")
	}

	if s.fieldTag.IsAutoIncrement() {
		switch s.fieldType.Value() {
		case util.TypeBooleanField, util.TypeStringField, util.TypeDateTimeField, util.TypeFloatField, util.TypeDoubleField, util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal auto_increment field type, type:%s", s.fieldType)
		default:
		}
	}

	if s.fieldTag.IsPrimaryKey() {
		switch s.fieldType.Value() {
		case util.TypeStructField, util.TypeSliceField:
			return fmt.Errorf("illegal primary key field type, type:%s", s.fieldType)
		default:
		}
	}

	return nil
}

func (s *fieldInfo) Copy() Field {
	var fieldValue FieldValue
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

// Append Append
func (s *Fields) Append(fieldInfo Field) {
	exist := false
	newField := fieldInfo.GetTag()
	for _, val := range *s {
		curField := val.GetTag()
		if curField.Name() == newField.Name() {
			exist = true
			break
		}
	}
	if exist {
		log.Fatalf("duplicate field tag,[%s]", fieldInfo.Dump())
	}

	*s = append(*s, fieldInfo)
}

// GetPrimaryField get primarykey field
func (s *Fields) GetPrimaryField() Field {
	for _, val := range *s {
		fieldTag := val.GetTag()
		if fieldTag.IsPrimaryKey() {
			return val
		}
	}

	return nil
}

// Copy Copy
func (s *Fields) Copy() Fields {
	ret := make(Fields, 0)
	for _, val := range *s {
		ret = append(ret, val.Copy())
	}
	return ret
}

// Dump Dump
func (s *Fields) Dump() {
	for _, v := range *s {
		fmt.Printf("\t%s\n", v.Dump())
	}
}

// GetFieldInfo GetFieldInfo
func GetFieldInfo(idx int, fieldType reflect.StructField, fieldVal *reflect.Value) (ret Field, err error) {
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
