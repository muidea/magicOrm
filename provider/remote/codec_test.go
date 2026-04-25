package remote

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/models"
)

func ptrOf[T any](v T) *T {
	return &v
}

func interfacePtr(v any) any {
	val := reflect.New(reflect.TypeOf(v))
	val.Elem().Set(reflect.ValueOf(v))
	return val.Interface()
}

func assertCodecEqual(t *testing.T, got, want any) {
	t.Helper()

	gotVal := reflect.ValueOf(got)
	wantVal := reflect.ValueOf(want)
	if wantVal.IsValid() && wantVal.Kind() == reflect.Ptr {
		if !gotVal.IsValid() || gotVal.Kind() != reflect.Ptr || gotVal.IsNil() || wantVal.IsNil() {
			t.Fatalf("value mismatch, got %#v want %#v", got, want)
		}
		if !reflect.DeepEqual(gotVal.Elem().Interface(), wantVal.Elem().Interface()) {
			t.Fatalf("value mismatch, got %#v want %#v", got, want)
		}
		return
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("value mismatch, got %#v want %#v", got, want)
	}
}

func TestRemoteCodecEncodeDecodeVMIBasicFields(t *testing.T) {
	product := loadVMIObject(t, "test/vmi/entity/product/product.json")

	nameType := requireField(t, product, "name").GetType()
	encodedName, err := EncodeValue("apple", nameType)
	if err != nil {
		t.Fatalf("EncodeValue(name) failed: %v", err)
	}
	if encodedName != "apple" {
		t.Fatalf("EncodeValue(name) mismatch, got %v", encodedName)
	}

	decodedName, err := DecodeValue(encodedName, nameType)
	if err != nil {
		t.Fatalf("DecodeValue(name) failed: %v", err)
	}
	if decodedName != "apple" {
		t.Fatalf("DecodeValue(name) mismatch, got %v", decodedName)
	}

	imageType := requireField(t, product, "image").GetType()
	encodedImage, err := EncodeValue([]string{"main.png", "thumb.png"}, imageType)
	if err != nil {
		t.Fatalf("EncodeValue(image) failed: %v", err)
	}
	if !reflect.DeepEqual(encodedImage, []string{"main.png", "thumb.png"}) {
		t.Fatalf("EncodeValue(image) mismatch, got %#v", encodedImage)
	}

	decodedImage, err := DecodeValue([]any{"main.png", "thumb.png"}, imageType)
	if err != nil {
		t.Fatalf("DecodeValue(image) failed: %v", err)
	}
	if !reflect.DeepEqual(decodedImage, []string{"main.png", "thumb.png"}) {
		t.Fatalf("DecodeValue(image) mismatch, got %#v", decodedImage)
	}
}

func TestRemoteCodecEncodeDecodePointerAndErrors(t *testing.T) {
	ptrStringType := &TypeImpl{
		Name:    "Name",
		PkgPath: "test",
		Value:   models.TypeStringValue,
		IsPtr:   true,
	}

	input := "apple"
	encodedPtr, err := EncodeValue(&input, ptrStringType)
	if err != nil {
		t.Fatalf("EncodeValue(pointer string) failed: %v", err)
	}
	encodedString, ok := encodedPtr.(*string)
	if !ok || encodedString == nil || *encodedString != "apple" {
		t.Fatalf("EncodeValue(pointer string) mismatch, got %#v", encodedPtr)
	}

	decodedPtr, err := DecodeValue("apple", ptrStringType)
	if err != nil {
		t.Fatalf("DecodeValue(pointer string) failed: %v", err)
	}
	decodedString, ok := decodedPtr.(*string)
	if !ok || decodedString == nil || *decodedString != "apple" {
		t.Fatalf("DecodeValue(pointer string) mismatch, got %#v", decodedPtr)
	}

	ptrSliceType := &TypeImpl{
		Name:    "Tags",
		PkgPath: "test",
		Value:   models.TypeSliceValue,
		IsPtr:   true,
		ElemType: &TypeImpl{
			Name:    "string",
			PkgPath: "",
			Value:   models.TypeStringValue,
		},
	}

	decodedSlice, err := DecodeValue([]any{"fresh", "fruit"}, ptrSliceType)
	if err != nil {
		t.Fatalf("DecodeValue(pointer string slice) failed: %v", err)
	}
	stringSlice, ok := decodedSlice.(*[]string)
	if !ok || stringSlice == nil || !reflect.DeepEqual(*stringSlice, []string{"fresh", "fruit"}) {
		t.Fatalf("DecodeValue(pointer string slice) mismatch, got %#v", decodedSlice)
	}

	statusType := requireField(t, loadVMIObject(t, "test/vmi/entity/product/product.json"), "status").GetType()
	if _, err := EncodeValue("illegal", statusType); err == nil {
		t.Fatalf("EncodeValue(struct type) should fail")
	}
	if _, err := DecodeValue("illegal", statusType); err == nil {
		t.Fatalf("DecodeValue(struct type) should fail")
	}
}

