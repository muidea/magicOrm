package consistency

import (
	"reflect"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/provider/local"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/muidea/magicOrm/utils"
)

func TestBasicTypeCodecConsistency(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		typeDecl models.TypeDeclare
	}{
		{"bool", true, models.TypeBooleanValue},
		{"int8", int8(8), models.TypeByteValue},
		{"int16", int16(16), models.TypeSmallIntegerValue},
		{"int32", int32(32), models.TypeInteger32Value},
		{"int64", int64(64), models.TypeBigIntegerValue},
		{"int", int(100), models.TypeIntegerValue},
		{"uint8", uint8(8), models.TypePositiveByteValue},
		{"uint16", uint16(16), models.TypePositiveSmallIntegerValue},
		{"uint32", uint32(32), models.TypePositiveInteger32Value},
		{"uint64", uint64(64), models.TypePositiveBigIntegerValue},
		{"uint", uint(100), models.TypePositiveIntegerValue},
		{"float32", float32(3.14), models.TypeFloatValue},
		{"float64", float64(3.14159), models.TypeDoubleValue},
		{"string", "hello", models.TypeStringValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vType := newTestType(tt.typeDecl, false)

			localEncoded, localErr := local.EncodeValue(tt.value, vType)
			if localErr != nil {
				t.Fatalf("local.EncodeValue failed: %v", localErr)
			}

			remoteEncoded, remoteErr := remote.EncodeValue(tt.value, vType)
			if remoteErr != nil {
				t.Fatalf("remote.EncodeValue failed: %v", remoteErr)
			}

			localDecoded, localErr := local.DecodeValue(localEncoded, vType)
			if localErr != nil {
				t.Fatalf("local.DecodeValue failed: %v", localErr)
			}

			remoteDecoded, remoteErr := remote.DecodeValue(remoteEncoded, vType)
			if remoteErr != nil {
				t.Fatalf("remote.DecodeValue failed: %v", remoteErr)
			}

			if !utils.IsSameValue(localDecoded, remoteDecoded) {
				t.Errorf("decoded values not equal: local=%v (%T), remote=%v (%T)",
					localDecoded, localDecoded, remoteDecoded, remoteDecoded)
			}
		})
	}
}

func TestPointerTypeCodecConsistency(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		typeDecl models.TypeDeclare
	}{
		{"*bool", ptr(true), models.TypeBooleanValue},
		{"*int", ptr(42), models.TypeIntegerValue},
		{"*string", ptr("test"), models.TypeStringValue},
		{"*float64", ptr(3.14), models.TypeDoubleValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vType := newTestType(tt.typeDecl, true)

			localEncoded, localErr := local.EncodeValue(tt.value, vType)
			if localErr != nil {
				t.Fatalf("local.EncodeValue failed: %v", localErr)
			}

			remoteEncoded, remoteErr := remote.EncodeValue(tt.value, vType)
			if remoteErr != nil {
				t.Fatalf("remote.EncodeValue failed: %v", remoteErr)
			}

			localDecoded, localErr := local.DecodeValue(localEncoded, vType)
			if localErr != nil {
				t.Fatalf("local.DecodeValue failed: %v", localErr)
			}

			remoteDecoded, remoteErr := remote.DecodeValue(remoteEncoded, vType)
			if remoteErr != nil {
				t.Fatalf("remote.DecodeValue failed: %v", remoteErr)
			}

			if !utils.IsSameValue(localDecoded, remoteDecoded) {
				t.Errorf("decoded pointer values not equal: local=%v, remote=%v",
					localDecoded, remoteDecoded)
			}
		})
	}
}

func TestSliceTypeCodecConsistency(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		elemDecl  models.TypeDeclare
		isElemPtr bool
	}{
		{"[]bool", []bool{true, false, true}, models.TypeBooleanValue, false},
		{"[]int", []int{1, 2, 3}, models.TypeIntegerValue, false},
		{"[]int8", []int8{-1, 0, 1}, models.TypeByteValue, false},
		{"[]int16", []int16{-10, 0, 10}, models.TypeSmallIntegerValue, false},
		{"[]int32", []int32{-100, 0, 100}, models.TypeInteger32Value, false},
		{"[]int64", []int64{-1000, 0, 1000}, models.TypeBigIntegerValue, false},
		{"[]uint", []uint{1, 2, 3}, models.TypePositiveIntegerValue, false},
		{"[]uint8", []uint8{1, 2, 3}, models.TypePositiveByteValue, false},
		{"[]uint16", []uint16{10, 20, 30}, models.TypePositiveSmallIntegerValue, false},
		{"[]uint32", []uint32{100, 200, 300}, models.TypePositiveInteger32Value, false},
		{"[]uint64", []uint64{1000, 2000, 3000}, models.TypePositiveBigIntegerValue, false},
		{"[]float32", []float32{1.1, 2.2, 3.3}, models.TypeFloatValue, false},
		{"[]float64", []float64{1.11, 2.22, 3.33}, models.TypeDoubleValue, false},
		{"[]string", []string{"a", "b", "c"}, models.TypeStringValue, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vType := newTestSliceType(tt.elemDecl, tt.isElemPtr)

			localEncoded, localErr := local.EncodeValue(tt.value, vType)
			if localErr != nil {
				t.Fatalf("local.EncodeValue failed: %v", localErr)
			}

			remoteEncoded, remoteErr := remote.EncodeValue(tt.value, vType)
			if remoteErr != nil {
				t.Fatalf("remote.EncodeValue failed: %v", remoteErr)
			}

			localDecoded, localErr := local.DecodeValue(localEncoded, vType)
			if localErr != nil {
				t.Fatalf("local.DecodeValue failed: %v", localErr)
			}

			remoteDecoded, remoteErr := remote.DecodeValue(remoteEncoded, vType)
			if remoteErr != nil {
				t.Fatalf("remote.DecodeValue failed: %v", remoteErr)
			}

			if !compareSliceValues(localDecoded, remoteDecoded) {
				t.Errorf("decoded slice values not equal:\nlocal=%v\nremote=%v",
					localDecoded, remoteDecoded)
			}
		})
	}
}

