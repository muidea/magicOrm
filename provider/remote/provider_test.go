package remote

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSimpleObjProvider(t *testing.T) {
	desc := "obj_desc"
	simple := Simple{Name: "simple", Desc: &desc, Age: 240, Flag: true, Add: []int{12, 34, 45}}

	simpleObj, simpleErr := GetObject(simple)
	if simpleErr != nil {
		t.Errorf("GetObject failed, err:%s", simpleErr.Error())
		return
	}

	data, dataErr := json.Marshal(simpleObj)
	if dataErr != nil {
		t.Errorf("marshal failed, err:%s", dataErr.Error())
		return
	}
	simpleInfo := &Object{}
	err := json.Unmarshal(data, simpleInfo)
	if err != nil {
		t.Errorf("unmarshal failed, err:%s", err.Error())
		return
	}

	simpleVal, simpleErr := GetObjectValue(simple)
	if simpleErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", simpleErr.Error())
		return
	}
	data, dataErr = json.Marshal(simpleVal)
	if dataErr != nil {
		t.Errorf("marshal failed, err:%s", dataErr.Error())
		return
	}
	simpleVal = &ObjectValue{}
	err = json.Unmarshal(data, simpleVal)
	if err != nil {
		t.Errorf("unmarshal failed, err:%s", err.Error())
		return
	}

	simpleModel, simpleErr := GetModel(reflect.ValueOf(simpleInfo))
	if simpleErr != nil {
		t.Errorf("GetModel failed, err:%s", simpleErr.Error())
		return
	}

	for _, val := range simpleModel.GetFields() {
		if val.IsAssigned() {
			t.Errorf("name:%s, check field assigned failed", val.GetName())
		}
	}

	simpleModel, simpleErr = SetModel(simpleModel, reflect.ValueOf(simpleVal))
	if simpleErr != nil {
		t.Errorf("SetModel failed, err:%s", simpleErr.Error())
		return
	}
	for _, val := range simpleModel.GetFields() {
		if !val.IsAssigned() {
			t.Errorf("name:%s, check field assigned failed", val.GetName())
		}
	}
}

func TestInterface(t *testing.T) {
	type AA struct {
		ID   int     `orm:"id key auto"`
		Name string  `orm:"name"`
		Desc *string `orm:"desc"`
	}

	type BB struct {
		ID   int    `orm:"id key auto"`
		Name string `orm:"name"`
		AA   *AA    `orm:"aa"`
	}

	bb := &BB{AA: &AA{}}
	bbVal, bbErr := GetObject(bb)
	if bbErr != nil {
		t.Errorf("GetObject failed, err:%s", bbErr.Error())
		return
	}

	valModel, objErr := GetModel(reflect.ValueOf(bbVal))
	if objErr != nil {
		t.Errorf("GetModel failed, err:%s", objErr.Error())
		return
	}

	for _, val := range valModel.GetFields() {
		vName := val.GetName()
		if val.IsAssigned() {
			if vName != "AA" {
				t.Errorf("name:%s, check field assigned failed", vName)
			}
		}
		if val.IsAssigned() {
			t.Errorf("name:%s, check field assigned failed", vName)
		}
	}
}
