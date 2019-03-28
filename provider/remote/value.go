package remote

import (
	"fmt"
	"reflect"
)

// ValueImpl ValueImpl
type ValueImpl struct {
	value reflect.Value
}

// IsNil IsNil
func (s *ValueImpl) IsNil() (ret bool) {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	ret = s.value.Kind() == reflect.Invalid

	return
}

// Set Set
func (s *ValueImpl) Set(val reflect.Value) (err error) {
	if val.Kind() == reflect.Invalid {
		err = fmt.Errorf("invalid set value")
		return
	}

	s.value = val
	return
}

// Update Update
func (s *ValueImpl) Update(val reflect.Value) (err error) {
	if s.value.Kind() == reflect.Invalid {
		err = fmt.Errorf("invalid current value")
		return
	}

	if val.Kind() == reflect.Invalid {
		err = fmt.Errorf("invalid update value")
		return
	}

	s.value.Set(val)

	return
}

// Get Get
func (s *ValueImpl) Get() (ret reflect.Value) {
	ret = s.value

	return
}

// Copy Copy
func (s *ValueImpl) Copy() (ret *ValueImpl) {
	ret = &ValueImpl{value: s.value}

	return
}
