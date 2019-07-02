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
