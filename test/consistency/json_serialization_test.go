package consistency

import (
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestObjectJSONRoundTrip(t *testing.T) {
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

	t.Logf("Object JSON: %s", string(jsonData))

	decodedObject, err := helper.DecodeObject(jsonData)
	if err != nil {
		t.Fatalf("DecodeObject failed: %v", err)
	}

	if !remote.CompareObject(object, decodedObject) {
		t.Error("Object round trip through JSON failed")
	}
}

func TestObjectValueJSONRoundTrip(t *testing.T) {
	original := NewBasicTypes()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	jsonData, err := remote.EncodeObjectValue(objValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON data should not be empty")
	}

	t.Logf("ObjectValue JSON: %s", string(jsonData))

	decodedValue, err := remote.DecodeObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}

	if !remote.CompareObjectValue(objValue, decodedValue) {
		t.Error("ObjectValue round trip through JSON failed")
	}
}

func TestEntityFullJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original any
		target   func() any
	}{
		{
			name:     "BasicTypes",
			original: NewBasicTypes(),
			target:   func() any { return &BasicTypes{} },
		},
		{
			name:     "PointerTypes",
			original: NewPointerTypes(),
			target:   func() any { return &PointerTypes{} },
		},
		{
			name:     "SliceTypes",
			original: NewSliceTypes(),
			target:   func() any { return &SliceTypes{} },
		},
		{
			name:     "NestedParent",
			original: NewNestedParent(),
			target:   func() any { return &NestedParent{} },
		},
		{
			name:     "NestedSliceParent",
			original: NewNestedSliceParent(),
			target:   func() any { return &NestedSliceParent{} },
		},
		{
			name:     "DeepLevel3",
			original: NewDeepLevel3(),
			target:   func() any { return &DeepLevel3{} },
		},
		{
			name:     "ComplexEntity",
			original: NewComplexEntity(),
			target:   func() any { return &ComplexEntity{} },
		},
		{
			name:     "AllInOne",
			original: NewAllInOne(),
			target:   func() any { return &AllInOne{} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objValue, err := helper.GetObjectValue(tt.original)
			if err != nil {
				t.Fatalf("GetObjectValue failed: %v", err)
			}

			jsonData, err := remote.EncodeObjectValue(objValue)
			if err != nil {
				t.Fatalf("EncodeObjectValue failed: %v", err)
			}

			t.Logf("%s JSON length: %d", tt.name, len(jsonData))

			decodedValue, err := remote.DecodeObjectValue(jsonData)
			if err != nil {
				t.Fatalf("DecodeObjectValue failed: %v", err)
			}

			target := tt.target()
			err = helper.UpdateEntity(decodedValue, target)
			if err != nil {
				t.Fatalf("UpdateEntity failed: %v", err)
			}

			if !compareEntities(tt.original, target) {
				t.Errorf("%s full JSON round trip failed", tt.name)
			}
		})
	}
}

func TestSliceObjectValueJSONRoundTrip(t *testing.T) {
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

	jsonData, err := remote.EncodeSliceObjectValue(sliceValue)
	if err != nil {
		t.Fatalf("EncodeSliceObjectValue failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON data should not be empty")
	}

	t.Logf("SliceObjectValue JSON: %s", string(jsonData))

	decodedValue, err := remote.DecodeSliceObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeSliceObjectValue failed: %v", err)
	}

	if !remote.CompareSliceObjectValue(sliceValue, decodedValue) {
		t.Error("SliceObjectValue round trip through JSON failed")
	}
}

func TestSliceFullJSONRoundTrip(t *testing.T) {
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
		if target[i].ID != original[i].ID {
			t.Errorf("[%d] ID mismatch: expected %d, got %d", i, original[i].ID, target[i].ID)
		}
		if target[i].Str != original[i].Str {
			t.Errorf("[%d] Str mismatch: expected %s, got %s", i, original[i].Str, target[i].Str)
		}
	}
}

