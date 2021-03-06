package remote

import (
	"encoding/json"
	"testing"
)

type Simple struct {
	//ID 唯一标示单元
	ID   int64   `orm:"id key"`
	Name string  `orm:"name"`
	Desc *string `orm:"desc"`
	Age  uint8   `orm:"age"`
	Flag bool    `orm:"flag"`
	Add  []int   `orm:"add"`
}

type ExtInfo struct {
	ID       int64     `orm:"id key"`
	Name     string    `orm:"name"`
	Obj      Simple    `orm:"obj"`
	ObjPtr   *Simple   `orm:"objPtr"`
	ObjArray []*Simple `orm:"array"`
}

func TestSimpleObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc, Age: 240}

	info, err := GetObject(obj)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}
	if info.GetName() != "remote.Simple" {
		t.Errorf("GetObject failed")
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}

	info2 := &Object{}
	err = json.Unmarshal(data, info2)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}

	if !compareObject(info, info2) {
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

	if info.GetName() != "remote.ExtInfo" {
		t.Errorf("get object failed")
		return
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}

	eInfo := &Object{}
	err = json.Unmarshal(data, eInfo)
	if err != nil {
		t.Errorf("unmarshal ext failed, err:%s", err.Error())
		return
	}

	if !compareObject(info, eInfo) {
		t.Errorf("unmarshal faile")
		return
	}
}
