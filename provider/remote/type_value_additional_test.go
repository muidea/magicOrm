package remote

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/models"
)

func TestValueImplUnpackAndAppend(t *testing.T) {
	nilValue := &ValueImpl{}
	if unpacked := nilValue.UnpackValue(); unpacked != nil {
		t.Fatalf("nil value should unpack to nil, got %#v", unpacked)
	}
	if err := nilValue.Append(1); err == nil {
		t.Fatalf("Append on nil value should fail")
	}

	basicSlice := NewValue([]int{1, 2})
	unpackedBasic := basicSlice.UnpackValue()
	if len(unpackedBasic) != 2 || unpackedBasic[0].Get() != 1 || unpackedBasic[1].Get() != 2 {
		t.Fatalf("basic slice unpack mismatch, got %#v", unpackedBasic)
	}
	if err := basicSlice.Append(3); err != nil {
		t.Fatalf("Append basic slice failed: %v", err)
	}
	if got := basicSlice.Get(); len(got.([]int)) != 3 || got.([]int)[2] != 3 {
		t.Fatalf("basic slice append mismatch, got %#v", got)
	}
	if err := basicSlice.Append("illegal"); err == nil {
		t.Fatalf("Append with mismatched basic type should fail")
	}

	objectValue := &ObjectValue{
		ID:      "1",
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(1)},
		},
	}
	objectValueImpl := NewValue(objectValue)
	unpackedObject := objectValueImpl.UnpackValue()
	if len(unpackedObject) != 1 {
		t.Fatalf("object value should unpack as one item, got %#v", unpackedObject)
	}
	gotObject, ok := unpackedObject[0].Get().(*ObjectValue)
	if !ok || !CompareObjectValue(objectValue, gotObject) {
		t.Fatalf("object unpack mismatch, got %#v", unpackedObject[0].Get())
	}

	sliceObjectValue := NewValue(&SliceObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Values: []*ObjectValue{
			objectValue.Copy(),
		},
	})
	unpackedSliceObject := sliceObjectValue.UnpackValue()
	if len(unpackedSliceObject) != 1 {
		t.Fatalf("slice object value should unpack items, got %#v", unpackedSliceObject)
	}
	gotSliceObject, ok := unpackedSliceObject[0].Get().(*ObjectValue)
	if !ok || !CompareObjectValue(objectValue, gotSliceObject) {
		t.Fatalf("slice object unpack mismatch, got %#v", unpackedSliceObject[0].Get())
	}
	if err := sliceObjectValue.Append(&ObjectValue{ID: "2", Name: "status", PkgPath: "/vmi"}); err != nil {
		t.Fatalf("Append object value failed: %v", err)
	}
	if len(sliceObjectValue.Get().(*SliceObjectValue).Values) != 2 {
		t.Fatalf("slice object append mismatch, got %#v", sliceObjectValue.Get())
	}
	if err := sliceObjectValue.Append(&ObjectValue{Name: "status", PkgPath: "/other"}); err == nil {
		t.Fatalf("Append with mismatched pkgPath should fail")
	}
	if err := sliceObjectValue.Append("illegal"); err == nil {
		t.Fatalf("Append with non object value should fail")
	}

	valueFormSlice := NewValue(SliceObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Values: []*ObjectValue{
			objectValue.Copy(),
		},
	})
	unpackedValueFormSlice := valueFormSlice.UnpackValue()
	if len(unpackedValueFormSlice) != 1 {
		t.Fatalf("value-form slice object should unpack items, got %#v", unpackedValueFormSlice)
	}
}

