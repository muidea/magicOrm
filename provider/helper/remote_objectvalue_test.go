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
	expectedPkgPath := "github.com/muidea/magicOrm/provider/helper/Simple"
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
