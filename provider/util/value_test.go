package util

import (
	"reflect"
	"testing"
)

func TestValue(t *testing.T) {
	var v reflect.Value

	valuePtr := NewValue(v)
	if !valuePtr.IsNil() {
		t.Errorf("NewValue failed")
		return
	}

	if valuePtr.IsBasic() {
		t.Errorf("IsBasic failed")
		return
	}

	iVal := 10
	iReflect := reflect.ValueOf(&iVal).Elem()
	valuePtr = NewValue(iReflect)
	if valuePtr.IsNil() {
		t.Errorf("NewValue failed")
	}

	if !valuePtr.IsBasic() {
		t.Errorf("IsBasic failed")
		return
	}

	var nulValue interface{} = nil
	nReflect := reflect.ValueOf(nulValue)
	value2Ptr := NewValue(nReflect)
	if !value2Ptr.IsNil() {
		t.Errorf("NewValue failed")
		return
	}
	if value2Ptr.IsBasic() {
		t.Errorf("IsBasic failed")
		return
	}

	valueErr := value2Ptr.Set(valuePtr.Get())
	if valueErr != nil {
		t.Errorf("set failed, err:%s", valueErr.Error())
		return
	}
	if value2Ptr.IsNil() {
		t.Errorf("NewValue failed")
		return
	}
	if !value2Ptr.IsBasic() {
		t.Errorf("IsBasic failed")
		return
	}

	iReflect2 := reflect.ValueOf(12)
	value2Err := value2Ptr.Set(iReflect2)
	if value2Err != nil {
		t.Errorf("update failed, err:%s", value2Err.Error())
		return
	}
}
