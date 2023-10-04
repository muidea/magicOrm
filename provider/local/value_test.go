package local

import (
	"reflect"
	"testing"
)

func TestValue(t *testing.T) {
	var v reflect.Value

	// nil
	if !NilValue.IsNil() {
		t.Errorf("illegal nilValue, is nil")
		return
	}

	// zero
	if !NilValue.IsZero() {
		t.Errorf("illegal nilValue, is zero")
		return
	}

	valuePtr := NewValue(v)
	// nil
	if !valuePtr.IsNil() {
		t.Errorf("NewValue failed, is nil")
		return
	}

	// zero
	if !valuePtr.IsZero() {
		t.Errorf("NewValue failed, is zero")
		return
	}

	// not basic
	if valuePtr.IsBasic() {
		t.Errorf("IsBasic failed, is not basic")
		return
	}

	iVal := 10
	iReflect := reflect.ValueOf(&iVal)
	valuePtr = NewValue(iReflect)
	// not nil
	if valuePtr.IsNil() {
		t.Errorf("NewValue failed, is not nil")
	}

	// not zero
	if valuePtr.IsZero() {
		t.Errorf("NewValue failed, is not zero")
	}

	// basic
	if !valuePtr.IsBasic() {
		t.Errorf("IsBasic failed")
		return
	}

	var nulValue int
	nReflect := reflect.ValueOf(&nulValue)
	value2Ptr := NewValue(nReflect)
	// not nil
	if value2Ptr.IsNil() {
		t.Errorf("NewValue failed, IsNil false")
		return
	}

	// zero
	if !value2Ptr.IsZero() {
		t.Errorf("NewValue failed, IsZero true")
		return
	}

	// basic
	if !value2Ptr.IsBasic() {
		t.Errorf("IsBasic failed")
		return
	}

	value2Ptr.Set(valuePtr.Get())
	// not nil
	if value2Ptr.IsNil() {
		t.Errorf("is not nil")
		return
	}
	// not zero
	if value2Ptr.IsZero() {
		t.Errorf("not zero")
	}

	// basic
	if !value2Ptr.IsBasic() {
		t.Errorf("IsBasic failed")
		return
	}

	iVal = 12
	iReflect2 := reflect.ValueOf(&iVal)
	value2Ptr.Set(iReflect2)
}
