package utils

import (
	"testing"
)

// TestSortFilter 测试排序过滤器
func TestSortFilter(t *testing.T) {
	// 创建排序过滤器
	filter := &SortFilter{
		FieldName: "test_field",
		AscFlag:   true,
	}

	// 测试Name方法
	if filter.Name() != "test_field" {
		t.Errorf("Expected field name to be 'test_field', but got '%s'", filter.Name())
		return
	}

	// 测试AscSort方法
	if !filter.AscSort() {
		t.Errorf("Expected AscSort to be true, but got false")
		return
	}

	// 测试降序
	filter.AscFlag = false
	if filter.AscSort() {
		t.Errorf("Expected AscSort to be false, but got true")
		return
	}
}

// TestPagination 测试分页
func TestPagination(t *testing.T) {
	// 测试有效值
	pagination := &Pagination{
		PageNum:  3,
		PageSize: 10,
	}

	if pagination.Limit() != 10 {
		t.Errorf("Expected Limit to be 10, but got %d", pagination.Limit())
		return
	}

	if pagination.Offset() != 20 {
		t.Errorf("Expected Offset to be 20, but got %d", pagination.Offset())
		return
	}

	// 测试负值处理
	pagination = &Pagination{
		PageNum:  -1,
		PageSize: -5,
	}

	if pagination.Limit() != 10 {
		t.Errorf("Expected Limit to be 10 (default) for negative value, but got %d", pagination.Limit())
		return
	}

	if pagination.Offset() != 0 {
		t.Errorf("Expected Offset to be 0 for negative values, but got %d", pagination.Offset())
		return
	}

	// 测试零值处理
	pagination = &Pagination{
		PageNum:  0,
		PageSize: 0,
	}

	if pagination.Limit() != 10 {
		t.Errorf("Expected Limit to be 10 (default) for zero value, but got %d", pagination.Limit())
		return
	}

	if pagination.Offset() != 0 {
		t.Errorf("Expected Offset to be 0 for zero values, but got %d", pagination.Offset())
		return
	}
}

// TestTypeChecks 测试类型检查函数
