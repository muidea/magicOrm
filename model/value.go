package model

import (
	"fmt"

	cd "github.com/muidea/magicCommon/def"
)

// Value Value
type Value interface {
	IsValid() bool
	IsZero() bool
	Get() any
	Set(val any) *cd.Error
	UnpackValue() []Value
}

func CompareValue(l, r Value) bool {
	if l != nil && r != nil {
		if l.IsValid() != r.IsValid() {
			return false
		}
		return fmt.Sprintf("+%v", l.Get()) == fmt.Sprintf("+%v", r.Get())
	}

	if l == nil && r == nil {
		return true
	}

	if l == nil && r != nil {
		return !r.IsValid()
	}

	if l != nil && r == nil {
		return !l.IsValid()
	}

	return false
}
