package helper

import (
	"testing"

	"github.com/muidea/magicOrm/provider/remote"
)

func TestSimpleObjProvider(t *testing.T) {
	desc := "obj_desc"
	simple := Simple{Name: "simple", Desc: &desc, Age: 240, Flag: true, Add: []int{12, 34, 45}}

	simpleObj, simpleErr := GetObject(simple)
	if simpleErr != nil {
		t.Errorf("GetObject failed, err:%s", simpleErr.Error())
		return
	}

	data, dataErr := remote.EncodeObject(simpleObj)
	if dataErr != nil {
		t.Errorf("marshal failed, err:%s", dataErr.Error())
		return
	}
	simpleInfo := &remote.Object{}
	simpleInfo, simpleErr = remote.DecodeObject(data)
	if simpleErr != nil {
		t.Errorf("unmarshal failed, err:%s", simpleErr.Error())
		return
	}

	simpleVal, simpleErr := GetObjectValue(simple)
	if simpleErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", simpleErr.Error())
		return
	}
	data, dataErr = remote.EncodeObjectValue(simpleVal)
	if dataErr != nil {
		t.Errorf("marshal failed, err:%s", dataErr.Error())
		return
	}
	simpleVal, simpleErr = remote.DecodeObjectValue(data)
	if simpleErr != nil {
		t.Errorf("unmarshal failed, err:%s", simpleErr.Error())
		return
	}

	simpleModel, simpleErr := remote.GetEntityModel(simpleInfo)
	if simpleErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", simpleErr.Error())
		return
	}

	simpleModel, simpleErr = remote.SetModelValue(simpleModel, remote.NewValue(simpleVal))
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

	_, objErr := remote.GetEntityModel(bbVal)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}
}
