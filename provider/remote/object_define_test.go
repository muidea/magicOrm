package remote

import (
	"encoding/json"
	"testing"
)

type SimpleObj struct {
	Name string  `orm:"name"`
	Desc *string `orm:"desc"`
	Age  uint8   `orm:"age"`
	Flag bool    `orm:"flag"`
	Add  []int   `orm:"add"`
}

type ExtObj struct {
	Name string    `orm:"name"`
	Obj  SimpleObj `orm:"obj"`
}

func TestSimpleObj(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc, Age: 240, Add: []int{12, 34, 45}}

	data, err := json.Marshal(&obj)
	if err != nil {
		t.Errorf("marshal obj failed, err:%s", err.Error())
		return
	}

	tt := &SimpleObj{}
	err = json.Unmarshal(data, tt)
	if err != nil {
		t.Errorf("unmarshal obj failed, err:%s", err.Error())
		return
	}

	if tt.Name != obj.Name {
		t.Errorf("unmarshal obj failed")
		return
	}

	t2 := &map[string]interface{}{}
	err = json.Unmarshal(data, t2)
	if err != nil {
		t.Errorf("unmarshal obj failed, err:%s", err.Error())
		return
	}
	if len(*t2) != 5 {
		t.Errorf("unmarshal obj failed")
	}
}

func TestSimpleObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc, Age: 240}

	info, err := GetObject(obj)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}
	if info.GetName() != "remote.SimpleObj" {
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
	}
}

func TestExtObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc}
	ext := &ExtObj{Name: "extObj", Obj: obj}

	info, err := GetObject(ext)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}

	if info.GetName() != "remote.ExtObj" {
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
