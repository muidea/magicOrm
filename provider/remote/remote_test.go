package remote

import (
	"encoding/json"
	"testing"

	"github.com/muidea/magicOrm/model"
)

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