func TestDateTimeCodecConsistency(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	vType := newTestType(models.TypeDateTimeValue, false)

	localEncoded, localErr := local.EncodeValue(testTime, vType)
	if localErr != nil {
		t.Fatalf("local.EncodeValue failed: %v", localErr)
	}

	remoteEncoded, remoteErr := remote.EncodeValue(testTime, vType)
	if remoteErr != nil {
		t.Fatalf("remote.EncodeValue failed: %v", remoteErr)
	}

	t.Logf("DateTime encoded - local: %v (%T), remote: %v (%T)",
		localEncoded, localEncoded, remoteEncoded, remoteEncoded)

	localDecoded, localErr := local.DecodeValue(localEncoded, vType)
	if localErr != nil {
		t.Fatalf("local.DecodeValue failed: %v", localErr)
	}

	remoteDecoded, remoteErr := remote.DecodeValue(remoteEncoded, vType)
	if remoteErr != nil {
		t.Fatalf("remote.DecodeValue failed: %v", remoteErr)
	}

	localTime, ok := localDecoded.(time.Time)
	if !ok {
		t.Fatalf("local decoded not time.Time: %T", localDecoded)
	}

	remoteStr, ok := remoteDecoded.(string)
	if !ok {
		t.Fatalf("remote decoded not string: %T", remoteDecoded)
	}

	t.Logf("DateTime decoded - local: %v, remote: %v", localTime, remoteStr)
}

func TestLocalRemoteRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		typeDecl models.TypeDeclare
		isPtr    bool
	}{
		{"bool", true, models.TypeBooleanValue, false},
		{"int", 42, models.TypeIntegerValue, false},
		{"string", "test", models.TypeStringValue, false},
		{"*int", ptr(100), models.TypeIntegerValue, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vType := newTestType(tt.typeDecl, tt.isPtr)

			localEncoded, _ := local.EncodeValue(tt.value, vType)
			remoteDecoded, err := remote.DecodeValue(localEncoded, vType)
			if err != nil {
				t.Fatalf("remote decode local encoded failed: %v", err)
			}
			roundTripEncoded, _ := local.EncodeValue(remoteDecoded, vType)

			if !utils.IsSameValue(localEncoded, roundTripEncoded) {
				t.Errorf("round trip failed: original=%v, roundtrip=%v",
					localEncoded, roundTripEncoded)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func newTestType(typeDecl models.TypeDeclare, isPtr bool) models.Type {
	rType := getTypeForDecl(typeDecl, isPtr)
	typ, _ := local.NewType(rType)
	return typ
}

func newTestSliceType(elemDecl models.TypeDeclare, isElemPtr bool) models.Type {
	elemType := getTypeForDecl(elemDecl, isElemPtr)
	sliceType := reflect.SliceOf(elemType)
	typ, _ := local.NewType(sliceType)
	return typ
}

func getTypeForDecl(typeDecl models.TypeDeclare, isPtr bool) reflect.Type {
	var rType reflect.Type
	switch typeDecl {
	case models.TypeBooleanValue:
		rType = reflect.TypeOf(false)
	case models.TypeByteValue:
		rType = reflect.TypeOf(int8(0))
	case models.TypeSmallIntegerValue:
		rType = reflect.TypeOf(int16(0))
	case models.TypeInteger32Value:
		rType = reflect.TypeOf(int32(0))
	case models.TypeBigIntegerValue:
		rType = reflect.TypeOf(int64(0))
	case models.TypeIntegerValue:
		rType = reflect.TypeOf(int(0))
	case models.TypePositiveByteValue:
		rType = reflect.TypeOf(uint8(0))
	case models.TypePositiveSmallIntegerValue:
		rType = reflect.TypeOf(uint16(0))
	case models.TypePositiveInteger32Value:
		rType = reflect.TypeOf(uint32(0))
	case models.TypePositiveBigIntegerValue:
		rType = reflect.TypeOf(uint64(0))
	case models.TypePositiveIntegerValue:
		rType = reflect.TypeOf(uint(0))
	case models.TypeFloatValue:
		rType = reflect.TypeOf(float32(0))
	case models.TypeDoubleValue:
		rType = reflect.TypeOf(float64(0))
	case models.TypeStringValue:
		rType = reflect.TypeOf("")
	case models.TypeDateTimeValue:
		rType = reflect.TypeOf(time.Time{})
	default:
		rType = reflect.TypeOf(int(0))
	}

	if isPtr {
		rType = reflect.PointerTo(rType)
	}
	return rType
}

func compareSliceValues(a, b any) bool {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	if aVal.Kind() == reflect.Ptr {
		aVal = aVal.Elem()
	}
	if bVal.Kind() == reflect.Ptr {
		bVal = bVal.Elem()
	}

	if aVal.Kind() != reflect.Slice || bVal.Kind() != reflect.Slice {
		return false
	}

	if aVal.Len() != bVal.Len() {
		return false
	}

	for i := 0; i < aVal.Len(); i++ {
		if !utils.IsSameValue(aVal.Index(i).Interface(), bVal.Index(i).Interface()) {
			return false
		}
	}

	return true
}

var _ = cd.NewError
