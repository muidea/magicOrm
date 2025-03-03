package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
)

// TestProviderStruct is a test struct for provider testing
type TestProviderStruct struct {
	ID         int       `orm:"id key"`
	Name       string    `orm:"name"`
	CreateTime time.Time `orm:"createTime"`
}

// TestNestedStruct is a struct containing another struct for testing nested models
type TestNestedStruct struct {
	ID     int                `orm:"id key"`
	Info   TestProviderStruct `orm:"info"`
	Active bool               `orm:"active"`
}

func TestGetModelFilter(t *testing.T) {
	// Create a test entity
	entity := TestProviderStruct{
		ID:         1,
		Name:       "test",
		CreateTime: time.Now(),
	}

	// Get model for the entity
	objModel, err := GetEntityModel(entity)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test GetModelFilter with default view
	filter, err := GetModelFilter(objModel, model.OriginView)
	if err != nil {
		t.Errorf("GetModelFilter failed: %s", err.Error())
		return
	}

	if filter == nil {
		t.Errorf("GetModelFilter returned nil filter")
		return
	}

	// Test GetModelFilter with detail view
	filter, err = GetModelFilter(objModel, model.DetailView)
	if err != nil {
		t.Errorf("GetModelFilter with DetailView failed: %s", err.Error())
		return
	}

	if filter == nil {
		t.Errorf("GetModelFilter with DetailView returned nil filter")
		return
	}
}

func TestSetModelValue(t *testing.T) {
	// Create a test entity
	entity := TestProviderStruct{
		ID:         1,
		Name:       "original",
		CreateTime: time.Now(),
	}

	// Get model for the entity
	objModel, err := GetEntityModel(entity)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Create a new value to set
	newEntity := TestProviderStruct{
		ID:         1,
		Name:       "updated",
		CreateTime: time.Now(),
	}

	entityValue, err := GetEntityValue(newEntity)
	if err != nil {
		t.Errorf("GetEntityValue failed: %s", err.Error())
		return
	}

	// Test SetModelValue
	updatedModel, err := SetModelValue(objModel, entityValue)
	if err != nil {
		t.Errorf("SetModelValue failed: %s", err.Error())
		return
	}

	if updatedModel == nil {
		t.Errorf("SetModelValue returned nil model")
		return
	}

	// Verify the updated value
	nameField := updatedModel.GetField("name")
	if nameField == nil {
		t.Errorf("GetField failed, name field not found")
		return
	}

	nameValue := nameField.GetValue()
	fieldValue := nameValue.Get().(reflect.Value)
	if fieldValue.String() != "updated" {
		t.Errorf("SetModelValue failed, expected name: %s, got: %s", "updated", fieldValue.String())
	}
}

func TestAppendSliceValue(t *testing.T) {
	// Create a slice value
	intSlice := []int{1, 2, 3}
	sliceReflectVal := reflect.ValueOf(intSlice)
	sliceValue := NewValue(sliceReflectVal)

	// Create a value to append
	newInt := 4
	newIntVal := reflect.ValueOf(newInt)
	newValue := NewValue(newIntVal)

	// Test AppendSliceValue
	resultValue, err := AppendSliceValue(sliceValue, newValue)
	if err != nil {
		t.Errorf("AppendSliceValue failed: %s", err.Error())
		return
	}

	// Verify the result
	resultSlice := resultValue.Get().(reflect.Value)
	if resultSlice.Len() != 4 {
		t.Errorf("AppendSliceValue failed, expected length: 4, got: %d", resultSlice.Len())
		return
	}

	lastElement := resultSlice.Index(3).Interface().(int)
	if lastElement != 4 {
		t.Errorf("AppendSliceValue failed, expected last element: 4, got: %d", lastElement)
	}

	// Test with non-slice value
	nonSliceValue := NewValue(reflect.ValueOf(1))
	_, err = AppendSliceValue(nonSliceValue, newValue)
	if err == nil {
		t.Errorf("AppendSliceValue should fail with non-slice value")
	}

	// Test with incompatible element type
	_, err = AppendSliceValue(sliceValue, NewValue(reflect.ValueOf("string")))
	if err == nil {
		t.Errorf("AppendSliceValue should fail with incompatible element type")
	}
}

