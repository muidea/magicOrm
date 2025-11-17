package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
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
	ID       int                      `orm:"id key auto" view:"detail,lite"`
	Name     string                   `orm:"name" view:"detail,lite"`
	Complex  ComplexFieldModelStruct  `orm:"complex" view:"detail"`
	ComplexP *ComplexFieldModelStruct `orm:"complexP" view:"detail"`
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

	entityValue := reflect.ValueOf(&entity)
	objModel, err := getValueModel(entityValue, models.MetaView)
	if err != nil {
		t.Errorf("getValueModel failed: %s", err.Error())
		return
	}

	// Test SetFieldValue
	newNameVal := "Updated Name"
	objModel.SetFieldValue("name", "Updated Name")

	// Check if the value was updated
	nameField := objModel.GetField("name")
	if nameField == nil {
		t.Errorf("GetField failed, name field not found")
		return
	}

	// Get the value as reflect.Value and convert to string
	updatedValue := nameField.GetValue()
	fieldValue := updatedValue.Get()
	if fieldValue.(string) != newNameVal {
		t.Errorf("SetFieldValue failed, expected: %s, got: %s", newNameVal, fieldValue.(string))
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

	entityValue := reflect.ValueOf(&entity)
	objModel, err := getValueModel(entityValue, models.OriginView)
	if err != nil {
		t.Errorf("getValueModel failed: %s", err.Error())
		return
	}

	// Test Interface with ptrValue=false, for Origin view
	originInterface := objModel.Interface(false)
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
	detailInterface := objModel.Interface(true)
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

	liteModel := objModel.Copy(models.LiteView)
	// Test Interface with ptrValue=false, for Lite view
	liteInterface := liteModel.Interface(false)
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

func TestComplexObjectCopy(t *testing.T) {
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

	entityValue := reflect.ValueOf(&entity)
	objModel, err := getValueModel(entityValue, models.OriginView)
	if err != nil {
		t.Errorf("getValueModel failed: %s", err.Error())
		return
	}

	// Test Copy with reset=false
	copiedModel := objModel.Copy(models.OriginView)
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
		idValField := idField.GetValue().Get()
		if idValField.(int) != entity.ID {
			t.Errorf("Copy failed, ID field value not preserved, expected: %d, got: %d", entity.ID, idValField.(int64))
		}
	}

	nameField := copiedModel.GetField("name")
	if nameField == nil {
		t.Errorf("Copy failed, Name field not found")
	} else {
		// Get the value as reflect.Value
		nameValField := nameField.GetValue().Get()
		if nameValField.(string) != entity.Name {
			t.Errorf("Copy failed, Name field value not preserved, expected: %s, got: %s", entity.Name, nameValField.(string))
		}
	}

	// Test Copy with reset=true
	resetModel := objModel.Copy(models.MetaView)
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

	entityValue := reflect.ValueOf(&nested)
	objModel, err := getValueModel(entityValue, models.OriginView)
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

	if !models.IsStructField(complexField) {
		t.Errorf("Field type check failed, complex should be a struct")
	}

	// Check the ComplexP field type
	complexPField := objModel.GetField("complexP")
	if complexPField == nil {
		t.Errorf("GetField failed, complexP field not found")
		return
	}

	if !models.IsStructField(complexPField) || !models.IsPtrField(complexPField) {
		t.Errorf("Field type check failed, complexP should be a struct pointer")
	}

	// Test Interface return for nested struct
	nestedInterface := objModel.Interface(false)
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
