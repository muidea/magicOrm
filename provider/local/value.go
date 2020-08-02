package local

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicOrm/util"
)

type valueImpl struct {
	valueImpl reflect.Value
}

func newValue(val reflect.Value) (ret *valueImpl, err error) {
	ret = &valueImpl{valueImpl: val}
	return
}

func (s *valueImpl) IsNil() (ret bool) {
	ret = util.IsNil(s.valueImpl)
	return
}

func (s *valueImpl) Set(val reflect.Value) (err error) {
	s.valueImpl = val
	return
}

func (s *valueImpl) Update(val reflect.Value) (err error) {
	if util.IsNil(val) {
		s.valueImpl = val
		return
	}

	preType := s.valueImpl.Type().String()
	curType := val.Type().String()
	if preType != curType {
		err = fmt.Errorf("illegal update value type, value type:%s, expect type:%s", curType, preType)
		return
	}

	s.valueImpl.Set(val)
	return
}

func (s *valueImpl) Get() (ret reflect.Value) {
	ret = s.valueImpl

	return
}

func (s *valueImpl) Copy() (ret *valueImpl) {
	ret = &valueImpl{valueImpl: s.valueImpl}

	return
}
