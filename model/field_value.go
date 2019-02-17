package model

import "reflect"

// Value Value
type Value interface {
	IsNil() bool
	Set(val reflect.Value) error
	Get() reflect.Value
	Str() (string, error)
}
