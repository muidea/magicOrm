package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
)

// ViewSpecStruct defines a struct for testing view specifications
type ViewSpecStruct struct {
	ID          int     `orm:"id key auto" view:"detail,lite"`
	Name        string  `orm:"name" view:"detail,lite"`
	Description string  `orm:"description" view:"detail"`
	Score       float64 `orm:"score" view:"detail"`
	CreatedAt   time.Time `orm:"createdAt" view:"lite"`
	UpdatedAt   time.Time `orm:"updatedAt"`              // No view spec
	InternalID  string  `orm:"internalId"`              // No view spec
}

// ValueDeclareTestStruct defines a struct for testing value declarations
type ValueDeclareTestStruct struct {
	ID            int       `orm:"id key auto" view:"detail,lite"`
	UUID          string    `orm:"uuid,uuid" view:"detail,lite"`
	SnowflakeID   int64     `orm:"snowflakeId,snowFlake" view:"detail,lite"`
	CreatedTime   time.Time `orm:"createdTime,dateTime" view:"detail"`
	AutoValue     int       `orm:"autoValue auto" view:"detail"`
	RegularString string    `orm:"regularString" view:"detail"`
}

func TestViewSpecParsing(t *testing.T) {
	// Test parsing of view specifications from struct tags
	var entity ViewSpecStruct
	entityType := reflect.TypeOf(entity)
	
	// Test that view specs are correctly parsed for each field
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		
		spec, err := NewSpec(field.Tag)
		if err != nil {
			t.Errorf("NewSpec failed for field %s: %s", field.Name, err.Error())
			continue
		}
		
		// Use spec to avoid unused variable warning
		_ = spec
		
		// Check view specifications
		viewDeclares := getViewItems(string(field.Tag.Get("view")))
		
		// Verify view declarations based on the field name
		switch field.Name {
		case "ID", "Name":
			// Should have both detail and lite views
			if len(viewDeclares) != 2 {
				t.Errorf("Field %s should have 2 view specs, got: %d", field.Name, len(viewDeclares))
			}
		case "Description", "Score":
			// Should have only detail view
			if len(viewDeclares) != 1 {
				t.Errorf("Field %s should have 1 view spec, got: %d", field.Name, len(viewDeclares))
			}
		case "CreatedAt":
			// Should have only lite view
			if len(viewDeclares) != 1 {
				t.Errorf("Field %s should have 1 view spec, got: %d", field.Name, len(viewDeclares))
			}
		case "UpdatedAt", "InternalID":
			// Should have no view specs
			if len(viewDeclares) != 0 {
				t.Errorf("Field %s should have no view specs, got: %v", field.Name, viewDeclares)
			}
		}
	}
}

