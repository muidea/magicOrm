package remote

import (
	"reflect"

	"github.com/muidea/magicOrm/model"
)

// ValueImpl ValueImpl
type ValueImpl struct {
	value interface{}
}

func GetValue(val reflect.Value) model.Value {
	return newValue(val)
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

// Update Update
func (s *ValueImpl) Update(val interface{}) (err error) {
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

func (s *ValueImpl) Type() (model.Type, error) {
	vType := reflect.TypeOf(s.value)
	if vType.Kind() == reflect.Interface {
		vType = vType.Elem()
	}

	return newType(vType)
}

// Copy Copy
func (s *ValueImpl) copy() (ret *ValueImpl) {
	ret = &ValueImpl{value: s.value}

	return
}
