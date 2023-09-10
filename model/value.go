package model

// Value Value
type Value interface {
	IsNil() bool
	IsZero() bool
	Set(val any) error
	Get() any
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
