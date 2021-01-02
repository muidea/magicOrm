package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/util"
)

func TestIntType(t *testing.T) {
	var iVal int
	iType, iErr := newType(reflect.TypeOf(iVal))
	if iErr != nil {
		t.Errorf("newType failed, err:%s", iErr.Error())
		return
	}

	if iType.GetValue() != util.TypeIntegerField {
		t.Errorf("get int type value failed.")
		return
	}

	newVal := iType.Interface().Get().(reflect.Value)
	riType, riErr := newType(newVal.Type())
	if riErr != nil {
		t.Errorf("newType failed, err:%s", riErr.Error())
		return
	}

	if riType.GetValue() != util.TypeIntegerField {
		t.Errorf("get int type value failed.")
		return
	}
	if iType.GetName() != riType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if iType.GetPkgPath() != riType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if iType.GetValue() != riType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}
}

func TestFloatType(t *testing.T) {
	var fVal float32
	fType, fErr := newType(reflect.TypeOf(fVal))
	if fErr != nil {
		t.Errorf("newType failed, err:%s", fErr.Error())
		return
	}

	if fType.GetValue() != util.TypeFloatField {
		t.Errorf("get float type value failed.")
		return
	}

	newVal := fType.Interface().Get().(reflect.Value)
	rfType, rfErr := newType(newVal.Type())
	if rfErr != nil {
		t.Errorf("newType failed, err:%s", rfErr.Error())
		return
	}

	if rfType.GetValue() != util.TypeFloatField {
		t.Errorf("get float type value failed.")
		return
	}
	if fType.GetName() != rfType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if fType.GetPkgPath() != rfType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if fType.GetValue() != rfType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}
}

func TestBoolType(t *testing.T) {
	var bVal bool
	bType, bErr := newType(reflect.TypeOf(bVal))
	if bErr != nil {
		t.Errorf("newType failed, err:%s", bErr.Error())
		return
	}

	if bType.GetValue() != util.TypeBooleanField {
		t.Errorf("get bool type value failed.")
		return
	}

	newVal := bType.Interface().Get().(reflect.Value)
	rbType, rbErr := newType(newVal.Type())
	if rbErr != nil {
		t.Errorf("newType failed, err:%s", rbErr.Error())
		return
	}

	if rbType.GetValue() != util.TypeBooleanField {
		t.Errorf("get bool type value failed.")
		return
	}
	if bType.GetName() != rbType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if bType.GetPkgPath() != rbType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if bType.GetValue() != rbType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}
}

func TestStringType(t *testing.T) {
	var strVal string
	strType, strErr := newType(reflect.TypeOf(strVal))
	if strErr != nil {
		t.Errorf("newType failed, err:%s", strErr.Error())
		return
	}

	if strType.GetValue() != util.TypeStringField {
		t.Errorf("get string type value failed.")
		return
	}

	newVal := strType.Interface().Get().(reflect.Value)
	rstrType, rstrErr := newType(newVal.Type())
	if rstrErr != nil {
		t.Errorf("newType failed, err:%s", rstrErr.Error())
		return
	}

	if rstrType.GetValue() != util.TypeStringField {
		t.Errorf("get string type value failed.")
		return
	}
	if strType.GetName() != rstrType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if strType.GetPkgPath() != rstrType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if strType.GetValue() != rstrType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}
}

func TestDateTimeType(t *testing.T) {
	var dtVal time.Time
	dtType, dtErr := newType(reflect.TypeOf(dtVal))
	if dtErr != nil {
		t.Errorf("newType failed, err:%s", dtErr.Error())
		return
	}

	if dtType.GetValue() != util.TypeDateTimeField {
		t.Errorf("get DateTime type value failed.")
		return
	}

	newVal := dtType.Interface().Get().(reflect.Value)
	rdtType, rdtErr := newType(newVal.Type())
	if rdtErr != nil {
		t.Errorf("newType failed, err:%s", rdtErr.Error())
		return
	}

	if rdtType.GetValue() != util.TypeDateTimeField {
		t.Errorf("get DateTime type value failed.")
		return
	}
	if dtType.GetName() != rdtType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if dtType.GetPkgPath() != rdtType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if dtType.GetValue() != rdtType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}
}

func TestStructType(t *testing.T) {
	type Base struct {
		iVal int
	}

	var structVal Base
	structType, structErr := newType(reflect.TypeOf(structVal))
	if structErr != nil {
		t.Errorf("newType failed, err:%s", structErr.Error())
		return
	}

	if structType.GetValue() != util.TypeStructField {
		t.Errorf("get DateTime type value failed.")
		return
	}

	newVal := structType.Interface().Get().(reflect.Value)
	rstructType, rstructErr := newType(newVal.Type())
	if rstructErr != nil {
		t.Errorf("newType failed, err:%s", rstructErr.Error())
		return
	}

	if rstructType.GetValue() != util.TypeStructField {
		t.Errorf("get DateTime type value failed.")
		return
	}
	if structType.GetName() != rstructType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if structType.GetPkgPath() != rstructType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if structType.GetValue() != rstructType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}

	if structType.IsPtrType() {
		t.Errorf("unexpect isPtrType")
		return
	}
}

