package test

import (
	"testing"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalGroup(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	provider := provider.NewLocalProvider("default")

	o1, err := orm.NewOrm(provider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Group{}, &User{}, &Status{}}
	registerModel(provider, objList)

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	gModel, gErr := provider.GetEntityModel(group1)
	if gErr != nil {
		t.Errorf("new Orm failed, err:%s", gErr.Error())
		return
	}

	err = o1.Drop(gModel)
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}

	err = o1.Create(gModel)
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	group1Model, group1Err := provider.GetEntityModel(group1)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group1Err)
		return
	}
	group1Model, group1Err = o1.Insert(group1Model)
	if group1Err != nil {
		t.Errorf("insert Group1 failed, err:%s", group1Err.Error())
		return
	}

	group2.Parent = group1Model.Interface(true, 0).(*Group)
	group2Model, group2Err := provider.GetEntityModel(group2)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group2Err)
		return
	}

	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group2Err.Error())
		return
	}
	group2 = group2Model.Interface(true, 0).(*Group)

	group3.Parent = group1Model.Interface(true, 0).(*Group)
	group3Model, group3Err := provider.GetEntityModel(group3)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group3Err)
		return
	}
	group3Model, group3Err = o1.Insert(group3Model)
	if group3Err != nil {
		t.Errorf("insert Group3 failed, err:%s", group3Err.Error())
		return
	}

	group3Model, group3Err = o1.Delete(group3Model)
	if group3Err != nil {
		t.Errorf("delete Group3 failed, err:%s", group3Err.Error())
		return
	}

	group4 := &Group{ID: group2.ID, Name: group2.Name}
	group4Model, group4Err := provider.GetEntityModel(group4)
	if group4Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group4Err)
		return
	}
	group4Model, group4Err = o1.Query(group4Model)
	if group4Err != nil {
		t.Errorf("query Group4 failed, err:%s", group4Err.Error())
		return
	}

	group42 := &Group{ID: group2.ID, Name: group2.Name, Parent: &Group{}}
	group42Model, group42Err := provider.GetEntityModel(group42)
	if group42Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group42Err)
		return
	}
	group42Model, group42Err = o1.Query(group42Model)
	if group42Err != nil {
		t.Errorf("query Group42 failed, err:%s", group42Err.Error())
		return
	}
	group42 = group42Model.Interface(true, 0).(*Group)
	if !group42.Equal(group2) {
		t.Errorf("query Group42 failed")
		return
	}

	group5 := &Group{Parent: &Group{ID: 1}}
	group5Model, group5Err := provider.GetEntityModel(group5)
	if group5Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group5Err)
		return
	}
	group5Model, group5Err = o1.Query(group5Model)
	if group5Err != nil {
		t.Errorf("query Group4 failed, err:%s", group5Err.Error())
		return
	}
	group5 = group5Model.Interface(true, 0).(*Group)
	if !group5.Equal(group2) {
		t.Errorf("query Group5 failed")
	}
}

func TestLocalUser(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	provider := provider.NewLocalProvider("default")

	o1, err := orm.NewOrm(provider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	status := &Status{Value: 10}
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	objList := []interface{}{&Group{}, &User{}, &Status{}}
	registerModel(provider, objList)

	statusModel, statusErr := provider.GetEntityModel(status)
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
	status = statusModel.Interface(true, 0).(*Status)

	groupModel, groupErr := provider.GetEntityModel(group1)
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
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}
	group1 = groupModel.Interface(true, 0).(*Group)

	group2Model, group2Err := provider.GetEntityModel(group2)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group2Err.Error())
		return
	}
	group2 = group2Model.Interface(true, 0).(*Group)

	group3Model, group3Err := provider.GetEntityModel(group3)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group3Model, group3Err = o1.Insert(group3Model)
	if group3Err != nil {
		t.Errorf("insert Group3 failed, err:%s", group3Err.Error())
		return
	}
	group3 = group3Model.Interface(true, 0).(*Group)

	user1 := &User{Name: "demo", EMail: "123@demo.com", Status: status, Group: []*Group{}}
	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)

	userModel, userErr := provider.GetEntityModel(user1)
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
	user1 = userModel.Interface(true, 0).(*User)

	user2 := &User{ID: user1.ID, Status: &Status{}, Group: []*Group{}}
	user2Model, user2Err := provider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Query(user2Model)
	if user2Err != nil {
		t.Errorf("query user2 failed, err:%s", user2Err.Error())
		return
	}
	user2 = user2Model.Interface(true, 0).(*User)

	if !user2.Equal(user1) {
		t.Errorf("query user2 failed")
		return
	}

	user1.Group = append(user1.Group, group3)
	user1Model, user1Err := provider.GetEntityModel(user1)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user1Model, user1Err = o1.Update(user1Model)
	if user1Err != nil {
		t.Errorf("update user1 failed, err:%s", user1Err.Error())
		return
	}
	user1 = user1Model.Interface(true, 0).(*User)

	user2Model, user2Err = provider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Query(user2Model)
	if user2Err != nil {
		t.Errorf("query user2 failed, err:%s", user2Err.Error())
		return
	}
	user2 = user2Model.Interface(true, 0).(*User)
	if len(user2.Group) != 3 {
		t.Errorf("query user2 failed")
		return
	}
	if !user2.Equal(user1) {
		t.Errorf("query user2 failed")
		return
	}

	group1Model, group1Err := provider.GetEntityModel(group1)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}
	group1Model, group1Err = o1.Delete(group1Model)
	if group1Err != nil {
		t.Errorf("delete group1 failed, err:%s", group1Err.Error())
		return
	}

	group2Model, group2Err = provider.GetEntityModel(group2)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group2Model, group2Err = o1.Delete(group2Model)
	if group2Err != nil {
		t.Errorf("delete group1 failed, err:%s", group2Err.Error())
		return
	}

	group3Model, group3Err = provider.GetEntityModel(group3)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group3Model, group3Err = o1.Delete(group3Model)
	if group3Err != nil {
		t.Errorf("delete group1 failed, err:%s", group3Err.Error())
		return
	}

	user2Model, user2Err = provider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Delete(user2Model)
	if user2Err != nil {
		t.Errorf("delete group1 failed, err:%s", user2Err.Error())
		return
	}
}

