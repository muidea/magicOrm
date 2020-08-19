package remote

import (
	"encoding/json"
	"testing"
)

type Simple struct {
	Name string  `orm:"name"`
	Desc *string `orm:"desc"`
	Age  uint8   `orm:"age"`
	Flag bool    `orm:"flag"`
	Add  []int   `orm:"add"`
}

type ExtInfo struct {
	Name string `orm:"name"`
	Obj  Simple `orm:"obj"`
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

	data, err := json.Marshal(&info)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}

	info2 := Object{}
	err = json.Unmarshal(data, &info2)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}

	if info2.GetName() != info.GetName() {
		t.Errorf("unmarshal failed")
		return
	}
}

func TestExtObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc}
	ext := &ExtInfo{Name: "extObj", Obj: obj}

	info, err := GetObject(ext)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}

	if info.GetName() != "remote.ExtInfo" {
		t.Errorf("get object failed")
		return
	}

	data, err := json.Marshal(&info)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}

	eInfo := Object{}
	err = json.Unmarshal(data, &eInfo)
	if err != nil {
		t.Errorf("unmarshal ext failed, err:%s", err.Error())
		return
	}

	if eInfo.GetName() != info.GetName() {
		t.Errorf("unmarshal faile")
		return
	}
}
