package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicOrm/model"
)

func TestIntType(t *testing.T) {
	var iVal int = 100
	iType, iErr := NewType(reflect.TypeOf(iVal))
	if iErr != nil {
		t.Errorf("NewType failed, err:%s", iErr.Error())
		return
	}

	if iType.GetValue() != model.TypeIntegerValue {
		t.Errorf("get int type value failed.")
		return
	}

	nVal, _ := iType.Interface(nil)
	switch nVal.Get().(reflect.Value).Interface().(type) {
	case int:
	default:
		t.Errorf("get int type value failed. type:%s", nVal.Get().(reflect.Value).Type().String())
		return
	}

	newVal := nVal.Get().(reflect.Value)
	riType, riErr := NewType(newVal.Type())
	if riErr != nil {
		t.Errorf("NewType failed, err:%s", riErr.Error())
		return
	}

	if riType.GetValue() != model.TypeIntegerValue {
		t.Errorf("get int type value failed.")
		return
	}
	if iType.GetName() != riType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if iType.GetPkgPath() != riType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if iType.GetValue() != riType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}

	var iValPtr *int
	iPtrType, iPtrErr := NewType(reflect.TypeOf(iValPtr))
	if iPtrErr != nil {
		t.Errorf("NewType failed, err:%s", iPtrErr.Error())
		return
	}
	if iPtrType.GetValue() != model.TypeIntegerValue {
		t.Errorf("get int type value failed. is integer value")
		return
	}
	if !iPtrType.IsPtrType() {
		t.Errorf("get int type value failed. is ptr")
		return
	}
	if !iPtrType.IsBasic() {
		t.Errorf("get int type value failed. is base")
		return
	}

	valPtr, _ := iPtrType.Interface(nil)
	switch valPtr.Interface().(type) {
	case *int:
	default:
		t.Errorf("get int type value failed. type:%s", valPtr.Get().(reflect.Value).Type().String())
		return
	}

	valPtr, _ = iPtrType.Interface(100)
	switch valPtr.Get().(reflect.Value).Interface().(type) {
	case *int:
	default:
		t.Errorf("get int type value failed. type:%s", valPtr.Get().(reflect.Value).Type().String())
		return
	}

	var iAny any
	iAny = 123
	valPtr, _ = iPtrType.Interface(iAny)
	switch valPtr.Get().(reflect.Value).Interface().(type) {
	case *int:
	default:
		t.Errorf("get int type value failed. type:%s", valPtr.Get().(reflect.Value).Type().String())
		return
	}

	valPtr.Set(reflect.ValueOf(&iVal))

	iAny = valPtr.Interface()
	switch iAny.(type) {
	case *int:
	default:
		t.Errorf("get int type value failed. type:%s", valPtr.Get().(reflect.Value).Type().String())
		return
	}

	intPtr, intOK := iAny.(*int)
	if !intOK {
		t.Errorf("get int type value failed. type:%s", valPtr.Get().(reflect.Value).Type().String())
		return
	}
	if *intPtr != iVal {
		t.Errorf("get int type value failed. type:%s", valPtr.Get().(reflect.Value).Type().String())
		return
	}

}

func TestFloatType(t *testing.T) {
	var fVal float32
	fType, fErr := NewType(reflect.TypeOf(fVal))
	if fErr != nil {
		t.Errorf("NewType failed, err:%s", fErr.Error())
		return
	}

	if fType.GetValue() != model.TypeFloatValue {
		t.Errorf("get float type value failed.")
		return
	}

	nVal, _ := fType.Interface(nil)
	newVal := nVal.Get().(reflect.Value)
	rfType, rfErr := NewType(newVal.Type())
	if rfErr != nil {
		t.Errorf("NewType failed, err:%s", rfErr.Error())
		return
	}

	if rfType.GetValue() != model.TypeFloatValue {
		t.Errorf("get float type value failed.")
		return
	}
	if fType.GetName() != rfType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if fType.GetPkgPath() != rfType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if fType.GetValue() != rfType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}
}

func TestBoolType(t *testing.T) {
	var bVal bool
	bType, bErr := NewType(reflect.TypeOf(bVal))
	if bErr != nil {
		t.Errorf("NewType failed, err:%s", bErr.Error())
		return
	}

	if bType.GetValue() != model.TypeBooleanValue {
		t.Errorf("get bool type value failed.")
		return
	}

	nVal, _ := bType.Interface(nil)
	newVal := nVal.Get().(reflect.Value)
	rbType, rbErr := NewType(newVal.Type())
	if rbErr != nil {
		t.Errorf("NewType failed, err:%s", rbErr.Error())
		return
	}

	if rbType.GetValue() != model.TypeBooleanValue {
		t.Errorf("get bool type value failed.")
		return
	}
	if bType.GetName() != rbType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if bType.GetPkgPath() != rbType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if bType.GetValue() != rbType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}
}

