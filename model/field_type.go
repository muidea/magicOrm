package model

import "reflect"

// FieldType FieldType
type FieldType interface {
	GetName() string
	GetValue() int
	GetPkgPath() string
	GetType() reflect.Type
	GetDepend() reflect.Type
	IsPtrType() bool
	String() string
	Copy() FieldType
}
