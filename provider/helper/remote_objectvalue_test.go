package helper

import (
	"testing"

	"github.com/muidea/magicOrm/provider/remote"
)

func TestSimpleValue(t *testing.T) {
	desc := "obj_desc"
	iVal := 123
	obj := Simple{Name: "obj", Desc: &desc, Age: 240, Add: []int{12, 34, 45}, AddPtr: &[]int{iVal, iVal}}

	rawVal, rawErr := GetObjectValue(obj)
	if rawErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", rawErr.Error())
		return
	}

	if !rawVal.IsAssigned() {
		t.Errorf("check object is assigned failed")
		return
	}

	data, err := remote.EncodeObjectValue(rawVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}

	curVal, curErr := remote.DecodeObjectValue(data)
	if curErr != nil {
		t.Errorf("decode obj failed, err:%s", curErr.Error())
		return
	}

	if !remote.CompareObjectValue(rawVal, curVal) {
		t.Errorf("compareObjectValue failed")
		return
	}

	obj2 := Simple{}
	rawVal, rawErr = GetObjectValue(obj2)
	if rawErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", rawErr.Error())
		return
	}

	if rawVal.IsAssigned() {
		t.Errorf("check object is assigned failed")
		return
	}
}

func TestExtObjValue(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc, Add: []int{12, 223, 456}}
	ext := &ExtInfo{Name: "extObj", Obj: obj, ObjArray: []*Simple{&obj, &obj}}

	objVal, objErr := GetObjectValue(ext)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	data, err := remote.EncodeObjectValue(objVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}

	objInfo, objErr := remote.DecodeObjectValue(data)
	if objErr != nil {
		t.Errorf("DecodeObjectValue failed, err:%s", objErr.Error())
		return
	}

	if !remote.CompareObjectValue(objVal, objInfo) {
		t.Errorf("compareObjectValue failed")
		return
	}
}

func TestSliceObjectValueManipulation(t *testing.T) {
	// Create test data
	simple1 := Simple{ID: 1, Name: "obj1", Age: 20}
	simple2 := Simple{ID: 2, Name: "obj2", Age: 30}

	// Test getting slice value
	simples := []Simple{simple1, simple2}

	sliceVal, err := GetSliceObjectValue(simples)
	if err != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", err.Error())
		return
	}

	// Check basic properties
	if sliceVal.GetName() != "Simple" {
		t.Errorf("SliceObjectValue name mismatch, expected 'Simple', got '%s'", sliceVal.GetName())
		return
	}

	// The PkgPath will be the package of this test file, which is "github.com/muidea/magicOrm/provider/remote"
	expectedPkgPath := "github.com/muidea/magicOrm/provider/helper"
	if sliceVal.GetPkgPath() != expectedPkgPath {
		t.Errorf("SliceObjectValue pkgPath mismatch, expected '%s', got '%s'", expectedPkgPath, sliceVal.GetPkgPath())
		return
	}

	// Check values
	values := sliceVal.GetValue()
	if len(values) != 2 {
		t.Errorf("SliceObjectValue should have 2 values, got %d", len(values))
		return
	}

	// Check individual values
	obj1 := values[0]
	if obj1.GetFieldValue("name") != "obj1" {
		t.Errorf("First object name mismatch, expected 'obj1', got '%v'", obj1.GetFieldValue("name"))
		return
	}

	obj2 := values[1]
	if obj2.GetFieldValue("name") != "obj2" {
		t.Errorf("Second object name mismatch, expected 'obj2', got '%v'", obj2.GetFieldValue("name"))
		return
	}

	// Test slice modification
	// Add a new field value to obj1
	obj1.SetFieldValue("age", uint8(25))
	updatedAge := obj1.GetFieldValue("age")
	if updatedAge != uint8(25) {
		t.Errorf("Failed to update field value, expected age 25, got %v", updatedAge)
		return
	}

	// Test serialization
	encodedData, encodeErr := remote.EncodeSliceObjectValue(sliceVal)
	if encodeErr != nil {
		t.Errorf("EncodeSliceObjectValue failed, err:%s", encodeErr.Error())
		return
	}

	decodedVal, decodeErr := remote.DecodeSliceObjectValue(encodedData)
	if decodeErr != nil {
		t.Errorf("DecodeSliceObjectValue failed, err:%s", decodeErr.Error())
		return
	}

	if !remote.CompareSliceObjectValue(sliceVal, decodedVal) {
		t.Errorf("SliceObjectValue serialization/deserialization failed")
		return
	}

	// Test Copy functionality
	copiedVal := sliceVal.Copy()
	if !remote.CompareSliceObjectValue(sliceVal, copiedVal) {
		t.Errorf("SliceObjectValue Copy() failed, copied value doesn't match original")
		return
	}

	// Modify original and verify copy is unaffected
	sliceVal.Values[0].SetFieldValue("name", "modified_obj1")
	if copiedVal.Values[0].GetFieldValue("name") == "modified_obj1" {
		t.Errorf("Copy not independent from original, modification affected the copy")
		return
	}
}