func TestStringType(t *testing.T) {
	var strVal string
	strType, strErr := NewType(reflect.TypeOf(strVal))
	if strErr != nil {
		t.Errorf("NewType failed, err:%s", strErr.Error())
		return
	}

	if strType.GetValue() != model.TypeStringValue {
		t.Errorf("get string type value failed.")
		return
	}

	nVal, _ := strType.Interface(nil)
	newVal := nVal.Get().(reflect.Value)
	rstrType, rstrErr := NewType(newVal.Type())
	if rstrErr != nil {
		t.Errorf("NewType failed, err:%s", rstrErr.Error())
		return
	}

	if rstrType.GetValue() != model.TypeStringValue {
		t.Errorf("get string type value failed.")
		return
	}
	if strType.GetName() != rstrType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if strType.GetPkgPath() != rstrType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if strType.GetValue() != rstrType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}
}

func TestDateTimeType(t *testing.T) {
	var dtVal time.Time
	dtType, dtErr := NewType(reflect.TypeOf(dtVal))
	if dtErr != nil {
		t.Errorf("NewType failed, err:%s", dtErr.Error())
		return
	}

	if dtType.GetValue() != model.TypeDateTimeValue {
		t.Errorf("get DateTime type value failed.")
		return
	}

	nVal, _ := dtType.Interface(nil)
	newVal := nVal.Get().(reflect.Value)
	rdtType, rdtErr := NewType(newVal.Type())
	if rdtErr != nil {
		t.Errorf("NewType failed, err:%s", rdtErr.Error())
		return
	}

	if rdtType.GetValue() != model.TypeDateTimeValue {
		t.Errorf("get DateTime type value failed.")
		return
	}
	if dtType.GetName() != rdtType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if dtType.GetPkgPath() != rdtType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if dtType.GetValue() != rdtType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}
}

func TestStructType(t *testing.T) {
	type Base struct {
		iVal int
	}

	var structVal Base
	structType, structErr := NewType(reflect.TypeOf(structVal))
	if structErr != nil {
		t.Errorf("NewType failed, err:%s", structErr.Error())
		return
	}

	if structType.GetValue() != model.TypeStructValue {
		t.Errorf("get DateTime type value failed.")
		return
	}

	nVal, _ := structType.Interface(nil)
	newVal := nVal.Get().(reflect.Value)
	rstructType, rstructErr := NewType(newVal.Type())
	if rstructErr != nil {
		t.Errorf("NewType failed, err:%s", rstructErr.Error())
		return
	}

	if rstructType.GetValue() != model.TypeStructValue {
		t.Errorf("get DateTime type value failed.")
		return
	}
	if structType.GetName() != rstructType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if structType.GetPkgPath() != rstructType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if structType.GetValue() != rstructType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}

	if structType.IsPtrType() {
		t.Errorf("unexpected isPtrType")
		return
	}
}

func TestSliceType(t *testing.T) {
	var sliceVal []*uint16
	sliceType, sliceErr := NewType(reflect.TypeOf(sliceVal))
	if sliceErr != nil {
		t.Errorf("NewType failed, err:%s", sliceErr.Error())
		return
	}

	if sliceType.GetValue() != model.TypeSliceValue {
		t.Errorf("get Slice type value failed.")
		return
	}

	nVal, _ := sliceType.Interface(nil)
	newVal := nVal.Get().(reflect.Value)
	rsliceType, rsliceErr := NewType(newVal.Type())
	if rsliceErr != nil {
		t.Errorf("NewType failed, err:%s", rsliceErr.Error())
		return
	}

	if rsliceType.GetValue() != model.TypeSliceValue {
		t.Errorf("get Slice type value failed.")
		return
	}
	if sliceType.GetName() != rsliceType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if sliceType.GetPkgPath() != rsliceType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if sliceType.GetValue() != rsliceType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}

	dependType := sliceType.Elem()
	if dependType == nil {
		t.Errorf("illegal depend")
		return
	}

	if dependType.GetValue() != model.TypePositiveSmallIntegerValue {
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

	if elemType.GetValue() != model.TypePositiveSmallIntegerValue {
		t.Errorf("illegal elem type value")
		return
	}

	if !elemType.IsPtrType() {
		t.Errorf("illegal elem type value")
		return
	}

	if sliceType.IsPtrType() {
		t.Errorf("unexpected isPtrType")
		return
	}
}

