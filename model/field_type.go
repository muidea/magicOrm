package model

import "reflect"

// FieldType FieldType
type FieldType interface {
	Name() string
	Value() int
	IsPtr() bool
	PkgPath() string
	String() string
	Depend() reflect.Type
	Copy() FieldType
}
