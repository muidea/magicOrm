package consistency

import (
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestSingleLevelNesting(t *testing.T) {
	original := NewNestedParent()

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
		t.Fatal("Child should not be nil")
	}
	if target.Child.ID != original.Child.ID {
		t.Errorf("Child.ID mismatch: expected %d, got %d", original.Child.ID, target.Child.ID)
	}
	if target.Child.Name != original.Child.Name {
		t.Errorf("Child.Name mismatch: expected %s, got %s", original.Child.Name, target.Child.Name)
	}
}

func TestMultiLevelNesting(t *testing.T) {
	original := NewDeepLevel3()

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

	target := &DeepLevel3{}
	err = helper.UpdateEntity(decodedValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.ID != original.ID {
		t.Errorf("Level3.ID mismatch: expected %d, got %d", original.ID, target.ID)
	}
	if target.Level == nil {
		t.Fatal("Level (Level2) should not be nil")
	}
	if target.Level.ID != original.Level.ID {
		t.Errorf("Level2.ID mismatch: expected %d, got %d", original.Level.ID, target.Level.ID)
	}
	if target.Level.Level == nil {
		t.Fatal("Level.Level (Level1) should not be nil")
	}
	if target.Level.Level.ID != original.Level.Level.ID {
		t.Errorf("Level1.ID mismatch: expected %d, got %d", original.Level.Level.ID, target.Level.Level.ID)
	}
	if target.Level.Level.Value != original.Level.Level.Value {
		t.Errorf("Level1.Value mismatch: expected %s, got %s", original.Level.Level.Value, target.Level.Level.Value)
	}
}

func TestNestedSlice(t *testing.T) {
	original := NewNestedSliceParent()

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

	target := &NestedSliceParent{}
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
	if len(target.Items) != len(original.Items) {
		t.Fatalf("Items length mismatch: expected %d, got %d", len(original.Items), len(target.Items))
	}

	for i := range original.Items {
		if target.Items[i].ID != original.Items[i].ID {
			t.Errorf("Items[%d].ID mismatch: expected %d, got %d", i, original.Items[i].ID, target.Items[i].ID)
		}
		if target.Items[i].Value != original.Items[i].Value {
			t.Errorf("Items[%d].Value mismatch: expected %s, got %s", i, original.Items[i].Value, target.Items[i].Value)
		}
	}
}

func TestPointerNesting(t *testing.T) {
	original := &NestedParent{
		ID:    1,
		Name:  "parent with nil child",
		Child: nil,
	}

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

	target := &NestedParent{}
	err = helper.UpdateEntity(decodedValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.Child != nil {
		t.Error("Child should remain nil")
	}
}

func TestNestedSliceWithNilElements(t *testing.T) {
	original := &NestedSliceParent{
		ID:    1,
		Name:  "parent with empty items",
		Items: nil,
	}

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

	target := &NestedSliceParent{}
	err = helper.UpdateEntity(decodedValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}
}

func TestComplexNestedEntity(t *testing.T) {
	original := NewComplexEntity()

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

	if (original.Count == nil) != (target.Count == nil) {
		t.Error("Count nil state mismatch")
	}
	if original.Count != nil && target.Count != nil {
		if *target.Count != *original.Count {
			t.Errorf("Count mismatch: expected %d, got %d", *original.Count, *target.Count)
		}
	}

	if len(target.Flags) != len(original.Flags) {
		t.Errorf("Flags length mismatch: expected %d, got %d", len(original.Flags), len(target.Flags))
	}

	if len(target.Numbers) != len(original.Numbers) {
		t.Errorf("Numbers length mismatch: expected %d, got %d", len(original.Numbers), len(target.Numbers))
	}

	if (original.Child == nil) != (target.Child == nil) {
		t.Error("Child nil state mismatch")
	}

	if len(target.Items) != len(original.Items) {
		t.Errorf("Items length mismatch: expected %d, got %d", len(original.Items), len(target.Items))
	}
}

func TestNestedSliceEntity(t *testing.T) {
	original := []*NestedParent{
		NewNestedParent(),
		{
			ID:    2,
			Name:  "parent2",
			Child: &NestedChild{ID: 20, Name: "child2"},
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

	decodedValue, err := remote.DecodeSliceObjectValue(jsonData)
	if err != nil {
		t.Fatalf("DecodeSliceObjectValue failed: %v", err)
	}

	target := []*NestedParent{}
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
		if target[i].Name != original[i].Name {
			t.Errorf("[%d] Name mismatch: expected %s, got %s", i, original[i].Name, target[i].Name)
		}
		if target[i].Child == nil {
			t.Errorf("[%d] Child should not be nil", i)
		}
	}
}

func TestAllInOneNested(t *testing.T) {
	original := NewAllInOne()

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

	target := &AllInOne{}
	err = helper.UpdateEntity(decodedValue, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.Child == nil {
		t.Error("Child should not be nil")
	} else {
		if target.Child.ID != original.Child.ID {
			t.Errorf("Child.ID mismatch: expected %d, got %d", original.Child.ID, target.Child.ID)
		}
		if target.Child.Name != original.Child.Name {
			t.Errorf("Child.Name mismatch: expected %s, got %s", original.Child.Name, target.Child.Name)
		}
	}

	if len(target.Children) != len(original.Children) {
		t.Errorf("Children length mismatch: expected %d, got %d", len(original.Children), len(target.Children))
	} else {
		for i := range original.Children {
			if target.Children[i].ID != original.Children[i].ID {
				t.Errorf("Children[%d].ID mismatch", i)
			}
			if target.Children[i].Name != original.Children[i].Name {
				t.Errorf("Children[%d].Name mismatch", i)
			}
		}
	}
}
