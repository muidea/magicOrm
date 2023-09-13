package local

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
)

func TestCodec(t *testing.T) {
	ii := 123
	iVal := NewValue(reflect.ValueOf(ii))
	iType, iErr := NewType(reflect.TypeOf(ii))
	if iErr != nil {
		t.Errorf("NewType failed, err:%s", iErr.Error())
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
	if fmt.Sprintf("%v", dVal.Get()) != fmt.Sprintf("%v", ii) {
		t.Errorf("Decode failed,")
		return
	}

	uii := uint16(123)
	uiVal := NewValue(reflect.ValueOf(uii))
	uiType, uiErr := NewType(reflect.TypeOf(uii))
	if uiErr != nil {
		t.Errorf("NewType failed, err:%s", uiErr.Error())
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
	fVal := NewValue(reflect.ValueOf(ff))
	fType, fErr := NewType(reflect.TypeOf(ff))
	if fErr != nil {
		t.Errorf("NewType failed, err:%s", fErr.Error())
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
	val := NewValue(reflect.ValueOf(bb))
	bType, err := NewType(reflect.TypeOf(bb))
	if err != nil {
		t.Errorf("NewType failed, err:%s", err.Error())
		return
	}

	valStr, valErr = _codec.Encode(val, bType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", 1) {
		t.Errorf("Encode failed, valStr:%v", valStr)
		return
	}

	ss := "hello world"
	sVal := NewValue(reflect.ValueOf(ss))
	sType, sErr := NewType(reflect.TypeOf(ss))
	if sErr != nil {
		t.Errorf("NewType failed, err:%s", sErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(sVal, sType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", ss) {
		t.Errorf("Encode failed,")
		return
	}

	tt := time.Now()
	tVal := NewValue(reflect.ValueOf(tt))
	tType, tErr := NewType(reflect.TypeOf(tt))
	if tErr != nil {
		t.Errorf("NewType failed, err:%s", fErr.Error())
		return
	}

	valStr, valErr = _codec.Encode(tVal, tType)
	if valErr != nil {
		t.Errorf("encode failed, err:%s", valErr.Error())
		return
	}
	if fmt.Sprintf("%v", valStr) != fmt.Sprintf("%v", tt.Format(util.CSTLayout)) {
		t.Errorf("Encode failed,")
		return
	}
}

func TestSliceCodec(t *testing.T) {
	ii := []int{123}
	iVal := NewValue(reflect.ValueOf(ii))
	iType, iErr := NewType(reflect.TypeOf(ii))
	if iErr != nil {
		t.Errorf("NewType failed, err:%s", iErr.Error())
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
	if reflect.Indirect(dVal.Get().(reflect.Value)).Len() != reflect.Indirect(iVal.Get().(reflect.Value)).Len() {
		t.Errorf("Decode failed,")
		return
	}

	ff := []float64{123.345}
	fVal := NewValue(reflect.ValueOf(ff))
	fType, fErr := NewType(reflect.TypeOf(ff))
	if fErr != nil {
		t.Errorf("NewType failed, err:%s", fErr.Error())
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
	val := NewValue(reflect.ValueOf(bb))
	bType, err := NewType(reflect.TypeOf(bb))
	if err != nil {
		t.Errorf("NewType failed, err:%s", err.Error())
		return
	}

	valStr, valErr = _codec.Encode(val, bType)
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
		t.Errorf("Decode failed, err:%s", dErr.Error())
		return
	}
	if dVal.Get().(reflect.Value).Len() != val.Get().(reflect.Value).Len() {
		t.Errorf("Decode failed,")
		return
	}

	ss := []string{"hello world"}
	sVal := NewValue(reflect.ValueOf(ss))
	sType, sErr := NewType(reflect.TypeOf(ss))
	if sErr != nil {
		t.Errorf("NewType failed, err:%s", sErr.Error())
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
	tVal := NewValue(reflect.ValueOf(tt))
	tType, tErr := NewType(reflect.TypeOf(tt))
	if tErr != nil {
		t.Errorf("NewType failed, err:%s", fErr.Error())
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

func TestBaseBoolCodec(t *testing.T) {
	val := false
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new bool type failed, err:%s", err.Error())
		return
	}

	// int8(0)
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode bool false, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int8:
		if eVal.(int8) != 0 {
			t.Errorf("encode false failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode false failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode bool false, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode bool false, is not nil")
		return
	}
	if !dVal.IsZero() {
		t.Errorf("decode bool false, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode bool false, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Bool() {
		t.Errorf("decode bool false, is false")
		return
	}

	val = true
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new bool type failed, err:%s", err.Error())
		return
	}

	// int8(1)
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode bool false, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int8:
		if eVal.(int8) != 1 {
			t.Errorf("encode false failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode false failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode bool false, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode bool false, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode bool false, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode bool false, is basic")
		return
	}
	if !dVal.Get().(reflect.Value).Bool() {
		t.Errorf("decode bool false, is true")
		return
	}
}

func TestBaseBoolPtrCodec(t *testing.T) {
	valFalse := false
	var val *bool = &valFalse
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new bool type failed, err:%s", err.Error())
		return
	}

	// int8(0)
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode bool false, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int8:
		if eVal.(int8) != 0 {
			t.Errorf("encode false failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode false failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode bool false, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode bool false, is not nil")
		return
	}
	if !dVal.IsZero() {
		t.Errorf("decode bool false, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode bool false, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Kind() != reflect.Ptr {
		t.Errorf("decode bool false, is false")
		return
	}

	if reflect.Indirect(dVal.Get().(reflect.Value)).Bool() {
		t.Errorf("decode bool false, is false")
		return
	}

	varTrue := true
	val = &varTrue
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new bool type failed, err:%s", err.Error())
		return
	}

	// int8(1)
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode bool false, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int8:
		if eVal.(int8) != 1 {
			t.Errorf("encode false failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode false failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode bool false, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode bool false, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode bool false, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode bool false, is basic")
		return
	}
	if !reflect.Indirect(dVal.Get().(reflect.Value)).Bool() {
		t.Errorf("decode bool false, is true")
		return
	}
}

func TestSliceBoolCodec(t *testing.T) {
	val := []bool{false}
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new []bool type failed, err:%s", err.Error())
		return
	}

	// string("0")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode []bool{false}, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "0" {
			t.Errorf("encode []bool{false} failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []bool{false} failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode []bool{false}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode []bool{false}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode []bool{false}, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode []bool{false}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 1 {
		t.Errorf("decode []bool{false}, is false")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Bool() != false {
		t.Errorf("decode []bool{false}, is false")
		return
	}

	val = []bool{true, false, true}
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new []bool type failed, err:%s", err.Error())
		return
	}

	// string(["1","0","1"])
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode []bool{true,false,true}, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "[1,0,1]" {
			t.Errorf("encode []bool{true,false,true} failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []bool{true,false,true} failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode  []bool{true,false,true}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode  []bool{true,false,true}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode  []bool{true,false,true}, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []bool{true,false,true}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 3 {
		t.Errorf("decode  []bool{true,false,true}, is true")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Bool() != true {
		t.Errorf("decode  []bool{true,false,true}, is true")
		return
	}
	if dVal.Get().(reflect.Value).Index(1).Bool() != false {
		t.Errorf("decode  []bool{true,false,true}, is true")
		return
	}
}

func TestSliceBoolPtrCodec(t *testing.T) {
	valFalse := false
	val := []*bool{&valFalse}
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new []bool type failed, err:%s", err.Error())
		return
	}

	// string("0")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode []bool{false}, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "0" {
			t.Errorf("encode []bool{false} failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []bool{false} failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode []bool{false}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode []bool{false}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode []bool{false}, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode []bool{false}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 1 {
		t.Errorf("decode []bool{false}, is false")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Kind() != reflect.Ptr {
		t.Errorf("decode []bool{false}, is false")
		return
	}
	if reflect.Indirect(dVal.Get().(reflect.Value).Index(0)).Bool() != false {
		t.Errorf("decode []bool{false}, is false")
		return
	}

	valTrue := true
	val = []*bool{&valTrue, &valFalse, &valTrue}
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new []bool type failed, err:%s", err.Error())
		return
	}

	// string(["1","0","1"])
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode []bool{true,false,true}, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "[1,0,1]" {
			t.Errorf("encode []bool{true,false,true} failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []bool{true,false,true} failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode  []bool{true,false,true}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode  []bool{true,false,true}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode  []bool{true,false,true}, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []bool{true,false,true}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 3 {
		t.Errorf("decode  []bool{true,false,true}, is true")
		return
	}
	if reflect.Indirect(dVal.Get().(reflect.Value).Index(0)).Bool() != true {
		t.Errorf("decode  []bool{true,false,true}, is true")
		return
	}
	if reflect.Indirect(dVal.Get().(reflect.Value).Index(1)).Bool() != false {
		t.Errorf("decode  []bool{true,false,true}, is true")
		return
	}
}

func TestBaseIntCodec(t *testing.T) {
	val := int16(0)
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new int16 type failed, err:%s", err.Error())
		return
	}

	// int64(0)
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode int16 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int64:
		if eVal.(int64) != 0 {
			t.Errorf("encode int16 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode 0 failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode int16 0, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode int16 0, is not nil")
		return
	}
	if !dVal.IsZero() {
		t.Errorf("decode int16 0, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode int16 0, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Int() != 0 {
		t.Errorf("decode int16 0, is 0")
		return
	}

	val = int16(123)
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new int16 type failed, err:%s", err.Error())
		return
	}

	// int64(123)
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode int16 123, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int64:
		if eVal.(int64) != 123 {
			t.Errorf("encode int16 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode int16 failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode int16 123, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode int16 123, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode int16 123, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode int16 123, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Int() != 123 {
		t.Errorf("decode int16 123, is 123")
		return
	}
}

func TestBaseIntPtrCodec(t *testing.T) {
	i0Val := int16(0)
	var val = &i0Val
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new int16 type failed, err:%s", err.Error())
		return
	}

	// int64(0)
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode int16 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int64:
		if eVal.(int64) != 0 {
			t.Errorf("encode int16 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode 0 failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode int16 0, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode int16 0, is not nil")
		return
	}
	if !dVal.IsZero() {
		t.Errorf("decode int16 0, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode int16 0, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Kind() != reflect.Ptr {
		t.Errorf("decode int16 123, is basic")
		return
	}
	if reflect.Indirect(dVal.Get().(reflect.Value)).Int() != 0 {
		t.Errorf("decode int16 0, is 0")
		return
	}

	i123Val := int16(123)
	val = &i123Val
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new int16 type failed, err:%s", err.Error())
		return
	}

	// int64(123)
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode int16 123, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case int64:
		if eVal.(int64) != 123 {
			t.Errorf("encode int16 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode int16 failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode int16 123, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode int16 123, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode int16 123, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode int16 123, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Kind() != reflect.Ptr {
		t.Errorf("decode int16 123, is basic")
		return
	}
	if reflect.Indirect(dVal.Get().(reflect.Value)).Int() != 123 {
		t.Errorf("decode int16 123, is 123")
		return
	}
}

func TestSliceIntCodec(t *testing.T) {
	val := []int16{123}
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new  []int16 type failed, err:%s", err.Error())
		return
	}

	// string("123")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode  []int16 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "123" {
			t.Errorf("encode  []int16 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []int16{123} failed, val:%v", eVal)
		return
	}

	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode  []int16{123}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode  []int16{123}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode  []int16{123}, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []int16{123}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 1 {
		t.Errorf("decode  []int16{123}, is 0")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Int() != 123 {
		t.Errorf("decode  []int16{123}, is 0")
		return
	}

	val = []int16{123, 456, -789}
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new []int16 type failed, err:%s", err.Error())
		return
	}

	// string([123,456,-789])
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode []int16{123,456,-789}, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "[123,456,-789]" {
			t.Errorf("encode []int16{123,456,-789} failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []int16{123,456,-789} failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode []int16{123,456,-789}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode []int16{123,456,-789}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode []int16{123,456,-789}, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []int16{123,456,-789}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 3 {
		t.Errorf("decode  []int16{123,456,-789}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Int() != 123 {
		t.Errorf("decode  []int16{123,456,-789}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(1).Int() != 456 {
		t.Errorf("decode  []int16{123,456,-789}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(2).Int() != -789 {
		t.Errorf("decode  []int16{123,456,-789}, is 123")
		return
	}
}

func TestBaseUIntCodec(t *testing.T) {
	val := uint32(0)
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new uint32 type failed, err:%s", err.Error())
		return
	}

	// uint64(0)
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode uint32 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case uint64:
		if eVal.(uint64) != 0 {
			t.Errorf("encode uint32 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode 0 failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode uint32 0, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode uint32 0, is not nil")
		return
	}
	if !dVal.IsZero() {
		t.Errorf("decode uint32 0, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode uint32 0, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Uint() != 0 {
		t.Errorf("decode uint32 0, is 0")
		return
	}

	val = uint32(123)
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new uint32 type failed, err:%s", err.Error())
		return
	}

	// uint64(123)
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode uint32 123, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case uint64:
		if eVal.(uint64) != 123 {
			t.Errorf("encode uint32 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode uint32 failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode uint32 123, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode uint32 123, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode uint32 123, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode int16 123, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Uint() != 123 {
		t.Errorf("decode uint32 123, is 123")
		return
	}
}

func TestSliceUIntCodec(t *testing.T) {
	val := []uint32{123}
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new  []uint32 type failed, err:%s", err.Error())
		return
	}

	// string("123")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode  []uint32 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "123" {
			t.Errorf("encode  []uint32 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []uint32{123} failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode  []uint32{123}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode  []uint32{123}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode  []uint32{123}, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []uint32{123}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 1 {
		t.Errorf("decode  []uint32{123}, is 0")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Uint() != 123 {
		t.Errorf("decode  []uint32{123}, is 0")
		return
	}

	val = []uint32{123, 456, 789000000}
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new []uint32 type failed, err:%s", err.Error())
		return
	}

	// string([123,456,789000000])
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode []uint32{123,456,789000000}, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "[123,456,789000000]" {
			t.Errorf("encode []uint32{123,456,789000000} failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []uint32{123,456,789000000} failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode []uint32{123,456,789000000}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode []uint32{123,456,789000000}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode []uint32{123,456,789000000}, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []uint32{123,456,789000000}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 3 {
		t.Errorf("decode  []uint32{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Uint() != 123 {
		t.Errorf("decode  []uint32{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(1).Uint() != 456 {
		t.Errorf("decode  []uint32{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(2).Uint() != 789000000 {
		t.Errorf("decode  []uint32{123,456,789000000}, is 123")
		return
	}
}

func TestBaseFloatCodec(t *testing.T) {
	val := float32(0)
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new float32 type failed, err:%s", err.Error())
		return
	}

	// float64(0)
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode float32 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case float64:
		if eVal.(float64) != 0 {
			t.Errorf("encode float32 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode 0 failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode float32 0, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode float32 0, is not nil")
		return
	}
	if !dVal.IsZero() {
		t.Errorf("decode float32 0, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode float32 0, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Float() != 0 {
		t.Errorf("decode float32 0, is 0")
		return
	}

	val = float32(123.456)
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new float32 type failed, err:%s", err.Error())
		return
	}

	// float64(123.456)
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode float32 123.456, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case float64:
		if math.Abs(eVal.(float64)-123.456) > 0.001 {
			t.Errorf("encode float32 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode float32 failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode float32 123.456, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode float32 123.456, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode float32 123.456, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode float32 123.456, is basic")
		return
	}
	if math.Abs(dVal.Get().(reflect.Value).Float()-123.456) > 0.001 {
		t.Errorf("decode float32 123.456, is 123.456")
		return
	}
}

func TestSliceFloatCodec(t *testing.T) {
	val := []float32{123}
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new  []float32 type failed, err:%s", err.Error())
		return
	}

	// string("123")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode  []float32 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "123" {
			t.Errorf("encode  []float32 failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []float32{123} failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode  []float32{123}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode  []float32{123}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode  []float32{123}, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []float32{123}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 1 {
		t.Errorf("decode  []float32{123}, is 0")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Float() != 123 {
		t.Errorf("decode  []float32{123}, is 0")
		return
	}

	val = []float32{123, 456, 789000000}
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new []float32 type failed, err:%s", err.Error())
		return
	}

	// string([123,456,789000000])
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode []float32{123,456,789000000}, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "[123,456,789000000]" {
			t.Errorf("encode []float32{123,456,789000000} failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []float32{123,456,789000000} failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode []float32{123,456,789000000}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode []float32{123,456,789000000}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode []float32{123,456,789000000}, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []float32{123,456,789000000}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 3 {
		t.Errorf("decode  []float32{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).Float() != 123 {
		t.Errorf("decode  []float32{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(1).Float() != 456 {
		t.Errorf("decode  []float32{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(2).Float() != 789000000 {
		t.Errorf("decode  []float32{123,456,789000000}, is 123")
		return
	}
}

func TestBaseStringCodec(t *testing.T) {
	val := ""
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new string type failed, err:%s", err.Error())
		return
	}

	// string(")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode string 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "" {
			t.Errorf("encode string failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode 0 failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode string 0, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode string 0, is not nil")
		return
	}
	if !dVal.IsZero() {
		t.Errorf("decode string 0, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode string 0, is basic")
		return
	}
	if dVal.Get().(reflect.Value).String() != "" {
		t.Errorf("decode string '', is ''")
		return
	}

	val = string("123.456")
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new string type failed, err:%s", err.Error())
		return
	}

	// string("123.456")
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode string 123.456, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "123.456" {
			t.Errorf("encode string failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode string failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode string 123.456, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode string 123.456, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode string 123.456, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode string 123.456, is basic")
		return
	}
	if dVal.Get().(reflect.Value).String() != "123.456" {
		t.Errorf("decode string '123.456', is '123.456'")
		return
	}
}

func TestSliceStringCodec(t *testing.T) {
	val := []string{"123"}
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new  []string type failed, err:%s", err.Error())
		return
	}

	// string("123")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode  []string 0, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "123" {
			t.Errorf("encode  []string failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []string{123} failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode  []string{123}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode  []string{123}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode  []string{123}, is zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []string{123}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 1 {
		t.Errorf("decode  []string{123}, is 0")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).String() != "123" {
		t.Errorf("decode  []string{123}, is 0")
		return
	}

	val = []string{"123", "456", "789000000"}
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new []string type failed, err:%s", err.Error())
		return
	}

	// string([123,456,789000000])
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode []string{123,456,789000000}, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "[\"123\",\"456\",\"789000000\"]" {
			t.Errorf("encode []string{123,456,789000000} failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode []string{123,456,789000000} failed, val:%v", eVal)
		return
	}

	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode []string{123,456,789000000}, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode []string{123,456,789000000}, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode []string{123,456,789000000}, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode  []string{123,456,789000000}, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 3 {
		t.Errorf("decode  []string{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(0).String() != "123" {
		t.Errorf("decode  []string{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(1).String() != "456" {
		t.Errorf("decode  []string{123,456,789000000}, is 123")
		return
	}
	if dVal.Get().(reflect.Value).Index(2).String() != "789000000" {
		t.Errorf("decode  []string{123,456,789000000}, is 123")
		return
	}
}

func TestBaseDateCodec(t *testing.T) {
	val, _ := time.Parse(util.CSTLayout, "2006-01-02 15:04:05")
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new dateTime type failed, err:%s", err.Error())
		return
	}

	// dateTime("2006-01-02 15:04:05")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode dateTime 2006-01-02 15:04:05, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "2006-01-02 15:04:05" {
			t.Errorf("encode dateTime failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode 2006-01-02 15:04:05 failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is basic")
		return
	}
	dtVal := dVal.Get().(reflect.Value).Interface().(time.Time).Format(util.CSTLayout)
	if dtVal != "2006-01-02 15:04:05" {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is 2006-01-02 15:04:05")
		return
	}
}

func TestSliceDateCodec(t *testing.T) {
	dt, _ := time.Parse(util.CSTLayout, "2006-01-02 15:04:05")
	val := []time.Time{dt}
	rVal := reflect.ValueOf(val)
	rType := reflect.TypeOf(val)

	valPtr := NewValue(rVal)
	typePtr, err := NewType(rType)
	if err != nil {
		t.Errorf("new dateTime type failed, err:%s", err.Error())
		return
	}

	// dateTime("2006-01-02 15:04:05")
	eVal, eErr := _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode dateTime 2006-01-02 15:04:05, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "2006-01-02 15:04:05" {
			t.Errorf("encode dateTime failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode 2006-01-02 15:04:05 failed, val:%v", eVal)
		return
	}
	dVal, dErr := _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is basic")
		return
	}
	dtVal := dVal.Get().(reflect.Value).Index(0).Interface().(time.Time).Format(util.CSTLayout)
	if dtVal != "2006-01-02 15:04:05" {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is 2006-01-02 15:04:05")
		return
	}

	dt1, _ := time.Parse(util.CSTLayout, "2006-01-02 15:04:05")
	dt2, _ := time.Parse(util.CSTLayout, "2007-01-02 15:04:05")
	val = []time.Time{dt1, dt2}
	rVal = reflect.ValueOf(val)
	rType = reflect.TypeOf(val)

	valPtr = NewValue(rVal)
	typePtr, err = NewType(rType)
	if err != nil {
		t.Errorf("new dateTime type failed, err:%s", err.Error())
		return
	}

	// dateTime(["2006-01-02 15:04:05","2007-01-02 15:04:05"])
	eVal, eErr = _codec.Encode(valPtr, typePtr)
	if eErr != nil {
		t.Errorf("encode dateTime 2006-01-02 15:04:05, err:%s", eErr.Error())
		return
	}
	switch eVal.(type) {
	case string:
		if eVal.(string) != "[\"2006-01-02 15:04:05\",\"2007-01-02 15:04:05\"]" {
			t.Errorf("encode dateTime failed, val:%v", eVal)
			return
		}
	default:
		t.Errorf("encode 2006-01-02 15:04:05 failed, val:%v", eVal)
		return
	}
	dVal, dErr = _codec.Decode(eVal, typePtr)
	if dErr != nil {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, err:%s", dErr.Error())
		return
	}
	if dVal.IsNil() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is not nil")
		return
	}
	if dVal.IsZero() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is not zero")
		return
	}
	if !dVal.IsBasic() {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is basic")
		return
	}
	if dVal.Get().(reflect.Value).Len() != 2 {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is 2006-01-02 15:04:05")
		return
	}
	dt1Val := dVal.Get().(reflect.Value).Index(0).Interface().(time.Time).Format(util.CSTLayout)
	if dt1Val != "2006-01-02 15:04:05" {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is 2006-01-02 15:04:05")
		return
	}
	dt2Val := dVal.Get().(reflect.Value).Index(1).Interface().(time.Time).Format(util.CSTLayout)
	if dt2Val != "2007-01-02 15:04:05" {
		t.Errorf("decode dateTime 2006-01-02 15:04:05, is 2006-01-02 15:04:05")
		return
	}
}