func TestRemoteCodecEncodeDecodeAdditionalBasicTypes(t *testing.T) {
	tests := []struct {
		name          string
		typ           *TypeImpl
		encodeInput   any
		decodeInput   any
		expectedValue any
	}{
		{name: "boolean", typ: &TypeImpl{Name: "boolean", Value: models.TypeBooleanValue}, encodeInput: true, decodeInput: true, expectedValue: true},
		{name: "int8", typ: &TypeImpl{Name: "int8", Value: models.TypeByteValue}, encodeInput: int8(8), decodeInput: int8(8), expectedValue: int8(8)},
		{name: "int16", typ: &TypeImpl{Name: "int16", Value: models.TypeSmallIntegerValue}, encodeInput: int16(16), decodeInput: int16(16), expectedValue: int16(16)},
		{name: "int32", typ: &TypeImpl{Name: "int32", Value: models.TypeInteger32Value}, encodeInput: int32(32), decodeInput: int32(32), expectedValue: int32(32)},
		{name: "int64", typ: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}, encodeInput: int64(64), decodeInput: int64(64), expectedValue: int64(64)},
		{name: "uint8", typ: &TypeImpl{Name: "uint8", Value: models.TypePositiveByteValue}, encodeInput: uint8(8), decodeInput: uint8(8), expectedValue: uint8(8)},
		{name: "uint16", typ: &TypeImpl{Name: "uint16", Value: models.TypePositiveSmallIntegerValue}, encodeInput: uint16(16), decodeInput: uint16(16), expectedValue: uint16(16)},
		{name: "uint32", typ: &TypeImpl{Name: "uint32", Value: models.TypePositiveInteger32Value}, encodeInput: uint32(32), decodeInput: uint32(32), expectedValue: uint32(32)},
		{name: "uint64", typ: &TypeImpl{Name: "uint64", Value: models.TypePositiveBigIntegerValue}, encodeInput: uint64(64), decodeInput: uint64(64), expectedValue: uint64(64)},
		{name: "float32", typ: &TypeImpl{Name: "float32", Value: models.TypeFloatValue}, encodeInput: float32(3.5), decodeInput: float32(3.5), expectedValue: float32(3.5)},
		{name: "float64", typ: &TypeImpl{Name: "float64", Value: models.TypeDoubleValue}, encodeInput: float64(7.5), decodeInput: float64(7.5), expectedValue: float64(7.5)},
		{name: "datetime", typ: &TypeImpl{Name: "datetime", Value: models.TypeDateTimeValue}, encodeInput: "2025-01-01T00:00:00Z", decodeInput: "2025-01-01T00:00:00Z", expectedValue: "2025-01-01T00:00:00Z"},
		{
			name:          "[]boolean",
			typ:           &TypeImpl{Name: "boolean", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "boolean", Value: models.TypeBooleanValue}},
			encodeInput:   []bool{true, false},
			decodeInput:   []any{true, false},
			expectedValue: []bool{true, false},
		},
		{
			name:          "[]int",
			typ:           &TypeImpl{Name: "int", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int", Value: models.TypeIntegerValue}},
			encodeInput:   []int{1, 2},
			decodeInput:   []any{1, 2},
			expectedValue: []int{1, 2},
		},
		{
			name:          "[]float64",
			typ:           &TypeImpl{Name: "float64", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "float64", Value: models.TypeDoubleValue}},
			encodeInput:   []float64{1.5, 2.5},
			decodeInput:   []any{1.5, 2.5},
			expectedValue: []float64{1.5, 2.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeValue(tt.encodeInput, tt.typ)
			if err != nil {
				t.Fatalf("EncodeValue(%s) failed: %v", tt.name, err)
			}
			if !reflect.DeepEqual(encoded, tt.expectedValue) {
				t.Fatalf("EncodeValue(%s) mismatch, got %#v want %#v", tt.name, encoded, tt.expectedValue)
			}

			decoded, err := DecodeValue(tt.decodeInput, tt.typ)
			if err != nil {
				t.Fatalf("DecodeValue(%s) failed: %v", tt.name, err)
			}
			if !reflect.DeepEqual(decoded, tt.expectedValue) {
				t.Fatalf("DecodeValue(%s) mismatch, got %#v want %#v", tt.name, decoded, tt.expectedValue)
			}
		})
	}
}