func TestGetObjectValuePreservesNilAndEmptyStructSlices(t *testing.T) {
	objWithNilSlices := &Compose{ID: 1, Name: "nil-slices"}

	nilVal, err := GetObjectValue(objWithNilSlices)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	baseArrayVal, ok := nilVal.GetFieldValue("baseArray").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("baseArray should be *remote.SliceObjectValue, got %T", nilVal.GetFieldValue("baseArray"))
	}
	if baseArrayVal.Values != nil {
		t.Fatalf("nil slice should remain unassigned, got %#v", baseArrayVal.Values)
	}

	basePtrArrayVal, ok := nilVal.GetFieldValue("basePtrArray").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("basePtrArray should be *remote.SliceObjectValue, got %T", nilVal.GetFieldValue("basePtrArray"))
	}
	if basePtrArrayVal.Values != nil {
		t.Fatalf("nil pointer slice should remain unassigned, got %#v", basePtrArrayVal.Values)
	}

	emptyPtrArray := []*Base{}
	objWithEmptySlices := &Compose{
		ID:              2,
		Name:            "empty-slices",
		BaseArray:       []Base{},
		BasePtrArray:    []*Base{},
		BasePtrArrayPtr: &emptyPtrArray,
	}

	emptyVal, err := GetObjectValue(objWithEmptySlices)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	baseArrayVal, ok = emptyVal.GetFieldValue("baseArray").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("baseArray should be *remote.SliceObjectValue, got %T", emptyVal.GetFieldValue("baseArray"))
	}
	if baseArrayVal.Values == nil || len(baseArrayVal.Values) != 0 {
		t.Fatalf("empty slice should remain assigned empty, got %#v", baseArrayVal.Values)
	}

	basePtrArrayVal, ok = emptyVal.GetFieldValue("basePtrArray").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("basePtrArray should be *remote.SliceObjectValue, got %T", emptyVal.GetFieldValue("basePtrArray"))
	}
	if basePtrArrayVal.Values == nil || len(basePtrArrayVal.Values) != 0 {
		t.Fatalf("empty pointer slice should remain assigned empty, got %#v", basePtrArrayVal.Values)
	}

	basePtrArrayPtrVal, ok := emptyVal.GetFieldValue("basePtrArrayPtr").(*remote.SliceObjectValue)
	if !ok {
		t.Fatalf("basePtrArrayPtr should be *remote.SliceObjectValue, got %T", emptyVal.GetFieldValue("basePtrArrayPtr"))
	}
	if basePtrArrayPtrVal.Values == nil || len(basePtrArrayPtrVal.Values) != 0 {
		t.Fatalf("empty pointer-to-slice should remain assigned empty, got %#v", basePtrArrayPtrVal.Values)
	}
}

