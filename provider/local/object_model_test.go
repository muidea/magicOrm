package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
)

// ComplexFieldModelStruct represents a struct with fields of various types 
// for testing object model implementation
type ComplexFieldModelStruct struct {
	ID            int       `orm:"id key auto" view:"detail,lite"`
	Name          string    `orm:"name" view:"detail,lite"`
	Score         float64   `orm:"score" view:"detail"`
	IsActive      bool      `orm:"isActive" view:"detail,lite"`
	CreatedTime   time.Time `orm:"createdTime" view:"detail"`
	IntPointer    *int      `orm:"intPointer" view:"detail"`
	StringPointer *string   `orm:"stringPointer" view:"detail"`
	IntSlice      []int     `orm:"intSlice" view:"detail"`
	StringSlice   []string  `orm:"stringSlice" view:"detail"`
}

// NestedModelStruct is a struct that contains a ComplexFieldModelStruct
type NestedModelStruct struct {
	ID       int                   `orm:"id key auto" view:"detail,lite"`
	Name     string                `orm:"name" view:"detail,lite"`
	Complex  ComplexFieldModelStruct `orm:"complex" view:"detail"`
	ComplexP *ComplexFieldModelStruct `orm:"complexP" view:"detail"`
}

func TestObjectGetters(t *testing.T) {
	var entity ComplexFieldModelStruct
	entityType := reflect.TypeOf(entity)
	
	objModel, err := getTypeModel(entityType)
	if err != nil {
		t.Errorf("getTypeModel failed: %s", err.Error())
		return
	}
	
	// Test GetName
	if objModel.GetName() != "ComplexFieldModelStruct" {
		t.Errorf("GetName failed, expected: ComplexFieldModelStruct, got: %s", objModel.GetName())
	}
	
	// Test GetPkgPath
	expectedPkgPath := "github.com/muidea/magicOrm/provider/local"
	if objModel.GetPkgPath() != expectedPkgPath {
		t.Errorf("GetPkgPath failed, expected: %s, got: %s", expectedPkgPath, objModel.GetPkgPath())
	}
	
	// Test GetPkgKey
	expectedPkgKey := expectedPkgPath + "/" + "ComplexFieldModelStruct"
	if objModel.GetPkgKey() != expectedPkgKey {
		t.Errorf("GetPkgKey failed, expected: %s, got: %s", expectedPkgKey, objModel.GetPkgKey())
	}
	
	// Test GetFields
	fields := objModel.GetFields()
	if len(fields) != 9 {
		t.Errorf("GetFields failed, expected 9 fields, got: %d", len(fields))
		for idx, f := range fields {
			t.Logf("Field[%d]: %s", idx, f.GetName())
		}
	}
	
	// Test GetPrimaryField
	primaryField := objModel.GetPrimaryField()
	if primaryField == nil {
		t.Errorf("GetPrimaryField failed, no primary field found")
	} else if primaryField.GetName() != "id" {
		t.Errorf("GetPrimaryField failed, expected: id, got: %s", primaryField.GetName())
	}
	
	// Test GetField
	nameField := objModel.GetField("name")
	if nameField == nil {
		t.Errorf("GetField failed, name field not found")
	} else if nameField.GetName() != "name" {
		t.Errorf("GetField failed, expected: name, got: %s", nameField.GetName())
	}
	
	// Test GetField for non-existent field
	nonExistentField := objModel.GetField("nonExistent")
	if nonExistentField != nil {
		t.Errorf("GetField failed, non-existent field should return nil")
	}
}

func TestObjectSetFieldValue(t *testing.T) {
	// Create a test entity
	entity := ComplexFieldModelStruct{
		ID:            1,
		Name:          "Original Name",
		Score:         95.5,
		IsActive:      true,
		CreatedTime:   time.Now(),
		IntPointer:    nil,
		StringPointer: nil,
		IntSlice:      nil,
		StringSlice:   nil,
	}
	
	entityValue := reflect.ValueOf(entity)
	objModel, err := getValueModel(entityValue)
	if err != nil {
		t.Errorf("getValueModel failed: %s", err.Error())
		return
	}
	
	// Test SetFieldValue
	newNameVal := "Updated Name"
	nameValue := reflect.ValueOf(newNameVal)
	objModel.SetFieldValue("name", NewValue(nameValue))
	
	// Check if the value was updated
	nameField := objModel.GetField("name")
	if nameField == nil {
		t.Errorf("GetField failed, name field not found")
		return
	}
	
	// Get the value as reflect.Value and convert to string
	updatedValue := nameField.GetValue()
	fieldValue := updatedValue.Get().(reflect.Value)
	if fieldValue.String() != newNameVal {
		t.Errorf("SetFieldValue failed, expected: %s, got: %s", newNameVal, fieldValue.String())
	}
}

