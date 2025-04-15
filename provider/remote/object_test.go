package remote

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/model"
)

func TestObjectImplementation(t *testing.T) {
	// Create test object
	obj := &Object{
		Name:        "TestObject",
		PkgPath:     "github.com/test/pkg",
		Description: "Test Object Description",
		Fields: []*Field{
			{
				Name:  "id",
				Type:  &TypeImpl{Name: "int64", Value: model.TypeBigIntegerValue},
				Spec:  &SpecImpl{FieldName: "id", PrimaryKey: true},
				value: NewValue(int64(123)),
			},
			{
				Name:  "name",
				Type:  &TypeImpl{Name: "string", Value: model.TypeStringValue},
				Spec:  &SpecImpl{FieldName: "name"},
				value: NewValue("test name"),
			},
		},
	}

	// Test GetName
	if obj.GetName() != "TestObject" {
		t.Errorf("GetName failed, expected 'TestObject', got '%s'", obj.GetName())
	}

	// Test GetDescription
	if obj.GetDescription() != "Test Object Description" {
		t.Errorf("GetDescription failed, expected 'Test Object Description', got '%s'", obj.GetDescription())
	}

	// Test GetPkgPath
	expectedPkgPath := "github.com/test/pkg"
	if obj.GetPkgPath() != expectedPkgPath {
		t.Errorf("GetPkgKey failed, expected '%s', got '%s'", expectedPkgPath, obj.GetPkgPath())
	}

	// Test GetFields
	fields := obj.GetFields()
	if len(fields) != 2 {
		t.Errorf("GetFields failed, expected 2 fields, got %d", len(fields))
		return
	}

	// Test GetPrimaryField
	pkField := obj.GetPrimaryField()
	if pkField == nil {
		t.Errorf("GetPrimaryField failed, no primary field found")
		return
	}
	if pkField.GetName() != "id" {
		t.Errorf("GetPrimaryField failed, expected 'id', got '%s'", pkField.GetName())
	}

	// Test GetField
	nameField := obj.GetField("name")
	if nameField == nil {
		t.Errorf("GetField failed, no field named 'name' found")
		return
	}
	if nameField.GetName() != "name" {
		t.Errorf("GetField failed, expected 'name', got '%s'", nameField.GetName())
	}

	// Test SetFieldValue
	obj.SetFieldValue("name", "updated name")
	updatedField := obj.GetField("name")
	if updatedField.GetValue().Get() != "updated name" {
		t.Errorf("SetFieldValue failed, expected 'updated name', got '%v'", updatedField.GetValue().Get())
	}

	// Test SetPrimaryFieldValue
	obj.SetPrimaryFieldValue(int64(456))
	updatedPKField := obj.GetPrimaryField()
	if updatedPKField.GetValue().Get() != int64(456) {
		t.Errorf("SetPrimaryFieldValue failed, expected 456, got %v", updatedPKField.GetValue().Get())
	}
}

func TestObjectCopy(t *testing.T) {
	// Create test object
	obj := &Object{
		Name:        "TestObject",
		PkgPath:     "github.com/test/pkg",
		Description: "Test Object Description",
		Fields: []*Field{
			{
				Name:  "id",
				Type:  &TypeImpl{Name: "int64", Value: model.TypeBigIntegerValue},
				Spec:  &SpecImpl{FieldName: "id", PrimaryKey: true},
				value: NewValue(int64(123)),
			},
			{
				Name:  "name",
				Type:  &TypeImpl{Name: "string", Value: model.TypeStringValue},
				Spec:  &SpecImpl{FieldName: "name"},
				value: NewValue("test name"),
			},
		},
	}

	// Test Copy with reset=true
	copiedObj := obj.Copy(model.MetaView)
	if copiedObj.GetName() != obj.GetName() ||
		copiedObj.GetPkgPath() != obj.GetPkgPath() ||
		copiedObj.GetDescription() != obj.GetDescription() {
		t.Errorf("Copy(true) failed, basic properties don't match")
	}

	// With reset=true, field structure is preserved but values are reset
	copiedFields := copiedObj.GetFields()
	if len(copiedFields) != len(obj.GetFields()) {
		t.Errorf("Copy(true) failed, field count mismatch: expected %d, got %d",
			len(obj.GetFields()), len(copiedFields))
	}

	// Check field values are reset
	nameField := copiedObj.GetField("name")
	if !nameField.GetValue().IsZero() {
		t.Errorf("Copy(true) failed, field value should be reset to zero")
	}

	// Test Copy with reset=false
	copiedObj2 := obj.Copy(model.OriginView)
	if copiedObj2.GetName() != obj.GetName() ||
		copiedObj2.GetPkgPath() != obj.GetPkgPath() ||
		copiedObj2.GetDescription() != obj.GetDescription() {
		t.Errorf("Copy(false) failed, basic properties don't match")
	}

	// With reset=false, field values should be preserved
	nameField2 := copiedObj2.GetField("name")
	originalField := obj.GetField("name")
	if nameField2.GetValue().Get() != originalField.GetValue().Get() {
		t.Errorf("Copy(false) failed, field value not preserved: expected %v, got %v",
			originalField.GetValue().Get(), nameField2.GetValue().Get())
	}
}

