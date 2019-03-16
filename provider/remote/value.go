package remote

import (
	"reflect"
)

// ItemValue ItemValue
type ItemValue struct {
	Value interface{}
}

// IsNil IsNil
func (s *ItemValue) IsNil() (ret bool) {
	return
}

// Set Set
func (s *ItemValue) Set(val reflect.Value) (err error) {
	return
}

// Get Get
func (s *ItemValue) Get() (ret reflect.Value) {
	return
}
