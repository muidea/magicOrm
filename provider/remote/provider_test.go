package remote

import (
	"reflect"
	"testing"

	pu "github.com/muidea/magicOrm/provider/util"
)

func TestSimpleObjProvider(t *testing.T) {
	desc := "obj_desc"
	simple := Simple{Name: "simple", Desc: &desc, Age: 240, Flag: true, Add: []int{12, 34, 45}}

	simpleObj, simpleErr := GetObject(simple)
	if simpleErr != nil {
		t.Errorf("GetObject failed, err:%s", simpleErr.Error())
		return
	}

	data, dataErr := EncodeObject(simpleObj)
	if dataErr != nil {
		t.Errorf("marshal failed, err:%s", dataErr.Error())
		return
	}
	simpleInfo := &Object{}
	simpleInfo, simpleErr = DecodeObject(data)
	if simpleErr != nil {
		t.Errorf("unmarshal failed, err:%s", simpleErr.Error())
		return
	}

	simpleVal, simpleErr := GetObjectValue(simple)
	if simpleErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", simpleErr.Error())
		return
	}
	data, dataErr = EncodeObjectValue(simpleVal)
	if dataErr != nil {
		t.Errorf("marshal failed, err:%s", dataErr.Error())
		return
	}
	simpleVal, simpleErr = DecodeObjectValue(data)
	if simpleErr != nil {
		t.Errorf("unmarshal failed, err:%s", simpleErr.Error())
		return
	}

	simpleModel, simpleErr := GetEntityModel(simpleInfo)
	if simpleErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", simpleErr.Error())
		return
	}

	sVal := reflect.ValueOf(simpleVal)
	simpleModel, simpleErr = SetModelValue(simpleModel, pu.NewValue(sVal))
	if simpleErr != nil {
		t.Errorf("SetModelValue failed, err:%s", simpleErr.Error())
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

	bb := &BB{AA: &AA{}}
	bbVal, bbErr := GetObject(bb)
	if bbErr != nil {
		t.Errorf("GetObject failed, err:%s", bbErr.Error())
		return
	}

	_, objErr := GetEntityModel(bbVal)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}
}
