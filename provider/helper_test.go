package provider

import (
	"github.com/muidea/magicOrm/provider/remote"
	"testing"
)

func TestUpdateExtObjValue(t *testing.T) {
	type Simple struct {
		//ID 唯一标示单元
		ID   int64   `orm:"id key"`
		Name string  `orm:"name"`
		Desc *string `orm:"desc"`
		Age  uint8   `orm:"age"`
		Flag bool    `orm:"flag"`
		Add  []int   `orm:"add"`
	}

	type ExtInfo struct {
		ID       int64     `orm:"id key"`
		Name     string    `orm:"name"`
		Obj      Simple    `orm:"obj"`
		ObjPtr   *Simple   `orm:"objPtr"`
		ObjArray []*Simple `orm:"array"`
	}

	desc := "obj_desc"
	obj := Simple{Name: "obj", Desc: &desc, Add: []int{12, 223, 456}}
	ext1 := &ExtInfo{ObjArray: []*Simple{}}
	ext2 := &ExtInfo{Name: "extObj", Obj: obj, ObjArray: []*Simple{&obj, &obj}}

	objVal, objErr := remote.GetObjectValue(ext2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	data, err := remote.EncodeObjectValue(objVal)
	if err != nil {
		t.Errorf("encode object value failed, err:%s", err.Error())
		return
	}

	objInfo, objErr := remote.DecodeObjectValue(data)
	if objErr != nil {
		t.Errorf("DecodeObjectValue failed, err:%s", objErr.Error())
		return
	}

	if !remote.CompareObjectValue(objVal, objInfo) {
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
