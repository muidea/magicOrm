package remote

import (
	"reflect"
	"testing"
)

func TestGetInitValue(t *testing.T) {
	var iVal int
	iType, iErr := newType(reflect.TypeOf(iVal))
	if iErr != nil {
		t.Errorf("newType failed, err:%s", iErr.Error())
		return
	}

	tVal := iType.Interface()
	if !tVal.Get().CanSet() || !tVal.Get().CanAddr() {
		t.Errorf("Interface value failed")
	}

	iValPtr := &iVal
	iType, iErr = newType(reflect.TypeOf(iValPtr))
	if iErr != nil {
		t.Errorf("newType failed, err:%s", iErr.Error())
		return
	}

	tValPtr := iType.Interface()
	if !tValPtr.Get().CanSet() || !tValPtr.Get().CanAddr() {
		t.Errorf("Interface value failed")
	}
}
