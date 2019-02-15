package model

import "reflect"

// FieldType FieldType
type FieldType interface {
	GetName() string
	GetValue() (int, error)
	GetPkgPath() string
	GetType() reflect.Type
	GetDepend() (Model, error)
	IsPtrType() bool
	Dump() string
}
