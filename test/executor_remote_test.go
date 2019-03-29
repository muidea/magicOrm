package test

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	orm "github.com/muidea/magicOrm"
	"github.com/muidea/magicOrm/provider/remote"
)

func getObjectValue(val interface{}) (ret *remote.ObjectValue, err error) {
	objVal, objErr := remote.GetObjectValue(val)
	if objErr != nil {
		err = objErr
		return
	}

	data, dataErr := json.Marshal(objVal)
	if dataErr != nil {
		err = dataErr
		return
	}

	ret = &remote.ObjectValue{}
	err = json.Unmarshal(data, ret)

	return
}

func TestRemoteExecutor(t *testing.T) {

	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	val := &Unit{ID: 10, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}

	objDef, objErr := remote.GetObject(val)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
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

	err = o1.Insert(objVal)
	if err != nil {
		t.Errorf("insert obj failed, err:%s", err.Error())
		return
	}

	log.Print(*objVal)

	err = remote.UpdateObject(objVal, val)
	if err != nil {
		t.Errorf("UpdateObject failed, err:%s", err.Error())
		return
	}

	val.Name = "abababa"
	val.Value = 100.000
	objVal, objErr = getObjectValue(val)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Update(objVal)
	if err != nil {
		t.Errorf("update obj failed, err:%s", err.Error())
		return
	}

	val2 := &Unit{ID: val.ID, Name: "", Value: 0.0}
	objVal2, objErr := getObjectValue(val2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Query(objVal2)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
		return
	}
	if val.Name != val2.Name || val.Value != val2.Value {
		t.Errorf("query obj failed, obj:%v, obj2:%v", val, val2)
		return
	}

	err = o1.Delete(objVal2)
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}

}

func TestRemoteDepends(t *testing.T) {
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	val := &Unit{ID: 10, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}
	extVal := &ExtUnit{Unit: val}

	extObjDef, objErr := remote.GetObject(extVal)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}
	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
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

	extObjVal, objErr := getObjectValue(extVal)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(extObjVal)
	if err != nil {
		t.Errorf("insert ext failed, err:%s", err.Error())
		return
	}

	extVal2 := &ExtUnitList{Unit: *val, UnitList: []Unit{}}
	extVal2.UnitList = append(extVal2.UnitList, *val)

	ext2ObjDef, objErr := remote.GetObject(extVal2)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}
	err = o1.Drop(ext2ObjDef)
	if err != nil {
		t.Errorf("drop ext2 failed, err:%s", err.Error())
		return
	}

	ext2ObjVal, objErr := getObjectValue(extVal2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Create(ext2ObjVal)
	if err != nil {
		t.Errorf("create ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(ext2ObjVal)
	if err != nil {
		t.Errorf("insert ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Delete(ext2ObjVal)
	if err != nil {
		t.Errorf("delete ext2 failed, err:%s", err.Error())
	}

}
