package test

import (
	"testing"

	orm "github.com/muidea/magicOrm"
)

func TestKPI(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	o1, err := orm.NewOrm()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Goal{}, &SpecialGoal{}, &KPI{}}
	registerModel(o1, objList, "default")

	goal := &Goal{Type: ByPiece, Value: 10}
	err = o1.Drop(goal, "default")
	if err != nil {
		t.Errorf("drop goal failed, err:%s", err.Error())
		return
	}

	err = o1.Create(goal, "default")
	if err != nil {
		t.Errorf("create goal failed, err:%s", err.Error())
		return
	}

	specailGoal := &SpecialGoal{CheckDistrict: []string{"123", "234"}, CheckProduct: []string{"111"}, CheckType: CheckSingle, CheckValue: *goal}
	err = o1.Drop(specailGoal, "default")
	if err != nil {
		t.Errorf("drop specailGoal failed, err:%s", err.Error())
		return
	}

	err = o1.Create(specailGoal, "default")
	if err != nil {
		t.Errorf("create specailGoal failed, err:%s", err.Error())
		return
	}

	kpi := &KPI{Title: "testKPI", JoinValue: *goal, PerMonthValue: *goal, SpecialValue: *specailGoal}
	err = o1.Drop(kpi, "default")
	if err != nil {
		t.Errorf("drop kpi failed, err:%s", err.Error())
		return
	}

	err = o1.Create(kpi, "default")
	if err != nil {
		t.Errorf("create kpi failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(kpi, "default")
	if err != nil {
		t.Errorf("insert kpi failed, err:%s", err.Error())
		return
	}

	goal1 := &Goal{Type: ByMoney, Value: 1234}
	kpi.JoinValue = *goal1
	err = o1.Update(kpi, "default")
	if err != nil {
		t.Errorf("update kpi failed, err:%s", err.Error())
		return
	}

	err = o1.Delete(kpi, "default")
	if err != nil {
		t.Errorf("delete kpi failed, err:%s", err.Error())
		return
	}
}