func TestSliceType(t *testing.T) {
	var sliceVal []*uint16
	sliceType, sliceErr := newType(reflect.TypeOf(sliceVal))
	if sliceErr != nil {
		t.Errorf("newType failed, err:%s", sliceErr.Error())
		return
	}

	if sliceType.GetValue() != util.TypeSliceField {
		t.Errorf("get Slice type value failed.")
		return
	}

	newVal := sliceType.Interface().Get().(reflect.Value)
	rsliceType, rsliceErr := newType(newVal.Type())
	if rsliceErr != nil {
		t.Errorf("newType failed, err:%s", rsliceErr.Error())
		return
	}

	if rsliceType.GetValue() != util.TypeSliceField {
		t.Errorf("get Slice type value failed.")
		return
	}
	if sliceType.GetName() != rsliceType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if sliceType.GetPkgPath() != rsliceType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if sliceType.GetValue() != rsliceType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}

	dependType := sliceType.Elem()
	if dependType == nil {
		t.Errorf("illegal depend")
		return
	}

	if dependType.GetValue() != util.TypePositiveSmallIntegerField {
		t.Errorf("illegal depend type value")
		return
	}

	if !dependType.IsPtrType() {
		t.Errorf("illegal depend type value")
		return
	}

	elemType := sliceType.Elem()
	if elemType == nil {
		t.Errorf("illegal elem")
		return
	}

	if elemType.GetValue() != util.TypePositiveSmallIntegerField {
		t.Errorf("illegal elem type value")
		return
	}

	if !elemType.IsPtrType() {
		t.Errorf("illegal elem type value")
		return
	}

	if sliceType.IsPtrType() {
		t.Errorf("unexpect isPtrType")
		return
	}
}

func TestPtrSliceType(t *testing.T) {
	var sliceVal *[]*uint16
	sliceType, sliceErr := newType(reflect.TypeOf(sliceVal))
	if sliceErr != nil {
		t.Errorf("newType failed, err:%s", sliceErr.Error())
		return
	}

	if sliceType.GetValue() != util.TypeSliceField {
		t.Errorf("get Slice type value failed.")
		return
	}

	newVal := sliceType.Interface().Get().(reflect.Value)
	rsliceType, rsliceErr := newType(newVal.Type())
	if rsliceErr != nil {
		t.Errorf("newType failed, err:%s", rsliceErr.Error())
		return
	}

	if rsliceType.GetValue() != util.TypeSliceField {
		t.Errorf("get Slice type value failed.")
		return
	}
	if sliceType.GetName() != rsliceType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if sliceType.GetPkgPath() != rsliceType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if sliceType.GetValue() != rsliceType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}

	dependType := sliceType.Elem()
	if dependType == nil {
		t.Errorf("illegal depend")
		return
	}

	if dependType.GetValue() != util.TypePositiveSmallIntegerField {
		t.Errorf("illegal depend type value")
		return
	}

	if !dependType.IsPtrType() {
		t.Errorf("illegal depend type value")
		return
	}

	elemType := sliceType.Elem()
	if elemType == nil {
		t.Errorf("illegal elem")
		return
	}

	if elemType.GetValue() != util.TypePositiveSmallIntegerField {
		t.Errorf("illegal elem type value")
		return
	}

	if !elemType.IsPtrType() {
		t.Errorf("illegal elem type value")
		return
	}

	if !sliceType.IsPtrType() {
		t.Errorf("unexpect isPtrType")
		return
	}
}

func TestSliceStructType(t *testing.T) {
	type Base struct {
		iVal int
	}
	var sliceVal []*Base
	sliceType, sliceErr := newType(reflect.TypeOf(sliceVal))
	if sliceErr != nil {
		t.Errorf("newType failed, err:%s", sliceErr.Error())
		return
	}

	if sliceType.GetValue() != util.TypeSliceField {
		t.Errorf("get Slice type value failed.")
		return
	}

	newVal := sliceType.Interface().Get().(reflect.Value)
	rsliceType, rsliceErr := newType(newVal.Type())
	if rsliceErr != nil {
		t.Errorf("newType failed, err:%s", rsliceErr.Error())
		return
	}

	if rsliceType.GetValue() != util.TypeSliceField {
		t.Errorf("get Slice type value failed.")
		return
	}
	if sliceType.GetName() != rsliceType.GetName() {
		t.Errorf("newType faild. illegal type name")
		return
	}
	if sliceType.GetPkgPath() != rsliceType.GetPkgPath() {
		t.Errorf("newType faild. illegal type pkgPath")
		return
	}
	if sliceType.GetValue() != rsliceType.GetValue() {
		t.Errorf("newType faild. illegal type value")
		return
	}

	dependType := sliceType.Elem()
	if dependType == nil {
		t.Errorf("illegal depend")
		return
	}

	if dependType.GetValue() != util.TypeStructField {
		t.Errorf("illegal depend type value")
		return
	}

	if !dependType.IsPtrType() {
		t.Errorf("illegal depend type value")
		return
	}

	elemType := sliceType.Elem()
	if elemType == nil {
		t.Errorf("illegal elem")
		return
	}

	if elemType.GetValue() != util.TypeStructField {
		t.Errorf("illegal elem type value")
		return
	}

	if !elemType.IsPtrType() {
		t.Errorf("illegal elem type value")
		return
	}
	if sliceType.IsPtrType() {
		t.Errorf("unexpect isPtrType")
		return
	}
}
