package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
)

// TypeValueTestStruct represents a test struct for type and value operations
type TypeValueTestStruct struct {
	ID         int       `orm:"id key auto"`
	Name       string    `orm:"name"`
	Value      float64   `orm:"value"`
	IsActive   bool      `orm:"isActive"`
	CreateTime time.Time `orm:"createTime"`
}

func TestGetValueType(t *testing.T) {
	// Test basic type values
	intVal := reflect.ValueOf(42)
	intType, err := getValueType(intVal)
	if err != nil {
		t.Errorf("getValueType failed for int: %s", err.Error())
	}
	if intType.GetValue() != model.TypeIntegerValue {
		t.Errorf("Wrong type for int, expected TypeIntegerValue, got: %v", intType.GetValue())
	}

	strVal := reflect.ValueOf("test string")
	strType, err := getValueType(strVal)
	if err != nil {
		t.Errorf("getValueType failed for string: %s", err.Error())
	}
	if strType.GetValue() != model.TypeStringValue {
		t.Errorf("Wrong type for string, expected TypeStringValue, got: %v", strType.GetValue())
	}

	boolVal := reflect.ValueOf(true)
	boolType, err := getValueType(boolVal)
	if err != nil {
		t.Errorf("getValueType failed for bool: %s", err.Error())
	}
	if boolType.GetValue() != model.TypeBooleanValue {
		t.Errorf("Wrong type for bool, expected TypeBooleanValue, got: %v", boolType.GetValue())
	}

	floatVal := reflect.ValueOf(float32(3.14))
	floatType, err := getValueType(floatVal)
	if err != nil {
		t.Errorf("getValueType failed for float64: %s", err.Error())
	}
	if floatType.GetValue() != model.TypeFloatValue {
		t.Errorf("Wrong type for float64, expected TypeFloatValue, got: %v", floatType.GetValue())
	}

	// Test time.Time value
	timeVal := reflect.ValueOf(time.Now())
	timeType, err := getValueType(timeVal)
	if err != nil {
		t.Errorf("getValueType failed for time.Time: %s", err.Error())
	}
	if timeType.GetValue() != model.TypeDateTimeValue {
		t.Errorf("Wrong type for time.Time, expected TypeDateTimeValue, got: %v", timeType.GetValue())
	}

	// Test struct value
	structVal := reflect.ValueOf(TypeValueTestStruct{ID: 1, Name: "test"})
	structType, err := getValueType(structVal)
	if err != nil {
		t.Errorf("getValueType failed for struct: %s", err.Error())
	}
	if structType.GetValue() != model.TypeStructValue {
		t.Errorf("Wrong type for struct, expected TypeStructValue, got: %v", structType.GetValue())
	}

	// Test pointer value
	intPtr := new(int)
	*intPtr = 42
	ptrVal := reflect.ValueOf(intPtr)
	ptrType, err := getValueType(ptrVal)
	if err != nil {
		t.Errorf("getValueType failed for pointer: %s", err.Error())
	}
	if !ptrType.IsPtrType() {
		t.Errorf("IsPtrType should return true for pointer values")
	}

	// Test slice value
	sliceVal := reflect.ValueOf([]int{1, 2, 3})
	sliceType, err := getValueType(sliceVal)
	if err != nil {
		t.Errorf("getValueType failed for slice: %s", err.Error())
	}
	if !sliceType.IsSlice() {
		t.Errorf("IsSlice should return true for slice values")
	}

	// Test nil value
	var nilVal *int = nil
	nilValue := reflect.ValueOf(nilVal)
	_, err = getValueType(nilValue)
	if err == nil {
		t.Errorf("getValueType should fail for nil values")
	}

	// Test unsupported type
	chanVal := reflect.ValueOf(make(chan int))
	_, err = getValueType(chanVal)
	if err == nil {
		t.Errorf("getValueType should fail for unsupported types like channels")
	}
}

