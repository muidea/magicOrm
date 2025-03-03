package local

import (
	"reflect"
	"testing"
	"time"
	
	"github.com/muidea/magicOrm/model"
)

// All these test structs represent different cases we need to test according
// to the readme.md conversion rules

// BasicTypes represents a struct with all basic data types
type BasicTypes struct {
	Id        int     `orm:"id key"` // Primary key is lowercase "id"
	IntValue  int     `orm:"intValue"`
	Int8Value int8    `orm:"int8Value"`
	Int16Value int16  `orm:"int16Value"`
	Int32Value int32  `orm:"int32Value"`
	Int64Value int64  `orm:"int64Value"`
	UintValue uint    `orm:"uintValue"`
	FloatValue float32 `orm:"floatValue"`
	Float64Value float64 `orm:"float64Value"`
	BoolValue bool    `orm:"boolValue"`
	StringValue string `orm:"stringValue"`
}

// TimeType represents a struct with a time.Time field
type TimeType struct {
	Id       int       `orm:"id key"`
	Created  time.Time `orm:"created"`
}

// NestedStruct represents a struct with nested struct fields
type NestedStruct struct {
	Id      int        `orm:"id key"`
	Basic   BasicTypes `orm:"basic"`
	TimeVal TimeType   `orm:"timeVal"`
}

// PointerTypes represents a struct with pointer types
type PointerTypes struct {
	Id          int         `orm:"id key"`
	IntPtr      *int        `orm:"intPtr"`
	StringPtr   *string     `orm:"stringPtr"`
	BasicPtr    *BasicTypes `orm:"basicPtr"`
}

// SliceTypes represents a struct with slice types
type SliceTypes struct {
	Id          int           `orm:"id key"`
	IntSlice    []int         `orm:"intSlice"`
	StringSlice []string      `orm:"stringSlice"`
	BasicSlice  []BasicTypes  `orm:"basicSlice"`
	TimeSlice   []time.Time   `orm:"timeSlice"`
}

// ComplexStruct represents a struct with a mix of all types
type ComplexStruct struct {
	Id         int           `orm:"id key"`
	Basic      BasicTypes    `orm:"basic"`
	Nested     NestedStruct  `orm:"nested"`
	TimeVal    time.Time     `orm:"timeVal"`
	PtrField   *BasicTypes   `orm:"ptrField"`
	SliceField []int         `orm:"sliceField"`
	PtrSlice   *[]string     `orm:"ptrSlice"`
}

// Invalid struct types for testing error cases
type NonStructType int

type NoFieldStruct struct{}

type NoPrimaryKeyStruct struct {
	Name string `orm:"name"`
	Age  int    `orm:"age"`
}

// ViewTagStruct represents a struct with view tag specifications
type ViewTagStruct struct {
	Id          int       `orm:"id key" view:"detail,lite"`
	Name        string    `orm:"name" view:"detail,lite"`
	Description string    `orm:"description" view:"detail"`
	InternalId  int       `orm:"internalId"`
	CreatedTime time.Time `orm:"createdTime" view:"lite"`
}

// ValueDeclareStruct represents a struct with different value declarations
type ValueDeclareStruct struct {
	Id        int       `orm:"id key"`
	UUIDField string    `orm:"uuidField,uuid"`
	SnowField int64     `orm:"snowField,snowFlake"`
	TimeField time.Time `orm:"timeField,dateTime"`
}

// TestStructToModelConversion tests rule 1: struct object to model.Model conversion
func TestStructToModelConversion(t *testing.T) {
	// Test BasicTypes struct conversion
	basicStruct := &BasicTypes{Id: 1, IntValue: 10, StringValue: "test"}
	
	// Get model from struct
 mdl, err := GetEntityModel(basicStruct)
	if err != nil {
		t.Fatalf("Failed to convert BasicTypes struct to model: %v", err)
	}
	
	// Test rule 1.1: struct name should become model name
	if mdl.GetName() != "BasicTypes" {
		t.Errorf("Rule 1.1 failed: Model name should be 'BasicTypes', got '%s'", mdl.GetName())
	}
	
	// Test rule 1.2: struct package path should become model package path
	if mdl.GetPkgPath() != reflect.TypeOf(BasicTypes{}).PkgPath() {
		t.Errorf("Rule 1.2 failed: Model package path should match struct package path")
	}
	
	// Test rule 1.3: PkgKey should be PkgPath + "/" + Name
	expectedPkgKey := mdl.GetPkgPath() + "/" + mdl.GetName()
	if mdl.GetPkgKey() != expectedPkgKey {
		t.Errorf("Rule 1.3 failed: Model PkgKey should be '%s', got '%s'", expectedPkgKey, mdl.GetPkgKey())
	}
}

