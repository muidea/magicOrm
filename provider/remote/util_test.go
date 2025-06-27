package remote

import (
	"reflect"
	"testing"
)

func TestCompareObjectFunctions(t *testing.T) {
	// Create test objects and object values
	obj1 := &Object{
		Name:    "TestObject",
		PkgPath: "github.com/test/pkg",
		Fields: []*Field{
			{
				Name:  "id",
				Type:  &TypeImpl{Name: "int64", Value: 0},
				Spec:  &SpecImpl{FieldName: "id", PrimaryKey: true},
				value: NewValue(int64(123)),
			},
			{
				Name:  "name",
				Type:  &TypeImpl{Name: "string", Value: 1},
				Spec:  &SpecImpl{FieldName: "name"},
				value: NewValue("test name"),
			},
		},
	}

	obj2 := &Object{
		Name:    "TestObject",
		PkgPath: "github.com/test/pkg",
		Fields: []*Field{
			{
				Name:  "id",
				Type:  &TypeImpl{Name: "int64", Value: 0},
				Spec:  &SpecImpl{FieldName: "id", PrimaryKey: true},
				value: NewValue(int64(123)),
			},
			{
				Name:  "name",
				Type:  &TypeImpl{Name: "string", Value: 1},
				Spec:  &SpecImpl{FieldName: "name"},
				value: NewValue("test name"),
			},
		},
	}

	// Create some object values
	objVal1 := &ObjectValue{
		Name:    "TestObjectValue",
		PkgPath: "github.com/test/pkg",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(123)},
			{Name: "name", Value: "test name"},
		},
	}

	objVal2 := &ObjectValue{
		Name:    "TestObjectValue",
		PkgPath: "github.com/test/pkg",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(123)},
			{Name: "name", Value: "test name"},
		},
	}

	// Create slice values
	sliceVal1 := &SliceObjectValue{
		Name:    "TestSliceValue",
		PkgPath: "github.com/test/pkg",
		Values:  []*ObjectValue{objVal1},
	}

	sliceVal2 := &SliceObjectValue{
		Name:    "TestSliceValue",
		PkgPath: "github.com/test/pkg",
		Values:  []*ObjectValue{objVal2},
	}

	// Test compare functions
	tests := []struct {
		name     string
		testFunc func() bool
		expected bool
	}{
		{
			name: "CompareObject with equal objects",
			testFunc: func() bool {
				return CompareObject(obj1, obj2)
			},
			expected: true,
		},
		{
			name: "CompareObject with different objects",
			testFunc: func() bool {
				diffObj := &Object{
					Name:    "DifferentObject",
					PkgPath: "github.com/test/pkg",
					Fields:  obj1.Fields,
				}
				return CompareObject(obj1, diffObj)
			},
			expected: false,
		},
		{
			name: "CompareObjectValue with equal values",
			testFunc: func() bool {
				return CompareObjectValue(objVal1, objVal2)
			},
			expected: true,
		},
		{
			name: "CompareObjectValue with different values",
			testFunc: func() bool {
				diffObjVal := &ObjectValue{
					Name:    "DifferentValue",
					PkgPath: "github.com/test/pkg",
					Fields:  objVal1.Fields,
				}
				return CompareObjectValue(objVal1, diffObjVal)
			},
			expected: false,
		},
		{
			name: "CompareSliceObjectValue with equal values",
			testFunc: func() bool {
				return CompareSliceObjectValue(sliceVal1, sliceVal2)
			},
			expected: true,
		},
		{
			name: "CompareSliceObjectValue with different values",
			testFunc: func() bool {
				diffSliceVal := &SliceObjectValue{
					Name:    "DifferentSlice",
					PkgPath: "github.com/test/pkg",
					Values:  sliceVal1.Values,
				}
				return CompareSliceObjectValue(sliceVal1, diffSliceVal)
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.testFunc()
			if result != test.expected {
				t.Errorf("%s failed, expected %v, got %v", test.name, test.expected, result)
			}
		})
	}
}

func TestEntityReflection(t *testing.T) {
	t.Run("Basic struct reflection", func(t *testing.T) {
		type TestStruct struct {
			ID   int
			Name string
		}

		val := TestStruct{ID: 1, Name: "test"}
		typeInfo := reflect.TypeOf(val)

		if typeInfo.Name() != "TestStruct" {
			t.Errorf("Type name mismatch, expected 'TestStruct', got '%s'", typeInfo.Name())
		}

		if typeInfo.Kind() != reflect.Struct {
			t.Errorf("Type kind mismatch, expected 'struct', got '%s'", typeInfo.Kind())
		}

		for i := 0; i < typeInfo.NumField(); i++ {
			field := typeInfo.Field(i)
			if field.Name == "ID" && field.Type.Kind() != reflect.Int {
				t.Errorf("Field 'ID' has incorrect type, expected 'int', got '%s'", field.Type.Kind())
			}
			if field.Name == "Name" && field.Type.Kind() != reflect.String {
				t.Errorf("Field 'Name' has incorrect type, expected 'string', got '%s'", field.Type.Kind())
			}
		}
	})

	t.Run("Pointer to struct reflection", func(t *testing.T) {
		type TestStruct struct {
			ID   int
			Name string
		}

		val := &TestStruct{ID: 1, Name: "test"}
		typeInfo := reflect.TypeOf(val)

		if typeInfo.Kind() != reflect.Ptr {
			t.Errorf("Type kind mismatch, expected 'ptr', got '%s'", typeInfo.Kind())
		}

		elemType := typeInfo.Elem()
		if elemType.Name() != "TestStruct" {
			t.Errorf("Element type name mismatch, expected 'TestStruct', got '%s'", elemType.Name())
		}

		if elemType.Kind() != reflect.Struct {
			t.Errorf("Element type kind mismatch, expected 'struct', got '%s'", elemType.Kind())
		}
	})
}
