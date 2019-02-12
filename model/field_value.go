package model

import "reflect"

// FieldValue FieldValue
type FieldValue interface {
	IsNil() bool
	Set(val reflect.Value) error
	Get() (reflect.Value, error)
	GetDepend() ([]reflect.Value, error)
	GetValueStr() (string, error)
	Copy() FieldValue
}