// TestFieldConversion tests rule 2: struct fields to model.Field conversion
func TestFieldConversion(t *testing.T) {
	basicStruct := &BasicTypes{
		Id:         1,
		IntValue:   2,
		Int8Value:  3,
		Int16Value: 4,
		Int32Value: 5,
		Int64Value: 6,
		UintValue:  7,
		FloatValue: 8.0,
		Float64Value: 9.0,
		BoolValue:  true,
		StringValue: "test",
	}
	
 mdl, err := GetEntityModel(basicStruct)
	if err != nil {
		t.Fatalf("Failed to convert struct to model: %v", err)
	}
	
	fields := mdl.GetFields()
	t.Logf("Number of fields in model: %d", len(fields))
	for i, field := range fields {
		t.Logf("Field[%d]: Name=%s, Type=%v, IsPrimaryKey=%v", 
			i, field.GetName(), field.GetType().GetValue(), field.IsPrimaryKey())
		
		// Test rule 2.1: field names should match the ORM tag names
		switch field.GetName() {
		case "id":
			if !field.IsPrimaryKey() {
				t.Errorf("Rule 2.1 failed: Field 'id' should be marked as primary key")
			}
		case "intValue":
			// Check that it's an integer type using GetValue()
			if field.GetType().GetValue() != model.TypeIntegerValue {
				t.Errorf("Rule 2.1 failed: Field 'intValue' should be of type int, got %v", field.GetType().GetValue())
			}
		case "floatValue":
			if field.GetType().GetValue() != model.TypeFloatValue {
				t.Errorf("Rule 2.1 failed: Field 'floatValue' should be of type float, got %v", field.GetType().GetValue())
			}
		case "boolValue":
			if field.GetType().GetValue() != model.TypeBooleanValue {
				t.Errorf("Rule 2.1 failed: Field 'boolValue' should be of type bool, got %v", field.GetType().GetValue())
			}
		case "stringValue":
			if field.GetType().GetValue() != model.TypeStringValue {
				t.Errorf("Rule 2.1 failed: Field 'stringValue' should be of type string, got %v", field.GetType().GetValue())
			}
		}
	}
}

// TestNestedStructConversion tests rule 2.5: nested struct conversion
func TestNestedStructConversion(t *testing.T) {
	nestedStruct := &NestedStruct{
		Id: 1,
		Basic: BasicTypes{Id: 10, IntValue: 20, StringValue: "nested"},
		TimeVal: TimeType{Id: 100, Created: time.Now()},
	}
	
 mdl, err := GetEntityModel(nestedStruct)
	if err != nil {
		t.Fatalf("Failed to convert NestedStruct to model: %v", err)
	}
	
	// Debug - print all fields in the model
	fields := mdl.GetFields()
	t.Logf("Number of fields in model: %d", len(fields))
	for i, field := range fields {
		t.Logf("Field[%d]: Name=%s, Type=%v, IsStruct=%v", 
			i, field.GetName(), field.GetType().GetValue(), field.IsStruct())
		
		// Check nested struct fields
		switch field.GetName() {
		case "basic":
			if !field.IsStruct() {
				t.Errorf("Rule 2.5 failed: Field 'basic' should be marked as a struct")
			}
		case "timeVal":
			if !field.IsStruct() {
				t.Errorf("Rule 2.5 failed: Field 'timeVal' should be marked as a struct")
			}
		}
	}
}

