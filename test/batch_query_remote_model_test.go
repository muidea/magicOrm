package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteBatchQuery(t *testing.T) {
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
	user0 := &User{}
	userDef, userObjectErr := helper.GetObject(user0)
	if userObjectErr != nil {
		t.Errorf("GetObject failed, err:%s", userObjectErr.Error())
		return
	}

	group0 := &Group{}
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	objList := []any{group0, user0, status}
	mList, mErr := registerRemoteModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registeregisterRemoteModelrLocalModel failed, err:%s", mErr.Error())
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

	group1Val, userObjectErr := getObjectValue(group1)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group1Model, group1Err := remoteProvider.GetEntityModel(group1Val, true)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}
	group1Model, group1Err = o1.Insert(group1Model)
	if group1Err != nil {
		t.Errorf("insert group failed, err:%s", group1Err.Error())
		return
	}
	err = helper.UpdateEntity(group1Model.Interface(true).(*remote.ObjectValue), group1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group2Val, userObjectErr := getObjectValue(group2)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group2Model, group2Err := remoteProvider.GetEntityModel(group2Val, true)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert group failed, err:%s", group2Err.Error())
		return
	}
	err = helper.UpdateEntity(group2Model.Interface(true).(*remote.ObjectValue), group2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	group3Val, userObjectErr := getObjectValue(group3)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group3Model, group3Err := remoteProvider.GetEntityModel(group3Val, true)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group3Model, group3Err = o1.Insert(group3Model)
	if group3Err != nil {
		t.Errorf("insert group failed, err:%s", group3Err.Error())
		return
	}
	err = helper.UpdateEntity(group3Model.Interface(true).(*remote.ObjectValue), group3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)

	err = o1.Drop(userDef)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef)
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	user1Val, userObjectErr := getObjectValue(user1)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	user1Model, user1Err := remoteProvider.GetEntityModel(user1Val, true)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user1Model, user1Err = o1.Insert(user1Model)
	if user1Err != nil {
		t.Errorf("insert user failed, err:%s", user1Err.Error())
		return
	}
	err = helper.UpdateEntity(user1Model.Interface(true).(*remote.ObjectValue), user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user2.Group = append(user2.Group, group1)
	user2.Group = append(user2.Group, group3)
	user2Val, userObjectErr := getObjectValue(user2)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	user2Model, user2Err := remoteProvider.GetEntityModel(user2Val, true)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Insert(user2Model)
	if user2Err != nil {
		t.Errorf("insert user failed, err:%s", user2Err.Error())
		return
	}
	err = helper.UpdateEntity(user2Model.Interface(true).(*remote.ObjectValue), user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	maskVal, maskErr := helper.GetObjectValue(&User{Group: []*Group{}})
	if maskErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", maskErr.Error())
		return
	}

	groupList := []*Group{group1, group2}
	groupListVal, groupListErr := helper.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	userObject, userObjectErr := helper.GetObject(&User{})
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}

	filter, err := remoteProvider.GetModelFilter(userObject)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	userModelList, userModelErr := o1.BatchQuery(filter)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	if len(userModelList) != 2 {
		t.Errorf("batch query user failed")
		return
	}

	err = filter.Equal("name", user1.Name)
	if err != nil {
		t.Errorf("filter.Equal err:%s", err.Error())
		return
	}

	err = filter.In("group", groupListVal)
	if err != nil {
		t.Errorf("filter.In err:%s", err.Error())
		return
	}
	err = filter.Like("email", user1.EMail)
	if err != nil {
		t.Errorf("filter.Like err:%s", err.Error())
		return
	}
	err = filter.ValueMask(maskVal)
	if err != nil {
		t.Errorf("filter.ValueMask err:%s", err.Error())
		return
	}

	filter.Pagination(0, 100)
	userModelList, userModelErr = o1.BatchQuery(filter)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	if len(userModelList) != 1 {
		t.Errorf("filter query user failed")
		return
	}

	groupList = []*Group{group1}
	groupListVal, groupListErr = helper.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	userFilter, err := remoteProvider.GetModelFilter(userObject)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	err = userFilter.In("group", groupListVal)
	if err != nil {
		t.Errorf("userFilter.In failed, err:%s", err.Error())
		return
	}

	userModelList, userModelErr = o1.BatchQuery(userFilter)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	if len(userModelList) != 2 {
		t.Errorf("filter query user failed")
		return
	}
}

