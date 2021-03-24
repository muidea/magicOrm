package test

import (
	"github.com/muidea/magicOrm/provider"
	"testing"

	"github.com/muidea/magicOrm/orm"
)

func TestKPI(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialize()

	config := orm.NewConfig("root", "rootkit", "localhost:3306", "testdb")
	provider := provider.NewLocalProvider("default")

	o1, err := orm.NewOrm(provider, config)
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Goal{}, &SpecialGoal{}, &KPI{}}
	registerModel(provider, objList)

	goal := &Goal{Type: ByPiece, Value: 10}
	goalModel, goalErr := provider.GetEntityModel(goal)
	if goalErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", goalErr.Error())
		return
	}

	err = o1.Drop(goalModel)
	if err != nil {
		t.Errorf("drop goal failed, err:%s", err.Error())
		return
	}

	err = o1.Create(goalModel)
	if err != nil {
		t.Errorf("create goal failed, err:%s", err.Error())
		return
	}

	specailGoal := &SpecialGoal{CheckDistrict: []string{"123", "234"}, CheckProduct: []string{"111"}, CheckType: CheckSingle, CheckValue: *goal}
	specailGoalModel, specailGoalErr := provider.GetEntityModel(specailGoal)
	if specailGoalErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", specailGoalErr.Error())
		return
	}
	err = o1.Drop(specailGoalModel)
	if err != nil {
		t.Errorf("drop specailGoal failed, err:%s", err.Error())
		return
	}

	err = o1.Create(specailGoalModel)
	if err != nil {
		t.Errorf("create specailGoal failed, err:%s", err.Error())
		return
	}

	kpi := &KPI{Title: "testKPI", JoinValue: *goal, PerMonthValue: *goal, SpecialValue: *specailGoal}
	kpiModel, kpiErr := provider.GetEntityModel(kpi)
	if kpiErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", kpiErr.Error())
		return
	}
	err = o1.Drop(kpiModel)
	if err != nil {
		t.Errorf("drop kpi failed, err:%s", err.Error())
		return
	}

	err = o1.Create(kpiModel)
	if err != nil {
		t.Errorf("create kpi failed, err:%s", err.Error())
		return
	}

	kpiModel, kpiErr = o1.Insert(kpiModel)
	if kpiErr != nil {
		t.Errorf("insert kpi failed, err:%s", kpiErr.Error())
		return
	}
	kpi = kpiModel.Interface(true).(*KPI)

	goal1 := &Goal{Type: ByMoney, Value: 1234}
	kpi.JoinValue = *goal1
	kpiModel, kpiErr = provider.GetEntityModel(kpi)
	if kpiErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", kpiErr.Error())
		return
	}
	kpiModel, kpiErr = o1.Update(kpiModel)
	if kpiErr != nil {
		t.Errorf("update kpi failed, err:%s", kpiErr.Error())
		return
	}

	_, err = o1.Delete(kpiModel)
	if err != nil {
		t.Errorf("delete kpi failed, err:%s", err.Error())
		return
	}
}
