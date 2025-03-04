package util

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/model"
)

type TestVal struct {
	BVal     bool    `json:"bVal"`
	IVal     int     `json:"iVal"`
	I16Val   int16   `json:"i16Val"`
	FVal     float32 `json:"fVal"`
	F64Val   float64 `json:"f64Val"`
	SVal     string  `json:"sVal"`
	ArrayVal []int64 `json:"arrayVal"`
}

func TestTypeEnumEdgeCases(t *testing.T) {
	// Test for pointer types
	var nilVal *int
	nilType := reflect.TypeOf(nilVal)
	_, err := GetTypeEnum(nilType)
	if err != nil {
		t.Errorf("GetTypeEnum for pointer should not error, got: %v", err)
	}
	if !IsPtr(nilType) {
		t.Errorf("IsPtr for pointer type should be true")
	}

	// Test for nested pointer types
	var nestedPtr **int
	nestedType := reflect.TypeOf(nestedPtr)
	_, err = GetTypeEnum(nestedType)
	if err != nil {
		t.Errorf("GetTypeEnum for nested pointer should not error, got: %v", err)
	}
	if !IsPtr(nestedType) {
		t.Errorf("IsPtr for nested pointer type should be true")
	}

	// Test for slice of pointers
	var slicePtr []*int
	sliceType := reflect.TypeOf(slicePtr)
	enumVal, enumErr := GetTypeEnum(sliceType)
	if enumErr != nil {
		t.Errorf("GetTypeEnum for slice of pointers should not error, got: %v", enumErr)
	}
	if enumVal != model.TypeSliceValue {
		t.Errorf("GetTypeEnum for slice of pointers expected TypeSliceValue, got: %v", enumVal)
	}

	// Test for map type
	var mapVal map[string]int
	mapType := reflect.TypeOf(mapVal)
	_, err = GetTypeEnum(mapType)
	if err != nil {
		t.Errorf("GetTypeEnum for map should not error, got: %v", err)
	}
}

func TestTypeConversionsRaw(t *testing.T) {
	// Test int to string conversion
	intVal := int64(42)
	str, err := GetRawString(reflect.ValueOf(intVal))
	if err != nil || str != "42" {
		t.Errorf("Failed to convert int64 to string: %v", err)
	}

	// Test string to int conversion
	strVal := "42"
	intResult, err := GetRawInt64(reflect.ValueOf(strVal))
	if err != nil || intResult != 42 {
		t.Errorf("Failed to convert string to int64: %v", err)
	}

	// Test float conversion - the string representation may vary slightly
	floatVal := 42.5
	floatStr, err := GetRawString(reflect.ValueOf(floatVal))
	if err != nil {
		t.Errorf("Failed to convert float64 to string: %v", err)
	}
	// Float to string conversion may have different formats, so we just check that it contains "42.5"
	if floatStr != "42.5" && floatStr != "42.500000" {
		t.Errorf("Unexpected string conversion for float: %s", floatStr)
	}

	// Test for string to bool conversion - for strings other than "1", it should return false, not error
	nonBoolStr := "notabool"
	boolVal, err := GetRawBool(reflect.ValueOf(nonBoolStr))
	if err != nil {
		t.Errorf("GetRawBool should not return error for string values, got: %v", err)
	}
	if boolVal != false {
		t.Errorf("GetRawBool for non-'1' string should return false")
	}

	// Test for actual error case (unsupported type)
	complexVal := complex(1, 2)
	_, err = GetRawBool(reflect.ValueOf(complexVal))
	if err == nil {
		t.Errorf("Expected error when converting complex to bool, but got none")
	}
}
