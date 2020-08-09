package model

import (
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
	// 更新值
	UpdateValue(val reflect.Value) error
}

// Fields field info collection
type Fields []Field

// Append Append
func (s *Fields) Append(fieldInfo Field) bool {
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
		return false
	}

	*s = append(*s, fieldInfo)
	return true
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
