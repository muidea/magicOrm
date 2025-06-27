package local

import (
	"testing"

	"github.com/muidea/magicOrm/model"
)

// TestModelProperties tests the basic properties of model
func TestModelProperties(t *testing.T) {
	// Create test structs
	type SimpleStruct struct {
		ID   int    `orm:"id key"`
		Name string `orm:"name"`
	}

	// Get model for simple struct
	simpleModel, err := GetEntityModel(&SimpleStruct{})
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test model basic properties
	if simpleModel.GetPkgPath() == "" {
		t.Errorf("Model should have a package path")
	}

	if simpleModel.GetName() == "" {
		t.Errorf("Model should have a name")
	}

	// Test field management
	fields := simpleModel.GetFields()
	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}

	idField := simpleModel.GetField("id")
	if idField == nil {
		t.Errorf("GetField(id) returned nil")
	} else if !model.IsPrimaryField(idField) {
		t.Errorf("id field should be primary key")
	}

	nameField := simpleModel.GetField("name")
	if nameField == nil {
		t.Errorf("GetField(name) returned nil")
	}

	// Test getting fields by index - alternative approach since GetFieldByIndex doesn't exist
	if len(fields) > 0 {
		field0 := fields[0]
		if field0 == nil {
			t.Errorf("First field is nil")
		}
	}

	if len(fields) > 1 {
		field1 := fields[1]
		if field1 == nil {
			t.Errorf("Second field is nil")
		}
	}

	// Test non-existent field
	nonExistentField := simpleModel.GetField("doesNotExist")
	if nonExistentField != nil {
		t.Errorf("GetField for non-existent field should return nil")
	}
}

// TestModelWithNestedStruct tests models with nested struct fields
func TestModelWithNestedStruct(t *testing.T) {
	// Create test struct with nested struct
	type Address struct {
		Street string `orm:"street"`
		City   string `orm:"city"`
	}

	type PersonWithAddress struct {
		ID      int     `orm:"id key"`
		Name    string  `orm:"name"`
		Address Address `orm:"address"`
	}

	// Get model for struct with nested struct
	nestedModel, err := GetEntityModel(&PersonWithAddress{})
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Check that the model has the expected fields
	fields := nestedModel.GetFields()
	if len(fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(fields))
	}

	// Check the nested struct field
	addressField := nestedModel.GetField("address")
	if addressField == nil {
		t.Errorf("GetField(address) returned nil")
	} else if !model.IsStructField(addressField) {
		t.Errorf("address field should be a struct type")
	}
}

// TestModelWithSliceFields tests models with slice fields
func TestModelWithSliceFields(t *testing.T) {
	// Create test struct with slice fields
	type WithSlices struct {
		ID     int      `orm:"id key"`
		Names  []string `orm:"names"`
		Scores []int    `orm:"scores"`
	}

	// Get model for struct with slices
	sliceModel, err := GetEntityModel(&WithSlices{})
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Check slice fields
	namesField := sliceModel.GetField("names")
	if namesField == nil {
		t.Errorf("GetField(names) returned nil")
	} else if !model.IsSliceField(namesField) {
		t.Errorf("names field should be a slice type")
	}

	scoresField := sliceModel.GetField("scores")
	if scoresField == nil {
		t.Errorf("GetField(scores) returned nil")
	} else if !model.IsSliceField(scoresField) {
		t.Errorf("scores field should be a slice type")
	}
}
