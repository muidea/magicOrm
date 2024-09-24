package local

import (
	"reflect"
	"testing"
)

func TestGetEntityValue(t *testing.T) {
	var entity interface{}
	eVal, eErr := GetEntityValue(entity)
	if eErr == nil {
		t.Errorf("GetEntityValue failed")
		return
	}
	iVal := 123
	entity = iVal
	eVal, eErr = GetEntityValue(entity)
	if eErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eErr.Error())
		return
	}

	iVal2 := 234
	eVal.Set(reflect.ValueOf(iVal2))

	niVal, niOK := eVal.Interface().(int)
	if !niOK || niVal != iVal2 {
		t.Errorf("GetEntityValue failed")
		return
	}

	iValArray := []int{iVal, iVal2}
	eArrayVal, eArrayErr := GetEntityValue(iValArray)
	if eArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eArrayErr.Error())
		return
	}
	niValArray, niValOK := eArrayVal.Interface().([]int)
	if !niValOK {
		t.Errorf("GetEntityValue failed")
		return
	}
	if len(niValArray) != len(iValArray) {
		t.Errorf("GetEntityValue failed")
		return
	}

	eArrayVal, eArrayErr = AppendSliceValue(eArrayVal, eVal)
	if eArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eArrayErr.Error())
		return
	}
	niValArray, niValOK = eArrayVal.Interface().([]int)
	if !niValOK {
		t.Errorf("GetEntityValue failed")
		return
	}
	if len(niValArray) != len(iValArray)+1 {
		t.Errorf("GetEntityValue failed")
		return
	}
}

func TestGetEntityType(t *testing.T) {
	var entity interface{}
	eVal, eErr := GetEntityValue(entity)
	if eErr == nil {
		t.Errorf("GetEntityValue failed")
		return
	}
	iVal := 123
	entity = iVal
	eVal, eErr = GetEntityValue(entity)
	if eErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eErr.Error())
		return
	}

	iVal2 := 234
	eVal.Set(reflect.ValueOf(iVal2))

	eType, eErr := GetEntityType(entity)
	if eErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eErr.Error())
		return
	}
	eVal, eErr = eType.Interface(iVal2)
	if eErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eErr.Error())
		return
	}

	niVal, niOK := eVal.Interface().(int)
	if !niOK || niVal != iVal2 {
		t.Errorf("GetEntityValue failed")
		return
	}

	iValArray := []int{iVal, iVal2}
	eArrayVal, eArrayErr := GetEntityValue(iValArray)
	if eArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eArrayErr.Error())
		return
	}
	eArrayType, eArrayErr := GetEntityType(iValArray)
	if eArrayErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eArrayErr.Error())
		return
	}
	eArrayVal, eArrayErr = eArrayType.Interface(iValArray)
	if eArrayErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eArrayErr.Error())
		return
	}

	niValArray, niValOK := eArrayVal.Interface().([]int)
	if !niValOK {
		t.Errorf("GetEntityValue failed")
		return
	}
	if len(niValArray) != len(iValArray) {
		t.Errorf("GetEntityValue failed")
		return
	}

	eArrayVal, eArrayErr = AppendSliceValue(eArrayVal, eVal)
	if eArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", eArrayErr.Error())
		return
	}
	niValArray, niValOK = eArrayVal.Interface().([]int)
	if !niValOK {
		t.Errorf("GetEntityValue failed")
		return
	}
	if len(niValArray) != len(iValArray)+1 {
		t.Errorf("GetEntityValue failed")
		return
	}

	iValPtrArray := []*int{&iVal, &iVal2}
	ePtrArrayVal, ePtrArrayErr := GetEntityValue(iValPtrArray)
	if ePtrArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", ePtrArrayErr.Error())
		return
	}
	ePtrArrayType, ePtrArrayErr := GetEntityType(iValPtrArray)
	if ePtrArrayErr != nil {
		t.Errorf("GetEntityType failed, error:%s", eArrayErr.Error())
		return
	}
	ePtrArrayVal, ePtrArrayErr = ePtrArrayType.Interface(iValPtrArray)
	if ePtrArrayErr != nil {
		t.Errorf("GetEntityType failed, error:%s", ePtrArrayErr.Error())
		return
	}

	niValPtrArray, niValPtrOK := ePtrArrayVal.Interface().([]*int)
	if !niValPtrOK {
		t.Errorf("GetEntityValue failed")
		return
	}
	if len(niValPtrArray) != len(iValPtrArray) {
		t.Errorf("GetEntityValue failed")
		return
	}

	ePtrVal, ePtrErr := GetEntityValue(&iVal)
	if ePtrErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", ePtrErr.Error())
		return
	}

	ePtrArrayVal, ePtrArrayErr = AppendSliceValue(ePtrArrayVal, ePtrVal)
	if ePtrArrayErr != nil {
		t.Errorf("GetEntityValue failed, error:%s", ePtrArrayErr.Error())
		return
	}
	niValPtrArray, niValPtrOK = ePtrArrayVal.Interface().([]*int)
	if !niValPtrOK {
		t.Errorf("GetEntityValue failed")
		return
	}
	if len(niValPtrArray) != len(iValPtrArray)+1 {
		t.Errorf("GetEntityValue failed")
		return
	}
}