func TestValueImplSetRewritePaths(t *testing.T) {
	objectValue := &ValueImpl{value: &ObjectValue{
		ID:      "1",
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(1)},
		},
	}}
	if err := objectValue.Set(ObjectValue{
		ID:      "2",
		Name:    "status",
		PkgPath: "/vmi",
		Fields: []*FieldValue{
			{Name: "id", Value: int64(2)},
		},
	}); err != nil {
		t.Fatalf("Set(ObjectValue) rewrite failed: %v", err)
	}
	if got := objectValue.Get().(*ObjectValue); got.ID != "1" || got.GetFieldValue("id") != int64(2) {
		t.Fatalf("ObjectValue rewrite mismatch, got %#v", got)
	}

	sliceValue := &ValueImpl{value: &SliceObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Values: []*ObjectValue{
			{ID: "1", Name: "status", PkgPath: "/vmi"},
		},
	}}
	if err := sliceValue.Set(SliceObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Values: []*ObjectValue{
			{ID: "2", Name: "status", PkgPath: "/vmi"},
			{ID: "3", Name: "status", PkgPath: "/vmi"},
		},
	}); err != nil {
		t.Fatalf("Set(SliceObjectValue) rewrite failed: %v", err)
	}
	if got := sliceValue.Get().(*SliceObjectValue); len(got.Values) != 2 || got.Values[0].ID != "2" || got.Values[1].ID != "3" {
		t.Fatalf("SliceObjectValue rewrite mismatch, got %#v", got)
	}

	nilObjectValue := &ValueImpl{}
	if err := nilObjectValue.Set(&ObjectValue{Name: "status", PkgPath: "/vmi"}); err != nil {
		t.Fatalf("Set(*ObjectValue) on nil holder failed: %v", err)
	}
	if _, ok := nilObjectValue.Get().(*ObjectValue); !ok {
		t.Fatalf("Set(*ObjectValue) on nil holder mismatch, got %#v", nilObjectValue.Get())
	}

	nilSliceObjectValue := &ValueImpl{}
	if err := nilSliceObjectValue.Set(&SliceObjectValue{Name: "status", PkgPath: "/vmi"}); err != nil {
		t.Fatalf("Set(*SliceObjectValue) on nil holder failed: %v", err)
	}
	if _, ok := nilSliceObjectValue.Get().(*SliceObjectValue); !ok {
		t.Fatalf("Set(*SliceObjectValue) on nil holder mismatch, got %#v", nilSliceObjectValue.Get())
	}

	if err := sliceValue.Set(nil); err != nil {
		t.Fatalf("Set(nil) failed: %v", err)
	}
	if sliceValue.Get() != nil {
		t.Fatalf("Set(nil) should clear value, got %#v", sliceValue.Get())
	}

	if err := (&ValueImpl{}).Set(func() {}); err == nil {
		t.Fatal("Set(illegal value) should fail")
	}
}

func TestValueImplNewValueAndCopyBranches(t *testing.T) {
	objectVal := NewValue(ObjectValue{Name: "status", PkgPath: "/vmi"})
	if _, ok := objectVal.Get().(*ObjectValue); !ok {
		t.Fatalf("NewValue(ObjectValue) should normalize to *ObjectValue, got %#v", objectVal.Get())
	}

	sliceObjectVal := NewValue(SliceObjectValue{Name: "status", PkgPath: "/vmi"})
	if _, ok := sliceObjectVal.Get().(*SliceObjectValue); !ok {
		t.Fatalf("NewValue(SliceObjectValue) should normalize to *SliceObjectValue, got %#v", sliceObjectVal.Get())
	}

	defer func() {
		if recover() == nil {
			t.Fatal("NewValue(func) should panic")
		}
	}()
	_ = NewValue(func() {})
}

