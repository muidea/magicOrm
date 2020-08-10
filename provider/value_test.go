package provider

import (
	"github.com/muidea/magicOrm/provider/local"
	"reflect"
	"testing"
	"time"
)

func TestGetValueStr(t *testing.T) {
	iVal := int(123)
	fiType, fiErr := local.GetType(reflect.ValueOf(iVal))
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	fiVal, fiErr := local.GetValue(reflect.ValueOf(iVal))
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	ret, _ := getBasicValue(fiType, fiVal.Get())
	if ret != "123" {
		t.Errorf("getValueStr failed, iVal:%d", iVal)
		return
	}

	fVal := 12.34
	ffType, ffErr := local.GetType(reflect.ValueOf(fVal))
	if ffErr != nil {
		t.Errorf("%s", ffErr.Error())
		return
	}
	ffVal, ffErr := local.GetValue(reflect.ValueOf(fVal))
	if ffErr != nil {
		t.Errorf("%s", ffErr.Error())
		return
	}
	ret, _ = getBasicValue(ffType, ffVal.Get())
	if ret != "12.340000" {
		t.Errorf("getValueStr failed, fVal:%f", fVal)
	}

	strVal := "abc"
	fstrType, fstrErr := local.GetType(reflect.ValueOf(strVal))
	if fstrErr != nil {
		t.Errorf("%s", fstrErr.Error())
		return
	}

	fstrVal, fstrErr := local.GetValue(reflect.ValueOf(strVal))
	if fstrErr != nil {
		t.Errorf("%s", fstrErr.Error())
		return
	}
	ret, _ = getBasicValue(fstrType, fstrVal.Get())
	if ret != "'abc'" {
		t.Errorf("getValueStr failed, ret:%s, strVal:%s", ret, strVal)
		return
	}

	bVal := true
	fbType, fbErr := local.GetType(reflect.ValueOf(bVal))
	if fbErr != nil {
		t.Errorf("%s", fbErr.Error())
		return
	}

	fbVal, fbErr := local.GetValue(reflect.ValueOf(bVal))
	if fbErr != nil {
		t.Errorf("%s", fbErr.Error())
		return
	}
	ret, _ = getBasicValue(fbType, fbVal.Get())
	if ret != "1" {
		t.Errorf("getValueStr failed, ret:%s, bVal:%v", ret, bVal)
		return
	}

	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-01-02 15:04:05", time.Local)
	ftimeType, ftimeErr := local.GetType(reflect.ValueOf(now))
	if ftimeErr != nil {
		t.Errorf("%s", ftimeErr.Error())
		return
	}
	ftimeVal, ftimeErr := local.GetValue(reflect.ValueOf(now))
	if ftimeErr != nil {
		t.Errorf("%s", ftimeErr.Error())
		return
	}
	ret, _ = getBasicValue(ftimeType, ftimeVal.Get())
	if ret != "'2018-01-02 15:04:05'" {
		t.Errorf("getValueStr failed, ret:%s, ftimeVal:%v", ret, now)
	}

	ii := 123
	var iiVal int
	iiVal = ii
	fiType, fiErr = local.GetType(reflect.ValueOf(iiVal))
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	fiVal, fiErr = local.GetValue(reflect.ValueOf(iiVal))
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	ret, _ = getBasicValue(fiType, fiVal.Get())
	if ret != "123" {
		t.Errorf("getValueStr failed, iVal:%d", iVal)
	}
}

