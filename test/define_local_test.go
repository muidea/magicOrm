package test

import (
	"testing"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
)

func TestDefine(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	localProvider := provider.NewLocalProvider("localOwner")

	o1, err := orm.NewOrm(localProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []any{&Sub{}, &Parent{}}
	modelList, modelErr := registerLocalModel(localProvider, objList)
	if modelErr != nil {
		err = modelErr
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	err = dropModel(o1, modelList)
	if err != nil {
		t.Errorf("drop model failed. err:%s", err.Error())
		return
	}

	err = createModel(o1, modelList)
	if err != nil {
		t.Errorf("create model failed. err:%s", err.Error())
		return
	}

}
