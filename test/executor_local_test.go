package test

import (
	"testing"
	"time"

	orm "github.com/muidea/magicOrm"
)

func registerMode(orm orm.Orm, objList []interface{}) {
	for _, val := range objList {
		orm.RegisterModel(val)
	}
}

func TestLocalExecutor(t *testing.T) {

	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	obj := &Unit{ID: 10, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Unit{}}
	registerMode(o1, objList)

	err = o1.Create(obj)
	if err != nil {
		t.Errorf("create obj failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(obj)
	if err != nil {
		t.Errorf("insert obj failed, err:%s", err.Error())
		return
	}

	obj.Name = "abababa"
	obj.Value = 100.000
	err = o1.Update(obj)
	if err != nil {
		t.Errorf("update obj failed, err:%s", err.Error())
		return
	}

	obj2 := &Unit{ID: obj.ID}
	err = o1.Query(obj2)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
		return
	}
	if obj.Name != obj2.Name || obj.Value != obj2.Value {
		t.Errorf("query obj failed, obj:%v, obj2:%v", obj, obj2)
		return
	}

	err = o1.Delete(obj)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}

}

func TestLocalDepends(t *testing.T) {
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	obj := &Unit{ID: 10, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}
	ext := &ExtUnit{Unit: obj}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Unit{}, &ExtUnit{}, &ExtUnitList{}}
	registerMode(o1, objList)

	err = o1.Drop(ext)
	if err != nil {
		t.Errorf("drop ext failed, err:%s", err.Error())
		return
	}

	err = o1.Create(ext)
	if err != nil {
		t.Errorf("create ext failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(ext)
	if err != nil {
		t.Errorf("insert ext failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(obj)
	if err != nil {
		t.Errorf("insert ext failed, err:%s", err.Error())
		return
	}

	ext2 := &ExtUnitList{Unit: *obj, UnitList: []Unit{}}
	ext2.UnitList = append(ext2.UnitList, *obj)
	err = o1.Drop(ext2)
	if err != nil {
		t.Errorf("drop ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Create(ext2)
	if err != nil {
		t.Errorf("create ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(ext2)
	if err != nil {
		t.Errorf("insert ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Delete(ext2)
	if err != nil {
		t.Errorf("delete ext2 failed, err:%s", err.Error())
	}

}
