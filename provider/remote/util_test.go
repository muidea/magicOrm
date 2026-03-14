package remote

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/models"
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
					Name:    "TestObjectValue",
					PkgPath: "github.com/test/pkg",
					Fields: []*FieldValue{
						{Name: "id", Value: int64(999)},
						{Name: "name", Value: "test name"},
					},
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

func TestConvertValueHelpers(t *testing.T) {
	intType := &TypeImpl{Name: "int", Value: models.TypeIntegerValue}
	convertedInt, err := convertValue(intType, float64(9))
	if err != nil {
		t.Fatalf("convertValue(int) failed: %v", err)
	}
	if convertedInt != 9 {
		t.Fatalf("convertValue(int) mismatch, got %#v", convertedInt)
	}

	stringSliceType := &TypeImpl{
		Name:  "string",
		Value: models.TypeStringValue,
	}
	convertedSlice, err := convertSliceValue(stringSliceType, []any{"a", "b"})
	if err != nil {
		t.Fatalf("convertSliceValue([]string) failed: %v", err)
	}
	if !reflect.DeepEqual(convertedSlice, []any{"a", "b"}) {
		t.Fatalf("convertSliceValue([]string) mismatch, got %#v", convertedSlice)
	}

	if _, err := convertSliceValue(stringSliceType, "not-slice"); err == nil {
		t.Fatal("convertSliceValue(non-slice) should fail")
	}
}

func TestInitializeValueHelpers(t *testing.T) {
	basicCases := []struct {
		name string
		typ  *TypeImpl
		want any
	}{
		{name: "bool", typ: &TypeImpl{Name: "bool", Value: models.TypeBooleanValue}, want: false},
		{name: "string", typ: &TypeImpl{Name: "string", Value: models.TypeStringValue}, want: ""},
		{name: "datetime", typ: &TypeImpl{Name: "datetime", Value: models.TypeDateTimeValue}, want: ""},
		{name: "int8", typ: &TypeImpl{Name: "int8", Value: models.TypeByteValue}, want: int8(0)},
		{name: "int16", typ: &TypeImpl{Name: "int16", Value: models.TypeSmallIntegerValue}, want: int16(0)},
		{name: "int32", typ: &TypeImpl{Name: "int32", Value: models.TypeInteger32Value}, want: int32(0)},
		{name: "int", typ: &TypeImpl{Name: "int", Value: models.TypeIntegerValue}, want: 0},
		{name: "int64", typ: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, want: int64(0)},
		{name: "uint8", typ: &TypeImpl{Name: "uint8", Value: models.TypePositiveByteValue}, want: uint8(0)},
		{name: "uint16", typ: &TypeImpl{Name: "uint16", Value: models.TypePositiveSmallIntegerValue}, want: uint16(0)},
		{name: "uint32", typ: &TypeImpl{Name: "uint32", Value: models.TypePositiveInteger32Value}, want: uint32(0)},
		{name: "uint", typ: &TypeImpl{Name: "uint", Value: models.TypePositiveIntegerValue}, want: uint(0)},
		{name: "uint64", typ: &TypeImpl{Name: "uint64", Value: models.TypePositiveBigIntegerValue}, want: uint64(0)},
		{name: "float32", typ: &TypeImpl{Name: "float32", Value: models.TypeFloatValue}, want: float32(0)},
		{name: "float64", typ: &TypeImpl{Name: "float64", Value: models.TypeDoubleValue}, want: float64(0)},
	}
	for _, tt := range basicCases {
		if got := getBasicInitValue(tt.typ); !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("getBasicInitValue(%s) mismatch, got %#v want %#v", tt.name, got, tt.want)
		}
	}

	sliceCases := []struct {
		name string
		typ  *TypeImpl
		want any
	}{
		{name: "[]bool", typ: &TypeImpl{Name: "bool", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "bool", Value: models.TypeBooleanValue}}, want: []bool{}},
		{name: "[]int8", typ: &TypeImpl{Name: "int8", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int8", Value: models.TypeByteValue}}, want: []int8{}},
		{name: "[]int16", typ: &TypeImpl{Name: "int16", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int16", Value: models.TypeSmallIntegerValue}}, want: []int16{}},
		{name: "[]int32", typ: &TypeImpl{Name: "int32", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int32", Value: models.TypeInteger32Value}}, want: []int32{}},
		{name: "[]int", typ: &TypeImpl{Name: "int", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int", Value: models.TypeIntegerValue}}, want: []int{}},
		{name: "[]string", typ: &TypeImpl{Name: "string", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "string", Value: models.TypeStringValue}}, want: []string{}},
		{name: "[]datetime", typ: &TypeImpl{Name: "datetime", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "datetime", Value: models.TypeDateTimeValue}}, want: []string{}},
		{name: "[]int64", typ: &TypeImpl{Name: "int64", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}}, want: []int64{}},
		{name: "[]uint8", typ: &TypeImpl{Name: "uint8", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint8", Value: models.TypePositiveByteValue}}, want: []uint8{}},
		{name: "[]uint16", typ: &TypeImpl{Name: "uint16", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint16", Value: models.TypePositiveSmallIntegerValue}}, want: []uint16{}},
		{name: "[]uint32", typ: &TypeImpl{Name: "uint32", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint32", Value: models.TypePositiveInteger32Value}}, want: []uint32{}},
		{name: "[]uint", typ: &TypeImpl{Name: "uint", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint", Value: models.TypePositiveIntegerValue}}, want: []uint{}},
		{name: "[]uint64", typ: &TypeImpl{Name: "uint64", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint64", Value: models.TypePositiveBigIntegerValue}}, want: []uint64{}},
		{name: "[]float32", typ: &TypeImpl{Name: "float32", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "float32", Value: models.TypeFloatValue}}, want: []float32{}},
		{name: "[]float64", typ: &TypeImpl{Name: "float64", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "float64", Value: models.TypeDoubleValue}}, want: []float64{}},
	}
	for _, tt := range sliceCases {
		if got := getSliceInitValue(tt.typ); !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("getSliceInitValue(%s) mismatch, got %#v want %#v", tt.name, got, tt.want)
		}
	}

	structType := &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue}
	if got := getInitializeValue(structType); got.(*ObjectValue).GetPkgPath() != "/vmi" {
		t.Fatalf("getInitializeValue(struct) mismatch, got %#v", got)
	}
	sliceStructType := &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue}}
	if got := getInitializeValue(sliceStructType); got.(*SliceObjectValue).GetPkgPath() != "/vmi" {
		t.Fatalf("getInitializeValue(slice struct) mismatch, got %#v", got)
	}
}

func TestInitializeValueHelperPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("getBasicInitValue(struct) should panic")
		}
	}()
	getBasicInitValue(&TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue})
}

func TestSliceInitializeValueHelperPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("getSliceInitValue(slice struct) should panic")
		}
	}()
	getSliceInitValue(&TypeImpl{Name: "items", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue}})
}

