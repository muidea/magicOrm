package remote

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
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ii) {
		t.Errorf("Encode failed,")
		return
	}

	vVal, vErr := _helper.Decode(valStr, iType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vii, vok := vVal.Get().Interface().(int)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vii != ii {
		t.Errorf("Decode failed")
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
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ii) {
		t.Errorf("Encode failed,")
		return
	}

	vVal, vErr = _helper.Decode(valStr, uiType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vuii, vok := vVal.Get().Interface().(uint16)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vuii != uii {
		t.Errorf("Decode failed")
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
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ff) {
		t.Errorf("Encode failed,")
		return
	}

	vVal, vErr = _helper.Decode(valStr, fType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vff, vok := vVal.Get().Interface().(float64)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vff != ff {
		t.Errorf("Decode failed")
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
	if fmt.Sprintf("%v", valStr) != "1" {
		t.Errorf("Encode failed,")
		return
	}
	vVal, vErr = _helper.Decode(valStr, bType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vbb, vok := vVal.Get().Interface().(bool)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vbb != bb {
		t.Errorf("Decode failed")
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
	if valStr.(string) != ss {
		t.Errorf("Encode failed,")
		return
	}
	vVal, vErr = _helper.Decode(valStr, sType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vss, vok := vVal.Get().Interface().(string)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vss != ss {
		t.Errorf("Decode failed")
		return
	}

	tt := time.Now()
	tVal := newValue(reflect.ValueOf("2006-01-02 15:04:05.000"))
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
	if valStr.(string) != "2006-01-02 15:04:05.000" {
		t.Errorf("Encode failed,")
		return
	}

	vVal, vErr = _helper.Decode(valStr, sType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vtt, vok := vVal.Get().Interface().(string)
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
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%d", ii) {
		t.Errorf("Encode failed,")
		return
	}
	vVal, vErr := _helper.Decode(valStr, iType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vii, vok := vVal.Get().Interface().(int)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vii != ii {
		t.Errorf("Decode failed")
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
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%d", ii) {
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
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ff) {
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
	if fmt.Sprintf("%v", valStr) != "1" {
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
	tVal := newValue(reflect.ValueOf("2006-01-02 15:04:05.000"))
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
	ii := &[]int{123}
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
	if valStr != "123" {
		t.Errorf("Encode failed,")
		return
	}

	vVal, vErr := _helper.Decode(valStr, iType)
	if vErr != nil {
		t.Errorf("Decode failed, err:%s", vErr.Error())
		return
	}
	vii, vok := vVal.Get().Interface().(*[]int)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if len(*vii) != len(*ii) {
		t.Errorf("Decode failed")
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
	if valStr != "123.345" {
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
	if valStr != "1" {
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
	if valStr != "hello world" {
		t.Errorf("Encode failed,")
		return
	}

	tt := []string{"2006-01-02 15:04:05.000"}
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
	if valStr != "2006-01-02 15:04:05.000" {
		t.Errorf("Encode failed,")
		return
	}
}

func TestRemoteSliceHelper(t *testing.T) {
	ii := []int{}
	iVal := newValue(reflect.ValueOf([]int{123}))
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
	if valStr != "123" {
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
	if valStr != "123.345" {
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
	if valStr != "1" {
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
	if valStr != "hello world" {
		t.Errorf("Encode failed,")
		return
	}

	tt := []string{"2006-01-02 15:04:05.000"}
	tVal := newValue(reflect.ValueOf(tt))
	tType, tErr := newType(reflect.TypeOf(tt))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", tErr.Error())
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

func TestSpecSliceHelper(t *testing.T) {
	v := 12
	ii := []int{v}
	tVal := newValue(reflect.ValueOf(ii))
	tType, tErr := newType(reflect.TypeOf(ii))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", tErr.Error())
		return
	}
	valStr, valErr := _helper.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "12" {
		t.Errorf("Encode failed,")
		return
	}

	dVal, dErr := _helper.Decode(valStr, tType)
	if dErr != nil {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}
	_, ok := dVal.Get().Interface().([]int)
	if !ok {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}

	iPtr := []*int{&v}
	tVal = newValue(reflect.ValueOf(iPtr))
	tType, tErr = newType(reflect.TypeOf(iPtr))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", tErr.Error())
		return
	}
	valStr, valErr = _helper.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "12" {
		t.Errorf("Encode failed,")
		return
	}

	dVal, dErr = _helper.Decode(valStr, tType)
	if dErr != nil {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}
	_, ok = dVal.Get().Interface().([]*int)
	if !ok {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}

	ptrii := &[]int{v}
	tVal = newValue(reflect.ValueOf(ptrii))
	tType, tErr = newType(reflect.TypeOf(ptrii))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", tErr.Error())
		return
	}
	valStr, valErr = _helper.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "12" {
		t.Errorf("Encode failed,")
		return
	}

	dVal, dErr = _helper.Decode(valStr, tType)
	if dErr != nil {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}
	_, ok = dVal.Get().Interface().(*[]int)
	if !ok {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}

	ptriiPtr := &[]*int{&v}
	tVal = newValue(reflect.ValueOf(ptriiPtr))
	tType, tErr = newType(reflect.TypeOf(ptriiPtr))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", tErr.Error())
		return
	}
	valStr, valErr = _helper.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "12" {
		t.Errorf("Encode failed,")
		return
	}

	dVal, dErr = _helper.Decode(valStr, tType)
	if dErr != nil {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}
	_, ok = dVal.Get().Interface().(*[]*int)
	if !ok {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}
}
