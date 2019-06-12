package test

import (
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

	data, dataErr := remote.EncodeObjectValue(objVal)
	if dataErr != nil {
		err = dataErr
		return
	}
	ret, err = remote.DecodeObjectValue(data)
	if err != nil {
		return
	}

	return
}

func TestRemoteExecutor(t *testing.T) {

	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	val := &Unit{ID: 10, I8: 1, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}

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

	objList := []interface{}{objDef}
	registerMode(o1, objList)

	err = o1.Drop(objDef, "default")
	if err != nil {
		t.Errorf("drop ext failed, err:%s", err.Error())
		return
	}

	err = o1.Create(objDef, "default")
	if err != nil {
		t.Errorf("create obj failed, err:%s", err.Error())
		return
	}

	objVal, objErr := getObjectValue(val)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(objVal, "default")
	if err != nil {
		t.Errorf("insert obj failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(objVal, val)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
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
	err = o1.Update(objVal, "default")
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

	err = o1.Query(objVal2, "default")
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
		return
	}

	err = remote.UpdateEntity(objVal2, val2)
	if err != nil {
		t.Errorf("UpdateEntity failed, err:%s", err.Error())
		return
	}

	if val.Name != val2.Name || val.Value != val2.Value {
		t.Errorf("query obj failed, obj:%v, obj2:%v", val, val2)
		return
	}

	err = o1.Delete(objVal2, "default")
	if err != nil {
		t.Errorf("query obj failed, err:%s", err.Error())
	}

}

func TestRemoteDepends(t *testing.T) {
	orm.Initialize("root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	val := &Unit{ID: 10, I64: uint64(78962222222), Name: "Hello world", Value: 12.3456, TimeStamp: now, Flag: true}
	extVal := &ExtUnit{Unit: val}

	objDef, objErr := remote.GetObject(val)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	extObjDef, objErr := remote.GetObject(extVal)
	if objErr != nil {
		t.Errorf("GetObject failed, err:%s", objErr.Error())
		return
	}

	extVal2 := &ExtUnitList{Unit: *val, UnitList: []Unit{}}
	ext2ObjDef, objErr := remote.GetObject(extVal2)
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

	objList := []interface{}{objDef, extObjDef, ext2ObjDef}
	registerMode(o1, objList)

	err = o1.Drop(objDef, "default")
	if err != nil {
		t.Errorf("drop unit failed, err:%s", err.Error())
		return
	}

	err = o1.Create(objDef, "default")
	if err != nil {
		t.Errorf("create unit failed, err:%s", err.Error())
		return
	}

	err = o1.Drop(extObjDef, "default")
	if err != nil {
		t.Errorf("drop ext failed, err:%s", err.Error())
		return
	}

	err = o1.Create(extObjDef, "default")
	if err != nil {
		t.Errorf("create ext failed, err:%s", err.Error())
		return
	}

	extObjVal, objErr := getObjectValue(extVal)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}

	err = o1.Insert(extObjVal, "default")
	if err != nil {
		t.Errorf("insert ext failed, err:%s", err.Error())
		return
	}

	extVal2.UnitList = append(extVal2.UnitList, *val)

	err = o1.Drop(ext2ObjDef, "default")
	if err != nil {
		t.Errorf("drop ext2 failed, err:%s", err.Error())
		return
	}

	ext2ObjVal, objErr := getObjectValue(extVal2)
	if objErr != nil {
		t.Errorf("GetObjectValue failed, err:%s", objErr.Error())
		return
	}
	err = o1.Create(ext2ObjDef, "default")
	if err != nil {
		t.Errorf("create ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(ext2ObjVal, "default")
	if err != nil {
		t.Errorf("insert ext2 failed, err:%s", err.Error())
		return
	}

	err = o1.Delete(ext2ObjVal, "default")
	if err != nil {
		t.Errorf("delete ext2 failed, err:%s", err.Error())
	}

}
