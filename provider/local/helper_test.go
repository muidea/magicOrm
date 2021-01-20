package local

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestHelper(t *testing.T) {
	ii := 123
	iVal := newValue(reflect.ValueOf(ii))
	iType, iErr := newType(reflect.TypeOf(ii))
	if iErr != nil {
		t.Errorf("newType failed, err:%s", iErr.Error())
		return
	}

	valStr, valErr := _helper.Encode(iVal, iType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != fmt.Sprintf("%d", ii) {
		t.Errorf("Encode failed,")
		return
	}

	uii := uint16(123)
	uiVal := newValue(reflect.ValueOf(uii))
	uiType, uiErr := newType(reflect.TypeOf(uii))
	if uiErr != nil {
		t.Errorf("newType failed, err:%s", uiErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(uiVal, uiType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != fmt.Sprintf("%d", ii) {
		t.Errorf("Encode failed,")
		return
	}

	ff := 123.345
	fVal := newValue(reflect.ValueOf(ff))
	fType, fErr := newType(reflect.TypeOf(ff))
	if fErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(fVal, fType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != fmt.Sprintf("%f", ff) {
		t.Errorf("Encode failed,")
		return
	}

	bb := true
	bVal := newValue(reflect.ValueOf(bb))
	bType, bErr := newType(reflect.TypeOf(bb))
	if bErr != nil {
		t.Errorf("newType failed, err:%s", bErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(bVal, bType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "1" {
		t.Errorf("Encode failed,")
		return
	}

	ss := "hello world"
	sVal := newValue(reflect.ValueOf(ss))
	sType, sErr := newType(reflect.TypeOf(ss))
	if sErr != nil {
		t.Errorf("newType failed, err:%s", sErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(sVal, sType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != ss {
		t.Errorf("Encode failed,")
		return
	}

	tt := time.Now()
	tVal := newValue(reflect.ValueOf(tt))
	tType, tErr := newType(reflect.TypeOf(tt))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != tt.Format("2006-01-02 15:04:05") {
		t.Errorf("Encode failed,")
		return
	}
}

func TestSliceHelper(t *testing.T) {
	ii := []int{123}
	iVal := newValue(reflect.ValueOf(ii))
	iType, iErr := newType(reflect.TypeOf(ii))
	if iErr != nil {
		t.Errorf("newType failed, err:%s", iErr.Error())
		return
	}

	valStr, valErr := _helper.Encode(iVal, iType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "[\"123\"]" {
		t.Errorf("Encode failed,")
		return
	}

	ff := []float64{123.345}
	fVal := newValue(reflect.ValueOf(ff))
	fType, fErr := newType(reflect.TypeOf(ff))
	if fErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(fVal, fType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "[\"123.345000\"]" {
		t.Errorf("Encode failed,")
		return
	}

	bb := []bool{true}
	bVal := newValue(reflect.ValueOf(bb))
	bType, bErr := newType(reflect.TypeOf(bb))
	if bErr != nil {
		t.Errorf("newType failed, err:%s", bErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(bVal, bType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "[\"1\"]" {
		t.Errorf("Encode failed,")
		return
	}

	ss := []string{"hello world"}
	sVal := newValue(reflect.ValueOf(ss))
	sType, sErr := newType(reflect.TypeOf(ss))
	if sErr != nil {
		t.Errorf("newType failed, err:%s", sErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(sVal, sType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "[\"hello world\"]" {
		t.Errorf("Encode failed,")
		return
	}

	tv, tErr := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
	if tErr != nil {
		t.Errorf("parse time failed, err:%s", tErr.Error())
		return
	}

	tt := []time.Time{tv}
	tVal := newValue(reflect.ValueOf(tt))
	tType, tErr := newType(reflect.TypeOf(tt))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "[\"2006-01-02 15:04:05\"]" {
		t.Errorf("Encode failed,")
		return
	}
}
