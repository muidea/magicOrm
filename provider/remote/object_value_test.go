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
	obj := Simple{Name: "obj", Desc: &desc, Add: []int{12, 223, 456}}
	ext := &ExtInfo{Name: "extObj", Obj: obj, ObjArray: []*Simple{&obj, &obj}}

	objVal, objErr := GetObjectValue(ext)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	data, err := EncodeObjectValue(objVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}

	objInfo, objErr := DecodeObjectValue(data)
	if objErr != nil {
		t.Errorf("DecodeObjectValue failed, err:%s", objErr.Error())
		return
	}

	if !compareObjectValue(objVal, objInfo) {
		t.Errorf("compareObjectValue failed")
		return
	}
}

func TestUpdateExtObjValue(t *testing.T) {
	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc, Add: []int{12, 223, 456}}
	ext1 := &ExtInfo{}
	ext2 := &ExtInfo{Name: "extObj", Obj: obj, ObjArray: []*Simple{&obj, &obj}}

	objVal, objErr := GetObjectValue(ext2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	data, err := EncodeObjectValue(objVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}

	objInfo, objErr := DecodeObjectValue(data)
	if objErr != nil {
		t.Errorf("DecodeObjectValue failed, err:%s", objErr.Error())
		return
	}

	if !compareObjectValue(objVal, objInfo) {
		t.Errorf("compareObjectValue failed")
		return
	}

	err = UpdateEntity(objInfo, ext1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if ext1.Name != ext2.Name {
		t.Errorf("updateEntity failed")
	}
	if ext1.Obj.Name != ext2.Obj.Name {
		t.Errorf("updateEntity failed")
	}
	if len(ext1.ObjArray) != len(ext2.ObjArray) {
		t.Errorf("updateEntity failed")
	}
}
