package provider

import (
	"testing"

	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestUpdateExtObjValue(t *testing.T) {
	newVal := &Compose{
		//BasePtrArrayPtr: &[]*Base{},
	}
	rawVal := composeVal
	objVal, objErr := helper.GetObjectValue(rawVal)
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

	err = UpdateLocalEntity(objInfo, newVal)
	if err != nil {
		t.Errorf("UpdateLocalEntity failed, err:%s", err.Error())
		return
	}

	if newVal.Name != rawVal.Name {
		t.Errorf("updateEntity failed")
		return
	}
	if newVal.Base.ID != rawVal.Base.ID {
		t.Errorf("updateEntity failed")
		return
	}
	if newVal.BasePtr == nil {
		t.Errorf("updateEntity failed")
		return
	}
	if len(newVal.BasePtrArray) != len(rawVal.BasePtrArray) {
		t.Errorf("updateEntity failed")
		return
	}
	if len(*newVal.BasePtrArrayPtr) != len(*rawVal.BasePtrArrayPtr) {
		t.Errorf("updateEntity failed")
		return
	}
}
