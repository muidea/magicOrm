package local

import (
	"testing"

	pu "github.com/muidea/magicOrm/provider/util"
)

// TestStruct for filter pagination and sort tests
type PaginationSortTestStruct struct {
	ID    int    `orm:"id key auto"`
	Name  string `orm:"name"`
	Value int    `orm:"value"`
}

func TestFilterPaginationExtended(t *testing.T) {
	// Create a test entity
	entity := PaginationSortTestStruct{
		ID:    1,
		Name:  "test",
		Value: 100,
	}

	// Create a filter
	modelVal, _ := GetEntityValue(entity)
	valImpl, _ := modelVal.(*ValueImpl)
	filterObj := newFilter(valImpl)

	// Test pagination settings
	expectedPageNum := 2
	expectedPageSize := 10
	filterObj.Pagination(expectedPageNum, expectedPageSize)

	// Verify pagination is correctly set
	pager := filterObj.Paginationer()
	if pager == nil {
		t.Errorf("Paginationer() returned nil")
		return
	}

	pageFilter, ok := pager.(*pu.Pagination)
	if !ok {
		t.Errorf("Paginationer() returned wrong type: %T", pager)
		return
	}

	if pageFilter.PageNum != expectedPageNum {
		t.Errorf("PageNum mismatch, expected: %d, got: %d", expectedPageNum, pageFilter.PageNum)
	}

	if pageFilter.PageSize != expectedPageSize {
		t.Errorf("PageSize mismatch, expected: %d, got: %d", expectedPageSize, pageFilter.PageSize)
	}

	// Test edge cases
	// Zero page size
	filterObj.Pagination(1, 0)
	pager = filterObj.Paginationer()
	pageFilter, _ = pager.(*pu.Pagination)
	if pageFilter.PageSize != 0 {
		t.Errorf("PageSize should be 0 when set to 0")
	}

	// Negative page number
	filterObj.Pagination(-1, 10)
	pager = filterObj.Paginationer()
	pageFilter, _ = pager.(*pu.Pagination)
	if pageFilter.PageNum != -1 {
		t.Errorf("PageNum should be -1 when set to -1")
	}
}

func TestFilterSortExtended(t *testing.T) {
	// Create a test entity
	entity := PaginationSortTestStruct{
		ID:    1,
		Name:  "test",
		Value: 100,
	}

	// Create a filter
	modelVal, _ := GetEntityValue(entity)
	valImpl, _ := modelVal.(*ValueImpl)
	filterObj := newFilter(valImpl)

	// Test sort settings - ascending order
	expectedFieldName := "name"
	expectedAscFlag := true
	filterObj.Sort(expectedFieldName, expectedAscFlag)

	// Verify sort is correctly set
	sorter := filterObj.Sorter()
	if sorter == nil {
		t.Errorf("Sorter() returned nil")
		return
	}

	sortFilter, ok := sorter.(*pu.SortFilter)
	if !ok {
		t.Errorf("Sorter() returned wrong type: %T", sorter)
		return
	}

	if sortFilter.FieldName != expectedFieldName {
		t.Errorf("FieldName mismatch, expected: %s, got: %s", expectedFieldName, sortFilter.FieldName)
	}

	if sortFilter.AscFlag != expectedAscFlag {
		t.Errorf("AscFlag mismatch, expected: %v, got: %v", expectedAscFlag, sortFilter.AscFlag)
	}

	// Test descending order
	expectedAscFlag = false
	filterObj.Sort(expectedFieldName, expectedAscFlag)
	sorter = filterObj.Sorter()
	sortFilter, _ = sorter.(*pu.SortFilter)
	if sortFilter.AscFlag != expectedAscFlag {
		t.Errorf("AscFlag mismatch for descending, expected: %v, got: %v", expectedAscFlag, sortFilter.AscFlag)
	}

	// Test with invalid field name
	filterObj.Sort("invalidField", true)
	sorter = filterObj.Sorter()
	sortFilter, _ = sorter.(*pu.SortFilter)
	if sortFilter.FieldName != "invalidField" {
		t.Errorf("FieldName mismatch for invalid field, expected: %s, got: %s", "invalidField", sortFilter.FieldName)
	}
}

