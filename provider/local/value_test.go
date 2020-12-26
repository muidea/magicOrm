package local

import (
	"reflect"
	"testing"
)

func TestValue(t *testing.T) {
	var v reflect.Value

	valuePtr, valueErr := newValue(v)
	if valueErr != nil {
		t.Errorf("newValue failed, err:%s", valueErr.Error())
		return
	}
	if !valuePtr.IsNil() {
		t.Errorf("newValue failed")
		return
	}

	iVal := 10
	iReflect := reflect.ValueOf(&iVal).Elem()
	valuePtr, valueErr = newValue(iReflect)
	if valueErr != nil {
		t.Errorf("newValue failed, err:%s", valueErr.Error())
		return
	}
	if valuePtr.IsNil() {
		t.Errorf("newValue failed")
	}

	var nulValue interface{} = nil
	nReflect := reflect.ValueOf(nulValue)
	value2Ptr, value2Err := newValue(nReflect)
	if value2Err != nil {
		t.Errorf("newValue failed, err:%s", value2Err.Error())
		return
	}
	if !value2Ptr.IsNil() {
		t.Errorf("newValue failed")
		return
	}
	valueErr = value2Ptr.Set(iReflect)
	if value2Err != nil {
		t.Errorf("set failed, err:%s", value2Err.Error())
		return
	}
	if value2Ptr.IsNil() {
		t.Errorf("newValue failed")
		return
	}

	iReflect2 := reflect.ValueOf(12)
	value2Err = value2Ptr.Update(iReflect2)
	if value2Err != nil {
		t.Errorf("update failed, err:%s", value2Err.Error())
		return
	}
}
