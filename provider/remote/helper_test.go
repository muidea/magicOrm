package remote

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestHelper(t *testing.T) {
	ii := 123
	iVal := newValue(ii)
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

	vVal, vErr := _helper.Decode(valStr, iType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vii, vok := vVal.Get().(int)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vii != ii {
		t.Errorf("Decode failed")
		return
	}

	uii := uint16(123)
	uiVal := newValue(uii)
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

	vVal, vErr = _helper.Decode(valStr, uiType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vuii, vok := vVal.Get().(uint16)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vuii != uii {
		t.Errorf("Decode failed")
		return
	}

	ff := 123.345
	fVal := newValue(ff)
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

	vVal, vErr = _helper.Decode(valStr, fType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vff, vok := vVal.Get().(float64)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vff != ff {
		t.Errorf("Decode failed")
		return
	}

	bb := true
	bVal := newValue(bb)
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
	vVal, vErr = _helper.Decode(valStr, bType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vbb, vok := vVal.Get().(bool)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vbb != bb {
		t.Errorf("Decode failed")
		return
	}

	ss := "hello world"
	sVal := newValue(ss)
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
	vVal, vErr = _helper.Decode(valStr, sType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vss, vok := vVal.Get().(string)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vss != ss {
		t.Errorf("Decode failed")
		return
	}

	tt := time.Now()
	tVal := newValue("2006-01-02 15:04:05.000")
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
	if valStr != "2006-01-02 15:04:05.000" {
		t.Errorf("Encode failed,")
		return
	}

	vVal, vErr = _helper.Decode(valStr, sType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vtt, vok := vVal.Get().(string)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vtt != "2006-01-02 15:04:05.000" {
		t.Errorf("Decode failed")
		return
	}
}

func TestRemoteHelper(t *testing.T) {
	ii := 123
	iVal := newValue(float64(ii))
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
	vVal, vErr := _helper.Decode(valStr, iType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vii, vok := vVal.Get().(int)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vii != ii {
		t.Errorf("Decode failed")
		return
	}

	uii := uint16(123)
	uiVal := newValue(float64(uii))
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
	fVal := newValue(ff)
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
	bVal := newValue(bb)
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
	sVal := newValue(ss)
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
	tVal := newValue("2006-01-02 15:04:05.000")
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
	if valStr != "2006-01-02 15:04:05.000" {
		t.Errorf("Encode failed,")
		return
	}
}

func TestSliceHelper(t *testing.T) {
	ii := []int{123}
	iVal := newValue(ii)
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

	vVal, vErr := _helper.Decode(valStr, iType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vii, vok := vVal.Get().([]int)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if len(vii) != len(ii) {
		t.Errorf("Decode failed")
		return
	}

	ff := []float64{123.345}
	fVal := newValue(ff)
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
	bVal := newValue(bb)
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
	sVal := newValue(ss)
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

	tt := []string{"2006-01-02 15:04:05.000"}
	tVal := newValue(tt)
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
	if valStr != "[\"2006-01-02 15:04:05.000\"]" {
		t.Errorf("Encode failed,")
		return
	}
}

func TestRemoteSliceHelper(t *testing.T) {
	ii := []int{}
	iVal := newValue([]float64{123})
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
	fVal := newValue(ff)
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
	bVal := newValue(bb)
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
	sVal := newValue(ss)
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

	tt := []string{"2006-01-02 15:04:05.000"}
	tVal := newValue(tt)
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
	if valStr != "[\"2006-01-02 15:04:05.000\"]" {
		t.Errorf("Encode failed,")
		return
	}
}
