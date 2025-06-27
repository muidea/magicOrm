package helper

import (
	"encoding/json"
	"testing"

	"github.com/muidea/magicOrm/model"
	"github.com/muidea/magicOrm/provider/remote"
	"github.com/stretchr/testify/assert"
)

type Simple struct {
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

	info2 := &remote.Object{}
	byteErr = json.Unmarshal(byteVal, info2)
	if byteErr != nil {
		t.Errorf("marshal info failed, err:%s", byteErr.Error())
		return
	}

	if !remote.CompareObject(info, info2) {
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

	eInfo := &remote.Object{}
	byteErr = json.Unmarshal(byteVal, eInfo)
	if byteErr != nil {
		t.Errorf("unmarshal ext failed, err:%s", byteErr.Error())
		return
	}

	if !remote.CompareObject(info, eInfo) {
		t.Errorf("unmarshal faile")
		return
	}
}

func TestGetObjectWithNilValue(t *testing.T) {
	var simplePtr *Simple = nil
	simpleValPtr, simpleValErr := GetObject(simplePtr)
	if simpleValErr != nil {
		t.Errorf("GetObject with nil value should return error, err:%s", simpleValErr.Error())
		return
	}
	assert.NotNil(t, simpleValPtr)

	var extPtr *ExtInfo = nil
	extValPtr, extValErr := GetObject(extPtr)
	if extValErr != nil {
		t.Errorf("GetObject with nil value should return error, err:%s", extValErr.Error())
		return
	}
	assert.NotNil(t, extValPtr)
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

	if !model.IsPtrField(objPtrField) {
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

	if !model.IsStructField(infoField) {
		t.Errorf("info field should be a struct type")
		return
	}

	arrayField := info.GetField("array")
	if arrayField == nil {
		t.Errorf("Failed to get array field")
		return
	}

	if !model.IsSliceField(arrayField) {
		t.Errorf("array field should be a slice type")
		return
	}

	// Test serialization and deserialization
	byteVal, byteErr := json.Marshal(info)
	if byteErr != nil {
		t.Errorf("marshal complex info failed, err:%s", byteErr.Error())
		return
	}

	deserializedInfo := &remote.Object{}
	byteErr = json.Unmarshal(byteVal, deserializedInfo)
	if byteErr != nil {
		t.Errorf("unmarshal complex info failed, err:%s", byteErr.Error())
		return
	}

	if !remote.CompareObject(info, deserializedInfo) {
		t.Errorf("deserialization failed for complex object")
		return
	}
}