// TestAllowedFieldTypes tests rule 2.6: allowed field types
func TestAllowedFieldTypes(t *testing.T) {
	// Test time.Time type
	timeStruct := &TimeType{Id: 1, Created: time.Now()}
 mdl, err := GetEntityModel(timeStruct)
	if err != nil {
		t.Errorf("Rule 2.6 failed: time.Time should be an allowed type: %v", err)
	}
	
	// Debug - print all fields in the model
	fields := mdl.GetFields()
	t.Logf("Number of fields in model: %d", len(fields))
	for i, field := range fields {
		t.Logf("Field[%d]: Name=%s, Type=%v", 
			i, field.GetName(), field.GetType().GetValue())
	}
	
	// Test pointer types
	intVal := 42
	strVal := "pointer"
	pointerStruct := &PointerTypes{
		Id: 1,
		IntPtr: &intVal,
		StringPtr: &strVal,
		BasicPtr: &BasicTypes{Id: 10},
	}
	
 mdl, err = GetEntityModel(pointerStruct)
	if err != nil {
		t.Errorf("Rule 2.6 failed: pointer types should be allowed: %v", err)
		return
	}
	
	// Debug - print all fields in the pointer model
	ptrFields := mdl.GetFields()
	t.Logf("Number of fields in pointer model: %d", len(ptrFields))
	for i, field := range ptrFields {
		t.Logf("PtrField[%d]: Name=%s, Type=%v, IsPtrType=%v", 
			i, field.GetName(), field.GetType().GetValue(), field.IsPtrType())
		
		// Check pointer types
		switch field.GetName() {
		case "intPtr":
			if !field.IsPtrType() {
				t.Errorf("Rule 2.6 failed: Field 'intPtr' should be marked as a pointer type")
			}
		case "stringPtr":
			if !field.IsPtrType() {
				t.Errorf("Rule 2.6 failed: Field 'stringPtr' should be marked as a pointer type")
			}
		case "basicPtr":
			if !field.IsPtrType() {
				t.Errorf("Rule 2.6 failed: Field 'basicPtr' should be marked as a pointer type")
			}
		}
	}
}

// TestSliceFieldTypes tests rule 2.7: allowed slice field types
func TestSliceFieldTypes(t *testing.T) {
	// Test slice types
	sliceStruct := &SliceTypes{
		Id: 1,
		IntSlice: []int{1, 2, 3},
		StringSlice: []string{"a", "b", "c"},
		BasicSlice: []BasicTypes{{Id: 10}, {Id: 20}},
		TimeSlice: []time.Time{time.Now(), time.Now()},
	}
	
 mdl, err := GetEntityModel(sliceStruct)
	if err != nil {
		t.Errorf("Rule 2.7 failed: slice types should be allowed: %v", err)
		return
	}
	
	// Debug - print all fields in the model
	fields := mdl.GetFields()
	t.Logf("Number of fields in model: %d", len(fields))
	for i, field := range fields {
		t.Logf("Field[%d]: Name=%s, Type=%v, IsSlice=%v", 
			i, field.GetName(), field.GetType().GetValue(), field.IsSlice())
		
		// Check slice types
		switch field.GetName() {
		case "intSlice":
			if !field.IsSlice() {
				t.Errorf("Rule 2.7 failed: Field 'intSlice' should be marked as a slice type")
			}
		case "stringSlice":
			if !field.IsSlice() {
				t.Errorf("Rule 2.7 failed: Field 'stringSlice' should be marked as a slice type")
			}
		case "basicSlice":
			if !field.IsSlice() {
				t.Errorf("Rule 2.7 failed: Field 'basicSlice' should be marked as a slice type")
			}
		case "timeSlice":
			if !field.IsSlice() {
				t.Errorf("Rule 2.7 failed: Field 'timeSlice' should be marked as a slice type")
			}
		}
	}
}

