package test

import (
	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"testing"
)

func TestDefine(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	localProvider := provider.NewLocalProvider(localOwner)

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Sub{}, &Parent{}}
	modelList, modelErr := registerModel(localProvider, objList)
	if modelErr != nil {
		t.Errorf("register model failed. err:%s", modelErr.Error())
		return
	}

	mErr := dropModel(o1, modelList)
	if mErr != nil {
		t.Errorf("drop model failed. err:%s", mErr.Error())
		return
	}

	mErr = createModel(o1, modelList)
	if mErr != nil {
		t.Errorf("create model failed. err:%s", mErr.Error())
		return
	}

}
