package local

import (
	"reflect"
	"testing"
	"time"
)

func TestGetValueStr(t *testing.T) {
	iVal := int(123)
	fiVal, fiErr := newValue(reflect.ValueOf(&iVal))
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
	} else {
		ret, _ := fiVal.Str()
		if ret != "123" {
			t.Errorf("GetValueStr failed, iVal:%d", iVal)
		}
	}

	fVal := 12.34
	ffVal, ffErr := newValue(reflect.ValueOf(&fVal))
	if ffErr != nil {
		t.Errorf("%s", ffErr.Error())
	} else {
		ret, _ := ffVal.Str()
		if ret != "12.340000" {
			t.Errorf("GetValueStr failed, fVal:%f", fVal)
		}
	}

	strVal := "abc"
	fstrVal, fstrErr := newValue(reflect.ValueOf(&strVal))
	if fstrErr != nil {
		t.Errorf("%s", fstrErr.Error())
	} else {
		ret, _ := fstrVal.Str()
		if ret != "'abc'" {
			t.Errorf("GetValueStr failed, ret:%s, strVal:%s", ret, strVal)
		}
	}

	bVal := true
	fbVal, fbErr := newValue(reflect.ValueOf(&bVal))
	if fbErr != nil {
		t.Errorf("%s", fbErr.Error())
	} else {
		ret, _ := fbVal.Str()
		if ret != "1" {
			t.Errorf("GetValueStr failed, ret:%s, bVal:%v", ret, bVal)
		}
	}

	now, _ := time.ParseInLocation("2006-01-02 15:04:05", "2018-01-02 15:04:05", time.Local)
	ftimeVal, ftimeErr := newValue(reflect.ValueOf(&now))
	if ftimeErr != nil {
		t.Errorf("%s", ftimeErr.Error())
	} else {
		ret, _ := ftimeVal.Str()
		if ret != "'2018-01-02 15:04:05'" {
			t.Errorf("GetValueStr failed, ret:%s, ftimeVal:%v", ret, now)
		}
	}

	ii := 123
	var iiVal int
	iiVal = ii
	fiVal, fiErr = newValue(reflect.ValueOf(&iiVal))
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
	} else {
		ret, err := fiVal.Str()
		if err != nil {
			t.Errorf("GetValueStr failed, err:%s", err.Error())
		} else {
			if ret != "123" {
				t.Errorf("GetValueStr failed, iVal:%d", iVal)
			}
		}
	}
}

func TestSetValue(t *testing.T) {
	var iVal int
	fiVal, fiErr := newValue(reflect.ValueOf(&iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
	} else {
		intVal := 123
		fiVal.Set(reflect.ValueOf(intVal))
		ret, _ := fiVal.Str()
		if ret != "123" {
			t.Errorf("GetValueStr failed, iVal:%d", iVal)
		}
		if iVal != 123 {
			t.Errorf("Set failed, iVal:%d", iVal)
		}
	}

	var fVal float32
	ffVal, ffErr := newValue(reflect.ValueOf(&fVal).Elem())
	if ffErr != nil {
		t.Errorf("%s", ffErr.Error())
	} else {
		fltVal := float32(12.34)
		ffVal.Set(reflect.ValueOf(fltVal))
		ret, _ := ffVal.Str()
		if ret != "12.340000" {
			t.Errorf("GetValueStr failed, fVal:%f", fVal)
		}
		if fVal != 12.34 {
			t.Errorf("Set failed, fVal:%f", fVal)
		}
	}

	var strVal string
	fstrVal, fstrErr := newValue(reflect.ValueOf(&strVal).Elem())
	if fstrErr != nil {
		t.Errorf("%s", fstrErr.Error())
	} else {
		stringVal := "abc"
		fstrVal.Set(reflect.ValueOf(stringVal))
		ret, _ := fstrVal.Str()
		if ret != "'abc'" {
			t.Errorf("GetValueStr failed, ret:%s, strVal:%s", ret, strVal)
		}
		if strVal != "abc" {
			t.Errorf("Set failed, strVal:%s", strVal)
		}
	}

	var bVal bool
	fbVal, fbErr := newValue(reflect.ValueOf(&bVal).Elem())
	if fbErr != nil {
		t.Errorf("%s", fbErr.Error())
	} else {
		boolVal := true
		fbVal.Set(reflect.ValueOf(boolVal))
		ret, _ := fbVal.Str()
		if ret != "1" {
			t.Errorf("GetValueStr failed, ret:%s, bVal:%v", ret, bVal)
		}
		if !bVal {
			t.Errorf("Set failed, bVal:%v", bVal)
		}
		bIntVal := false
		fbVal.Set(reflect.ValueOf(bIntVal))
		ret, _ = fbVal.Str()
		if ret != "0" {
			t.Errorf("GetValueStr failed, ret:%s, bVal:%v", ret, bVal)
		}
		if bVal {
			t.Errorf("Set failed, bVal:%v", bVal)
		}
	}
}

func TestPtr(t *testing.T) {
	ii := 10
	jj := 20
	var iVal *int

	iVal = &jj
	fiVal, fiErr := newValue(reflect.ValueOf(&iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
		return
	}
	ret, err := fiVal.Str()
	if err != nil {
		t.Errorf("GetValueStr exception, err:%s", err.Error())
		return
	}

	fiVal.Set(reflect.ValueOf(&ii))
	ret, err = fiVal.Str()
	if err != nil {
		t.Errorf("GetValueStr failed, err:%s", err.Error())
		return
	}
	if ret != "10" {
		t.Errorf("GetValueStr exception, iVal:%d, ret:%s", *iVal, ret)
		return
	}

	if *iVal != ii {
		t.Errorf("GetValueStr exception, iVal:%d, ii:%d", *iVal, ii)
		return
	}

	iVal = &ii
	fiVal, fiErr = newValue(reflect.ValueOf(iVal).Elem())
	if fiErr != nil {
		t.Errorf("%s", fiErr.Error())
	} else {
		ret, err := fiVal.Str()
		if err != nil {
			t.Errorf("GetValueStr failed, err:%s", err.Error())
		} else {
			if ret != "10" {
				t.Errorf("GetValueStr failed, ret:%s, iVal:%d", ret, iVal)
			}
			if *iVal != 10 {
				t.Errorf("Set failed, iVal:%d", iVal)
			}
		}

		intVal := 123
		fiVal.Set(reflect.ValueOf(intVal))
		ret, err = fiVal.Str()
		if err != nil {
			t.Errorf("GetValueStr failed, err:%s", err.Error())
		} else {
			if ret != "123" {
				t.Errorf("GetValueStr failed, ret:%s, iVal:%d", ret, iVal)
			}
			if *iVal != 123 {
				t.Errorf("Set failed, iVal:%d", iVal)
			}

		}
	}
}
