package remote

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/muidea/magicOrm/model"
)

func toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

type Simple struct {
	//ID 唯一标示单元
	ID     int64   `orm:"id key" view:"detail,lite"`
	Name   string  `orm:"name" view:"detail,lite"`
	Desc   *string `orm:"desc" view:"detail"`
	Age    uint8   `orm:"age" view:"detail,lite"`
	Flag   bool    `orm:"flag" view:"detail,lite"`
	Add    []int   `orm:"add" view:"detail,lite"`
	AddPtr *[]int  `orm:"addPtr" view:"detail,lite"`
}

type ExtInfo struct {
	ID       int64     `orm:"id key" view:"detail,lite"`
	Name     string    `orm:"name" view:"detail,lite"`
	Obj      Simple    `orm:"obj" view:"detail,lite"`
	ObjPtr   *Simple   `orm:"objPtr" view:"detail,lite"`
	ObjArray []*Simple `orm:"array" view:"detail,lite"`
}

type Complex struct {
	ID       int64      `orm:"id key" view:"detail,lite"`
	Name     string     `orm:"name" view:"detail,lite"`
	Info     ExtInfo    `orm:"info" view:"detail"`
	InfoPtr  *ExtInfo   `orm:"infoPtr" view:"detail"`
	Array    []ExtInfo  `orm:"array" view:"detail"`
	ArrayPtr []*ExtInfo `orm:"arrayPtr" view:"detail"`
}

func TestSpec(t *testing.T) {
	spec := ""
	_, err := getOrmSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}

	spec = "test"
	itemSpec, err := getOrmSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}
	if itemSpec.GetFieldName() != "test" {
		t.Errorf("illegal spec name")
		return
	}
	if itemSpec.IsPrimaryKey() {
		t.Errorf("illegal spec define")
		return
	}
	if itemSpec.GetValueDeclare() == model.AutoIncrement {
		t.Errorf("illegal spec define")
		return
	}

	spec = "test auto key"
	itemSpec, err = getOrmSpec(spec)
	if err != nil {
		t.Errorf("illegal spec value")
		return
	}
	if itemSpec.GetFieldName() != "test" {
		t.Errorf("illegal spec name")
		return
	}
	if !itemSpec.IsPrimaryKey() {
		t.Errorf("illegal spec define")
		return
	}
	if itemSpec.GetValueDeclare() != model.AutoIncrement {
		t.Errorf("illegal spec define")
		return
	}
}

func TestSimpleObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc, Age: 240}

	info, err := GetObject(obj)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}
	if info.GetName() != "Simple" {
		t.Errorf("GetObject failed")
	}

	byteVal, byteErr := json.Marshal(info)
	if byteErr != nil {
		t.Errorf("marshal info failed, err:%s", byteErr.Error())
		return
	}

	info2 := &Object{}
	byteErr = json.Unmarshal(byteVal, info2)
	if byteErr != nil {
		t.Errorf("marshal info failed, err:%s", byteErr.Error())
		return
	}

	if !CompareObject(info, info2) {
		t.Errorf("unmarshal failed")
		return
	}
}

func TestExtObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc}
	ext := &ExtInfo{Name: "extObj", Obj: obj, ObjArray: []*Simple{&obj, &obj}}

	info, err := GetObject(ext)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}

	if info.GetName() != "ExtInfo" {
		t.Errorf("get object failed")
		return
	}

	byteVal, byteErr := json.Marshal(info)
	if byteErr != nil {
		t.Errorf("marshal info failed, err:%s", byteErr.Error())
		return
	}

	eInfo := &Object{}
	byteErr = json.Unmarshal(byteVal, eInfo)
	if byteErr != nil {
		t.Errorf("unmarshal ext failed, err:%s", byteErr.Error())
		return
	}

	if !CompareObject(info, eInfo) {
		t.Errorf("unmarshal faile")
		return
	}
}

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

	data, err := EncodeObjectValue(rawVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}

	curVal, curErr := DecodeObjectValue(data)
	if curErr != nil {
		t.Errorf("decode obj failed, err:%s", curErr.Error())
		return
	}

	if !CompareObjectValue(rawVal, curVal) {
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

	data, err := EncodeObjectValue(objVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}

	objInfo, objErr := DecodeObjectValue(data)
	if objErr != nil {
		t.Errorf("DecodeObjectValue failed, err:%s", objErr.Error())
		return
	}

	if !CompareObjectValue(objVal, objInfo) {
		t.Errorf("compareObjectValue failed")
		return
	}
}

