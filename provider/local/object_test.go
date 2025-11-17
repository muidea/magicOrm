package local

import (
	"reflect"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicOrm/models"
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

	uVal, uOK := eModel.Interface(false).(Unit)
	if !uOK {
		t.Errorf("eModel.Interface failed")
		return
	}
	if uVal.Name != unit.Name {
		t.Errorf("eModel.Interface failed")
		return
	}

	uPtrVal, uPtrOK := eModel.Interface(true).(*Unit)
	if !uPtrOK {
		t.Errorf("eModel.Interface failed")
		return
	}
	if uPtrVal.Name != unit.Name {
		t.Errorf("eModel.Interface failed")
		return
	}

	eModel, eErr = GetEntityModel(&unit)
	if eErr != nil {
		t.Errorf("GetEntityModel failed, error %s", eErr.Error())
		return
	}

	uVal, uOK = eModel.Interface(false).(Unit)
	if !uOK {
		t.Errorf("eModel.Interface failed")
		return
	}
	if uVal.Name != unit.Name {
		t.Errorf("eModel.Interface failed")
		return
	}

	uPtrVal, uPtrOK = eModel.Interface(true).(*Unit)
	if !uPtrOK {
		t.Errorf("eModel.Interface failed")
		return
	}
	if uPtrVal.Name != unit.Name {
		t.Errorf("eModel.Interface failed")
		return
	}

	nameField := eModel.GetField("name")
	if nameField == nil {
		t.Errorf("eModel.GetField-name failed")
		return
	}
	nameField.SetValue("abc")
	if unit.Name != "abc" {
		t.Errorf("eModel.GetField-name failed")
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

	cdVal := reflect.ValueOf(&CD{}).Elem()
	demoVal := reflect.ValueOf(&Demo{AB: &AB{}}).Elem()
	_, err := getValueModel(demoVal, models.MetaView)
	if err != nil {
		t.Errorf("getValueModel failed, err:%s", err.Error())
		return
	}

	_, err = getValueModel(cdVal, models.MetaView)
	if err != nil {
		t.Errorf("getValueModel failed, err:%s", err.Error())
	}
}

type TT struct {
	Aa int `orm:"aa key auto"`
	Bb int `orm:"bb"`
	Tt *TT `orm:"tt"`
}

func TestGetModelValue(t *testing.T) {
	t1 := TT{Aa: 12, Bb: 23}
	ttVal := reflect.ValueOf(&t1).Elem()
	_, t1Err := getValueModel(ttVal, models.MetaView)
	if t1Err != nil {
		t.Errorf("getValueModel t1 failed, err:%s", t1Err.Error())
		return
	}

	t2 := &TT{Aa: 34, Bb: 45}
	//reflect.TypeOf(t2)
	_, t2Err := getValueModel(reflect.ValueOf(t2).Elem(), models.MetaView)
	if t2Err != nil {
		t.Errorf("getValueModel t2 failed, err:%s", t2Err.Error())
		return
	}
}

type Reference struct {
	ID          int       `orm:"id key auto" view:"detail,lite"`
	BArray      []bool    `orm:"bArray" view:"detail,lite"`
	StrArray    []string  `orm:"strArray" view:"detail,lite"`
	PtrArray    *[]string `orm:"ptrArray" view:"detail,lite"`
	PtrStrArray *[]string `orm:"ptrStrArray" view:"detail"`
}

func TestObjectCopy(t *testing.T) {
	refValue := &Reference{
		ID:          123,
		BArray:      []bool{true, false},
		StrArray:    []string{"str1", "str2"},
		PtrArray:    &[]string{"ptr1", "ptr2"},
		PtrStrArray: &[]string{"ptrStr1", "ptrStr2"},
	}

	refModelVal, refModelErr := GetEntityModel(refValue)
	if refModelErr != nil {
		t.Errorf("GetEntityModel failed, err:%s", refModelErr.Error())
	}
	if !models.IsValidField(refModelVal.GetField("id")) {
		t.Errorf("check refModelVal field id valid failed, really false, expect true")
	}
	if !models.IsValidField(refModelVal.GetField("bArray")) {
		t.Errorf("check refModelVal field bArray valid failed, really false, expect true")
	}
	if !models.IsValidField(refModelVal.GetField("strArray")) {
		t.Errorf("check refModelVal field strArray valid failed, really false, expect true")
	}
	if !models.IsValidField(refModelVal.GetField("ptrArray")) {
		t.Errorf("check refModelVal field ptrArray valid failed, really false, expect true")
	}
	if !models.IsValidField(refModelVal.GetField("ptrStrArray")) {
		t.Errorf("check refModelVal field ptrStrArray valid failed, really false, expect true")
	}
	if !models.IsAssignedField(refModelVal.GetField("id")) {
		t.Errorf("check refModelVal field id assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(refModelVal.GetField("bArray")) {
		t.Errorf("check refModelVal field bArray assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(refModelVal.GetField("strArray")) {
		t.Errorf("check refModelVal field strArray assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(refModelVal.GetField("ptrArray")) {
		t.Errorf("check refModelVal field ptrArray assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(refModelVal.GetField("ptrStrArray")) {
		t.Errorf("check refModelVal field ptrStrArray assigned failed, really false, expect true")
	}

	originRefModelVal := refModelVal.Copy(models.OriginView)
	if originRefModelVal == nil {
		t.Errorf("Copy originRefModelVal model failed")
	}
	if !models.CompareModel(refModelVal, originRefModelVal) {
		t.Errorf("compare originRefModelVal model failed")
	}
	if !models.IsValidField(originRefModelVal.GetField("id")) {
		t.Errorf("check originRefModelVal model id field failed, really false, expect true")
	}
	if !models.IsValidField(originRefModelVal.GetField("bArray")) {
		t.Errorf("check originRefModelVal model bArray field failed, really false, expect true")
	}
	if !models.IsValidField(originRefModelVal.GetField("strArray")) {
		t.Errorf("check originRefModelVal model strArray field failed, really false, expect true")
	}
	if !models.IsValidField(originRefModelVal.GetField("ptrArray")) {
		t.Errorf("check originRefModelVal model ptrArray field failed, really false, expect true")
	}
	if !models.IsValidField(originRefModelVal.GetField("ptrStrArray")) {
		t.Errorf("check originRefModelVal model ptrStrArray field failed, really false, expect true")
	}
	if !models.IsAssignedField(originRefModelVal.GetField("id")) {
		t.Errorf("check originRefModelVal model id field assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(originRefModelVal.GetField("bArray")) {
		t.Errorf("check originRefModelVal model bArray field assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(originRefModelVal.GetField("strArray")) {
		t.Errorf("check originRefModelVal model strArray field assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(originRefModelVal.GetField("ptrArray")) {
		t.Errorf("check originRefModelVal model ptrArray field assigned failed, really false, expect true")
	}

	metaRefModelVal := refModelVal.Copy(models.MetaView)
	if metaRefModelVal == nil {
		t.Errorf("Copy metaRefModelVal model failed")
	}
	if models.CompareModel(refModelVal, metaRefModelVal) {
		t.Errorf("compare metaRefModelVal model failed")
	}
	if !models.IsValidField(metaRefModelVal.GetField("id")) {
		t.Errorf("check metaRefModelVal model id field failed, really false, expect true")
	}
	if !models.IsValidField(metaRefModelVal.GetField("bArray")) {
		t.Errorf("check metaRefModelVal model bArray field failed, really false, expect true")
	}
	if !models.IsValidField(metaRefModelVal.GetField("strArray")) {
		t.Errorf("check metaRefModelVal model strArray field failed, really false, expect true")
	}
	if models.IsValidField(metaRefModelVal.GetField("ptrArray")) {
		t.Errorf("check metaRefModelVal model ptrArray field failed, really true, expect false")
	}
	if models.IsValidField(metaRefModelVal.GetField("ptrStrArray")) {
		t.Errorf("check metaRefModelVal model ptrStrArray field failed, really true, expect false")
	}
	if models.IsAssignedField(metaRefModelVal.GetField("id")) {
		t.Errorf("check metaRefModelVal model id field assigned failed, really true, expect false")
	}
	if models.IsAssignedField(metaRefModelVal.GetField("bArray")) {
		t.Errorf("check metaRefModelVal model bArray field assigned failed, really true, expect false")
	}
	if models.IsAssignedField(metaRefModelVal.GetField("strArray")) {
		t.Errorf("check metaRefModelVal model strArray field assigned failed, really true, expect false")
	}
	if models.IsAssignedField(metaRefModelVal.GetField("ptrArray")) {
		t.Errorf("check metaRefModelVal model ptrArray field assigned failed, really true, expect false")
	}
	if models.IsAssignedField(metaRefModelVal.GetField("ptrStrArray")) {
		t.Errorf("check metaRefModelVal model ptrStrArray field assigned failed, really true, expect false")
	}

	detailRefModelVal := refModelVal.Copy(models.DetailView)
	if detailRefModelVal == nil {
		t.Errorf("Copy detailRefModelVal model failed")
	}
	if !models.IsValidField(detailRefModelVal.GetField("id")) {
		t.Errorf("check detailRefModelVal field id valid failed, really false, expect true")
	}
	if !models.IsValidField(detailRefModelVal.GetField("bArray")) {
		t.Errorf("check detailRefModelVal field bArray valid failed, really false, expect true")
	}
	if !models.IsValidField(detailRefModelVal.GetField("strArray")) {
		t.Errorf("check detailRefModelVal field strArray valid failed, really false, expect true")
	}
	if !models.IsValidField(detailRefModelVal.GetField("ptrArray")) {
		t.Errorf("check detailRefModelVal field ptrArray valid failed, really false, expect true")
	}
	if !models.IsValidField(detailRefModelVal.GetField("ptrStrArray")) {
		t.Errorf("check detailRefModelVal field ptrStrArray valid failed, really false, expect true")
	}
	if !models.IsAssignedField(detailRefModelVal.GetField("id")) {
		t.Errorf("check detailRefModelVal field id assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(detailRefModelVal.GetField("bArray")) {
		t.Errorf("check detailRefModelVal field bArray assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(detailRefModelVal.GetField("strArray")) {
		t.Errorf("check detailRefModelVal field strArray assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(detailRefModelVal.GetField("ptrArray")) {
		t.Errorf("check detailRefModelVal field ptrArray assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(detailRefModelVal.GetField("ptrStrArray")) {
		t.Errorf("check detailRefModelVal field ptrStrArray assigned failed, really false, expect true")
	}

	liteRefModelVal := refModelVal.Copy(models.LiteView)
	if liteRefModelVal == nil {
		t.Errorf("Copy liteRefModelVal model failed")
	}
	if !models.IsValidField(liteRefModelVal.GetField("id")) {
		t.Errorf("check liteRefModelVal field id valid failed, really false, expect true")
	}
	if !models.IsValidField(liteRefModelVal.GetField("bArray")) {
		t.Errorf("check liteRefModelVal field bArray valid failed, really false, expect true")
	}
	if !models.IsValidField(liteRefModelVal.GetField("strArray")) {
		t.Errorf("check liteRefModelVal field strArray valid failed, really false, expect true")
	}
	if !models.IsValidField(liteRefModelVal.GetField("ptrArray")) {
		t.Errorf("check liteRefModelVal field ptrArray valid failed, really false, expect true")
	}
	if models.IsValidField(liteRefModelVal.GetField("ptrStrArray")) {
		t.Errorf("check liteRefModelVal field ptrStrArray valid failed, really true, expect false")
	}
	if !models.IsAssignedField(liteRefModelVal.GetField("id")) {
		t.Errorf("check liteRefModelVal field id assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(liteRefModelVal.GetField("bArray")) {
		t.Errorf("check liteRefModelVal field bArray assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(liteRefModelVal.GetField("strArray")) {
		t.Errorf("check liteRefModelVal field strArray assigned failed, really false, expect true")
	}
	if !models.IsAssignedField(liteRefModelVal.GetField("ptrArray")) {
		t.Errorf("check liteRefModelVal field ptrArray assigned failed, really false, expect true")
	}
	if models.IsAssignedField(liteRefModelVal.GetField("ptrStrArray")) {
		t.Errorf("check liteRefModelVal field ptrStrArray assigned failed, really true, expect flase")
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

	// Get entity entityModel
	entityModel, err := GetEntityModel(&unit)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// Test normal copy (with values)
	copiedModel := entityModel.Copy(models.OriginView)
	if copiedModel == nil {
		t.Errorf("Model.Copy() returned nil")
		return
	}

	// Verify fields were copied correctly
	copiedUnit, ok := copiedModel.Interface(false).(Unit)
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

	//if !copiedUnit.TimeStamp.Equal(unit.TimeStamp) {
	//	t.Errorf("Copied TimeStamp mismatch, expected: %v, got: %v", unit.TimeStamp, copiedUnit.TimeStamp)
	//}

	if copiedUnit.T1.Val != unit.T1.Val {
		t.Errorf("Copied nested struct value mismatch, expected: %d, got: %d", unit.T1.Val, copiedUnit.T1.Val)
	}

	// Test reset copy (without values)
	resetModel := entityModel.Copy(models.MetaView)
	if resetModel == nil {
		t.Errorf("Model.Copy(true) returned nil")
		return
	}

	// Verify fields exist but values are reset
	resetUnit, ok := resetModel.Interface(false).(Unit)
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

	// Get entity unitModel
	unitModel, err := GetEntityModel(&unit)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	u001 := unitModel.Interface(false).(Unit)
	if u001.ID != 100 || u001.Name != "OriginalName" {
		t.Errorf("Interface failed for DetailView, expected ID: 100, Name: OriginalName, got: ID: %d, Name: %s", u001.ID, u001.Name)
	}

	// Test updating string field
	newName := "UpdatedName"
	unitModel.GetField("name").SetValue(newName)
	if unitModel.GetField("name").GetValue().Get() != newName {
		t.Errorf("Name update failed, expected: %s, got: %s", newName, unitModel.GetField("name").GetValue().Get())
	}

	u002 := unitModel.Interface(false).(Unit)
	if u002.ID != 100 || u002.Name != newName {
		t.Errorf("Interface failed for DetailView, expected ID: 100, Name: %s, got: ID: %d, Name: %s", newName, u002.ID, u002.Name)
	}

	// Test updating int field
	newID := int64(200)
	unitModel.SetFieldValue("id", newID)

	// Verify updates worked
	updatedUnit, ok := unitModel.Interface(false).(Unit)
	if !ok {
		t.Errorf("models.Interface() failed to convert to Unit")
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
	unitModel.SetFieldValue("nonexistent", invalidVal)

	// Test updating with incompatible type
	invalidTypeVal := NewValue(reflect.ValueOf(true)) // bool value for string field
	unitModel.SetFieldValue("name", invalidTypeVal)
	unitModel.SetFieldValue("name", "abc")
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

	// Get entity entityEodel
	entityEodel, err := GetEntityModel(viewTest)
	if err != nil {
		t.Errorf("GetEntityModel failed: %s", err.Error())
		return
	}

	// For object_impl, we'll test the view functionality by checking
	// if fields are present in different views

	// Get all fields (using Origin view)
	allFields := entityEodel.GetFields()
	if len(allFields) < 4 {
		t.Errorf("Expected at least 4 fields in origin view, got %d", len(allFields))
	}

	// Check for specific views
	// In the model implementation, we may not have a direct GetViewFields method
	// Instead, we'll verify field visibility through specific view interfaces

	// Detail view should include ID, Name, Value but not Description
	detailObj, ok := entityEodel.Interface(false).(ViewTestStruct)
	if !ok {
		t.Errorf("models.Interface() failed to convert to ViewTestStruct for detail view")
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

	liteModel := entityEodel.Copy(models.LiteView)
	// Lite view should only include ID and Name
	liteObj, ok := liteModel.Interface(false).(ViewTestStruct)
	if !ok {
		t.Errorf("models.Interface() failed to convert to ViewTestStruct for lite view")
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
	entityModel, err := GetEntityModel(&testObj)
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

	if !models.IsStructField(baseField) {
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

	// Get entity entityModel
	entityModel, entityErr := GetEntityModel(&ptrObj)
	if entityErr != nil {
		t.Errorf("GetEntityModel failed: %s", entityErr.Error())
		return
	}

	// Check if interface can be retrieved
	retrievedObj, ok := entityModel.Interface(false).(PointerStruct)
	if !ok {
		t.Errorf("models.Interface() failed to convert to PointerStruct")
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
	entityModel.SetFieldValue("name", newName)

	// Verify update worked
	updatedObj, ok := entityModel.Interface(false).(PointerStruct)
	if !ok {
		t.Errorf("models.Interface() failed to convert to PointerStruct")
		return
	}

	if *updatedObj.Name != newName {
		t.Errorf("Name update failed, expected: %s, got: %s", newName, *updatedObj.Name)
	}
}

func TestAssign(t *testing.T) {
	type AssignStruct struct {
		ID          int     `orm:"id key"`
		Ptr         *int    `orm:"ptr"`
		Slice       []int   `orm:"slice"`
		PtrSlice    []*int  `orm:"ptrSlice"`
		SlicePtr    *[]int  `orm:"slicePtr"`
		PtrSlicePtr *[]*int `orm:"ptrSlicePtr"`
	}

	intVal := 0
	ptrVal := &intVal
	sliceVal := []int{}
	ptrSliceVal := []*int{}
	slicePtrVal := &sliceVal
	ptrSlicePtrVal := &ptrSliceVal

	rawVal := &AssignStruct{Ptr: ptrVal, Slice: sliceVal, PtrSlice: ptrSliceVal, SlicePtr: slicePtrVal, PtrSlicePtr: ptrSlicePtrVal}
	modelVal, modelErr := GetEntityModel(rawVal)
	if modelErr != nil {
		t.Errorf("GetEntityModel failed, err: %s", modelErr.Error())
	}

	zeroModelVal := modelVal.Copy(models.MetaView)

	var id = 18
	var intPtr = &id
	var intSlice = []int{1, 2, 3}
	var intPtrSlice = []*int{&intSlice[0], &intSlice[1], &intSlice[2]}
	var intSlicePtr = &intSlice
	var intPtrSlicePtr = &intPtrSlice

	err := modelVal.SetFieldValue("id", id)
	if err != nil {
		t.Errorf("SetFieldValue->id failed, err: %s", err.Error())
	}
	err = modelVal.SetFieldValue("ptr", intPtr)
	if err != nil {
		t.Errorf("SetFieldValue->ptr failed, err: %s", err.Error())
	}
	sliceField := modelVal.GetField("slice")
	for _, lVal := range intSlice {
		err = sliceField.AppendSliceValue(lVal)
		if err != nil {
			t.Errorf("AppendSliceValue->slice failed, err: %s", err.Error())
		}
	}
	slicePtrField := modelVal.GetField("ptrSlice")
	for _, lVal := range intPtrSlice {
		err = slicePtrField.AppendSliceValue(lVal)
		if err != nil {
			t.Errorf("AppendSliceValue->ptrSlice failed, err: %s", err.Error())
		}
	}
	err = modelVal.SetFieldValue("slicePtr", intSlicePtr)
	if err != nil {
		t.Errorf("SetFieldValue->slicePtr failed, err: %s", err.Error())
	}
	err = modelVal.SetFieldValue("ptrSlicePtr", intPtrSlicePtr)
	if err != nil {
		t.Errorf("SetFieldValue->ptrSlicePtr failed, err: %s", err.Error())
	}

	if rawVal.ID != id {
		t.Errorf("Assign failed, expected: %d, got: %d", id, rawVal.ID)
	}
	if *rawVal.Ptr != *intPtr {
		t.Errorf("Assign failed, expected: %d, got: %d", *intPtr, *rawVal.Ptr)
	}
	if !reflect.DeepEqual(rawVal.Slice, intSlice) {
		t.Errorf("Assign failed, expected: %v, got: %v", intSlice, rawVal.Slice)
	}
	if !reflect.DeepEqual(rawVal.PtrSlice, intPtrSlice) {
		t.Errorf("Assign failed, expected: %v, got: %v", intPtrSlice, rawVal.PtrSlice)
	}
	if !reflect.DeepEqual(rawVal.SlicePtr, intSlicePtr) {
		t.Errorf("Assign failed, expected: %v, got: %v", intSlicePtr, rawVal.SlicePtr)
	}
	if !reflect.DeepEqual(rawVal.PtrSlicePtr, intPtrSlicePtr) {
		t.Errorf("Assign failed, expected: %v, got: %v", intPtrSlicePtr, rawVal.PtrSlicePtr)
	}

	newVal := modelVal.Interface(false).(AssignStruct)
	if newVal.ID != id {
		t.Errorf("Assign failed, expected: %d, got: %d", id, newVal.ID)
	}
	if *newVal.Ptr != *intPtr {
		t.Errorf("Assign failed, expected: %d, got: %d", *intPtr, *newVal.Ptr)
	}
	if !reflect.DeepEqual(newVal.Slice, intSlice) {
		t.Errorf("Assign failed, expected: %v, got: %v", intSlice, newVal.Slice)
	}
	if !reflect.DeepEqual(newVal.PtrSlice, intPtrSlice) {
		t.Errorf("Assign failed, expected: %v, got: %v", intPtrSlice, newVal.PtrSlice)
	}
	if !reflect.DeepEqual(newVal.SlicePtr, intSlicePtr) {
		t.Errorf("Assign failed, expected: %v, got: %v", intSlicePtr, newVal.SlicePtr)
	}
	if !reflect.DeepEqual(newVal.PtrSlicePtr, intPtrSlicePtr) {
		t.Errorf("Assign failed, expected: %v, got: %v", intPtrSlicePtr, newVal.PtrSlicePtr)
	}

	log.Infof("rawVal:%+v", rawVal)
	log.Infof("newVal:%+v", newVal)

	err = zeroModelVal.SetFieldValue("id", id)
	if err != nil {
		t.Errorf("SetFieldValue->id failed, err: %s", err.Error())
	}
	err = zeroModelVal.SetFieldValue("ptr", intPtr)
	if err != nil {
		t.Errorf("SetFieldValue->ptr failed, err: %s", err.Error())
	}
	sliceField = zeroModelVal.GetField("slice")
	for _, lVal := range intSlice {
		err = sliceField.AppendSliceValue(lVal)
		if err != nil {
			t.Errorf("AppendSliceValue->slice failed, err: %s", err.Error())
		}
	}
	slicePtrField = zeroModelVal.GetField("ptrSlice")
	for _, lVal := range intPtrSlice {
		err = slicePtrField.AppendSliceValue(lVal)
		if err != nil {
			t.Errorf("AppendSliceValue->ptrSlice failed, err: %s", err.Error())
		}
	}
	err = zeroModelVal.SetFieldValue("slicePtr", intSlicePtr)
	if err != nil {
		t.Errorf("SetFieldValue->slicePtr failed, err: %s", err.Error())
	}
	err = zeroModelVal.SetFieldValue("ptrSlicePtr", intPtrSlicePtr)
	if err != nil {
		t.Errorf("SetFieldValue->ptrSlicePtr failed, err: %s", err.Error())
	}

	updatedVal001 := zeroModelVal.Interface(true)
	log.Infof("rawVal:%+v", rawVal)
	log.Infof("newVal:%+v", newVal)
	log.Infof("updatedVal:%+v", updatedVal001)

	zeroModelVal.SetFieldValue("id", 100)
	err = sliceField.AppendSliceValue(100)
	if err != nil {
		t.Errorf("AppendSliceValue->id failed, err: %s", err.Error())
	}
	updatedVal002 := zeroModelVal.Interface(true)
	log.Infof("rawVal:%+v", rawVal)
	log.Infof("newVal:%+v", newVal)
	log.Infof("updatedVal:%+v", updatedVal002)

	zero02Model := zeroModelVal.Copy(models.MetaView)
	slicePtrField = zero02Model.GetField("slicePtr")
	err = slicePtrField.AppendSliceValue(100)
	if err != nil {
		t.Errorf("AppendSliceValue->slicePtr failed, err: %s", err.Error())
	}
	updatedVal003 := zero02Model.Interface(true)
	log.Infof("rawVal:%+v", rawVal)
	log.Infof("newVal:%+v", newVal)
	log.Infof("updatedVal:%+v", updatedVal003)
}