func TestCombinedPaginationAndSortExtended(t *testing.T) {
	// Create a test entity
	entity := PaginationSortTestStruct{
		ID:    1,
		Name:  "test",
		Value: 100,
	}

	// Create a filter with both pagination and sort
	modelVal, _ := GetEntityValue(entity)
	valImpl, _ := modelVal.(*ValueImpl)
	filterObj := newFilter(valImpl)

	// Set pagination and sort
	expectedPageNum := 3
	expectedPageSize := 15
	expectedFieldName := "value"
	expectedAscFlag := false

	filterObj.Pagination(expectedPageNum, expectedPageSize)
	filterObj.Sort(expectedFieldName, expectedAscFlag)

	// Verify both settings are correctly set
	pager := filterObj.Paginationer()
	if pager == nil {
		t.Errorf("Paginationer() returned nil")
		return
	}

	pageFilter, ok := pager.(*pu.Pagination)
	if !ok {
		t.Errorf("Paginationer() returned wrong type: %T", pager)
		return
	}

	if pageFilter.PageNum != expectedPageNum {
		t.Errorf("PageNum mismatch, expected: %d, got: %d", expectedPageNum, pageFilter.PageNum)
	}

	if pageFilter.PageSize != expectedPageSize {
		t.Errorf("PageSize mismatch, expected: %d, got: %d", expectedPageSize, pageFilter.PageSize)
	}

	sorter := filterObj.Sorter()
	if sorter == nil {
		t.Errorf("Sorter() returned nil")
		return
	}

	sortFilter, ok := sorter.(*pu.SortFilter)
	if !ok {
		t.Errorf("Sorter() returned wrong type: %T", sorter)
		return
	}

	if sortFilter.FieldName != expectedFieldName {
		t.Errorf("FieldName mismatch, expected: %s, got: %s", expectedFieldName, sortFilter.FieldName)
	}

	if sortFilter.AscFlag != expectedAscFlag {
		t.Errorf("AscFlag mismatch, expected: %v, got: %v", expectedAscFlag, sortFilter.AscFlag)
	}
}

func TestMaskModelInterfaceExtended(t *testing.T) {
	// Create test entity
	testEntity := PaginationSortTestStruct{
		ID:    1,
		Name:  "test",
		Value: 100,
	}

	// Create a filter for masking
	modelVal, _ := GetEntityValue(testEntity)
	valImpl, _ := modelVal.(*ValueImpl)
	filterObj := newFilter(valImpl)

	// Mask testEntityModel with filter
	testEntityModel, _ := GetEntityModel(testEntity)
	err := filterObj.ValueMask(testEntity)
	if err != nil {
		t.Errorf("Mask failed: %s", err.Error())
		return
	}

	// Interface the model to get updated values
	interfaceVal := testEntityModel.Interface(false, "")
	resultEntity, ok := interfaceVal.(PaginationSortTestStruct)
	if !ok {
		t.Errorf("Interface() returned wrong type: %T", interfaceVal)
		return
	}

	// Verify masked values
	if resultEntity.Name != "test" {
		t.Errorf("Name not masked correctly, expected: %s, got: %s", "test", resultEntity.Name)
	}

	if resultEntity.Value != 100 {
		t.Errorf("Value not masked correctly, expected: %d, got: %d", 100, resultEntity.Value)
	}

	// ID should remain unchanged
	if resultEntity.ID != 1 {
		t.Errorf("ID should not be affected by mask, expected: %d, got: %d", 1, resultEntity.ID)
	}

	// Test masking with invalid value type
	var maskValues = make(map[string]interface{})
	maskValues["id"] = 100
	maskValues["name"] = "masked_name"
	maskValues["value"] = "200" // Will be converted to int

	// Mask model with filter
	err = filterObj.ValueMask(&PaginationSortTestStruct{})
	if err != nil {
		t.Errorf("Mask failed: %s", err.Error())
		return
	}

	// Test masking with invalid value type
	maskValues["value"] = "invalid_int"
	err = filterObj.ValueMask(&PaginationSortTestStruct{})
	if err != nil {
		t.Errorf("Mask should fail with invalid int value")
	}

	// Test masking non-existent field
	delete(maskValues, "value")
	maskValues["nonexistent"] = "value"
	err = filterObj.ValueMask(&PaginationSortTestStruct{})
	if err != nil {
		t.Errorf("Mask should succeed with non-existent field: %s", err.Error())
	}
}
