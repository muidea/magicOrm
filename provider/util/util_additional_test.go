package util

import (
	"reflect"
	"testing"
	"time"

	fu "github.com/muidea/magicCommon/foundation/util"
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

	if pagination.Limit() != 30 {
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
func TestTypeChecks(t *testing.T) {
	// 测试IsInteger
	if !IsInteger(reflect.TypeOf(int(0))) {
		t.Errorf("int should be identified as integer")
	}
	if !IsInteger(reflect.TypeOf(int8(0))) {
		t.Errorf("int8 should be identified as integer")
	}
	if !IsInteger(reflect.TypeOf(int16(0))) {
		t.Errorf("int16 should be identified as integer")
	}
	if !IsInteger(reflect.TypeOf(int32(0))) {
		t.Errorf("int32 should be identified as integer")
	}
	if !IsInteger(reflect.TypeOf(int64(0))) {
		t.Errorf("int64 should be identified as integer")
	}
	if IsInteger(reflect.TypeOf(uint(0))) {
		t.Errorf("uint should not be identified as integer")
	}

	// 测试IsUInteger
	if !IsUInteger(reflect.TypeOf(uint(0))) {
		t.Errorf("uint should be identified as unsigned integer")
	}
	if !IsUInteger(reflect.TypeOf(uint8(0))) {
		t.Errorf("uint8 should be identified as unsigned integer")
	}
	if !IsUInteger(reflect.TypeOf(uint16(0))) {
		t.Errorf("uint16 should be identified as unsigned integer")
	}
	if !IsUInteger(reflect.TypeOf(uint32(0))) {
		t.Errorf("uint32 should be identified as unsigned integer")
	}
	if !IsUInteger(reflect.TypeOf(uint64(0))) {
		t.Errorf("uint64 should be identified as unsigned integer")
	}
	if IsUInteger(reflect.TypeOf(int(0))) {
		t.Errorf("int should not be identified as unsigned integer")
	}

	// 测试IsFloat
	if !IsFloat(reflect.TypeOf(float32(0))) {
		t.Errorf("float32 should be identified as float")
	}
	if !IsFloat(reflect.TypeOf(float64(0))) {
		t.Errorf("float64 should be identified as float")
	}
	if IsFloat(reflect.TypeOf(int(0))) {
		t.Errorf("int should not be identified as float")
	}

	// 测试IsNumber
	if !IsNumber(reflect.TypeOf(int(0))) {
		t.Errorf("int should be identified as number")
	}
	if !IsNumber(reflect.TypeOf(uint(0))) {
		t.Errorf("uint should be identified as number")
	}
	if !IsNumber(reflect.TypeOf(float32(0))) {
		t.Errorf("float32 should be identified as number")
	}
	if IsNumber(reflect.TypeOf("string")) {
		t.Errorf("string should not be identified as number")
	}

	// 测试IsBool
	if !IsBool(reflect.TypeOf(bool(true))) {
		t.Errorf("bool should be identified as bool")
	}
	if IsBool(reflect.TypeOf(int(0))) {
		t.Errorf("int should not be identified as bool")
	}

	// 测试IsString
	if !IsString(reflect.TypeOf("")) {
		t.Errorf("string should be identified as string")
	}
	if IsString(reflect.TypeOf(int(0))) {
		t.Errorf("int should not be identified as string")
	}

	// 测试IsDateTime
	if !IsDateTime(reflect.TypeOf(time.Time{})) {
		t.Errorf("time.Time should be identified as DateTime")
	}
	if IsDateTime(reflect.TypeOf("")) {
		t.Errorf("string should not be identified as DateTime")
	}

	// 测试IsSlice
	if !IsSlice(reflect.TypeOf([]int{})) {
		t.Errorf("[]int should be identified as slice")
	}
	if IsSlice(reflect.TypeOf(int(0))) {
		t.Errorf("int should not be identified as slice")
	}

	// 测试IsStruct
	type testStruct struct{}
	if !IsStruct(reflect.TypeOf(testStruct{})) {
		t.Errorf("struct should be identified as struct")
	}
	if IsStruct(reflect.TypeOf(int(0))) {
		t.Errorf("int should not be identified as struct")
	}

	// 测试IsMap
	if !IsMap(reflect.TypeOf(map[string]int{})) {
		t.Errorf("map should be identified as map")
	}
	if IsMap(reflect.TypeOf(int(0))) {
		t.Errorf("int should not be identified as map")
	}

	// 测试IsPtr
	var x int
	if !IsPtr(reflect.TypeOf(&x)) {
		t.Errorf("pointer should be identified as pointer")
	}
	if IsPtr(reflect.TypeOf(x)) {
		t.Errorf("non-pointer should not be identified as pointer")
	}
}

// TestIsValueFunctions 测试IsNil和IsZero函数
func TestIsValueFunctions(t *testing.T) {
	// 测试IsNil
	var nilSlice []int
	if !IsNil(reflect.ValueOf(nilSlice)) {
		t.Errorf("nil slice should be identified as nil")
	}

	nonNilSlice := []int{1, 2, 3}
	if IsNil(reflect.ValueOf(nonNilSlice)) {
		t.Errorf("non-nil slice should not be identified as nil")
	}

	var nilMap map[string]int
	if !IsNil(reflect.ValueOf(nilMap)) {
		t.Errorf("nil map should be identified as nil")
	}

	var nilPtr *int
	if !IsNil(reflect.ValueOf(nilPtr)) {
		t.Errorf("nil pointer should be identified as nil")
	}

	// 测试IsZero
	emptyString := ""
	if !IsZero(reflect.ValueOf(emptyString)) {
		t.Errorf("empty string should be identified as zero")
	}

	nonEmptyString := "test"
	if IsZero(reflect.ValueOf(nonEmptyString)) {
		t.Errorf("non-empty string should not be identified as zero")
	}

	zeroInt := 0
	if !IsZero(reflect.ValueOf(zeroInt)) {
		t.Errorf("zero int should be identified as zero")
	}

	nonZeroInt := 123
	if IsZero(reflect.ValueOf(nonZeroInt)) {
		t.Errorf("non-zero int should not be identified as zero")
	}

	zeroTime := time.Time{}
	if !IsZero(reflect.ValueOf(zeroTime)) {
		t.Errorf("zero time should be identified as zero")
	}

	nonZeroTime := time.Now()
	if IsZero(reflect.ValueOf(nonZeroTime)) {
		t.Errorf("non-zero time should not be identified as zero")
	}
}

// TestIsSameValue 测试IsSameValue函数
func TestIsSameValue(t *testing.T) {
	// 测试基本类型
	if !IsSameValue(123, 123) {
		t.Errorf("Same integers should be identified as same")
	}

	if IsSameValue(123, 456) {
		t.Errorf("Different integers should not be identified as same")
	}

	if !IsSameValue("test", "test") {
		t.Errorf("Same strings should be identified as same")
	}

	if IsSameValue("test", "different") {
		t.Errorf("Different strings should not be identified as same")
	}

	// 测试结构体
	type testStruct struct {
		ID   int
		Name string
	}

	s1 := testStruct{ID: 1, Name: "test"}
	s2 := testStruct{ID: 1, Name: "test"}
	s3 := testStruct{ID: 2, Name: "different"}

	if !IsSameValue(s1, s2) {
		t.Errorf("Structs with same values should be identified as same")
	}

	if IsSameValue(s1, s3) {
		t.Errorf("Structs with different values should not be identified as same")
	}

	// 测试切片
	slice1 := []int{1, 2, 3}
	slice2 := []int{1, 2, 3}
	slice3 := []int{4, 5, 6}

	if !IsSameValue(slice1, slice2) {
		t.Errorf("Slices with same values should be identified as same")
	}

	if IsSameValue(slice1, slice3) {
		t.Errorf("Slices with different values should not be identified as same")
	}
}

// TestGetRawValueFunctions 测试GetRaw系列函数
func TestGetRawValueFunctions(t *testing.T) {
	// 测试GetRawBool
	boolVal := true
	rawBool, boolErr := GetRawBool(reflect.ValueOf(boolVal))
	if boolErr != nil {
		t.Errorf("GetRawBool failed: %s", boolErr.Error())
	}
	if !rawBool {
		t.Errorf("Expected true, got false")
	}

	// 测试GetRawInt
	intVal := 123
	rawInt, intErr := GetRawInt(reflect.ValueOf(intVal))
	if intErr != nil {
		t.Errorf("GetRawInt failed: %s", intErr.Error())
	}
	if rawInt != 123 {
		t.Errorf("Expected 123, got %d", rawInt)
	}

	// 测试GetRawString
	strVal := "test"
	rawStr, strErr := GetRawString(reflect.ValueOf(strVal))
	if strErr != nil {
		t.Errorf("GetRawString failed: %s", strErr.Error())
	}
	if rawStr != "test" {
		t.Errorf("Expected 'test', got '%s'", rawStr)
	}

	// 测试GetRawDateTime
	timeVal := time.Now()
	rawTime, timeErr := GetRawDateTime(reflect.ValueOf(timeVal))
	if timeErr != nil {
		t.Errorf("GetRawDateTime failed: %s", timeErr.Error())
	}
	if !rawTime.Equal(timeVal) {
		t.Errorf("Times don't match")
	}

	// 测试类型转换
	var int8Val int8 = 123
	rawIntFrom8, intErr8 := GetRawInt(reflect.ValueOf(int8Val))
	if intErr8 != nil {
		t.Errorf("GetRawInt failed for int8: %s", intErr8.Error())
	}
	if rawIntFrom8 != 123 {
		t.Errorf("Expected 123, got %d", rawIntFrom8)
	}
}

// TestGetValueFunctions 测试Get系列函数
func TestGetValueFunctions(t *testing.T) {
	// 测试GetBool
	boolVal, boolErr := GetBool(true)
	if boolErr != nil {
		t.Errorf("GetBool failed: %s", boolErr.Error())
	}
	if val, ok := boolVal.Value().(bool); !ok || !val {
		t.Errorf("Expected true, got something else")
	}

	// 测试GetInt
	intVal, intErr := GetInt(123)
	if intErr != nil {
		t.Errorf("GetInt failed: %s", intErr.Error())
	}
	if val, ok := intVal.Value().(int); !ok || val != 123 {
		t.Errorf("Expected 123, got something else")
	}

	// 测试GetString
	strVal, strErr := GetString("test")
	if strErr != nil {
		t.Errorf("GetString failed: %s", strErr.Error())
	}
	if val, ok := strVal.Value().(string); !ok || val != "test" {
		t.Errorf("Expected 'test', got something else")
	}

	// 测试GetDateTime
	timeVal := time.Now()
	dateVal, dateErr := GetDateTime(timeVal)
	if dateErr != nil {
		t.Errorf("GetDateTime failed: %s", dateErr.Error())
	}
	if val, ok := dateVal.Value().(time.Time); !ok || !val.Equal(timeVal) {
		t.Errorf("Time values don't match")
	}

	// 测试无效类型
	_, invalidErr := GetInt("not an int")
	if invalidErr == nil {
		t.Errorf("Expected error for invalid type, but got none")
	}
}

// TestTimeHelpers 测试时间辅助函数
func TestTimeHelpers(t *testing.T) {
	// 测试GetCurrentDateTime
	now := time.Now()
	dateTime := GetCurrentDateTime()
	// 应该近似相等，允许几秒误差
	diff := dateTime.Sub(now)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("GetCurrentDateTime should return current time, diff: %v", diff)
	}

	// 测试GetCurrentDateTimeStr
	dateTimeStr := GetCurrentDateTimeStr()
	// 应该是合法的时间字符串
	_, parseErr := time.Parse(fu.CSTLayout, dateTimeStr)
	if parseErr != nil {
		t.Errorf("GetCurrentDateTimeStr should return valid RFC3339 time string: %s", parseErr.Error())
	}
}

// TestIDGenerators 测试ID生成器
func TestIDGenerators(t *testing.T) {
	// 测试GetNewUUID
	uuid1 := GetNewUUID()
	uuid2 := GetNewUUID()
	// UUID应该不同
	if uuid1 == uuid2 {
		t.Errorf("GetNewUUID should return different values: %s vs %s", uuid1, uuid2)
	}
	// UUID应该是32字符长度
	if len(uuid1) != 32 {
		t.Errorf("UUID should be 32 characters long, got %d: %s", len(uuid1), uuid1)
	}

	// 测试GetNewSnowFlakeID
	id1 := GetNewSnowFlakeID()
	id2 := GetNewSnowFlakeID()
	// SnowFlake ID应该不同
	if id1 == id2 {
		t.Errorf("GetNewSnowFlakeID should return different values: %d vs %d", id1, id2)
	}
	// ID应该大于0
	if id1 <= 0 || id2 <= 0 {
		t.Errorf("SnowFlake IDs should be positive: %d, %d", id1, id2)
	}
}
