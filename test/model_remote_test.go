package test

import (
	"testing"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteGroup(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("default")

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	status := &Status{Value: 10}
	statusDef, objErr := helper.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	userDef, objErr := helper.GetObject(&User{})
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	groupDef, objErr := helper.GetObject(&Group{})
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	objList := []interface{}{groupDef, userDef, statusDef}
	_, mErr := registerModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registerModel failed, err:%s", mErr.Error())
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

	statusVal, objErr := getObjectValue(status)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	statusModel, statusErr := remoteProvider.GetEntityModel(statusVal)
	if statusErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", statusErr.Error())
		return
	}

	status2Model, status2Err := o1.Insert(statusModel)
	if status2Err != nil {
		t.Errorf("insert Group1 failed, err:%s", status2Err.Error())
		return
	}

	mErr = helper.UpdateEntity(status2Model.Interface(true).(*remote.ObjectValue), status)
	if mErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", mErr.Error())
		return
	}

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	group1Model, group1Err := remoteProvider.GetEntityModel(group1Val)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}
	group11Model, group11Err := o1.Insert(group1Model)
	if group11Err != nil {
		t.Errorf("insert Group1 failed, err:%s", group11Err.Error())
		return
	}

	mErr = helper.UpdateEntity(group11Model.Interface(true).(*remote.ObjectValue), group1)
	if mErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", mErr.Error())
		return
	}

	qGroup1 := &Group{ID: group1.ID, Parent: &Group{}}
	qGroup1Val, qObjErr := getObjectValue(qGroup1)
	if qObjErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", qObjErr.Error())
		return
	}
	qGroup1Model, qGroup1Err := remoteProvider.GetEntityModel(qGroup1Val)
	if qGroup1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", qGroup1Err.Error())
		return
	}
	qGroup11Model, qGroup11Err := o1.Query(qGroup1Model)
	if qGroup11Err != nil {
		t.Errorf("insert Group1 failed, err:%s", qGroup11Err.Error())
		return
	}
	mErr = helper.UpdateEntity(qGroup11Model.Interface(true).(*remote.ObjectValue), qGroup1)
	if mErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", mErr.Error())
		return
	}

	group2.Parent = group1
	group2Val, objErr := getObjectValue(group2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group2Model, group2Err := remoteProvider.GetEntityModel(group2Val)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group22Model, group22Err := o1.Insert(group2Model)
	if group22Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group22Err.Error())
		return
	}
	mErr = helper.UpdateEntity(group22Model.Interface(true).(*remote.ObjectValue), group2)
	if mErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", mErr.Error())
		return
	}

	group3.Parent = group1
	group3Val, objErr := getObjectValue(group3)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group3Model, group3Err := remoteProvider.GetEntityModel(group3Val)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group33Model, group33Err := o1.Insert(group3Model)
	if group33Err != nil {
		t.Errorf("insert Group3 failed, err:%s", group33Err.Error())
		return
	}

	mErr = helper.UpdateEntity(group33Model.Interface(true).(*remote.ObjectValue), group3)
	if mErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", mErr.Error())
		return
	}
	group33Model, group33Err = o1.Delete(group3Model)
	if group33Err != nil {
		t.Errorf("delete Group3 failed, err:%s", group33Err.Error())
		return
	}

	group4 := &Group{ID: group2.ID, Parent: &Group{}}
	group4Val, objErr := getObjectValue(group4)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group4Model, group4Err := remoteProvider.GetEntityModel(group4Val)
	if group4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group4Err.Error())
		return
	}
	group44Model, group44Err := o1.Query(group4Model)
	if group44Err != nil {
		t.Errorf("query Group4 failed, err:%s", group44Err.Error())
		return
	}

	mErr = helper.UpdateEntity(group44Model.Interface(true).(*remote.ObjectValue), group4)
	if mErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", mErr.Error())
		return
	}

	group5 := &Group{ID: group2.ID, Parent: &Group{}}
	group5Val, objErr := getObjectValue(group5)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group5Model, group5Err := remoteProvider.GetEntityModel(group5Val)
	if group5Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group5Err.Error())
		return
	}
	group55Model, group55Err := o1.Query(group5Model)
	if group55Err != nil {
		t.Errorf("query Group5 failed, err:%s", group55Err.Error())
		return
	}

	mErr = helper.UpdateEntity(group55Model.Interface(true).(*remote.ObjectValue), group5)
	if mErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", mErr.Error())
		return
	}
	if !group5.Equal(group2) {
		t.Errorf("query Group5 failed")
	}
}

