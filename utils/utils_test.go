package utils

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/log"
)

func TestInterface(t *testing.T) {
	interfaceSlice := []any{}
	interfaceSlice = append(interfaceSlice, "test")
	interfaceSlice = append(interfaceSlice, 1)
	interfaceSlice = append(interfaceSlice, 1.1)
	interfaceSlice = append(interfaceSlice, true)
	interfaceSlice = append(interfaceSlice, time.Now())

	rVal := reflect.ValueOf(interfaceSlice)
	for idx := range rVal.Len() {
		t.Logf("idx:%d, type:%v, value:%v", idx, rVal.Index(idx).Type(), rVal.Index(idx).Interface())

		rawVal := reflect.ValueOf(rVal.Index(idx).Interface())
		t.Logf("raw type:%v, value:%v", rawVal.Type(), rawVal.Interface())
	}
}

// TestIsReallyValid tests the IsReallyValid function based on its documented behavior
func TestIsReallyValid(t *testing.T) {
	var strPtr *string
	strPtrRVal := reflect.ValueOf(strPtr)
	log.Infof("strPtrRVal Type:%v", strPtrRVal.Type())
	strPtrRVal = reflect.New(strPtrRVal.Type().Elem()).Elem()
	log.Infof("strPtrRVal Type:%v", strPtrRVal.Type())
	if !IsReallyValidValueForReflect(strPtrRVal) {
		t.Errorf("strPtrRVal should be valid")
	}
	strPtrRVal = strPtrRVal.Addr()
	log.Infof("strPtrRVal Type:%v", strPtrRVal.Type())
	if !IsReallyValidValueForReflect(strPtrRVal) {
		t.Errorf("strPtrRVal should be valid")
	}

	var interfaceVal any

	emptyIntSlice := []int{}
	noEmptyIntSlice := []int{1, 2, 3}

	emptyArray := [0]int{}
	float32Array := [2]float32{12.34, 23.45}

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		// Basic value types tests
		{"nil", nil, false},
		{"bool true", true, true},
		{"bool false", false, true},
		{"int zero", 0, true},
		{"int non-zero", 42, true},
		{"int8", int8(8), true},
		{"int16", int16(16), true},
		{"int32", int32(32), true},
		{"int64", int64(64), true},
		{"uint", uint(1), true},
		{"uint8", uint8(8), true},
		{"uint16", uint16(16), true},
		{"uint32", uint32(32), true},
		{"uint64", uint64(64), true},
		{"float32", float32(3.14), true},
		{"float64", float64(3.14), true},
		{"empty string", "", true},
		{"non-empty string", "hello", true},

		// Pointer type tests
		{"nil pointer", (*int)(nil), false},
		{"pointer to int", func() any { v := 42; return &v }(), true},
		{"pointer to string", func() any { v := "hello"; return &v }(), true},
		{"nil pointer time.Time", (*time.Time)(nil), false},
		{"pointer to time.Time", func() any { v := time.Now(); return &v }(), true},

		{"interface, empty int slice", func() any { interfaceVal = emptyIntSlice; return interfaceVal }(), true},
		{"interface, noEmpty int slice", func() any { interfaceVal = noEmptyIntSlice; return interfaceVal }(), true},
		{"interface, empty int array", func() any { interfaceVal = emptyArray; return interfaceVal }(), true},
		{"interface, noEmpty float32 array", func() any { interfaceVal = float32Array; return interfaceVal }(), true},

		// Slice type tests
		{"nil slice", ([]int)(nil), false},
		{"empty slice", []int{}, true},
		{"int slice", []int{1, 2, 3}, true},
		{"string slice", []string{"a", "b", "c"}, true},
		{"empty float32 array", [0]float32{}, true},

		// Invalid types
		{"map", map[string]int{"a": 1}, false},
		{"struct", struct{ Name string }{"John"}, false},
		{"chan", make(chan int), false},
		{"func", func() {}, false},
		{"complex", complex(1, 2), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReallyValidValue(tt.value); got != tt.want {
				t.Errorf("IsReallyValid() = %v, name:%v, want %v for %v", got, tt.name, tt.want, tt.value)
			}
		})
	}
}

