package local

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
	pu "github.com/muidea/magicOrm/provider/util"
)

func TestCodec(t *testing.T) {
	ii := 123
	iVal := pu.NewValue(reflect.ValueOf(ii))
	iType, iErr := newType(reflect.TypeOf(ii))
	if iErr != nil {
		t.Errorf("newType failed, err:%s", iErr.Error())
		return
	}

	valStr, valErr := _codec.Encode(iVal, iType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ii) {
		t.Errorf("Encode failed,")
		return
	}

	dVal, dErr := _codec.Decode(valStr, iType)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if dVal.Get().Int() != iVal.Get().Int() {
		t.Errorf("Decode failed,")
		return
	}

	uii := uint16(123)
	uiVal := pu.NewValue(reflect.ValueOf(uii))
	uiType, uiErr := newType(reflect.TypeOf(uii))
	if uiErr != nil {
		t.Errorf("newType failed, err:%s", uiErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(uiVal, uiType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", uii) {
		t.Errorf("Encode failed,")
		return
	}

	ff := 123.345
	fVal := pu.NewValue(reflect.ValueOf(ff))
	fType, fErr := newType(reflect.TypeOf(ff))
	if fErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(fVal, fType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ff) {
		t.Errorf("Encode failed,")
		return
	}

	bb := true
	bVal := pu.NewValue(reflect.ValueOf(bb))
	bType, bErr := newType(reflect.TypeOf(bb))
	if bErr != nil {
		t.Errorf("newType failed, err:%s", bErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(bVal, bType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != 1 {
		t.Errorf("Encode failed,")
		return
	}

	ss := "hello world"
	sVal := pu.NewValue(reflect.ValueOf(ss))
	sType, sErr := newType(reflect.TypeOf(ss))
	if sErr != nil {
		t.Errorf("newType failed, err:%s", sErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(sVal, sType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != ss {
		t.Errorf("Encode failed,")
		return
	}

	tt := time.Now()
	tVal := pu.NewValue(reflect.ValueOf(tt))
	tType, tErr := newType(reflect.TypeOf(tt))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != tt.Format("2006-01-02 15:04:05") {
		t.Errorf("Encode failed,")
		return
	}
}

func TestSliceCodec(t *testing.T) {
	ii := []int{123}
	iVal := pu.NewValue(reflect.ValueOf(ii))
	iType, iErr := newType(reflect.TypeOf(ii))
	if iErr != nil {
		t.Errorf("newType failed, err:%s", iErr.Error())
		return
	}

	valStr, valErr := _codec.Encode(iVal, iType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "123" {
		t.Errorf("Encode failed,")
		return
	}

	dVal, dErr := _codec.Decode(valStr, iType)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if reflect.Indirect(dVal.Get()).Len() != reflect.Indirect(iVal.Get()).Len() {
		t.Errorf("Decode failed,")
		return
	}

	ff := []float64{123.345}
	fVal := pu.NewValue(reflect.ValueOf(ff))
	fType, fErr := newType(reflect.TypeOf(ff))
	if fErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(fVal, fType)
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

	valStr, valErr = _codec.Encode(bVal, bType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "1" {
		t.Errorf("Encode failed,")
		return
	}
	dVal, dErr = _codec.Decode(valStr, bType)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if dVal.Get().Len() != bVal.Get().Len() {
		t.Errorf("Decode failed,")
		return
	}

	ss := []string{"hello world"}
	sVal := pu.NewValue(reflect.ValueOf(ss))
	sType, sErr := newType(reflect.TypeOf(ss))
	if sErr != nil {
		t.Errorf("newType failed, err:%s", sErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(sVal, sType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "hello world" {
		t.Errorf("Encode failed,")
		return
	}

	tv, tErr := time.Parse(util.CSTLayout, "2006-01-02 15:04:05")
	if tErr != nil {
		t.Errorf("parse time failed, err:%s", tErr.Error())
		return
	}

	tt := []time.Time{tv}
	tVal := pu.NewValue(reflect.ValueOf(tt))
	tType, tErr := newType(reflect.TypeOf(tt))
	if tErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if valStr != "2006-01-02 15:04:05" {
		t.Errorf("Encode failed,")
		return
	}
}