func TestRemoteCodecExhaustivePointerVariants(t *testing.T) {
	scalarCases := []struct {
		name   string
		value  models.TypeDeclare
		raw    any
		expect any
	}{
		{name: "boolean", value: models.TypeBooleanValue, raw: true, expect: true},
		{name: "int8", value: models.TypeByteValue, raw: int8(8), expect: int8(8)},
		{name: "int16", value: models.TypeSmallIntegerValue, raw: int16(16), expect: int16(16)},
		{name: "int32", value: models.TypeInteger32Value, raw: int32(32), expect: int32(32)},
		{name: "int", value: models.TypeIntegerValue, raw: int(64), expect: int(64)},
		{name: "int64", value: models.TypeBigIntegerValue, raw: int64(128), expect: int64(128)},
		{name: "uint8", value: models.TypePositiveByteValue, raw: uint8(8), expect: uint8(8)},
		{name: "uint16", value: models.TypePositiveSmallIntegerValue, raw: uint16(16), expect: uint16(16)},
		{name: "uint32", value: models.TypePositiveInteger32Value, raw: uint32(32), expect: uint32(32)},
		{name: "uint", value: models.TypePositiveIntegerValue, raw: uint(64), expect: uint(64)},
		{name: "uint64", value: models.TypePositiveBigIntegerValue, raw: uint64(128), expect: uint64(128)},
		{name: "float32", value: models.TypeFloatValue, raw: float32(3.5), expect: float32(3.5)},
		{name: "float64", value: models.TypeDoubleValue, raw: float64(7.5), expect: float64(7.5)},
		{name: "string", value: models.TypeStringValue, raw: "apple", expect: "apple"},
		{name: "datetime", value: models.TypeDateTimeValue, raw: "2025-01-01T00:00:00Z", expect: "2025-01-01T00:00:00Z"},
	}
	for _, tt := range scalarCases {
		t.Run("scalar/"+tt.name, func(t *testing.T) {
			baseType := &TypeImpl{Name: tt.name, Value: tt.value}
			encoded, err := EncodeValue(tt.raw, baseType)
			if err != nil {
				t.Fatalf("EncodeValue(%s) failed: %v", tt.name, err)
			}
			assertCodecEqual(t, encoded, tt.expect)

			decoded, err := DecodeValue(tt.raw, baseType)
			if err != nil {
				t.Fatalf("DecodeValue(%s) failed: %v", tt.name, err)
			}
			assertCodecEqual(t, decoded, tt.expect)

			ptrType := &TypeImpl{Name: tt.name, Value: tt.value, IsPtr: true}
			expectedPtr := interfacePtr(tt.expect)
			encodedPtr, err := EncodeValue(expectedPtr, ptrType)
			if err != nil {
				t.Fatalf("EncodeValue(ptr %s) failed: %v", tt.name, err)
			}
			assertCodecEqual(t, encodedPtr, expectedPtr)

			decodedPtr, err := DecodeValue(tt.raw, ptrType)
			if err != nil {
				t.Fatalf("DecodeValue(ptr %s) failed: %v", tt.name, err)
			}
			assertCodecEqual(t, decodedPtr, expectedPtr)
		})
	}

	sliceCases := []struct {
		name   string
		elem   models.TypeDeclare
		encode any
		decode any
		expect any
	}{
		{name: "boolean", elem: models.TypeBooleanValue, encode: []bool{true, false}, decode: []any{true, false}, expect: []bool{true, false}},
		{name: "int8", elem: models.TypeByteValue, encode: []int8{1, 2}, decode: []any{int8(1), int8(2)}, expect: []int8{1, 2}},
		{name: "int16", elem: models.TypeSmallIntegerValue, encode: []int16{1, 2}, decode: []any{int16(1), int16(2)}, expect: []int16{1, 2}},
		{name: "int32", elem: models.TypeInteger32Value, encode: []int32{1, 2}, decode: []any{int32(1), int32(2)}, expect: []int32{1, 2}},
		{name: "int", elem: models.TypeIntegerValue, encode: []int{1, 2}, decode: []any{1, 2}, expect: []int{1, 2}},
		{name: "int64", elem: models.TypeBigIntegerValue, encode: []int64{1, 2}, decode: []any{int64(1), int64(2)}, expect: []int64{1, 2}},
		{name: "uint8", elem: models.TypePositiveByteValue, encode: []uint8{1, 2}, decode: []any{uint8(1), uint8(2)}, expect: []uint8{1, 2}},
		{name: "uint16", elem: models.TypePositiveSmallIntegerValue, encode: []uint16{1, 2}, decode: []any{uint16(1), uint16(2)}, expect: []uint16{1, 2}},
		{name: "uint32", elem: models.TypePositiveInteger32Value, encode: []uint32{1, 2}, decode: []any{uint32(1), uint32(2)}, expect: []uint32{1, 2}},
		{name: "uint", elem: models.TypePositiveIntegerValue, encode: []uint{1, 2}, decode: []any{uint(1), uint(2)}, expect: []uint{1, 2}},
		{name: "uint64", elem: models.TypePositiveBigIntegerValue, encode: []uint64{1, 2}, decode: []any{uint64(1), uint64(2)}, expect: []uint64{1, 2}},
		{name: "float32", elem: models.TypeFloatValue, encode: []float32{1.5, 2.5}, decode: []any{float32(1.5), float32(2.5)}, expect: []float32{1.5, 2.5}},
		{name: "float64", elem: models.TypeDoubleValue, encode: []float64{1.5, 2.5}, decode: []any{1.5, 2.5}, expect: []float64{1.5, 2.5}},
		{name: "string", elem: models.TypeStringValue, encode: []string{"a", "b"}, decode: []any{"a", "b"}, expect: []string{"a", "b"}},
		{name: "datetime", elem: models.TypeDateTimeValue, encode: []string{"2025-01-01", "2025-01-02"}, decode: []any{"2025-01-01", "2025-01-02"}, expect: []string{"2025-01-01", "2025-01-02"}},
	}
	for _, tt := range sliceCases {
		t.Run("slice/"+tt.name, func(t *testing.T) {
			baseType := &TypeImpl{
				Name:     tt.name,
				Value:    models.TypeSliceValue,
				ElemType: &TypeImpl{Name: tt.name, Value: tt.elem},
			}
			encoded, err := EncodeValue(tt.encode, baseType)
			if err != nil {
				t.Fatalf("EncodeValue([]%s) failed: %v", tt.name, err)
			}
			assertCodecEqual(t, encoded, tt.expect)

			decoded, err := DecodeValue(tt.decode, baseType)
			if err != nil {
				t.Fatalf("DecodeValue([]%s) failed: %v", tt.name, err)
			}
			assertCodecEqual(t, decoded, tt.expect)

			ptrType := &TypeImpl{
				Name:     tt.name,
				Value:    models.TypeSliceValue,
				IsPtr:    true,
				ElemType: &TypeImpl{Name: tt.name, Value: tt.elem},
			}
			expectedPtr := interfacePtr(tt.expect)
			encodedPtr, err := EncodeValue(expectedPtr, ptrType)
			if err != nil {
				t.Fatalf("EncodeValue(ptr []%s) failed: %v", tt.name, err)
			}
			assertCodecEqual(t, encodedPtr, expectedPtr)

			decodedPtr, err := DecodeValue(tt.decode, ptrType)
			if err != nil {
				t.Fatalf("DecodeValue(ptr []%s) failed: %v", tt.name, err)
			}
			assertCodecEqual(t, decodedPtr, expectedPtr)
		})
	}
}