// TestIsReallyZero tests the IsReallyZero function based on its documented behavior
func TestIsReallyZero(t *testing.T) {
	var interfaceVal any

	emptyIntSlice := []int{}
	noEmptyIntSlice := []int{1, 2, 3}

	interfaceVal = emptyIntSlice
	if !IsReallyZeroValue(interfaceVal) {
		t.Error("IsReallyZero() should return true for empty slice")
	}
	interfaceVal = noEmptyIntSlice
	if IsReallyZeroValue(interfaceVal) {
		t.Error("IsReallyZero() should return false for non-empty slice")
	}

	tests := []struct {
		name  string
		value any
		want  bool
	}{
		// Basic value types tests
		{"nil", nil, true},
		{"bool false", false, true},
		{"bool true", true, false},
		{"int zero", 0, true},
		{"int non-zero", 42, false},
		{"int8 zero", int8(0), true},
		{"int8 non-zero", int8(8), false},
		{"int16 zero", int16(0), true},
		{"int16 non-zero", int16(16), false},
		{"int32 zero", int32(0), true},
		{"int32 non-zero", int32(32), false},
		{"int64 zero", int64(0), true},
		{"int64 non-zero", int64(64), false},
		{"uint zero", uint(0), true},
		{"uint non-zero", uint(1), false},
		{"uint8 zero", uint8(0), true},
		{"uint8 non-zero", uint8(8), false},
		{"uint16 zero", uint16(0), true},
		{"uint16 non-zero", uint16(16), false},
		{"uint32 zero", uint32(0), true},
		{"uint32 non-zero", uint32(32), false},
		{"uint64 zero", uint64(0), true},
		{"uint64 non-zero", uint64(64), false},
		{"float32 zero", float32(0), true},
		{"float32 non-zero", float32(3.14), false},
		{"float64 zero", float64(0), true},
		{"float64 non-zero", float64(3.14), false},
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"empty time.Time", time.Time{}, true},
		{"non-empty time.Time", time.Now(), false},

		// Pointer type tests
		{"nil pointer", (*int)(nil), true},
		{"pointer to zero", func() any { v := 0; return &v }(), true},
		{"pointer to non-zero", func() any { v := 42; return &v }(), false},
		{"nil pointer time.Time", (*time.Time)(nil), true},
		{"pointer to non-empty time.Time", func() any { v := time.Now(); return &v }(), false},

		{"empty slice", func() any { interfaceVal = emptyIntSlice; return interfaceVal }(), true},
		{"noEmpty slice", func() any { interfaceVal = noEmptyIntSlice; return interfaceVal }(), false},

		// Invalid types (should return default Go zero value determination)
		{"map", map[string]int{"a": 1}, false},
		{"empty map", map[string]int{}, true},
		{"struct with values", struct{ Name string }{"John"}, false},
		{"empty struct", struct{ Name string }{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReallyZeroValue(tt.value); got != tt.want {
				t.Errorf("IsReallyZero() name:%s got = %v, want %v for %v", tt.name, got, tt.want, tt.value)
			}
		})
	}
}

// TestIsReallyValidType tests the IsReallyValidType function based on its documented behavior
func TestIsReallyValidType(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		// Basic value types tests
		{"nil", nil, false},
		{"bool", true, true},
		{"int", 42, true},
		{"int8", int8(8), true},
		{"int16", int16(16), true},
		{"int32", int32(32), true},
		{"int64", int64(64), true},
		{"uint", uint(1), true},
		{"uint8", uint8(8), true},
		{"uint16", uint16(16), true},
		{"uint32", uint32(32), true},
		{"uint64", uint64(64), true},
		{"float32", float32(3.14), true},
		{"float64", float64(3.14), true},
		{"string", "hello", true},

		// Pointer type tests
		{"nil pointer", (*int)(nil), true},
		{"pointer to int", func() any { v := 42; return &v }(), true},
		{"pointer to string", func() any { v := "hello"; return &v }(), true},

		// Slice type tests
		{"nil slice", ([]int)(nil), true},
		{"empty slice", []int{}, true},
		{"int slice", []int{1, 2, 3}, true},
		{"string slice", []string{"a", "b", "c"}, true},

		// Invalid types
		{"map", map[string]int{"a": 1}, false},
		{"struct", struct{ Name string }{"John"}, false},
		{"chan", make(chan int), false},
		{"func", func() {}, false},
		{"complex", complex(1, 2), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReallyValidType(tt.value); got != tt.want {
				t.Errorf("IsReallyValidType() = %v, want %v for %v of type %T", got, tt.want, tt.name, tt.value)
			}
		})
	}

	sliceArray := []any{}
	sliceArray = append(sliceArray, 1, 2, 3)
	if IsReallyValidType(sliceArray) {
		t.Errorf("IsReallyValidType([]any) = true, want false")
	}
}

