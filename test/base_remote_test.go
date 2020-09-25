package test

import (
	orm "github.com/muidea/magicOrm"
	"github.com/muidea/magicOrm/provider/remote"
	"testing"
)

func TestRemoteSimple(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	simple := &Simple{}
	simpleDef, simpleErr := remote.GetObject(simple)
	if simpleErr != nil {
		t.Errorf("GetObject failed, err:%s", simpleErr.Error())
		return
	}

	objList := []interface{}{simpleDef}
	registerModel(o1, objList)

}
