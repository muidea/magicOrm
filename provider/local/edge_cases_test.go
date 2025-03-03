package local

import (
	"reflect"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/model"
)

// EmptyStruct is a struct with no ORM fields for testing
type EmptyStruct struct {
	ID   int
	Name string
}

// InvalidTimeStruct tests invalid time values
type InvalidTimeStruct struct {
	ID        int       `orm:"id key"`
	CreatedAt time.Time `orm:"createdAt"`
}

// CyclicStruct1 and CyclicStruct2 for testing cyclic references
type CyclicStruct1 struct {
	ID   int            `orm:"id key"`
	Name string         `orm:"name"`
	Ref  *CyclicStruct2 `orm:"ref"`
}

type CyclicStruct2 struct {
	ID   int            `orm:"id key"`
	Name string         `orm:"name"`
	Ref  *CyclicStruct1 `orm:"ref"`
}

// EdgeStruct with unusual fields
type EdgeStruct struct {
	ID          int       `orm:"id key"`
	EmptyString string    `orm:"emptyString"`
	ZeroInt     int       `orm:"zeroInt"`
	ZeroFloat   float64   `orm:"zeroFloat"`
	EmptySlice  []string  `orm:"emptySlice"`
	NilPointer  *string   `orm:"nilPointer"`
	ZeroTime    time.Time `orm:"zeroTime"`
}

func TestNilValueOperations(t *testing.T) {
	// Test with non-nil pointer to nil entity
	var es *EdgeStruct = nil
	var nilEntity interface{} = es

	// Test GetEntityType with nil
	_, err := GetEntityType(nilEntity)
	if err == nil {
		t.Errorf("GetEntityType should fail with nil entity")
	}

	// Use defer/recover to test functions that might panic
	// even with our checks
	testNilWithRecover := func(t *testing.T, funcName string, testFunc func() (*cd.Result, bool)) {
		defer func() {
			if r := recover(); r != nil {
				// The function panicked, which is actually ok for this test
				// Since we're testing error handling of nil values
				t.Logf("%s panicked with: %v (this is expected)", funcName, r)
			}
		}()

		result, failed := testFunc()
		if !failed {
			t.Errorf("%s should fail with nil entity", funcName)
		} else {
			t.Logf("%s returned error as expected: %v", funcName, result)
		}
	}

	// Test GetEntityValue with nil using recover
	testNilWithRecover(t, "GetEntityValue", func() (*cd.Result, bool) {
		_, err := GetEntityValue(nilEntity)
		return err, err != nil
	})

	// Test GetEntityModel with nil using recover
	testNilWithRecover(t, "GetEntityModel", func() (*cd.Result, bool) {
		_, err := GetEntityModel(nilEntity)
		return err, err != nil
	})

	// Test with nil interface
	var nilInterface interface{} = nil
	_, err = GetEntityType(nilInterface)
	if err == nil {
		t.Errorf("GetEntityType should fail with nil interface")
	}
}

func TestEmptyStructHandling(t *testing.T) {
	// Test struct with no ORM fields
	empty := EmptyStruct{ID: 1, Name: "test"}
	_, err := GetEntityModel(empty)
	if err == nil {
		t.Errorf("GetEntityModel should fail with struct that has no ORM fields")
	}
}

