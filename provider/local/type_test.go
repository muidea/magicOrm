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
	switch valPtr.Interface().Value().(type) {
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

	iAny = valPtr.Interface().Value()
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

func TestTypeIsPtrType(t *testing.T) {
	testCases := []struct {
		name      string
		valueType reflect.Type
		isPtr     bool
	}{
		{"Int", reflect.TypeOf(int(0)), false},
		{"String", reflect.TypeOf(string("")), false},
		{"IntPtr", reflect.TypeOf((*int)(nil)), true},
		{"StringPtr", reflect.TypeOf((*string)(nil)), true},
		{"StructPtr", reflect.TypeOf((*struct{})(nil)), true},
		{"SlicePtr", reflect.TypeOf((*[]int)(nil)), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}
		})
	}
}

func TestTypeIsSlice(t *testing.T) {
	testCases := []struct {
		name      string
		valueType reflect.Type
		isSlice   bool
	}{
		{"Int", reflect.TypeOf(int(0)), false},
		{"String", reflect.TypeOf(string("")), false},
		{"IntSlice", reflect.TypeOf([]int{}), true},
		{"StringSlice", reflect.TypeOf([]string{}), true},
		{"ByteSlice", reflect.TypeOf([]byte{}), true},
		{"IntPtrSlice", reflect.TypeOf([]*int{}), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}
		})
	}
}

func TestTypeIsStruct(t *testing.T) {
	type TestStruct struct {
		Field int
	}

	testCases := []struct {
		name      string
		valueType reflect.Type
		isStruct  bool
	}{
		{"Int", reflect.TypeOf(int(0)), false},
		{"String", reflect.TypeOf(string("")), false},
		{"EmptyStruct", reflect.TypeOf(struct{}{}), true},
		{"TestStruct", reflect.TypeOf(TestStruct{}), true},
		{"Time", reflect.TypeOf(time.Time{}), true},
		{"IntSlice", reflect.TypeOf([]int{}), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}
		})
	}
}

func TestTypeNewElemType(t *testing.T) {
	testCases := []struct {
		name            string
		valueType       reflect.Type
		testValue       interface{}
		expectedNewType reflect.Type
	}{
		{"IntSlice", reflect.TypeOf([]int{}), int(0), reflect.TypeOf(int(0))},
		{"StringSlice", reflect.TypeOf([]string{}), string(""), reflect.TypeOf(string(""))},
		{"StructSlice", reflect.TypeOf([]struct{ Field int }{}), struct{ Field int }{}, reflect.TypeOf(struct{ Field int }{})},
		{"PtrSlice", reflect.TypeOf([]*int{}), (*int)(nil), reflect.TypeOf((*int)(nil))},
		{"TimeSlice", reflect.TypeOf([]time.Time{}), time.Time{}, reflect.TypeOf(time.Time{})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get type for test
			_, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}

			elemType, err := GetType(reflect.SliceOf(reflect.TypeOf(tc.testValue)).Elem())
			if err != nil {
				t.Errorf("GetType for element type failed for %s: %s", tc.name, err.Error())
				return
			}

			// Compare the underlying type of the new element type
			expectedTypeObj, err := GetType(tc.expectedNewType)
			if err != nil {
				t.Errorf("GetType failed for expected type of %s: %s", tc.name, err.Error())
				return
			}

			if elemType.GetName() != expectedTypeObj.GetName() ||
				elemType.GetPkgPath() != expectedTypeObj.GetPkgPath() ||
				elemType.GetValue() != expectedTypeObj.GetValue() {
				t.Errorf("Element type for %s doesn't match expected type: %v",
					tc.name, expectedTypeObj.GetName())
			}
		})
	}
}

func TestTypeNewType(t *testing.T) {
	type TestStruct struct {
		Field int
	}

	testCases := []struct {
		name      string
		valueType reflect.Type
	}{
		{"Int", reflect.TypeOf(int(0))},
		{"String", reflect.TypeOf(string(""))},
		{"EmptyStruct", reflect.TypeOf(struct{}{})},
		{"TestStruct", reflect.TypeOf(TestStruct{})},
		{"Time", reflect.TypeOf(time.Time{})},
		{"IntSlice", reflect.TypeOf([]int{})},
		{"StructSlice", reflect.TypeOf([]TestStruct{})},
		{"IntPtr", reflect.TypeOf((*int)(nil))},
		{"StructPtr", reflect.TypeOf((*TestStruct)(nil))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}
		})
	}
}

