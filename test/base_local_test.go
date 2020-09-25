package test

import (
	orm "github.com/muidea/magicOrm"
	"testing"
	"time"
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

	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	s1 := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: now, Flag: true}

	err = o1.Insert(s1, "test")
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}

	s1.Name = "hello"
	err = o1.Update(s1, "test")
	if err != nil {
		t.Errorf("update simple failed, err:%s", err.Error())
		return
	}

	s2 := &Simple{ID: s1.ID}
	err = o1.Query(s2, "test")
	if err != nil {
		t.Errorf("query simple failed, err:%s", err.Error())
		return
	}
}
