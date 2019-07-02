package remote

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

func TestSimpleObjProvider(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc, Age: 240, Add: []int{12, 34, 45}}

	provider := New("default")

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

	err = provider.RegisterModel(localInfo)
	if err != nil {
		log.Printf("RegisterModel failed, err:%s", err.Error())
		return
	}

	infoModel, objErr := provider.GetEntityModel(localInfo)
	if objErr != nil {
		log.Printf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	valModel, objErr := provider.GetValueModel(reflect.ValueOf(localVal))
	if objErr != nil {
		log.Printf("GetValueModel failed, err:%s", objErr.Error())
		return
	}

	if infoModel.GetName() != valModel.GetName() {
		log.Printf("get value model failed. infoModel name:%s, valModel name:%s", infoModel.GetName(), valModel.GetName())
		return
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

	aaDef, aaErr := GetObject(&AA{})
	if aaErr != nil {
		t.Errorf("GetObject failed, err:%s", aaErr.Error())
		return
	}

	bbDef, bbErr := GetObject(&BB{})
	if bbErr != nil {
		t.Errorf("GetObject failed, err:%s", bbErr.Error())
		return
	}

	provider := New("default")

	err := provider.RegisterModel(aaDef)
	if err != nil {
		t.Errorf("RegisterModel failed, err:%s", err.Error())
		return
	}
	err = provider.RegisterModel(bbDef)
	if err != nil {
		t.Errorf("RegisterModel failed, err:%s", err.Error())
		return
	}

	bb := &BB{AA: &AA{}}
	bbVal, bbErr := GetObjectValue(bb)
	if bbErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", bbErr.Error())
		return
	}

	_, bbErr = provider.GetEntityModel(bbVal)
	if bbErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", bbErr.Error())
		return
	}
}
