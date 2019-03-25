package local

import (
	"log"
	"reflect"
	"testing"
)

func TestIntSlice(t *testing.T) {
	data := []int64{112, 223}

	strVal, strErr := EncodeSliceValue(reflect.ValueOf(data))
	if strErr != nil {
		t.Errorf("marshal failed, err:%s", strErr.Error())
		return
	}

	fType, fErr := newType(reflect.TypeOf(data))
	if fErr != nil {
		t.Errorf("illegal data type")
		return
	}

	ret, err := DecodeSliceValue(strVal, fType)
	if err != nil {
		t.Errorf("DecodeSliceValue failed, err:%s", err.Error())
		return
	}

	log.Print(ret.Interface())
}

func TestStrSlice(t *testing.T) {
	data := []string{"aab", "ccd"}

	strVal, strErr := EncodeSliceValue(reflect.ValueOf(data))
	if strErr != nil {
		t.Errorf("marshal failed, err:%s", strErr.Error())
		return
	}

	fType, fErr := newType(reflect.TypeOf(data))
	if fErr != nil {
		t.Errorf("illegal data type")
		return
	}

	ret, err := DecodeSliceValue(strVal, fType)
	if err != nil {
		t.Errorf("DecodeSliceValue failed, err:%s", err.Error())
		return
	}

	log.Print(ret.Interface())
}
