package remote

import (
	"encoding/json"
	"log"
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

	log.Print(obj)

	log.Print(string(data))

	tt := &SimpleObj{}
	err = json.Unmarshal(data, tt)
	if err != nil {
		t.Errorf("unmarshal obj failed, err:%s", err.Error())
		return
	}

	log.Print(*tt)
	t2 := &map[string]interface{}{}
	err = json.Unmarshal(data, t2)
	if err != nil {
		t.Errorf("unmarshal obj failed, err:%s", err.Error())
		return
	}

	log.Print(*t2)
}

func TestSimpleObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc, Age: 240}

	cache := NewCache()
	info, err := GetObject(obj, cache)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}
	log.Print(info)

	data, err := json.Marshal(&info)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}
	log.Print(string(data))

	info2 := Object{}
	err = json.Unmarshal(data, &info2)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}
	log.Print(info2)

	data, err = json.Marshal(info2)
	if err != nil {
		t.Errorf("marshal info2 failed, err:%s", err.Error())
		return
	}
	log.Print(string(data))
}

func TestExtObjInfo(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc}
	ext := &ExtObj{Name: "extObj", Obj: obj}

	cache := NewCache()
	info, err := GetObject(ext, cache)
	if err != nil {
		t.Errorf("GetObject failed, err:%s", err.Error())
		return
	}
	log.Print(info)

	data, err := json.Marshal(&info)
	if err != nil {
		t.Errorf("marshal info failed, err:%s", err.Error())
		return
	}
	log.Print(string(data))

	eInfo := Object{}
	json.Unmarshal(data, &eInfo)
	if err != nil {
		t.Errorf("unmarshal ext failed, err:%s", err.Error())
		return
	}
	log.Print(eInfo)

	data, err = json.Marshal(eInfo)
	if err != nil {
		t.Errorf("marshal eInfo failed, err:%s", err.Error())
		return
	}
	log.Print(string(data))
}
