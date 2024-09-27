package model

// Value Value
type Value interface {
	IsValid() bool
	IsZero() bool
	Set(val any)
	Get() any
	Addr() Value
	Interface() any
	IsBasic() bool
}

func CompareValue(l, r Value) bool {
	if l != nil && r != nil {
		return !l.IsValid() == !r.IsValid()
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
