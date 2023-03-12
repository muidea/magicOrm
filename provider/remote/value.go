package remote

import (
	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/util"
	"reflect"
)

// ValueImpl ValueImpl
type ValueImpl struct {
	value reflect.Value
}

func newValue(val reflect.Value) (ret *ValueImpl) {
	ret = &ValueImpl{value: reflect.Indirect(val)}
	return
}

// IsNil IsNil
func (s *ValueImpl) IsNil() (ret bool) {
	ret = util.IsNil(s.value)

	return
}

// Set Set
func (s *ValueImpl) Set(val reflect.Value) (err error) {
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = val.Elem()
		val = reflect.Indirect(val)
	}

	if util.IsNil(s.value) || util.IsNil(val) {
		s.value = val
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

func (s *ValueImpl) Addr() model.Value {
	impl := &ValueImpl{value: s.value.Addr()}
	return impl
}

func (s *ValueImpl) IsBasic() bool {
	if util.IsNil(s.value) {
		return false
	}

	trueType := s.value.Type()
	if s.value.Kind() == reflect.Interface {
		trueType = s.value.Elem().Type()
	}
	return !util.IsStruct(trueType)
}

// Copy Copy
func (s *ValueImpl) copy() (ret *ValueImpl) {
	if !util.IsNil(s.value) {
		ret = &ValueImpl{value: reflect.New(s.value.Type()).Elem()}
		ret.value.Set(s.value)
		return
	}

	ret = &ValueImpl{}
	return
}
