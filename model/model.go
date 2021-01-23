package model

// Model Model
type Model interface {
	GetName() string
	GetPkgPath() string
	GetFields() Fields
	SetFieldValue(name string, val Value) error
	GetPrimaryField() Field
	GetField(name string) Field
	Interface(ptrValue bool) interface{}
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
