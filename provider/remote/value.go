package remote

import (
	"fmt"
	"reflect"
)

// ItemValue ItemValue
type ItemValue struct {
	value reflect.Value
}

// IsNil IsNil
func (s *ItemValue) IsNil() (ret bool) {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	ret = s.value.Kind() == reflect.Invalid

	return
}

// Set Set
func (s *ItemValue) Set(val reflect.Value) (err error) {
	if val.Kind() == reflect.Invalid {
		err = fmt.Errorf("invalid set value")
		return
	}

	s.value = val
	return
}

// Update Update
func (s *ItemValue) Update(val reflect.Value) (err error) {
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
func (s *ItemValue) Get() (ret reflect.Value) {
	ret = s.value

	return
}

// Copy Copy
func (s *ItemValue) Copy() (ret *ItemValue) {
	ret = &ItemValue{value: s.value}

	return
}
