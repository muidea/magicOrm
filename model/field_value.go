package model

import "reflect"

// FieldValue FieldValue
type FieldValue interface {
	IsNil() bool
	Set(val reflect.Value) error
	Get() (reflect.Value, error)
	Str() (string, error)
	Dump() string
}
