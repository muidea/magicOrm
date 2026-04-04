package test

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteReference(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	remoteProvider := provider.NewRemoteProvider(remoteOwner, nil)

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Reference{}}
	mList, mErr := registerRemoteModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	err = dropModel(o1, mList)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = createModel(o1, mList)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
		return
	}

	ts, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []string{strValue, strValue}
	s1 := &Reference{
		Name:        strValue,
		FValue:      fValue,
		F64:         23.456,
		TimeStamp:   ts,
		Flag:        flag,
		IArray:      iArray,
		FArray:      fArray,
		StrArray:    strArray,
		BArray:      bArray,
		PtrArray:    &strArray,
		StrPtrArray: strPtrArray,
		PtrStrArray: &strPtrArray,
	}

	s1Value, s1Err := getObjectValue(s1)
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
	err = helper.UpdateEntity(s1Model.Interface(true).(*remote.ObjectValue), s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	s1.Name = "hello"
	s1Value, s1Err = getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err = remoteProvider.GetEntityModel(s1Value, true)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	s1Model, s1Err = o1.Update(s1Model)
	if s1Err != nil {
		err = s1Err
		t.Errorf("update reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s1Model.Interface(true).(*remote.ObjectValue), s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	s2 := &Reference{ID: s1.ID}
	s2Value, s2Err := getObjectValue(s2)
	if s2Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s2Err.Error())
		return
	}
	s2Model, s2Err := remoteProvider.GetEntityModel(s2Value, true)
	if s2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s2Err.Error())
		return
	}
	s2Model, s2Err = o1.Query(s2Model)
	if s2Err != nil {
		err = s2Err
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s2Model.Interface(true).(*remote.ObjectValue), s2)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if !s1.IsSame(s2) {
		t.Errorf("Query reference failed.")
		return
	}

	s4 := &Reference{ID: s1.ID}
	s4Value, s4Err := getObjectValue(s4)
	if s4Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s4Err.Error())
		return
	}
	s4Model, s4Err := remoteProvider.GetEntityModel(s4Value, true)
	if s4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s4Err.Error())
		return
	}
	s4Model, s4Err = o1.Query(s4Model)
	if s4Err != nil {
		err = s4Err
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	err = helper.UpdateEntity(s4Model.Interface(true).(*remote.ObjectValue), s4)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if !s1.IsSame(s4) {
		t.Errorf("query reference failed")
		return
	}
}
