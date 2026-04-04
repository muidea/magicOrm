package test

import (
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalReference(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider(localOwner, nil)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Reference{}}
	_, err = registerLocalModel(localProvider, objList)
	if err != nil {
		t.Errorf("register model failed. err:%s", err.Error())
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

	s1Model, s1Err := localProvider.GetEntityModel(s1, true)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}
	err = o1.Drop(s1Model)
	if err != nil {
		t.Errorf("drop simple schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(s1Model)
	if err != nil {
		t.Errorf("create simple schema failed, err:%s", err.Error())
		return
	}

	s1Model, err = o1.Insert(s1Model)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(true).(*Reference)

	s1.Name = "hello"
	s1Model, s1Err = localProvider.GetEntityModel(s1, true)
	if s1Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s1Err.Error())
		return
	}

	s1Model, err = o1.Update(s1Model)
	if err != nil {
		t.Errorf("update simple failed, err:%s", err.Error())
		return
	}
	s1 = s1Model.Interface(true).(*Reference)

	s2 := Reference{ID: s1.ID}

	s2Model, s2Err := localProvider.GetEntityModel(&s2, true)
	if s2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s2Err.Error())
		return
	}

	s2Model, err = o1.Query(s2Model)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	s2 = s2Model.Interface(false).(Reference)

	if !s1.IsSame(&s2) {
		t.Errorf("Query reference failed.")
		return
	}

	s2Model, err = o1.Insert(s2Model)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	s2 = s2Model.Interface(false).(Reference)
	if s1.IsSame(&s2) {
		t.Errorf("Query reference failed.")
		return
	}

	s4 := Reference{ID: s1.ID}
	s4Model, s4Err := localProvider.GetEntityModel(&s4, true)
	if s4Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", s4Err.Error())
		return
	}

	s4Model, err = o1.Query(s4Model)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	s4 = s4Model.Interface(false).(Reference)
	if !s1.IsSame(&s4) {
		t.Errorf("query reference failed")
		return
	}
}