// TestNestedTypes tests the behavior of the utility functions with nested types
func TestNestedTypes(t *testing.T) {
	// Nested slices
	nestedSlice := [][]int{{1, 2, 3}}
	if IsReallyValidType(nestedSlice) {
		t.Errorf("IsReallyValidType([][]int) = true, want false")
	}

	// Pointers to slices
	slicePtr := &[]int{1, 2, 3}
	if !IsReallyValidType(slicePtr) {
		t.Errorf("IsReallyValidType(&[]int) = true, want false")
	}

	// Slices of pointers to basic types
	intVal := 42
	sliceOfPtrs := []*int{&intVal}
	if !IsReallyValidType(sliceOfPtrs) {
		t.Errorf("IsReallyValidType([]*int) = false, want true")
	}
}

// TestBasicTypesFull provides an exhaustive test of all basic types
func TestBasicTypesFull(t *testing.T) {
	// Test all basic value types are considered valid
	validTypes := []any{
		// Bool
		true, false,
		// Integer types
		int8(8), int16(16), int32(32), int(42), int64(64),
		uint8(8), uint16(16), uint32(32), uint(42), uint64(64),
		// Float types
		float32(3.14), float64(3.14),
		// String
		"hello",
		time.Now(),
	}

	for i, v := range validTypes {
		if !IsReallyValidType(v) {
			t.Errorf("Case %d: IsReallyValidType(%T) = false, want true", i, v)
		}
	}

	// Test all non-basic types are considered invalid
	invalidTypes := []any{
		struct{}{},
		map[string]int{},
		[]struct{}{},
		func() {},
		make(chan int),
		complex(1, 2),
	}

	for i, v := range invalidTypes {
		if IsReallyValidType(v) {
			t.Errorf("Case %d: IsReallyValidType(%T) = true, want false", i, v)
		}
	}
}

// TestPointersDeep tests the behavior with multi-level pointers
func TestPointersDeep(t *testing.T) {
	// Single level pointer
	intVal := 42
	ptrInt := &intVal

	if !IsReallyValidType(ptrInt) {
		t.Errorf("IsReallyValidType(*int) = false, want true")
	}

	// Double level pointer
	ptrPtrInt := &ptrInt
	if IsReallyValidType(ptrPtrInt) {
		t.Errorf("IsReallyValidType(**int) = true, want false")
	}
}

// TestEdgeCases tests edge cases for each function
func TestEdgeCases(t *testing.T) {
	// IsReallyValidType with any
	var iface any = 42
	if !IsReallyValidType(iface) {
		t.Errorf("IsReallyValidType(any=42) = false, want true")
	}

	// IsReallyValid with any
	if !IsReallyValidValue(iface) {
		t.Errorf("IsReallyValid(any=42) = false, want true")
	}

	// IsReallyZero with any
	iface = 0
	if !IsReallyZeroValue(iface) {
		t.Errorf("IsReallyZero(any=0) = false, want true")
	}

	// Empty slices of various types
	if !IsReallyValidType([]int{}) {
		t.Errorf("IsReallyValidType([]int{}) = false, want true")
	}

	if !IsReallyValidType([]string{}) {
		t.Errorf("IsReallyValidType([]string{}) = false, want true")
	}
}

func TestDeepCopy(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{"int", 42},
		{"string", "hello"},
		{"bool", true},
		{"float64", 3.14},
		{"slice", []int{1, 2, 3}},
		{"pointer", func() any { v := 10; return &v }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copiedVal, copiedErr := DeepCopy(tt.value)
			if copiedErr != nil {
				t.Errorf("DeepCopy() error = %v, want nil", copiedErr)
			}

			if !reflect.DeepEqual(tt.value, copiedVal) {
				t.Errorf("DeepCopy() = %v, want %v", copiedVal, tt.value)
			}

			// Check that the copied value is not the same instance
			if reflect.ValueOf(tt.value).Kind() == reflect.Ptr {
				if reflect.ValueOf(tt.value).Pointer() == reflect.ValueOf(copiedVal).Pointer() {
					t.Errorf("DeepCopy() returned same pointer for %v", tt.name)
				}
			}
		})
	}
}

func TestDeepCopyForReflect(t *testing.T) {
	type MyStruct struct {
		A int
		B string
		C []int
	}

	original := MyStruct{
		A: 42,
		B: "hello",
		C: []int{1, 2, 3},
	}

	srcValue := reflect.ValueOf(original)
	copyValue := DeepCopyForReflect(srcValue)

	// 验证拷贝结果
	copied := copyValue.Interface().(MyStruct)
	if !IsSameValue(original, copied) {
		t.Errorf("DeepCopyForReflect() = %v, want %v", copied, original)
	}
}