// TestComplexStructConversion tests all rules together with a complex struct
func TestComplexStructConversion(t *testing.T) {
	// Create a string slice for the pointer slice field
	strSlice := []string{"item1", "item2"}
	
	complexStruct := &ComplexStruct{
		Id: 1,
		Basic: BasicTypes{Id: 10, IntValue: 20, StringValue: "basic"},
		Nested: NestedStruct{
			Id: 100,
			Basic: BasicTypes{Id: 101, StringValue: "nested-basic"},
			TimeVal: TimeType{Id: 102, Created: time.Now()},
		},
		TimeVal: time.Now(),
		PtrField: &BasicTypes{Id: 200, BoolValue: true},
		SliceField: []int{1, 2, 3},
		PtrSlice: &strSlice,
	}
	
 mdl, err := GetEntityModel(complexStruct)
	if err != nil {
		t.Fatalf("Failed to convert ComplexStruct to model: %v", err)
	}
	
	// Debug - print all fields in the model
	fields := mdl.GetFields()
	t.Logf("Number of fields in model: %d", len(fields))
	for i, field := range fields {
		t.Logf("Field[%d]: Name=%s, Type=%v, IsStruct=%v, IsSlice=%v, IsPtrType=%v", 
			i, field.GetName(), field.GetType().GetValue(), field.IsStruct(), 
			field.IsSlice(), field.IsPtrType())
		
		// Comprehensive test - verify all field types are correctly identified
		switch field.GetName() {
		case "id":
			if field.IsStruct() || field.IsSlice() || field.IsPtrType() {
				t.Errorf("Field 'id' should not be a struct, slice, or pointer type")
			}
		case "basic":
			if !field.IsStruct() {
				t.Errorf("Field 'basic' should be marked as a struct")
			}
		case "nested":
			if !field.IsStruct() {
				t.Errorf("Field 'nested' should be marked as a struct")
			}
		case "timeVal":
			if field.IsStruct() || field.IsSlice() || field.IsPtrType() {
				t.Errorf("Field 'timeVal' should not be a struct, slice, or pointer type")
			}
		case "ptrField":
			if !field.IsStruct() || !field.IsPtrType() {
				t.Errorf("Field 'ptrField' should be marked as a struct and pointer type")
			}
		case "sliceField":
			if !field.IsSlice() {
				t.Errorf("Field 'sliceField' should be marked as a slice type")
			}
		case "ptrSlice":
			if !field.IsSlice() || !field.IsPtrType() {
				t.Errorf("Field 'ptrSlice' should be marked as a slice and pointer type")
			}
		}
	}
	
	// Test round-trip conversion
	// Convert model back to struct and verify properties are preserved
	roundTripped := mdl.Interface(true, "origin")
	convertedStruct, ok := roundTripped.(*ComplexStruct)
	if !ok {
		t.Errorf("Model could not be converted back to ComplexStruct, got type %T", roundTripped)
		return
	}
	
	// Check some values to make sure they were preserved
	if convertedStruct.Id != complexStruct.Id {
		t.Errorf("After round-trip conversion, Id should be %d, got %d", 
			complexStruct.Id, convertedStruct.Id)
	}
	
	if convertedStruct.Basic.StringValue != complexStruct.Basic.StringValue {
		t.Errorf("After round-trip conversion, Basic.StringValue should be '%s', got '%s'", 
			complexStruct.Basic.StringValue, convertedStruct.Basic.StringValue)
	}
}

// TestEntityModelErrorCases tests error cases for GetEntityModel
func TestEntityModelErrorCases(t *testing.T) {
	// Test case 1: Non-struct type
	var nonStruct NonStructType = 42
 mdl, err := GetEntityModel(nonStruct)
	if err == nil {
		t.Errorf("GetEntityModel should return error for non-struct type")
	}
	
	// Test case 2: Nil value
 mdl, err = GetEntityModel(nil)
	if err == nil {
		t.Errorf("GetEntityModel should return error for nil value")
	}
	
	// Test case 3: Struct with no fields
	emptyStruct := &NoFieldStruct{}
 mdl, err = GetEntityModel(emptyStruct)
	if err == nil {
		t.Errorf("GetEntityModel should return error for struct with no fields")
	}
	
	// Test case 4: Struct with no primary key
	noPkStruct := &NoPrimaryKeyStruct{Name: "test", Age: 30}
 mdl, err = GetEntityModel(noPkStruct)
	if err == nil {
		t.Errorf("GetEntityModel should return error for struct with no primary key")
	}
	
	// Test case 5: Value instead of pointer
	directStruct := BasicTypes{Id: 1}
 mdl, err = GetEntityModel(&directStruct)
	if err != nil {
		t.Errorf("GetEntityModel should handle non-pointer struct types, got error: %v", err)
	}
	
	if mdl == nil {
		t.Errorf("GetEntityModel should return a valid model for non-pointer struct types")
	} else {
		// Verify model is correct
		if mdl.GetName() != "BasicTypes" {
			t.Errorf("Expected model name to be 'BasicTypes', got '%s'", mdl.GetName())
		}
	}
}