func TestValueImplCopyBranches(t *testing.T) {
	nilValue, err := (&ValueImpl{}).copy()
	if err != nil || nilValue.Get() != nil {
		t.Fatalf("copy(nil) mismatch, got %#v err=%v", nilValue, err)
	}

	objectCopy, err := (&ValueImpl{value: &ObjectValue{Name: "status", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(1)}}}}).copy()
	if err != nil {
		t.Fatalf("copy(*ObjectValue) failed: %v", err)
	}
	if !CompareObjectValue(objectCopy.Get().(*ObjectValue), &ObjectValue{Name: "status", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(1)}}}) {
		t.Fatalf("copy(*ObjectValue) mismatch, got %#v", objectCopy.Get())
	}

	sliceCopy, err := (&ValueImpl{value: &SliceObjectValue{Name: "status", PkgPath: "/vmi", Values: []*ObjectValue{{Name: "status", PkgPath: "/vmi"}}}}).copy()
	if err != nil {
		t.Fatalf("copy(*SliceObjectValue) failed: %v", err)
	}
	if !CompareSliceObjectValue(sliceCopy.Get().(*SliceObjectValue), &SliceObjectValue{Name: "status", PkgPath: "/vmi", Values: []*ObjectValue{{Name: "status", PkgPath: "/vmi"}}}) {
		t.Fatalf("copy(*SliceObjectValue) mismatch, got %#v", sliceCopy.Get())
	}

	basicCopy, err := (&ValueImpl{value: []string{"a", "b"}}).copy()
	if err != nil {
		t.Fatalf("copy(basic slice) failed: %v", err)
	}
	if got := basicCopy.Get().([]string); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("copy(basic slice) mismatch, got %#v", basicCopy.Get())
	}
}

func TestTypeImplInterfaceAdditionalBranches(t *testing.T) {
	intType := &TypeImpl{Name: "int", Value: models.TypeIntegerValue}
	intVal, err := intType.Interface(float64(12))
	if err != nil {
		t.Fatalf("Interface(int) failed: %v", err)
	}
	if got := intVal.Get(); got != 12 {
		t.Fatalf("Interface(int) mismatch, got %#v", got)
	}

	stringSliceType := &TypeImpl{
		Name:  "string",
		Value: models.TypeSliceValue,
		ElemType: &TypeImpl{
			Name:  "string",
			Value: models.TypeStringValue,
		},
	}
	stringSliceVal, err := stringSliceType.Interface([]any{"a", "b"})
	if err != nil {
		t.Fatalf("Interface([]string) failed: %v", err)
	}
	gotStrings, ok := stringSliceVal.Get().([]string)
	if !ok || len(gotStrings) != 2 || gotStrings[0] != "a" || gotStrings[1] != "b" {
		t.Fatalf("Interface([]string) mismatch, got %#v", stringSliceVal.Get())
	}

	structType := &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue, Description: "status desc"}
	if structType.GetDescription() != "status desc" {
		t.Fatalf("GetDescription mismatch, got %q", structType.GetDescription())
	}
	structVal, err := structType.Interface(map[string]any{"id": int64(9), "name": "published"})
	if err != nil {
		t.Fatalf("Interface(struct map) failed: %v", err)
	}
	gotStruct, ok := structVal.Get().(*ObjectValue)
	if !ok || gotStruct.GetFieldValue("id") != int64(9) || gotStruct.GetFieldValue("name") != "published" {
		t.Fatalf("Interface(struct map) mismatch, got %#v", structVal.Get())
	}

	sliceStructType := &TypeImpl{
		Name:    "skuInfo",
		PkgPath: "/vmi/product",
		Value:   models.TypeSliceValue,
		ElemType: &TypeImpl{
			Name:    "skuInfo",
			PkgPath: "/vmi/product",
			Value:   models.TypeStructValue,
		},
	}
	sliceStructVal, err := sliceStructType.Interface([]map[string]any{{"sku": "sku-001"}, {"sku": "sku-002"}})
	if err != nil {
		t.Fatalf("Interface(slice struct) failed: %v", err)
	}
	gotSliceStruct, ok := sliceStructVal.Get().(*SliceObjectValue)
	if !ok || len(gotSliceStruct.Values) != 2 || gotSliceStruct.Values[0].GetFieldValue("sku") != "sku-001" || gotSliceStruct.Values[1].GetFieldValue("sku") != "sku-002" {
		t.Fatalf("Interface(slice struct) mismatch, got %#v", sliceStructVal.Get())
	}

	nilStructVal, err := structType.Interface(nil)
	if err != nil {
		t.Fatalf("Interface(nil struct) failed: %v", err)
	}
	gotNilStruct, ok := nilStructVal.Get().(*ObjectValue)
	if !ok || gotNilStruct.Name != "status" || gotNilStruct.PkgPath != "/vmi" {
		t.Fatalf("Interface(nil struct) mismatch, got %#v", nilStructVal.Get())
	}

	nilSliceStructVal, err := sliceStructType.Interface(nil)
	if err != nil {
		t.Fatalf("Interface(nil slice struct) failed: %v", err)
	}
	gotNilSliceStruct, ok := nilSliceStructVal.Get().(*SliceObjectValue)
	if !ok || gotNilSliceStruct.Name != "skuInfo" || gotNilSliceStruct.PkgPath != "/vmi/product" || gotNilSliceStruct.Values != nil {
		t.Fatalf("Interface(nil slice struct) mismatch, got %#v", nilSliceStructVal.Get())
	}

	boolType := &TypeImpl{Name: "bool", Value: models.TypeBooleanValue}
	if _, err := boolType.Interface(map[string]any{"v": true}); err == nil {
		t.Fatalf("Interface(bool invalid) should fail")
	}
}

