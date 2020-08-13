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
	Interface() reflect.Value
}

func CompareModel(l, r Model) bool {
	if l.GetName() != r.GetName() || l.GetPkgPath() != r.GetPkgPath() {
		return false
	}

	lFields := l.GetFields()
	rFields := r.GetFields()
	if len(lFields) != len(rFields) {
		return false
	}

	for idx := range lFields {
		if !CompareField(lFields[idx], rFields[idx]) {
			return false
		}
	}

	return true
}
