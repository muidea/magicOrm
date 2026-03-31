package remote

import (
	"errors"
	"testing"

	"github.com/muidea/magicOrm/models"
)

type fieldTestValidator struct {
	called int
	err    error
}

func (m *fieldTestValidator) Register(k models.Key, fn models.ValidatorFunc) {}

func (m *fieldTestValidator) ValidateValue(val any, directives []models.Directive) error {
	m.called++
	return m.err
}

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

func TestFieldSliceHelpers(t *testing.T) {
	sliceField := &Field{
		Name: "tags",
		Type: &TypeImpl{
			Name:  "string",
			Value: models.TypeSliceValue,
			ElemType: &TypeImpl{
				Name:  "string",
				Value: models.TypeStringValue,
			},
		},
		value: NewValue([]string{"a", "b"}),
	}

	values := sliceField.GetSliceValue()
	if len(values) != 2 || values[0].Get() != "a" || values[1].Get() != "b" {
		t.Fatalf("GetSliceValue mismatch, got %#v", values)
	}

	if err := sliceField.AppendSliceValue("c"); err != nil {
		t.Fatalf("AppendSliceValue failed: %v", err)
	}
	gotValues := sliceField.GetValue().Get().([]string)
	if len(gotValues) != 3 || gotValues[2] != "c" {
		t.Fatalf("AppendSliceValue mismatch, got %#v", gotValues)
	}

	if err := sliceField.AppendSliceValue(nil); err == nil {
		t.Fatal("AppendSliceValue(nil) should fail")
	}

	basicField := &Field{
		Name:  "name",
		Type:  &TypeImpl{Name: "string", Value: models.TypeStringValue},
		value: NewValue("alice"),
	}
	if got := basicField.GetSliceValue(); got != nil {
		t.Fatalf("GetSliceValue(non-slice) should be nil, got %#v", got)
	}
	if err := basicField.AppendSliceValue("x"); err == nil {
		t.Fatal("AppendSliceValue(non-slice) should fail")
	}
}

