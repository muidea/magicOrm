package remote

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

func TestSimpleObjProvider(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc, Age: 240, Flag: true, Add: []int{12, 34, 45}}

	objInfo, objErr := GetObject(obj)
	if objErr != nil {
		log.Printf("GetObject failed, err:%s", objErr.Error())
		return
	}

	data, dataErr := json.Marshal(objInfo)
	if dataErr != nil {
		log.Printf("marshal failed, err:%s", dataErr.Error())
		return
	}
	localInfo := &Object{}
	err := json.Unmarshal(data, localInfo)
	if err != nil {
		log.Printf("unmarshal failed, err:%s", err.Error())
		return
	}

	objVal, objErr := GetObjectValue(obj)
	if objErr != nil {
		log.Printf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}
	data, dataErr = json.Marshal(objVal)
	if dataErr != nil {
		log.Printf("marshal failed, err:%s", dataErr.Error())
		return
	}
	localVal := &ObjectValue{}
	err = json.Unmarshal(data, localVal)
	if err != nil {
		log.Printf("unmarshal failed, err:%s", err.Error())
		return
	}

	infoModel, infoErr := GetModel(reflect.ValueOf(localInfo))
	if infoErr != nil {
		log.Printf("GetModel failed, err:%s", infoErr.Error())
		return
	}

	for _, val := range infoModel.GetFields() {
		if val.IsAssigned() {
			t.Errorf("name:%s, check field assigned failed", val.GetName())
		}
	}

	infoModel, infoErr = SetModel(infoModel, reflect.ValueOf(localVal))
	if infoErr != nil {
		log.Printf("SetModel failed, err:%s", infoErr.Error())
		return
	}
	for _, val := range infoModel.GetFields() {
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
	bbVal, bbErr := GetObjectValue(bb)
	if bbErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", bbErr.Error())
		return
	}

	valModel, objErr := GetModel(reflect.ValueOf(bbVal))
	if objErr != nil {
		log.Printf("GetModel failed, err:%s", objErr.Error())
		return
	}

	for _, val := range valModel.GetFields() {
		vName := val.GetName()
		if val.IsAssigned() {
			if vName != "AA" {
				t.Errorf("name:%s, check field assigned failed", vName)
			}
		}
		if !val.IsAssigned() {
			if vName != "ID" && vName != "Name" {
				t.Errorf("name:%s, check field assigned failed", vName)
			}
		}
	}
}
