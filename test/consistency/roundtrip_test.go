package consistency

import (
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestLocalRemoteRoundTripBasicTypes(t *testing.T) {
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

func TestLocalRemoteRoundTripWithJSON(t *testing.T) {
	original := NewBasicTypes()

	objValue, err := helper.GetObjectValue(original)
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

	if !compareBasicTypes(original, target) {
		t.Error("BasicTypes round trip with JSON failed")
	}
}

func TestMultipleRoundTripsAllTypes(t *testing.T) {
	types := []struct {
		name     string
		original any
		target   func() any
		compare  func(a, b any) bool
	}{
		{
			name:     "BasicTypes",
			original: NewBasicTypes(),
			target:   func() any { return &BasicTypes{} },
			compare: func(a, b any) bool {
				return compareBasicTypes(a.(*BasicTypes), b.(*BasicTypes))
			},
		},
		{
			name:     "PointerTypes",
			original: NewPointerTypes(),
			target:   func() any { return &PointerTypes{} },
			compare: func(a, b any) bool {
				return comparePointerTypes(a.(*PointerTypes), b.(*PointerTypes))
			},
		},
		{
			name:     "NestedParent",
			original: NewNestedParent(),
			target:   func() any { return &NestedParent{} },
			compare: func(a, b any) bool {
				return compareNestedParent(a.(*NestedParent), b.(*NestedParent))
			},
		},
		{
			name:     "ComplexEntity",
			original: NewComplexEntity(),
			target:   func() any { return &ComplexEntity{} },
			compare: func(a, b any) bool {
				return compareComplexEntity(a.(*ComplexEntity), b.(*ComplexEntity))
			},
		},
	}

	for _, tt := range types {
		t.Run(tt.name, func(t *testing.T) {
			current := tt.original

			for i := 0; i < 3; i++ {
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

				next := tt.target()
				err = helper.UpdateEntity(decodedValue, next)
				if err != nil {
					t.Fatalf("Round %d: UpdateEntity failed: %v", i+1, err)
				}

				current = next
			}

			if !tt.compare(tt.original, current) {
				t.Errorf("After 3 round trips, data mismatch for %s", tt.name)
			}
		})
	}
}

func TestSliceRoundTrip(t *testing.T) {
	original := []*BasicTypes{
		NewBasicTypes(),
		{ID: 2, Str: "test2"},
		{ID: 3, Str: "test3"},
	}

	sliceValue, err := helper.GetSliceObjectValue(original)
	if err != nil {
		t.Fatalf("GetSliceObjectValue failed: %v", err)
	}

	jsonData, err := remote.EncodeSliceObjectValue(sliceValue)
	if err != nil {
		t.Fatalf("EncodeSliceObjectValue failed: %v", err)
	}

	decodedValue, err := remote.DecodeSliceObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeSliceObjectValue failed: %v", err)
	}

	target := []*BasicTypes{}
	err = helper.UpdateSliceEntity(decodedValue, &target)
	if err != nil {
		t.Fatalf("UpdateSliceEntity failed: %v", err)
	}

	if len(target) != len(original) {
		t.Fatalf("length mismatch: expected %d, got %d", len(original), len(target))
	}

	for i := range original {
		if !compareBasicTypes(original[i], target[i]) {
			t.Errorf("[%d] round trip failed", i)
		}
	}
}

func TestCrossProviderConsistency(t *testing.T) {
	original := NewBasicTypes()

	localObjValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	jsonData, err := remote.EncodeObjectValue(localObjValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}

	remoteObjValue, err := remote.DecodeObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}

	if !remote.CompareObjectValue(localObjValue, remoteObjValue) {
		t.Error("Local and remote ObjectValue mismatch after JSON round trip")
	}

	targetFromLocal := &BasicTypes{}
	err = helper.UpdateEntity(localObjValue, targetFromLocal)
	if err != nil {
		t.Fatalf("UpdateEntity from local failed: %v", err)
	}

	targetFromRemote := &BasicTypes{}
	err = helper.UpdateEntity(remoteObjValue, targetFromRemote)
	if err != nil {
		t.Fatalf("UpdateEntity from remote failed: %v", err)
	}

	if !compareBasicTypes(targetFromLocal, targetFromRemote) {
		t.Error("Entities from local and remote ObjectValue mismatch")
	}
}

func TestObjectRoundTrip(t *testing.T) {
	original := NewBasicTypes()

	object, err := helper.GetObject(original)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	jsonData, err := helper.EncodeObject(object)
	if err != nil {
		t.Fatalf("EncodeObject failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON data should not be empty")
	}

	decodedObject, err := helper.DecodeObject(jsonData)
	if err != nil {
		t.Fatalf("DecodeObject failed: %v", err)
	}

	if !remote.CompareObject(object, decodedObject) {
		t.Error("Object round trip failed")
	}
}

func TestFullRoundTripChain(t *testing.T) {
	original := NewComplexEntity()

	objectValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	objValueJSON, err := remote.EncodeObjectValue(objectValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}

	decodedObjValue, err := remote.DecodeObjectValue(objValueJSON)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}

	target := &ComplexEntity{}
	err = helper.UpdateEntity(decodedObjValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.ID != original.ID {
		t.Errorf("ID mismatch after full chain: expected %d, got %d", original.ID, target.ID)
	}
	if target.Name != original.Name {
		t.Errorf("Name mismatch after full chain: expected %s, got %s", original.Name, target.Name)
	}
}

func TestStressRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	original := NewAllInOne()

	for i := 0; i < 100; i++ {
		objValue, err := helper.GetObjectValue(original)
		if err != nil {
			t.Fatalf("Iteration %d: GetObjectValue failed: %v", i, err)
		}

		jsonData, err := remote.EncodeObjectValue(objValue)
		if err != nil {
			t.Fatalf("Iteration %d: EncodeObjectValue failed: %v", i, err)
		}

		decodedValue, err := remote.DecodeObjectValue(jsonData)
		if err != nil {
			t.Fatalf("Iteration %d: DecodeObjectValue failed: %v", i, err)
		}

		target := &AllInOne{}
		err = helper.UpdateEntity(decodedValue, target)
		if err != nil {
			t.Fatalf("Iteration %d: UpdateEntity failed: %v", i, err)
		}

		if !compareAllInOne(original, target) {
			t.Fatalf("Iteration %d: data mismatch", i)
		}
	}

	t.Log("100 iterations completed successfully")
}