func TestObjectInterface(t *testing.T) {
	// Create an instance with values
	intVal := 42
	strVal := "test string"
	entity := ComplexFieldModelStruct{
		ID:            1,
		Name:          "Test Name",
		Score:         95.5,
		IsActive:      true,
		CreatedTime:   time.Now(),
		IntPointer:    &intVal,
		StringPointer: &strVal,
		IntSlice:      []int{1, 2, 3},
		StringSlice:   []string{"a", "b", "c"},
	}
	
	entityValue := reflect.ValueOf(entity)
	objModel, err := getValueModel(entityValue)
	if err != nil {
		t.Errorf("getValueModel failed: %s", err.Error())
		return
	}
	
	// Test Interface with ptrValue=false, for Origin view
	originInterface := objModel.Interface(false, model.OriginView)
	if originInterface == nil {
		t.Errorf("Interface failed for OriginView, returned nil")
		return
	}
	
	originStruct, ok := originInterface.(ComplexFieldModelStruct)
	if !ok {
		t.Errorf("Interface failed, expected ComplexFieldModelStruct type, got: %T", originInterface)
		return
	}
	
	if originStruct.ID != entity.ID || originStruct.Name != entity.Name {
		t.Errorf("Interface returned incorrect data for OriginView")
	}
	
	// Test Interface with ptrValue=true, for Detail view
	detailInterface := objModel.Interface(true, model.DetailView)
	if detailInterface == nil {
		t.Errorf("Interface failed for DetailView, returned nil")
		return
	}
	
	detailStruct, ok := detailInterface.(*ComplexFieldModelStruct)
	if !ok {
		t.Errorf("Interface failed, expected *ComplexFieldModelStruct type, got: %T", detailInterface)
		return
	}
	
	if detailStruct.ID != entity.ID || detailStruct.Name != entity.Name {
		t.Errorf("Interface returned incorrect data for DetailView")
	}
	
	// Test Interface with ptrValue=false, for Lite view
	liteInterface := objModel.Interface(false, model.LiteView)
	if liteInterface == nil {
		t.Errorf("Interface failed for LiteView, returned nil")
		return
	}
	
	liteStruct, ok := liteInterface.(ComplexFieldModelStruct)
	if !ok {
		t.Errorf("Interface failed, expected ComplexFieldModelStruct type, got: %T", liteInterface)
		return
	}
	
	// In Lite view, only ID, Name, and IsActive should be included
	if liteStruct.ID != entity.ID || liteStruct.Name != entity.Name || !liteStruct.IsActive {
		t.Errorf("Interface returned incorrect data for LiteView")
	}
	
	// Score should not be included in Lite view (should be zero value)
	zeroScore := float64(0)
	if liteStruct.Score != zeroScore {
		t.Errorf("Interface included field not in LiteView, Score should be zero, got: %v", liteStruct.Score)
	}
}

