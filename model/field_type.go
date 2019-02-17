package model

import "reflect"

// FieldType FieldType
type FieldType interface {
	GetName() string
	GetValue() int
	GetPkgPath() string
	GetType() reflect.Type
	IsPtrType() bool
	Dump() string
}
