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

	provider := New()

	objInfo, objErr := GetObject(obj, provider.modelCache)
	if objErr != nil {
		log.Printf("GetObject failed, err:%s", objErr.Error())
		return
	}

	data, dataErr := json.Marshal(objInfo)
	if dataErr != nil {
		log.Printf("marshal failed, err:%s", dataErr.Error())
		return
	}
	log.Print(string(data))
	localInfo := &Object{}
	err := json.Unmarshal(data, localInfo)
	if err != nil {
		log.Printf("unmarshal failed, err:%s", err.Error())
		return
	}

	objVal, objErr := GetObjectValue(obj)
	if objErr != nil {
		log.Printf("GetObjectModel failed, err:%s", objErr.Error())
		return
	}
	data, dataErr = json.Marshal(objVal)
	if dataErr != nil {
		log.Printf("marshal failed, err:%s", dataErr.Error())
		return
	}
	log.Print(string(data))
	localVal := &ObjectValue{}
	err = json.Unmarshal(data, localVal)
	if err != nil {
		log.Printf("unmarshal failed, err:%s", err.Error())
		return
	}

	infoModel, objErr := provider.GetObjectModel(localInfo)
	if objErr != nil {
		log.Printf("GetObjectModel failed, err:%s", objErr.Error())
		return
	}
	log.Printf("GetObjectModel name:%s,pkgPath:%s", infoModel.GetName(), infoModel.GetPkgPath())

	valModel, objErr := provider.GetValueModel(reflect.ValueOf(localVal))
	if objErr != nil {
		log.Printf("GetValueModel failed, err:%s", objErr.Error())
		return
	}
	log.Printf("GetValueModel name:%s,pkgPath:%s", valModel.GetName(), valModel.GetPkgPath())

	if infoModel.GetName() != valModel.GetName() {
		log.Printf("get value model failed. infoModel name:%s, valModel name:%s", infoModel.GetName(), valModel.GetName())
		return
	}

}
