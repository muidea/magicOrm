package remote

import (
	"fmt"
	"reflect"
	"testing"

	pu "github.com/muidea/magicOrm/provider/util"
)

func TestHelper(t *testing.T) {
	/*
		float64
		bool,
		string
	*/
	ii := 123.00
	iVal := pu.NewValue(reflect.ValueOf(ii))
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
	vii, vok := vVal.Interface().(float64)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vii != ii {
		t.Errorf("Decode failed")
		return
	}

	ff := 123.345
	fVal := pu.NewValue(reflect.ValueOf(ff))
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
	vff, vok := vVal.Interface().(float64)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vff != ff {
		t.Errorf("Decode failed")
		return
	}

	bb := true
	bVal := pu.NewValue(reflect.ValueOf(bb))
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
	vbb, vok := vVal.Interface().(bool)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vbb != bb {
		t.Errorf("Decode failed")
		return
	}

	ss := "hello world"
	sVal := pu.NewValue(reflect.ValueOf(ss))
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
	vss, vok := vVal.Interface().(string)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if vss != ss {
		t.Errorf("Decode failed")
		return
	}
}

func TestSliceHelper(t *testing.T) {
	ii := &[]float64{123}
	iVal := pu.NewValue(reflect.ValueOf(ii))
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
	vii, vok := vVal.Interface().(*[]float64)
	if !vok {
		t.Errorf("Decode failed")
		return
	}
	if len(*vii) != len(*ii) {
		t.Errorf("Decode failed")
		return
	}

	ff := []float64{123.345}
	fVal := pu.NewValue(reflect.ValueOf(ff))
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
	bVal := pu.NewValue(reflect.ValueOf(bb))
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

	bb = []bool{true, false, true}
	bVal = pu.NewValue(reflect.ValueOf(bb))
	bType, bErr = newType(reflect.TypeOf(bb))
	if bErr != nil {
		t.Errorf("newType failed, err:%s", bErr.Error())
		return
	}

	valStr, valErr = _helper.Encode(bVal, bType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "[1,0,1]" {
		t.Errorf("Encode failed,")
		return
	}

	ss := []string{"hello world"}
	sVal := pu.NewValue(reflect.ValueOf(ss))
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
	tVal := pu.NewValue(reflect.ValueOf(tt))
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
	ii := []float64{}
	iVal := pu.NewValue(reflect.ValueOf([]float64{123}))
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
	fVal := pu.NewValue(reflect.ValueOf(ff))
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
	bVal := pu.NewValue(reflect.ValueOf(bb))
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

	ss := []string{"hello world", "hello world"}
	sVal := pu.NewValue(reflect.ValueOf(ss))
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
	if valStr != "[\"hello world\",\"hello world\"]" {
		t.Errorf("Encode failed,")
		return
	}

	tt := []string{"2006-01-02 15:04:05.000"}
	tVal := pu.NewValue(reflect.ValueOf(tt))
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
	v := float64(12)
	ii := []float64{v}
	tVal := pu.NewValue(reflect.ValueOf(ii))
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
	_, ok := dVal.Get().Interface().([]float64)
	if !ok {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}

	iPtr := []*float64{&v}
	tVal = pu.NewValue(reflect.ValueOf(iPtr))
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
	_, ok = dVal.Get().Interface().([]*float64)
	if !ok {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}

	ptrii := &[]float64{v}
	tVal = pu.NewValue(reflect.ValueOf(ptrii))
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
	_, ok = dVal.Get().Interface().(*[]float64)
	if !ok {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}

	ptriiPtr := &[]*float64{&v}
	tVal = pu.NewValue(reflect.ValueOf(ptriiPtr))
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
	_, ok = dVal.Get().Interface().(*[]*float64)
	if !ok {
		t.Errorf("decode failed, err:%s", dErr.Error())
		return
	}
}
