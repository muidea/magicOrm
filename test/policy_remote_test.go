package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemotePolicy(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider("default", nil)

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

	objList := []any{valueItemDef, valueScopeDef, statusDef, rewardPolicyDef}
	_, err = registerLocalModel(remoteProvider, objList)
	if err != nil {
		t.Errorf("registerLocalModel failed, err:%s", err.Error())
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
	s1Model, s1Err := remoteProvider.GetEntityModel(s1Value, true)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err = o1.Insert(s1Model)
	if s1Err != nil {
		err = s1Err
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s1Model.Interface(true).(*remote.ObjectValue), status)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
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
	rewardPolicyModel, rewardPolicyErr := remoteProvider.GetEntityModel(rewardPolicyValue, true)
	if rewardPolicyErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", rewardPolicyErr.Error())
		return
	}
	rewardPolicyModel, rewardPolicyErr = o1.Insert(rewardPolicyModel)
	if rewardPolicyErr != nil {
		err = rewardPolicyErr
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(rewardPolicyModel.Interface(true).(*remote.ObjectValue), rewardPolicy)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
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
	filter, err := remoteProvider.GetModelFilter(rewardPolicyObject)
	if err != nil {
		t.Errorf("GetEntityFilter failed, err:%s", err.Error())
		return
	}

	err = filter.Equal("status", statusValue)
	if err != nil {
		t.Errorf("filter.Equal failed, err:%s", err.Error())
		return
	}

	err = filter.ValueMask(maskVal)
	if err != nil {
		t.Errorf("filter.ValueMask failed, err:%s", err.Error())
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
