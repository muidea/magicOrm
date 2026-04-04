package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalGroup(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	provider := provider.NewLocalProvider("default", nil)

	o1, err := orm.NewOrm(provider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Group{}, &User{}, &Status{}}
	registerLocalModel(provider, objList)

	group1 := &Group{Name: "testGroup1"}
	group2 := &Group{Name: "testGroup2"}
	group3 := &Group{Name: "testGroup3"}

	gModel, gErr := provider.GetEntityModel(group1, true)
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

	group1Model, group1Err := provider.GetEntityModel(group1, true)
	if group1Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group1Err)
		return
	}
	group1Model, group1Err = o1.Insert(group1Model)
	if group1Err != nil {
		t.Errorf("insert Group1 failed, err:%s", group1Err.Error())
		return
	}

	group2.Parent = group1Model.Interface(true).(*Group)
	group2Model, group2Err := provider.GetEntityModel(group2, true)
	if group2Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group2Err)
		return
	}

	group2Model, group2Err = o1.Insert(group2Model)
	if group2Err != nil {
		t.Errorf("insert Group2 failed, err:%s", group2Err.Error())
		return
	}
	group2 = group2Model.Interface(true).(*Group)

	group3.Parent = group1Model.Interface(true).(*Group)
	group3Model, group3Err := provider.GetEntityModel(group3, true)
	if group3Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group3Err)
		return
	}
	group3Model, group3Err = o1.Insert(group3Model)
	if group3Err != nil {
		t.Errorf("insert Group3 failed, err:%s", group3Err.Error())
		return
	}

	_, group3Err = o1.Delete(group3Model)
	if group3Err != nil {
		t.Errorf("delete Group3 failed, err:%s", group3Err.Error())
		return
	}

	group4 := &Group{ID: group2.ID, Name: group2.Name}
	group4Model, group4Err := provider.GetEntityModel(group4, true)
	if group4Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group4Err)
		return
	}
	_, group4Err = o1.Query(group4Model)
	if group4Err != nil {
		t.Errorf("query Group4 failed, err:%s", group4Err.Error())
		return
	}

	group42 := &Group{ID: group2.ID}
	group42Model, group42Err := provider.GetEntityModel(group42, true)
	if group42Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group42Err)
		return
	}

	group42Model, group42Err = o1.Query(group42Model)
	if group42Err != nil {
		t.Errorf("query Group42 failed, err:%s", group42Err.Error())
		return
	}
	group42 = group42Model.Interface(true).(*Group)
	if group42.ID != group2.ID || group42.Name != group2.Name {
		t.Errorf("query Group42 basic fields failed")
		return
	}
	if group42.Parent == nil || group2.Parent == nil {
		t.Errorf("query Group42 parent failed")
		return
	}
	if group42.Parent.ID != group2.Parent.ID || group42.Parent.Name != group2.Parent.Name {
		t.Errorf("query Group42 failed")
		return
	}

	group5 := &Group{Parent: &Group{ID: 1}}
	group5Model, group5Err := provider.GetEntityModel(group5, true)
	if group5Err != nil {
		t.Errorf("GetEntityModel failed,err:%s", group5Err)
		return
	}

	group5Model, group5Err = o1.Query(group5Model)
	if group5Err != nil {
		t.Errorf("query Group4 failed, err:%s", group5Err.Error())
		return
	}
	group5 = group5Model.Interface(true).(*Group)
	if group5.ID != group2.ID || group5.Name != group2.Name {
		t.Errorf("query Group5 basic fields failed")
		return
	}
	if group5.Parent == nil || group2.Parent == nil {
		t.Errorf("query Group5 parent failed")
		return
	}
	if group5.Parent.ID != group2.Parent.ID || group5.Parent.Name != group2.Parent.Name {
		t.Errorf("query Group5 failed")
	}
}
