package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/models"
)

// TestFieldProperties tests the properties of fields
func TestFieldProperties(t *testing.T) {
	// Create test struct with various field types
	type TestStruct struct {
		ID        int        `orm:"id key"`
		Name      string     `orm:"name"`
		Active    bool       `orm:"active"`
		Value     float64    `orm:"value"`
		Created   time.Time  `orm:"created"`
		UpdatedAt *time.Time `orm:"updatedAt"`
	}

	// Get model for test struct
	testModel, err := GetEntityModel(&TestStruct{}, nil)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test properties of all fields
	fields := testModel.GetFields()
	for _, field := range fields {
		// Verify field has basic properties
		if field.GetName() == "" {
			t.Errorf("Field name should not be empty")
		}

		// Check field types are properly detected
		switch field.GetName() {
		case "id":
			if !models.IsBasicField(field) || models.IsSliceField(field) || models.IsStructField(field) {
				t.Errorf("ID field should be a basic type")
			}
			if !models.IsPrimaryField(field) {
				t.Errorf("ID field should be a primary key")
			}
		case "name":
			if !models.IsBasicField(field) || models.IsPtrField(field) {
				t.Errorf("Name field should be a basic non-pointer type")
			}
		case "updatedAt":
			if !models.IsPtrField(field) {
				t.Errorf("UpdatedAt field should be a pointer type")
			}
		case "created":
			if !models.IsBasicField(field) {
				t.Errorf("Created field should be a struct type")
			}
		}
	}
}

// TestFieldCompare tests comparing fields of the same model
func TestFieldCompare(t *testing.T) {
	// Create test struct
	type TestStruct struct {
		ID   int    `orm:"id key"`
		Name string `orm:"name"`
	}

	// Get model for test struct
	testModel, err := GetEntityModel(&TestStruct{}, nil)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Get fields
	idField := testModel.GetField("id")
	nameField := testModel.GetField("name")

	// Fields should be different
	if idField.GetName() == nameField.GetName() {
		t.Errorf("Field names should be different")
	}

	// Field types should be different
	if reflect.DeepEqual(idField.GetType(), nameField.GetType()) {
		t.Errorf("Field types should be different")
	}

	// Primary key status should be different
	if models.IsPrimaryField(idField) == models.IsPrimaryField(nameField) {
		t.Errorf("Primary key status should be different")
	}

	idField.SetValue(100)
	if idField.GetValue().Get().(int) != 100 {
		t.Errorf("Field value should be 100")
	}

	testVal := testModel.Interface(false)
	if testVal.(TestStruct).ID != 100 {
		t.Errorf("Interface failed for DetailView, expected ID: 100, got: %d", testVal.(TestStruct).ID)
	}
	testPtrVal := testModel.Interface(true)
	if testPtrVal.(*TestStruct).ID != 100 {
		t.Errorf("Interface failed for DetailView, expected ID: 100, got: %d", testPtrVal.(*TestStruct).ID)
	}
}

// TestFieldValueDeclare tests value declarations on fields
func TestFieldValueDeclare(t *testing.T) {
	// Create test struct with different value declarations
	type TestDeclareStruct struct {
		ID        int       `orm:"id key auto"`
		UUID      string    `orm:"uuid uuid"`
		Snowflake int64     `orm:"snowflake snowflake"`
		Created   time.Time `orm:"created datetime"`
	}

	// Get model for test struct
	testModel, err := GetEntityModel(&TestDeclareStruct{
		UUID:      "uuid",
		Snowflake: 123,
		Created:   time.Now(),
	}, nil)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test cases
	testCases := []struct {
		fieldName    string
		valueDeclare models.ValueDeclare
	}{
		{"id", models.AutoIncrement},
		{"uuid", models.UUID},
		{"snowflake", models.Snowflake},
		{"created", models.DateTime},
	}

	for _, tc := range testCases {
		field := testModel.GetField(tc.fieldName)
		if field == nil {
			t.Errorf("GetField(%s) returned nil", tc.fieldName)
			continue
		}

		// Check the field's value declaration
		fieldSpec := field.GetSpec()
		if fieldSpec.GetValueDeclare() != tc.valueDeclare {
			t.Errorf("Field.GetSpec().GetValueDeclare() for %s, expected: %v, got: %v",
				tc.fieldName, tc.valueDeclare, fieldSpec.GetValueDeclare())
		}
	}
}
