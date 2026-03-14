package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteSystem(t *testing.T) {
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
	group0 := &Group{}
	sys0 := &System{}
	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	objList := []any{group0, user0, status, sys0}
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
		t.Errorf("insert user failed, err:%s", user1Err.Error())
		return
	}
	err = helper.UpdateEntity(user1Model.Interface(true).(*remote.ObjectValue), user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

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

	users := []User{*user1, *user2}
	sys1 := &System{Name: "sys1", Tags: []string{"aab", "ccd"}}
	sys1.Users = &users

	sys1Val, objErr := getObjectValue(sys1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	sys1Model, sys1Err := remoteProvider.GetEntityModel(sys1Val, true)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys1Model, sys1Err = o1.Insert(sys1Model)
	if sys1Err != nil {
		t.Errorf("insert user failed, err:%s", sys1Err.Error())
		return
	}
	err = helper.UpdateEntity(sys1Model.Interface(true).(*remote.ObjectValue), sys1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	users = append(users, *user1)
	users = append(users, *user2)
	sys1.Users = &users
	sys1Val, objErr = getObjectValue(sys1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	sys1Model, sys1Err = remoteProvider.GetEntityModel(sys1Val, true)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys1Model, sys1Err = o1.Update(sys1Model)
	if sys1Err != nil {
		t.Errorf("update system failed, err:%s", sys1Err.Error())
		return
	}
	err = helper.UpdateEntity(sys1Model.Interface(true).(*remote.ObjectValue), sys1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	sys2 := &System{ID: sys1.ID, Users: &[]User{}, Tags: []string{}}
	sys2Val, objErr := getObjectValue(sys2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	sys2Model, sys2Err := remoteProvider.GetEntityModel(sys2Val, true)
	if sys2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys2Err.Error())
		return
	}
	sys2Model, sys2Err = o1.Query(sys2Model)
	if sys2Err != nil {
		t.Errorf("query system failed, err:%s", sys2Err.Error())
		return
	}
	err = helper.UpdateEntity(sys2Model.Interface(true).(*remote.ObjectValue), sys2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if !sys1.Equal(sys2) {
		t.Error("query sys2 faield")
		return
	}

	_, sys2Err = o1.Delete(sys2Model)
	if sys2Err != nil {
		t.Errorf("delete system failed, err:%s", sys2Err.Error())
		return
	}
	_, user1Err = o1.Delete(user1Model)
	if user1Err != nil {
		t.Errorf("delete user1 failed, err:%s", user1Err.Error())
		return
	}
	_, user2Err = o1.Delete(user2Model)
	if user2Err != nil {
		t.Errorf("delete user2 failed, err:%s", user2Err.Error())
	}
}
