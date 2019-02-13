package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

type structImpl struct {
	value reflect.Value
}

func (s *structImpl) IsNil() bool {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	return false
}
func (s *structImpl) Set(val reflect.Value) (err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't set nil ptr")
		return
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Struct:
		if rawVal.Type().String() == val.Type().String() {
			rawVal.Set(val)
		} else {
			err = fmt.Errorf("can't convert %s to %s", val.Type().String(), rawVal.Type().String())
		}
	default:
		err = fmt.Errorf("can't convert %s to %s", val.Type().String(), rawVal.Type().String())
	}
	return
}

func (s *structImpl) Get() (reflect.Value, error) {
	return s.value, nil
}

func (s *structImpl) GetDepend() (ret []reflect.Value, err error) {
	if s.value.Kind() == reflect.Ptr {
		if s.value.IsNil() {
			return
		}
	}

	ret = append(ret, s.value)

	return
}

func (s *structImpl) GetValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value")
		return
	}

	rawVal := reflect.Indirect(s.value)
	pkField, pkErr := getStructPrimaryKey(rawVal)
	if pkErr != nil {
		err = pkErr
		return
	}

	ret, err = pkField.GetValue().GetValueStr()
	return
}

func (s *structImpl) Copy() model.FieldValue {
	return &structImpl{value: s.value}
}