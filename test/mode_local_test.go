package test

import (
	"log"
	"testing"

	"github.com/muidea/magicCommon/foundation/util"
	orm "github.com/muidea/magicOrm"
)

func TestLocalGroup(t *testing.T) {
	//orm.Initialize("root", "rootkit", "localhost:9696", "testdb")
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Group{}, &User{}, &Status{}}
	registerMode(o1, objList)

	err = o1.Drop(group1, "default")
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}

	err = o1.Create(group1, "default")
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(group1, "default")
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}
	log.Printf("group1:%v", group1)
	group2.Parent = group1
	err = o1.Insert(group2, "default")
	if err != nil {
		t.Errorf("insert Group2 failed, err:%s", err.Error())
		return
	}
	log.Printf("group2:%v", group2)

	group3.Parent = group1
	err = o1.Insert(group3, "default")
	if err != nil {
		t.Errorf("insert Group3 failed, err:%s", err.Error())
		return
	}
	log.Printf("group3:%v", group3)

	err = o1.Delete(group3, "default")
	if err != nil {
		t.Errorf("delete Group3 failed, err:%s", err.Error())
		return
	}

	group4 := &Group{ID: group2.ID, Name: group2.Name}
	err = o1.Query(group4, "default")
	if err != nil {
		t.Errorf("query Group4 failed, err:%s", err.Error())
		return
	}

	group42 := &Group{ID: group2.ID, Name: group2.Name, Parent: &Group{}}
	err = o1.Query(group42, "default")
	if err != nil {
		t.Errorf("query Group42 failed, err:%s", err.Error())
		return
	}

	group5 := &Group{Parent: &Group{ID: 1}}
	err = o1.Query(group5, "default")
	if err != nil {
		t.Errorf("query Group5 failed, err:%s", err.Error())
		return
	}

	if !group5.Equal(group2) {
		t.Errorf("query Group5 failed")
	}
}