func TestFieldInnerSetValueAndCopyBranches(t *testing.T) {
	validator := &fieldTestValidator{err: errors.New("rejected")}
	field := &Field{
		Name:           "name",
		Type:           &TypeImpl{Name: "string", Value: models.TypeStringValue},
		Spec:           &SpecImpl{FieldName: "name", Constraint: "req"},
		valueValidator: validator,
	}
	if err := field.innerSetValue("apple", false); err == nil {
		t.Fatal("innerSetValue with failing validator should fail")
	}
	if validator.called != 1 {
		t.Fatalf("validator should be called once, got %d", validator.called)
	}

	if err := field.innerSetValue("apple", true); err != nil {
		t.Fatalf("innerSetValue disableValidator failed: %v", err)
	}
	if got := field.GetValue().Get(); got != "apple" {
		t.Fatalf("innerSetValue disableValidator mismatch, got %#v", got)
	}

	plainField := &Field{
		Name:           "plain",
		Type:           &TypeImpl{Name: "string", Value: models.TypeStringValue},
		valueValidator: validator,
	}
	if err := plainField.innerSetValue("banana", false); err != nil {
		t.Fatalf("innerSetValue without constraints failed: %v", err)
	}

	ptrField := &Field{
		Name: "status",
		Type: &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue, IsPtr: true},
		Spec: &SpecImpl{FieldName: "status"},
	}
	metaCopy, err := ptrField.copy(models.MetaView)
	if err != nil {
		t.Fatalf("Field.copy(meta ptr) failed: %v", err)
	}
	if metaCopy.GetValue().IsValid() {
		t.Fatalf("meta ptr copy should remain invalid, got %#v", metaCopy.GetValue().Get())
	}

	detailDisabled, err := (&Field{
		Name: "expire",
		Type: &TypeImpl{Name: "int", Value: models.TypeIntegerValue},
		Spec: &SpecImpl{FieldName: "expire", ViewDeclare: []models.ViewDeclare{models.MetaView}},
	}).copy(models.DetailView)
	if err != nil {
		t.Fatalf("Field.copy(detail disabled) failed: %v", err)
	}
	if detailDisabled.GetValue().IsValid() {
		t.Fatalf("detail-disabled field should remain invalid, got %#v", detailDisabled.GetValue().Get())
	}

	detailEnabled, err := (&Field{
		Name: "expire",
		Type: &TypeImpl{Name: "int", Value: models.TypeIntegerValue},
		Spec: &SpecImpl{FieldName: "expire", ViewDeclare: []models.ViewDeclare{models.DetailView}},
	}).copy(models.DetailView)
	if err != nil {
		t.Fatalf("Field.copy(detail enabled) failed: %v", err)
	}
	if !detailEnabled.GetValue().IsValid() || detailEnabled.GetValue().Get() != 0 {
		t.Fatalf("detail-enabled field should be initialized, got %#v", detailEnabled.GetValue().Get())
	}

	nilSpecCopy, err := (&Field{
		Name: "id",
		Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
	}).copy(models.MetaView)
	if err != nil {
		t.Fatalf("Field.copy(nil spec) failed: %v", err)
	}
	if nilSpecCopy.GetSpec() == nil || nilSpecCopy.GetSpec().IsPrimaryKey() || nilSpecCopy.GetSpec().GetValueDeclare() != models.Customer {
		t.Fatalf("Field.copy(nil spec) should use emptySpec, got %#v", nilSpecCopy.GetSpec())
	}

	originNonPtr, err := (&Field{
		Name: "count",
		Type: &TypeImpl{Name: "int", Value: models.TypeIntegerValue},
	}).copy(models.OriginView)
	if err != nil {
		t.Fatalf("Field.copy(origin non-ptr) failed: %v", err)
	}
	if !originNonPtr.GetValue().IsValid() || originNonPtr.GetValue().Get() != 0 {
		t.Fatalf("origin non-ptr field should be initialized, got %#v", originNonPtr.GetValue().Get())
	}

	originPtr, err := (&Field{
		Name: "count",
		Type: &TypeImpl{Name: "int", Value: models.TypeIntegerValue, IsPtr: true},
	}).copy(models.OriginView)
	if err != nil {
		t.Fatalf("Field.copy(origin ptr) failed: %v", err)
	}
	if originPtr.GetValue().IsValid() {
		t.Fatalf("origin ptr field should remain invalid, got %#v", originPtr.GetValue().Get())
	}

	unsupportedView, err := (&Field{
		Name:  "count",
		Type:  &TypeImpl{Name: "int", Value: models.TypeIntegerValue},
		Spec:  &SpecImpl{FieldName: "count"},
		value: NewValue(3),
	}).copy(models.ViewDeclare("unsupported"))
	if err != nil {
		t.Fatalf("Field.copy(unsupported view) failed: %v", err)
	}
	if unsupportedView.GetValue().IsValid() {
		t.Fatalf("unsupported view should not initialize value, got %#v", unsupportedView.GetValue().Get())
	}
}

