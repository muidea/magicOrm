package consistency

import (
	"math"
	"testing"
	"time"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestNilValueHandling(t *testing.T) {
	t.Run("nil pointer in struct", func(t *testing.T) {
		entity := &PointerTypes{
			ID:   1,
			Bool: nil,
			Int:  nil,
			Str:  nil,
			Time: nil,
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &PointerTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.Bool != nil || target.Int != nil || target.Str != nil || target.Time != nil {
			t.Error("nil pointers should remain nil after round trip")
		}
	})

	t.Run("nil nested struct", func(t *testing.T) {
		entity := &NestedParent{
			ID:    1,
			Name:  "test",
			Child: nil,
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &NestedParent{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.Child != nil {
			t.Error("nil child should remain nil")
		}
	})
}

func TestZeroValueHandling(t *testing.T) {
	t.Run("zero values in basic types", func(t *testing.T) {
		entity := &BasicTypes{
			ID:     0,
			Bool:   false,
			Int8:   0,
			Int16:  0,
			Int32:  0,
			Int64:  0,
			Int:    0,
			UInt8:  0,
			UInt16: 0,
			UInt32: 0,
			UInt64: 0,
			UInt:   0,
			Float:  0,
			Double: 0,
			Str:    "",
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &BasicTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.Bool != false {
			t.Error("zero bool should be false")
		}
		if target.Int != 0 {
			t.Error("zero int should be 0")
		}
		if target.Str != "" {
			t.Error("zero string should be empty")
		}
	})
}

func TestEmptySliceEdgeCases(t *testing.T) {
	t.Run("empty slices", func(t *testing.T) {
		entity := &SliceTypes{
			ID:    1,
			Bools: []bool{},
			Ints:  []int{},
			Strs:  []string{},
			Times: []time.Time{},
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &SliceTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.Bools != nil && len(target.Bools) != 0 {
			t.Errorf("empty bools slice handling incorrect")
		}
		if target.Ints != nil && len(target.Ints) != 0 {
			t.Errorf("empty ints slice handling incorrect")
		}
	})

	t.Run("nil slices", func(t *testing.T) {
		entity := &SliceTypes{
			ID:    1,
			Bools: nil,
			Ints:  nil,
			Strs:  nil,
			Times: nil,
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &SliceTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}
	})
}

func TestLargeValues(t *testing.T) {
	t.Run("max int values", func(t *testing.T) {
		entity := &BasicTypes{
			ID:     math.MaxInt,
			Int8:   math.MaxInt8,
			Int16:  math.MaxInt16,
			Int32:  math.MaxInt32,
			Int64:  math.MaxInt64,
			UInt8:  math.MaxUint8,
			UInt16: math.MaxUint16,
			UInt32: math.MaxUint32,
			UInt64: math.MaxUint64,
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &BasicTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.Int8 != math.MaxInt8 {
			t.Errorf("Int8 max value mismatch: expected %d, got %d", math.MaxInt8, target.Int8)
		}
		if target.Int16 != math.MaxInt16 {
			t.Errorf("Int16 max value mismatch: expected %d, got %d", math.MaxInt16, target.Int16)
		}
		if target.Int32 != math.MaxInt32 {
			t.Errorf("Int32 max value mismatch: expected %d, got %d", math.MaxInt32, target.Int32)
		}
		if target.Int64 != math.MaxInt64 {
			t.Errorf("Int64 max value mismatch: expected %d, got %d", math.MaxInt64, target.Int64)
		}
	})

	t.Run("min int values", func(t *testing.T) {
		entity := &BasicTypes{
			Int8:  math.MinInt8,
			Int16: math.MinInt16,
			Int32: math.MinInt32,
			Int64: math.MinInt64,
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &BasicTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.Int8 != math.MinInt8 {
			t.Errorf("Int8 min value mismatch: expected %d, got %d", math.MinInt8, target.Int8)
		}
		if target.Int16 != math.MinInt16 {
			t.Errorf("Int16 min value mismatch: expected %d, got %d", math.MinInt16, target.Int16)
		}
		if target.Int32 != math.MinInt32 {
			t.Errorf("Int32 min value mismatch: expected %d, got %d", math.MinInt32, target.Int32)
		}
		if target.Int64 != math.MinInt64 {
			t.Errorf("Int64 min value mismatch: expected %d, got %d", math.MinInt64, target.Int64)
		}
	})
}

func TestSpecialStrings(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"empty string", ""},
		{"unicode", "你好世界"},
		{"emoji", "😀🎉🚀"},
		{"newlines", "line1\nline2\nline3"},
		{"tabs", "col1\tcol2\tcol3"},
		{"quotes", `"quoted" and 'single'`},
		{"backslash", `path\to\file`},
		{"json special", `{"key": "value"}`},
		{"html", `<html><body>test</body></html>`},
		{"long string", string(make([]byte, 1000))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := &BasicTypes{
				ID:  1,
				Str: tt.value,
			}

			objValue, err := helper.GetObjectValue(entity)
			if err != nil {
				t.Fatalf("GetObjectValue failed: %v", err)
			}

			jsonData, err := remote.EncodeObjectValue(objValue)
			if err != nil {
				t.Fatalf("EncodeObjectValue failed: %v", err)
			}

			decodedValue, err := remote.DecodeObjectValue(jsonData)
			if err != nil {
				t.Fatalf("DecodeObjectValue failed: %v", err)
			}

			target := &BasicTypes{}
			err = helper.UpdateEntity(decodedValue, target)
			if err != nil {
				t.Fatalf("UpdateEntity failed: %v", err)
			}

			if target.Str != tt.value {
				t.Errorf("String mismatch: expected %q, got %q", tt.value, target.Str)
			}
		})
	}
}

func TestFloatSpecialValues(t *testing.T) {
	t.Run("very small float", func(t *testing.T) {
		entity := &BasicTypes{
			ID:     1,
			Float:  1e-38,
			Double: 1e-308,
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &BasicTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.Float == 0 {
			t.Error("very small float should not become zero")
		}
	})

	t.Run("very large float", func(t *testing.T) {
		entity := &BasicTypes{
			ID:     1,
			Float:  math.MaxFloat32,
			Double: math.MaxFloat64,
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &BasicTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.Float != math.MaxFloat32 {
			t.Errorf("Float32 max value mismatch: expected %v, got %v", math.MaxFloat32, target.Float)
		}
	})
}

func TestSlicePointerFields(t *testing.T) {
	t.Run("slice pointer fields", func(t *testing.T) {
		boolSlice := []bool{true, false}
		intSlice := []int{1, 2, 3}
		strSlice := []string{"a", "b"}

		entity := &SlicePointerTypes{
			ID:      1,
			BoolPtr: &boolSlice,
			IntPtr:  &intSlice,
			StrPtr:  &strSlice,
			TimePtr: nil,
		}

		objValue, err := helper.GetObjectValue(entity)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		target := &SlicePointerTypes{}
		err = helper.UpdateEntity(objValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.IntPtr == nil {
			t.Error("IntPtr should not be nil")
		} else if len(*target.IntPtr) != 3 {
			t.Errorf("IntPtr length mismatch: expected 3, got %d", len(*target.IntPtr))
		}
	})
}

func TestMultipleRoundTrips(t *testing.T) {
	original := NewBasicTypes()

	current := original
	for i := 0; i < 5; i++ {
		objValue, err := helper.GetObjectValue(current)
		if err != nil {
			t.Fatalf("Round %d: GetObjectValue failed: %v", i+1, err)
		}

		jsonData, err := remote.EncodeObjectValue(objValue)
		if err != nil {
			t.Fatalf("Round %d: EncodeObjectValue failed: %v", i+1, err)
		}

		decodedValue, err := remote.DecodeObjectValue(jsonData)
		if err != nil {
			t.Fatalf("Round %d: DecodeObjectValue failed: %v", i+1, err)
		}

		next := &BasicTypes{}
		err = helper.UpdateEntity(decodedValue, next)
		if err != nil {
			t.Fatalf("Round %d: UpdateEntity failed: %v", i+1, err)
		}

		current = next
	}

	if current.ID != original.ID {
		t.Errorf("After 5 round trips, ID mismatch: expected %d, got %d", original.ID, current.ID)
	}
	if current.Str != original.Str {
		t.Errorf("After 5 round trips, Str mismatch: expected %s, got %s", original.Str, current.Str)
	}
}