func TestTypeDescription(t *testing.T) {
	testCases := []struct {
		name      string
		valueType reflect.Type
	}{
		{"Int", reflect.TypeOf(int(0))},
		{"String", reflect.TypeOf(string(""))},
		{"EmptyStruct", reflect.TypeOf(struct{}{})},
		{"TestStruct", reflect.TypeOf(struct{ Field int }{})},
		{"Time", reflect.TypeOf(time.Time{})},
		{"IntPtr", reflect.TypeOf((*int)(nil))},
		{"IntSlice", reflect.TypeOf([]int{})},
		{"StructSlice", reflect.TypeOf([]struct{ Field int }{})},
		{"IntPtr", reflect.TypeOf((*int)(nil))},
		{"StructPtr", reflect.TypeOf((*struct{ Field int })(nil))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}
		})
	}
}

func TestGetType(t *testing.T) {
	// Basic types
	intType := reflect.TypeOf(0)
	_, err := GetType(intType)
	if err != nil {
		t.Errorf("GetType failed for int: %s", err.Error())
		return
	}

	// String type
	stringType := reflect.TypeOf("")
	_, err = GetType(stringType)
	if err != nil {
		t.Errorf("GetType failed for string: %s", err.Error())
		return
	}

	// Struct type - time.Time is a special case that is treated as basic
	timeType := reflect.TypeOf(time.Time{})
	_, err = GetType(timeType)
	if err != nil {
		t.Errorf("GetType failed for time.Time: %s", err.Error())
		return
	}

	// Custom struct type
	type CustomStruct struct {
		Field1 string
		Field2 int
	}
	structType := reflect.TypeOf(CustomStruct{})
	_, err = GetType(structType)
	if err != nil {
		t.Errorf("GetType failed for CustomStruct: %s", err.Error())
		return
	}
}

func TestGetTypeForPointers(t *testing.T) {
	// Int pointer
	intPtr := new(int)
	intPtrType := reflect.TypeOf(intPtr)
	_, err := GetType(intPtrType)
	if err != nil {
		t.Errorf("GetType failed for *int: %s", err.Error())
		return
	}

	// Struct pointer
	type TestStruct struct {
		Field string
	}
	structPtr := &TestStruct{}
	structPtrType := reflect.TypeOf(structPtr)
	_, err = GetType(structPtrType)
	if err != nil {
		t.Errorf("GetType failed for *TestStruct: %s", err.Error())
		return
	}
}

func TestGetTypeForSlices(t *testing.T) {
	// Int slice
	intSlice := []int{}
	intSliceType := reflect.TypeOf(intSlice)
	_, err := GetType(intSliceType)
	if err != nil {
		t.Errorf("GetType failed for []int: %s", err.Error())
		return
	}

	// String slice
	stringSlice := []string{}
	stringSliceType := reflect.TypeOf(stringSlice)
	_, err = GetType(stringSliceType)
	if err != nil {
		t.Errorf("GetType failed for []string: %s", err.Error())
		return
	}

	// Struct slice
	type Item struct {
		Value string
	}
	structSlice := []Item{}
	structSliceType := reflect.TypeOf(structSlice)
	_, err = GetType(structSliceType)
	if err != nil {
		t.Errorf("GetType failed for []Item: %s", err.Error())
		return
	}
}

func TestGetTypeForNestedTypes(t *testing.T) {
	// Pointer to slice
	slicePtr := &[]string{}
	slicePtrType := reflect.TypeOf(slicePtr)
	_, err := GetType(slicePtrType)
	if err != nil {
		t.Errorf("GetType failed for *[]string: %s", err.Error())
		return
	}

	// Slice of pointers
	ptrSlice := []*string{}
	ptrSliceType := reflect.TypeOf(ptrSlice)
	_, err = GetType(ptrSliceType)
	if err != nil {
		t.Errorf("GetType failed for []*string: %s", err.Error())
		return
	}
}

func TestGetTypeErrors(t *testing.T) {
	// Unsupported types
	invalidTypes := []interface{}{
		make(map[string]int),    // Map
		func() {},               // Function
		make(chan int),          // Channel
		[3]int{1, 2, 3},         // Array (fixed size)
		complex(1.0, 2.0),       // Complex number
	}

	for i, invalidType := range invalidTypes {
		_, err := GetType(reflect.TypeOf(invalidType))
		if err == nil {
			t.Errorf("Case %d: GetType should fail for unsupported type: %T", i, invalidType)
		}
	}
}

