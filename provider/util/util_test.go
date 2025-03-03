package util

import (
	"reflect"
	"testing"

	"github.com/muidea/magicCommon/foundation/log"
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

func TestNilValue(t *testing.T) {
	var val reflect.Value
	log.Infof("IsValid(val), val:%v", val)
	// nil
	if !IsNil(val) {
		t.Errorf("Check val is nil failed")
		return
	}

	var iVal int
	log.Infof("IsValid(reflect.ValueOf(iVal)), val:%v", iVal)
	// not nil
	if IsNil(reflect.ValueOf(iVal)) {
		t.Errorf("Check int is nil failed")
		return
	}
	log.Infof("IsValid(reflect.ValueOf(&iVal)), val:%v", &iVal)
	// not nil
	if IsNil(reflect.ValueOf(&iVal)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	var iValPtr *int
	log.Infof("!IsValid(reflect.ValueOf(iValPtr)), val:%v", iValPtr)
	// nil
	if !IsNil(reflect.ValueOf(iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	iValPtr = &iVal
	log.Infof("IsValid(reflect.ValueOf(iValPtr)), val:%v", iValPtr)
	// not nil
	if IsNil(reflect.ValueOf(iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	log.Infof("IsValid(reflect.ValueOf(&iValPtr)), val:%v", &iValPtr)
	// not nil
	if IsNil(reflect.ValueOf(&iValPtr)) {
		t.Errorf("Check int ptr is nil failed")
		return
	}

	var interfaceVal interface{}
	log.Infof("!IsValid(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	// nil
	if !IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	interfaceVal = iVal
	log.Infof("IsValid(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	// not nil
	if IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	interfaceVal = nil
	log.Infof("IsValid(reflect.ValueOf(interfaceVal)), val:%v", interfaceVal)
	// nil
	if !IsNil(reflect.ValueOf(interfaceVal)) {
		t.Errorf("Check interface is nil failed")
		return
	}

	var arrayIntVal []int
	log.Infof("IsValid(reflect.ValueOf(arrayIntVal)), val:%v", arrayIntVal)
	// nil
	if !IsNil(reflect.ValueOf(arrayIntVal)) {
		t.Errorf("Check arrayIntVal is nil failed")
		return
	}

	var arrayIntInterfaceVal interface{}
	arrayIntInterfaceVal = arrayIntVal
	log.Infof("IsValid(reflect.ValueOf(arrayIntInterfaceVal)), val:%v", arrayIntInterfaceVal)
	// nil
	if !IsNil(reflect.ValueOf(arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal interface is nil failed")
		return
	}

	arrayIntInterfaceVal = &arrayIntVal
	log.Infof("IsValid(reflect.ValueOf(arrayIntInterfaceVal)), val:%v", arrayIntInterfaceVal)
	// not nil
	if IsNil(reflect.ValueOf(arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal ptr interface is nil failed")
		return
	}

	log.Infof("IsValid(reflect.ValueOf(&arrayIntInterfaceVal)), val:%v", &arrayIntInterfaceVal)
	// not nil
	if IsNil(reflect.ValueOf(&arrayIntInterfaceVal)) {
		t.Errorf("Check arrayIntVal ptr interface is nil failed")
		return
	}

	var mapVal map[string]string
	log.Infof("IsValid(reflect.ValueOf(mapVal)), val:%v", mapVal)
	// nil
	if !IsNil(reflect.ValueOf(mapVal)) {
		t.Errorf("Check mapVal is nil failed")
		return
	}

	log.Infof("IsValid(reflect.ValueOf(&mapVal)), val:%v", &mapVal)
	// not nil
	if IsNil(reflect.ValueOf(&mapVal)) {
		t.Errorf("Check mapVal ptr is nil failed")
		return
	}

	intSlice := []int64{}
	log.Infof("IsValid(reflect.ValueOf(intSlice)), val:%v", &intSlice)
	// not nil
	if IsNil(reflect.ValueOf(intSlice)) {
		t.Errorf("Check intSlice ptr is nil failed")
		return
	}
}

func TestStructNilValue(t *testing.T) {
	type Demo struct {
		IntVal       int
		PtrVal       *int
		InterfaceVal interface{}
		ArrayVal     []int
	}

	demo1 := Demo{}
	dv := reflect.ValueOf(demo1)
	intVal := dv.FieldByName("IntVal")
	log.Infof("IsValid(intVal), val:%v", intVal.Interface())
	// not nil
	if IsNil(intVal) {
		t.Errorf("Check intVal is nil failed")
		return
	}

	ptrVal := dv.FieldByName("PtrVal")
	log.Infof("!IsValid(ptrVal), val:%v", ptrVal.Interface())
	// nil
	if !IsNil(ptrVal) {
		t.Errorf("Check ptrVal is nil failed")
		return
	}

	interfaceVal := dv.FieldByName("InterfaceVal")
	log.Infof("!IsValid(interfaceVal), val:%v", interfaceVal.Interface())
	// nil
	if !IsNil(interfaceVal) {
		t.Errorf("Check interfaceVal is nil failed")
		return
	}

	arrayVal := dv.FieldByName("ArrayVal")
	log.Infof("IsValid(arrayVal), val:%v", arrayVal.Interface())
	// nil
	if !IsNil(arrayVal) {
		t.Errorf("Check arrayVal is nil failed")
		return
	}

	ii := 10
	demo2 := Demo{PtrVal: &ii}
	dv2 := reflect.ValueOf(demo2)
	intVal2 := dv2.FieldByName("IntVal")
	log.Infof("IsValid(intVal2), val:%v", intVal2.Interface())
	// not nil
	if IsNil(intVal2) {
		t.Errorf("Check intVal2 is nil failed")
		return
	}

	ptrVal2 := dv2.FieldByName("PtrVal")
	log.Infof("IsValid(ptrVal2), val:%v", ptrVal2.Interface())
	// not nil
	if IsNil(ptrVal2) {
		t.Errorf("Check ptrVal2 is nil failed")
		return
	}

	interfaceVal2 := dv2.FieldByName("InterfaceVal")
	log.Infof("!IsValid(interfaceVal2), val:%v", interfaceVal2.Interface())
	// nil
	if !IsNil(interfaceVal2) {
		t.Errorf("Check interfaceVal2 is nil failed")
		return
	}

	arrayVal2 := dv2.FieldByName("ArrayVal")
	log.Infof("IsValid(arrayVal2), val:%v", arrayVal2.Interface())
	// nil
	if !IsNil(arrayVal2) {
		t.Errorf("Check arrayVal2 is nil failed")
		return
	}
}

func TestTypeEnumEdgeCases(t *testing.T) {
	// Test for pointer types
	var nilVal *int
	nilType := reflect.TypeOf(nilVal)
	enum, err := GetTypeEnum(nilType)
	if err != nil {
		t.Errorf("GetTypeEnum for pointer should not error, got: %v", err)
	}
	if !IsPtr(nilType) {
		t.Errorf("IsPtr for pointer type should be true")
	}

	// Test for nested pointer types
	var nestedPtr **int
	nestedType := reflect.TypeOf(nestedPtr)
	enum, err = GetTypeEnum(nestedType)
	if err != nil {
		t.Errorf("GetTypeEnum for nested pointer should not error, got: %v", err)
	}
	if !IsPtr(nestedType) {
		t.Errorf("IsPtr for nested pointer type should be true")
	}

	// Test for slice of pointers
	var slicePtr []*int
	sliceType := reflect.TypeOf(slicePtr)
	enum, err = GetTypeEnum(sliceType)
	if err != nil {
		t.Errorf("GetTypeEnum for slice of pointers should not error, got: %v", err)
	}
	if enum != model.TypeSliceValue {
		t.Errorf("GetTypeEnum for slice of pointers expected TypeSliceValue, got: %v", enum)
	}

	// Test for interface type
	var iface interface{}
	_ = reflect.TypeOf(&iface).Elem() // Just to check if it's nil
	if !IsNil(reflect.ValueOf(iface)) {
		t.Errorf("IsNil for nil interface should be true")
	}

	// Test for map type
	var mapVal map[string]int
	mapType := reflect.TypeOf(mapVal)
	_, err = GetTypeEnum(mapType)
	if err != nil {
		t.Errorf("GetTypeEnum for map should not error, got: %v", err)
	}
}

func TestIsZeroFunction(t *testing.T) {
	// IsZero tests for basic types
	if !IsZero(reflect.ValueOf(0)) {
		t.Errorf("IsZero for 0 should be true")
	}
	if !IsZero(reflect.ValueOf("")) {
		t.Errorf("IsZero for empty string should be true")
	}
	if !IsZero(reflect.ValueOf(false)) {
		t.Errorf("IsZero for false should be true")
	}
	if IsZero(reflect.ValueOf(1)) {
		t.Errorf("IsZero for 1 should be false")
	}
	if IsZero(reflect.ValueOf("test")) {
		t.Errorf("IsZero for non-empty string should be false")
	}

	// Test for struct types
	emptyStruct := TestVal{}
	if !IsZero(reflect.ValueOf(emptyStruct)) {
		t.Errorf("IsZero for empty struct should be true")
	}

	nonEmptyStruct := TestVal{IVal: 5, SVal: "test"}
	if IsZero(reflect.ValueOf(nonEmptyStruct)) {
		t.Errorf("IsZero for non-empty struct should be false")
	}

	// Test for slice types
	var emptySlice []int
	if !IsZero(reflect.ValueOf(emptySlice)) {
		t.Errorf("IsZero for nil slice should be true")
	}

	nonEmptySlice := []int{1, 2, 3}
	if IsZero(reflect.ValueOf(nonEmptySlice)) {
		t.Errorf("IsZero for non-empty slice should be false")
	}

	// Test for pointer types
	var nilPtr *int
	if !IsZero(reflect.ValueOf(nilPtr)) {
		t.Errorf("IsZero for nil pointer should be true")
	}

	intVal := 5
	ptr := &intVal
	if IsZero(reflect.ValueOf(ptr)) {
		t.Errorf("IsZero for non-nil pointer should be false")
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
