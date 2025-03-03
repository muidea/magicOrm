package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
)

type FieldTestStruct struct {
	ID          int     `orm:"id key auto" view:"detail,lite"`
	Name        string  `orm:"name" view:"detail,lite"`
	Value       float64 `orm:"value" view:"detail"`
	Description string  `orm:"description"`
}

func TestFieldMethods(t *testing.T) {
	testEntity := &FieldTestStruct{
		ID:          1,
		Name:        "Test Name",
		Value:       123.45,
		Description: "Test Description",
	}

	entityType := reflect.TypeOf(testEntity).Elem()
	entityValue := reflect.ValueOf(testEntity).Elem()

	// Test ID field
	idField := entityType.Field(0)
	idValue := entityValue.Field(0)

	fieldSpec, specErr := NewSpec(idField.Tag)
	if specErr != nil {
		t.Errorf("NewSpec failed: %s", specErr.Error())
		return
	}

	fieldType, typeErr := NewType(idValue.Type())
	if typeErr != nil {
		t.Errorf("NewType failed: %s", typeErr.Error())
		return
	}

	field := &field{
		index:    0,
		name:     "id",
		typePtr:  fieldType,
		specPtr:  fieldSpec,
		valuePtr: NewValue(idValue),
	}

	// Test basic field properties
	if field.GetIndex() != 0 {
		t.Errorf("GetIndex failed, expected: 0, got: %d", field.GetIndex())
	}

	if field.GetName() != "id" {
		t.Errorf("GetName failed, expected: id, got: %s", field.GetName())
	}

	// Test IsPrimaryKey
	if !field.IsPrimaryKey() {
		t.Errorf("IsPrimaryKey failed, expected: true, got: false")
	}

	// Test IsBasic
	if !field.IsBasic() {
		t.Errorf("IsBasic failed, expected: true, got: false")
	}

	// Test IsStruct
	if field.IsStruct() {
		t.Errorf("IsStruct failed, expected: false, got: true")
	}

	// Test IsSlice
	if field.IsSlice() {
		t.Errorf("IsSlice failed, expected: false, got: true")
	}

	// Test IsPtrType
	if field.IsPtrType() {
		t.Errorf("IsPtrType failed, expected: false, got: true")
	}

	// Test GetSpec
	spec := field.GetSpec()
	if !spec.IsPrimaryKey() {
		t.Errorf("GetSpec.IsPrimaryKey failed, expected: true, got: false")
	}

	if spec.GetValueDeclare() != model.AutoIncrement {
		t.Errorf("GetSpec.GetValueDeclare failed, expected: AutoIncrement, got: %v", spec.GetValueDeclare())
	}

	// Test Field value
	value := field.GetValue()
	if !value.IsValid() {
		t.Errorf("GetValue.IsValid failed, expected: true, got: false")
	}

	if value.IsZero() {
		t.Errorf("GetValue.IsZero failed, expected: false, got: true")
	}

	// Test for view tags
	structField := entityType.Field(0) // Get the ID field which has the view tag
	viewDecl := getViewItems(string(structField.Tag.Get("view")))
	if len(viewDecl) != 2 {
		t.Errorf("View declarations mismatch, expected 2 views, got: %d", len(viewDecl))
	} else {
		containsDetail := false
		containsLite := false
		for _, v := range viewDecl {
			if v == "detail" {
				containsDetail = true
			}
			if v == "lite" {
				containsLite = true
			}
		}

		if !containsDetail || !containsLite {
			t.Errorf("View tag specification failed, expected 'detail' and 'lite', got: %v", viewDecl)
		}
	}
}

func TestFieldVerify(t *testing.T) {
	// Test verification of different field types
	testEntity := &FieldTestStruct{
		ID:          1,
		Name:        "Test Name",
		Value:       123.45,
		Description: "Test Description",
	}

	entityType := reflect.TypeOf(testEntity).Elem()
	entityValue := reflect.ValueOf(testEntity).Elem()

	// Test each field
	for i := 0; i < entityType.NumField(); i++ {
		structField := entityType.Field(i)
		fieldValue := entityValue.Field(i)

		fieldSpec, specErr := NewSpec(structField.Tag)
		if specErr != nil {
			t.Errorf("NewSpec failed: %s", specErr.Error())
			continue
		}

		fieldType, typeErr := NewType(fieldValue.Type())
		if typeErr != nil {
			t.Errorf("NewType failed: %s", typeErr.Error())
			continue
		}

		field := &field{
			index:    i,
			name:     fieldSpec.GetFieldName(),
			typePtr:  fieldType,
			specPtr:  fieldSpec,
			valuePtr: NewValue(fieldValue),
		}

		err := field.verify()
		if err != nil {
			t.Errorf("field.verify failed for field %s: %s", structField.Name, err.Error())
		}
	}
}