func TestFieldCompareAndFieldValueCopy(t *testing.T) {
	base := &Field{
		Name:     "id",
		ShowName: "ID",
		Type:     &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
		Spec:     &SpecImpl{FieldName: "id", PrimaryKey: true},
	}
	if !compareItem(base, base.copyMust(t, models.OriginView)) {
		t.Fatal("compareItem identical field should be true")
	}
	if compareItem(base, &Field{Name: "other", ShowName: "ID", Type: base.Type.Copy(), Spec: base.Spec.Copy()}) {
		t.Fatal("compareItem different name should be false")
	}
	if compareItem(base, &Field{Name: "id", ShowName: "Other", Type: base.Type.Copy(), Spec: base.Spec.Copy()}) {
		t.Fatal("compareItem different showName should be false")
	}
	if compareItem(base, &Field{Name: "id", ShowName: "ID", Type: &TypeImpl{Name: "string", Value: models.TypeStringValue}, Spec: base.Spec.Copy()}) {
		t.Fatal("compareItem different type should be false")
	}
	if compareItem(base, &Field{Name: "id", ShowName: "ID", Type: base.Type.Copy(), Spec: &SpecImpl{FieldName: "id"}}) {
		t.Fatal("compareItem different spec should be false")
	}

	nilValue := (&FieldValue{Name: "name"}).copy()
	if nilValue.Value != nil {
		t.Fatalf("FieldValue.copy(nil) mismatch, got %#v", nilValue.Value)
	}

	objectValue := (&FieldValue{Name: "status", Value: &ObjectValue{Name: "status", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(1)}}}}).copy()
	if !CompareObjectValue(objectValue.Value.(*ObjectValue), &ObjectValue{Name: "status", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(1)}}}) {
		t.Fatalf("FieldValue.copy(object) mismatch, got %#v", objectValue.Value)
	}

	sliceObjectValue := (&FieldValue{Name: "items", Value: &SliceObjectValue{Name: "item", PkgPath: "/vmi", Values: []*ObjectValue{{Name: "item", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(1)}}}}}}).copy()
	if !CompareSliceObjectValue(sliceObjectValue.Value.(*SliceObjectValue), &SliceObjectValue{Name: "item", PkgPath: "/vmi", Values: []*ObjectValue{{Name: "item", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(1)}}}}}) {
		t.Fatalf("FieldValue.copy(slice object) mismatch, got %#v", sliceObjectValue.Value)
	}

	basicValue := (&FieldValue{Name: "tags", Value: []string{"a", "b"}}).copy()
	tags, ok := basicValue.Value.([]string)
	if !ok || len(tags) != 2 || tags[0] != "a" || tags[1] != "b" {
		t.Fatalf("FieldValue.copy(basic) mismatch, got %#v", basicValue.Value)
	}
}

func (s *Field) copyMust(t *testing.T, viewSpec models.ViewDeclare) *Field {
	t.Helper()

	ret, err := s.copy(viewSpec)
	if err != nil {
		t.Fatalf("Field.copy failed: %v", err)
	}
	return ret
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
			name: "Valid Snowflake field",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue},
				Spec: &SpecImpl{FieldName: "id", ValueDeclare: models.Snowflake},
			},
			expectedToPass: true,
		},
		{
			name: "Invalid Snowflake on string field",
			field: &Field{
				Name: "id",
				Type: &TypeImpl{Name: "string", Value: models.TypeStringValue},
				Spec: &SpecImpl{FieldName: "id", ValueDeclare: models.Snowflake},
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

func TestFieldValueSetSupportsBasicCollectionOperand(t *testing.T) {
	fieldVal := &FieldValue{Name: "ids"}
	fieldVal.Set([]any{int64(1), int64(2), int64(3)})

	got, ok := fieldVal.Get().([]any)
	if !ok {
		t.Fatalf("FieldValue.Set(collection) should preserve []any, got %T", fieldVal.Get())
	}
	if len(got) != 3 || got[0] != int64(1) || got[1] != int64(2) || got[2] != int64(3) {
		t.Fatalf("FieldValue.Set(collection) mismatch, got %#v", got)
	}
	if !fieldVal.IsValid() || fieldVal.IsZero() {
		t.Fatalf("FieldValue.Set(collection) validity mismatch, valid=%v zero=%v", fieldVal.IsValid(), fieldVal.IsZero())
	}

	fieldVal.Set([]any{})
	emptyCollection, ok := fieldVal.Get().([]any)
	if !ok {
		t.Fatalf("FieldValue.Set(empty collection) should preserve []any, got %T", fieldVal.Get())
	}
	if len(emptyCollection) != 0 {
		t.Fatalf("FieldValue.Set(empty collection) mismatch, got %#v", emptyCollection)
	}
	if !fieldVal.IsValid() || fieldVal.IsZero() {
		t.Fatalf("FieldValue.Set(empty collection) validity mismatch, valid=%v zero=%v", fieldVal.IsValid(), fieldVal.IsZero())
	}
}