// TestViewTagSpecification tests rule 3.2: field tags with view specifications
func TestViewTagSpecification(t *testing.T) {
	viewStruct := ViewTagStruct{
		Id:          1,
		Name:        "Test Name",
		Description: "Test Description",
		InternalId:  42,
		CreatedTime: time.Now(),
	}
	
 mdl, err := GetEntityModel(&viewStruct)
	if err != nil {
		t.Fatalf("Failed to convert struct to model: %v", err)
	}
	
	// Test that view tags are correctly parsed
	fields := mdl.GetFields()
	
	// Debug - print all fields with their view specifications
	t.Logf("Number of fields in model: %d", len(fields))
	for i, field := range fields {
		t.Logf("Field[%d]: Name=%s, Origin=%v, Detail=%v, Lite=%v", 
			i, field.GetName(), 
			field.GetSpec().EnableView("origin"),
			field.GetSpec().EnableView("detail"),
			field.GetSpec().EnableView("lite"))
		
		// Check view specifications based on field name
		switch field.GetName() {
		case "id":
			if !field.GetSpec().EnableView("detail") || !field.GetSpec().EnableView("lite") {
				t.Errorf("Field 'id' should support detail and lite views")
			}
		case "name":
			if !field.GetSpec().EnableView("detail") || !field.GetSpec().EnableView("lite") {
				t.Errorf("Field 'name' should support detail and lite views")
			}
		case "description":
			if !field.GetSpec().EnableView("detail") || field.GetSpec().EnableView("lite") {
				t.Errorf("Field 'description' should support only detail view, not lite view")
			}
		case "internalId":
			if field.GetSpec().EnableView("detail") || field.GetSpec().EnableView("lite") {
				t.Errorf("Field 'internalId' should not support any views")
			}
		case "createdTime":
			if field.GetSpec().EnableView("detail") || !field.GetSpec().EnableView("lite") {
				t.Errorf("Field 'createdTime' should support only lite view, not detail view")
			}
		}
	}
	
	// Test Interface function with different views
	// OriginView - should include all fields
	originObj := mdl.Interface(false, "origin")
	_, isPointer := originObj.(*ViewTagStruct)
	if isPointer {
		t.Errorf("Interface(false, OriginView) should return a value, not a pointer")
	}
	
	// Test with different view specifications
	detailObj := mdl.Interface(false, "detail").(ViewTagStruct)
	if detailObj.Id != 1 || detailObj.Name != "Test Name" || detailObj.Description != "Test Description" {
		t.Errorf("Interface with DetailView should return fields tagged with detail")
	}
	
	liteObj := mdl.Interface(false, "lite").(ViewTagStruct)
	if liteObj.Id != 1 || liteObj.Name != "Test Name" {
		t.Errorf("Interface with LiteView should return fields tagged with lite")
	}
}

