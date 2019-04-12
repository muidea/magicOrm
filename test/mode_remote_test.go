package test

import (
	"log"
	"testing"

	"github.com/muidea/magicCommon/foundation/util"
	orm "github.com/muidea/magicOrm"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteGroup(t *testing.T) {
	//orm.Initialize("root", "rootkit", "localhost:9696", "testdb")
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

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

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef}
	registerMode(o1, objList)

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

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(group1Val)
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(group1Val, group1)
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
	err = o1.Insert(group2Val)
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
	err = o1.Insert(group3Val)
	if err != nil {
		t.Errorf("insert Group3 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(group3Val, group3)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(group3Val)
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
	err = o1.Query(group4Val)
	if err != nil {
		t.Errorf("query Group4 failed, err:%s", err.Error())
		return
	}

	group5 := &Group{ID: group2.ID, Parent: &Group{}}
	group5Val, objErr := getObjectValue(group5)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Query(group5Val)
	if err != nil {
		t.Errorf("query Group5 failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(group5Val, group5)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}
	if !group5.Equle(group2) {
		t.Errorf("query Group5 failed")
	}
}

func TestRemoteUser(t *testing.T) {
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

	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef}
	registerMode(o1, objList)

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

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(group1Val)
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
	err = o1.Insert(group2Val)
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
	err = o1.Insert(group3Val)
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
	user1Val, objErr := getObjectValue(user1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(user1Val)
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

	err = o1.Query(user2Val)
	if err != nil {
		t.Errorf("query user2 failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user2Val, user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if !user2.Equle(user1) {
		t.Errorf("query user2 failed")
		return
	}

	user1.Group = append(user1.Group, group3)
	user1Val, objErr = getObjectValue(user1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Update(user1Val)
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

	err = o1.Query(user2Val)
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
	if !user2.Equle(user1) {
		t.Errorf("query user2 failed")
		return
	}

	err = o1.Delete(group1Val)
	if err != nil {
		t.Errorf("delete group1 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(group2Val)
	if err != nil {
		t.Errorf("delete group2 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(group3Val)
	if err != nil {
		t.Errorf("delete group3 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(user2Val)
	if err != nil {
		t.Errorf("delete user2 failed, err:%s", err.Error())
	}

}

func TestRemoteSystem(t *testing.T) {
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

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

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef, sysDef}
	registerMode(o1, objList)

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

	user1Val, objErr := getObjectValue(user1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(user1Val)
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
	err = o1.Insert(user2Val)
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
	err = o1.Insert(sys1Val)
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
	err = o1.Update(sys1Val)
	if err != nil {
		t.Errorf("update system failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(sys1Val, sys1)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	sys2 := &System{ID: sys1.ID, Users: &[]User{}}
	sys2Val, objErr := getObjectValue(sys2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Query(sys2Val)
	if err != nil {
		t.Errorf("query system failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(sys2Val, sys2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if !sys1.Equle(sys2) {
		t.Error("query sys2 faield")
		return
	}

	err = o1.Delete(sys2Val)
	if err != nil {
		t.Errorf("delete system failed, err:%s", err.Error())
		return
	}

	err = o1.Delete(user1Val)
	if err != nil {
		t.Errorf("delete user1 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(user2Val)
	if err != nil {
		t.Errorf("delete user2 failed, err:%s", err.Error())
	}
}

func TestRemoteBatchQuery(t *testing.T) {
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

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

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{groupDef, userDef}
	registerMode(o1, objList)

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

	group1Val, objErr := getObjectValue(group1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(group1Val)
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
	err = o1.Insert(group2Val)
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
	err = o1.Insert(group3Val)
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

	user1Val, objErr := getObjectValue(user1)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Insert(user1Val)
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
	err = o1.Insert(user2Val)
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(user2Val, user2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	userList := []User{}
	filter := orm.NewFilter()
	filter.Equle("Name", &user1.Name)
	filter.In("Group", user1.Group)
	filter.Like("EMail", user1.EMail)

	pageFilter := &util.PageFilter{PageNum: 0, PageSize: 100}
	filter.PageFilter(pageFilter)

	userListVal, objErr := remote.GetSliceObjectValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}
	retVal, retErr := o1.BatchQuery(userListVal, nil)
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	log.Print(retVal)

	if len(userList) != 2 {
		t.Errorf("batch query user failed")
		return
	}

	userList = []User{}
	userListVal, objErr = remote.GetSliceObjectValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}

	retVal, retErr = o1.BatchQuery(&userListVal, filter)
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	log.Print(retVal)
	if len(userList) != 1 {
		t.Errorf("filter query user failed")
		return
	}
	if userList[0].Name != user1.Name || len(userList[0].Group) != len(user1.Group) {
		t.Errorf("filter query user failed")
		return
	}

	gs := []*Group{group1}
	filter2 := orm.NewFilter()
	filter2.In("Group", gs)
	userList = []User{}
	userListVal, objErr = remote.GetSliceObjectValue(userList)
	if objErr != nil {
		t.Errorf("GetSliceObjectValue failed, err:%s", objErr.Error())
		return
	}
	retVal, retErr = o1.BatchQuery(&userListVal, filter2)
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}
	log.Print(retVal)
	log.Print(userList)
}
