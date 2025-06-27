package remote

import (
	"reflect"
	"testing"
)

func TestValue(t *testing.T) {
	var v any
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

	iVal := 10
	valuePtr = NewValue(iVal)
	// not nil
	if !valuePtr.IsValid() {
		t.Errorf("NewValue failed, is not nil")
	}

	// not zero
	if valuePtr.IsZero() {
		t.Errorf("NewValue failed, is not zero")
	}

	var nulValue int
	value2Ptr := NewValue(nulValue)
	// not nil
	if !value2Ptr.IsValid() {
		t.Errorf("NewValue failed, IsValid false")
		return
	}

	// not zero
	if !value2Ptr.IsZero() {
		t.Errorf("NewValue failed, IsZero true")
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

	iReflect2 := 12
	value2Ptr.Set(iReflect2)

	// Test Interface method
	if valuePtr.Get() != iVal {
		t.Errorf("Interface failed, value mismatch: got %v, expected %v", valuePtr.Get(), iVal)
		return
	}
}

func TestValueWithVariousTypes(t *testing.T) {
	// Test boolean values
	boolVal := true
	boolValuePtr := NewValue(boolVal)
	if !boolValuePtr.IsValid() || boolValuePtr.IsZero() {
		t.Errorf("Boolean value handling failed")
		return
	}

	// Test zero boolean
	zeroBool := false
	zeroBoolValuePtr := NewValue(zeroBool)
	if !zeroBoolValuePtr.IsValid() || !zeroBoolValuePtr.IsZero() {
		t.Errorf("Zero boolean value handling failed")
		return
	}

	// Test string values
	strVal := "test string"
	strValuePtr := NewValue(strVal)
	if !strValuePtr.IsValid() || strValuePtr.IsZero() {
		t.Errorf("String value handling failed")
		return
	}

	// Test empty string
	emptyStr := ""
	emptyStrValuePtr := NewValue(emptyStr)
	if !emptyStrValuePtr.IsValid() || !emptyStrValuePtr.IsZero() {
		t.Errorf("Empty string value handling failed")
		return
	}

	// Test float values
	floatVal := 123.456
	floatValuePtr := NewValue(floatVal)
	if !floatValuePtr.IsValid() || floatValuePtr.IsZero() {
		t.Errorf("Float value handling failed")
		return
	}

	// Test zero float
	zeroFloat := 0.0
	zeroFloatValuePtr := NewValue(zeroFloat)
	if !zeroFloatValuePtr.IsValid() || !zeroFloatValuePtr.IsZero() {
		t.Errorf("Zero float value handling failed")
		return
	}

	// Test slice values
	sliceVal := []int{1, 2, 3}
	sliceValuePtr := NewValue(sliceVal)
	if !sliceValuePtr.IsValid() || sliceValuePtr.IsZero() {
		t.Errorf("Slice value handling failed")
		return
	}

	// Test empty slice
	emptySlice := []int{}
	emptySliceValuePtr := NewValue(emptySlice)
	if !emptySliceValuePtr.IsValid() || !emptySliceValuePtr.IsZero() {
		t.Errorf("Empty slice value handling failed")
		return
	}
}

func TestValueCopy(t *testing.T) {
	// Test copy of basic types
	intVal := 42
	intValuePtr := NewValue(intVal)

	copyValue, err := intValuePtr.copy()
	if err != nil {
		t.Errorf("Copy of int value failed: %v", err)
		return
	}

	if !reflect.DeepEqual(copyValue.Get(), intValuePtr.Get()) {
		t.Errorf("Copy of int value failed, values don't match: got %v, expected %v",
			copyValue.Get(), intValuePtr.Get())
		return
	}

	// Test copy of string
	strVal := "test string"
	strValuePtr := NewValue(strVal)

	copyStrValue, err := strValuePtr.copy()
	if err != nil {
		t.Errorf("Copy of string value failed: %v", err)
		return
	}

	if !reflect.DeepEqual(copyStrValue.Get(), strValuePtr.Get()) {
		t.Errorf("Copy of string value failed, values don't match: got %v, expected %v",
			copyStrValue.Get(), strValuePtr.Get())
		return
	}

	// Test copy of slice
	sliceVal := []int{1, 2, 3}
	sliceValuePtr := NewValue(sliceVal)

	copySliceValue, err := sliceValuePtr.copy()
	if err != nil {
		t.Errorf("Copy of slice value failed: %v", err)
		return
	}

	if !reflect.DeepEqual(copySliceValue.Get(), sliceValuePtr.Get()) {
		t.Errorf("Copy of slice value failed, values don't match: got %v, expected %v",
			copySliceValue.Get(), sliceValuePtr.Get())
		return
	}
}

func TestObjectValueHandling(t *testing.T) {
	// Create a simple ObjectValue
	objVal := &ObjectValue{
		Name:    "TestObject",
		PkgPath: "test/pkg",
		Fields: []*FieldValue{
			{Name: "field1", Value: 42},
		},
	}

	objValuePtr := NewValue(objVal)

	if !objValuePtr.IsValid() {
		t.Errorf("ObjectValue should be valid")
		return
	}

	if objValuePtr.IsZero() {
		t.Errorf("ObjectValue should not be zero")
		return
	}

	// Test copy of ObjectValue
	copyObjValue, err := objValuePtr.copy()
	if err != nil {
		t.Errorf("Copy of ObjectValue failed: %v", err)
		return
	}

	originalPtr := objValuePtr.Get().(*ObjectValue)
	copyPtr := copyObjValue.Get().(*ObjectValue)

	if originalPtr.Name != copyPtr.Name || originalPtr.PkgPath != copyPtr.PkgPath {
		t.Errorf("ObjectValue copy failed, values don't match")
		return
	}
}

func TestSliceObjectValueHandling(t *testing.T) {
	// Create a simple SliceObjectValue
	sliceObjVal := &SliceObjectValue{
		Name:    "TestSliceObject",
		PkgPath: "test/pkg",
		Values: []*ObjectValue{
			{
				Name:    "TestObject",
				PkgPath: "test/pkg",
				Fields: []*FieldValue{
					{Name: "field1", Value: 42},
				},
			},
		},
	}

	sliceObjValuePtr := NewValue(sliceObjVal)

	if !sliceObjValuePtr.IsValid() {
		t.Errorf("SliceObjectValue should be valid")
		return
	}

	if sliceObjValuePtr.IsZero() {
		t.Errorf("SliceObjectValue should not be zero with values")
		return
	}

	// Test copy of SliceObjectValue
	copySliceObjValue, err := sliceObjValuePtr.copy()
	if err != nil {
		t.Errorf("Copy of SliceObjectValue failed: %v", err)
		return
	}

	originalPtr := sliceObjValuePtr.Get().(*SliceObjectValue)
	copyPtr := copySliceObjValue.Get().(*SliceObjectValue)

	if originalPtr.Name != copyPtr.Name || originalPtr.PkgPath != copyPtr.PkgPath {
		t.Errorf("SliceObjectValue copy failed, values don't match")
		return
	}

	// Test empty SliceObjectValue
	emptySliceObjVal := &SliceObjectValue{
		Name:    "TestEmptySliceObject",
		PkgPath: "test/pkg",
		Values:  []*ObjectValue{},
	}

	emptySliceObjValuePtr := NewValue(emptySliceObjVal)

	if !emptySliceObjValuePtr.IsValid() {
		t.Errorf("Empty SliceObjectValue should be valid")
		return
	}

	if !emptySliceObjValuePtr.IsZero() {
		t.Errorf("Empty SliceObjectValue should be zero")
		return
	}
}