// Test the value declaration fields in ValueDeclareStruct
func TestValueDeclarations(t *testing.T) {
	// Create a struct with values
	declareStruct := ValueDeclareStruct{
		Id:        1,
		UUIDField: "test-uuid-value",
		SnowField: 12345678901234,
		TimeField: time.Now(),
	}
	
	// Print struct for debugging
	t.Logf("Test struct: %+v", declareStruct)
	
	// Convert to model
 mdl, err := GetEntityModel(&declareStruct)
	if err != nil {
		t.Fatalf("Failed to convert struct to model: %v", err)
	}
	
	// Debug - print all fields in the model
	fields := mdl.GetFields()
	t.Logf("Number of fields in model: %d", len(fields))
	for i, field := range fields {
		t.Logf("Field[%d]: Name=%s, Type=%v, ValueDeclare=%v", 
			i, field.GetName(), field.GetType().GetName(), field.GetSpec().GetValueDeclare())
		
		// Check value declarations based on field name
		switch field.GetName() {
		case "uuidField":
			if field.GetSpec().GetValueDeclare() != 2 { // UUID = 2
				t.Errorf("Field 'uuidField' should have UUID value declaration, got %d", field.GetSpec().GetValueDeclare())
			}
		case "snowField":
			if field.GetSpec().GetValueDeclare() != 3 { // SnowFlake = 3
				t.Errorf("Field 'snowField' should have SnowFlake value declaration, got %d", field.GetSpec().GetValueDeclare())
			}
		case "timeField":
			if field.GetSpec().GetValueDeclare() != 4 { // DateTime = 4
				t.Errorf("Field 'timeField' should have DateTime value declaration, got %d", field.GetSpec().GetValueDeclare())
			}
		}
	}
}

// TestModelCopyFunction tests rule 3.4: Model Copy function
func TestModelCopyFunction(t *testing.T) {
	// Create a basic struct with values
	basicStruct := BasicTypes{
		Id:        1,
		IntValue:  42,
		FloatValue: 3.14,
		BoolValue: true,
		StringValue: "test string",
	}
	
	// Convert to model
 mdl, err := GetEntityModel(&basicStruct)
	if err != nil {
		t.Fatalf("Failed to convert struct to model: %v", err)
	}
	
	// Test Copy with reset=false (should preserve values)
	copiedMdl := mdl.Copy(false)
	
	// Convert back to struct and verify values are preserved
	copiedStruct := copiedMdl.Interface(false, "origin").(BasicTypes)
	
	if copiedStruct.Id != 1 || 
	   copiedStruct.IntValue != 42 ||
	   copiedStruct.FloatValue != 3.14 ||
	   copiedStruct.BoolValue != true ||
	   copiedStruct.StringValue != "test string" {
		t.Errorf("Copy(false) should preserve all values in the model")
	}
	
	// Test Copy with reset=true (should reset all values to defaults)
	resetMdl := mdl.Copy(true)
	
	// Convert back to struct and verify values are reset
	resetStruct := resetMdl.Interface(false, "origin").(BasicTypes)
	
	if resetStruct.Id != 0 || 
	   resetStruct.IntValue != 0 ||
	   resetStruct.FloatValue != 0 ||
	   resetStruct.BoolValue != false ||
	   resetStruct.StringValue != "" {
		t.Errorf("Copy(true) should reset all values to defaults, got: %+v", resetStruct)
	}
}

// TestInterfaceFunction tests rule 3.3: Model Interface function
func TestInterfaceFunction(t *testing.T) {
	// Create a struct with values
	testStruct := ViewTagStruct{
		Id:          123,
		Name:        "Test Object",
		Description: "This is a test object",
		InternalId:  456,
		CreatedTime: time.Now(),
	}
	
	// Convert to model
 mdl, err := GetEntityModel(&testStruct)
	if err != nil {
		t.Fatalf("Failed to convert struct to model: %v", err)
	}
	
	// Test Interface with ptrValue=false, should return a value
	valueObj := mdl.Interface(false, "origin")
	_, isPointer := valueObj.(*ViewTagStruct)
	if isPointer {
		t.Errorf("Interface(false, OriginView) should return a value, not a pointer")
	}
	
	// Test Interface with ptrValue=true, should return a pointer
	ptrObj := mdl.Interface(true, "origin")
	_, isPointer = ptrObj.(*ViewTagStruct)
	if !isPointer {
		t.Errorf("Interface(true, OriginView) should return a pointer")
	}
	
	// Test with different view specifications
	detailObj := mdl.Interface(false, "detail").(ViewTagStruct)
	if detailObj.Id != 123 || detailObj.Name != "Test Object" || detailObj.Description != "This is a test object" {
		t.Errorf("Interface with DetailView should return fields tagged with detail")
	}
	
	liteObj := mdl.Interface(false, "lite").(ViewTagStruct)
	if liteObj.Id != 123 || liteObj.Name != "Test Object" {
		t.Errorf("Interface with LiteView should return fields tagged with lite")
	}
}