func TestLocalUser(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	status := &Status{Value: 10}
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Group{}, &User{}, &Status{}}
	registerMode(o1, objList)

	err = o1.Drop(status, "default")
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}

	err = o1.Create(status, "default")
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(status, "default")
	if err != nil {
		t.Errorf("insert status failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(group1, "default")
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}

	err = o1.Create(group1, "default")
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(group1, "default")
	if err != nil {
		t.Errorf("insert Group1 failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(group2, "default")
	if err != nil {
		t.Errorf("insert Group2 failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(group3, "default")
	if err != nil {
		t.Errorf("insert group3 failed, err:%s", err.Error())
		return
	}

	user1 := &User{Name: "demo", EMail: "123@demo.com", Status: status, Group: []*Group{}}
	err = o1.Drop(user1, "default")
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(user1, "default")
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)
	err = o1.Insert(user1, "default")
	if err != nil {
		t.Errorf("insert user1 failed, err:%s", err.Error())
		return
	}

	user2 := &User{ID: user1.ID, Status: &Status{}}
	err = o1.Query(user2, "default")
	if err != nil {
		t.Errorf("query user2 failed, err:%s", err.Error())
		return
	}

	if !user2.Equal(user1) {
		t.Errorf("query user2 failed")
		return
	}

	user1.Group = append(user1.Group, group3)
	err = o1.Update(user1, "default")
	if err != nil {
		t.Errorf("update user1 failed, err:%s", err.Error())
		return
	}

	err = o1.Query(user2, "default")
	if err != nil {
		t.Errorf("query user2 failed, err:%s", err.Error())
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

	err = o1.Delete(group1, "default")
	if err != nil {
		t.Errorf("delete group1 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(group2, "default")
	if err != nil {
		t.Errorf("delete group2 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(group3, "default")
	if err != nil {
		t.Errorf("delete group3 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(user2, "default")
	if err != nil {
		t.Errorf("delete user2 failed, err:%s", err.Error())
	}

}

func TestLoalSystem(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Group{}, &User{}, &System{}, &Status{}}
	registerMode(o1, objList)

	err = o1.Drop(user1, "default")
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	sys1 := &System{Name: "sys1", Tags: []string{"aab", "ccd"}}
	err = o1.Drop(sys1, "default")
	if err != nil {
		t.Errorf("drop system failed, err:%s", err.Error())
		return
	}

	err = o1.Create(user1, "default")
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(user1, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}
	err = o1.Insert(user2, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}

	users := []User{*user1, *user2}
	sys1.Users = &users

	err = o1.Create(sys1, "default")
	if err != nil {
		t.Errorf("create system failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(sys1, "default")
	if err != nil {
		t.Errorf("insert system failed, err:%s", err.Error())
		return
	}

	users = append(users, *user1)
	users = append(users, *user2)
	sys1.Users = &users
	err = o1.Update(sys1, "default")
	if err != nil {
		t.Errorf("update system failed, err:%s", err.Error())
		return
	}

	sys2 := &System{ID: sys1.ID, Users: &[]User{}}
	err = o1.Query(sys2, "default")
	if err != nil {
		t.Errorf("query system failed, err:%s", err.Error())
		return
	}

	if !sys1.Equal(sys2) {
		t.Error("query sys2 faield")
		return
	}

	err = o1.Delete(sys2, "default")
	if err != nil {
		t.Errorf("delete system failed, err:%s", err.Error())
		return
	}

	err = o1.Delete(user1, "default")
	if err != nil {
		t.Errorf("delete user1 failed, err:%s", err.Error())
		return
	}
	err = o1.Delete(user2, "default")
	if err != nil {
		t.Errorf("delete user2 failed, err:%s", err.Error())
	}
}

func TestLocalBatchQuery(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	status := &Status{Value: 10}
	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}

	user1 := &User{Name: "demo1", EMail: "123@demo.com"}
	user2 := &User{Name: "demo2", EMail: "123@demo.com"}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Group{}, &User{}, &Status{}}
	registerMode(o1, objList)

	err = o1.Drop(status, "default")
	if err != nil {
		t.Errorf("drop status failed, err:%s", err.Error())
		return
	}
	err = o1.Create(status, "default")
	if err != nil {
		t.Errorf("create status failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(status, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(group1, "default")
	if err != nil {
		t.Errorf("drop group failed, err:%s", err.Error())
		return
	}
	err = o1.Create(group1, "default")
	if err != nil {
		t.Errorf("create group failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(group1, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(group2, "default")
	if err != nil {
		t.Errorf("insert group failed, err:%s", err.Error())
		return
	}

	user1.Group = append(user1.Group, group1)
	user1.Group = append(user1.Group, group2)
	user1.Status = status

	err = o1.Drop(user1, "default")
	if err != nil {
		t.Errorf("drop user failed, err:%s", err.Error())
		return
	}

	err = o1.Create(user1, "default")
	if err != nil {
		t.Errorf("create user failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(user1, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(user2, "default")
	if err != nil {
		t.Errorf("insert user failed, err:%s", err.Error())
		return
	}

	valueMask := User{Status: &Status{}}
	var userList []User
	filter := o1.QueryFilter("default")
	filter.Equal("Name", &user1.Name)
	filter.In("Group", user1.Group)
	filter.Like("EMail", user1.EMail)
	filter.Equal("Status", status)
	filter.ValueMask(valueMask)

	pageFilter := &util.PageFilter{PageNum: 0, PageSize: 100}
	filter.Page(pageFilter)

	retErr := o1.BatchQuery(&userList, filter, "default")
	//retErr := o1.BatchQuery(&userList, nil, "default")
	if retErr != nil {
		err = retErr
		t.Errorf("batch query user failed, err:%s", err.Error())
		return
	}

	if len(userList) != 1 {
		t.Errorf("filter query user failed")
		return
	}
	if userList[0].Status == nil {
		t.Errorf("filter valueMask failed")
		return
	}
}