func TestTypeImplInterfaceNumericConversions(t *testing.T) {
	tests := []struct {
		name     string
		typ      *TypeImpl
		input    any
		expected any
	}{
		{name: "int8", typ: &TypeImpl{Name: "int8", Value: models.TypeByteValue}, input: 8, expected: int8(8)},
		{name: "int16", typ: &TypeImpl{Name: "int16", Value: models.TypeSmallIntegerValue}, input: 16, expected: int16(16)},
		{name: "int32", typ: &TypeImpl{Name: "int32", Value: models.TypeInteger32Value}, input: 32, expected: int32(32)},
		{name: "int64", typ: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, input: 64, expected: int64(64)},
		{name: "uint8", typ: &TypeImpl{Name: "uint8", Value: models.TypePositiveByteValue}, input: 8, expected: uint8(8)},
		{name: "uint16", typ: &TypeImpl{Name: "uint16", Value: models.TypePositiveSmallIntegerValue}, input: 16, expected: uint16(16)},
		{name: "uint32", typ: &TypeImpl{Name: "uint32", Value: models.TypePositiveInteger32Value}, input: 32, expected: uint32(32)},
		{name: "uint", typ: &TypeImpl{Name: "uint", Value: models.TypePositiveIntegerValue}, input: 42, expected: uint(42)},
		{name: "uint64", typ: &TypeImpl{Name: "uint64", Value: models.TypePositiveBigIntegerValue}, input: 64, expected: uint64(64)},
		{name: "float32", typ: &TypeImpl{Name: "float32", Value: models.TypeFloatValue}, input: 3.5, expected: float32(3.5)},
		{name: "float64", typ: &TypeImpl{Name: "float64", Value: models.TypeDoubleValue}, input: 6, expected: float64(6)},
		{name: "datetime", typ: &TypeImpl{Name: "datetime", Value: models.TypeDateTimeValue}, input: 123, expected: "123"},
		{name: "string ptr", typ: &TypeImpl{Name: "string", Value: models.TypeStringValue, IsPtr: true}, input: 123, expected: func() any { s := "123"; return &s }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.typ.Interface(tt.input)
			if err != nil {
				t.Fatalf("Interface(%s) failed: %v", tt.name, err)
			}

			got := value.Get()
			if reflect.TypeOf(tt.expected).Kind() == reflect.Ptr {
				gotVal := reflect.ValueOf(got)
				wantVal := reflect.ValueOf(tt.expected)
				if gotVal.Kind() != reflect.Ptr || gotVal.IsNil() || wantVal.IsNil() || gotVal.Elem().Interface() != wantVal.Elem().Interface() {
					t.Fatalf("Interface(%s) mismatch, got %#v want %#v", tt.name, got, tt.expected)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("Interface(%s) mismatch, got %#v want %#v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestCompareTypeAdditionalBranches(t *testing.T) {
	base := &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue, IsPtr: true}
	if !compareType(base, base.Copy()) {
		t.Fatalf("compareType identical types should be true")
	}
	if compareType(base, &TypeImpl{Name: "other", PkgPath: "/vmi", Value: models.TypeStructValue, IsPtr: true}) {
		t.Fatalf("compareType different names should be false")
	}
	if compareType(base, &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue, IsPtr: false}) {
		t.Fatalf("compareType different ptr flags should be false")
	}
	if compareType(
		&TypeImpl{Name: "items", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue}},
		&TypeImpl{Name: "items", Value: models.TypeSliceValue},
	) {
		t.Fatalf("compareType elemType mismatch should be false")
	}
}

func TestTypeImplInterfacePointerSlicesAndStructSources(t *testing.T) {
	ptrSliceCases := []struct {
		name     string
		typ      *TypeImpl
		input    any
		expected any
	}{
		{
			name: "[]int8 ptr",
			typ: &TypeImpl{Name: "int8", Value: models.TypeSliceValue, IsPtr: true, ElemType: &TypeImpl{Name: "int8", Value: models.TypeByteValue}},
			input: []any{1, 2},
			expected: func() any {
				v := []int8{1, 2}
				return &v
			}(),
		},
		{
			name: "[]uint ptr",
			typ: &TypeImpl{Name: "uint", Value: models.TypeSliceValue, IsPtr: true, ElemType: &TypeImpl{Name: "uint", Value: models.TypePositiveIntegerValue}},
			input: []any{1, 2},
			expected: func() any {
				v := []uint{1, 2}
				return &v
			}(),
		},
		{
			name: "[]float32 ptr",
			typ: &TypeImpl{Name: "float32", Value: models.TypeSliceValue, IsPtr: true, ElemType: &TypeImpl{Name: "float32", Value: models.TypeFloatValue}},
			input: []any{1.5, 2.5},
			expected: func() any {
				v := []float32{1.5, 2.5}
				return &v
			}(),
		},
	}
	for _, tt := range ptrSliceCases {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.typ.Interface(tt.input)
			if err != nil {
				t.Fatalf("Interface(%s) failed: %v", tt.name, err)
			}
			got := reflect.ValueOf(value.Get())
			want := reflect.ValueOf(tt.expected)
			if got.Kind() != reflect.Ptr || got.IsNil() || !reflect.DeepEqual(got.Elem().Interface(), want.Elem().Interface()) {
				t.Fatalf("Interface(%s) mismatch, got %#v want %#v", tt.name, value.Get(), tt.expected)
			}
		})
	}

	objectSeed := &ObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Fields:  []*FieldValue{{Name: "id", Value: int64(1)}, {Name: "name", Value: "ready"}},
	}
	structType := &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue}
	gotObject, err := structType.convertRawStruct(objectSeed)
	if err != nil {
		t.Fatalf("convertRawStruct(*ObjectValue) failed: %v", err)
	}
	if !CompareObjectValue(gotObject, objectSeed) {
		t.Fatalf("convertRawStruct(*ObjectValue) mismatch, got %#v", gotObject)
	}

	gotObjectValue, err := structType.convertRawStruct(*objectSeed)
	if err != nil {
		t.Fatalf("convertRawStruct(ObjectValue) failed: %v", err)
	}
	if !CompareObjectValue(gotObjectValue, objectSeed) {
		t.Fatalf("convertRawStruct(ObjectValue) mismatch, got %#v", gotObjectValue)
	}

	sliceSeed := &SliceObjectValue{
		Name:    "status",
		PkgPath: "/vmi",
		Values: []*ObjectValue{
			objectSeed.Copy(),
			{Name: "status", PkgPath: "/vmi", Fields: []*FieldValue{{Name: "id", Value: int64(2)}}},
		},
	}
	sliceStructType := &TypeImpl{
		Name:     "status",
		PkgPath:  "/vmi",
		Value:    models.TypeSliceValue,
		ElemType: &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue},
	}
	gotSlice, err := sliceStructType.convertRawStructToSlice(sliceSeed)
	if err != nil {
		t.Fatalf("convertRawStructToSlice(*SliceObjectValue) failed: %v", err)
	}
	if !CompareSliceObjectValue(gotSlice, sliceSeed) {
		t.Fatalf("convertRawStructToSlice(*SliceObjectValue) mismatch, got %#v", gotSlice)
	}

	gotSliceValue, err := sliceStructType.convertRawStructToSlice(*sliceSeed)
	if err != nil {
		t.Fatalf("convertRawStructToSlice(SliceObjectValue) failed: %v", err)
	}
	if !CompareSliceObjectValue(gotSliceValue, sliceSeed) {
		t.Fatalf("convertRawStructToSlice(SliceObjectValue) mismatch, got %#v", gotSliceValue)
	}

	if _, err := sliceStructType.convertRawStructToSlice(123); err == nil {
		t.Fatal("convertRawStructToSlice(non-slice) should fail")
	}
	if _, err := structType.convertRawStruct(123); err != nil {
		t.Fatalf("convertRawStruct(non-struct) should currently return nil,nil, got %v", err)
	}
}

