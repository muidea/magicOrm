package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalSystem(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider("default", nil)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	objList := []any{&Group{}, &User{}, &System{}, &Status{}}
	registerLocalModel(localProvider, objList)

	userModel, userErr := localProvider.GetEntityModel(User{}, true)
	if userErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", userErr.Error())
		return
	}

	err = o1.Drop(userModel)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	sysModel, sysErr := localProvider.GetEntityModel(System{}, true)
	if sysErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", sysErr.Error())
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

	err = o1.Drop(sysModel)
	if err != nil {
		t.Errorf("drop system failed, err:%s", err.Error())
		return
	}
	err = o1.Create(sysModel)
	if err != nil {
		t.Errorf("create system failed, err:%s", err.Error())
		return
	}

	user1Model, user1Err := localProvider.GetEntityModel(user1, true)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user1Model, user1Err = o1.Insert(user1Model)
	if user1Err != nil {
		t.Errorf("insert user failed, err:%s", user1Err.Error())
		return
	}
	user1 = user1Model.Interface(true).(*User)

	user2Model, user2Err := localProvider.GetEntityModel(user2, true)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Insert(user2Model)
	if user2Err != nil {
		t.Errorf("insert user failed, err:%s", user2Err.Error())
		return
	}
	user2 = user2Model.Interface(true).(*User)

	sys1 := &System{Name: "sys1", Tags: []string{"aab", "ccd"}}

	users := []User{*user1, *user2}
	sys1.Users = &users
	sys1Model, sys1Err := localProvider.GetEntityModel(sys1, true)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}

	sys1Model, sys1Err = o1.Insert(sys1Model)
	if sys1Err != nil {
		t.Errorf("insert user failed, err:%s", sys1Err.Error())
		return
	}
	sys1 = sys1Model.Interface(true).(*System)

	users = append(users, *user1)
	users = append(users, *user2)
	sys1.Users = &users
	sys1Model, sys1Err = localProvider.GetEntityModel(sys1, true)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys1Model, sys1Err = o1.Update(sys1Model)
	if sys1Err != nil {
		t.Errorf("insert user failed, err:%s", sys1Err.Error())
		return
	}
	sys1 = sys1Model.Interface(true).(*System)

	sys2 := &System{ID: sys1.ID, Users: &[]User{}, Tags: []string{}}
	sys2Model, sys2Err := localProvider.GetEntityModel(sys2, true)
	if sys2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys2Err.Error())
		return
	}
	sys2Model, sys2Err = o1.Query(sys2Model)
	if sys2Err != nil {
		t.Errorf("query user failed, err:%s", sys2Err.Error())
		return
	}
	sys2 = sys2Model.Interface(true).(*System)

	if !sys1.Equal(sys2) {
		t.Error("query sys2 faield")
		return
	}

	sys2Model, sys2Err = localProvider.GetEntityModel(sys2, true)
	if sys2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys2Err.Error())
		return
	}
	_, sys2Err = o1.Delete(sys2Model)
	if sys2Err != nil {
		t.Errorf("insert user failed, err:%s", sys2Err.Error())
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
		return
	}
}
