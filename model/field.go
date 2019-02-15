package model

import (
	"log"
	"reflect"
)

// Field Field
type Field interface {
	GetIndex() int
	GetName() string
	GetType() FieldType
	GetTag() FieldTag
	GetValue() FieldValue
	SetValue(val reflect.Value) error
	GetDepend() (Model, error)
	IsPrimary() bool
	Copy() Field
}

// Fields field info collection
type Fields []Field

// Append Append
func (s *Fields) Append(fieldInfo Field) {
	exist := false
	newField := fieldInfo.GetTag()
	for _, val := range *s {
		curField := val.GetTag()
		if curField.GetName() == newField.GetName() {
			exist = true
			break
		}
	}
	if exist {
		log.Printf("duplicate field tag,[%s]", newField.GetName())
		return
	}

	*s = append(*s, fieldInfo)
}

// GetPrimaryField get primarykey field
func (s *Fields) GetPrimaryField() Field {
	for _, val := range *s {
		if val.IsPrimary() {
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