func TestTypeImplInterfaceExhaustivePointerScalars(t *testing.T) {
	tests := []struct {
		name  string
		typ   *TypeImpl
		input any
		want  any
	}{
		{name: "bool", typ: &TypeImpl{Name: "bool", Value: models.TypeBooleanValue, IsPtr: true}, input: 1, want: func() any { v := true; return &v }()},
		{name: "int8", typ: &TypeImpl{Name: "int8", Value: models.TypeByteValue, IsPtr: true}, input: 8, want: func() any { v := int8(8); return &v }()},
		{name: "int16", typ: &TypeImpl{Name: "int16", Value: models.TypeSmallIntegerValue, IsPtr: true}, input: 16, want: func() any { v := int16(16); return &v }()},
		{name: "int32", typ: &TypeImpl{Name: "int32", Value: models.TypeInteger32Value, IsPtr: true}, input: 32, want: func() any { v := int32(32); return &v }()},
		{name: "int", typ: &TypeImpl{Name: "int", Value: models.TypeIntegerValue, IsPtr: true}, input: 64, want: func() any { v := int(64); return &v }()},
		{name: "int64", typ: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue, IsPtr: true}, input: 128, want: func() any { v := int64(128); return &v }()},
		{name: "uint8", typ: &TypeImpl{Name: "uint8", Value: models.TypePositiveByteValue, IsPtr: true}, input: 8, want: func() any { v := uint8(8); return &v }()},
		{name: "uint16", typ: &TypeImpl{Name: "uint16", Value: models.TypePositiveSmallIntegerValue, IsPtr: true}, input: 16, want: func() any { v := uint16(16); return &v }()},
		{name: "uint32", typ: &TypeImpl{Name: "uint32", Value: models.TypePositiveInteger32Value, IsPtr: true}, input: 32, want: func() any { v := uint32(32); return &v }()},
		{name: "uint", typ: &TypeImpl{Name: "uint", Value: models.TypePositiveIntegerValue, IsPtr: true}, input: 64, want: func() any { v := uint(64); return &v }()},
		{name: "uint64", typ: &TypeImpl{Name: "uint64", Value: models.TypePositiveBigIntegerValue, IsPtr: true}, input: 128, want: func() any { v := uint64(128); return &v }()},
		{name: "float32", typ: &TypeImpl{Name: "float32", Value: models.TypeFloatValue, IsPtr: true}, input: 3.5, want: func() any { v := float32(3.5); return &v }()},
		{name: "float64", typ: &TypeImpl{Name: "float64", Value: models.TypeDoubleValue, IsPtr: true}, input: 7, want: func() any { v := float64(7); return &v }()},
		{name: "datetime", typ: &TypeImpl{Name: "datetime", Value: models.TypeDateTimeValue, IsPtr: true}, input: 123, want: func() any { v := "123"; return &v }()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.typ.Interface(tt.input)
			if err != nil {
				t.Fatalf("Interface(%s) failed: %v", tt.name, err)
			}
			got := reflect.ValueOf(value.Get())
			want := reflect.ValueOf(tt.want)
			if got.Kind() != reflect.Ptr || got.IsNil() || !reflect.DeepEqual(got.Elem().Interface(), want.Elem().Interface()) {
				t.Fatalf("Interface(%s) mismatch, got %#v want %#v", tt.name, value.Get(), tt.want)
			}
		})
	}
}