func TestUpdateEntityPreservesAssignedEmptyStructSlices(t *testing.T) {
	emptyPtrArray := []*Base{}
	source := &Compose{
		ID:              3,
		Name:            "assigned-empty",
		BaseArray:       []Base{},
		BasePtrArray:    []*Base{},
		BasePtrArrayPtr: &emptyPtrArray,
	}

	objVal, err := GetObjectValue(source)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	target := &Compose{}
	err = UpdateEntity(objVal, target)
	if err != nil {
		t.Fatalf("UpdateEntity failed: %v", err)
	}

	if target.BaseArray == nil || len(target.BaseArray) != 0 {
		t.Fatalf("BaseArray should remain assigned empty, got %#v", target.BaseArray)
	}
	if target.BasePtrArray == nil || len(target.BasePtrArray) != 0 {
		t.Fatalf("BasePtrArray should remain assigned empty, got %#v", target.BasePtrArray)
	}
	if target.BasePtrArrayPtr == nil || *target.BasePtrArrayPtr == nil || len(*target.BasePtrArrayPtr) != 0 {
		t.Fatalf("BasePtrArrayPtr should remain assigned empty, got %#v", target.BasePtrArrayPtr)
	}
}

func TestGetObjectValueTypedNilReturnsError(t *testing.T) {
	var simplePtr *Simple
	if _, err := GetObjectValue(simplePtr); err == nil {
		t.Fatal("GetObjectValue((*Simple)(nil)) should return error")
	}

	var remoteObjPtr *remote.Object
	if _, err := GetObjectValue(remoteObjPtr); err == nil {
		t.Fatal("GetObjectValue((*remote.Object)(nil)) should return error")
	}

	var remoteObjValuePtr *remote.ObjectValue
	if _, err := GetObjectValue(remoteObjValuePtr); err == nil {
		t.Fatal("GetObjectValue((*remote.ObjectValue)(nil)) should return error")
	}
}

func TestGetObjectValueLeavesNilPointerFieldsUnassigned(t *testing.T) {
	person := &Person{
		ID:   7,
		Name: "tester",
	}

	objVal, err := GetObjectValue(person)
	if err != nil {
		t.Fatalf("GetObjectValue failed: %v", err)
	}

	var addrField *remote.FieldValue
	for _, field := range objVal.Fields {
		if field.Name == "addr" {
			addrField = field
			break
		}
	}
	if addrField == nil {
		t.Fatal("addr field should be exported")
	}
	if addrField.Value != nil {
		t.Fatalf("nil pointer field should keep nil value, got %#v", addrField.Value)
	}
	if addrField.Assigned {
		t.Fatalf("helper-generated nil pointer field should stay unassigned, got %#v", addrField)
	}
}

func TestGetObjectValueMarksExplicitZeroPointersAndSlicesAssigned(t *testing.T) {
	age := 0
	intArray := []int{}
	base := &Base{
		ID:        9,
		IArray:    []int{},
		IArrayPtr: &intArray,
	}
	person := &Person{
		ID:   8,
		Name: "zero",
		Age:  &age,
	}

	baseVal, err := GetObjectValue(base)
	if err != nil {
		t.Fatalf("GetObjectValue(base) failed: %v", err)
	}
	if iArray := findFieldValue(baseVal, "iArray"); iArray == nil || !iArray.Assigned {
		t.Fatalf("empty basic slice should be treated as assigned, got %#v", iArray)
	}
	if iArrayPtr := findFieldValue(baseVal, "iArrayPtr"); iArrayPtr == nil || !iArrayPtr.Assigned {
		t.Fatalf("empty pointer slice should be treated as assigned, got %#v", iArrayPtr)
	}

	personVal, err := GetObjectValue(person)
	if err != nil {
		t.Fatalf("GetObjectValue(person) failed: %v", err)
	}
	if ageField := findFieldValue(personVal, "age"); ageField == nil || !ageField.Assigned {
		t.Fatalf("non-nil zero pointer should be treated as assigned, got %#v", ageField)
	}
}

func findFieldValue(objVal *remote.ObjectValue, name string) *remote.FieldValue {
	for _, field := range objVal.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func TestGetSliceObjectValueTypedNilReturnsError(t *testing.T) {
	var slicePtr *[]*Simple
	if _, err := GetSliceObjectValue(slicePtr); err == nil {
		t.Fatal("GetSliceObjectValue((*[]*Simple)(nil)) should return error")
	}

	var remoteSlicePtr *remote.SliceObjectValue
	if _, err := GetSliceObjectValue(remoteSlicePtr); err == nil {
		t.Fatal("GetSliceObjectValue((*remote.SliceObjectValue)(nil)) should return error")
	}
}
