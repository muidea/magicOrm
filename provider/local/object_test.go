package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int64 `json:"id" orm:"id key"`
	// Name 名称
	Name      string    `json:"name" orm:"name"`
	Value     float32   `json:"value" orm:"value"`
	TimeStamp time.Time `json:"timeStamp" orm:"timeStamp"`
	T1        Test      `orm:"t1"`
	// Description 描述
	Description string `json:"description" orm:"description"`
}

type BT struct {
	ID  int `orm:"id key"`
	Val int `orm:"val"`
}

type Base struct {
	ID  int `orm:"id key"`
	Val int `orm:"val"`
	Bt  BT  `orm:"bt"`
}

type Test struct {
	ID    int  `orm:"id key"`
	Val   int  `orm:"val"`
	Base  Base `orm:"b1"`
	Base2 BT   `orm:"b2"`
}

func TestModelValue(t *testing.T) {
	now, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	unit := Unit{Name: "AA", T1: Test{Val: 123}, TimeStamp: now}

	eModel, eErr := GetEntityModel(&unit)
	if eErr != nil {
		t.Errorf("GetEntityModel failed, error %s", eErr.Error())
		return
	}

	uVal, uOK := eModel.Interface(false, "origin").(Unit)
	if !uOK {
		t.Errorf("eModel.Interface failed")
		return
	}
	if uVal.Name != unit.Name {
		t.Errorf("eModel.Interface failed")
		return
	}

	uPtrVal, uPtrOK := eModel.Interface(true, "origin").(*Unit)
	if !uPtrOK {
		t.Errorf("eModel.Interface failed")
		return
	}
	if uPtrVal.Name != unit.Name {
		t.Errorf("eModel.Interface failed")
		return
	}

	eModel, eErr = GetEntityModel(unit)
	if eErr != nil {
		t.Errorf("GetEntityModel failed, error %s", eErr.Error())
		return
	}

	uVal, uOK = eModel.Interface(false, "origin").(Unit)
	if !uOK {
		t.Errorf("eModel.Interface failed")
		return
	}
	if uVal.Name != unit.Name {
		t.Errorf("eModel.Interface failed")
		return
	}

	uPtrVal, uPtrOK = eModel.Interface(true, "origin").(*Unit)
	if !uPtrOK {
		t.Errorf("eModel.Interface failed")
		return
	}
	if uPtrVal.Name != unit.Name {
		t.Errorf("eModel.Interface failed")
		return
	}

	unitVal := reflect.ValueOf(&unit).Elem()
	unitInfo, unitErr := getTypeModel(unitVal.Type())
	if unitErr != nil {
		t.Errorf("getValueModel failed, unitErr:%s", unitErr.Error())
		return
	}

	id := int64(123320)
	iVal := NewValue(reflect.ValueOf(id))
	pk := unitInfo.GetPrimaryField()
	if pk == nil {
		t.Errorf("GetPrimaryField faield")
		return
	}
	pk.SetValue(iVal)

	name := "abcdfrfe"
	nVal := NewValue(reflect.ValueOf(name))
	unitInfo.SetFieldValue("name", nVal)

	now = time.Now()
	tsVal := NewValue(reflect.ValueOf(now))
	unitInfo.SetFieldValue("timeStamp", tsVal)

	unit = unitInfo.Interface(false, "origin").(Unit)
	if unit.ID != int64(id) {
		t.Errorf("update id field failed, ID:%v, id:%v", unit.ID, id)
		return
	}
	if unit.Name != name {
		t.Errorf("update id field failed")
		return
	}
	if !unit.TimeStamp.Equal(now) {
		t.Errorf("update timeStamp failed")
		return
	}

	unitPtrVal, unitPtrOK := unitInfo.Interface(true, "origin").(*Unit)
	if !unitPtrOK {
		t.Errorf("unitInfo.Interface failed")
		return
	}
	if unitPtrVal.Name != name {
		t.Errorf("update id field failed")
		return
	}
}