func TestObjectCopy(t *testing.T) {
	// Create an instance with values
	intVal := 42
	strVal := "test string"
	entity := ComplexFieldModelStruct{
		ID:            1,
		Name:          "Test Name",
		Score:         95.5,
		IsActive:      true,
		CreatedTime:   time.Now(),
		IntPointer:    &intVal,
		StringPointer: &strVal,
		IntSlice:      []int{1, 2, 3},
		StringSlice:   []string{"a", "b", "c"},
	}
	
	entityValue := reflect.ValueOf(entity)
	objModel, err := getValueModel(entityValue)
	if err != nil {
		t.Errorf("getValueModel failed: %s", err.Error())
		return
	}
	
	// Test Copy with reset=false
	copiedModel := objModel.Copy(false)
	if copiedModel == nil {
		t.Errorf("Copy failed, returned nil")
		return
	}
	
	// Check if field values are preserved
	idField := copiedModel.GetField("id")
	if idField == nil {
		t.Errorf("Copy failed, ID field not found")
	} else {
		// Get the value as reflect.Value
		idValField := idField.GetValue().Get().(reflect.Value)
		if idValField.Int() != int64(entity.ID) {
			t.Errorf("Copy failed, ID field value not preserved, expected: %d, got: %d", entity.ID, idValField.Int())
		}
	}
	
	nameField := copiedModel.GetField("name")
	if nameField == nil {
		t.Errorf("Copy failed, Name field not found")
	} else {
		// Get the value as reflect.Value
		nameValField := nameField.GetValue().Get().(reflect.Value)
		if nameValField.String() != entity.Name {
			t.Errorf("Copy failed, Name field value not preserved, expected: %s, got: %s", entity.Name, nameValField.String())
		}
	}
	
	// Test Copy with reset=true
	resetModel := objModel.Copy(true)
	if resetModel == nil {
		t.Errorf("Copy with reset failed, returned nil")
		return
	}
	
	// When reset=true, the field values should not be copied,
	// but the fields themselves should still exist
	idResetField := resetModel.GetField("id")
	if idResetField == nil {
		t.Errorf("Copy with reset failed, ID field not found")
	}
	
	nameResetField := resetModel.GetField("name")
	if nameResetField == nil {
		t.Errorf("Copy with reset failed, Name field not found")
	}
}

func TestNestedObjectModel(t *testing.T) {
	// Create a nested structure
	intVal := 42
	strVal := "test string"
	nested := NestedModelStruct{
		ID:   1,
		Name: "Nested Test",
		Complex: ComplexFieldModelStruct{
			ID:          2,
			Name:        "Inner Complex",
			Score:       85.0,
			IsActive:    true,
			IntPointer:  &intVal,
			StringSlice: []string{"x", "y", "z"},
		},
		ComplexP: &ComplexFieldModelStruct{
			ID:            3,
			Name:          "Inner Pointer Complex",
			Score:         75.0,
			StringPointer: &strVal,
			IntSlice:      []int{4, 5, 6},
		},
	}
	
	entityValue := reflect.ValueOf(nested)
	objModel, err := getValueModel(entityValue)
	if err != nil {
		t.Errorf("getValueModel failed for nested struct: %s", err.Error())
		return
	}
	
	// Test that fields are correctly identified
	fields := objModel.GetFields()
	if len(fields) != 4 {
		t.Errorf("GetFields failed for nested struct, expected 4 fields, got: %d", len(fields))
		for idx, f := range fields {
			t.Logf("Field[%d]: %s", idx, f.GetName())
		}
	}
	
	// Check the Complex field type
	complexField := objModel.GetField("complex")
	if complexField == nil {
		t.Errorf("GetField failed, complex field not found")
		return
	}
	
	if !complexField.IsStruct() {
		t.Errorf("Field type check failed, complex should be a struct")
	}
	
	// Check the ComplexP field type
	complexPField := objModel.GetField("complexP")
	if complexPField == nil {
		t.Errorf("GetField failed, complexP field not found")
		return
	}
	
	if !complexPField.IsStruct() || !complexPField.IsPtrType() {
		t.Errorf("Field type check failed, complexP should be a struct pointer")
	}
	
	// Test Interface return for nested struct
	nestedInterface := objModel.Interface(false, model.DetailView)
	if nestedInterface == nil {
		t.Errorf("Interface failed for nested struct with DetailView, returned nil")
		return
	}
	
	nestedResult, ok := nestedInterface.(NestedModelStruct)
	if !ok {
		t.Errorf("Interface failed, expected NestedModelStruct type, got: %T", nestedInterface)
		return
	}
	
	// Check nested struct values
	if nestedResult.ID != nested.ID || nestedResult.Name != nested.Name {
		t.Errorf("Interface returned incorrect data for nested struct")
	}
	
	// Check complex field within nested struct
	if nestedResult.Complex.ID != nested.Complex.ID || nestedResult.Complex.Name != nested.Complex.Name {
		t.Errorf("Interface returned incorrect data for nested complex field")
	}
	
	// Check complex pointer field within nested struct
	if nestedResult.ComplexP == nil || nestedResult.ComplexP.ID != nested.ComplexP.ID || nestedResult.ComplexP.Name != nested.ComplexP.Name {
		t.Errorf("Interface returned incorrect data for nested complex pointer field")
	}
}
