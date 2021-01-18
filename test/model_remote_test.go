package test

import (
	"testing"

	"github.com/muidea/magicCommon/foundation/util"
	orm "github.com/muidea/magicOrm"
	"github.com/muidea/magicOrm/provider/remote"
)

func getSliceObjectValue(val interface{}) (ret *remote.SliceObjectValue, err error) {
	objVal, objErr := remote.GetSliceObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	data, dataErr := remote.EncodeSliceObjectValue(objVal)
	if dataErr != nil {
		err = dataErr
		return
	}

	ret, err = remote.DecodeSliceObjectValue(data)
	if err != nil {
		return
	}

	return
}

func getSliceObjectPtrValue(val interface{}) (ret *remote.SliceObjectValue, err error) {
	objVal, objErr := remote.GetSliceObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	data, dataErr := remote.EncodeSliceObjectValue(objVal)
	if dataErr != nil {
		err = dataErr
		return
	}

	ret, err = remote.DecodeSliceObjectValue(data)
	if err != nil {
		return
	}

	return
}

func TestRemoteGroup(t *testing.T) {
	//orm.Initialize("root", "rootkit", "localhost:9696", "testdb")
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	status := &Status{Value: 10}
	statusDef, objErr := remote.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	user1 := &User{}
	userDef, objErr := remote.GetObject(user1)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	groupDef, objErr := remote.GetObject(group1)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	o1, err := orm.NewOrm()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef, statusDef}
	registerModel(o1, objList, "default")

	err = o1.Drop(statusDef, "default")
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}

	err = o1.Create(statusDef, "default")
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(userDef, "default")
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef, "default")
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(groupDef, "default")
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}

	err = o1.Create(groupDef, "default")
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	statusVal, objErr := getObjectValue(status)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(statusVal, "default")
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(statusVal, status)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(group1Val, "default")
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(group1Val, group1)
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
	err = o1.Query(qGroup1Val, "default")
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(qGroup1Val, qGroup1)
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
	err = o1.Insert(group2Val, "default")
	if err != nil {
		t.Errorf("insert Group2 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group2Val, group2)
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
	err = o1.Insert(group3Val, "default")
	if err != nil {
		t.Errorf("insert Group3 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(group3Val, group3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(group3Val, "default")
	if err != nil {
		t.Errorf("delete Group3 failed, err:%s", err.Error())
		return
	}

	group4 := &Group{ID: group2.ID, Parent: &Group{}}
	group4Val, objErr := getObjectValue(group4)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Query(group4Val, "default")
	if err != nil {
		t.Errorf("query Group4 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(group4Val, group4)
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
	err = o1.Query(group5Val, "default")
	if err != nil {
		t.Errorf("query Group5 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(group5Val, group5)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	if !group5.Equal(group2) {
		t.Errorf("query Group5 failed")
	}
}

func TestRemoteUser(t *testing.T) {
	status := &Status{Value: 10}
	statusDef, objErr := remote.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	user0 := &User{}
	userDef, objErr := remote.GetObject(user0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	groupDef, objErr := remote.GetObject(group1)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	o1, err := orm.NewOrm()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef, statusDef}
	registerModel(o1, objList, "default")

	err = o1.Drop(statusDef, "default")
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}

	err = o1.Create(statusDef, "default")
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(userDef, "default")
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef, "default")
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(groupDef, "default")
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}

	err = o1.Create(groupDef, "default")
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	statusVal, objErr := getObjectValue(status)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(statusVal, "default")
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(statusVal, status)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(group1Val, "default")
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group1Val, group1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group2Val, objErr := getObjectValue(group2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group2Val, "default")
	if err != nil {
		t.Errorf("insert Group2 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group2Val, group2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group3Val, objErr := getObjectValue(group3)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group3Val, "default")
	if err != nil {
		t.Errorf("insert group3 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group3Val, group3)
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

	err = o1.Insert(user1Val, "default")
	if err != nil {
		t.Errorf("insert user1 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user1Val, user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user2 := &User{ID: user1.ID}
	user2Val, objErr := getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Query(user2Val, "default")
	if err != nil {
		t.Errorf("query user2 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user2Val, user2)
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
	err = o1.Update(user1Val, "default")
	if err != nil {
		t.Errorf("update user1 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user1Val, user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	user2Val, objErr = getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Query(user2Val, "default")
	if err != nil {
		t.Errorf("query user2 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user2Val, user2)
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

	err = o1.Delete(group1Val, "default")
	if err != nil {
		t.Errorf("delete group1 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(group2Val, "default")
	if err != nil {
		t.Errorf("delete group2 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(group3Val, "default")
	if err != nil {
		t.Errorf("delete group3 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(user2Val, "default")
	if err != nil {
		t.Errorf("delete user2 failed, err:%s", err.Error())
	}

}

func TestRemoteSystem(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	status := &Status{Value: 10}
	statusDef, objErr := remote.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	user0 := &User{}
	userDef, objErr := remote.GetObject(user0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group0 := &Group{}
	groupDef, objErr := remote.GetObject(group0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	sys0 := &System{}
	sysDef, objErr := remote.GetObject(sys0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	o1, err := orm.NewOrm()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef, statusDef, sysDef}
	registerModel(o1, objList, "default")

	err = o1.Drop(statusDef, "default")
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}

	err = o1.Create(statusDef, "default")
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(userDef, "default")
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(sysDef, "default")
	if err != nil {
		t.Errorf("drop system failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef, "default")
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	statusVal, objErr := getObjectValue(status)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(statusVal, "default")
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(statusVal, status)
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

	err = o1.Insert(user1Val, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user1Val, user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	user2Val, objErr := getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(user2Val, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user2Val, user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	users := []User{*user1, *user2}
	sys1 := &System{Name: "sys1", Tags: []string{"aab", "ccd"}}
	sys1.Users = &users

	err = o1.Create(sysDef, "default")
	if err != nil {
		t.Errorf("create system failed, err:%s", err.Error())
		return
	}

	sys1Val, objErr := getObjectValue(sys1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(sys1Val, "default")
	if err != nil {
		t.Errorf("insert system failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(sys1Val, sys1)
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
	err = o1.Update(sys1Val, "default")
	if err != nil {
		t.Errorf("update system failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(sys1Val, sys1)
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
	err = o1.Query(sys2Val, "default")
	if err != nil {
		t.Errorf("query system failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(sys2Val, sys2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if !sys1.Equal(sys2) {
		t.Error("query sys2 faield")
		return
	}

	err = o1.Delete(sys2Val, "default")
	if err != nil {
		t.Errorf("delete system failed, err:%s", err.Error())
		return
	}

	err = o1.Delete(user1Val, "default")
	if err != nil {
		t.Errorf("delete user1 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(user2Val, "default")
	if err != nil {
		t.Errorf("delete user2 failed, err:%s", err.Error())
	}
}

func TestRemoteBatchQuery(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	status := &Status{Value: 10}
	statusDef, objErr := remote.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	user0 := &User{}
	userDef, objErr := remote.GetObject(user0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group0 := &Group{}
	groupDef, objErr := remote.GetObject(group0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	o1, err := orm.NewOrm()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef, statusDef}
	registerModel(o1, objList, "default")

	err = o1.Drop(statusDef, "default")
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}

	err = o1.Create(statusDef, "default")
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(groupDef, "default")
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}
	err = o1.Create(groupDef, "default")
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group1Val, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group1Val, group1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group2Val, objErr := getObjectValue(group2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group2Val, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group2Val, group2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	group3Val, objErr := getObjectValue(group3)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group3Val, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group3Val, group3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)

	err = o1.Drop(userDef, "default")
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef, "default")
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	user1Val, objErr := getObjectValue(user1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(user1Val, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user1Val, user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user2.Group = append(user2.Group, group1)
	user2.Group = append(user2.Group, group3)
	user2Val, objErr := getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(user2Val, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user2Val, user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	maskVal, maskErr := remote.GetObjectValue(&User{Group: []*Group{}})
	if maskErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", maskErr.Error())
		return
	}

	groupList := []*Group{group1, group2}
	groupListVal, groupListErr := remote.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	userList := &[]User{}
	filter := o1.QueryFilter("default")
	filter.Equal("Name", &user1.Name)
	filter.In("Group", groupListVal)
	filter.Like("EMail", user1.EMail)
	filter.ValueMask(maskVal)

	pageFilter := &util.PageFilter{PageNum: 0, PageSize: 100}
	filter.Page(pageFilter)

	userListVal, objErr := getSliceObjectValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}
	retErr := o1.BatchQuery(userListVal, nil, "default")
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	retErr = remote.UpdateSliceEntity(userListVal, userList)
	if retErr != nil {
		err = retErr
		t.Errorf("UpdateSliceEntity failed, err:%s", err.Error())
		return
	}

	if len(*userList) != 2 {
		t.Errorf("batch query user failed")
		return
	}

	userList = &[]User{}
	userListVal, objErr = getSliceObjectValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}

	retErr = o1.BatchQuery(userListVal, filter, "default")
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	retErr = remote.UpdateSliceEntity(userListVal, userList)
	if retErr != nil {
		err = retErr
		t.Errorf("UpdateSliceEntity failed, err:%s", err.Error())
		return
	}
	if len(*userList) != 1 {
		t.Errorf("filter query user failed")
		return
	}
	if (*userList)[0].Name != user1.Name || len((*userList)[0].Group) != len(user1.Group) {
		t.Errorf("filter query user failed")
		return
	}

	groupList = []*Group{group1}
	groupListVal, groupListErr = remote.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}
	filter2 := o1.QueryFilter("default")
	filter2.In("Group", groupListVal)
	userList = &[]User{}
	userListVal, objErr = getSliceObjectValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}
	retErr = o1.BatchQuery(userListVal, filter2, "default")
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	retErr = remote.UpdateSliceEntity(userListVal, userList)
	if retErr != nil {
		err = retErr
		t.Errorf("UpdateSliceEntity failed, err:%s", err.Error())
		return
	}
	if len(*userList) != 2 {
		t.Errorf("filter query user failed")
		return
	}
}

func TestRemoteBatchQueryPtr(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	status := &Status{Value: 10}
	statusDef, objErr := remote.GetObject(status)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	user0 := &User{}
	userDef, objErr := remote.GetObject(user0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group0 := &Group{}
	groupDef, objErr := remote.GetObject(group0)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	o1, err := orm.NewOrm()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef, statusDef}
	registerModel(o1, objList, "default")

	err = o1.Drop(statusDef, "default")
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}

	err = o1.Create(statusDef, "default")
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(groupDef, "default")
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}
	err = o1.Create(groupDef, "default")
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	statusVal, objErr := getObjectValue(status)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(statusVal, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(statusVal, status)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	user1.Status = status

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group1Val, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group1Val, group1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	group2Val, objErr := getObjectValue(group2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group2Val, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group2Val, group2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	group3Val, objErr := getObjectValue(group3)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group3Val, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(group3Val, group3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)

	err = o1.Drop(userDef, "default")
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(userDef, "default")
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	user1Val, objErr := getObjectValue(user1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(user1Val, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user1Val, user1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	user2.Status = status
	user2.Group = append(user2.Group, group1)
	user2.Group = append(user2.Group, group3)
	user2Val, objErr := getObjectValue(user2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(user2Val, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user2Val, user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	maskValue := &User{Status: &Status{}}

	maskVal, _ := getObjectValue(maskValue)
	groupList := []*Group{group1, group2}
	groupListVal, groupListErr := remote.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	userList := &[]*User{}
	filter := o1.QueryFilter("default")
	filter.Equal("Name", &user1.Name)
	filter.In("Group", groupListVal)
	filter.Like("EMail", user1.EMail)
	filter.ValueMask(maskVal)

	pageFilter := &util.PageFilter{PageNum: 0, PageSize: 100}
	filter.Page(pageFilter)

	userListVal, objErr := getSliceObjectPtrValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}
	retErr := o1.BatchQuery(userListVal, nil, "default")
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	retErr = remote.UpdateSliceEntity(userListVal, userList)
	if retErr != nil {
		err = retErr
		t.Errorf("UpdateSlicePtrEntity failed, err:%s", err.Error())
		return
	}

	if len(*userList) != 2 {
		t.Errorf("batch query user failed")
		return
	}

	userList = &[]*User{}
	userListVal, objErr = getSliceObjectPtrValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}

	retErr = o1.BatchQuery(userListVal, filter, "default")
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	retErr = remote.UpdateSliceEntity(userListVal, userList)
	if retErr != nil {
		err = retErr
		t.Errorf("UpdateSliceEntity failed, err:%s", err.Error())
		return
	}
	if len(*userList) != 1 {
		t.Errorf("filter query user failed")
		return
	}
	if (*userList)[0].Name != user1.Name || len((*userList)[0].Group) != len(user1.Group) {
		t.Errorf("filter query user failed")
		return
	}
	if (*userList)[0].Status == nil {
		t.Errorf("valueMask failed")
		return
	}

	groupList = []*Group{group1}
	groupListVal, groupListErr = remote.GetSliceObjectValue(groupList)
	if groupListErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", groupListErr.Error())
		return
	}

	filter2 := o1.QueryFilter("default")
	filter2.In("Group", groupListVal)
	userList = &[]*User{}
	userListVal, objErr = getSliceObjectPtrValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}
	retErr = o1.BatchQuery(userListVal, filter2, "default")
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	retErr = remote.UpdateSliceEntity(userListVal, userList)
	if retErr != nil {
		err = retErr
		t.Errorf("UpdateSlicePtrEntity failed, err:%s", err.Error())
		return
	}
	if len(*userList) != 2 {
		t.Errorf("filter query user failed")
		return
	}
}
