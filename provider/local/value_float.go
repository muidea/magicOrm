package local

import (
	"fmt"
	"reflect"

	"muidea.com/magicOrm/model"
)

type floatImpl struct {
	value reflect.Value
}

func (s *floatImpl) IsNil() bool {
	if s.value.Kind() == reflect.Ptr {
		return s.value.IsNil()
	}

	return false
}
func (s *floatImpl) Set(val reflect.Value) (err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't set nil ptr")
		return
	}

	rawVal := reflect.Indirect(s.value)
	val = reflect.Indirect(val)
	switch val.Kind() {
	case reflect.Float32, reflect.Float64:
		rawVal.SetFloat(val.Float())
	default:
		err = fmt.Errorf("can't convert %s to float", val.Type().String())
	}
	return
}

func (s *floatImpl) Get() (reflect.Value, error) {
	return s.value, nil
}

func (s *floatImpl) Depend() (ret []reflect.Value, err error) {
	// noting todo
	return
}

func (s *floatImpl) ValueStr() (ret string, err error) {
	if s.IsNil() {
		err = fmt.Errorf("can't get nil ptr value")
		return
	}

	rawVal := reflect.Indirect(s.value)
	ret = fmt.Sprintf("%f", rawVal.Float())

	return
}

func (s *floatImpl) Copy() model.FieldValue {
	return &floatImpl{value: s.value}
}
