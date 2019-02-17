package model

import "reflect"

// Type Type
type Type interface {
	GetName() string
	GetValue() int
	GetPkgPath() string
	GetType() reflect.Type
	IsPtrType() bool
}