func TestFieldCopy(t *testing.T) {
	testEntity := &FieldTestStruct{
		ID:          1,
		Name:        "Test Name",
		Value:       123.45,
		Description: "Test Description",
	}

	entityType := reflect.TypeOf(testEntity).Elem()
	entityValue := reflect.ValueOf(testEntity).Elem()

	// Get a field
	structField := entityType.Field(0)
	fieldValue := entityValue.Field(0)

	fieldSpec, specErr := NewSpec(structField.Tag)
	if specErr != nil {
		t.Errorf("NewSpec failed: %s", specErr.Error())
		return
	}

	fieldType, typeErr := NewType(fieldValue.Type())
	if typeErr != nil {
		t.Errorf("NewType failed: %s", typeErr.Error())
		return
	}

	originalField := &field{
		index:    0,
		name:     fieldSpec.GetFieldName(),
		typePtr:  fieldType,
		specPtr:  fieldSpec,
		valuePtr: NewValue(fieldValue),
	}

	// Copy the field
	copiedField := originalField.copy(false)

	// Check if properties match
	if copiedField.GetIndex() != originalField.GetIndex() {
		t.Errorf("copy failed: index mismatch")
	}

	if copiedField.GetName() != originalField.GetName() {
		t.Errorf("copy failed: name mismatch")
	}

	if copiedField.IsPrimaryKey() != originalField.IsPrimaryKey() {
		t.Errorf("copy failed: primary key flag mismatch")
	}

	// Check if value was preserved
	if copiedField.GetValue().IsZero() != originalField.GetValue().IsZero() {
		t.Errorf("copy failed: value zero state mismatch")
	}

	// Now test copy with reset
	resetField := originalField.copy(true)

	// Value should be zero
	if !resetField.GetValue().IsZero() {
		t.Errorf("copy with reset failed: value should be zero")
	}
}

func TestFieldGetterSetter(t *testing.T) {
	// Create test struct
	type TestStruct struct {
		ID       int        `orm:"id key"`
		Name     string     `orm:"name"`
		Active   bool       `orm:"active"`
		Created  time.Time  `orm:"created"`
		Modified *time.Time `orm:"modified"`
	}

	testEntity := TestStruct{
		ID:      1,
		Name:    "test",
		Active:  true,
		Created: time.Now(),
	}

	// Get model
	testModel, err := GetEntityModel(testEntity)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test GetField and field properties
	fields := []struct {
		name      string
		isPrimary bool
		isBasic   bool
		isStruct  bool
		isSlice   bool
		isPtrType bool
	}{
		{"id", true, true, false, false, false},
		{"name", false, true, false, false, false},
		{"active", false, true, false, false, false},
		{"created", false, true, false, false, false},
		{"modified", false, true, false, false, true},
	}

	for _, f := range fields {
		field := testModel.GetField(f.name)
		if field == nil {
			t.Errorf("GetField(%s) returned nil", f.name)
			continue
		}

		// Test field name
		if field.GetName() != f.name {
			t.Errorf("Field.GetName() expected: %s, got: %s", f.name, field.GetName())
		}

		// Test primary key
		if field.IsPrimaryKey() != f.isPrimary {
			t.Errorf("Field.IsPrimaryKey() for %s, expected: %v, got: %v",
				f.name, f.isPrimary, field.IsPrimaryKey())
		}

		// Test IsBasic
		if field.IsBasic() != f.isBasic {
			t.Errorf("Field.IsBasic() for %s, expected: %v, got: %v",
				f.name, f.isBasic, field.IsBasic())
		}

		// Test IsStruct
		if field.IsStruct() != f.isStruct {
			t.Errorf("Field.IsStruct() for %s, expected: %v, got: %v",
				f.name, f.isStruct, field.IsStruct())
		}

		// Test IsSlice
		if field.IsSlice() != f.isSlice {
			t.Errorf("Field.IsSlice() for %s, expected: %v, got: %v",
				f.name, f.isSlice, field.IsSlice())
		}

		// Test IsPtrType
		if field.IsPtrType() != f.isPtrType {
			t.Errorf("Field.IsPtrType() for %s, expected: %v, got: %v",
				f.name, f.isPtrType, field.IsPtrType())
		}
	}
}

