package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteUser(t *testing.T) {
	status := &Status{Value: 10}
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	user0 := &User{}
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider("default", nil)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{group1, user0, status}
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
		t.Errorf("insert Group2 failed, err:%s", group3Err.Error())
		return
	}
	err = helper.UpdateEntity(group3Model.Interface(true).(*remote.ObjectValue), group3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user1 := &User{Name: "demo", EMail: "123@demo.com", Group: []*Group{}}
	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)
	user1.Status = status
	user1Val, objErr := getObjectValue(user1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	user1Model, user1Err := remoteProvider.GetEntityModel(user1Val, true)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user1Model, user1Err = o1.Insert(user1Model)
	if user1Err != nil {
		t.Errorf("insert user1 failed, err:%s", user1Err.Error())
		return
	}
	err = helper.UpdateEntity(user1Model.Interface(true).(*remote.ObjectValue), user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user2 := &User{ID: user1.ID, Status: &Status{}, Group: []*Group{}}
	user2Val, objErr := getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	user2Model, user2Err := remoteProvider.GetEntityModel(user2Val, true)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Query(user2Model)
	if user2Err != nil {
		t.Errorf("query user2 failed, err:%s", user2Err.Error())
		return
	}
	err = helper.UpdateEntity(user2Model.Interface(true).(*remote.ObjectValue), user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if !user2.Equal(user1) {
		t.Errorf("query user2 failed")
		return
	}

	user1.Group = append(user1.Group, group3)
	user1Val, objErr = getObjectValue(user1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	user1Model, user1Err = remoteProvider.GetEntityModel(user1Val, true)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user1Model, user1Err = o1.Update(user1Model)
	if user1Err != nil {
		t.Errorf("update user1 failed, err:%s", user1Err.Error())
		return
	}
	err = helper.UpdateEntity(user1Model.Interface(true).(*remote.ObjectValue), user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	user2Val, objErr = getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	user2Model, user2Err = remoteProvider.GetEntityModel(user2Val, true)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Query(user2Model)
	if user2Err != nil {
		t.Errorf("query user2 failed, err:%s", user2Err.Error())
		return
	}
	err = helper.UpdateEntity(user2Model.Interface(true).(*remote.ObjectValue), user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if len(user2.Group) != 3 {
		t.Errorf("query user2 failed")
		return
	}
	if !user2.Equal(user1) {
		t.Errorf("query user2 failed")
		return
	}

	userObject, userErr := helper.GetObject(&User{Status: &Status{}, Group: []*Group{}})
	if userErr != nil {
		t.Errorf("helper.GetObject failed, error:%s", userErr.Error())
		return
	}

	userModel, _ := remoteProvider.GetEntityModel(userObject, true)
	user2Filter, user2Err := remoteProvider.GetModelFilter(userModel)
	if user2Err != nil {
		t.Errorf("remoteProvider.GetModelFilter failed, error:%s", user2Err.Error())
		return
	}

	maskValue, maskErr := helper.GetObjectValue(&User{Status: &Status{}, Group: []*Group{}})
	if maskErr != nil {
		t.Errorf("helper.GetObjectValue failed, error:%s", maskErr.Error())
		return
	}

	maskErr = user2Filter.ValueMask(maskValue)
	if maskErr != nil {
		t.Errorf("user2Filter.ValueMask failed, error:%s", maskErr.Error())
		return
	}

	userModelList, userModelErr := o1.BatchQuery(user2Filter)
	if userModelErr != nil {
		t.Errorf("o1.BatchQuery failed, error:%s", userModelErr.Error())
		return
	}
	if len(userModelList) != 1 {
		t.Errorf("o1.BatchQuery failed")
		return
	}

	userValueList := []*remote.ObjectValue{}
	for _, val := range userModelList {
		userValueList = append(userValueList, val.Interface(true).(*remote.ObjectValue))
	}

	userList := []*User{}
	userSliceValue := remote.TransferObjectValue(maskValue.GetName(), maskValue.GetPkgPath(), userValueList)
	user2Err = helper.UpdateSliceEntity(userSliceValue, &userList)
	if user2Err != nil {
		t.Errorf("helper.UpdateSliceEntity failed, error:%s", user2Err.Error())
		return
	}

	_, group1Err = o1.Delete(group1Model)
	if group1Err != nil {
		t.Errorf("delete group1 failed, err:%s", group1Err.Error())
		return
	}
	_, group2Err = o1.Delete(group2Model)
	if group2Err != nil {
		t.Errorf("delete group2 failed, err:%s", group2Err.Error())
		return
	}
	_, group3Err = o1.Delete(group3Model)
	if group3Err != nil {
		t.Errorf("delete group3 failed, err:%s", group3Err.Error())
		return
	}
	_, user2Err = o1.Delete(user2Model)
	if user2Err != nil {
		t.Errorf("delete user2 failed, err:%s", user2Err.Error())
		return
	}
}