func TestPtrSliceType(t *testing.T) {
	var sliceVal *[]*uint16
	sliceType, sliceErr := NewType(reflect.TypeOf(sliceVal))
	if sliceErr != nil {
		t.Errorf("NewType failed, err:%s", sliceErr.Error())
		return
	}

	if sliceType.GetValue() != model.TypeSliceValue {
		t.Errorf("get Slice type value failed.")
		return
	}

	nVal, _ := sliceType.Interface(nil)
	newVal := nVal.Get().(reflect.Value)
	rsliceType, rsliceErr := NewType(newVal.Type())
	if rsliceErr != nil {
		t.Errorf("NewType failed, err:%s", rsliceErr.Error())
		return
	}

	if rsliceType.GetValue() != model.TypeSliceValue {
		t.Errorf("get Slice type value failed.")
		return
	}
	if sliceType.GetName() != rsliceType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if sliceType.GetPkgPath() != rsliceType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if sliceType.GetValue() != rsliceType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}

	dependType := sliceType.Elem()
	if dependType == nil {
		t.Errorf("illegal depend")
		return
	}

	if dependType.GetValue() != model.TypePositiveSmallIntegerValue {
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

	if elemType.GetValue() != model.TypePositiveSmallIntegerValue {
		t.Errorf("illegal elem type value")
		return
	}

	if !elemType.IsPtrType() {
		t.Errorf("illegal elem type value")
		return
	}

	if !sliceType.IsPtrType() {
		t.Errorf("unexpected isPtrType")
		return
	}
}

func TestSliceStructType(t *testing.T) {
	type Base struct {
		iVal int
	}
	var sliceVal []*Base
	sliceType, sliceErr := NewType(reflect.TypeOf(sliceVal))
	if sliceErr != nil {
		t.Errorf("NewType failed, err:%s", sliceErr.Error())
		return
	}

	if sliceType.GetValue() != model.TypeSliceValue {
		t.Errorf("get Slice type value failed.")
		return
	}

	nVal, _ := sliceType.Interface(nil)
	newVal := nVal.Get().(reflect.Value)
	rsliceType, rsliceErr := NewType(newVal.Type())
	if rsliceErr != nil {
		t.Errorf("NewType failed, err:%s", rsliceErr.Error())
		return
	}

	if rsliceType.GetValue() != model.TypeSliceValue {
		t.Errorf("get Slice type value failed.")
		return
	}
	if sliceType.GetName() != rsliceType.GetName() {
		t.Errorf("NewType faild. illegal type name")
		return
	}
	if sliceType.GetPkgPath() != rsliceType.GetPkgPath() {
		t.Errorf("NewType faild. illegal type pkgPath")
		return
	}
	if sliceType.GetValue() != rsliceType.GetValue() {
		t.Errorf("NewType faild. illegal type value")
		return
	}

	dependType := sliceType.Elem()
	if dependType == nil {
		t.Errorf("illegal depend")
		return
	}

	if dependType.GetValue() != model.TypeStructValue {
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

	if elemType.GetValue() != model.TypeStructValue {
		t.Errorf("illegal elem type value")
		return
	}

	if !elemType.IsPtrType() {
		t.Errorf("illegal elem type value")
		return
	}
	if sliceType.IsPtrType() {
		t.Errorf("unexpected isPtrType")
		return
	}
}

func TestTypeImpl_Interface(t *testing.T) {
	var iVal int
	iType, iErr := NewType(reflect.TypeOf(iVal))
	if iErr != nil {
		t.Errorf("NewType failed, err:%s", iErr.Error())
		return
	}

	tVal, _ := iType.Interface(nil)
	if !tVal.Get().(reflect.Value).CanSet() || !tVal.Get().(reflect.Value).CanAddr() {
		t.Errorf("Interface value failed")
		return
	}
	if tVal.Get().(reflect.Value).Type().String() != "int" {
		t.Errorf("Interface value failed")
		return
	}

	iValPtr := &iVal
	iType, iErr = NewType(reflect.TypeOf(iValPtr))
	if iErr != nil {
		t.Errorf("NewType failed, err:%s", iErr.Error())
		return
	}

	tValPtr, _ := iType.Interface(nil)
	if tValPtr.Get().(reflect.Value).Type().String() != "*int" {
		t.Errorf("Interface value failed")
		return
	}
}
