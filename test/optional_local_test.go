package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestLocalOptional(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider(localOwner, nil)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Optional{}}
	_, err = registerLocalModel(localProvider, objList)
	if err != nil {
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	opt001 := Optional{Name: "abc", StrArry: []string{"a", "b", "c"}}
	optionalModel, err := localProvider.GetEntityModel(opt001, true)
	if err != nil {
		t.Errorf("GetEntityModel failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(optionalModel)
	if err != nil {
		t.Errorf("drop optional failed, err:%s", err.Error())
		return
	}

	err = o1.Create(optionalModel)
	if err != nil {
		t.Errorf("create optional failed, err:%s", err.Error())
		return
	}

	newOpt000 := optionalModel.Interface(true).(*Optional)
	if newOpt000.Name != opt001.Name {
		t.Errorf("insert optional failed, missmatch name, expect:%s, actual:%s", opt001.Name, newOpt000.Name)
	}
	if newOpt000.Optional != nil {
		t.Errorf("insert optional failed, missmatch optional, expect nil, actual:%+v", newOpt000.Optional)
	}
	if newOpt000.OptionnalStrArray != nil {
		t.Errorf("insert optional failed, missmatch optionnalStrArray, expect nil, actual:%+v", newOpt000.OptionnalStrArray)
	}

	optionalModel, err = o1.Insert(optionalModel)
	if err != nil {
		t.Errorf("insert optional failed, err:%s", err.Error())
	}
	newOpt001 := optionalModel.Interface(true).(*Optional)
	if newOpt001.Name != opt001.Name {
		t.Errorf("insert optional failed, missmatch name, expect:%s, actual:%s", opt001.Name, newOpt001.Name)
	}
	if newOpt001.Optional != nil {
		t.Errorf("insert optional failed, missmatch optional, expect nil, actual:%+v", newOpt001.Optional)
	}
	if newOpt001.OptionnalStrArray != nil {
		t.Errorf("insert optional failed, missmatch optionnalStrArray, expect nil, actual:%+v", newOpt001.OptionnalStrArray)
	}

	err = optionalModel.SetFieldValue("name", "def")
	if err != nil {
		t.Errorf("set optional failed, err:%s", err.Error())
	}
	optionalModel, err = o1.Update(optionalModel)
	if err != nil {
		t.Errorf("update optional failed, err:%s", err.Error())
	}
	newOpt002 := optionalModel.Interface(true).(*Optional)
	if newOpt002.Name != "def" {
		t.Errorf("update optional failed, missmatch name, expect:%s, actual:%s", "def", newOpt002.Name)
	}

	err = optionalModel.SetFieldValue("name", "ghi")
	if err != nil {
		t.Errorf("set optional failed, err:%s", err.Error())
	}
	newPtr := &opt001.Name
	err = optionalModel.SetFieldValue("optional", newPtr)
	if err != nil {
		t.Errorf("set optional failed, err:%s", err.Error())
	}
	optionalModel, err = o1.Update(optionalModel)
	if err != nil {
		t.Errorf("update optional failed, err:%s", err.Error())
	}
	newOpt003 := optionalModel.Interface(true).(*Optional)
	if newOpt003.Name != "ghi" {
		t.Errorf("update optional failed, missmatch name, expect:%s, actual:%s", "ghi", newOpt003.Name)
	}
	if *newOpt003.Optional != opt001.Name {
		t.Errorf("update optional failed, missmatch optional, expect:%s, actual:%s", opt001.Name, *newOpt003.Optional)
	}
}
