package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

type nilImpl struct {
	value      reflect.Value
	fieldValue model.FieldValue
	modelCache Cache
}

func (s *nilImpl) IsNil() bool {
	return s.fieldValue == nil
}

func (s *nilImpl) Set(val reflect.Value) (err error) {
	if val.Kind() != reflect.Ptr {
		err = fmt.Errorf("can't convert %s to %s", val.Type().String(), s.value.Type().String())
		return
	}

	if s.value.Type().String() == val.Type().String() {
		s.value.Set(val)
	}

	fieldValue, fieldErr := NewFieldValue(val, s.modelCache)
	if fieldErr != nil {
		err = fieldErr
		return
	}
	s.fieldValue = fieldValue

	return
}

func (s *nilImpl) Get() (ret reflect.Value, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil value")
		return
	}

	return s.value, nil
}

func (s *nilImpl) GetDepend() (ret []reflect.Value, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil depend")
		return
	}

	return s.fieldValue.GetDepend()
}

func (s *nilImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value string")
		return
	}

	return s.fieldValue.GetValueStr()
}

func (s *nilImpl) Copy() model.FieldValue {
	var fieldValue model.FieldValue
	if s.fieldValue != nil {
		fieldValue = s.fieldValue.Copy()
	}
	return &nilImpl{value: s.value, fieldValue: fieldValue, modelCache: s.modelCache}
}
