package test

import (
	"testing"
	"time"

	orm "github.com/muidea/magicOrm"
)

func TestLocalExecutor(t *testing.T) {

	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
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
	registerModel(o1, objList, "default")

	err = o1.Create(obj, "default")
	if err != nil {
		t.Errorf("create obj failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(obj, "default")
	if err != nil {
		t.Errorf("insert obj failed, err:%s", err.Error())
		return
	}

	obj.Name = "abababa"
	obj.Value = 100.000
	err = o1.Update(obj, "default")
	if err != nil {
		t.Errorf("update obj failed, err:%s", err.Error())
		return
	}

	obj2 := &Unit{ID: obj.ID}
	err = o1.Query(obj2, "default")
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
		return
	}
	if obj.Name != obj2.Name || obj.Value != obj2.Value {
		t.Errorf("query obj failed, obj:%v, obj2:%v", obj, obj2)
		return
	}

	_, countErr := o1.Count(obj2, nil, "default")
	if countErr != nil {
		t.Errorf("count object failed, err:%s", countErr.Error())
		return
	}

	err = o1.Delete(obj, "default")
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}

}

func TestLocalDepends(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
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
	registerModel(o1, objList, "default")

	err = o1.Drop(ext, "default")
	if err != nil {
		t.Errorf("drop ext failed, err:%s", err.Error())
		return
	}

	err = o1.Create(ext, "default")
	if err != nil {
		t.Errorf("create ext failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(ext, "default")
	if err != nil {
		t.Errorf("insert ext failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(obj, "default")
	if err != nil {
		t.Errorf("insert ext failed, err:%s", err.Error())
		return
	}

	ext2 := &ExtUnitList{Unit: *obj, UnitList: []Unit{}}
	ext2.UnitList = append(ext2.UnitList, *obj)
	err = o1.Drop(ext2, "default")
	if err != nil {
		t.Errorf("drop ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Create(ext2, "default")
	if err != nil {
		t.Errorf("create ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(ext2, "default")
	if err != nil {
		t.Errorf("insert ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Delete(ext2, "default")
	if err != nil {
		t.Errorf("delete ext2 failed, err:%s", err.Error())
		return
	}

}
