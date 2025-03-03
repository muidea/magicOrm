package local

import (
	"reflect"
	"testing"
)

func TestValue(t *testing.T) {
	var v reflect.Value

	// nil
	if NilValue.IsValid() {
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
	if valuePtr.IsValid() {
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
	if !valuePtr.IsValid() {
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
	if !value2Ptr.IsValid() {
		t.Errorf("NewValue failed, IsValid false")
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
	if !value2Ptr.IsValid() {
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

// TestValueInterface tests the Interface method
func TestValueInterface(t *testing.T) {
	// Test with nil value
	var v reflect.Value
	valuePtr := NewValue(v)
	if valuePtr.Interface() != nil {
		t.Errorf("Interface() for nil value should return nil")
	}

	// Test with non-nil value
	iVal := 10
	iReflect := reflect.ValueOf(iVal)
	valuePtr = NewValue(iReflect)
	rawVal := valuePtr.Interface()
	if rawVal == nil {
		t.Errorf("Interface() for non-nil value should not return nil")
		return
	}

	if rawVal.Value() != iVal {
		t.Errorf("Interface() returned wrong value, expected: %v, got: %v", iVal, rawVal.Value())
	}
}

// TestValueAddr tests the Addr method
func TestValueAddr(t *testing.T) {
	// Setup addressable value
	iVal := 10
	slice := []int{iVal}
	sliceReflectVal := reflect.ValueOf(slice)
	elemReflectVal := sliceReflectVal.Index(0)
	valuePtr := NewValue(elemReflectVal)

	// Test Addr
	addrValue := valuePtr.Addr()
	if addrValue == nil {
		t.Errorf("Addr() should not return nil for addressable value")
		return
	}

	// Verify the value can be dereferenced
	rawVal := addrValue.Interface()
	if rawVal == nil {
		t.Errorf("Addr().Interface() should not return nil")
		return
	}

	// Verify the value is correct
	val := rawVal.Value().(*int)
	if *val != iVal {
		t.Errorf("Addr() returned wrong value, expected: %v, got: %v", iVal, *val)
	}
}

// TestValueCopy tests the Copy method
func TestValueCopy(t *testing.T) {
	// Test with nil value
	var v reflect.Value
	valuePtr := NewValue(v)
	copyVal := valuePtr.Copy()
	if copyVal.IsValid() {
		t.Errorf("Copy() of nil value should not be valid")
	}

	// Test with non-nil value
	iVal := 10
	iReflect := reflect.ValueOf(iVal)
	valuePtr = NewValue(iReflect)
	copyVal = valuePtr.Copy()

	// Verify copy is valid
	if !copyVal.IsValid() {
		t.Errorf("Copy() of non-nil value should be valid")
		return
	}

	// Verify value equality
	originalVal := valuePtr.Get().(reflect.Value).Interface()
	copiedVal := copyVal.Get().(reflect.Value).Interface()
	if originalVal != copiedVal {
		t.Errorf("Copy() value mismatch, expected: %v, got: %v", originalVal, copiedVal)
	}

	// Verify deep copy - changing original should not affect copy
	iVal = 20
	if copiedVal != 10 {
		t.Errorf("Copy() should be a deep copy, expected copied value to remain 10, got: %v", copiedVal)
	}
}

// TestValueIsBasic tests the IsBasic method with various types
func TestValueIsBasic(t *testing.T) {
	// Test with basic types
	testCases := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"Int", 10, true},
		{"String", "test", true},
		{"Bool", true, true},
		{"Float", 10.5, true},
		{"Struct", struct{ Name string }{"test"}, false},
		{"Slice of basic", []int{1, 2, 3}, true},
		{"Slice of struct", []struct{ Name string }{{"test"}}, false},
		{"Pointer to basic", func() interface{} { i := 10; return &i }(), true},
		{"Pointer to struct", func() interface{} { s := struct{ Name string }{"test"}; return &s }(), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valuePtr := NewValue(reflect.ValueOf(tc.value))
			if valuePtr.IsBasic() != tc.expected {
				t.Errorf("IsBasic() for %s, expected: %v, got: %v", tc.name, tc.expected, valuePtr.IsBasic())
			}
		})
	}
}

// TestValueWithNilPointer tests value methods with nil pointers
func TestValueWithNilPointer(t *testing.T) {
	// Create nil pointer of type *int
	var ptr *int
	valuePtr := NewValue(reflect.ValueOf(ptr))

	// Test IsValid
	if valuePtr.IsValid() {
		t.Errorf("IsValid() for nil pointer should be false (value is invalid if nil)")
	}

	// Test IsZero
	if !valuePtr.IsZero() {
		t.Errorf("IsZero() for nil pointer should be true")
	}

	iVal := 10
	ptr = &iVal
	valuePtr = NewValue(reflect.ValueOf(ptr))
	if !valuePtr.IsValid() {
		t.Errorf("IsValid() for non-nil pointer should be true")
	}
	if valuePtr.IsZero() {
		t.Errorf("IsZero() for non-nil pointer should be false")
	}
}