func TestRemoteBatchQueryPtr(t *testing.T) {
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
	statusDef, userObjectErr := helper.GetObject(status)
	if userObjectErr != nil {
		t.Errorf("GetObject failed, err:%s", userObjectErr.Error())
		return
	}

	user0 := &User{}
	userDef, userObjectErr := helper.GetObject(user0)
	if userObjectErr != nil {
		t.Errorf("GetObject failed, err:%s", userObjectErr.Error())
		return
	}

	group0 := &Group{}
	groupDef, userObjectErr := helper.GetObject(group0)
	if userObjectErr != nil {
		t.Errorf("GetObject failed, err:%s", userObjectErr.Error())
		return
	}

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	objList := []any{groupDef, userDef, statusDef}
	_, err = registerLocalModel(remoteProvider, objList)
	if err != nil {
		t.Errorf("registerLocalModel failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(statusDef)
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}

	err = o1.Create(statusDef)
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(groupDef)
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}
	err = o1.Create(groupDef)
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	statusVal, userObjectErr := getObjectValue(status)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	statusModel, statusErr := remoteProvider.GetEntityModel(statusVal, true)
	if statusErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", statusErr.Error())
		return
	}

	statusModel, statusErr = o1.Insert(statusModel)
	if statusErr != nil {
		t.Errorf("insert group failed, err:%s", statusErr.Error())
		return
	}
	err = helper.UpdateEntity(statusModel.Interface(true).(*remote.ObjectValue), status)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	user1.Status = status

	group1Val, userObjectErr := getObjectValue(group1)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group1Model, group1Err := remoteProvider.GetEntityModel(group1Val, true)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}

	group1Model, group1Err = o1.Insert(group1Model)
	if group1Err != nil {
		t.Errorf("insert group failed, err:%s", group1Err.Error())
		return
	}
	err = helper.UpdateEntity(group1Model.Interface(true).(*remote.ObjectValue), group1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group2Val, userObjectErr := getObjectValue(group2)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group2Model, group2Err := remoteProvider.GetEntityModel(group2Val, true)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}

	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert group failed, err:%s", group2Err.Error())
		return
	}
	err = helper.UpdateEntity(group2Model.Interface(true).(*remote.ObjectValue), group2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	group3Val, userObjectErr := getObjectValue(group3)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group3Model, group3Err := remoteProvider.GetEntityModel(group3Val, true)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}

	group3Model, group3Err = o1.Insert(group3Model)
	if group3Err != nil {
		t.Errorf("insert group failed, err:%s", group3Err.Error())
		return
	}
	err = helper.UpdateEntity(group3Model.Interface(true).(*remote.ObjectValue), group3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)

	err = o1.Drop(userDef)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef)
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	user1Val, userObjectErr := getObjectValue(user1)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	user1Model, user1Err := remoteProvider.GetEntityModel(user1Val, true)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}

	user1Model, user1Err = o1.Insert(user1Model)
	if user1Err != nil {
		t.Errorf("insert group failed, err:%s", user1Err.Error())
		return
	}
	err = helper.UpdateEntity(user1Model.Interface(true).(*remote.ObjectValue), user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user2.Status = status
	user2.Group = append(user2.Group, group1)
	user2.Group = append(user2.Group, group3)
	user2Val, userObjectErr := getObjectValue(user2)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	user2Model, user2Err := remoteProvider.GetEntityModel(user2Val, true)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}

	user2Model, user2Err = o1.Insert(user2Model)
	if user2Err != nil {
		t.Errorf("insert group failed, err:%s", user2Err.Error())
		return
	}
	err = helper.UpdateEntity(user2Model.Interface(true).(*remote.ObjectValue), user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	maskValue := &User{Status: &Status{}}

	maskVal, _ := getObjectValue(maskValue)
	groupList := []*Group{group1, group2}
	groupListVal, groupListErr := helper.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	userObject, userObjectErr := helper.GetObject(&User{})
	if userObjectErr != nil {
		t.Errorf("GetObject failed, err:%s", userObjectErr.Error())
		return
	}
	filter, err := remoteProvider.GetModelFilter(userObject)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	filter.Pagination(0, 100)
	userModelList, userModelErr := o1.BatchQuery(filter)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	if len(userModelList) != 2 {
		t.Errorf("batch query user failed")
		return
	}

	err = filter.Equal("name", user1.Name)
	if err != nil {
		t.Errorf("filter.Equal failed, err:%s", err.Error())
		return
	}
	err = filter.In("group", groupListVal)
	if err != nil {
		t.Errorf("filter.In failed, err:%s", err.Error())
		return
	}
	err = filter.Like("email", user1.EMail)
	if err != nil {
		t.Errorf("filter.Like failed, err:%s", err.Error())
		return
	}
	err = filter.ValueMask(maskVal)
	if err != nil {
		t.Errorf("filter.ValueMask failed, err:%s", err.Error())
		return
	}

	userModelList, userModelErr = o1.BatchQuery(filter)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	if len(userModelList) != 1 {
		t.Errorf("filter query user failed")
		return
	}

	groupList = []*Group{group1}
	groupListVal, groupListErr = helper.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	filter2, err := remoteProvider.GetModelFilter(userObject)
	if err != nil {
		t.Errorf("GetModelFilter failed, err:%s", err.Error())
		return
	}

	err = filter2.In("group", groupListVal)
	if err != nil {
		t.Errorf("filter.In failed, err:%s", err.Error())
		return
	}

	userModelList, userModelErr = o1.BatchQuery(filter2)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	if len(userModelList) != 2 {
		t.Errorf("filter query user failed")
		return
	}
}