func TestEncodeDecodeWithNestedStructs(t *testing.T) {
	// Create a test nested entity
	entity := TestNestedStruct{
		ID:     1,
		Info:   TestProviderStruct{ID: 2, Name: "nested", CreateTime: time.Now()},
		Active: true,
	}

	// Get type and value
	entityType, err := GetEntityType(entity)
	if err != nil {
		t.Errorf("GetEntityType failed: %s", err.Error())
		return
	}

	entityValue, err := GetEntityValue(entity)
	if err != nil {
		t.Errorf("GetEntityValue failed: %s", err.Error())
		return
	}

	// Create a simple mock cache for testing
	mockCache := model.NewCache()

	// Get the model for our test entity
	entityModel, err := GetEntityModel(entity)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Add model to mock cache
	mockCache.Put(entityModel.GetPkgKey(), entityModel)

	// Also add model for the nested type
	nestedModel, err := GetEntityModel(TestProviderStruct{})
	if err != nil {
		t.Errorf("GetEntityModel for nested type failed: %s", err.Error())
		return
	}
	mockCache.Put(nestedModel.GetPkgKey(), nestedModel)

	// Test EncodeValue with nested struct
	rawVal, err := EncodeValue(entityValue, entityType, mockCache)
	if err != nil {
		t.Errorf("EncodeValue failed: %s", err.Error())
		return
	}

	if rawVal == nil {
		t.Errorf("EncodeValue returned nil value")
		return
	}

	// Test DecodeValue with nested struct
	decodedValue, err := DecodeValue(rawVal, entityType, mockCache)
	if err != nil {
		t.Errorf("DecodeValue failed: %s", err.Error())
		return
	}

	if decodedValue == nil {
		t.Errorf("DecodeValue returned nil value")
		return
	}

	// Verify the decoded value is the same as original
	decodedEntity := decodedValue.Interface().Value().(TestNestedStruct)

	if decodedEntity.ID != entity.ID {
		t.Errorf("ID mismatch: expected %d, got %d", entity.ID, decodedEntity.ID)
	}
}

func TestEncodeDecodeSlice(t *testing.T) {
	// Create a slice of test entities
	entities := []TestProviderStruct{
		{ID: 1, Name: "first", CreateTime: time.Now()},
		{ID: 2, Name: "second", CreateTime: time.Now()},
		{ID: 3, Name: "third", CreateTime: time.Now()},
	}

	// Get type and value
	entityType, err := GetEntityType(entities)
	if err != nil {
		t.Errorf("GetEntityType failed: %s", err.Error())
		return
	}

	entityValue, err := GetEntityValue(entities)
	if err != nil {
		t.Errorf("GetEntityValue failed: %s", err.Error())
		return
	}

	// Create a simple mock cache for testing
	mockCache := model.NewCache()

	// Get the model for our element type
	elementModel, err := GetEntityModel(TestProviderStruct{})
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	mockCache.Put(elementModel.GetPkgKey(), elementModel)
	// Test EncodeValue with slice
	rawVal, err := EncodeValue(entityValue, entityType, mockCache)
	if err != nil {
		t.Errorf("EncodeValue failed: %s", err.Error())
		return
	}

	if rawVal == nil {
		t.Errorf("EncodeValue returned nil value")
		return
	}

	// Test DecodeValue with slice
	decodedValue, err := DecodeValue(rawVal, entityType, mockCache)
	if err != nil {
		t.Errorf("DecodeValue failed: %s", err.Error())
		return
	}

	if decodedValue == nil {
		t.Errorf("DecodeValue returned nil value")
		return
	}

	// Verify the decoded value is the same as original
	decodedSlice := decodedValue.Interface().Value().([]TestProviderStruct)

	if len(decodedSlice) != len(entities) {
		t.Errorf("Slice length mismatch: expected %d, got %d", len(entities), len(decodedSlice))
		return
	}

	for i, entity := range entities {
		if decodedSlice[i].ID != entity.ID {
			t.Errorf("ID mismatch at index %d: expected %d, got %d", i, entity.ID, decodedSlice[i].ID)
		}
	}
}

func TestGetNewValue(t *testing.T) {
	testCases := []struct {
		name           string
		valueDeclare   model.ValueDeclare
		expectedIsZero bool
	}{
		{"Customer", model.Customer, true},
		{"AutoIncrement", model.AutoIncrement, true},
		{"UUID", model.UUID, false},                // UUID generates a new value, so not zero
		{"SnowFlake", model.SnowFlake, false},      // SnowFlake generates a new value, so not zero
		{"DateTime", model.DateTime, false},        // DateTime gets current time, so not zero
		{"Invalid", model.ValueDeclare(999), true}, // Test with invalid value declare
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := GetNewValue(tc.valueDeclare)

			if value == nil {
				t.Errorf("GetNewValue returned nil for %s", tc.name)
				return
			}

			if value.IsZero() != tc.expectedIsZero {
				t.Errorf("GetNewValue for %s, expected IsZero(): %v, got: %v",
					tc.name, tc.expectedIsZero, value.IsZero())
			}
		})
	}
}
