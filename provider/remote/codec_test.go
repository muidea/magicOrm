package remote

import (
	"fmt"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
)

func TestCodec(t *testing.T) {
	ii := 123
	iValueVal, iValueErr := GetEntityValue(ii)
	if iValueErr != nil {
		t.Errorf("GetEntityValue(ii) failed, error:%s", iValueErr.Error())
	}

	iTypeVal, iTypeErr := GetEntityType(ii)
	if iTypeErr != nil {
		t.Errorf("NewType failed, err:%s", iTypeErr.Error())
		return
	}

	valStr, valErr := _codec.Encode(iValueVal, iTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ii) {
		t.Errorf("Encode failed,")
		return
	}

	dVal, dErr := _codec.Decode(valStr, iTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if fmt.Sprintf("%v", dVal.Interface()) != fmt.Sprintf("%v", ii) {
		t.Errorf("Decode failed,")
		return
	}

	uii := uint16(123)
	uiValueVal, uiValueErr := GetEntityValue(uii)
	if uiValueErr != nil {
		t.Errorf("GetEntityValue(uii) failed, error:%s", uiValueErr.Error())
		return
	}
	uiTypeVal, uiTypeErr := GetEntityType(uii)
	if uiTypeErr != nil {
		t.Errorf("NewType failed, err:%s", uiTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(uiValueVal, uiTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", uii) {
		t.Errorf("Encode failed,")
		return
	}
	dVal, dErr = _codec.Decode(valStr, uiTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if fmt.Sprintf("%v", dVal.Interface()) != fmt.Sprintf("%v", uii) {
		t.Errorf("Decode failed,")
		return
	}

	ff := 123.345
	fValueVal, fValueErr := GetEntityValue(ff)
	if fValueErr != nil {
		t.Errorf("GetEntityValue(ff) failed, error:%s", fValueErr.Error())
		return
	}

	fTypeVal, fTypeErr := GetEntityType(ff)
	if fTypeErr != nil {
		t.Errorf("NewType failed, err:%s", fTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(fValueVal, fTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ff) {
		t.Errorf("Encode failed,")
		return
	}
	dVal, dErr = _codec.Decode(valStr, fTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if fmt.Sprintf("%v", dVal.Interface()) != fmt.Sprintf("%v", ff) {
		t.Errorf("Decode failed,")
		return
	}

	bb := true
	bValueVal, bValueErr := GetEntityValue(bb)
	if bValueErr != nil {
		t.Errorf("GetEntityValue(bb) failed, error:%s", bValueErr.Error())
		return
	}
	bTypeVal, bTypeErr := GetEntityType(bb)
	if bTypeErr != nil {
		t.Errorf("GetEntityType(bb) failed, err:%s", bTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(bValueVal, bTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", 1) {
		t.Errorf("Encode failed, valStr:%v", valStr)
		return
	}
	dVal, dErr = _codec.Decode(valStr, bTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if fmt.Sprintf("%v", dVal.Interface()) != fmt.Sprintf("%v", bb) {
		t.Errorf("Decode failed,")
		return
	}

	ss := "hello world"
	sValueVal, sValueErr := GetEntityValue(ss)
	if sValueErr != nil {
		t.Errorf("GetEntityValue(ss) failed, error:%s", sValueErr.Error())
		return
	}
	sTypeVal, sTypeErr := GetEntityType(ss)
	if sTypeErr != nil {
		t.Errorf("GetEntityType(ss) failed, error:%s", sTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(sValueVal, sTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("'%v'", ss) {
		t.Errorf("Encode failed,")
		return
	}
	dVal, dErr = _codec.Decode(valStr, sTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if fmt.Sprintf("%v", dVal.Interface()) != fmt.Sprintf("'%v'", ss) {
		t.Errorf("Decode failed,")
		return
	}

	tt := time.Now()
	tValueVal, tValueErr := GetEntityValue(tt)
	if tValueErr != nil {

	}
	tTypeValue, tTypeErr := GetEntityType(tt)
	if tTypeErr != nil {
		t.Errorf("NewType failed, err:%s", tTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(tValueVal, tTypeValue)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("'%s'", tt.Format(util.CSTLayout)) {
		t.Errorf("Encode failed,")
		return
	}
	dVal, dErr = _codec.Decode(valStr, tTypeValue)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	if dVal.Interface().Value().(string) != tt.Format(util.CSTLayout) {
		t.Errorf("Decode failed,")
		return
	}

}

func TestSliceCodec(t *testing.T) {
	ii := []int{123, 234}
	iValueVal, iValueErr := GetEntityValue(ii)
	if iValueErr != nil {
		t.Errorf("GetEntityValue(ii) failed, error:%s", iValueErr.Error())
		return
	}
	iTypeVal, iTypeErr := GetEntityType(ii)
	if iTypeErr != nil {
		t.Errorf("NewType failed, bTypeErr:%s", iTypeErr.Error())
		return
	}

	valStr, valErr := _codec.Encode(iValueVal, iTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, bTypeErr:%s", valErr.Error())
		return
	}
	fStr := fmt.Sprintf("%v", valStr)
	if fStr != "[123 234]" {
		t.Errorf("Encode failed,")
		return
	}

	dVal, dErr := _codec.Decode(valStr, iTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
	switch dVal.Interface().Value().(type) {
	case []int:
		t.Logf("%+v", dVal.Interface().Value())
	default:
		t.Errorf("Decode failed,")
	}
	i16 := []int16{123, 234}
	i16ValueVal, i16ValueErr := GetEntityValue(i16)
	if i16ValueErr != nil {
		t.Errorf("GetEntityValue(i16) failed, error:%s", i16ValueErr.Error())
		return
	}
	i16TypeVal, i16TypeErr := GetEntityType(i16)
	if i16TypeErr != nil {
		t.Errorf("NewType failed, bTypeErr:%s", i16TypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(i16ValueVal, i16TypeVal)
	if valErr != nil {
		t.Errorf("encode failed, bTypeErr:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != "[123 234]" {
		t.Errorf("Encode failed,")
		return
	}

	_, dErr = _codec.Decode(valStr, i16TypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}

	ff := []float64{123.345}
	fValueVal, fValueErr := GetEntityValue(ff)
	if fValueErr != nil {
		t.Errorf("GetEntityValue(ff) failed, error:%s", fValueErr.Error())
		return
	}
	fTypeVal, fTypeErr := GetEntityType(ff)
	if fTypeErr != nil {
		t.Errorf("NewType failed, bTypeErr:%s", fTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(fValueVal, fTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, bTypeErr:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != "[123.345]" {
		t.Errorf("Encode failed,")
		return
	}
	_, dErr = _codec.Decode(valStr, fTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}

	bb := []bool{true}
	bValueVal, bValueErr := GetEntityValue(bb)
	if bValueErr != nil {
		t.Errorf("GetEntityValue(bb) failed, error:%s", bValueErr.Error())
		return
	}
	bTypeVal, bTypeErr := GetEntityType(bb)
	if bTypeErr != nil {
		t.Errorf("GetEntityType(bb) failed, bTypeErr:%s", bTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(bValueVal, bTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, bTypeErr:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != "[1]" {
		t.Errorf("Encode failed,")
		return
	}
	_, dErr = _codec.Decode(valStr, bTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed, bTypeErr:%s", dErr.Error())
		return
	}

	ss := []string{"hello world"}
	sValueVal, sValueErr := GetEntityValue(ss)
	if sValueErr != nil {
		t.Errorf("GetEntityValue(ss) failed, error:%s", sValueErr.Error())
		return
	}
	sTypeVal, sTypeErr := GetEntityType(ss)
	if sTypeErr != nil {
		t.Errorf("GetEntityType(ss) failed, bTypeErr:%s", sTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(sValueVal, sTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, bTypeErr:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != "[hello world]" {
		t.Errorf("Encode failed,")
		return
	}
	_, dErr = _codec.Decode(valStr, sTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}

	tv, tvErr := time.Parse(util.CSTLayout, "2006-01-02 15:04:05")
	if tvErr != nil {
		t.Errorf("parse time failed, bTypeErr:%s", tvErr.Error())
		return
	}

	tt := []time.Time{tv, tv}
	tValueVal, tValueErr := GetEntityValue(tt)
	if tValueErr != nil {
		t.Errorf("GetEntityValue(tt) failed, error:%s", tValueErr.Error())
		return
	}
	tTypeVal, tTypeErr := GetEntityType(tt)
	if tTypeErr != nil {
		t.Errorf("GetEntityType(tt) failed, bTypeErr:%s", tTypeErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(tValueVal, tTypeVal)
	if valErr != nil {
		t.Errorf("encode failed, bTypeErr:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != "[2006-01-02 15:04:05 2006-01-02 15:04:05]" {
		t.Errorf("Encode failed,")
		return
	}
	_, dErr = _codec.Decode(valStr, tTypeVal)
	if dErr != nil {
		t.Errorf("Decode failed,")
		return
	}
}