func TestGetTypeExtended(t *testing.T) {
	// Test GetType with various reflect.Type inputs
	intType := reflect.TypeOf(int(0))
	typeVal, err := GetType(intType)
	if err != nil {
		t.Errorf("GetType failed for int: %s", err.Error())
	}
	if typeVal.GetValue() != model.TypeIntegerValue {
		t.Errorf("Wrong type for int, expected TypeIntegerValue, got: %v", typeVal.GetValue())
	}

	structType := reflect.TypeOf(TypeValueTestStruct{})
	typeVal, err = GetType(structType)
	if err != nil {
		t.Errorf("GetType failed for struct: %s", err.Error())
	}
	if typeVal.GetValue() != model.TypeStructValue {
		t.Errorf("Wrong type for struct, expected TypeStructValue, got: %v", typeVal.GetValue())
	}

	// Test with pointer type
	ptrType := reflect.TypeOf(&TypeValueTestStruct{})
	typeVal, err = GetType(ptrType)
	if err != nil {
		t.Errorf("GetType failed for pointer: %s", err.Error())
	}
	if !typeVal.IsPtrType() {
		t.Errorf("IsPtrType should return true for pointer types")
	}

	// Test with slice type
	sliceType := reflect.TypeOf([]int{})
	typeVal, err = GetType(sliceType)
	if err != nil {
		t.Errorf("GetType failed for slice: %s", err.Error())
	}
	if !typeVal.IsSlice() {
		t.Errorf("IsSlice should return true for slice types")
	}

	// Test with unsupported type
	chanType := reflect.TypeOf(make(chan int))
	_, err = GetType(chanType)
	if err == nil {
		t.Errorf("GetType should fail for unsupported types like channels")
	}
}

func TestGetEntityTypeExtended(t *testing.T) {
	// Test GetEntityType with various entity inputs
	entity := TypeValueTestStruct{ID: 1, Name: "test"}
	typeVal, err := GetEntityType(entity)
	if err != nil {
		t.Errorf("GetEntityType failed for struct: %s", err.Error())
	}
	if typeVal.GetValue() != model.TypeStructValue {
		t.Errorf("Wrong type for struct, expected TypeStructValue, got: %v", typeVal.GetValue())
	}

	// Test with pointer entity
	ptrEntity := &TypeValueTestStruct{ID: 1, Name: "test"}
	typeVal, err = GetEntityType(ptrEntity)
	if err != nil {
		t.Errorf("GetEntityType failed for pointer struct: %s", err.Error())
	}
	if typeVal.GetValue() != model.TypeStructValue {
		t.Errorf("Wrong type for pointer struct, expected TypeStructValue, got: %v", typeVal.GetValue())
	}

	// Test with nil entity
	var nilEntity *TypeValueTestStruct = nil
	_, err = GetEntityType(nilEntity)
	if err == nil {
		t.Errorf("GetEntityType should fail for nil entity")
	}

	// Test with interface entity
	var interfaceEntity interface{} = ptrEntity
	typeVal, err = GetEntityType(interfaceEntity)
	if err != nil {
		t.Errorf("GetEntityType failed for interface entity: %s", err.Error())
	}
	if typeVal.GetValue() != model.TypeStructValue {
		t.Errorf("Wrong type for interface entity, expected TypeStructValue, got: %v", typeVal.GetValue())
	}

	// Test with basic type (should work but not be a struct type)
	basicEntity := 42
	typeVal, err = GetEntityType(basicEntity)
	if err != nil {
		t.Errorf("GetEntityType failed for basic type: %s", err.Error())
	}
	if typeVal.GetValue() != model.TypeIntegerValue {
		t.Errorf("Wrong type for basic entity, expected TypeIntegerValue, got: %v", typeVal.GetValue())
	}
}

func TestGetEntityValueExtended(t *testing.T) {
	// Test GetEntityValue with various entity inputs
	entity := TypeValueTestStruct{ID: 1, Name: "test"}
	valueVal, err := GetEntityValue(entity)
	if err != nil {
		t.Errorf("GetEntityValue failed for struct: %s", err.Error())
	}
	if !valueVal.IsValid() {
		t.Errorf("Value should be valid for struct entity")
	}

	// Test with pointer entity
	ptrEntity := &TypeValueTestStruct{ID: 1, Name: "test"}
	valueVal, err = GetEntityValue(ptrEntity)
	if err != nil {
		t.Errorf("GetEntityValue failed for pointer struct: %s", err.Error())
	}
	if !valueVal.IsValid() {
		t.Errorf("Value should be valid for pointer struct entity")
	}

	// Test with nil entity
	var nilEntity *TypeValueTestStruct = nil
	_, err = GetEntityValue(nilEntity)
	if err == nil {
		t.Errorf("GetEntityValue should fail for nil entity")
	}

	// Test with interface entity
	var interfaceEntity interface{} = ptrEntity
	valueVal, err = GetEntityValue(interfaceEntity)
	if err != nil {
		t.Errorf("GetEntityValue failed for interface entity: %s", err.Error())
	}
	if !valueVal.IsValid() {
		t.Errorf("Value should be valid for interface entity")
	}

	// Test with basic type
	basicEntity := 42
	valueVal, err = GetEntityValue(basicEntity)
	if err != nil {
		t.Errorf("GetEntityValue failed for basic type: %s", err.Error())
	}
	if !valueVal.IsValid() {
		t.Errorf("Value should be valid for basic entity")
	}
}