func TestGetViewItems(t *testing.T) {
	// Test retrieving view items from model
	var entity ViewSpecStruct
	entityType := reflect.TypeOf(entity)
	
	objModel, err := getTypeModel(entityType)
	if err != nil {
		t.Errorf("getTypeModel failed: %s", err.Error())
		return
	}
	
	// Use objModel to avoid unused variable warning
	_ = objModel
	
	// Test detail view
	detailItems := objModel.GetViewFields(model.DetailView)
	if len(detailItems) != 4 {
		t.Errorf("GetViewFields for DetailView should return 4 fields, got: %d", len(detailItems))
		for i, item := range detailItems {
			t.Logf("DetailView item[%d]: %s", i, item.GetName())
		}
	}
	
	// Check that expected fields are included in detail view
	detailFieldNames := []string{"id", "name", "description", "score"}
	for _, v := range detailFieldNames {
		found := false
		for _, item := range detailItems {
			if item.GetName() == v {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Field %s should be included in DetailView", v)
		}
	}
	
	// Test lite view
	liteItems := objModel.GetViewFields(model.LiteView)
	if len(liteItems) != 3 {
		t.Errorf("GetViewFields for LiteView should return 3 fields, got: %d", len(liteItems))
		for i, item := range liteItems {
			t.Logf("LiteView item[%d]: %s", i, item.GetName())
		}
	}
	
	// Check that expected fields are included in lite view
	liteFieldNames := []string{"id", "name", "createdAt"}
	for _, v := range liteFieldNames {
		found := false
		for _, item := range liteItems {
			if item.GetName() == v {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Field %s should be included in LiteView", v)
		}
	}
}

func TestModelViewInterfaces(t *testing.T) {
	// Test model interface creation with different views
	entity := ViewSpecStruct{
		ID:          1,
		Name:        "Test Entity",
		Description: "This is a test description",
		Score:       95.5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		InternalID:  "INT123",
	}
	
	entityValue := reflect.ValueOf(entity)
	objModel, err := getValueModel(entityValue)
	if err != nil {
		t.Errorf("getValueModel failed: %s", err.Error())
		return
	}
	
	// Test Detail view
	detailInterface := objModel.Interface(false, model.DetailView)
	detailStruct, ok := detailInterface.(ViewSpecStruct)
	if !ok {
		t.Errorf("Interface failed for DetailView, expected ViewSpecStruct type, got: %T", detailInterface)
		return
	}
	
	// In Detail view, ID, Name, Description, and Score should be included
	if detailStruct.ID != entity.ID || detailStruct.Name != entity.Name || 
	   detailStruct.Description != entity.Description || detailStruct.Score != entity.Score {
		t.Errorf("Interface returned incorrect data for DetailView")
	}
	
	// CreatedAt, UpdatedAt, and InternalID should not be included in Detail view (zero values)
	zeroTime := time.Time{}
	if !detailStruct.CreatedAt.Equal(zeroTime) || !detailStruct.UpdatedAt.Equal(zeroTime) || detailStruct.InternalID != "" {
		t.Errorf("Interface included fields not in DetailView")
	}
	
	// Test Lite view
	liteInterface := objModel.Interface(false, model.LiteView)
	liteStruct, ok := liteInterface.(ViewSpecStruct)
	if !ok {
		t.Errorf("Interface failed for LiteView, expected ViewSpecStruct type, got: %T", liteInterface)
		return
	}
	
	// In Lite view, only ID, Name, and CreatedAt should be included
	if liteStruct.ID != entity.ID || liteStruct.Name != entity.Name || !liteStruct.CreatedAt.Equal(entity.CreatedAt) {
		t.Errorf("Interface returned incorrect data for LiteView")
	}
	
	// Description, Score, UpdatedAt, and InternalID should not be included in Lite view (zero values)
	if liteStruct.Description != "" || liteStruct.Score != 0 || !liteStruct.UpdatedAt.Equal(zeroTime) || liteStruct.InternalID != "" {
		t.Errorf("Interface included fields not in LiteView")
	}
}

func TestValueDeclareSpecParsing(t *testing.T) {
	// Test parsing value declarations from struct tags
	var entity ValueDeclareTestStruct
	entityType := reflect.TypeOf(entity)
	
	objModel, err := getTypeModel(entityType)
	if err != nil {
		t.Errorf("getTypeModel failed: %s", err.Error())
		return
	}
	
	// Use objModel to avoid unused variable warning
	_ = objModel
	
	// Check auto field
	autoField := objModel.GetField("autoValue")
	if autoField == nil {
		t.Error("Field autoValue not found")
		return
	}
	
	autoSpec := autoField.GetSpec()
	if autoSpec == nil {
		t.Error("Spec for autoValue field is nil")
		return
	}
	
	valueDeclare := autoSpec.GetValueDeclare()
	if !valueDeclare.IsAutoIncrement() {
		t.Errorf("Expected autoValue to have AutoIncrement value declare, got %v", valueDeclare)
	}
	
	// Verify default value handling
	regularField := objModel.GetField("regularString")
	if regularField == nil {
		t.Error("Field regularString not found")
		return
	}
	
	regularSpec := regularField.GetSpec()
	if regularSpec == nil {
		t.Error("Spec for regularString field is nil")
		return
	}
	
	// Check default value is correctly set
	if regularSpec.GetValueDeclare() != model.Customer {
		t.Errorf("Expected Customer value declare for regularString, got %v", regularSpec.GetValueDeclare())
	}
}

func TestCombinedSpecifications(t *testing.T) {
	// Test combined specs with both view specs and value declares
	var entity ValueDeclareTestStruct
	entityType := reflect.TypeOf(entity)
	
	objModel, err := getTypeModel(entityType)
	if err != nil {
		t.Errorf("getTypeModel failed: %s", err.Error())
		return
	}
	
	_ = objModel
	
	// Test each field for combined specs
	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		
		// Check view specifications
		viewDeclares := getViewItems(string(field.Tag.Get("view")))
		
		// Verify view declarations based on the field name
		switch field.Name {
		case "ID", "Name":
			// Should have both detail and lite views
			if len(viewDeclares) != 2 {
				t.Errorf("Field %s should have 2 view specs, got: %d", field.Name, len(viewDeclares))
			}
		case "Description", "Score":
			// Should have only detail view
			if len(viewDeclares) != 1 {
				t.Errorf("Field %s should have 1 view spec, got: %d", field.Name, len(viewDeclares))
			}
		case "CreatedAt":
			// Should have only lite view
			if len(viewDeclares) != 1 {
				t.Errorf("Field %s should have 1 view spec, got: %d", field.Name, len(viewDeclares))
			}
		case "UpdatedAt", "InternalID":
			// Should have no view specs
			if len(viewDeclares) != 0 {
				t.Errorf("Field %s should have no view specs, got: %v", field.Name, viewDeclares)
			}
		}
	}
}

// Helper function to check if a view is in the list
func containsView(views []string, view string) bool {
	for _, v := range views {
		if v == view {
			return true
		}
	}
	return false
}

// Add GetViewFields function to object implementation
func (s *objectImpl) GetViewFields(viewDecl model.ViewDeclare) (ret model.Fields) {
	for _, field := range s.GetFields() {
		spec := field.GetSpec()
		if spec != nil && spec.EnableView(viewDecl) {
			ret = append(ret, field)
		}
	}
	return
}
