package remote

import (
	"encoding/json"
	"testing"
)

func TestSimpleObjValue(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc, Age: 240, Add: []int{12, 34, 45}}

	objVal, objErr := GetObjectValue(obj)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	if !objVal.IsAssigned() {
		t.Errorf("check object is assigned failed")
		return
	}

	data, err := json.Marshal(objVal)
	if err != nil {
		t.Errorf("marshal obj failed, err:%s", err.Error())
		return
	}

	val := &ObjectValue{}
	err = json.Unmarshal(data, val)
	if err != nil {
		t.Errorf("marshal obj failed, err:%s", err.Error())
		return
	}

	obj2 := SimpleObj{}
	objVal, objErr = GetObjectValue(obj2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	if objVal.IsAssigned() {
		t.Errorf("check object is assigned failed")
		return
	}
}

func TestExtObjValue(t *testing.T) {
	desc := "obj_desc"
	obj := SimpleObj{Name: "obj", Desc: &desc}
	ext := &ExtObj{Name: "extObj", Obj: obj}

	objVal, objErr := GetObjectValue(ext)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	_, err := json.Marshal(&objVal)
	if err != nil {
		t.Errorf("marshal obj failed, err:%s", err.Error())
		return
	}
}