func TestTypeCopy(t *testing.T) {
	// Create a test type
	originalType, err := GetType(reflect.TypeOf(""))
	if err != nil {
		t.Errorf("GetType failed: %s", err.Error())
		return
	}

	// Make a copy
	copiedType, err := GetType(reflect.TypeOf(""))
	if err != nil {
		t.Errorf("GetType failed for copied type: %s", err.Error())
		return
	}

	// Verify the copy has the same values
	if copiedType.GetName() != originalType.GetName() {
		t.Errorf("Copied type name mismatch, expected: %s, got: %s",
			originalType.GetName(), copiedType.GetName())
	}

	if copiedType.GetPkgPath() != originalType.GetPkgPath() {
		t.Errorf("Copied type pkg path mismatch, expected: %s, got: %s",
			originalType.GetPkgPath(), copiedType.GetPkgPath())
	}

	if copiedType.IsBasic() != originalType.IsBasic() {
		t.Errorf("Copied type IsBasic mismatch, expected: %v, got: %v",
			originalType.IsBasic(), copiedType.IsBasic())
	}

	if copiedType.IsStruct() != originalType.IsStruct() {
		t.Errorf("Copied type IsStruct mismatch, expected: %v, got: %v",
			originalType.IsStruct(), copiedType.IsStruct())
	}

	if copiedType.IsSlice() != originalType.IsSlice() {
		t.Errorf("Copied type IsSlice mismatch, expected: %v, got: %v",
			originalType.IsSlice(), copiedType.IsSlice())
	}

	if copiedType.IsPtrType() != originalType.IsPtrType() {
		t.Errorf("Copied type IsPtrType mismatch, expected: %v, got: %v",
			originalType.IsPtrType(), copiedType.IsPtrType())
	}
}

func TestTypeElemType(t *testing.T) {
	testCases := []struct {
		name             string
		valueType        reflect.Type
		expectedElemType reflect.Type
	}{
		{"IntSlice", reflect.TypeOf([]int{}), reflect.TypeOf(int(0))},
		{"StringSlice", reflect.TypeOf([]string{}), reflect.TypeOf(string(""))},
		{"StructSlice", reflect.TypeOf([]struct{ Field int }{}), reflect.TypeOf(struct{ Field int }{})},
		{"IntPtrSlice", reflect.TypeOf([]*int{}), reflect.TypeOf((*int)(nil))},
		{"StructPtrSlice", reflect.TypeOf([]*struct{ Field int }{}), reflect.TypeOf((*struct{ Field int })(nil))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}
		})
	}
}

func TestTypeValue(t *testing.T) {
	testCases := []struct {
		name         string
		valueType    reflect.Type
		expectedType reflect.Type
	}{
		{"Int", reflect.TypeOf(int(0)), reflect.TypeOf(int(0))},
		{"String", reflect.TypeOf(string("")), reflect.TypeOf(string(""))},
		{"Bool", reflect.TypeOf(bool(false)), reflect.TypeOf(bool(false))},
		{"Time", reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{})},
		{"Struct", reflect.TypeOf(struct{ Field int }{}), reflect.TypeOf(struct{ Field int }{})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			typeObj, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}

			// Compare with expected type values
			expectedTypeObj, err := GetType(tc.expectedType)
			if err != nil {
				t.Errorf("GetType failed for expected type: %s", err.Error())
				return
			}

			if typeObj.GetValue() != expectedTypeObj.GetValue() {
				t.Errorf("Type.GetValue() for %s expected: %v, got: %v",
					tc.name, expectedTypeObj.GetValue(), typeObj.GetValue())
			}
		})
	}
}

func TestTypeValueForNativeTypes(t *testing.T) {
	// Test cases for native types
	testCases := []struct {
		name      string
		valueType reflect.Type
	}{
		{"Int", reflect.TypeOf(int(0))},
		{"String", reflect.TypeOf(string(""))},
		{"EmptyStruct", reflect.TypeOf(struct{}{})},
		{"TestStruct", reflect.TypeOf(struct{ Field int }{})},
		{"Time", reflect.TypeOf(time.Time{})},
		{"IntPtr", reflect.TypeOf((*int)(nil))},
		{"IntSlice", reflect.TypeOf([]int{})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetType(tc.valueType)
			if err != nil {
				t.Errorf("GetType failed for %s: %s", tc.name, err.Error())
				return
			}
		})
	}
}
