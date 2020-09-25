package test

import (
	orm "github.com/muidea/magicOrm"
	"testing"
)

func TestLocalSimple(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Simple{}}
	registerModel(o1, objList)

}
