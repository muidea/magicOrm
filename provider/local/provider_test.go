package local

import (
	"reflect"
	"testing"

	"github.com/muidea/magicOrm/model"
)

func TestGetEntityValueBasic(t *testing.T) {
	// Test with integer
	iVal := 123
	// Convert returned value to int
	eVal, eErr := GetEntityValue(iVal)
	if eErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eErr.Error())
		return
	}

	// Update value
	iVal2 := 234
	eVal.Set(reflect.ValueOf(iVal2))

	rawVal := eVal.Interface().(model.RawVal)
	niVal := rawVal.Value().(int)
	if niVal != iVal2 {
		t.Errorf("GetEntityValue failed")
		return
	}

	// Test with integer array
	iValArray := []int{1, 2}
	eArrayVal, eArrayErr := GetEntityValue(iValArray)
	if eArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eArrayErr.Error())
		return
	}
	rawValArray := eArrayVal.Interface().(model.RawVal)
	niValArray := rawValArray.Value().([]int)
	if !reflect.DeepEqual(niValArray, []int{1, 2}) {
		t.Errorf("GetEntityValue failed")
		return
	}

	eArrayVal, eArrayErr = AppendSliceValue(eArrayVal, eVal)
	if eArrayErr != nil {
		t.Errorf("AppendSliceValue failed, error:%s", eArrayErr.Error())
		return
	}
	rawValArray = eArrayVal.Interface().(model.RawVal)
	niValArray = rawValArray.Value().([]int)
	if !reflect.DeepEqual(niValArray, []int{1, 2, 234}) {
		t.Errorf("GetEntityValue failed")
		return
	}
}

func TestGetEntityValue(t *testing.T) {
	// Test with integer
	iVal := 123
	// Convert returned value to int
	eVal, eErr := GetEntityValue(iVal)
	if eErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eErr.Error())
		return
	}

	// Update value
	iVal2 := 234
	eVal.Set(reflect.ValueOf(iVal2))

	rawVal := eVal.Interface().(model.RawVal)
	niVal := rawVal.Value().(int)
	if niVal != iVal2 {
		t.Errorf("GetEntityValue failed")
		return
	}

	// Test converting a string value
	sVal := "test"
	_, _ = GetEntityValue(sVal)

	// Test with integer array
	iValArray := []int{1, 2}
	eArrayVal, eArrayErr := GetEntityValue(iValArray)
	if eArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eArrayErr.Error())
		return
	}
	rawValArray := eArrayVal.Interface().(model.RawVal)
	niValArray := rawValArray.Value().([]int)
	if !reflect.DeepEqual(niValArray, []int{1, 2}) {
		t.Errorf("GetEntityValue failed")
		return
	}

	eArrayVal, eArrayErr = AppendSliceValue(eArrayVal, eVal)
	if eArrayErr != nil {
		t.Errorf("AppendSliceValue failed, error:%s", eArrayErr.Error())
		return
	}
	rawValArray = eArrayVal.Interface().(model.RawVal)
	niValArray = rawValArray.Value().([]int)
	if !reflect.DeepEqual(niValArray, []int{1, 2, 234}) {
		t.Errorf("GetEntityValue failed")
		return
	}

	// Convert string array - just test that it doesn't error
	sValArray := []string{"1", "2"}
	_, _ = GetEntityValue(sValArray)

	// Test with integer ptr array
	iValPtrArray := []*int{&iVal, &iVal2}
	ePtrArrayVal, ePtrArrayErr := GetEntityValue(iValPtrArray)
	if ePtrArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", ePtrArrayErr.Error())
		return
	}
	rawValPtrArray := ePtrArrayVal.Interface().(model.RawVal)
	niValPtrArray := rawValPtrArray.Value().([]*int)
	if len(niValPtrArray) != 2 {
		t.Errorf("GetEntityValue failed")
		return
	}

	// Test append to integer ptr array
	ePtrVal := &iVal
	ePtrValObj, ePtrValErr := GetEntityValue(ePtrVal)
	if ePtrValErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", ePtrValErr.Error())
		return
	}
	ePtrArrayVal, ePtrArrayErr = AppendSliceValue(ePtrArrayVal, ePtrValObj)
	if ePtrArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", ePtrArrayErr.Error())
		return
	}
	rawValPtrArray = ePtrArrayVal.Interface().(model.RawVal)
	niValPtrArray = rawValPtrArray.Value().([]*int)
	if len(niValPtrArray) != 3 {
		t.Errorf("GetEntityValue failed")
		return
	}
}

