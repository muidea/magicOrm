package remote

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/util"
)

// ValueImpl ValueImpl
type ValueImpl struct {
	value reflect.Value
}

// IsNil IsNil
func (s *ValueImpl) IsNil() (ret bool) {
	ret = util.IsNil(s.value)

	return
}

// Set Set
func (s *ValueImpl) Set(val reflect.Value) (err error) {
	if util.IsNil(val) {
		err = fmt.Errorf("invalid set value")
		return
	}

	s.value = val
	return
}

// Update Update
func (s *ValueImpl) Update(val reflect.Value) (err error) {
	if util.IsNil(s.value) {
		err = fmt.Errorf("invalid current value")
		return
	}

	if util.IsNil(val) {
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