func TestGetObjectWithNilValue(t *testing.T) {
	var ptr *Simple = nil
	_, err := GetObject(ptr)
	if err == nil {
		t.Errorf("GetObject with nil value should return error")
		return
	}

	var extPtr *ExtInfo = nil
	_, err = GetObject(extPtr)
	if err == nil {
		t.Errorf("GetObject with nil value should return error")
		return
	}
}

func TestGetObjectWithStructPointers(t *testing.T) {
	desc := "obj_desc"
	obj := &Simple{Name: "obj", Desc: &desc, Age: 240}

	info, err := GetObject(obj)
	if err != nil {
		t.Errorf("GetObject failed with pointer, err:%s", err.Error())
		return
	}
	if info.GetName() != "Simple" {
		t.Errorf("GetObject failed with pointer")
	}

	// Test with a nested struct pointer
	ext := ExtInfo{Name: "extObj", ObjPtr: obj}
	extInfo, extErr := GetObject(ext)
	if extErr != nil {
		t.Errorf("GetObject failed with nested pointer, err:%s", extErr.Error())
		return
	}

	objPtrField := extInfo.GetField("objPtr")
	if objPtrField == nil {
		t.Errorf("Failed to get objPtr field")
		return
	}

	if !objPtrField.IsPtrType() {
		t.Errorf("objPtr field should be a pointer type")
		return
	}
}

func TestComplexObjInfo(t *testing.T) {
	desc := "obj_desc"
	simple := Simple{Name: "simple", Desc: &desc, Age: 240}
	ext := ExtInfo{Name: "extObj", Obj: simple, ObjPtr: &simple, ObjArray: []*Simple{&simple}}
	complex := Complex{
		Name:     "complex",
		Info:     ext,
		InfoPtr:  &ext,
		Array:    []ExtInfo{ext},
		ArrayPtr: []*ExtInfo{&ext},
	}

	info, err := GetObject(complex)
	if err != nil {
		t.Errorf("GetObject failed for complex obj, err:%s", err.Error())
		return
	}

	if info.GetName() != "Complex" {
		t.Errorf("GetObject failed for complex obj")
		return
	}

	// Check nested fields
	infoField := info.GetField("info")
	if infoField == nil {
		t.Errorf("Failed to get info field")
		return
	}

	if !infoField.IsStruct() {
		t.Errorf("info field should be a struct type")
		return
	}

	arrayField := info.GetField("array")
	if arrayField == nil {
		t.Errorf("Failed to get array field")
		return
	}

	if !arrayField.IsSlice() {
		t.Errorf("array field should be a slice type")
		return
	}

	// Test serialization and deserialization
	byteVal, byteErr := json.Marshal(info)
	if byteErr != nil {
		t.Errorf("marshal complex info failed, err:%s", byteErr.Error())
		return
	}

	deserializedInfo := &Object{}
	byteErr = json.Unmarshal(byteVal, deserializedInfo)
	if byteErr != nil {
		t.Errorf("unmarshal complex info failed, err:%s", byteErr.Error())
		return
	}

	if !CompareObject(info, deserializedInfo) {
		t.Errorf("deserialization failed for complex object")
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
	expectedPkgPath := "github.com/muidea/magicOrm/provider/remote"
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
	encodedData, encodeErr := EncodeSliceObjectValue(sliceVal)
	if encodeErr != nil {
		t.Errorf("EncodeSliceObjectValue failed, err:%s", encodeErr.Error())
		return
	}

	decodedVal, decodeErr := DecodeSliceObjectValue(encodedData)
	if decodeErr != nil {
		t.Errorf("DecodeSliceObjectValue failed, err:%s", decodeErr.Error())
		return
	}

	if !CompareSliceObjectValue(sliceVal, decodedVal) {
		t.Errorf("SliceObjectValue serialization/deserialization failed")
		return
	}

	// Test Copy functionality
	copiedVal := sliceVal.Copy()
	if !CompareSliceObjectValue(sliceVal, copiedVal) {
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