func TestObjectInterface(t *testing.T) {
	// Create test object
	obj := &Object{
		Name:        "TestObject",
		PkgPath:     "github.com/test/pkg",
		Description: "Test Object Description",
		Fields: []*Field{
			{
				Name:  "id",
				Type:  &TypeImpl{Name: "int64", Value: model.TypeBigIntegerValue},
				Spec:  &SpecImpl{FieldName: "id", PrimaryKey: true, ViewDeclare: []model.ViewDeclare{model.DetailView, model.LiteView}},
				value: NewValue(int64(123)),
			},
			{
				Name:  "name",
				Type:  &TypeImpl{Name: "string", Value: model.TypeStringValue},
				Spec:  &SpecImpl{FieldName: "name", ViewDeclare: []model.ViewDeclare{model.DetailView, model.LiteView}},
				value: NewValue("test name"),
			},
			{
				Name:  "description",
				Type:  &TypeImpl{Name: "string", Value: model.TypeStringValue},
				Spec:  &SpecImpl{FieldName: "description", ViewDeclare: []model.ViewDeclare{model.DetailView}},
				value: NewValue("test description"),
			},
		},
	}

	// Test Interface with OriginView
	result := obj.Interface(false)
	objVal, ok := result.(*ObjectValue)
	if !ok {
		t.Errorf("Interface failed, expected *ObjectValue, got %T", result)
		return
	}

	if objVal.Name != obj.Name || objVal.PkgPath != obj.PkgPath {
		t.Errorf("Interface failed, basic properties don't match")
	}

	if len(objVal.Fields) != 3 {
		t.Errorf("Interface with OriginView failed, expected 3 fields, got %d", len(objVal.Fields))
	}

	detailView := obj.Copy(model.DetailView)
	// Test Interface with DetailView
	result = detailView.Interface(false)
	objVal, ok = result.(*ObjectValue)
	if !ok {
		t.Errorf("Interface failed, expected *ObjectValue, got %T", result)
		return
	}

	if len(objVal.Fields) != 3 {
		t.Errorf("Interface with DetailView failed, expected 3 fields, got %d", len(objVal.Fields))
	}

	liteView := obj.Copy(model.LiteView)
	// Test Interface with LiteView
	result = liteView.Interface(false)
	objVal, ok = result.(*ObjectValue)
	if !ok {
		t.Errorf("Interface failed, expected *ObjectValue, got %T", result)
		return
	}

	if len(objVal.Fields) != 2 {
		t.Errorf("Interface with LiteView failed, expected 2 fields, got %d", len(objVal.Fields))
	}

	// Verify fields in LiteView
	hasDescription := false
	for _, f := range objVal.Fields {
		if f.Name == "description" {
			hasDescription = true
			break
		}
	}
	if hasDescription {
		t.Errorf("Interface with LiteView failed, 'description' field should not be included")
	}
}

