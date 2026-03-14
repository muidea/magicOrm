package consistency

import (
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestSliceToSliceObjectValueBasicTypes(t *testing.T) {
	original := []*BasicTypes{
		NewBasicTypes(),
		{
			ID:     2,
			Bool:   false,
			Int:    -200,
			Str:    "test2",
			Float:  2.71,
			Double: 2.718281828,
		},
	}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	if sliceValue.GetName() != "BasicTypes" {
		t.Errorf("expected name 'BasicTypes', got '%s'", sliceValue.GetName())
	}

	values := sliceValue.GetValue()
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}

	id1 := values[0].GetFieldValue("id")
	if id1 != int(1) {
		t.Errorf("first element id mismatch: %v", id1)
	}

	id2 := values[1].GetFieldValue("id")
	if id2 != int(2) {
		t.Errorf("second element id mismatch: %v", id2)
	}
}

func TestSliceToSliceObjectValuePointerTypes(t *testing.T) {
	original := []*PointerTypes{NewPointerTypes()}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	values := sliceValue.GetValue()
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}

	id := values[0].GetFieldValue("id")
	if id != int(1) {
		t.Errorf("id mismatch: %v", id)
	}
}

func TestSliceToSliceObjectValueNestedTypes(t *testing.T) {
	original := []*NestedParent{NewNestedParent()}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	values := sliceValue.GetValue()
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}

	name := values[0].GetFieldValue("name")
	if name != "parent" {
		t.Errorf("name mismatch: %v", name)
	}
}

func TestSliceObjectValueToSliceBasicTypes(t *testing.T) {
	original := []*BasicTypes{
		NewBasicTypes(),
		{
			ID:     2,
			Bool:   false,
			Int:    -200,
			Str:    "test2",
			Float:  2.71,
			Double: 2.718281828,
		},
	}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	target := []*BasicTypes{}
	err = helper.UpdateSliceEntity(sliceValue, &target)
	if err != nil {
		t.Fatalf("UpdateSliceEntity failed: %v", err)
	}

	if len(target) != len(original) {
		t.Fatalf("length mismatch: expected %d, got %d", len(original), len(target))
	}

	for i := range original {
		if target[i].ID != original[i].ID {
			t.Errorf("[%d] ID mismatch: expected %d, got %d", i, original[i].ID, target[i].ID)
		}
		if target[i].Str != original[i].Str {
			t.Errorf("[%d] Str mismatch: expected %s, got %s", i, original[i].Str, target[i].Str)
		}
	}
}

func TestSliceObjectValueToSlicePointerTypes(t *testing.T) {
	original := []*PointerTypes{NewPointerTypes()}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	target := []*PointerTypes{}
	err = helper.UpdateSliceEntity(sliceValue, &target)
	if err != nil {
		t.Fatalf("UpdateSliceEntity failed: %v", err)
	}

	if len(target) != len(original) {
		t.Fatalf("length mismatch: expected %d, got %d", len(original), len(target))
	}

	if target[0].ID != original[0].ID {
		t.Errorf("ID mismatch: expected %d, got %d", original[0].ID, target[0].ID)
	}
}

func TestSliceRoundTripBasicTypes(t *testing.T) {
	original := []*BasicTypes{
		NewBasicTypes(),
		{
			ID:     2,
			Bool:   false,
			Int:    -200,
			Str:    "test2",
			Float:  2.71,
			Double: 2.718281828,
		},
	}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	target := []*BasicTypes{}
	err = helper.UpdateSliceEntity(sliceValue, &target)
	if err != nil {
		t.Fatalf("UpdateSliceEntity failed: %v", err)
	}

	if !compareBasicTypesSlice(original, target) {
		t.Error("BasicTypes slice round trip failed")
	}
}

