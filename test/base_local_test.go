package test

import (
	orm "github.com/muidea/magicOrm"
	"testing"
	"time"
)

const localOwner = "local"

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
	err = registerModel(o1, objList, localOwner)
	if err != nil {
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	ts, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	s1 := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	err = o1.Drop(s1, localOwner)
	if err != nil {
		t.Errorf("drop simple schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(s1, localOwner)
	if err != nil {
		t.Errorf("create simple schema failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(s1, localOwner)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}

	s1.Name = "hello"
	err = o1.Update(s1, localOwner)
	if err != nil {
		t.Errorf("update simple failed, err:%s", err.Error())
		return
	}

	s2 := &Simple{ID: s1.ID}
	err = o1.Query(s2, localOwner)
	if err != nil {
		t.Errorf("query simple failed, err:%s", err.Error())
		return
	}
	if !s1.IsSame(s2) {
		t.Errorf("Query simple failed.")
	}
}

func TestLocalReference(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", true)
	defer orm.Uninitialize()

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	objList := []interface{}{&Reference{}}
	err = registerModel(o1, objList, localOwner)
	if err != nil {
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	ts, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	strValue := "test code"
	fValue := float32(12.34)
	flag := true
	iArray := []int{12, 23, 34}
	fArray := []float32{12.34, 23, 45, 45, 67}
	strArray := []string{"Abc", "Bcd"}
	bArray := []bool{true, true, false, false}
	strPtrArray := []*string{&strValue, &strValue}
	s1 := &Reference{
		Name:        strValue,
		FValue:      &fValue,
		F64:         23.456,
		TimeStamp:   &ts,
		Flag:        &flag,
		IArray:      iArray,
		FArray:      fArray,
		StrArray:    strArray,
		BArray:      bArray,
		PtrArray:    &strArray,
		StrPtrArray: strPtrArray,
		PtrStrArray: &strPtrArray,
	}

	err = o1.Drop(s1, localOwner)
	if err != nil {
		t.Errorf("drop simple schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(s1, localOwner)
	if err != nil {
		t.Errorf("create simple schema failed, err:%s", err.Error())
		return
	}

	err = o1.Insert(s1, localOwner)
	if err != nil {
		t.Errorf("insert simple failed, err:%s", err.Error())
		return
	}

	s1.Name = "hello"
	err = o1.Update(s1, localOwner)
	if err != nil {
		t.Errorf("update simple failed, err:%s", err.Error())
		return
	}

	fValue2 := float32(0.0)
	var ts2 time.Time
	var bVal bool
	var strArray2 []string
	var ptrArray2 []*string
	s2 := &Reference{
		ID:          s1.ID,
		FValue:      &fValue2,
		TimeStamp:   &ts2,
		Flag:        &bVal,
		PtrArray:    &strArray2,
		PtrStrArray: &ptrArray2,
	}

	err = o1.Query(s2, localOwner)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	if !s1.IsSame(s2) {
		t.Errorf("Query reference failed.")
		return
	}

	err = o1.Insert(s2, localOwner)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	if s1.IsSame(s2) {
		t.Errorf("Query reference failed.")
		return
	}

	s4 := &Reference{
		ID: s1.ID,
	}
	err = o1.Query(s4, localOwner)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	if s4.Name != s2.Name {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	if s4.FValue != nil || s4.TimeStamp != nil || s4.Flag != nil || s4.PtrStrArray != nil || s4.PtrArray != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
}
