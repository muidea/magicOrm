package helper

import (
	"reflect"
	"testing"
)

func TestAssignValue(t *testing.T) {
	// int -> bool
	iVal := 10
	var bVal bool
	var fVal float64
	var strVal string

	iReflectVal := reflect.ValueOf(iVal)
	bReflectVal := reflect.ValueOf(&bVal).Elem()
	fReflectVal := reflect.ValueOf(&fVal).Elem()
	strReflectVal := reflect.ValueOf(&strVal).Elem()

	bReflectVal, err := AssignValue(iReflectVal, bReflectVal)
	if err != nil {
		t.Errorf("assign int to bool failed,err:%s", err.Error())
		return
	}
	if !bReflectVal.Bool() {
		t.Errorf("assign int to bool failed, convert unexpect")
		return
	}

	fReflectVal, err = AssignValue(iReflectVal, fReflectVal)
	if err != nil {
		t.Errorf("assign int to float failed,err:%s", err.Error())
		return
	}
	if fReflectVal.Float() != 10.00 {
		t.Errorf("assign int to float failed, convert unexpect")
		return
	}
	strReflectVal, err = AssignValue(iReflectVal, strReflectVal)
	if err != nil {
		t.Errorf("assign int to string failed,err:%s", err.Error())
		return
	}
	if strReflectVal.String() != "10" {
		t.Errorf("assign int to string failed, convert unexpect")
		return
	}

	iVal = 0
	iReflectVal = reflect.ValueOf(iVal)
	bReflectVal = reflect.ValueOf(&bVal).Elem()
	fReflectVal = reflect.ValueOf(&fVal).Elem()
	strReflectVal = reflect.ValueOf(&strVal).Elem()
	bReflectVal, err = AssignValue(iReflectVal, bReflectVal)
	if err != nil {
		t.Errorf("assign int to bool failed,err:%s", err.Error())
		return
	}
	if bReflectVal.Bool() {
		t.Errorf("assign int to bool failed, convert unexpect")
		return
	}
	fReflectVal, err = AssignValue(iReflectVal, fReflectVal)
	if err != nil {
		t.Errorf("assign int to float failed,err:%s", err.Error())
		return
	}
	if fReflectVal.Float() == 10.00 {
		t.Errorf("assign int to float failed, convert unexpect")
		return
	}
	strReflectVal, err = AssignValue(iReflectVal, strReflectVal)
	if err != nil {
		t.Errorf("assign int to string failed,err:%s", err.Error())
		return
	}
	if strReflectVal.String() != "0" {
		t.Errorf("assign int to string failed, convert unexpect")
		return
	}

	bValPtr := &bVal
	fValPtr := &fVal
	strValPtr := &strVal
	iVal = 1234
	iReflectVal = reflect.ValueOf(iVal)
	bReflectVal = reflect.ValueOf(&bValPtr).Elem()
	fReflectVal = reflect.ValueOf(&fValPtr).Elem()
	strReflectVal = reflect.ValueOf(&strValPtr).Elem()
	bReflectVal, err = AssignValue(iReflectVal, bReflectVal)
	if err != nil {
		t.Errorf("assign int to bool failed,err:%s", err.Error())
		return
	}
	if !bReflectVal.Elem().Bool() {
		t.Errorf("assign int to bool failed, convert unexpect")
		return
	}
	fReflectVal, err = AssignValue(iReflectVal, fReflectVal)
	if err != nil {
		t.Errorf("assign int to float failed,err:%s", err.Error())
		return
	}
	if fReflectVal.Elem().Float() != 1234.00 {
		t.Errorf("assign int to float failed, convert unexpect")
		return
	}
	strReflectVal, err = AssignValue(iReflectVal, strReflectVal)
	if err != nil {
		t.Errorf("assign int to string failed,err:%s", err.Error())
		return
	}
	if strReflectVal.Elem().String() != "1234" {
		t.Errorf("assign int to string failed, convert unexpect")
		return
	}
}
