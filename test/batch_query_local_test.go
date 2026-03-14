package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalBatchQuery(t *testing.T) {
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

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

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
		t.Errorf("insert group failed, err:%s", statusErr.Error())
		return
	}
	status = statusModel.Interface(true).(*Status)

	group1Model, group1Err := localProvider.GetEntityModel(group1, true)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}
	err = o1.Drop(group1Model)
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}
	err = o1.Create(group1Model)
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	group1Model, group1Err = o1.Insert(group1Model)
	if group1Err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	group1 = group1Model.Interface(true).(*Group)

	group2Model, group2Err := localProvider.GetEntityModel(group2, true)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert group failed, err:%s", group2Err.Error())
		return
	}
	group2 = group2Model.Interface(true).(*Group)

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)
	user1.Status = status

	user1Model, user1Err := localProvider.GetEntityModel(user1, true)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}

	err = o1.Drop(user1Model)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(user1Model)
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
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
	_, user2Err = o1.Insert(user2Model)
	if user2Err != nil {
		t.Errorf("insert user failed, err:%s", user2Err.Error())
		return
	}

	valueMask := &User{Status: &Status{}}
	uModel, _ := localProvider.GetEntityModel(&User{}, true)
	filter, err := localProvider.GetModelFilter(uModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}
	err = filter.Equal("name", &user1.Name)
	if err != nil {
		t.Errorf("Equal name failed, error:%s", err.Error())
		return
	}
	err = filter.In("group", user1.Group)
	if err != nil {
		t.Errorf("Equal group failed, error:%s", err.Error())
		return
	}
	err = filter.Like("email", user1.EMail)
	if err != nil {
		t.Errorf("like email failed, error:%s", err.Error())
		return
	}

	err = filter.Equal("status", status)
	if err != nil {
		t.Errorf("Equal status failed, error:%s", err.Error())
		return
	}

	err = filter.ValueMask(valueMask)
	if err != nil {
		t.Errorf("ValueMask failed, error:%s", err.Error())
		return
	}

	filter.Pagination(0, 100)

	userModelList, userModelErr := o1.BatchQuery(filter)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	if len(userModelList) != 1 {
		t.Errorf("filter query user failed")
		return
	}
}