func TestTypeImplInterfaceErrorBranches(t *testing.T) {
	errorTypes := []struct {
		name string
		typ  *TypeImpl
	}{
		{name: "bool", typ: &TypeImpl{Name: "bool", Value: models.TypeBooleanValue}},
		{name: "int8", typ: &TypeImpl{Name: "int8", Value: models.TypeByteValue}},
		{name: "int16", typ: &TypeImpl{Name: "int16", Value: models.TypeSmallIntegerValue}},
		{name: "int32", typ: &TypeImpl{Name: "int32", Value: models.TypeInteger32Value}},
		{name: "int", typ: &TypeImpl{Name: "int", Value: models.TypeIntegerValue}},
		{name: "int64", typ: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}},
		{name: "uint8", typ: &TypeImpl{Name: "uint8", Value: models.TypePositiveByteValue}},
		{name: "uint16", typ: &TypeImpl{Name: "uint16", Value: models.TypePositiveSmallIntegerValue}},
		{name: "uint32", typ: &TypeImpl{Name: "uint32", Value: models.TypePositiveInteger32Value}},
		{name: "uint", typ: &TypeImpl{Name: "uint", Value: models.TypePositiveIntegerValue}},
		{name: "uint64", typ: &TypeImpl{Name: "uint64", Value: models.TypePositiveBigIntegerValue}},
		{name: "float32", typ: &TypeImpl{Name: "float32", Value: models.TypeFloatValue}},
		{name: "float64", typ: &TypeImpl{Name: "float64", Value: models.TypeDoubleValue}},
		{name: "string", typ: &TypeImpl{Name: "string", Value: models.TypeStringValue}},
		{name: "datetime", typ: &TypeImpl{Name: "datetime", Value: models.TypeDateTimeValue}},
	}
	for _, tt := range errorTypes {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.typ.Interface(map[string]any{"bad": true}); err == nil {
				t.Fatalf("Interface(%s invalid) should fail", tt.name)
			}
		})
	}

	if _, err := (&TypeImpl{Name: "items", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int", Value: models.TypeIntegerValue}}).Interface("bad"); err == nil {
		t.Fatal("Interface(slice basic invalid) should fail")
	}
	if _, err := (&TypeImpl{Name: "items", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int", Value: models.TypeIntegerValue}}).Interface([]any{1, "bad"}); err == nil {
		t.Fatal("Interface(slice basic mixed values) should fail")
	}

	sliceStructType := &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue}}
	if _, err := sliceStructType.Interface("bad"); err == nil {
		t.Fatal("Interface(slice struct invalid) should fail")
	}
	if _, err := sliceStructType.convertRawStructToSlice([]any{struct{}{}}); err == nil {
		t.Fatal("convertRawStructToSlice(invalid slice item) should fail")
	}

	structType := &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue}
	if _, err := structType.convertRawStruct(struct{}{}); err == nil {
		t.Fatal("convertRawStruct(plain struct) should fail")
	}
	if got, err := (&TypeImpl{Name: "unknown", Value: models.TypeDeclare(999)}).Interface("bad"); err != nil {
		t.Fatalf("Interface(unknown type) failed: %v", err)
	} else if got.Get() != (*ObjectValue)(nil) {
		t.Fatalf("Interface(unknown type) mismatch, got %#v", got.Get())
	}
}