func TestSetValue(t *testing.T) {
	var iVal int
	fiType, fiErr := local.GetType(reflect.ValueOf(&iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	fiVal, fiErr := local.GetValue(reflect.ValueOf(&iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	intVal := 123
	fiVal.Update(reflect.ValueOf(intVal))
	ret, _ := getBasicValue(fiType, fiVal.Get())
	if ret != "123" {
		t.Errorf("getValueStr failed, iVal:%d", iVal)
		return
	}
	if iVal != 123 {
		t.Errorf("Set failed, iVal:%d", iVal)
	}

	var fVal float32
	ffType, ffErr := local.GetType(reflect.ValueOf(&fVal).Elem())
	if ffErr != nil {
		t.Errorf("%s", ffErr.Error())
		return
	}
	ffVal, ffErr := local.GetValue(reflect.ValueOf(&fVal).Elem())
	if ffErr != nil {
		t.Errorf("%s", ffErr.Error())
		return
	}
	fltVal := float32(12.34)
	ffVal.Update(reflect.ValueOf(fltVal))
	ret, _ = getBasicValue(ffType, ffVal.Get())
	if ret != "12.340000" {
		t.Errorf("getValueStr failed, fVal:%f", fVal)
		return
	}
	if fVal != 12.34 {
		t.Errorf("Set failed, fVal:%f", fVal)
	}

	var strVal string
	fstrType, fstrErr := local.GetType(reflect.ValueOf(&strVal).Elem())
	if fstrErr != nil {
		t.Errorf("%s", fstrErr.Error())
		return
	}
	fstrVal, fstrErr := local.GetValue(reflect.ValueOf(&strVal).Elem())
	if fstrErr != nil {
		t.Errorf("%s", fstrErr.Error())
		return
	}

	stringVal := "abc"
	fstrVal.Update(reflect.ValueOf(stringVal))
	ret, _ = getBasicValue(fstrType, fstrVal.Get())
	if ret != "'abc'" {
		t.Errorf("getValueStr failed, ret:%s, strVal:%s", ret, strVal)
		return
	}
	if strVal != "abc" {
		t.Errorf("Set failed, strVal:%s", strVal)
		return
	}

	var bVal bool
	fbType, fbErr := local.GetType(reflect.ValueOf(&bVal).Elem())
	if fbErr != nil {
		t.Errorf("%s", fbErr.Error())
		return
	}
	fbVal, fbErr := local.GetValue(reflect.ValueOf(&bVal).Elem())
	if fbErr != nil {
		t.Errorf("%s", fbErr.Error())
		return
	}
	boolVal := true
	fbVal.Update(reflect.ValueOf(boolVal))
	ret, _ = getBasicValue(fbType, fbVal.Get())
	if ret != "1" {
		t.Errorf("getValueStr failed, ret:%s, bVal:%v", ret, bVal)
		return
	}
	if !bVal {
		t.Errorf("Set failed, bVal:%v", bVal)
		return
	}
	bIntVal := false
	fbVal.Update(reflect.ValueOf(bIntVal))
	ret, _ = getBasicValue(fbType, fbVal.Get())
	if ret != "0" {
		t.Errorf("getValueStr failed, ret:%s, bVal:%v", ret, bVal)
	}
	if bVal {
		t.Errorf("Set failed, bVal:%v", bVal)
	}
}

func TestPtr(t *testing.T) {
	ii := 10
	jj := 20
	var iVal *int

	iVal = &jj
	reVal := &ii
	fiType, fiErr := local.GetType(reflect.ValueOf(&iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	fiVal, fiErr := local.GetValue(reflect.ValueOf(&iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}

	err := fiVal.Update(reflect.ValueOf(reVal))
	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}

	ret, err := getBasicValue(fiType, fiVal.Get())
	if err != nil {
		t.Errorf("%s", err.Error())
		return
	}
	if ret != "10" {
		t.Errorf("getValueStr exception, iVal:%d, ret:%s", *iVal, ret)
		return
	}

	if *iVal != ii {
		t.Errorf("getValueStr exception, iVal:%d, ii:%d", *iVal, ii)
		return
	}

	iVal = &ii
	fiType, fiErr = local.GetType(reflect.ValueOf(iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	fiVal, fiErr = local.GetValue(reflect.ValueOf(iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	ret, _ = getBasicValue(fiType, fiVal.Get())
	if ret != "10" {
		t.Errorf("getValueStr failed, ret:%s, iVal:%d", ret, iVal)
	}
	if *iVal != 10 {
		t.Errorf("Set failed, iVal:%d", iVal)
	}

	intVal := 123
	fiVal.Update(reflect.ValueOf(intVal))
	ret, _ = getBasicValue(fiType, fiVal.Get())
	if ret != "123" {
		t.Errorf("getValueStr failed, ret:%s, iVal:%d", ret, iVal)
	}
	if *iVal != 123 {
		t.Errorf("Set failed, iVal:%d", iVal)
	}

}
