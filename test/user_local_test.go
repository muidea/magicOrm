package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalUser(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider("default", nil)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	status := &Status{Value: 10}
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	objList := []any{&Group{}, &User{}, &Status{}}
	registerLocalModel(localProvider, objList)

	statusModel, statusErr := localProvider.GetEntityModel(status, true)
	if statusErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", statusErr.Error())
		return
	}
	err = o1.Drop(statusModel)
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}

	err = o1.Create(statusModel)
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	statusModel, statusErr = o1.Insert(statusModel)
	if statusErr != nil {
		t.Errorf("insert status failed, err:%s", statusErr.Error())
		return
	}
	status = statusModel.Interface(true).(*Status)

	groupModel, groupErr := localProvider.GetEntityModel(group1, true)
	if groupErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", groupErr.Error())
		return
	}

	err = o1.Drop(groupModel)
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}

	err = o1.Create(groupModel)
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	groupModel, groupErr = o1.Insert(groupModel)
	if groupErr != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}
	group1 = groupModel.Interface(true).(*Group)

	group2Model, group2Err := localProvider.GetEntityModel(group2, true)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group2Err.Error())
		return
	}
	group2 = group2Model.Interface(true).(*Group)

	group3Model, group3Err := localProvider.GetEntityModel(group3, true)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group3Model, group3Err = o1.Insert(group3Model)
	if group3Err != nil {
		t.Errorf("insert Group3 failed, err:%s", group3Err.Error())
		return
	}
	group3 = group3Model.Interface(true).(*Group)

	user1 := User{Name: "demo", EMail: "123@demo.com", Status: status, Group: []*Group{}}
	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)

	userModel, userErr := localProvider.GetEntityModel(user1, true)
	if userErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", userErr.Error())
		return
	}
	err = o1.Drop(userModel)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userModel)
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	userModel, userErr = o1.Insert(userModel)
	if userErr != nil {
		t.Errorf("insert user1 failed, err:%s", userErr.Error())
		return
	}
	user1 = userModel.Interface(false).(User)

	user2 := User{ID: user1.ID, Status: &Status{}, Group: []*Group{}}
	user2Model, user2Err := localProvider.GetEntityModel(user2, true)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Query(user2Model)
	if user2Err != nil {
		t.Errorf("query user2 failed, err:%s", user2Err.Error())
		return
	}
	user2 = user2Model.Interface(false).(User)

	if !user2.Equal(&user1) {
		t.Errorf("query user2 failed")
		return
	}

	user1.Group = append(user1.Group, group3)
	user1Model, user1Err := localProvider.GetEntityModel(user1, true)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user1Model, user1Err = o1.Update(user1Model)
	if user1Err != nil {
		t.Errorf("update user1 failed, err:%s", user1Err.Error())
		return
	}
	newUser1 := user1Model.Interface(true).(*User)

	user2Model, user2Err = localProvider.GetEntityModel(user2, true)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Query(user2Model)
	if user2Err != nil {
		t.Errorf("query user2 failed, err:%s", user2Err.Error())
		return
	}
	newUser2 := user2Model.Interface(true).(*User)
	if len(newUser2.Group) != 3 {
		t.Errorf("query user2 failed")
		return
	}
	if !newUser2.Equal(newUser1) {
		t.Errorf("query user2 failed")
		return
	}

	group1Model, group1Err := localProvider.GetEntityModel(group1, true)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}
	_, group1Err = o1.Delete(group1Model)
	if group1Err != nil {
		t.Errorf("delete group1 failed, err:%s", group1Err.Error())
		return
	}

	group2Model, group2Err = localProvider.GetEntityModel(group2, true)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	_, group2Err = o1.Delete(group2Model)
	if group2Err != nil {
		t.Errorf("delete group1 failed, err:%s", group2Err.Error())
		return
	}

	group3Model, group3Err = localProvider.GetEntityModel(group3, true)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	_, group3Err = o1.Delete(group3Model)
	if group3Err != nil {
		t.Errorf("delete group1 failed, err:%s", group3Err.Error())
		return
	}

	user2Model, user2Err = localProvider.GetEntityModel(user2, true)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	_, user2Err = o1.Delete(user2Model)
	if user2Err != nil {
		t.Errorf("delete group1 failed, err:%s", user2Err.Error())
		return
	}
}