func TestJSONNilFieldHandling(t *testing.T) {
	t.Run("nil pointer fields", func(t *testing.T) {
		original := &PointerTypes{
			ID:   1,
			Bool: nil,
			Int:  nil,
			Str:  nil,
			Time: nil,
		}

		objValue, err := helper.GetObjectValue(original)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		jsonData, err := remote.EncodeObjectValue(objValue)
		if err != nil {
			t.Fatalf("EncodeObjectValue failed: %v", err)
		}

		t.Logf("Nil fields JSON: %s", string(jsonData))

		decodedValue, err := remote.DecodeObjectValue(jsonData)
		if err != nil {
			t.Fatalf("DecodeObjectValue failed: %v", err)
		}

		target := &PointerTypes{}
		err = helper.UpdateEntity(decodedValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.ID != original.ID {
			t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
		}
	})
}

func TestJSONEmptySliceHandling(t *testing.T) {
	t.Run("empty slice fields", func(t *testing.T) {
		original := &SliceTypes{
			ID:    1,
			Ints:  []int{},
			Strs:  []string{},
			Bools: []bool{},
		}

		objValue, err := helper.GetObjectValue(original)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		jsonData, err := remote.EncodeObjectValue(objValue)
		if err != nil {
			t.Fatalf("EncodeObjectValue failed: %v", err)
		}

		t.Logf("Empty slice JSON: %s", string(jsonData))

		decodedValue, err := remote.DecodeObjectValue(jsonData)
		if err != nil {
			t.Fatalf("DecodeObjectValue failed: %v", err)
		}

		target := &SliceTypes{}
		err = helper.UpdateEntity(decodedValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.ID != original.ID {
			t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
		}
	})
}

func TestJSONNestedStructures(t *testing.T) {
	t.Run("nested struct", func(t *testing.T) {
		original := NewNestedParent()

		objValue, err := helper.GetObjectValue(original)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		jsonData, err := remote.EncodeObjectValue(objValue)
		if err != nil {
			t.Fatalf("EncodeObjectValue failed: %v", err)
		}

		t.Logf("Nested JSON: %s", string(jsonData))

		decodedValue, err := remote.DecodeObjectValue(jsonData)
		if err != nil {
			t.Fatalf("DecodeObjectValue failed: %v", err)
		}

		target := &NestedParent{}
		err = helper.UpdateEntity(decodedValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.ID != original.ID {
			t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
		}
		if target.Name != original.Name {
			t.Errorf("Name mismatch: expected %s, got %s", original.Name, target.Name)
		}
		if target.Child == nil {
			t.Error("Child should not be nil")
		} else if target.Child.ID != original.Child.ID {
			t.Errorf("Child.ID mismatch: expected %d, got %d", original.Child.ID, target.Child.ID)
		}
	})

	t.Run("nested slice", func(t *testing.T) {
		original := NewNestedSliceParent()

		objValue, err := helper.GetObjectValue(original)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		jsonData, err := remote.EncodeObjectValue(objValue)
		if err != nil {
			t.Fatalf("EncodeObjectValue failed: %v", err)
		}

		t.Logf("Nested slice JSON: %s", string(jsonData))

		decodedValue, err := remote.DecodeObjectValue(jsonData)
		if err != nil {
			t.Fatalf("DecodeObjectValue failed: %v", err)
		}

		target := &NestedSliceParent{}
		err = helper.UpdateEntity(decodedValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.ID != original.ID {
			t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
		}
		if len(target.Items) != len(original.Items) {
			t.Errorf("Items length mismatch: expected %d, got %d", len(original.Items), len(target.Items))
		}
	})

	t.Run("deep nested", func(t *testing.T) {
		original := NewDeepLevel3()

		objValue, err := helper.GetObjectValue(original)
		if err != nil {
			t.Fatalf("GetObjectValue failed: %v", err)
		}

		jsonData, err := remote.EncodeObjectValue(objValue)
		if err != nil {
			t.Fatalf("EncodeObjectValue failed: %v", err)
		}

		t.Logf("Deep nested JSON: %s", string(jsonData))

		decodedValue, err := remote.DecodeObjectValue(jsonData)
		if err != nil {
			t.Fatalf("DecodeObjectValue failed: %v", err)
		}

		target := &DeepLevel3{}
		err = helper.UpdateEntity(decodedValue, target)
		if err != nil {
			t.Fatalf("UpdateEntity failed: %v", err)
		}

		if target.ID != original.ID {
			t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
		}
		if target.Level == nil {
			t.Error("Level should not be nil")
		} else if target.Level.Level == nil {
			t.Error("Level.Level should not be nil")
		}
	})
}

func TestJSONComplexEntity(t *testing.T) {
	original := NewComplexEntity()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	jsonData, err := remote.EncodeObjectValue(objValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}

	t.Logf("ComplexEntity JSON length: %d", len(jsonData))

	decodedValue, err := remote.DecodeObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}

	target := &ComplexEntity{}
	err = helper.UpdateEntity(decodedValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.ID != original.ID {
		t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
	}
	if target.Name != original.Name {
		t.Errorf("Name mismatch: expected %s, got %s", original.Name, target.Name)
	}
	if len(target.Flags) != len(original.Flags) {
		t.Errorf("Flags length mismatch: expected %d, got %d", len(original.Flags), len(target.Flags))
	}
	if len(target.Numbers) != len(original.Numbers) {
		t.Errorf("Numbers length mismatch: expected %d, got %d", len(original.Numbers), len(target.Numbers))
	}
	if target.Child == nil {
		t.Error("Child should not be nil")
	}
	if len(target.Items) != len(original.Items) {
		t.Errorf("Items length mismatch: expected %d, got %d", len(original.Items), len(target.Items))
	}
}

func TestJSONAllInOne(t *testing.T) {
	original := NewAllInOne()

	objValue, err := helper.GetObjectValue(original)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	jsonData, err := remote.EncodeObjectValue(objValue)
	if err != nil {
		t.Fatalf("EncodeObjectValue failed: %v", err)
	}

	t.Logf("AllInOne JSON length: %d", len(jsonData))

	decodedValue, err := remote.DecodeObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeObjectValue failed: %v", err)
	}

	target := &AllInOne{}
	err = helper.UpdateEntity(decodedValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.ID != original.ID {
		t.Errorf("ID mismatch: expected %d, got %d", original.ID, target.ID)
	}
	if target.Bool != original.Bool {
		t.Errorf("Bool mismatch: expected %v, got %v", original.Bool, target.Bool)
	}
	if target.Int != original.Int {
		t.Errorf("Int mismatch: expected %d, got %d", original.Int, target.Int)
	}
	if target.Str != original.Str {
		t.Errorf("Str mismatch: expected %s, got %s", original.Str, target.Str)
	}
	if len(target.BoolSlice) != len(original.BoolSlice) {
		t.Errorf("BoolSlice length mismatch: expected %d, got %d", len(original.BoolSlice), len(target.BoolSlice))
	}
	if len(target.IntSlice) != len(original.IntSlice) {
		t.Errorf("IntSlice length mismatch: expected %d, got %d", len(original.IntSlice), len(target.IntSlice))
	}
	if len(target.StrSlice) != len(original.StrSlice) {
		t.Errorf("StrSlice length mismatch: expected %d, got %d", len(original.StrSlice), len(target.StrSlice))
	}
	if target.Child == nil {
		t.Error("Child should not be nil")
	}
	if len(target.Children) != len(original.Children) {
		t.Errorf("Children length mismatch: expected %d, got %d", len(original.Children), len(target.Children))
	}
}

func compareEntities(a, b any) bool {
	switch va := a.(type) {
	case *BasicTypes:
		vb, ok := b.(*BasicTypes)
		if !ok {
			return false
		}
		return compareBasicTypes(va, vb)
	case *PointerTypes:
		vb, ok := b.(*PointerTypes)
		if !ok {
			return false
		}
		return comparePointerTypes(va, vb)
	case *SliceTypes:
		vb, ok := b.(*SliceTypes)
		if !ok {
			return false
		}
		return compareSliceTypes(va, vb)
	case *NestedParent:
		vb, ok := b.(*NestedParent)
		if !ok {
			return false
		}
		return compareNestedParent(va, vb)
	case *NestedSliceParent:
		vb, ok := b.(*NestedSliceParent)
		if !ok {
			return false
		}
		return compareNestedSliceParent(va, vb)
	case *DeepLevel3:
		vb, ok := b.(*DeepLevel3)
		if !ok {
			return false
		}
		return compareDeepLevel3(va, vb)
	case *ComplexEntity:
		vb, ok := b.(*ComplexEntity)
		if !ok {
			return false
		}
		return compareComplexEntity(va, vb)
	case *AllInOne:
		vb, ok := b.(*AllInOne)
		if !ok {
			return false
		}
		return compareAllInOne(va, vb)
	default:
		return false
	}
}

func compareNestedParent(a, b *NestedParent) bool {
	if a.ID != b.ID || a.Name != b.Name {
		return false
	}
	if (a.Child == nil) != (b.Child == nil) {
		return false
	}
	if a.Child != nil && b.Child != nil {
		if a.Child.ID != b.Child.ID || a.Child.Name != b.Child.Name {
			return false
		}
	}
	return true
}

func compareNestedSliceParent(a, b *NestedSliceParent) bool {
	if a.ID != b.ID || a.Name != b.Name {
		return false
	}
	if len(a.Items) != len(b.Items) {
		return false
	}
	for i := range a.Items {
		if a.Items[i].ID != b.Items[i].ID || a.Items[i].Value != b.Items[i].Value {
			return false
		}
	}
	return true
}

func compareNestedSlicePtrParent(a, b *NestedSlicePtrParent) bool {
	if a.ID != b.ID || a.Name != b.Name {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if (a.Children[i] == nil) != (b.Children[i] == nil) {
			return false
		}
		if a.Children[i] != nil && b.Children[i] != nil {
			if a.Children[i].ID != b.Children[i].ID || a.Children[i].Name != b.Children[i].Name {
				return false
			}
		}
	}
	return true
}

func compareDeepLevel3(a, b *DeepLevel3) bool {
	if a.ID != b.ID {
		return false
	}
	if (a.Level == nil) != (b.Level == nil) {
		return false
	}
	if a.Level != nil && b.Level != nil {
		if a.Level.ID != b.Level.ID {
			return false
		}
		if (a.Level.Level == nil) != (b.Level.Level == nil) {
			return false
		}
		if a.Level.Level != nil && b.Level.Level != nil {
			if a.Level.Level.ID != b.Level.Level.ID || a.Level.Level.Value != b.Level.Level.Value {
				return false
			}
		}
	}
	return true
}

func compareComplexEntity(a, b *ComplexEntity) bool {
	if a.ID != b.ID || a.Name != b.Name {
		return false
	}
	if (a.Count == nil) != (b.Count == nil) {
		return false
	}
	if a.Count != nil && b.Count != nil && *a.Count != *b.Count {
		return false
	}
	if len(a.Flags) != len(b.Flags) || len(a.Numbers) != len(b.Numbers) || len(a.Items) != len(b.Items) {
		return false
	}
	for i := range a.Flags {
		if a.Flags[i] != b.Flags[i] {
			return false
		}
	}
	for i := range a.Numbers {
		if a.Numbers[i] != b.Numbers[i] {
			return false
		}
	}
	// 嵌套对象 Child
	if (a.Child == nil) != (b.Child == nil) {
		return false
	}
	if a.Child != nil && b.Child != nil {
		if a.Child.ID != b.Child.ID || a.Child.Name != b.Child.Name {
			return false
		}
	}
	// 对象切片 Items
	for i := range a.Items {
		if a.Items[i].ID != b.Items[i].ID || a.Items[i].Value != b.Items[i].Value {
			return false
		}
	}
	return true
}

func compareAllInOne(a, b *AllInOne) bool {
	if a.ID != b.ID || a.Bool != b.Bool || a.Int != b.Int || a.Str != b.Str {
		return false
	}
	if len(a.BoolSlice) != len(b.BoolSlice) || len(a.IntSlice) != len(b.IntSlice) || len(a.StrSlice) != len(b.StrSlice) {
		return false
	}
	for i := range a.BoolSlice {
		if a.BoolSlice[i] != b.BoolSlice[i] {
			return false
		}
	}
	for i := range a.IntSlice {
		if a.IntSlice[i] != b.IntSlice[i] {
			return false
		}
	}
	for i := range a.StrSlice {
		if a.StrSlice[i] != b.StrSlice[i] {
			return false
		}
	}
	// 嵌套对象 Child
	if (a.Child == nil) != (b.Child == nil) {
		return false
	}
	if a.Child != nil && b.Child != nil {
		if a.Child.ID != b.Child.ID || a.Child.Name != b.Child.Name {
			return false
		}
	}
	// 对象切片 Children
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if a.Children[i].ID != b.Children[i].ID || a.Children[i].Name != b.Children[i].Name {
			return false
		}
	}
	return true
}
