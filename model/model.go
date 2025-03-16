package model

import cd "github.com/muidea/magicCommon/def"

type Model interface {
	GetName() string
	GetShowName() string
	GetPkgPath() string
	GetDescription() string
	GetFields() Fields
	SetFieldValue(name string, val any) *cd.Result
	SetPrimaryFieldValue(val any) *cd.Result
	GetPrimaryField() Field
	GetField(name string) Field
	Interface(ptrValue bool) any
	Copy(viewSpec ViewDeclare) Model
	Reset()
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
