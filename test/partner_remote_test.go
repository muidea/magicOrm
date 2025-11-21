package test

import (
	"encoding/json"
	"testing"

	"github.com/muidea/magicOrm/models"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/remote"
)

var statusDef = `
{
  "id": 0,
  "name": "status",
  "version": "0.0.1",
  "status": 2,
  "pkgPath": "/vmi",
  "description": "",
  "fields": [
    {
      "name": "id",
      "description": "",
      "type": {
        "name": "int64",
        "pkgPath": "",
        "description": "",
        "value": 105,
        "isPtr": false,
        "elemType": null
      },
      "spec": {
        "primaryKey": true,
        "valueDeclare": "auto",
        "viewDeclare": [
          "detail",
          "lite"
        ]
      }
    },
    {
      "name": "value",
      "description": "",
      "type": {
        "name": "int",
        "pkgPath": "",
        "description": "",
        "value": 104,
        "isPtr": false,
        "elemType": null
      },
      "spec": {
        "viewDeclare": [
          "detail",
          "lite"
        ]
      }
    },
    {
      "name": "name",
      "description": "",
      "type": {
        "name": "string",
        "pkgPath": "",
        "description": "",
        "value": 113,
        "isPtr": false,
        "elemType": null
      },
      "spec": {
        "viewDeclare": [
          "detail",
          "lite"
        ]
      }
    }
  ]
}
`

var partnerDef = `
{
  "id": 0,
  "name": "partner",
  "version": "0.0.1",
  "status": 2,
  "pkgPath": "/vmi",
  "description": "会员信息",
  "fields": [
    {
      "name": "id",
      "description": "",
      "type": {
        "name": "int64",
        "pkgPath": "",
        "description": "",
        "value": 105,
        "isPtr": false,
        "elemType": null
      },
      "spec": {
        "primaryKey": true,
        "valueDeclare": "auto",
        "viewDeclare": [
          "detail",
          "lite"
         ]
      }
    },
    {
      "name": "code",
      "description": "",
      "type": {
        "name": "string",
        "pkgPath": "",
        "description": "",
        "value": 113,
        "isPtr": false,
        "elemType": null
      },
      "spec": {
        "viewDeclare": [
          "detail",
          "lite"
         ]
      }
    },
    {
      "name": "status",
      "description": "",
      "type": {
        "name": "status",
        "pkgPath": "/vmi",
        "description": "",
        "value": 115,
        "isPtr": true,
        "elemType": null
      },
      "spec": {
        "viewDeclare": [
          "detail",
          "lite"
         ],
        "defaultValue": 2
      }
    }
  ]
}
`

func TestPartner(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider("default")
	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	statusObject := &remote.Object{}
	jsonErr := json.Unmarshal([]byte(statusDef), statusObject)
	if jsonErr != nil {
		t.Errorf("unmarshal status object failed, error:%s", jsonErr.Error())
		return
	}

	partnerObject := &remote.Object{}
	jsonErr = json.Unmarshal([]byte(partnerDef), partnerObject)
	if jsonErr != nil {
		t.Errorf("unmarshal partner object failed, error:%s", jsonErr.Error())
		return
	}

	objList := []any{statusObject, partnerObject}
	mList, mErr := registerLocalModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("register mode failed, err:%s", mErr.Error())
		return
	}

	err = dropModel(o1, mList)
	if err != nil {
		t.Errorf("drop model failed, err:%s", err.Error())
		return
	}
	err = createModel(o1, mList)
	if err != nil {
		t.Errorf("create model failed, err:%s", err.Error())
		return
	}

	status001ObjectValue := &remote.ObjectValue{
		Name:    statusObject.Name,
		PkgPath: statusObject.PkgPath,
		Fields: []*remote.FieldValue{
			{
				Name:  "id",
				Value: 1,
			},
			{
				Name:  "value",
				Value: 110,
			},
			{
				Name:  "name",
				Value: "t110",
			},
		},
	}
	status001Model, status001Err := remoteProvider.GetEntityModel(status001ObjectValue)
	if status001Err != nil {
		t.Errorf("remoteProvider.GetEntityModel failed, error:%s", status001Err.Error())
		return
	}

	_, status001Err = o1.Insert(status001Model)
	if status001Err != nil {
		t.Errorf("Insert status failed, error:%s", status001Err.Error())
		return
	}

	status002ObjectValue := &remote.ObjectValue{
		Name:    statusObject.Name,
		PkgPath: statusObject.PkgPath,
		Fields: []*remote.FieldValue{
			{
				Name:  "id",
				Value: 2,
			},
			{
				Name:  "value",
				Value: 220,
			},
			{
				Name:  "name",
				Value: "t220",
			},
		},
	}
	status002Model, status002Err := remoteProvider.GetEntityModel(status002ObjectValue)
	if status002Err != nil {
		t.Errorf("remoteProvider.GetEntityModel failed, error:%s", status002Err.Error())
		return
	}

	_, status002Err = o1.Insert(status002Model)
	if status002Err != nil {
		t.Errorf("Insert status failed, error:%s", status002Err.Error())
		return
	}

	partnerObjectValue := &remote.ObjectValue{
		Name:    partnerObject.Name,
		PkgPath: partnerObject.PkgPath,
		Fields: []*remote.FieldValue{
			{
				Name:  "code",
				Value: "10001",
			},
		},
	}
	partnerModel, partnerErr := remoteProvider.GetEntityModel(partnerObjectValue)
	if partnerErr != nil {
		t.Errorf("remoteProvider.GetEntityModel failed, error:%s", partnerErr.Error())
		return
	}

	partnerModel, partnerErr = o1.Insert(partnerModel)
	if partnerErr != nil {
		t.Errorf("Insert partner failed, error:%s", partnerErr.Error())
		return
	}

	partnerModel, partnerErr = o1.Query(partnerModel)
	if partnerErr != nil {
		t.Errorf("Insert partner failed, error:%s", partnerErr.Error())
		return
	}

	partnerFilter, partnerErr := remoteProvider.GetEntityFilter(partnerModel.Interface(true), models.MetaView)
	if partnerErr != nil {
		t.Errorf("remoteProvider.GetEntityFilter failed, error:%s", partnerErr.Error())
		return
	}

	partnerList, partnerErr := o1.BatchQuery(partnerFilter)
	if partnerErr != nil {
		t.Errorf("BatchQuery partner failed, error:%s", partnerErr.Error())
		return
	}
	if len(partnerList) != 1 {
		t.Errorf("BatchQuery partner failed, error:%s", partnerErr.Error())
		return
	}
}