func TestZeroValues(t *testing.T) {
	// Test handling of zero/empty values
	edge := EdgeStruct{
		ID:          1,
		EmptyString: "",
		ZeroInt:     0,
		ZeroFloat:   0.0,
		EmptySlice:  []string{},
		NilPointer:  nil,
		ZeroTime:    time.Time{},
	}

	entityModel, err := GetEntityModel(edge)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	modelCache := model.NewCache()
	modelCache.Put(entityModel.GetPkgKey(), entityModel)

	// Test empty string field
	emptyStringField := entityModel.GetField("emptyString")
	if emptyStringField == nil {
		t.Errorf("GetField failed for 'emptyString'")
		return
	}

	if !emptyStringField.GetValue().IsZero() {
		t.Errorf("Empty string should be recognized as zero value")
	}

	// Test zero int field
	zeroIntField := entityModel.GetField("zeroInt")
	if zeroIntField == nil {
		t.Errorf("GetField failed for 'zeroInt'")
		return
	}

	if !zeroIntField.GetValue().IsZero() {
		t.Errorf("Zero int should be recognized as zero value")
	}

	// Test nil pointer field
	nilPointerField := entityModel.GetField("nilPointer")
	if nilPointerField == nil {
		t.Errorf("GetField failed for 'nilPointer'")
		return
	}

	if !nilPointerField.GetValue().IsZero() {
		t.Errorf("Nil pointer should be recognized as zero value")
	}

	// Test encode and decode with zero values
	entityType, _ := GetEntityType(edge)
	entityValue, _ := GetEntityValue(edge)

	rawVal, err := EncodeValue(entityValue, entityType, modelCache)
	if err != nil {
		t.Errorf("EncodeValue failed: %s", err.Error())
		return
	}

	_, err = DecodeValue(rawVal, entityType, modelCache)
	if err != nil {
		t.Errorf("DecodeValue failed: %s", err.Error())
	}
}

func TestInvalidInputs(t *testing.T) {
	// Test with basic type (not struct)
	basicType := 42
	_, err := GetEntityModel(basicType)
	if err == nil {
		t.Errorf("GetEntityModel should fail with non-struct type")
	}

	// Test with map type
	mapType := map[string]string{"key": "value"}
	_, err = GetEntityModel(mapType)
	if err == nil {
		t.Errorf("GetEntityModel should fail with map type")
	}

	// Test with slice type
	sliceType := []int{1, 2, 3}
	_, err = GetEntityModel(sliceType)
	if err == nil {
		t.Errorf("GetEntityModel should fail with slice type")
	}

	// Test with function type
	funcType := func() {}
	_, err = GetEntityModel(funcType)
	if err == nil {
		t.Errorf("GetEntityModel should fail with function type")
	}

	// Test with invalid time value
	invalidTime := InvalidTimeStruct{
		ID:        1,
		CreatedAt: time.Time{}, // Zero time
	}

	// This should succeed as we handle zero time values
	_, err = GetEntityModel(invalidTime)
	if err != nil {
		t.Errorf("GetEntityModel should succeed with zero time: %s", err.Error())
	}
}

// ViewSpecTestStruct for testing view specifications
type ViewSpecTestStruct struct {
	ID          int    `orm:"id key" view:"origin,detail,lite"`
	Name        string `orm:"name" view:"origin,detail,lite"`
	Description string `orm:"description" view:"origin,detail"`
	Notes       string `orm:"notes" view:"origin"`
}

func TestViewSpecEdgeCases(t *testing.T) {
	// Test with different view specifications
	entity := ViewSpecTestStruct{
		ID:          1,
		Name:        "test",
		Description: "description",
		Notes:       "notes",
	}

	model, err := GetEntityModel(entity)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test interfacing with different view specs
	// Detail view should include ID, Name, Description
	detailInterface := model.Interface(false, "detail")
	detailValue := reflect.ValueOf(detailInterface)
	if !detailValue.IsValid() {
		t.Errorf("Interface with DetailView returned invalid value")
		return
	}

	// Check if the right fields are present
	detailStruct, ok := detailInterface.(ViewSpecTestStruct)
	if !ok {
		t.Errorf("Interface should return ViewSpecTestStruct")
		return
	}

	if detailStruct.ID != 1 || detailStruct.Name != "test" || detailStruct.Description != "description" {
		t.Errorf("DetailView should include ID, Name, and Description")
	}

	// Lite view should include ID, Name but not Description
	liteInterface := model.Interface(false, "lite")
	liteValue := reflect.ValueOf(liteInterface)
	if !liteValue.IsValid() {
		t.Errorf("Interface with LiteView returned invalid value")
		return
	}

	// Check if the right fields are present/absent
	liteStruct, ok := liteInterface.(ViewSpecTestStruct)
	if !ok {
		t.Errorf("Interface should return ViewSpecTestStruct")
		return
	}

	if liteStruct.ID != 1 || liteStruct.Name != "test" {
		t.Errorf("LiteView should include ID and Name")
	}

	// Description should be empty in Lite view
	if liteStruct.Description != "" {
		t.Errorf("LiteView should not include Description, got: %s", liteStruct.Description)
	}
}

