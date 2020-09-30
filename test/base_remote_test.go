package test

import (
	orm "github.com/muidea/magicOrm"
	"github.com/muidea/magicOrm/provider/remote"
	"testing"
	"time"
)

const remoteOwner = "remote"

func TestRemoteSimple(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	reference := &Simple{}
	simpleDef, simpleErr := remote.GetObject(reference)
	if simpleErr != nil {
		t.Errorf("GetObject failed, err:%s", simpleErr.Error())
		return
	}

	objList := []interface{}{simpleDef}
	registerModel(o1, objList, remoteOwner)

	err = o1.Drop(simpleDef, remoteOwner)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(simpleDef, remoteOwner)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
		return
	}

	ts, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	s1 := &Simple{I8: 12, I16: 23, I32: 34, I64: 45, Name: "test code", Value: 12.345, F64: 23.456, TimeStamp: ts, Flag: true}

	s1Value, s1Err := getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}

	err = o1.Insert(s1Value, remoteOwner)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(s1Value, s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	s1.Name = "hello"
	s1Value, s1Err = getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}
	err = o1.Update(s1Value, remoteOwner)
	if err != nil {
		t.Errorf("update reference failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(s1Value, s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	s2 := &Simple{ID: s1.ID}
	s2Value, s2Err := getObjectValue(s2)
	if s2Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s2Err.Error())
		return
	}
	err = o1.Query(s2Value, remoteOwner)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(s2Value, s2)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if !s1.IsSame(s2) {
		t.Errorf("Query reference failed.")
	}

}

func TestRemoteReference(t *testing.T) {
	orm.Initialize(50, "root", "rootkit", "localhost:3306", "testdb", false)
	defer orm.Uninitialize()

	o1, err := orm.New()
	defer o1.Release()
	if err != nil {
		t.Errorf("new Orm failed, err:%s", err.Error())
		return
	}

	ref := &Reference{}
	refDef, refErr := remote.GetObject(ref)
	if refErr != nil {
		t.Errorf("GetObject failed, err:%s", refErr.Error())
		return
	}

	objList := []interface{}{refDef}
	err = registerModel(o1, objList, remoteOwner)
	if err != nil {
		t.Errorf("register model failed. err:%s", err.Error())
		return
	}

	err = o1.Drop(refDef, remoteOwner)
	if err != nil {
		t.Errorf("drop reference schema failed, err:%s", err.Error())
		return
	}

	err = o1.Create(refDef, remoteOwner)
	if err != nil {
		t.Errorf("create reference schema failed, err:%s", err.Error())
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

	s1Value, s1Err := getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}
	err = o1.Insert(s1Value, remoteOwner)
	if err != nil {
		t.Errorf("insert reference failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(s1Value, s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}

	s1.Name = "hello"
	s1Value, s1Err = getObjectValue(s1)
	if s1Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s1Err.Error())
		return
	}
	err = o1.Update(s1Value, remoteOwner)
	if err != nil {
		t.Errorf("update reference failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(s1Value, s1)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
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
	s2Value, s2Err := getObjectValue(s2)
	if s2Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s2Err.Error())
		return
	}
	err = o1.Query(s2Value, remoteOwner)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(s2Value, s2)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
		return
	}
	if !s1.IsSame(s2) {
		t.Errorf("Query reference failed.")
		return
	}

	s4 := &Reference{
		ID: s1.ID,
	}
	s4Value, s4Err := getObjectValue(s4)
	if s4Err != nil {
		t.Errorf("getObjectValue failed, err:%s", s4Err.Error())
		return
	}
	err = o1.Query(s4Value, remoteOwner)
	if err != nil {
		t.Errorf("query reference failed, err:%s", err.Error())
		return
	}
	err = remote.UpdateEntity(s4Value, s4)
	if err != nil {
		t.Errorf("updateEntity failed, err:%s", err.Error())
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
