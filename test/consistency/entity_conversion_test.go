package consistency

import (
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestEntityToObjectValueBasicTypes(t *testing.T) {
	original := NewBasicTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	if objValue.GetName() != "BasicTypes" {
		t.Errorf("expected name 'BasicTypes', got '%s'", objValue.GetName())
	}

	if objValue.GetPkgPath() != "github.com/muidea/magicOrm/test/consistency" {
		t.Errorf("unexpected pkgPath: %s", objValue.GetPkgPath())
	}

	fields := objValue.GetValue()
	if len(fields) != 16 {
		t.Errorf("expected 16 fields, got %d", len(fields))
	}

	fieldMap := make(map[string]any)
	for _, f := range fields {
		fieldMap[f.GetName()] = f.Get()
	}

	if fieldMap["id"] != int(1) {
		t.Errorf("id field mismatch: %v", fieldMap["id"])
	}
	if fieldMap["bool"] != true {
		t.Errorf("bool field mismatch: %v", fieldMap["bool"])
	}
	if fieldMap["str"] != "hello world" {
		t.Errorf("str field mismatch: %v", fieldMap["str"])
	}
}

func TestEntityToObjectValuePointerTypes(t *testing.T) {
	original := NewPointerTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	fields := objValue.GetValue()
	fieldMap := make(map[string]any)
	for _, f := range fields {
		fieldMap[f.GetName()] = f.Get()
	}

	if fieldMap["id"] != int(1) {
		t.Errorf("id field mismatch: %v", fieldMap["id"])
	}

	if fieldMap["bool"] == nil {
		t.Error("bool field should not be nil")
	}
}

func TestEntityToObjectValueSliceTypes(t *testing.T) {
	original := NewSliceTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	fields := objValue.GetValue()
	fieldMap := make(map[string]any)
	for _, f := range fields {
		fieldMap[f.GetName()] = f.Get()
	}

	ints, ok := fieldMap["ints"].([]int)
	if !ok {
		t.Errorf("ints field should be []int, got: %T", fieldMap["ints"])
	} else if len(ints) != 3 {
		t.Errorf("ints slice length mismatch: %d", len(ints))
	}

	strs, ok := fieldMap["strs"].([]string)
	if !ok {
		t.Errorf("strs field should be []string, got: %T", fieldMap["strs"])
	} else if len(strs) != 3 {
		t.Errorf("strs slice length mismatch: %d", len(strs))
	}
}

func TestObjectValueToEntityBasicTypes(t *testing.T) {
	original := NewBasicTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	target := &BasicTypes{}
	err = helper.UpdateEntity(objValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.ID != original.ID {
		t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
	}
	if target.Bool != original.Bool {
		t.Errorf("Bool mismatch: expected %v, got %v", original.Bool, target.Bool)
	}
	if target.Str != original.Str {
		t.Errorf("Str mismatch: expected %s, got %s", original.Str, target.Str)
	}
	if target.Int != original.Int {
		t.Errorf("Int mismatch: expected %d, got %d", original.Int, target.Int)
	}
	if target.Float != original.Float {
		t.Errorf("Float mismatch: expected %f, got %f", original.Float, target.Float)
	}
	if target.Double != original.Double {
		t.Errorf("Double mismatch: expected %f, got %f", original.Double, target.Double)
	}
}

func TestObjectValueToEntityPointerTypes(t *testing.T) {
	original := NewPointerTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	target := &PointerTypes{}
	err = helper.UpdateEntity(objValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.ID != original.ID {
		t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
	}

	if target.Int == nil {
		t.Error("Int pointer should not be nil")
	} else if *target.Int != *original.Int {
		t.Errorf("Int mismatch: expected %d, got %d", *original.Int, *target.Int)
	}

	if target.Str == nil {
		t.Error("Str pointer should not be nil")
	} else if *target.Str != *original.Str {
		t.Errorf("Str mismatch: expected %s, got %s", *original.Str, *target.Str)
	}
}

func TestObjectValueToEntitySliceTypes(t *testing.T) {
	original := NewSliceTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	target := &SliceTypes{}
	err = helper.UpdateEntity(objValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if len(target.Ints) != len(original.Ints) {
		t.Errorf("Ints length mismatch: expected %d, got %d", len(original.Ints), len(target.Ints))
	}
	for i := range original.Ints {
		if target.Ints[i] != original.Ints[i] {
			t.Errorf("Ints[%d] mismatch: expected %d, got %d", i, original.Ints[i], target.Ints[i])
		}
	}

	if len(target.Strs) != len(original.Strs) {
		t.Errorf("Strs length mismatch: expected %d, got %d", len(original.Strs), len(target.Strs))
	}
}

func TestEntityRoundTripBasicTypes(t *testing.T) {
	original := NewBasicTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	target := &BasicTypes{}
	err = helper.UpdateEntity(objValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if !compareBasicTypes(original, target) {
		t.Error("BasicTypes round trip failed")
	}
}

func TestEntityRoundTripPointerTypes(t *testing.T) {
	original := NewPointerTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	target := &PointerTypes{}
	err = helper.UpdateEntity(objValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if !comparePointerTypes(original, target) {
		t.Error("PointerTypes round trip failed")
	}
}

func TestEntityRoundTripSliceTypes(t *testing.T) {
	original := NewSliceTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	target := &SliceTypes{}
	err = helper.UpdateEntity(objValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if !compareSliceTypes(original, target) {
		t.Error("SliceTypes round trip failed")
	}
}

func compareBasicTypes(a, b *BasicTypes) bool {
	return a.ID == b.ID &&
		a.Bool == b.Bool &&
		a.Int8 == b.Int8 &&
		a.Int16 == b.Int16 &&
		a.Int32 == b.Int32 &&
		a.Int64 == b.Int64 &&
		a.Int == b.Int &&
		a.UInt8 == b.UInt8 &&
		a.UInt16 == b.UInt16 &&
		a.UInt32 == b.UInt32 &&
		a.UInt64 == b.UInt64 &&
		a.UInt == b.UInt &&
		a.Float == b.Float &&
		a.Double == b.Double &&
		a.Str == b.Str
}

func comparePointerTypes(a, b *PointerTypes) bool {
	if a.ID != b.ID {
		return false
	}

	if (a.Bool == nil) != (b.Bool == nil) {
		return false
	}
	if a.Bool != nil && b.Bool != nil && *a.Bool != *b.Bool {
		return false
	}

	if (a.Int == nil) != (b.Int == nil) {
		return false
	}
	if a.Int != nil && b.Int != nil && *a.Int != *b.Int {
		return false
	}

	if (a.Str == nil) != (b.Str == nil) {
		return false
	}
	if a.Str != nil && b.Str != nil && *a.Str != *b.Str {
		return false
	}

	return true
}

func compareSliceTypes(a, b *SliceTypes) bool {
	if len(a.Ints) != len(b.Ints) {
		return false
	}
	for i := range a.Ints {
		if a.Ints[i] != b.Ints[i] {
			return false
		}
	}

	if len(a.Strs) != len(b.Strs) {
		return false
	}
	for i := range a.Strs {
		if a.Strs[i] != b.Strs[i] {
			return false
		}
	}

	return true
}

var _ = remote.EncodeObjectValue
