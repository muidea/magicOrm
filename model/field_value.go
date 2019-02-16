package model

import "reflect"

// FieldValue FieldValue
type FieldValue interface {
	IsNil() bool
	Set(val reflect.Value) error
	Get() reflect.Value
	Str() (string, error)
	Dump() string
}
