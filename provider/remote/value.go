package remote

import (
	"github.com/muidea/magicOrm/model"
)

// ValueImpl ValueImpl
type ValueImpl struct {
	value interface{}
}

func newValue(v interface{}) (ret *ValueImpl) {
	ret = &ValueImpl{value: v}
	return
}

// IsNil IsNil
func (s *ValueImpl) IsNil() (ret bool) {
	ret = s.value == nil

	return
}

// Set Set
func (s *ValueImpl) Set(val interface{}) (err error) {
	s.value = val
	return
}

// Get Get
func (s *ValueImpl) Get() (ret interface{}) {
	ret = s.value
	return
}

func (s *ValueImpl) Addr() model.Value {
	impl := &ValueImpl{value: &s.value}
	return impl
}

// Copy Copy
func (s *ValueImpl) copy() (ret *ValueImpl) {
	ret = &ValueImpl{value: s.value}

	return
}