func TestRemoteCodecInternalErrorBranches(t *testing.T) {
	intType := &TypeImpl{Name: "int", Value: models.TypeIntegerValue}
	if _, err := encodeValue(reflect.Value{}, intType); err == nil {
		t.Fatal("encodeValue(invalid) should fail")
	}
	if _, err := decodeValue(reflect.Value{}, intType); err == nil {
		t.Fatal("decodeValue(invalid) should fail")
	}

	structType := &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue}
	if _, err := encodeSliceValue(reflect.ValueOf([]int{1}), structType); err == nil {
		t.Fatal("encodeSliceValue(non-basic type) should fail")
	}
	if _, err := decodeSliceValue(reflect.ValueOf([]int{1}), structType); err == nil {
		t.Fatal("decodeSliceValue(non-basic type) should fail")
	}

	sliceStructType := &TypeImpl{
		Name:     "status",
		Value:    models.TypeSliceValue,
		ElemType: &TypeImpl{Name: "status", PkgPath: "/vmi", Value: models.TypeStructValue},
	}
	if _, err := encodeSliceValue(reflect.ValueOf([]int{1}), sliceStructType); err == nil {
		t.Fatal("encodeSliceValue(slice struct) should fail")
	}
	if _, err := decodeSliceValue(reflect.ValueOf([]int{1}), sliceStructType); err == nil {
		t.Fatal("decodeSliceValue(slice struct) should fail")
	}

	if _, err := encodeSliceTemplate[int8](reflect.ValueOf([]int{1}), &TypeImpl{Name: "string", Value: models.TypeStringValue}, int8(0)); err == nil {
		t.Fatal("encodeSliceTemplate(type mismatch) should fail")
	}
	if _, err := decodeSliceTemplate[int8](reflect.ValueOf([]any{"1"}), &TypeImpl{Name: "items", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "string", Value: models.TypeStringValue}}, int8(0)); err == nil {
		t.Fatal("decodeSliceTemplate(type mismatch) should fail")
	}
}

