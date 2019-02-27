package local

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

func TestIntSlice(t *testing.T) {
	data := []int64{112, 223}

	byteVal, byteErr := json.Marshal(data)
	if byteErr != nil {
		t.Errorf("marshal failed, err:%s", byteErr.Error())
		return
	}

	fType, fErr := newType(reflect.TypeOf(data))
	if fErr != nil {
		t.Errorf("illegal data type")
		return
	}

	ret, err := getSliceFromString(reflect.ValueOf(string(byteVal)), fType)
	if err != nil {
		t.Errorf("getSliceFromString failed, err:%s", err.Error())
		return
	}

	log.Print(ret.Interface())
}

func TestStrSlice(t *testing.T) {
	data := []string{"aab", "ccd"}

	byteVal, byteErr := json.Marshal(data)
	if byteErr != nil {
		t.Errorf("marshal failed, err:%s", byteErr.Error())
		return
	}

	fType, fErr := newType(reflect.TypeOf(data))
	if fErr != nil {
		t.Errorf("illegal data type")
		return
	}

	ret, err := getSliceFromString(reflect.ValueOf(string(byteVal)), fType)
	if err != nil {
		t.Errorf("getSliceFromString failed, err:%s", err.Error())
		return
	}

	log.Print(ret.Interface())
}