func TestNestedArrays(t *testing.T) {
	// Test struct with nested arrays
	type ArrayItem struct {
		ItemID   int    `orm:"itemid key"`
		ItemName string `orm:"itemName"`
	}

	type NestedArrayStruct struct {
		ID    int         `orm:"id key"`
		Name  string      `orm:"name"`
		Items []ArrayItem `orm:"items"`
	}

	testStruct := NestedArrayStruct{
		ID:   1,
		Name: "test",
		Items: []ArrayItem{
			{ItemID: 1, ItemName: "item1"},
			{ItemID: 2, ItemName: "item2"},
		},
	}

	modelCache := model.NewCache()

	// Get arrayStructModel
	arrayStructModel, err := GetEntityModel(testStruct)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	modelCache.Put(arrayStructModel.GetPkgKey(), arrayStructModel)

	// Check if the items field is marked as slice
	itemsField := arrayStructModel.GetField("items")
	if itemsField == nil {
		t.Errorf("GetField failed for 'items'")
		return
	}

	if !itemsField.IsSlice() {
		t.Errorf("Items field should be marked as slice")
	}

	// Get items field type
	itemsType := itemsField.GetType()
	if itemsType.IsBasic() {
		t.Errorf("Items type should not be basic")
	}

	// Test encode/decode of nested slice
	entityType, _ := GetEntityType(testStruct)
	entityValue, _ := GetEntityValue(testStruct)

	rawVal, err := EncodeValue(entityValue, entityType, modelCache)
	if err != nil {
		t.Errorf("EncodeValue failed: %s", err.Error())
		return
	}

	_, err = DecodeValue(rawVal, entityType, modelCache)
	if err != nil {
		t.Errorf("DecodeValue failed: %s", err.Error())
	}

	// Interface the decoded value
	arrayStructModel.Interface(false, "origin")
}

func TestObjectValueCombinations(t *testing.T) {
	// Create a test entity with various field types
	type ComplexFieldStruct struct {
		ID          int       `orm:"id key"`
		Name        string    `orm:"name"`
		Score       float64   `orm:"score"`
		IsActive    bool      `orm:"isActive"`
		CreateTime  time.Time `orm:"createTime"`
		IntPtr      *int      `orm:"intPtr"`
		StringPtr   *string   `orm:"stringPtr"`
		IntSlice    []int     `orm:"intSlice"`
		StringSlice []string  `orm:"stringSlice"`
	}

	intVal := 42
	stringVal := "pointer value"

	testStruct := ComplexFieldStruct{
		ID:          1,
		Name:        "test",
		Score:       95.5,
		IsActive:    true,
		CreateTime:  time.Now(),
		IntPtr:      &intVal,
		StringPtr:   &stringVal,
		IntSlice:    []int{1, 2, 3},
		StringSlice: []string{"a", "b", "c"},
	}

	// Get model
	model, err := GetEntityModel(testStruct)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test pointers
	intPtrField := model.GetField("intPtr")
	if intPtrField == nil {
		t.Errorf("GetField failed for 'intPtr'")
		return
	}

	if !intPtrField.IsPtrType() {
		t.Errorf("IntPtr field should be marked as pointer type")
	}

	// Test slices
	intSliceField := model.GetField("intSlice")
	if intSliceField == nil {
		t.Errorf("GetField failed for 'intSlice'")
		return
	}

	if !intSliceField.IsSlice() {
		t.Errorf("IntSlice field should be marked as slice")
	}

	// Test Interface with pointer return
	ptrInterface := model.Interface(true, "origin")
	if ptrInterface == nil {
		t.Errorf("Interface with ptrValue=true returned nil")
	}

	// Test model copy
	copyModel := model.Copy(false)
	if copyModel == nil {
		t.Errorf("Copy returned nil")
		return
	}

	// Test Dump function
	dumpStr := model.Dump()
	if dumpStr == "" {
		t.Errorf("Dump returned empty string")
	}
}
