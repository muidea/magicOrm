package model

import (
	"log"
	"reflect"
)

// Field Field
type Field interface {
	// Index
	GetIndex() int
	// Name
	GetName() string
	// Type
	GetType() Type
	// Tag
	GetTag() Tag
	// Value
	GetValue() Value
	// 是否主键
	IsPrimary() bool
	// 是否已赋值
	IsAssigned() bool
	// 设置值
	SetValue(val reflect.Value) error
	// 更新值
	UpdateValue(val reflect.Value) error
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

// GetPrimaryField get primary key field
func (s *Fields) GetPrimaryField() Field {
	for _, val := range *s {
		if val.IsPrimary() {
			return val
		}
	}

	return nil
}
