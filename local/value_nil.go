package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

type nilImpl struct {
	value      reflect.Value
	fieldValue model.FieldValue
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

	fieldValue, fieldErr := NewFieldValue(val)
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

func (s *nilImpl) Depend() (ret []reflect.Value, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil depend")
		return
	}

	return s.fieldValue.Depend()
}

func (s *nilImpl) ValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value string")
		return
	}

	return s.fieldValue.ValueStr()
}

func (s *nilImpl) Copy() model.FieldValue {
	var fieldValue model.FieldValue
	if s.fieldValue != nil {
		fieldValue = s.fieldValue.Copy()
	}
	return &nilImpl{value: s.value, fieldValue: fieldValue}
}