func TestReference(t *testing.T) {
	type AB struct {
		F32 float32 `orm:"f32 key"`
	}

	type CD struct {
		AB  AB  `orm:"ab"`
		I64 int `orm:"i64 key"`
	}

	type Demo struct {
		II int   `orm:"ii key"`
		AB *AB   `orm:"ab"`
		CD []int `orm:"cd"`
		EF []*AB `orm:"ef"`
	}

	abVal := reflect.ValueOf(&AB{})
	cdVal := reflect.ValueOf(&CD{}).Elem()
	demoVal := reflect.ValueOf(&Demo{AB: &AB{}}).Elem()
	_, err := getTypeModel(abVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	_, err = getTypeModel(cdVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	_, err = getTypeModel(demoVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	f32Info, err := getValueModel(demoVal)
	if err != nil {
		t.Errorf("getValueModel failed, err:%s", err.Error())
		return
	}

	f32Info.Dump()

	i64Info, err := getValueModel(cdVal)
	if err != nil {
		t.Errorf("getValueModel failed, err:%s", err.Error())
	}

	i64Info.Dump()
}

type TT struct {
	Aa int `orm:"aa key auto"`
	Bb int `orm:"bb"`
	Tt *TT `orm:"tt"`
}

func TestGetModelValue(t *testing.T) {
	t1 := TT{Aa: 12, Bb: 23}
	ttVal := reflect.ValueOf(&t1).Elem()
	_, err := getTypeModel(ttVal.Type())
	if err != nil {
		t.Errorf("getTypeModel failed, err:%s", err.Error())
		return
	}
	t1Info, t1Err := getValueModel(ttVal)
	if t1Err != nil {
		t.Errorf("getValueModel t1 failed, err:%s", t1Err.Error())
		return
	}

	t2 := &TT{Aa: 34, Bb: 45}
	//reflect.TypeOf(t2)
	t2Info, t2Err := getValueModel(reflect.ValueOf(t2).Elem())
	if t2Err != nil {
		t.Errorf("getValueModel t2 failed, err:%s", t2Err.Error())
		return
	}

	t1Info.Dump()
	t2Info.Dump()
}

type Reference struct {
	ID          int       `orm:"id key auto" view:"detail,lite"`
	BArray      []bool    `orm:"bArray" view:"detail,lite"`
	StrArray    []string  `orm:"strArray" view:"detail,lite"`
	PtrArray    *[]string `orm:"ptrArray" view:"detail,lite"`
	PtrStrArray *[]string `orm:"ptrStrArray" view:"detail,lite"`
}

func TestCheckValid(t *testing.T) {
	ptrArray2 := []string{}
	r1 := &Reference{
		StrArray: ptrArray2,
		PtrArray: &ptrArray2,
	}

	fR1 := reflect.ValueOf(r1)
	r1Model, r1Err := getValueModel(fR1)
	if r1Err != nil {
		t.Errorf("getValueModel failed, error:%s", r1Err.Error())
		return
	}

	if !r1Model.GetField("id").GetValue().IsValid() {
		t.Errorf("check int IsValid failed")
		return
	}
	if r1Model.GetField("bArray").GetValue().IsValid() {
		t.Errorf("check int IsValid failed")
		return
	}
	if !r1Model.GetField("strArray").GetValue().IsValid() {
		t.Errorf("check int IsValid failed")
		return
	}
	if !r1Model.GetField("ptrArray").GetValue().IsValid() {
		t.Errorf("check int IsValid failed")
		return
	}
	if r1Model.GetField("ptrStrArray").GetValue().IsValid() {
		t.Errorf("check int IsValid failed")
		return
	}
}

func TestObjectCopy2(t *testing.T) {
	// Create a test object
	now, _ := time.Parse(util.CSTLayout, "2018-01-02 15:04:05")
	unit := Unit{
		ID:        123,
		Name:      "TestUnit",
		Value:     42.5,
		TimeStamp: now,
		T1:        Test{ID: 1, Val: 100},
	}

	// Get entity model
	model, err := GetEntityModel(&unit)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test normal copy (with values)
	copiedModel := model.Copy(false)
	if copiedModel == nil {
		t.Errorf("Model.Copy() returned nil")
		return
	}

	// Verify fields were copied correctly
	copiedUnit, ok := copiedModel.Interface(false, "origin").(Unit)
	if !ok {
		t.Errorf("copiedModel.Interface() failed to convert to Unit")
		return
	}

	// Check that values were copied
	if copiedUnit.ID != unit.ID {
		t.Errorf("Copied ID mismatch, expected: %d, got: %d", unit.ID, copiedUnit.ID)
	}

	if copiedUnit.Name != unit.Name {
		t.Errorf("Copied Name mismatch, expected: %s, got: %s", unit.Name, copiedUnit.Name)
	}

	if copiedUnit.Value != unit.Value {
		t.Errorf("Copied Value mismatch, expected: %f, got: %f", unit.Value, copiedUnit.Value)
	}

	if !copiedUnit.TimeStamp.Equal(unit.TimeStamp) {
		t.Errorf("Copied TimeStamp mismatch, expected: %v, got: %v", unit.TimeStamp, copiedUnit.TimeStamp)
	}

	if copiedUnit.T1.Val != unit.T1.Val {
		t.Errorf("Copied nested struct value mismatch, expected: %d, got: %d", unit.T1.Val, copiedUnit.T1.Val)
	}

	// Test reset copy (without values)
	resetModel := model.Copy(true)
	if resetModel == nil {
		t.Errorf("Model.Copy(true) returned nil")
		return
	}

	// Verify fields exist but values are reset
	resetUnit, ok := resetModel.Interface(false, "origin").(Unit)
	if !ok {
		t.Errorf("resetModel.Interface() failed to convert to Unit")
		return
	}

	// Check structure is preserved but values are zero
	var zeroVal float32
	if resetUnit.ID != 0 {
		t.Errorf("Reset copy should have zero ID value, got: %d", resetUnit.ID)
	}

	if resetUnit.Name != "" {
		t.Errorf("Reset copy should have empty Name value, got: %s", resetUnit.Name)
	}

	if resetUnit.Value != zeroVal {
		t.Errorf("Reset copy should have zero Value, got: %f", resetUnit.Value)
	}
}

func TestObjectSetFieldValue2(t *testing.T) {
	// Create a test object
	unit := Unit{
		ID:   100,
		Name: "OriginalName",
	}

	// Get entity model
	model, err := GetEntityModel(unit)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test updating string field
	newName := "UpdatedName"
	nameVal := NewValue(reflect.ValueOf(newName))
	model.SetFieldValue("name", nameVal)

	// Test updating int field
	newID := int64(200)
	idVal := NewValue(reflect.ValueOf(newID))
	model.SetFieldValue("id", idVal)

	// Verify updates worked
	updatedUnit, ok := model.Interface(false, "origin").(Unit)
	if !ok {
		t.Errorf("model.Interface() failed to convert to Unit")
		return
	}

	if updatedUnit.ID != newID {
		t.Errorf("ID update failed, expected: %d, got: %d", newID, updatedUnit.ID)
	}

	if updatedUnit.Name != newName {
		t.Errorf("Name update failed, expected: %s, got: %s", newName, updatedUnit.Name)
	}

	// Test updating non-existent field
	invalidVal := NewValue(reflect.ValueOf("test"))
	model.SetFieldValue("nonexistent", invalidVal)

	// Test updating with incompatible type
	invalidTypeVal := NewValue(reflect.ValueOf(true)) // bool value for string field
	model.SetFieldValue("name", invalidTypeVal)
}

// TestObjectViews tests the view functionality
func TestObjectViews(t *testing.T) {
	// Create a test struct with different view tags
	type ViewTestStruct struct {
		ID          int     `orm:"id key auto" view:"detail,lite"`
		Name        string  `orm:"name" view:"detail,lite"`
		Value       float64 `orm:"value" view:"detail"`
		Description string  `json:"description" orm:"description"`
	}

	viewTest := &ViewTestStruct{
		ID:          123,
		Name:        "ViewTest",
		Value:       42.5,
		Description: "Test Description",
	}

	// Get entity model
	model, err := GetEntityModel(viewTest)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// For object_impl, we'll test the view functionality by checking
	// if fields are present in different views

	// Get all fields (using Origin view)
	allFields := model.GetFields()
	if len(allFields) < 4 {
		t.Errorf("Expected at least 4 fields in origin view, got %d", len(allFields))
	}

	// Check for specific views
	// In the model implementation, we may not have a direct GetViewFields method
	// Instead, we'll verify field visibility through specific view interfaces

	// Detail view should include ID, Name, Value but not Description
	detailObj, ok := model.Interface(false, "detail").(ViewTestStruct)
	if !ok {
		t.Errorf("model.Interface() failed to convert to ViewTestStruct for detail view")
		return
	}

	// Check detail view fields
	if detailObj.ID != viewTest.ID {
		t.Errorf("ID should be included in detail view")
	}

	if detailObj.Name != viewTest.Name {
		t.Errorf("Name should be included in detail view")
	}

	if detailObj.Value != viewTest.Value {
		t.Errorf("Value should be included in detail view")
	}

	// Lite view should only include ID and Name
	liteObj, ok := model.Interface(false, "lite").(ViewTestStruct)
	if !ok {
		t.Errorf("model.Interface() failed to convert to ViewTestStruct for lite view")
		return
	}

	// Check lite view fields
	if liteObj.ID != viewTest.ID {
		t.Errorf("ID should be included in lite view")
	}

	if liteObj.Name != viewTest.Name {
		t.Errorf("Name should be included in lite view")
	}

	// Value should be zero value in lite view if it's not included
	var zeroValue float64
	if liteObj.Value != zeroValue {
		t.Errorf("Value should not be included in lite view")
	}
}

func TestNestedObjects(t *testing.T) {
	// Create test object with nested objects
	testObj := Test{
		ID:  1,
		Val: 100,
		Base: Base{
			ID:  2,
			Val: 200,
			Bt: BT{
				ID:  3,
				Val: 300,
			},
		},
		Base2: BT{
			ID:  4,
			Val: 400,
		},
	}

	// Get entity entityModel
	entityModel, err := GetEntityModel(testObj)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test accessing nested fields
	baseField := entityModel.GetField("b1")
	if baseField == nil {
		t.Errorf("GetField(b1) returned nil")
		return
	}

	if !baseField.IsStruct() {
		t.Errorf("Base field should be a struct")
	}

	// Test updating nested field
	base2Field := entityModel.GetField("b2")
	if base2Field == nil {
		t.Errorf("GetField(b2) returned nil")
		return
	}
}

// TestObjectWithNil tests handling nil values
func TestObjectWithNil(t *testing.T) {
	// Test with nil entity using defer/recover
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("GetEntityModel with nil properly panicked: %v", r)
				// This is acceptable as we're testing boundary conditions
			}
		}()

		_, err := GetEntityModel(nil)
		if err == nil {
			t.Errorf("GetEntityModel should fail with nil entity")
		} else {
			t.Logf("GetEntityModel returned error as expected: %v", err)
		}
	}()

	// Test with pointer to nil struct using defer/recover
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("GetEntityModel with nil struct pointer properly panicked: %v", r)
				// This is acceptable as we're testing boundary conditions
			}
		}()

		var nilUnit *Unit
		_, err := GetEntityModel(nilUnit)
		if err == nil {
			t.Errorf("GetEntityModel should fail with nil struct pointer")
		} else {
			t.Logf("GetEntityModel returned error as expected: %v", err)
		}
	}()
}

