package remote

import (
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestFieldImplementation(t *testing.T) {
	// Create a test Field
	field := &Field{
		Name:        "testField",
		ShowName:    "Test Field",
		Description: "Test field description",
		Type:        &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
		Spec:        &SpecImpl{FieldName: "testField", PrimaryKey: true},
		value:       NewValue(int64(123)),
	}

	// Test GetName
	if field.GetName() != "testField" {
		t.Errorf("GetName failed, expected 'testField', got '%s'", field.GetName())
	}

	// Test GetShowName
	if field.GetShowName() != "Test Field" {
		t.Errorf("GetShowName failed, expected 'Test Field', got '%s'", field.GetShowName())
	}

	// Test GetDescription
	if field.GetDescription() != "Test field description" {
		t.Errorf("GetDescription failed, expected 'Test field description', got '%s'", field.GetDescription())
	}

	// Test GetType
	fieldType := field.GetType()
	if fieldType == nil {
		t.Errorf("GetType failed, returned nil")
		return
	}
	if fieldType.GetName() != "int64" {
		t.Errorf("GetType failed, expected type name 'int64', got '%s'", fieldType.GetName())
	}

	// Test GetSpec
	fieldSpec := field.GetSpec()
	if fieldSpec == nil {
		t.Errorf("GetSpec failed, returned nil")
		return
	}
	if !fieldSpec.IsPrimaryKey() {
		t.Errorf("GetSpec failed, expected primary key, got non-primary key")
	}

	// Test GetValue
	fieldValue := field.GetValue()
	if fieldValue == nil {
		t.Errorf("GetValue failed, returned nil")
		return
	}
	if fieldValue.Get() != int64(123) {
		t.Errorf("GetValue failed, expected 123, got %v", fieldValue.Get())
	}

	// Test SetValue
	field.SetValue(int64(456))
	updatedValue := field.GetValue()
	if updatedValue.Get() != int64(456) {
		t.Errorf("SetValue failed, expected 456, got %v", updatedValue.Get())
	}

	// Test IsPrimaryField
	if !models.IsPrimaryField(field) {
		t.Errorf("IsPrimaryField failed, expected true, got false")
	}

	// Test IsBasic
	if !models.IsBasicField(field) {
		t.Errorf("IsBasic failed, expected true, got false")
	}

	// Test IsStruct
	if models.IsStructField(field) {
		t.Errorf("IsStruct failed, expected false, got true")
	}

	// Test IsSlice
	if models.IsSliceField(field) {
		t.Errorf("IsSlice failed, expected false, got true")
	}

	// Test IsPtrType
	if models.IsPtrField(field) {
		t.Errorf("IsPtrType failed, expected false, got true")
	}
}

func TestFieldCopy(t *testing.T) {
	// Create a test Field
	field := &Field{
		Name:        "testField",
		ShowName:    "Test Field",
		Description: "Test field description",
		Type:        &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
		Spec:        &SpecImpl{FieldName: "testField", PrimaryKey: true},
		value:       NewValue(int64(123)),
	}

	// Test copy with reset=false
	copiedField, err := field.copy(models.OriginView)
	if err != nil {
		t.Errorf("Field copy(false) failed with error: %v", err)
		return
	}

	if copiedField.Name != field.Name ||
		copiedField.ShowName != field.ShowName ||
		copiedField.Description != field.Description {
		t.Errorf("Field copy(false) failed, basic properties don't match")
	}

	if copiedField.GetValue().Get() != field.GetValue().Get() {
		t.Errorf("Field copy(false) failed, values don't match: expected %v, got %v",
			field.GetValue().Get(), copiedField.GetValue().Get())
	}

	// Test copy with reset=true
	resetField, err := field.copy(models.MetaView)
	if err != nil {
		t.Errorf("Field copy(true) failed with error: %v", err)
		return
	}

	if resetField.Name != field.Name ||
		resetField.ShowName != field.ShowName ||
		resetField.Description != field.Description {
		t.Errorf("Field copy(true) failed, basic properties don't match")
	}

	// With reset=true, the value should be reset
	if !resetField.GetValue().IsZero() {
		t.Errorf("Field copy(true) failed, value should be zero, got %v", resetField.GetValue().Get())
	}
}

func TestFieldValueVerification(t *testing.T) {
	// Create various field definitions for validation testing
	tests := []struct {
		name           string
		field          *Field
		expectedToPass bool
	}{
		{
			name: "Valid integer primary key",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &SpecImpl{FieldName: "id", PrimaryKey: true},
			},
			expectedToPass: true,
		},
		{
			name: "Valid string primary key",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{FieldName: "id", PrimaryKey: true},
			},
			expectedToPass: true,
		},
		{
			name: "Valid auto-increment field",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &SpecImpl{FieldName: "id", ValueDeclare: models.AutoIncrement},
			},
			expectedToPass: true,
		},
		{
			name: "Invalid auto-increment on string field",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{FieldName: "id", ValueDeclare: models.AutoIncrement},
			},
			expectedToPass: false,
		},
		{
			name: "Valid UUID field",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{FieldName: "id", ValueDeclare: models.UUID},
			},
			expectedToPass: true,
		},
		{
			name: "Invalid UUID on integer field",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &SpecImpl{FieldName: "id", ValueDeclare: models.UUID},
			},
			expectedToPass: false,
		},
		{
			name: "Valid SnowFlake field",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &SpecImpl{FieldName: "id", ValueDeclare: models.SnowFlake},
			},
			expectedToPass: true,
		},
		{
			name: "Invalid SnowFlake on string field",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{FieldName: "id", ValueDeclare: models.SnowFlake},
			},
			expectedToPass: false,
		},
		{
			name: "Valid DateTime field",
			field: &Field{
				Name: "createdAt",
				Type: &TypeImpl{Name: "time.Time", Value: models.TypeDateTimeValue},
				Spec: &SpecImpl{FieldName: "createdAt", ValueDeclare: models.DateTime},
			},
			expectedToPass: true,
		},
		{
			name: "Invalid DateTime on string field",
			field: &Field{
				Name: "createdAt",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{FieldName: "createdAt", ValueDeclare: models.DateTime},
			},
			expectedToPass: false,
		},
		{
			name: "Invalid primary key on struct field",
			field: &Field{
				Name: "objField",
				Type: &TypeImpl{Name: "TestStruct", Value: models.TypeStructValue},
				Spec: &SpecImpl{FieldName: "objField", PrimaryKey: true},
			},
			expectedToPass: false,
		},
		{
			name: "Invalid primary key on slice field",
			field: &Field{
				Name: "sliceField",
				Type: &TypeImpl{Name: "[]int", Value: models.TypeSliceValue},
				Spec: &SpecImpl{FieldName: "sliceField", PrimaryKey: true},
			},
			expectedToPass: false,
		},
		{
			name: "Field without type",
			field: &Field{
				Name: "missingType",
				Spec: &SpecImpl{FieldName: "missingType"},
			},
			expectedToPass: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.field.verify()
			if test.expectedToPass && err != nil {
				t.Errorf("Expected field to pass verification, but got error: %v", err)
			} else if !test.expectedToPass && err == nil {
				t.Errorf("Expected field to fail verification, but it passed")
			}
		})
	}
}

