package remote

import (
	"log"
	"reflect"
	"testing"
)

func TestItem(t *testing.T) {
	idx := 12
	name := "Tt"
	iTag, iErr := newTag("tt")
	if iErr != nil {
		t.Errorf("newTag failed, err:%s", iErr.Error())
	}
	iType, iErr := newType(reflect.TypeOf(idx))
	if iErr != nil {
		t.Errorf("newType failed, err:%s", iErr.Error())
		return
	}

	item := newItem(idx, name, iTag, iType)
	if item.IsAssigned() {
		t.Errorf("check item assigned failed")
		return
	}

	fValue := 12.00
	iErr = item.SetValue(reflect.ValueOf(&fValue).Elem())
	if iErr != nil {
		t.Errorf("SetValue failed, err:%s", iErr.Error())
		return
	}

	iErr = item.UpdateValue(reflect.ValueOf(23.00))
	if iErr != nil {
		t.Errorf("UpdateValue failed, err:%s", iErr.Error())
		return
	}

	if fValue != 23.00 {
		t.Errorf("assigned value failed")
		return
	}

	log.Printf("curValue:%0.2f", fValue)
}