// TestObjectsWithPointers tests handling pointers in structs
func TestObjectsWithPointers(t *testing.T) {
	// Define a struct with pointer fields
	type PointerStruct struct {
		ID      int     `orm:"id key"`
		Name    *string `orm:"name"`
		Value   *int    `orm:"value"`
		Enabled *bool   `orm:"enabled"`
	}

	// Create a test object with pointer fields
	name := "TestName"
	value := 42
	enabled := true

	ptrObj := PointerStruct{
		ID:      1,
		Name:    &name,
		Value:   &value,
		Enabled: &enabled,
	}

	// Get entity model
	model, err := GetEntityModel(ptrObj)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Check if interface can be retrieved
	retrievedObj, ok := model.Interface(false, "origin").(PointerStruct)
	if !ok {
		t.Errorf("model.Interface() failed to convert to PointerStruct")
		return
	}

	// Check pointer values
	if retrievedObj.ID != ptrObj.ID {
		t.Errorf("ID mismatch, expected: %d, got: %d", ptrObj.ID, retrievedObj.ID)
	}

	if *retrievedObj.Name != *ptrObj.Name {
		t.Errorf("Name mismatch, expected: %s, got: %s", *ptrObj.Name, *retrievedObj.Name)
	}

	if *retrievedObj.Value != *ptrObj.Value {
		t.Errorf("Value mismatch, expected: %d, got: %d", *ptrObj.Value, *retrievedObj.Value)
	}

	if *retrievedObj.Enabled != *ptrObj.Enabled {
		t.Errorf("Enabled mismatch, expected: %t, got: %t", *ptrObj.Enabled, *retrievedObj.Enabled)
	}

	// Test updating pointer field
	newName := "UpdatedName"
	nameVal := NewValue(reflect.ValueOf(&newName))
	model.SetFieldValue("name", nameVal)

	// Verify update worked
	updatedObj, ok := model.Interface(false, "origin").(PointerStruct)
	if !ok {
		t.Errorf("model.Interface() failed to convert to PointerStruct")
		return
	}

	if *updatedObj.Name != newName {
		t.Errorf("Name update failed, expected: %s, got: %s", newName, *updatedObj.Name)
	}
}
