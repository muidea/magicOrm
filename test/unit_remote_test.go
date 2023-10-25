package test

import (
	"testing"
	"time"

	"github.com/muidea/magicOrm/orm"
	"github.com/muidea/magicOrm/provider"
	"github.com/muidea/magicOrm/provider/helper"
	"github.com/muidea/magicOrm/provider/remote"
)

func TestRemoteExecutor(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("default")

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	val := &Unit{ID: 10, I8: 1, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}

	objDef, objErr := helper.GetObject(val)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	objList := []interface{}{objDef}
	_, mErr := registerModel(remoteProvider, objList)
	if mErr != nil {
		t.Errorf("register mode failed, err:%s", mErr.Error())
		return
	}

	err = o1.Drop(objDef)
	if err != nil {
		t.Errorf("drop ext failed, err:%s", err.Error())
		return
	}

	err = o1.Create(objDef)
	if err != nil {
		t.Errorf("create obj failed, err:%s", err.Error())
		return
	}

	objVal, objErr := getObjectValue(val)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	objModel, objErr := remoteProvider.GetEntityModel(objVal)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}

	obj1Model, obj1Err := o1.Insert(objModel)
	if obj1Err != nil {
		t.Errorf("insert obj failed, err:%s", obj1Err.Error())
		return
	}

	eErr := helper.UpdateEntity(obj1Model.Interface(true).(*remote.ObjectValue), val)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	val.I8 = int8(124)
	val.Name = "abababa"
	val.Value = 100.000
	objVal, objErr = getObjectValue(val)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	objModel, objErr = remoteProvider.GetEntityModel(objVal)
	if objErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", objErr.Error())
		return
	}
	objModel, objErr = o1.Update(objModel)
	if err != nil {
		t.Errorf("update obj failed, err:%s", err.Error())
		return
	}

	val2 := &Unit{ID: val.ID, Name: "", Value: 0.0}
	obj2Val, objErr := getObjectValue(val2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	obj2Model, obj2Err := remoteProvider.GetEntityModel(obj2Val)
	if obj2Err != nil {
		t.Errorf("GetEntityModel failed, err:%s", obj2Err.Error())
		return
	}

	obj22Model, obj22Err := o1.Query(obj2Model)
	if obj22Err != nil {
		t.Errorf("query obj failed, err:%s", obj22Err.Error())
		return
	}

	eErr = helper.UpdateEntity(obj22Model.Interface(true).(*remote.ObjectValue), val2)
	if eErr != nil {
		t.Errorf("UpdateEntity failed, err:%s", eErr.Error())
		return
	}

	if val.Name != val2.Name || val.Value != val2.Value {
		t.Errorf("query obj failed, obj:%v, obj2:%v", val, val2)
		return
	}

	_, err = o1.Delete(obj2Model)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}

}

func TestRemoteDepends(t *testing.T) {
	orm.Initialize()
	defer orm.Uninitialized()

	config := orm.NewConfig("localhost:3306", "testdb", "root", "rootkit", "")
	remoteProvider := provider.NewRemoteProvider("default")

	o1, err := orm.NewOrm(remoteProvider, config, "abc")
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	val := &Unit{ID: 10, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}
	extVal := &ExtUnit{Unit: val}

	objDef, objErr := helper.GetObject(val)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	extObjDef, objErr := helper.GetObject(extVal)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	extVal2 := &ExtUnitList{Unit: *val, UnitList: []Unit{}}
	ext2ObjDef, objErr := helper.GetObject(extVal2)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	objList := []interface{}{objDef, extObjDef, ext2ObjDef}
	registerModel(remoteProvider, objList)

	err = o1.Drop(objDef)
	if err != nil {
		t.Errorf("drop unit failed, err:%s", err.Error())
		return
	}

	err = o1.Create(objDef)
	if err != nil {
		t.Errorf("create unit failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(extObjDef)
	if err != nil {
		t.Errorf("drop ext failed, err:%s", err.Error())
		return
	}

	err = o1.Create(extObjDef)
	if err != nil {
		t.Errorf("create ext failed, err:%s", err.Error())
		return
	}

	extObjVal, extObjErr := getObjectValue(extVal)
	if extObjErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", extObjErr.Error())
		return
	}

	extObjModel, extObjErr := remoteProvider.GetEntityModel(extObjVal)
	if extObjErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", extObjErr.Error())
		return
	}

	_, extObj1Err := o1.Insert(extObjModel)
	if extObj1Err != nil {
		t.Errorf("insert ext failed, err:%s", extObj1Err.Error())
		return
	}

	extVal2.UnitList = append(extVal2.UnitList, *val)

	err = o1.Drop(ext2ObjDef)
	if err != nil {
		t.Errorf("drop ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Create(ext2ObjDef)
	if err != nil {
		t.Errorf("create ext2 failed, err:%s", err.Error())
		return
	}

	ext2ObjVal, ext2ObjErr := getObjectValue(extVal2)
	if ext2ObjErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", ext2ObjErr.Error())
		return
	}
	ext2ObjModel, ext2ObjErr := remoteProvider.GetEntityModel(ext2ObjVal)
	if ext2ObjErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", ext2ObjErr.Error())
		return
	}

	ext22ObjModel, ext22ObjErr := o1.Insert(ext2ObjModel)
	if ext22ObjErr != nil {
		t.Errorf("insert ext2 failed, err:%s", ext22ObjErr.Error())
		return
	}

	_, err = o1.Delete(ext22ObjModel)
	if err != nil {
		t.Errorf("delete ext2 failed, err:%s", err.Error())
	}

}