func TestObjectValueImplementation(t *testing.T) {
	// Create test ObjectValue
	objVal := &ObjectValue{
		ID:      "123",
		Name:    "TestObject",
		PkgPath: "github.com/test/pkg",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(123)},
			{Name: "name", Value: "test name"},
		},
	}

	// Test GetName
	if objVal.GetName() != "TestObject" {
		t.Errorf("GetName failed, expected 'TestObject', got '%s'", objVal.GetName())
	}

	// Test GetPkgPath
	expectedPkgPath := "github.com/test/pkg"
	if objVal.GetPkgPath() != expectedPkgPath {
		t.Errorf("GetPkgPathy failed, expected '%s', got '%s'", expectedPkgPath, objVal.GetPkgPath())
	}

	// Test GetValue
	fields := objVal.GetValue()
	if len(fields) != 2 {
		t.Errorf("GetValue failed, expected 2 fields, got %d", len(fields))
	}

	// Test GetFieldValue
	nameVal := objVal.GetFieldValue("name")
	if nameVal != "test name" {
		t.Errorf("GetFieldValue failed, expected 'test name', got '%v'", nameVal)
	}

	// Test SetFieldValue
	objVal.SetFieldValue("name", "updated name")
	updatedVal := objVal.GetFieldValue("name")
	if updatedVal != "updated name" {
		t.Errorf("SetFieldValue failed, expected 'updated name', got '%v'", updatedVal)
	}

	// Test IsAssigned
	if !objVal.IsAssigned() {
		t.Errorf("IsAssigned failed, should be true")
	}

	// Create empty ObjectValue
	emptyObjVal := &ObjectValue{
		Name:    "EmptyObject",
		PkgPath: "github.com/test/pkg",
		Fields:  []*FieldValue{},
	}

	// Test IsAssigned with empty object
	if emptyObjVal.IsAssigned() {
		t.Errorf("IsAssigned failed for empty object, should be false")
	}

	// Test Copy
	copiedVal := objVal.Copy()
	if copiedVal.Name != objVal.Name || copiedVal.PkgPath != objVal.PkgPath {
		t.Errorf("Copy failed, basic properties don't match")
	}

	if len(copiedVal.Fields) != len(objVal.Fields) {
		t.Errorf("Copy failed, field count mismatch")
	}

	// Verify the copy is a deep copy
	objVal.SetFieldValue("name", "modified after copy")
	if copiedVal.GetFieldValue("name") == "modified after copy" {
		t.Errorf("Copy failed, copy should not be affected by changes to original")
	}
}

func TestSliceObjectValueImplementation(t *testing.T) {
	// Create test SliceObjectValue
	sliceObjVal := &SliceObjectValue{
		Name:    "TestSliceObject",
		PkgPath: "github.com/test/pkg",
		Values: []*ObjectValue{
			{
				ID:      "123",
				Name:    "TestObject1",
				PkgPath: "github.com/test/pkg",
				Fields: []*FieldValue{
					{Name: "id", Value: int64(123)},
					{Name: "name", Value: "test name 1"},
				},
			},
			{
				ID:      "456",
				Name:    "TestObject2",
				PkgPath: "github.com/test/pkg",
				Fields: []*FieldValue{
					{Name: "id", Value: int64(456)},
					{Name: "name", Value: "test name 2"},
				},
			},
		},
	}

	// Test GetName
	if sliceObjVal.GetName() != "TestSliceObject" {
		t.Errorf("GetName failed, expected 'TestSliceObject', got '%s'", sliceObjVal.GetName())
	}

	// Test GetPkgKey
	expectedPkgPath := "github.com/test/pkg"
	if sliceObjVal.GetPkgPath() != expectedPkgPath {
		t.Errorf("GetPkgKey failed, expected '%s', got '%s'", expectedPkgPath, sliceObjVal.GetPkgPath())
	}

	// Test GetValue
	values := sliceObjVal.GetValue()
	if len(values) != 2 {
		t.Errorf("GetValue failed, expected 2 values, got %d", len(values))
	}

	// Test IsAssigned
	if !sliceObjVal.IsAssigned() {
		t.Errorf("IsAssigned failed, should be true")
	}

	// Create empty SliceObjectValue
	emptySliceObjVal := &SliceObjectValue{
		Name:    "EmptySliceObject",
		PkgPath: "github.com/test/pkg",
		Values:  []*ObjectValue{},
	}

	// Test IsAssigned with empty slice
	if emptySliceObjVal.IsAssigned() {
		t.Errorf("IsAssigned failed for empty slice, should be false")
	}

	// Test Copy
	copiedVal := sliceObjVal.Copy()
	if copiedVal.Name != sliceObjVal.Name || copiedVal.PkgPath != sliceObjVal.PkgPath {
		t.Errorf("Copy failed, basic properties don't match")
	}

	if len(copiedVal.Values) != len(sliceObjVal.Values) {
		t.Errorf("Copy failed, value count mismatch")
	}

	// Verify the copy is a deep copy by modifying the original and checking the copy
	if len(sliceObjVal.Values) > 0 && len(sliceObjVal.Values[0].Fields) > 0 {
		originalFirstObjField := sliceObjVal.Values[0].Fields[0]
		originalFieldVal := originalFirstObjField.Value
		originalFirstObjField.Value = "modified"

		copiedFirstObjField := copiedVal.Values[0].Fields[0]
		if reflect.DeepEqual(copiedFirstObjField.Value, originalFirstObjField.Value) {
			t.Errorf("Copy failed, copy should not be affected by changes to original")
		}

		// Restore original for other tests
		originalFirstObjField.Value = originalFieldVal
	}

	// Test TransferObjectValue function
	transferredVal := TransferObjectValue("TransferredObject", "github.com/test/transfer", sliceObjVal.Values)
	if transferredVal.Name != "TransferredObject" || transferredVal.PkgPath != "github.com/test/transfer" {
		t.Errorf("TransferObjectValue failed, basic properties don't match")
	}

	if len(transferredVal.Values) != len(sliceObjVal.Values) {
		t.Errorf("TransferObjectValue failed, value count mismatch")
	}
}

