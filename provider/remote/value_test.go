package remote

import (
	"testing"
)

func TestValue(t *testing.T) {
	var v any

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
	valuePtr = NewValue(iVal)
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
	value2Ptr := NewValue(nulValue)
	// not nil
	if value2Ptr.IsNil() {
		t.Errorf("NewValue failed, IsNil false")
		return
	}

	// not zero
	if !value2Ptr.IsZero() {
		t.Errorf("NewValue failed, IsZero true")
		return
	}

	// basic
	if !value2Ptr.IsBasic() {
		t.Errorf("IsBasic failed")
		return
	}

	valueErr := value2Ptr.Set(valuePtr.Get())
	if valueErr != nil {
		t.Errorf("set failed, err:%s", valueErr.Error())
		return
	}
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

	iReflect2 := 12
	value2Err := value2Ptr.Set(iReflect2)
	if value2Err != nil {
		t.Errorf("update failed, err:%s", value2Err.Error())
		return
	}
}