func TestRemoteCodecExhaustiveErrorVariants(t *testing.T) {
	scalarTypes := []*TypeImpl{
		{Name: "boolean", Value: models.TypeBooleanValue},
		{Name: "int8", Value: models.TypeByteValue},
		{Name: "int16", Value: models.TypeSmallIntegerValue},
		{Name: "int32", Value: models.TypeInteger32Value},
		{Name: "int", Value: models.TypeIntegerValue},
		{Name: "int64", Value: models.TypeBigIntegerValue},
		{Name: "uint8", Value: models.TypePositiveByteValue},
		{Name: "uint16", Value: models.TypePositiveSmallIntegerValue},
		{Name: "uint32", Value: models.TypePositiveInteger32Value},
		{Name: "uint", Value: models.TypePositiveIntegerValue},
		{Name: "uint64", Value: models.TypePositiveBigIntegerValue},
		{Name: "float32", Value: models.TypeFloatValue},
		{Name: "float64", Value: models.TypeDoubleValue},
		{Name: "string", Value: models.TypeStringValue},
		{Name: "datetime", Value: models.TypeDateTimeValue},
	}
	for _, typ := range scalarTypes {
		t.Run("scalar/"+typ.Name, func(t *testing.T) {
			if _, err := EncodeValue(map[string]any{"bad": true}, typ); err == nil {
				t.Fatalf("EncodeValue(%s invalid) should fail", typ.Name)
			}
			if _, err := DecodeValue(map[string]any{"bad": true}, typ); err == nil {
				t.Fatalf("DecodeValue(%s invalid) should fail", typ.Name)
			}
		})
	}

	sliceTypes := []struct {
		name  string
		typ   *TypeImpl
		input any
	}{
		{name: "boolean", typ: &TypeImpl{Name: "boolean", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "boolean", Value: models.TypeBooleanValue}}, input: []any{true, map[string]any{"bad": true}}},
		{name: "int8", typ: &TypeImpl{Name: "int8", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int8", Value: models.TypeByteValue}}, input: []any{int8(1), "bad"}},
		{name: "int16", typ: &TypeImpl{Name: "int16", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int16", Value: models.TypeSmallIntegerValue}}, input: []any{int16(1), "bad"}},
		{name: "int32", typ: &TypeImpl{Name: "int32", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int32", Value: models.TypeInteger32Value}}, input: []any{int32(1), "bad"}},
		{name: "int", typ: &TypeImpl{Name: "int", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int", Value: models.TypeIntegerValue}}, input: []any{1, "bad"}},
		{name: "int64", typ: &TypeImpl{Name: "int64", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "int64", Value: models.TypeBigIntegerValue}}, input: []any{int64(1), "bad"}},
		{name: "uint8", typ: &TypeImpl{Name: "uint8", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint8", Value: models.TypePositiveByteValue}}, input: []any{uint8(1), "bad"}},
		{name: "uint16", typ: &TypeImpl{Name: "uint16", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint16", Value: models.TypePositiveSmallIntegerValue}}, input: []any{uint16(1), "bad"}},
		{name: "uint32", typ: &TypeImpl{Name: "uint32", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint32", Value: models.TypePositiveInteger32Value}}, input: []any{uint32(1), "bad"}},
		{name: "uint", typ: &TypeImpl{Name: "uint", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint", Value: models.TypePositiveIntegerValue}}, input: []any{uint(1), "bad"}},
		{name: "uint64", typ: &TypeImpl{Name: "uint64", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "uint64", Value: models.TypePositiveBigIntegerValue}}, input: []any{uint64(1), "bad"}},
		{name: "float32", typ: &TypeImpl{Name: "float32", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "float32", Value: models.TypeFloatValue}}, input: []any{float32(1.5), "bad"}},
		{name: "float64", typ: &TypeImpl{Name: "float64", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "float64", Value: models.TypeDoubleValue}}, input: []any{1.5, "bad"}},
		{name: "string", typ: &TypeImpl{Name: "string", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "string", Value: models.TypeStringValue}}, input: []any{"ok", map[string]any{"bad": true}}},
		{name: "datetime", typ: &TypeImpl{Name: "datetime", Value: models.TypeSliceValue, ElemType: &TypeImpl{Name: "datetime", Value: models.TypeDateTimeValue}}, input: []any{"2025-01-01", map[string]any{"bad": true}}},
	}
	for _, tt := range sliceTypes {
		t.Run("slice/"+tt.name, func(t *testing.T) {
			if _, err := EncodeValue(tt.input, tt.typ); err == nil {
				t.Fatalf("EncodeValue([]%s invalid) should fail", tt.name)
			}
			if _, err := DecodeValue(tt.input, tt.typ); err == nil {
				t.Fatalf("DecodeValue([]%s invalid) should fail", tt.name)
			}
		})
	}
}