func TestObjectValueEncoding(t *testing.T) {
	// Create test ObjectValue
	objVal := &ObjectValue{
		ID:      "123",
		Name:    "TestObject",
		PkgPath: "github.com/test/pkg",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(123)},
			{Name: "name", Value: "test name"},
		},
	}

	// Test EncodeObjectValue
	encodedData, err := EncodeObjectValue(objVal)
	if err != nil {
		t.Errorf("EncodeObjectValue failed with error: %v", err)
		return
	}

	// Test DecodeObjectValue
	decodedVal, decodeErr := DecodeObjectValue(encodedData)
	if decodeErr != nil {
		t.Errorf("DecodeObjectValue failed with error: %v", decodeErr)
		return
	}

	if !CompareObjectValue(objVal, decodedVal) {
		t.Errorf("Encode/Decode cycle failed, objects don't match")
	}

	// Test ConvertObjectValue
	convertedVal, convertErr := ConvertObjectValue(objVal)
	if convertErr != nil {
		t.Errorf("ConvertObjectValue failed with error: %v", convertErr)
		return
	}

	if !CompareObjectValue(objVal, convertedVal) {
		t.Errorf("ConvertObjectValue failed, objects don't match")
	}
}

func TestSliceObjectValueEncoding(t *testing.T) {
	// Create test SliceObjectValue
	sliceObjVal := &SliceObjectValue{
		Name:    "TestSliceObject",
		PkgPath: "github.com/test/pkg",
		Values: []*ObjectValue{
			{
				ID:      "123",
				Name:    "TestObject1",
				PkgPath: "github.com/test/pkg",
				Fields: []*FieldValue{
					{Name: "id", Value: int64(123)},
					{Name: "name", Value: "test name 1"},
				},
			},
		},
	}

	// Test EncodeSliceObjectValue
	encodedData, err := EncodeSliceObjectValue(sliceObjVal)
	if err != nil {
		t.Errorf("EncodeSliceObjectValue failed with error: %v", err)
		return
	}

	// Test DecodeSliceObjectValue
	decodedVal, decodeErr := DecodeSliceObjectValue(encodedData)
	if decodeErr != nil {
		t.Errorf("DecodeSliceObjectValue failed with error: %v", decodeErr)
		return
	}

	if !CompareSliceObjectValue(sliceObjVal, decodedVal) {
		t.Errorf("Encode/Decode cycle failed, objects don't match")
	}

	// Test ConvertSliceObjectValue
	convertedVal, convertErr := ConvertSliceObjectValue(sliceObjVal)
	if convertErr != nil {
		t.Errorf("ConvertSliceObjectValue failed with error: %v", convertErr)
		return
	}

	if !CompareSliceObjectValue(sliceObjVal, convertedVal) {
		t.Errorf("ConvertSliceObjectValue failed, objects don't match")
	}
}