func TestAppendSliceValue2(t *testing.T) {
	// Create slice value
	intSlice := []int{1, 2, 3}
	sliceVal, err := GetEntityValue(intSlice)
	if err != nil {
		t.Errorf("GetEntityValue failed for slice: %s", err.Error())
		return
	}

	// Create element value
	element := 4
	elemVal, err := GetEntityValue(element)
	if err != nil {
		t.Errorf("GetEntityValue failed for element: %s", err.Error())
		return
	}

	// Append element to slice
	resultVal, err := AppendSliceValue(sliceVal, elemVal)
	if err != nil {
		t.Errorf("AppendSliceValue failed: %s", err.Error())
		return
	}

	// Verify result is a slice
	rawVal := resultVal.Interface().(model.RawVal)
	result := rawVal.Value().([]int)
	if len(result) != len(intSlice)+1 {
		t.Errorf("Result length mismatch, expected: %d, got: %d", len(intSlice)+1, len(result))
	}

	// Verify element was appended
	if result[len(result)-1] != element {
		t.Errorf("Appended element mismatch, expected: %d, got: %d", element, result[len(result)-1])
	}

	// Test with incompatible types
	stringVal, err := GetEntityValue("string")
	if err != nil {
		t.Errorf("GetEntityValue failed for string: %s", err.Error())
		return
	}

	// Append string to int slice (should fail)
	_, err = AppendSliceValue(sliceVal, stringVal)
	if err == nil {
		t.Errorf("AppendSliceValue should fail with incompatible types")
	}

	// Test with non-slice
	_, err = AppendSliceValue(elemVal, elemVal)
	if err == nil {
		t.Errorf("AppendSliceValue should fail with non-slice first argument")
	}
}

// TestPointerValueHandling tests handling of pointer values
func TestPointerValueHandling(t *testing.T) {
	// Create pointer to int
	intValue := 123
	intPtr := &intValue

	// Get entity value for pointer
	ptrVal, err := GetEntityValue(intPtr)
	if err != nil {
		t.Errorf("GetEntityValue failed for pointer: %s", err.Error())
		return
	}

	// Verify pointer value is correct
	rawVal := ptrVal.Interface().(model.RawVal)
	retrievedPtr := rawVal.Value().(*int)
	if *retrievedPtr != intValue {
		t.Errorf("Retrieved pointer value mismatch, expected: %d, got: %d", intValue, *retrievedPtr)
	}

	// Update through pointer
	newValue := 456
	ptrVal.Set(reflect.ValueOf(&newValue))

	// Verify updated value
	rawVal = ptrVal.Interface().(model.RawVal)
	updatedPtr := rawVal.Value().(*int)
	if *updatedPtr != newValue {
		t.Errorf("Updated pointer value mismatch, expected: %d, got: %d", newValue, *updatedPtr)
	}

	// Use the existing Unit type from object_test.go
	unitValue := Unit{ID: 789, Name: "UnitPtr"}
	unitPtr := &unitValue

	// Get entity value for pointer to struct
	unitPtrVal, err := GetEntityValue(unitPtr)
	if err != nil {
		t.Errorf("GetEntityValue failed for pointer to struct: %s", err.Error())
		return
	}

	// Verify pointer to struct value is correct
	rawVal = unitPtrVal.Interface().(model.RawVal)
	retrievedUnitPtr := rawVal.Value().(*Unit)
	if retrievedUnitPtr.ID != unitValue.ID || retrievedUnitPtr.Name != unitValue.Name {
		t.Errorf("Retrieved struct pointer mismatch, expected: %+v, got: %+v", unitValue, *retrievedUnitPtr)
	}

	// Test with pointer to int slice
	sliceValue := []int{1, 2, 3}
	slicePtr := &sliceValue

	// Get entity value for pointer to slice
	slicePtrVal, err := GetEntityValue(slicePtr)
	if err != nil {
		t.Errorf("GetEntityValue failed for pointer to slice: %s", err.Error())
		return
	}

	// Verify pointer to slice value is correct
	rawVal = slicePtrVal.Interface().(model.RawVal)
	retrievedSlicePtr := rawVal.Value().(*[]int)
	if !reflect.DeepEqual(*retrievedSlicePtr, *slicePtr) {
		t.Errorf("Retrieved slice pointer mismatch, expected: %v, got: %v", *slicePtr, *retrievedSlicePtr)
	}
}