func TestFieldSetValue(t *testing.T) {
	// Create test struct
	type TestStruct struct {
		ID   int    `orm:"id key"`
		Name string `orm:"name"`
	}

	testEntity := TestStruct{ID: 1, Name: "test"}

	// Get model
	testModel, err := GetEntityModel(testEntity)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Get the ID field
	idField := testModel.GetField("id")
	if idField == nil {
		t.Errorf("GetField(id) returned nil")
		return
	}

	// Create new value
	newID := 2
	newIDValue, err := GetEntityValue(newID)
	if err != nil {
		t.Errorf("GetEntityValue failed: %s", err.Error())
		return
	}

	// Set new value
	idField.SetValue(newIDValue)

	// Get and verify new value
	fieldValue := idField.GetValue()
	if fieldValue == nil {
		t.Errorf("GetValue() returned nil after SetValue")
		return
	}

	val := fieldValue.Get().(reflect.Value)
	if val.Int() != int64(newID) {
		t.Errorf("Field value after SetValue, expected: %d, got: %d", newID, val.Int())
	}
}

func TestFieldTypes(t *testing.T) {
	// Define test struct with various field types
	type NestedStruct struct {
		Value string `orm:"value"`
	}

	type TestAllFieldTypes struct {
		ID         int           `orm:"id key"`
		Int8Val    int8          `orm:"int8Val"`
		Int16Val   int16         `orm:"int16Val"`
		Int32Val   int32         `orm:"int32Val"`
		Int64Val   int64         `orm:"int64Val"`
		UintVal    uint          `orm:"uintVal"`
		Uint8Val   uint8         `orm:"uint8Val"`
		Uint16Val  uint16        `orm:"uint16Val"`
		Uint32Val  uint32        `orm:"uint32Val"`
		Uint64Val  uint64        `orm:"uint64Val"`
		Float32Val float32       `orm:"float32Val"`
		Float64Val float64       `orm:"float64Val"`
		BoolVal    bool          `orm:"boolVal"`
		StringVal  string        `orm:"stringVal"`
		TimeVal    time.Time     `orm:"timeVal"`
		StructVal  NestedStruct  `orm:"structVal"`
		SliceVal   []int         `orm:"sliceVal"`
		PtrVal     *string       `orm:"ptrVal"`
		StructPtr  *NestedStruct `orm:"structPtr"`
		SlicePtr   *[]int        `orm:"slicePtr"`
	}

	testEntity := TestAllFieldTypes{
		ID:        1,
		StringVal: "test",
		TimeVal:   time.Now(),
		StructVal: NestedStruct{Value: "nested"},
		SliceVal:  []int{1, 2, 3},
	}

	// Get model
	testModel, err := GetEntityModel(testEntity)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Verify all fields exist and have correct types
	allFields := testModel.GetFields()
	if len(allFields) != 20 { // Should match the number of fields in TestAllFieldTypes
		t.Errorf("Expected %d fields, got %d", 20, len(allFields))
	}

	// Test basic field type functions
	testBasic := func(field model.Field, name string, isPtr bool) {
		if field.GetName() != name {
			t.Errorf("Field name mismatch, expected: %s, got: %s", name, field.GetName())
		}

		if field.IsPtrType() != isPtr {
			t.Errorf("Field.IsPtrType() for %s, expected: %v, got: %v",
				name, isPtr, field.IsPtrType())
		}

		fieldType := field.GetType()
		if fieldType == nil {
			t.Errorf("Field.GetType() for %s returned nil", name)
			return
		}
	}

	// Check specific fields
	testFields := map[string]struct {
		isPtr    bool
		isSlice  bool
		isStruct bool
	}{
		"id":        {false, false, false},
		"stringVal": {false, false, false},
		"structVal": {false, false, true},
		"sliceVal":  {false, true, false},
		"ptrVal":    {true, false, false},
		"structPtr": {true, false, true},
		"slicePtr":  {true, true, false},
	}

	for name, expected := range testFields {
		field := testModel.GetField(name)
		if field == nil {
			t.Errorf("GetField(%s) returned nil", name)
			continue
		}

		testBasic(field, name, expected.isPtr)

		if field.IsSlice() != expected.isSlice {
			t.Errorf("Field.IsSlice() for %s, expected: %v, got: %v",
				name, expected.isSlice, field.IsSlice())
		}

		if field.IsStruct() != expected.isStruct {
			t.Errorf("Field.IsStruct() for %s, expected: %v, got: %v",
				name, expected.isStruct, field.IsStruct())
		}
	}
}
