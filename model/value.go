package model

import "fmt"

// Value Value
type Value interface {
	IsValid() bool
	IsZero() bool
	Set(val any)
	Get() any
	Addr() Value
	Interface() RawVal
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

type RawVal interface {
	Value() any
}

type rawValImpl struct {
	rawVal any
}

func (s *rawValImpl) Value() any {
	return s.rawVal
}

func (s *rawValImpl) String() string {
	switch s.rawVal.(type) {
	case string:
		return fmt.Sprintf("'%v'", s.rawVal)
	default:
	}
	return fmt.Sprintf("%v", s.rawVal)
}

func NewRawVal(val any) RawVal {
	return &rawValImpl{
		rawVal: val,
	}
}