func TestEmptySliceHandling(t *testing.T) {
	t.Run("empty slice to SliceObjectValue", func(t *testing.T) {
		original := []*BasicTypes{}

		sliceValue, err := helper.GetSliceObjectValue(original)
		if err != nil {
			t.Fatalf("GetSliceObjectValue failed: %v", err)
		}

		if sliceValue == nil {
			t.Error("SliceObjectValue should not be nil for empty slice")
		}

		values := sliceValue.GetValue()
		if len(values) != 0 {
			t.Errorf("expected 0 values, got %d", len(values))
		}
	})

	t.Run("empty SliceObjectValue to slice", func(t *testing.T) {
		sliceValue := &remote.SliceObjectValue{
			Name:    "BasicTypes",
			PkgPath: "github.com/muidea/magicOrm/test/consistency",
			Values:  []*remote.ObjectValue{},
		}

		target := []*BasicTypes{}
		err := helper.UpdateSliceEntity(sliceValue, &target)
		if err != nil {
			t.Fatalf("UpdateSliceEntity failed: %v", err)
		}

		if len(target) != 0 {
			t.Errorf("expected 0 elements, got %d", len(target))
		}
	})

	t.Run("empty SliceObjectValue clears existing slice", func(t *testing.T) {
		sliceValue := &remote.SliceObjectValue{
			Name:    "BasicTypes",
			PkgPath: "github.com/muidea/magicOrm/test/consistency",
			Values:  []*remote.ObjectValue{},
		}

		target := []*BasicTypes{NewBasicTypes()}
		err := helper.UpdateSliceEntity(sliceValue, &target)
		if err != nil {
			t.Fatalf("UpdateSliceEntity failed: %v", err)
		}

		if target == nil || len(target) != 0 {
			t.Fatalf("expected assigned empty slice, got %#v", target)
		}
	})

	t.Run("nil SliceObjectValue leaves existing slice untouched", func(t *testing.T) {
		sliceValue := &remote.SliceObjectValue{
			Name:    "BasicTypes",
			PkgPath: "github.com/muidea/magicOrm/test/consistency",
		}

		target := []*BasicTypes{NewBasicTypes()}
		err := helper.UpdateSliceEntity(sliceValue, &target)
		if err != nil {
			t.Fatalf("UpdateSliceEntity failed: %v", err)
		}

		if len(target) != 1 || target[0] == nil || target[0].ID != 1 {
			t.Fatalf("nil slice should not overwrite target, got %#v", target)
		}
	})

	t.Run("non-empty SliceObjectValue replaces existing slice", func(t *testing.T) {
		original := []*BasicTypes{
			{ID: 9, Str: "replace-1"},
			{ID: 10, Str: "replace-2"},
		}
		sliceValue, err := helper.GetSliceObjectValue(original)
		if err != nil {
			t.Fatalf("GetSliceObjectValue failed: %v", err)
		}

		target := []*BasicTypes{NewBasicTypes()}
		err = helper.UpdateSliceEntity(sliceValue, &target)
		if err != nil {
			t.Fatalf("UpdateSliceEntity failed: %v", err)
		}

		if len(target) != len(original) {
			t.Fatalf("expected %d elements, got %d", len(original), len(target))
		}
		if target[0].ID != 9 || target[1].ID != 10 {
			t.Fatalf("expected replacement values, got %#v", target)
		}
	})
}

func TestSingleElementSlice(t *testing.T) {
	original := []*BasicTypes{NewBasicTypes()}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	target := []*BasicTypes{}
	err = helper.UpdateSliceEntity(sliceValue, &target)
	if err != nil {
		t.Fatalf("UpdateSliceEntity failed: %v", err)
	}

	if len(target) != 1 {
		t.Fatalf("expected 1 element, got %d", len(target))
	}

	if !compareBasicTypes(original[0], target[0]) {
		t.Error("single element slice round trip failed")
	}
}

func TestLargeSlice(t *testing.T) {
	original := make([]*BasicTypes, 100)
	for i := range original {
		original[i] = &BasicTypes{
			ID:     i + 1,
			Bool:   i%2 == 0,
			Int:    i,
			Str:    "test",
			Float:  float32(i) / 10,
			Double: float64(i) / 100,
		}
	}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	target := []*BasicTypes{}
	err = helper.UpdateSliceEntity(sliceValue, &target)
	if err != nil {
		t.Fatalf("UpdateSliceEntity failed: %v", err)
	}

	if len(target) != len(original) {
		t.Fatalf("length mismatch: expected %d, got %d", len(original), len(target))
	}

	for i := range original {
		if target[i].ID != original[i].ID {
			t.Errorf("[%d] ID mismatch", i)
		}
	}
}

func compareBasicTypesSlice(a, b []*BasicTypes) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !compareBasicTypes(a[i], b[i]) {
			return false
		}
	}
	return true
}
