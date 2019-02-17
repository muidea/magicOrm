package remote

import (
	"reflect"

	"muidea.com/magicOrm/model"
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
func (s *ItemValue) Get() (ret reflect.Value, err error) {
	return
}

// GetDepend GetDepend
func (s *ItemValue) GetDepend() (ret []reflect.Value, err error) {
	return
}

// GetValueStr GetValueStr
func (s *ItemValue) GetValueStr() (ret string, err error) {
	return
}

// Copy Copy
func (s *ItemValue) Copy() (ret model.Value) {
	return
}
