package test

import (
	"github.com/muidea/magicOrm/provider"
	"testing"
	"time"

	"github.com/muidea/magicOrm/orm"
)

func TestLocalExecutor(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit")
	provider := provider.NewLocalProvider("default")

	o1, err := orm.NewOrm(provider, config)
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	obj := &Unit{ID: 10, I8: 8, I16: 1600, I32: 323200, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, F64: 12.45678, TimeStamp: now, Flag: true}

	objList := []interface{}{&Unit{}}
	registerModel(provider, objList)

	objModel, objErr := provider.GetEntityModel(obj)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	err = o1.Create(objModel)
	if err != nil {
		t.Errorf("create obj failed, err:%s", err.Error())
		return
	}

	objModel, objErr = o1.Insert(objModel)
	if err != nil {
		t.Errorf("insert obj failed, err:%s", err.Error())
		return
	}
	obj = objModel.Interface(true).(*Unit)

	obj.Name = "abababa"
	obj.Value = 100.000
	objModel, objErr = provider.GetEntityModel(obj)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}
	objModel, objErr = o1.Update(objModel)
	if objErr != nil {
		t.Errorf("update obj failed, err:%s", objErr.Error())
		return
	}

	obj2 := &Unit{ID: obj.ID}
	obj2Model, obj2Err := provider.GetEntityModel(obj2)
	if obj2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", obj2Err.Error())
		return
	}
	obj2Model, obj2Err = o1.Query(obj2Model)
	if obj2Err != nil {
		t.Errorf("query obj failed, err:%s", obj2Err.Error())
		return
	}
	obj2 = obj2Model.Interface(true).(*Unit)
	if obj.Name != obj2.Name || obj.Value != obj2.Value {
		t.Errorf("query obj failed, obj:%v, obj2:%v", obj, obj2)
		return
	}

	_, countErr := o1.Count(obj2Model, nil)
	if countErr != nil {
		t.Errorf("count object failed, err:%s", countErr.Error())
		return
	}

	_, err = o1.Delete(obj2Model)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}

}

func TestLocalDepends(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit")
	provider := provider.NewLocalProvider("default")

	o1, err := orm.NewOrm(provider, config)
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	now, _ := time.Parse("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000")
	obj := &Unit{ID: 10, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}
	ext := &ExtUnit{Unit: obj}

	objList := []interface{}{&Unit{}, &ExtUnit{}, &ExtUnitList{}}
	registerModel(provider, objList)

	extModel, extErr := provider.GetEntityModel(ext)
	if extErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", extErr.Error())
		return
	}

	err = o1.Drop(extModel)
	if err != nil {
		t.Errorf("drop ext failed, err:%s", err.Error())
		return
	}

	err = o1.Create(extModel)
	if err != nil {
		t.Errorf("create ext failed, err:%s", err.Error())
		return
	}

	extModel, extErr = o1.Insert(extModel)
	if extErr != nil {
		t.Errorf("insert ext failed, err:%s", extErr.Error())
		return
	}
	ext = extModel.Interface(true).(*ExtUnit)

	objModel, objErr := provider.GetEntityModel(obj)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}
	objModel, objErr = o1.Insert(objModel)
	if objErr != nil {
		t.Errorf("insert ext failed, err:%s", objErr.Error())
		return
	}
	obj = objModel.Interface(true).(*Unit)

	ext2 := &ExtUnitList{Unit: *obj, UnitList: []Unit{}}
	ext2.UnitList = append(ext2.UnitList, *obj)
	ext2Model, ext2Err := provider.GetEntityModel(ext2)
	if ext2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", ext2Err.Error())
		return
	}

	err = o1.Drop(ext2Model)
	if err != nil {
		t.Errorf("drop ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Create(ext2Model)
	if err != nil {
		t.Errorf("create ext2 failed, err:%s", err.Error())
		return
	}

	ext2Model, ext2Err = o1.Insert(ext2Model)
	if ext2Err != nil {
		t.Errorf("insert ext2 failed, err:%s", ext2Err.Error())
		return
	}

	_, err = o1.Delete(ext2Model)
	if err != nil {
		t.Errorf("delete ext2 failed, err:%s", err.Error())
		return
	}

}
