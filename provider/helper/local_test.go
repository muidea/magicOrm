package helper

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/muidea/magicOrm/provider/remote"
)

/*
	{
	    "uuid": "5b915b5f2c8349c4ae543a25eb04d79a",
	    "name": "magicTest",
	    "shortName": "test",
	    "icon": "/static/file/share/test/icon.svg",
	    "version": "v1.3.0",
	    "domain": "mulife.vip",
	    "email": "rangh@mulife.vip",
	    "author": "rangh",
	    "description": "test application"
	}
*/

const appVal = `
	{
	    "uuid": "5b915b5f2c8349c4ae543a25eb04d79a",
	    "name": "magicTest",
	    "shortName": "test",
	    "icon": "/static/file/share/test/icon.svg",
	    "version": "v1.3.0",
	    "domain": "mulife.vip",
	    "email": "rangh@mulife.vip",
	    "author": "rangh",
	    "description": "test application"
	}
`

type ApplicationDefine struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Icon        string `json:"icon"`
	Version     string `json:"version"`
	Domain      string `json:"domain"`
	EMail       string `json:"email"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

func TestApplication(t *testing.T) {
	valPtr, valErr := GetObjectValue(&ApplicationDefine{})
	assert.Nil(t, valErr)
	assert.NotNil(t, valPtr)

	appPtr := &ApplicationDefine{}
	err := json.Unmarshal([]byte(appVal), appPtr)
	assert.Nil(t, err)

	valPtr, valErr = GetObjectValue(appPtr)
	assert.Nil(t, valErr)
	assert.NotNil(t, valPtr)
}

func TestUpdateExtObjValue(t *testing.T) {
	newVal := &Compose{
		BasePtrArrayPtr: &[]*Base{},
	}
	rawVal := composeVal
	objVal, objErr := remote.GetObjectValue(rawVal)
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

	err = UpdateEntity(objInfo, newVal)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if newVal.Name != rawVal.Name {
		t.Errorf("updateEntity failed, newVal.Name:%s, rawVal.Name:%s", newVal.Name, rawVal.Name)
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

	sliceObjectValue := &remote.SliceObjectValue{
		Name:    objVal.Name,
		PkgPath: objVal.PkgPath,
		Values:  []*remote.ObjectValue{objVal, objInfo},
	}

	composeList := []*Compose{}
	err = UpdateSliceEntity(sliceObjectValue, &composeList)
	if err != nil {
		t.Errorf("UpdateSliceEntity failed, err:%s", err.Error())
		return
	}
	if len(composeList) != len(sliceObjectValue.Values) {
		t.Errorf("UpdateSliceEntity failed")
	}
}
