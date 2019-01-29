package model

import "reflect"

// FieldType FieldType
type FieldType interface {
	Name() string
	Value() int
	IsPtr() bool
	PkgPath() string
	String() string
	Type() reflect.Type
	Depend() FieldType
	Copy() FieldType
}
