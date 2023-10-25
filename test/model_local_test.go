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
	group11Model, group11Err := o1.Insert(group1Model)
	if group11Err != nil {
		t.Errorf("insert Group1 failed, err:%s", group11Err.Error())
		return
	}

	group2.Parent = group11Model.Interface(true).(*Group)
	group2Model, group2Err := provider.GetEntityModel(group2)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group2Err)
		return
	}

	group22Model, group22Err := o1.Insert(group2Model)
	if group22Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group22Err.Error())
		return
	}
	group2 = group22Model.Interface(true).(*Group)

	group3.Parent = group1Model.Interface(true).(*Group)
	group3Model, group3Err := provider.GetEntityModel(group3)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group3Err)
		return
	}
	group33Model, group33Err := o1.Insert(group3Model)
	if group33Err != nil {
		t.Errorf("insert Group3 failed, err:%s", group33Err.Error())
		return
	}

	group33Model, group33Err = o1.Delete(group33Model)
	if group33Err != nil {
		t.Errorf("delete Group3 failed, err:%s", group33Err.Error())
		return
	}

	group4 := &Group{ID: group2.ID, Name: group2.Name}
	group4Model, group4Err := provider.GetEntityModel(group4)
	if group4Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group4Err)
		return
	}
	_, group44Err := o1.Query(group4Model)
	if group44Err != nil {
		t.Errorf("query Group4 failed, err:%s", group44Err.Error())
		return
	}

	group42 := &Group{ID: group2.ID, Name: group2.Name, Parent: &Group{}}
	group42Model, group42Err := provider.GetEntityModel(group42)
	if group42Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group42Err)
		return
	}
	group422Model, group422Err := o1.Query(group42Model)
	if group422Err != nil {
		t.Errorf("query Group42 failed, err:%s", group422Err.Error())
		return
	}
	group42 = group422Model.Interface(true).(*Group)
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
	group55Model, group55Err := o1.Query(group5Model)
	if group55Err != nil {
		t.Errorf("query Group5 failed, err:%s", group55Err.Error())
		return
	}
	group5 = group55Model.Interface(true).(*Group)
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

	status2Model, status2Err := o1.Insert(statusModel)
	if status2Err != nil {
		t.Errorf("insert status failed, err:%s", status2Err.Error())
		return
	}
	status = status2Model.Interface(true).(*Status)

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
	group1 = groupModel.Interface(true).(*Group)

	group2Model, group2Err := provider.GetEntityModel(group2)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group22Model, group22Err := o1.Insert(group2Model)
	if group22Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group22Err.Error())
		return
	}
	group2 = group22Model.Interface(true).(*Group)

	group3Model, group3Err := provider.GetEntityModel(group3)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group33Model, group33Err := o1.Insert(group3Model)
	if group33Err != nil {
		t.Errorf("insert Group3 failed, err:%s", group33Err.Error())
		return
	}
	group3 = group33Model.Interface(true).(*Group)

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

	user20Model, user20Err := o1.Insert(userModel)
	if user20Err != nil {
		t.Errorf("insert user1 failed, err:%s", user20Err.Error())
		return
	}
	user1 = user20Model.Interface(true).(*User)

	user2 := &User{ID: user1.ID, Status: &Status{}, Group: []*Group{}}
	user2Model, user2Err := provider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user22Model, user22Err := o1.Query(user2Model)
	if user22Err != nil {
		t.Errorf("query user2 failed, err:%s", user22Err.Error())
		return
	}
	user2 = user22Model.Interface(true).(*User)

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
	user11Model, user11Err := o1.Update(user1Model)
	if user11Err != nil {
		t.Errorf("update user1 failed, err:%s", user11Err.Error())
		return
	}
	user1 = user11Model.Interface(true).(*User)

	user2Model, user2Err = provider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user22Model, user22Err = o1.Query(user2Model)
	if user22Err != nil {
		t.Errorf("query user2 failed, err:%s", user22Err.Error())
		return
	}
	user2 = user22Model.Interface(true).(*User)
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
	_, group11Err := o1.Delete(group1Model)
	if group11Err != nil {
		t.Errorf("delete group1 failed, err:%s", group11Err.Error())
		return
	}

	group2Model, group2Err = provider.GetEntityModel(group2)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group22Model, group22Err = o1.Delete(group2Model)
	if group22Err != nil {
		t.Errorf("delete group1 failed, err:%s", group22Err.Error())
		return
	}

	group3Model, group3Err = provider.GetEntityModel(group3)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group33Model, group33Err = o1.Delete(group3Model)
	if group33Err != nil {
		t.Errorf("delete group1 failed, err:%s", group33Err.Error())
		return
	}

	user2Model, user2Err = provider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user22Model, user22Err = o1.Delete(user2Model)
	if user22Err != nil {
		t.Errorf("delete group1 failed, err:%s", user22Err.Error())
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
	user11Model, user11Err := o1.Insert(user1Model)
	if user11Err != nil {
		t.Errorf("insert user failed, err:%s", user11Err.Error())
		return
	}
	user1 = user11Model.Interface(true).(*User)

	user2Model, user2Err := localProvider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user22Model, user22Err := o1.Insert(user2Model)
	if user22Err != nil {
		t.Errorf("insert user failed, err:%s", user22Err.Error())
		return
	}
	user2 = user22Model.Interface(true).(*User)

	sys1 := &System{Name: "sys1", Tags: []string{"aab", "ccd"}}

	users := []User{*user1, *user2}
	sys1.Users = &users
	sys1Model, sys1Err := localProvider.GetEntityModel(sys1)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys11Model, sys11Err := o1.Insert(sys1Model)
	if sys11Err != nil {
		t.Errorf("insert user failed, err:%s", sys11Err.Error())
		return
	}
	sys1 = sys11Model.Interface(true).(*System)

	users = append(users, *user1)
	users = append(users, *user2)
	sys1.Users = &users
	sys1Model, sys1Err = localProvider.GetEntityModel(sys1)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys11Model, sys11Err = o1.Update(sys1Model)
	if sys11Err != nil {
		t.Errorf("insert user failed, err:%s", sys11Err.Error())
		return
	}
	sys1 = sys11Model.Interface(true).(*System)

	sys2 := &System{ID: sys1.ID, Users: &[]User{}, Tags: []string{}}
	sys2Model, sys2Err := localProvider.GetEntityModel(sys2)
	if sys2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys2Err.Error())
		return
	}
	sys22Model, sys22Err := o1.Query(sys2Model)
	if sys22Err != nil {
		t.Errorf("query user failed, err:%s", sys22Err.Error())
		return
	}
	sys2 = sys22Model.Interface(true).(*System)

	if !sys1.Equal(sys2) {
		t.Error("query sys2 faield")
		return
	}

	sys2Model, sys2Err = localProvider.GetEntityModel(sys2)
	if sys2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys2Err.Error())
		return
	}
	sys22Model, sys22Err = o1.Delete(sys2Model)
	if sys22Err != nil {
		t.Errorf("insert user failed, err:%s", sys22Err.Error())
		return
	}

	user11Model, user11Err = o1.Delete(user1Model)
	if user11Err != nil {
		t.Errorf("delete user1 failed, err:%s", user11Err.Error())
		return
	}
	user22Model, user22Err = o1.Delete(user2Model)
	if user22Err != nil {
		t.Errorf("delete user2 failed, err:%s", user22Err.Error())
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

	status2Model, status2Err := o1.Insert(statusModel)
	if status2Err != nil {
		t.Errorf("insert group failed, err:%s", status2Err.Error())
		return
	}
	status = status2Model.Interface(true).(*Status)

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
	group1 = group1Model.Interface(true).(*Group)

	group2Model, group2Err := localProvider.GetEntityModel(group2)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group22Model, group22Err := o1.Insert(group2Model)
	if group22Err != nil {
		t.Errorf("insert group failed, err:%s", group22Err.Error())
		return
	}
	group2 = group22Model.Interface(true).(*Group)

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

	user11Model, user11Err := o1.Insert(user1Model)
	if user11Err != nil {
		t.Errorf("insert user failed, err:%s", user11Err.Error())
		return
	}
	user1 = user11Model.Interface(true).(*User)

	user2Model, user2Err := localProvider.GetEntityModel(user2)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	_, user22Err := o1.Insert(user2Model)
	if user22Err != nil {
		t.Errorf("insert user failed, err:%s", user22Err.Error())
		return
	}

	valueMask := &User{Status: &Status{}}
	uModel, _ := localProvider.GetEntityModel(&User{})
	filterVal, filterErr := localProvider.GetModelFilter(uModel)
	if filterErr != nil {
		t.Errorf("GetEntityFilter failed, err:%s", filterErr.Error())
		return
	}
	filterVal.Equal("name", &user1.Name)
	filterVal.In("group", user1.Group)
	filterVal.Like("email", user1.EMail)
	filterVal.Equal("status", status)
	filterVal.ValueMask(valueMask)

	pageFilter := &util.Pagination{PageNum: 0, PageSize: 100}
	filterVal.Page(pageFilter)

	userModelList, userModelErr := o1.BatchQuery(filterVal)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	if len(userModelList) != 1 {
		t.Errorf("filterVal query user failed")
		return
	}
}
