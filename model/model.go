package model

import cd "github.com/muidea/magicCommon/def"

type Model interface {
	GetName() string
	GetPkgPath() string
	GetDescription() string
	GetPkgKey() string
	GetFields() Fields
	SetFieldValue(name string, val Value) *cd.Result
	GetPrimaryField() Field
	GetField(name string) Field
	Interface(ptrValue bool) any
	Copy() Model
	Dump() string
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
