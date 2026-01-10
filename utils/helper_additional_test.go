package utils

import (
	"reflect"
	"testing"
	"time"

	fu "github.com/muidea/magicCommon/foundation/util"
	"github.com/stretchr/testify/assert"
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
	rawBool, boolErr := ConvertToBool(reflect.ValueOf(boolVal))
	if boolErr != nil {
		t.Errorf("GetRawBool failed: %s", boolErr.Error())
	}
	if !rawBool {
		t.Errorf("Expected true, got false")
	}

	// 测试GetRawInt
	intVal := 123
	rawInt, intErr := ConvertToInt(reflect.ValueOf(intVal))
	if intErr != nil {
		t.Errorf("GetRawInt failed: %s", intErr.Error())
	}
	if rawInt != 123 {
		t.Errorf("Expected 123, got %d", rawInt)
	}

	// 测试GetRawString
	strVal := "test"
	rawStr, strErr := ConvertRawToString(strVal)
	if strErr != nil {
		t.Errorf("GetRawString failed: %s", strErr.Error())
	}
	if rawStr != "test" {
		t.Errorf("Expected 'test', got '%s'", rawStr)
	}

	// 测试GetRawDateTime
	timeVal := time.Now()
	rawTime, timeErr := ConvertRawToDateTime(timeVal)
	if timeErr != nil {
		t.Errorf("GetRawDateTime failed: %s", timeErr.Error())
	}
	if !rawTime.Equal(timeVal) {
		t.Errorf("Times don't match")
	}

	// 测试类型转换
	var int8Val int8 = 123
	rawIntFrom8, intErr8 := ConvertToInt(reflect.ValueOf(int8Val))
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
	boolVal, boolErr := ConvertRawToBool(true)
	if boolErr != nil {
		t.Errorf("GetBool failed: %s", boolErr.Error())
	}
	if !boolVal {
		t.Errorf("Expected true, got something else")
	}

	// 测试GetInt
	intVal, intErr := ConvertRawToInt(123)
	if intErr != nil {
		t.Errorf("GetInt failed: %s", intErr.Error())
	}
	if intVal != 123 {
		t.Errorf("Expected 123, got %d", intVal)
	}

	// 测试GetString
	strVal, strErr := ConvertRawToString("test")
	if strErr != nil {
		t.Errorf("GetString failed: %s", strErr.Error())
	}
	if strVal != "test" {
		t.Errorf("Expected 'test', got %s", strVal)
	}

	// 测试GetDateTime
	timeVal := time.Now()
	dateVal, dateErr := ConvertRawToDateTime(timeVal)
	if dateErr != nil {
		t.Errorf("GetDateTime failed: %s", dateErr.Error())
	}
	if !dateVal.Equal(timeVal) {
		t.Errorf("Time values don't match")
	}

	// 测试无效类型
	_, invalidErr := ConvertRawToInt("not an int")
	if invalidErr == nil {
		t.Errorf("Expected error for invalid type, but got none")
	}
}

// TestTimeHelpers 测试时间辅助函数
func TestTimeHelpers(t *testing.T) {
	// 测试GetCurrentDateTime
	now := time.Now()
	datetime := GetCurrentDateTime()
	// 应该近似相等，允许几秒误差
	diff := datetime.Sub(now)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("GetCurrentDateTime should return current time, diff: %v", diff)
	}

	// 测试GetCurrentDateTimeStr
	dateTimeStr := GetCurrentDateTimeStr()
	// 应该是合法的时间字符串
	_, parseErr := time.Parse(fu.CSTLayoutWithMillisecond, dateTimeStr)
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

	// 测试GetNewSnowflakeID
	id1 := GetNewSnowflakeID()
	id2 := GetNewSnowflakeID()
	// Snowflake ID应该不同
	if id1 == id2 {
		t.Errorf("GetNewSnowflakeID should return different values: %d vs %d", id1, id2)
	}
	// ID应该大于0
	if id1 <= 0 || id2 <= 0 {
		t.Errorf("Snowflake IDs should be positive: %d, %d", id1, id2)
	}
}