func TestRemoteUser(t *testing.T) {
	status := &Status{Value: 10}
	statusDef, objErr := helper.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	user0 := &User{}
	userDef, objErr := helper.GetObject(user0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	groupDef, objErr := helper.GetObject(group1)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("default")

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef, statusDef}
	_, mErr := registerModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registerModel failed, err:%s", mErr.Error())
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

	err = o1.Drop(userDef)
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef)
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
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

	statusVal, objErr := getObjectValue(status)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	statusModel, statusErr := remoteProvider.GetEntityModel(statusVal)
	if statusErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", statusErr.Error())
		return
	}

	status2Model, status2Err := o1.Insert(statusModel)
	if status2Err != nil {
		t.Errorf("insert Group1 failed, err:%s", status2Err.Error())
		return
	}

	eErr := helper.UpdateEntity(status2Model.Interface(true).(*remote.ObjectValue), status)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	group1Model, group1Err := remoteProvider.GetEntityModel(group1Val)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}

	group11Model, group11Err := o1.Insert(group1Model)
	if group11Err != nil {
		t.Errorf("insert Group1 failed, err:%s", group11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(group11Model.Interface(true).(*remote.ObjectValue), group1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	group2Val, objErr := getObjectValue(group2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group2Model, group2Err := remoteProvider.GetEntityModel(group2Val)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group22Model, group22Err := o1.Insert(group2Model)
	if group22Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(group22Model.Interface(true).(*remote.ObjectValue), group2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	group3Val, objErr := getObjectValue(group3)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	group3Model, group3Err := remoteProvider.GetEntityModel(group3Val)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group33Model, group33Err := o1.Insert(group3Model)
	if group33Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group33Err.Error())
		return
	}
	eErr = helper.UpdateEntity(group33Model.Interface(true).(*remote.ObjectValue), group3)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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
	user1Model, user1Err := remoteProvider.GetEntityModel(user1Val)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user11Model, user11Err := o1.Insert(user1Model)
	if user11Err != nil {
		t.Errorf("insert user1 failed, err:%s", user11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user11Model.Interface(true).(*remote.ObjectValue), user1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	user2 := &User{ID: user1.ID, Status: &Status{}, Group: []*Group{}}
	user2Val, objErr := getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	user2Model, user2Err := remoteProvider.GetEntityModel(user2Val)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user22Model, user22Err := o1.Query(user2Model)
	if user22Err != nil {
		t.Errorf("query user2 failed, err:%s", user22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user22Model.Interface(true).(*remote.ObjectValue), user2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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
	user1Model, user1Err = remoteProvider.GetEntityModel(user1Val)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user11Model, user11Err = o1.Update(user1Model)
	if user11Err != nil {
		t.Errorf("update user1 failed, err:%s", user11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user11Model.Interface(true).(*remote.ObjectValue), user1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}
	user2Val, objErr = getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	user2Model, user2Err = remoteProvider.GetEntityModel(user2Val)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user22Model, user22Err = o1.Query(user2Model)
	if user22Err != nil {
		t.Errorf("query user2 failed, err:%s", user22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user22Model.Interface(true).(*remote.ObjectValue), user2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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

	userModel, _ := remoteProvider.GetEntityModel(userObject)
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

	group11Model, group11Err = o1.Delete(group1Model)
	if group11Err != nil {
		t.Errorf("delete group1 failed, err:%s", group11Err.Error())
		return
	}
	group22Model, group22Err = o1.Delete(group2Model)
	if group22Err != nil {
		t.Errorf("delete group2 failed, err:%s", group22Err.Error())
		return
	}
	group33Model, group33Err = o1.Delete(group3Model)
	if group33Err != nil {
		t.Errorf("delete group3 failed, err:%s", group33Err.Error())
		return
	}
	user22Model, user22Err = o1.Delete(user2Model)
	if user22Err != nil {
		t.Errorf("delete user2 failed, err:%s", user22Err.Error())
		return
	}
}

func TestRemoteSystem(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("default")

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	status := &Status{Value: 10}
	statusDef, objErr := helper.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	user0 := &User{}
	userDef, objErr := helper.GetObject(user0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group0 := &Group{}
	groupDef, objErr := helper.GetObject(group0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	sys0 := &System{}
	sysDef, objErr := helper.GetObject(sys0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	objList := []interface{}{groupDef, userDef, statusDef, sysDef}
	_, mErr := registerModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registerModel failed, err:%s", mErr.Error())
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

	err = o1.Drop(userDef)
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(sysDef)
	if err != nil {
		t.Errorf("drop system failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef)
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	statusVal, objErr := getObjectValue(status)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	statusModel, statusErr := remoteProvider.GetEntityModel(statusVal)
	if statusErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", statusErr.Error())
		return
	}

	status2Model, status2Err := o1.Insert(statusModel)
	if status2Err != nil {
		t.Errorf("insert status failed, err:%s", status2Err.Error())
		return
	}

	eErr := helper.UpdateEntity(status2Model.Interface(true).(*remote.ObjectValue), status)
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
	user1Model, user1Err := remoteProvider.GetEntityModel(user1Val)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user11Model, user11Err := o1.Insert(user1Model)
	if user11Err != nil {
		t.Errorf("insert user failed, err:%s", user11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user11Model.Interface(true).(*remote.ObjectValue), user1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	user2Val, objErr := getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	user2Model, user2Err := remoteProvider.GetEntityModel(user2Val)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user22Model, user22Err := o1.Insert(user2Model)
	if user22Err != nil {
		t.Errorf("insert user2 failed, err:%s", user22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user22Model.Interface(true).(*remote.ObjectValue), user2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	users := []User{*user1, *user2}
	sys1 := &System{Name: "sys1", Tags: []string{"aab", "ccd"}}
	sys1.Users = &users

	err = o1.Create(sysDef)
	if err != nil {
		t.Errorf("create system failed, err:%s", err.Error())
		return
	}

	sys1Val, objErr := getObjectValue(sys1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	sys1Model, sys1Err := remoteProvider.GetEntityModel(sys1Val)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys11Model, sys11Err := o1.Insert(sys1Model)
	if sys11Err != nil {
		t.Errorf("insert user failed, err:%s", sys11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(sys11Model.Interface(true).(*remote.ObjectValue), sys1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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
	sys1Model, sys1Err = remoteProvider.GetEntityModel(sys1Val)
	if sys1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys1Err.Error())
		return
	}
	sys11Model, sys11Err = o1.Update(sys1Model)
	if sys11Err != nil {
		t.Errorf("update system failed, err:%s", sys11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(sys11Model.Interface(true).(*remote.ObjectValue), sys1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	sys2 := &System{ID: sys1.ID, Users: &[]User{}, Tags: []string{}}
	sys2Val, objErr := getObjectValue(sys2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	sys2Model, sys2Err := remoteProvider.GetEntityModel(sys2Val)
	if sys2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", sys2Err.Error())
		return
	}
	sys22Model, sys22Err := o1.Query(sys2Model)
	if sys22Err != nil {
		t.Errorf("query system failed, err:%s", sys22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(sys22Model.Interface(true).(*remote.ObjectValue), sys2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	if !sys1.Equal(sys2) {
		t.Error("query sys2 faield")
		return
	}

	sys2Model, sys2Err = o1.Delete(sys2Model)
	if err != nil {
		t.Errorf("delete system failed, err:%s", err.Error())
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
	}
}

func TestRemoteBatchQuery(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("default")

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

	objList := []interface{}{groupDef, userDef, statusDef}
	_, mErr := registerModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registerModel failed, err:%s", mErr.Error())
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

	group1Val, userObjectErr := getObjectValue(group1)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group1Model, group1Err := remoteProvider.GetEntityModel(group1Val)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}
	group11Model, group11Err := o1.Insert(group1Model)
	if group11Err != nil {
		t.Errorf("insert group failed, err:%s", group11Err.Error())
		return
	}
	eErr := helper.UpdateEntity(group11Model.Interface(true).(*remote.ObjectValue), group1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	group2Val, userObjectErr := getObjectValue(group2)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group2Model, group2Err := remoteProvider.GetEntityModel(group2Val)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}
	group22Model, group22Err := o1.Insert(group2Model)
	if group22Err != nil {
		t.Errorf("insert group2 failed, err:%s", group22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(group22Model.Interface(true).(*remote.ObjectValue), group2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}
	group3Val, userObjectErr := getObjectValue(group3)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group3Model, group3Err := remoteProvider.GetEntityModel(group3Val)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}
	group33Model, group33Err := o1.Insert(group3Model)
	if group33Err != nil {
		t.Errorf("insert group failed, err:%s", group33Err.Error())
		return
	}
	eErr = helper.UpdateEntity(group33Model.Interface(true).(*remote.ObjectValue), group3)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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
	user1Model, user1Err := remoteProvider.GetEntityModel(user1Val)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}
	user11Model, user11Err := o1.Insert(user1Model)
	if user11Err != nil {
		t.Errorf("insert user failed, err:%s", user11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user11Model.Interface(true).(*remote.ObjectValue), user1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	user2.Group = append(user2.Group, group1)
	user2.Group = append(user2.Group, group3)
	user2Val, userObjectErr := getObjectValue(user2)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	user2Model, user2Err := remoteProvider.GetEntityModel(user2Val)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}
	user22Model, user22Err := o1.Insert(user2Model)
	if user22Err != nil {
		t.Errorf("insert user failed, err:%s", user22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user22Model.Interface(true).(*remote.ObjectValue), user2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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

	filterVal, filterErr := remoteProvider.GetModelFilter(userObject)
	if filterErr != nil {
		t.Errorf("GetEntityFilter failed, err:%s", filterErr.Error())
		return
	}

	userModelList, userModelErr := o1.BatchQuery(filterVal)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	if len(userModelList) != 2 {
		t.Errorf("batch query user failed")
		return
	}

	fErr := filterVal.Equal("name", user1.Name)
	if fErr != nil {
		t.Errorf("filterVal.Equal err:%s", err.Error())
		return
	}

	fErr = filterVal.In("group", groupListVal)
	if fErr != nil {
		t.Errorf("filterVal.In err:%s", err.Error())
		return
	}
	fErr = filterVal.Like("email", user1.EMail)
	if fErr != nil {
		t.Errorf("filterVal.Like err:%s", err.Error())
		return
	}
	fErr = filterVal.ValueMask(maskVal)
	if fErr != nil {
		t.Errorf("filterVal.ValueMask err:%s", err.Error())
		return
	}

	pageFilter := &util.Pagination{PageNum: 0, PageSize: 100}
	filterVal.Page(pageFilter)
	userModelList, userModelErr = o1.BatchQuery(filterVal)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	if len(userModelList) != 1 {
		t.Errorf("filterVal query user failed")
		return
	}

	groupList = []*Group{group1}
	groupListVal, groupListErr = helper.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	userFilter, filterErr := remoteProvider.GetModelFilter(userObject)
	if filterErr != nil {
		t.Errorf("GetEntityFilter failed, err:%s", filterErr.Error())
		return
	}

	fErr = userFilter.In("group", groupListVal)
	if fErr != nil {
		t.Errorf("userFilter.In failed, err:%s", fErr.Error())
		return
	}

	userModelList, userModelErr = o1.BatchQuery(userFilter)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	if len(userModelList) != 2 {
		t.Errorf("filterVal query user failed")
		return
	}
}

func TestRemoteBatchQueryPtr(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("default")

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

	objList := []interface{}{groupDef, userDef, statusDef}
	_, mErr := registerModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registerModel failed, err:%s", mErr.Error())
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
	statusModel, statusErr := remoteProvider.GetEntityModel(statusVal)
	if statusErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", statusErr.Error())
		return
	}

	status2Model, status2Err := o1.Insert(statusModel)
	if status2Err != nil {
		t.Errorf("insert group failed, err:%s", status2Err.Error())
		return
	}
	eErr := helper.UpdateEntity(status2Model.Interface(true).(*remote.ObjectValue), status)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}
	user1.Status = status

	group1Val, userObjectErr := getObjectValue(group1)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group1Model, group1Err := remoteProvider.GetEntityModel(group1Val)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group1Err.Error())
		return
	}

	group11Model, group11Err := o1.Insert(group1Model)
	if group11Err != nil {
		t.Errorf("insert group failed, err:%s", group11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(group11Model.Interface(true).(*remote.ObjectValue), group1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group2Val, userObjectErr := getObjectValue(group2)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group2Model, group2Err := remoteProvider.GetEntityModel(group2Val)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group2Err.Error())
		return
	}

	group22Model, group22Err := o1.Insert(group2Model)
	if group22Err != nil {
		t.Errorf("insert group failed, err:%s", group22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(group22Model.Interface(true).(*remote.ObjectValue), group2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	group3Val, userObjectErr := getObjectValue(group3)
	if userObjectErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", userObjectErr.Error())
		return
	}
	group3Model, group3Err := remoteProvider.GetEntityModel(group3Val)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", group3Err.Error())
		return
	}

	group33Model, group33Err := o1.Insert(group3Model)
	if group33Err != nil {
		t.Errorf("insert group failed, err:%s", group33Err.Error())
		return
	}
	eErr = helper.UpdateEntity(group33Model.Interface(true).(*remote.ObjectValue), group3)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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
	user1Model, user1Err := remoteProvider.GetEntityModel(user1Val)
	if user1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user1Err.Error())
		return
	}

	user11Model, user11Err := o1.Insert(user1Model)
	if user11Err != nil {
		t.Errorf("insert group failed, err:%s", user11Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user11Model.Interface(true).(*remote.ObjectValue), user1)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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
	user2Model, user2Err := remoteProvider.GetEntityModel(user2Val)
	if user2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", user2Err.Error())
		return
	}

	user22Model, user22Err := o1.Insert(user2Model)
	if user22Err != nil {
		t.Errorf("insert group failed, err:%s", user22Err.Error())
		return
	}
	eErr = helper.UpdateEntity(user22Model.Interface(true).(*remote.ObjectValue), user2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
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
	filterVal, filterErr := remoteProvider.GetModelFilter(userObject)
	if filterErr != nil {
		t.Errorf("GetModelFilter failed, err:%s", filterErr.Error())
		return
	}

	pageFilter := &util.Pagination{PageNum: 0, PageSize: 100}
	filterVal.Page(pageFilter)
	userModelList, userModelErr := o1.BatchQuery(filterVal)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	if len(userModelList) != 2 {
		t.Errorf("batch query user failed")
		return
	}

	fErr := filterVal.Equal("name", user1.Name)
	if fErr != nil {
		t.Errorf("filterVal.Equal failed, err:%s", fErr.Error())
		return
	}
	fErr = filterVal.In("group", groupListVal)
	if fErr != nil {
		t.Errorf("filterVal.In failed, err:%s", fErr.Error())
		return
	}
	fErr = filterVal.Like("email", user1.EMail)
	if fErr != nil {
		t.Errorf("filterVal.Like failed, err:%s", fErr.Error())
		return
	}
	fErr = filterVal.ValueMask(maskVal)
	if fErr != nil {
		t.Errorf("filterVal.ValueMask failed, err:%s", fErr.Error())
		return
	}

	userModelList, userModelErr = o1.BatchQuery(filterVal)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	if len(userModelList) != 1 {
		t.Errorf("filterVal query user failed")
		return
	}

	groupList = []*Group{group1}
	groupListVal, groupListErr = helper.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	filter2, fErr := remoteProvider.GetModelFilter(userObject)
	if fErr != nil {
		t.Errorf("GetModelFilter failed, err:%s", fErr.Error())
		return
	}

	fErr = filter2.In("group", groupListVal)
	if fErr != nil {
		t.Errorf("filterVal.In failed, err:%s", fErr.Error())
		return
	}

	userModelList, userModelErr = o1.BatchQuery(filter2)
	if userModelErr != nil {
		err = userModelErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	if len(userModelList) != 2 {
		t.Errorf("filterVal query user failed")
		return
	}
}

func TestPolicy(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("default")

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	valueItem := &ValueItem{}
	valueScope := &ValueScope{}
	status := &Status{}
	rewardPolicy := &RewardPolicy{}

	valueItemDef, valueItemErr := helper.GetObject(valueItem)
	if valueItemErr != nil {
		t.Errorf("GetObject failed, err:%s", valueItemErr.Error())
		return
	}
	valueScopeDef, valueScopeErr := helper.GetObject(valueScope)
	if valueScopeErr != nil {
		t.Errorf("GetObject failed, err:%s", valueScopeErr.Error())
		return
	}
	statusDef, statusErr := helper.GetObject(status)
	if statusErr != nil {
		t.Errorf("GetObject failed, err:%s", statusErr.Error())
		return
	}
	rewardPolicyDef, rewardPolicyErr := helper.GetObject(rewardPolicy)
	if rewardPolicyErr != nil {
		t.Errorf("GetObject failed, err:%s", rewardPolicyErr.Error())
		return
	}

	objList := []interface{}{valueItemDef, valueScopeDef, statusDef, rewardPolicyDef}
	_, mErr := registerModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("registerModel failed, err:%s", mErr.Error())
		return
	}

	err = o1.Drop(valueItemDef)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(valueItemDef)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(valueScopeDef)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(valueScopeDef)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(statusDef)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(statusDef)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
		return
	}
	err = o1.Drop(rewardPolicyDef)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(rewardPolicyDef)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
		return
	}

	status.Value = 12
	s1Value, s1Err := getObjectValue(status)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err := remoteProvider.GetEntityModel(s1Value)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s11Model, s11Err := o1.Insert(s1Model)
	if s11Err != nil {
		t.Errorf("insert reference failed, err:%s", s11Err.Error())
		return
	}
	eErr := helper.UpdateEntity(s11Model.Interface(true).(*remote.ObjectValue), status)
	if eErr != nil {
		t.Errorf("updateEntity failed, err:%s", eErr.Error())
		return
	}

	rewardPolicy.Name = "testPolicy"
	rewardPolicy.Description = "desc"
	rewardPolicy.ValueItem = append(rewardPolicy.ValueItem, ValueItem{Level: 1, Type: 1, Value: 12.34})
	rewardPolicy.ValueItem = append(rewardPolicy.ValueItem, ValueItem{Level: 2, Type: 1, Value: 12.34})
	rewardPolicy.ValueItem = append(rewardPolicy.ValueItem, ValueItem{Level: 3, Type: 1, Value: 12.34})
	rewardPolicy.ValueScope = ValueScope{LowValue: 12.34, HighValue: 34.56}
	rewardPolicy.Status = status
	rewardPolicy.Creater = 12
	rewardPolicy.UpdateTime = 1234
	rewardPolicy.Namespace = "test"

	rewardPolicyObject, _ := helper.GetObject(rewardPolicy)
	rewardPolicyValue, rewardPolicyErr := getObjectValue(rewardPolicy)
	if rewardPolicyErr != nil {
		t.Errorf("getObjectValue failed, err:%s", rewardPolicyErr.Error())
		return
	}
	rewardPolicyModel, rewardPolicyErr := remoteProvider.GetEntityModel(rewardPolicyValue)
	if rewardPolicyErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", rewardPolicyErr.Error())
		return
	}
	rewardPolicy2Model, rewardPolicy2Err := o1.Insert(rewardPolicyModel)
	if rewardPolicy2Err != nil {
		t.Errorf("insert reference failed, err:%s", rewardPolicy2Err.Error())
		return
	}
	eErr = helper.UpdateEntity(rewardPolicy2Model.Interface(true).(*remote.ObjectValue), rewardPolicy)
	if eErr != nil {
		t.Errorf("updateEntity failed, err:%s", eErr.Error())
		return
	}

	maskVal, maskErr := helper.GetObjectValue(&RewardPolicy{Status: &Status{}, ValueItem: []ValueItem{}})
	if maskErr != nil {
		t.Errorf("getObjectValue failed, err:%s", maskErr.Error())
		return
	}

	statusValue, statusErr := getObjectValue(status)
	if statusErr != nil {
		t.Errorf("getObjectValue failed, err:%s", statusErr.Error())
		return
	}
	filter, fErr := remoteProvider.GetModelFilter(rewardPolicyObject)
	if fErr != nil {
		t.Errorf("GetEntityFilter failed, err:%s", fErr.Error())
		return
	}

	fErr = filter.Equal("status", statusValue)
	if fErr != nil {
		t.Errorf("filter.Equal failed, err:%s", fErr.Error())
		return
	}

	fErr = filter.ValueMask(maskVal)
	if fErr != nil {
		t.Errorf("filter.ValueMask failed, err:%s", fErr.Error())
		return
	}

	cModelList, cModelErr := o1.BatchQuery(filter)
	if cModelErr != nil {
		t.Errorf("batch query compose failed, err:%s", cModelErr.Error())
		return
	}
	if len(cModelList) != 1 {
		t.Errorf("batch query compose failed")
		return
	}
}
