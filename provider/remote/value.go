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

// Get Get
func (s *ItemValue) Get() (ret reflect.Value) {
	ret = s.value

	return
}

func (s *ItemValue) update(val reflect.Value) (err error) {
	if s.value.Kind() == reflect.Invalid {
		err = fmt.Errorf("invalid current value")
		return
	}

	if val.Kind() == reflect.Invalid {
		err = fmt.Errorf("invalid update value")
		return
	}

	valTypeName := val.Type().String()
	expectTypeName := s.value.Type().String()
	if expectTypeName != valTypeName {
		err = fmt.Errorf("illegal value type, type:%s, expect:%s", valTypeName, expectTypeName)
		return
	}

	s.value.Set(val)

	return
}