func TestFieldValueImplementation(t *testing.T) {
	// Create test FieldValue
	fieldVal := &FieldValue{
		Name:  "testField",
		Value: int64(123),
	}

	// Test String method
	strRepresentation := fieldVal.String()
	if strRepresentation == "" {
		t.Errorf("String() failed, returned empty string")
	}

	// Test IsNil
	if !fieldVal.IsValid() {
		t.Errorf("IsNil failed, expected false, got true")
	}

	// Test with nil value
	nilFieldVal := &FieldValue{
		Name:  "nilField",
		Value: nil,
	}
	if nilFieldVal.IsValid() {
		t.Errorf("IsNil failed for nil value, expected true, got false")
	}

	// Test IsZero
	nonZeroFieldVal := &FieldValue{
		Name:  "nonZeroField",
		Value: int64(123),
	}
	if nonZeroFieldVal.IsZero() {
		t.Errorf("IsZero failed for non-zero value, expected false, got true")
	}

	zeroFieldVal := &FieldValue{
		Name:  "zeroField",
		Value: int64(0),
	}
	if !zeroFieldVal.IsZero() {
		t.Errorf("IsZero failed for zero value, expected true, got false")
	}

	// Test Set and Get
	testFieldVal := &FieldValue{
		Name:  "testField",
		Value: int64(123),
	}

	if testFieldVal.Get() != int64(123) {
		t.Errorf("Get failed, expected 123, got %v", testFieldVal.Get())
	}

	testFieldVal.Set(int64(456))
	if testFieldVal.Get() != int64(456) {
		t.Errorf("Set failed, expected 456 after setting, got %v", testFieldVal.Get())
	}

	// Test GetName
	if testFieldVal.GetName() != "testField" {
		t.Errorf("GetName failed, expected 'testField', got '%s'", testFieldVal.GetName())
	}

	// Test copy
	copiedFieldVal := testFieldVal.copy()
	if copiedFieldVal.Name != testFieldVal.Name || copiedFieldVal.Value != testFieldVal.Value {
		t.Errorf("copy failed, values don't match")
	}

	// Verify the copy is independent
	testFieldVal.Set(int64(789))
	if copiedFieldVal.Value == testFieldVal.Value {
		t.Errorf("copy failed, copy should not be affected by changes to original")
	}
}