func TestUtilityRewriteAndValidityHelpers(t *testing.T) {
	rawObject := &ObjectValue{Name: "status", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(1)}}}
	srcObject := &ObjectValue{Name: "status", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(2)}}}
	if err := rewriteObjectValue(rawObject, srcObject); err != nil {
		t.Fatalf("rewriteObjectValue failed: %v", err)
	}
	if rawObject.GetFieldValue("id") != int64(2) {
		t.Fatalf("rewriteObjectValue mismatch, got %#v", rawObject)
	}
	if err := rewriteObjectValue(nil, srcObject); err != nil {
		t.Fatalf("rewriteObjectValue(nil) should be ignored, got %v", err)
	}

	rawSlice := &SliceObjectValue{Name: "status", PkgPath: "/vmi", Values: []*ObjectValue{{Name: "status", PkgPath: "/vmi"}}}
	srcSlice := &SliceObjectValue{Name: "status", PkgPath: "/vmi", Values: []*ObjectValue{{Name: "status", PkgPath: "/vmi"}, {Name: "status", PkgPath: "/vmi"}}}
	if err := rewriteSliceObjectValue(rawSlice, srcSlice); err != nil {
		t.Fatalf("rewriteSliceObjectValue failed: %v", err)
	}
	if len(rawSlice.Values) != 2 {
		t.Fatalf("rewriteSliceObjectValue mismatch, got %#v", rawSlice)
	}
	if err := rewriteSliceObjectValue(nil, srcSlice); err != nil {
		t.Fatalf("rewriteSliceObjectValue(nil) should be ignored, got %v", err)
	}

	if !isValid(ObjectValue{Name: "status", PkgPath: "/vmi"}) {
		t.Fatal("isValid(ObjectValue) should be true")
	}
	if !isValid(SliceObjectValue{Name: "status", PkgPath: "/vmi"}) {
		t.Fatal("isValid(SliceObjectValue) should be true")
	}
	if !isZero(&ObjectValue{Name: "status", PkgPath: "/vmi"}) {
		t.Fatal("isZero(empty ObjectValue) should be true")
	}
	if isZero(&SliceObjectValue{Name: "skuInfo", PkgPath: "/vmi", Values: []*ObjectValue{}}) {
		t.Fatal("isZero(explicit empty SliceObjectValue) should be false")
	}

	if got, err := convertValue(&TypeImpl{Name: "string", Value: models.TypeStringValue}, nil); err != nil || got != nil {
		t.Fatalf("convertValue(nil) mismatch, got %#v err=%v", got, err)
	}
	if _, err := convertValue(&TypeImpl{Name: "bool", Value: models.TypeBooleanValue}, map[string]any{"bad": true}); err == nil {
		t.Fatal("convertValue(bool invalid) should fail")
	}
	if got, err := convertSliceValue(&TypeImpl{Name: "string", Value: models.TypeStringValue}, nil); err != nil || got != nil {
		t.Fatalf("convertSliceValue(nil) mismatch, got %#v err=%v", got, err)
	}
}
