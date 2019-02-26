package model

import "reflect"

// Model Model
type Model interface {
	GetName() string
	GetPkgPath() string
	GetFields() Fields
	SetFieldValue(idx int, val reflect.Value) error
	UpdateFieldValue(name string, val reflect.Value) error
	GetPrimaryField() Field
	IsPtrModel() bool
	Interface() reflect.Value
}
