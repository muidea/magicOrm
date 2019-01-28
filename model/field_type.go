package model

import "reflect"

// FieldType FieldType
type FieldType interface {
	Name() string
	Value() int
	IsPtr() bool
	PkgPath() string
	String() string
	Depend() (dependType reflect.Type, isTypePtr bool)
	Copy() FieldType
}