func TestLocalSystem(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider("default")

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	objList := []interface{}{&Group{}, &User{}, &System{}, &Status{}}
	registerModel(localProvider, objList)

	userModel, userErr := localProvider.GetEntityModel(User{})
	if userErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", userErr.Error())
		return
	}

	err = o1.Drop(userModel)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	sysModel, sysErr := localProvider.GetEntityModel(System{})
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

	user1Model, user1Err := localProvider.GetEntityModel(user1)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user1Model, user1Err = o1.Insert(user1Model)
	if user1Err != nil {
		t.Errorf("insert user failed, err:%s", user1Err.Error())
		return
	}
	user1 = user1Model.Interface(true, 0).(*User)

	user2Model, user2Err := localProvider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Insert(user2Model)
	if user2Err != nil {
		t.Errorf("insert user failed, err:%s", user2Err.Error())
		return
	}
	user2 = user2Model.Interface(true, 0).(*User)

	sys1 := &System{Name: "sys1", Tags: []string{"aab", "ccd"}}

	users := []User{*user1, *user2}
	sys1.Users = &users
	sys1Model, sys1Err := localProvider.GetEntityModel(sys1)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys1Model, sys1Err = o1.Insert(sys1Model)
	if sys1Err != nil {
		t.Errorf("insert user failed, err:%s", sys1Err.Error())
		return
	}
	sys1 = sys1Model.Interface(true, 0).(*System)

	users = append(users, *user1)
	users = append(users, *user2)
	sys1.Users = &users
	sys1Model, sys1Err = localProvider.GetEntityModel(sys1)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys1Model, sys1Err = o1.Update(sys1Model)
	if sys1Err != nil {
		t.Errorf("insert user failed, err:%s", sys1Err.Error())
		return
	}
	sys1 = sys1Model.Interface(true, 0).(*System)

	sys2 := &System{ID: sys1.ID, Users: &[]User{}, Tags: []string{}}
	sys2Model, sys2Err := localProvider.GetEntityModel(sys2)
	if sys2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys2Err.Error())
		return
	}
	sys2Model, sys2Err = o1.Query(sys2Model)
	if sys2Err != nil {
		t.Errorf("query user failed, err:%s", sys2Err.Error())
		return
	}
	sys2 = sys2Model.Interface(true, 0).(*System)

	if !sys1.Equal(sys2) {
		t.Error("query sys2 faield")
		return
	}

	sys2Model, sys2Err = localProvider.GetEntityModel(sys2)
	if sys2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys2Err.Error())
		return
	}
	sys2Model, sys2Err = o1.Delete(sys2Model)
	if sys2Err != nil {
		t.Errorf("insert user failed, err:%s", sys2Err.Error())
		return
	}

	user1Model, user1Err = o1.Delete(user1Model)
	if user1Err != nil {
		t.Errorf("delete user1 failed, err:%s", user1Err.Error())
		return
	}
	user2Model, user2Err = o1.Delete(user2Model)
	if user2Err != nil {
		t.Errorf("delete user2 failed, err:%s", user2Err.Error())
		return
	}
}

func TestLocalBatchQuery(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider("default")

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

	objList := []interface{}{&Group{}, &User{}, &Status{}}
	registerModel(localProvider, objList)

	statusModel, statusErr := localProvider.GetEntityModel(status)
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
	status = statusModel.Interface(true, 0).(*Status)

	group1Model, group1Err := localProvider.GetEntityModel(group1)
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
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	group1 = group1Model.Interface(true, 0).(*Group)

	group2Model, group2Err := localProvider.GetEntityModel(group2)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert group failed, err:%s", group2Err.Error())
		return
	}
	group2 = group2Model.Interface(true, 0).(*Group)

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)
	user1.Status = status

	user1Model, user1Err := localProvider.GetEntityModel(user1)
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
	user1 = user1Model.Interface(true, 0).(*User)

	user2Model, user2Err := localProvider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user2Model, user2Err = o1.Insert(user2Model)
	if user2Err != nil {
		t.Errorf("insert user failed, err:%s", user2Err.Error())
		return
	}

	valueMask := &User{Status: &Status{}}
	uModel, _ := localProvider.GetEntityModel(&User{})
	filter, err := localProvider.GetModelFilter(uModel)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}
	filter.Equal("name", &user1.Name)
	filter.In("group", user1.Group)
	filter.Like("email", user1.EMail)
	filter.Equal("status", status)
	filter.ValueMask(valueMask)

	pageFilter := &util.Pagination{PageNum: 0, PageSize: 100}
	filter.Page(pageFilter)

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