// TestConvertNumberVal 测试 convertNumberVal 函数
func TestConvertNumberVal(t *testing.T) {
	tests := []struct {
		name     string
		kind     reflect.Kind
		input    any
		expected any
		hasError bool
	}{
		{"Bool to Int", reflect.Int, true, int(1), false},
		{"Bool to Uint", reflect.Uint, false, uint(0), false},
		{"Bool to Float", reflect.Float64, true, float64(1), false},
		{"Int to Int", reflect.Int32, int64(42), int32(42), false},
		{"Int to Uint", reflect.Uint16, int64(42), uint16(42), false},
		{"Int to Float", reflect.Float32, int64(42), float32(42), false},
		{"Uint to Int", reflect.Int8, uint64(42), int8(42), false},
		{"Uint to Uint", reflect.Uint64, uint64(42), uint64(42), false},
		{"Uint to Float", reflect.Float64, uint64(42), float64(42), false},
		{"Float to Int", reflect.Int64, float64(42.5), int64(42), false},
		{"Float to Uint", reflect.Uint32, float64(42.5), uint32(42), false},
		{"Float to Float", reflect.Float32, float64(42.5), float32(42.5), false},
		{"String to Int", reflect.Int16, "42", int16(42), false},
		{"String to Uint", reflect.Uint8, "42", uint8(42), false},
		{"String to Float", reflect.Float64, "42.5", float64(42.5), false},
		{"Invalid String to Int", reflect.Int, "not a number", nil, true},
		{"Unsupported Kind", reflect.String, int64(42), nil, true},
		{"Bool to Int8", reflect.Int8, true, int8(1), false},
		{"Bool to Int16", reflect.Int16, false, int16(0), false},
		{"Bool to Int32", reflect.Int32, true, int32(1), false},
		{"Bool to Int64", reflect.Int64, false, int64(0), false},
		{"Bool to Uint8", reflect.Uint8, true, uint8(1), false},
		{"Bool to Uint16", reflect.Uint16, false, uint16(0), false},
		{"Bool to Uint32", reflect.Uint32, true, uint32(1), false},
		{"Bool to Uint64", reflect.Uint64, false, uint64(0), false},
		{"Bool to Float32", reflect.Float32, true, float32(1), false},
		{"Int to Int8", reflect.Int8, int64(127), int8(127), false},
		{"Int to Int16", reflect.Int16, int64(32767), int16(32767), false},
		{"Int to Uint8", reflect.Uint8, int64(255), uint8(255), false},
		{"Int to Uint32", reflect.Uint32, int64(4294967295), uint32(4294967295), false},
		{"Int to Float64", reflect.Float64, int64(42), float64(42), false},
		{"Uint to Int16", reflect.Int16, uint64(32767), int16(32767), false},
		{"Uint to Int32", reflect.Int32, uint64(2147483647), int32(2147483647), false},
		{"Uint to Uint8", reflect.Uint8, uint64(255), uint8(255), false},
		{"Uint to Uint16", reflect.Uint16, uint64(65535), uint16(65535), false},
		{"Uint to Uint32", reflect.Uint32, uint64(4294967295), uint32(4294967295), false},
		{"Uint to Float32", reflect.Float32, uint64(42), float32(42), false},
		{"Float to Int8", reflect.Int8, float64(127.9), int8(127), false},
		{"Float to Int16", reflect.Int16, float64(32767.9), int16(32767), false},
		{"Float to Int32", reflect.Int32, float64(2147483647.9), int32(2147483647), false},
		{"Float to Uint8", reflect.Uint8, float64(255.9), uint8(255), false},
		{"Float to Uint16", reflect.Uint16, float64(65535.9), uint16(65535), false},
		{"String to Int8", reflect.Int8, "127", int8(127), false},
		{"String to Int32", reflect.Int32, "2147483647", int32(2147483647), false},
		{"String to Int64", reflect.Int64, "9223372036854775807", int64(9223372036854775807), false},
		{"String to Uint16", reflect.Uint16, "65535", uint16(65535), false},
		{"String to Uint32", reflect.Uint32, "4294967295", uint32(4294967295), false},
		{"String to Uint64", reflect.Uint64, "18446744073709551615", uint64(18446744073709551615), false},
		{"String to Float32", reflect.Float32, "3.14159", float32(3.14159), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rVal := reflect.ValueOf(tt.input)
			result, err := convertNumberVal(tt.kind, rVal)

			if tt.hasError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
