package model

import (
	"fmt"
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
	Copy() Field
	Dump() string
}

// Fields field info collection
type Fields []Field

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
		return
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
