package remote

import (
	"testing"
)

func TestSimpleValue(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc, Age: 240, Add: []int{12, 34, 45}}

	rawVal, rawErr := GetObjectValue(obj)
	if rawErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", rawErr.Error())
		return
	}

	if !rawVal.IsAssigned() {
		t.Errorf("check object is assigned failed")
		return
	}

	data, err := EncodeObjectValue(rawVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}

	curVal, curErr := DecodeObjectValue(data)
	if curErr != nil {
		t.Errorf("decode obj failed, err:%s", curErr.Error())
		return
	}

	if !compareObjectValue(rawVal, curVal) {
		t.Errorf("CompareObjectValue failed")
		return
	}

	obj2 := Simple{}
	rawVal, rawErr = GetObjectValue(obj2)
	if rawErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", rawErr.Error())
		return
	}

	if rawVal.IsAssigned() {
		t.Errorf("check object is assigned failed")
		return
	}
}

func TestExtObjValue(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc}
	ext := &ExtInfo{Name: "extObj", Obj: obj}

	objVal, objErr := GetObjectValue(ext)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	_, err := EncodeObjectValue(objVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}
}
