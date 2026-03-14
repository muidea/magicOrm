package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteGroup(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider("default", nil)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	status := &Status{Value: 10}
	_, objErr := helper.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	objList := []any{&Group{}, &User{}, status}
	mList, mErr := registerRemoteModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registerRemoteModel failed, err:%s", mErr.Error())
		return
	}

	err = dropModel(o1, mList)
	if err != nil {
		t.Errorf("dropModel failed, err:%s", err.Error())
		return
	}

	err = createModel(o1, mList)
	if err != nil {
		t.Errorf("createModel failed, err:%s", err.Error())
		return
	}

	statusVal, objErr := getObjectValue(status)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	statusModel, statusErr := remoteProvider.GetEntityModel(statusVal, true)
	if statusErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", statusErr.Error())
		return
	}

	statusModel, statusErr = o1.Insert(statusModel)
	if statusErr != nil {
		t.Errorf("insert Group1 failed, err:%s", statusErr.Error())
		return
	}

	err = helper.UpdateEntity(statusModel.Interface(true).(*remote.ObjectValue), status)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	group1Model, group1Err := remoteProvider.GetEntityModel(group1Val, true)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}
	group1Model, group1Err = o1.Insert(group1Model)
	if group1Err != nil {
		t.Errorf("insert Group1 failed, err:%s", group1Err.Error())
		return
	}

	err = helper.UpdateEntity(group1Model.Interface(true).(*remote.ObjectValue), group1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	qGroup1 := &Group{ID: group1.ID, Parent: &Group{}}
	qGroup1Val, qObjErr := getObjectValue(qGroup1)
	if qObjErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", qObjErr.Error())
		return
	}
	qGroup1Model, qGroup1Err := remoteProvider.GetEntityModel(qGroup1Val, true)
	if qGroup1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", qGroup1Err.Error())
		return
	}
	qGroup1Model, qGroup1Err = o1.Query(qGroup1Model)
	if qGroup1Err != nil {
		t.Errorf("insert Group1 failed, err:%s", qGroup1Err.Error())
		return
	}
	err = helper.UpdateEntity(qGroup1Model.Interface(true).(*remote.ObjectValue), qGroup1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group2.Parent = group1
	group2Val, objErr := getObjectValue(group2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group2Model, group2Err := remoteProvider.GetEntityModel(group2Val, true)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group2Err.Error())
		return
	}
	err = helper.UpdateEntity(group2Model.Interface(true).(*remote.ObjectValue), group2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group3.Parent = group1
	group3Val, objErr := getObjectValue(group3)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group3Model, group3Err := remoteProvider.GetEntityModel(group3Val, true)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group3Model, group3Err = o1.Insert(group3Model)
	if group3Err != nil {
		t.Errorf("insert Group3 failed, err:%s", group3Err.Error())
		return
	}

	err = helper.UpdateEntity(group3Model.Interface(true).(*remote.ObjectValue), group3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	_, group3Err = o1.Delete(group3Model)
	if group3Err != nil {
		t.Errorf("delete Group3 failed, err:%s", group3Err.Error())
		return
	}

	group4 := &Group{ID: group2.ID, Parent: &Group{}}
	group4Val, objErr := getObjectValue(group4)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group4Model, group4Err := remoteProvider.GetEntityModel(group4Val, true)
	if group4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group4Err.Error())
		return
	}
	group4Model, group4Err = o1.Query(group4Model)
	if group4Err != nil {
		t.Errorf("query Group4 failed, err:%s", group4Err.Error())
		return
	}

	err = helper.UpdateEntity(group4Model.Interface(true).(*remote.ObjectValue), group4)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group5 := &Group{ID: group2.ID, Parent: &Group{}}
	group5Val, objErr := getObjectValue(group5)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group5Model, group5Err := remoteProvider.GetEntityModel(group5Val, true)
	if group5Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group5Err.Error())
		return
	}
	group5Model, group5Err = o1.Query(group5Model)
	if group5Err != nil {
		t.Errorf("query Group5 failed, err:%s", group5Err.Error())
		return
	}

	err = helper.UpdateEntity(group5Model.Interface(true).(*remote.ObjectValue), group5)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	if !group5.Equal(group2) {
		t.Errorf("query Group5 failed")
	}
}
