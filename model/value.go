package model

import "reflect"

// Value Value
type Value interface {
	IsNil() bool
	Set(val reflect.Value) error
	Get() reflect.Value
	Addr() Value
	Interface() any
	IsBasic() bool
}

func CompareValue(l, r Value) bool {
	if l != nil && r != nil {
		return l.IsNil() == r.IsNil()
	}

	if l == nil && r == nil {
		return true
	}

	if l == nil && r != nil {
		return r.IsNil()
	}

	if l != nil && r == nil {
		return l.IsNil()
	}

	return false
}
