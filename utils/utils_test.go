package utils

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type BaseInfo struct {
	IntVal          int            `orm:"intVal"`
	EmptyIntVal     int            `orm:"EmptyIntVal"`
	StrVal          string         `orm:"strVal"`
	EmptyStrVal     string         `orm:"EmptyStrVal"`
	BoolVal         bool           `orm:"boolVal"`
	EmptyBoolVal    bool           `orm:"EmptyBoolVal"`
	Float32Val      float32        `orm:"float32Val"`
	EmptyFloat32Val float32        `orm:"EmptyFloat32Val"`
	Float64Val      float64        `orm:"float64Val"`
	EmptyFloat64Val float64        `orm:"EmptyFloat64Val"`
	TimeVal         time.Time      `orm:"timeVal"`
	EmptyTimeVal    time.Time      `orm:"EmptyTimeVal"`
	SliceVal        []int          `orm:"sliceVal"`
	EmptySliceVal   []int          `orm:"EmptySliceVal"`
	MapVal          map[string]int `orm:"mapVal"`
	EmptyMapVal     map[string]int `orm:"EmptyMapVal"`
}

// TestIsReallyValid tests the IsReallyValid function
func TestIsReallyValid(t *testing.T) {
	var sliceVal []int
	var mapVal map[string]int
	var structVal struct{}

	baseInfo := BaseInfo{
		IntVal:     1,
		StrVal:     "a",
		BoolVal:    true,
		Float32Val: 1.0,
		Float64Val: 1.0,
		TimeVal:    time.Now(),
		SliceVal:   []int{1},
		MapVal:     map[string]int{"a": 1},
	}

	tests := []struct {
		name string
		val  interface{}
		want bool
	}{
		{"Nil", nil, false},
		{"NilPointer", (*int)(nil), false},
		{"NonNilPointer", new(int), true},
		{"PointerToNonNilPointer", func() interface{} { i := 0; p := &i; return &p }(), true},
		{"ValidInt", 42, true},
		{"ValidString", "hello", true},
		{"ZeroInt", 0, true},
		{"EmptyString", "", true},
		{"EmptySlice", []int{}, true},
		{"sliceVal", sliceVal, false},
		{"NonEmptySlice", []int{1, 2, 3}, true},
		{"EmptyMap", map[string]int{}, true},
		{"mapVal", mapVal, false},
		{"NonEmptyMap", map[string]int{"a": 1}, true},
		{"EmptyStruct", struct{}{}, false},
		{"structVal", structVal, false},
		{"NonEmptyStruct", struct{ Name string }{Name: "test"}, true},
		{"NilFunc", (func())(nil), false},
		{"NilChan", (chan int)(nil), false},
		{"baseInfo", baseInfo, true},
		{"baseInfo.IntVal", baseInfo.IntVal, true},
		{"baseInfo.EmptyIntVal", baseInfo.EmptyIntVal, true},
		{"baseInfo.StrVal", baseInfo.StrVal, true},
		{"baseInfo.EmptyStrVal", baseInfo.EmptyStrVal, true},
		{"baseInfo.BoolVal", baseInfo.BoolVal, true},
		{"baseInfo.EmptyBoolVal", baseInfo.EmptyBoolVal, true},
		{"baseInfo.Float32Val", baseInfo.Float32Val, true},
		{"baseInfo.EmptyFloat32Val", baseInfo.EmptyFloat32Val, true},
		{"baseInfo.Float64Val", baseInfo.Float64Val, true},
		{"baseInfo.EmptyFloat64Val", baseInfo.EmptyFloat64Val, true},
		{"baseInfo.TimeVal", baseInfo.TimeVal, true},
		{"baseInfo.EmptyTimeVal", baseInfo.EmptyTimeVal, true},
		{"baseInfo.SliceVal", baseInfo.SliceVal, true},
		{"baseInfo.EmptySliceVal", baseInfo.EmptySliceVal, false},
		{"baseInfo.MapVal", baseInfo.MapVal, true},
		{"baseInfo.EmptyMapVal", baseInfo.EmptyMapVal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReallyValid(tt.val); got != tt.want {
				t.Errorf("name:%s IsReallyValid() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestIsReallyValidForReflect tests the IsReallyValidForReflect function
func TestIsReallyValidForReflect(t *testing.T) {
	var sliceVal []int
	var mapVal map[string]int
	var structVal struct{}

	baseInfo := BaseInfo{
		IntVal:     1,
		StrVal:     "a",
		BoolVal:    true,
		Float32Val: 1.0,
		Float64Val: 1.0,
		TimeVal:    time.Now(),
		SliceVal:   []int{1},
		MapVal:     map[string]int{"a": 1},
	}

	tests := []struct {
		name string
		val  reflect.Value
		want bool
	}{
		{"NilPointer", reflect.ValueOf((*int)(nil)), false},
		{"NonNilPointer", reflect.ValueOf(new(int)), true},
		{"ValidInt", reflect.ValueOf(42), true},
		{"ValidString", reflect.ValueOf("hello"), true},
		{"ValidStruct", reflect.ValueOf(struct{ Name string }{Name: "test"}), true},
		{"ZeroInt", reflect.ValueOf(0), true},
		{"EmptyString", reflect.ValueOf(""), true},
		{"EmptySlice", reflect.ValueOf([]int{}), true},
		{"NonEmptySlice", reflect.ValueOf([]int{1, 2, 3}), true},
		{"EmptyMap", reflect.ValueOf(map[string]int{}), true},
		{"NonEmptyMap", reflect.ValueOf(map[string]int{"a": 1}), true},
		{"EmptyStruct", reflect.ValueOf(struct{}{}), false},
		{"ZeroTime", reflect.ValueOf(time.Time{}), true},
		{"NonZeroTime", reflect.ValueOf(time.Now()), true},
		{"sliceVal", reflect.ValueOf(sliceVal), false},
		{"mapVal", reflect.ValueOf(mapVal), false},
		{"structVal", reflect.ValueOf(structVal), false},
		{"baseInfo", reflect.ValueOf(baseInfo), true},
		{"baseInfo.IntVal", reflect.ValueOf(baseInfo.IntVal), true},
		{"baseInfo.EmptyIntVal", reflect.ValueOf(baseInfo.EmptyIntVal), true},
		{"baseInfo.StrVal", reflect.ValueOf(baseInfo.StrVal), true},
		{"baseInfo.EmptyStrVal", reflect.ValueOf(baseInfo.EmptyStrVal), true},
		{"baseInfo.BoolVal", reflect.ValueOf(baseInfo.BoolVal), true},
		{"baseInfo.EmptyBoolVal", reflect.ValueOf(baseInfo.EmptyBoolVal), true},
		{"baseInfo.Float32Val", reflect.ValueOf(baseInfo.Float32Val), true},
		{"baseInfo.EmptyFloat32Val", reflect.ValueOf(baseInfo.EmptyFloat32Val), true},
		{"baseInfo.Float64Val", reflect.ValueOf(baseInfo.Float64Val), true},
		{"baseInfo.EmptyFloat64Val", reflect.ValueOf(baseInfo.EmptyFloat64Val), true},
		{"baseInfo.TimeVal", reflect.ValueOf(baseInfo.TimeVal), true},
		{"baseInfo.EmptyTimeVal", reflect.ValueOf(baseInfo.EmptyTimeVal), true},
		{"baseInfo.SliceVal", reflect.ValueOf(baseInfo.SliceVal), true},
		{"baseInfo.EmptySliceVal", reflect.ValueOf(baseInfo.EmptySliceVal), false},
		{"baseInfo.MapVal", reflect.ValueOf(baseInfo.MapVal), true},
		{"baseInfo.EmptyMapVal", reflect.ValueOf(baseInfo.EmptyMapVal), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReallyValidForReflect(tt.val); got != tt.want {
				t.Errorf("name:%s IsReallyValidForReflect() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}

	// Test for nil channel and function
	var funcPtr func() int
	if IsReallyValidForReflect(reflect.ValueOf(funcPtr)) {
		t.Errorf("nil function should not be valid")
	}

	var chanVal chan int
	if IsReallyValidForReflect(reflect.ValueOf(chanVal)) {
		t.Errorf("nil channel should not be valid")
	}
}

// TestIsReallyZero tests the IsReallyZero function
func TestIsReallyZero(t *testing.T) {
	var sliceVal []int
	var mapVal map[string]int
	var structVal struct{}

	baseInfo := BaseInfo{
		IntVal:     1,
		StrVal:     "a",
		BoolVal:    true,
		Float32Val: 1.0,
		Float64Val: 1.0,
		TimeVal:    time.Now(),
		SliceVal:   []int{1},
		MapVal:     map[string]int{"a": 1},
	}

	tests := []struct {
		name string
		val  interface{}
		want bool
	}{
		{"Nil", nil, true},
		{"NilPointer", (*int)(nil), true},
		{"NonNilPointerToZero", func() interface{} { i := 0; return &i }(), true},
		{"NonNilPointerToNonZero", func() interface{} { i := 1; return &i }(), false},
		{"ZeroInt", 0, true},
		{"NonZeroInt", 1, false},
		{"EmptyString", "", true},
		{"NonEmptyString", "hello", false},
		{"EmptySlice", []int{}, true},
		{"NonEmptySlice", []int{1, 2, 3}, false},
		{"EmptyMap", map[string]int{}, true},
		{"NonEmptyMap", map[string]int{"a": 1}, false},
		{"EmptyStruct", struct{}{}, true},
		{"ZeroTime", time.Time{}, true},
		{"NonZeroTime", time.Now(), false},
		{"StructWithZeroFields", struct{ Name string }{}, true},
		{"StructWithNonZeroFields", struct{ Name string }{Name: "test"}, false},
		{"ComplexNestedStruct", struct {
			Name  string
			Age   int
			Items []string
			Data  map[string]int
			Ptr   *int
		}{}, true},
		{"NilFunc", (func())(nil), true},
		{"NonNilFunc", func() {}, false},
		{"NilChan", (chan int)(nil), true},
		{"NonNilChan", make(chan int), false},
		{"Array", [3]int{0, 0, 0}, true},
		{"NonZeroArray", [3]int{0, 1, 0}, false},
		{"sliceVal", sliceVal, true},
		{"mapVal", mapVal, true},
		{"structVal", structVal, true},
		{"baseInfo.IntVal", baseInfo.IntVal, false},
		{"baseInfo.EmptyIntVal", baseInfo.EmptyIntVal, true},
		{"baseInfo.StrVal", baseInfo.StrVal, false},
		{"baseInfo.EmptyStrVal", baseInfo.EmptyStrVal, true},
		{"baseInfo.BoolVal", baseInfo.BoolVal, false},
		{"baseInfo.EmptyBoolVal", baseInfo.EmptyBoolVal, true},
		{"baseInfo.Float32Val", baseInfo.Float32Val, false},
		{"baseInfo.EmptyFloat32Val", baseInfo.EmptyFloat32Val, true},
		{"baseInfo.Float64Val", baseInfo.Float64Val, false},
		{"baseInfo.EmptyFloat64Val", baseInfo.EmptyFloat64Val, true},
		{"baseInfo.TimeVal", baseInfo.TimeVal, false},
		{"baseInfo.EmptyTimeVal", baseInfo.EmptyTimeVal, true},
		{"baseInfo.SliceVal", baseInfo.SliceVal, false},
		{"baseInfo.EmptySliceVal", baseInfo.EmptySliceVal, true},
		{"baseInfo.MapVal", baseInfo.MapVal, false},
		{"baseInfo.EmptyMapVal", baseInfo.EmptyMapVal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReallyZero(tt.val); got != tt.want {
				t.Errorf("name:%s IsReallyZero() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}

	zero := reflect.ValueOf(time.Time{})
	cur := reflect.ValueOf(time.Now())
	if IsReallyZeroForReflect(cur) {
		t.Errorf("name:NowTime IsReallyZero() = %v, want %v", cur, false)
	}
	if !IsReallyZeroForReflect(zero) {
		t.Errorf("name:ZeroTime IsReallyZero() = %v, want %v", zero, true)
	}
}

// TestIsReallyZeroForReflect tests the IsReallyZeroForReflect function
func TestIsReallyZeroForReflect(t *testing.T) {
	var sliceVal []int
	var mapVal map[string]int
	var structVal struct{}

	baseInfo := BaseInfo{
		IntVal:     1,
		StrVal:     "a",
		BoolVal:    true,
		Float32Val: 1.0,
		Float64Val: 1.0,
		TimeVal:    time.Now(),
		SliceVal:   []int{1},
		MapVal:     map[string]int{"a": 1},
	}

	tests := []struct {
		name string
		val  reflect.Value
		want bool
	}{
		{"NilPointer", reflect.ValueOf((*int)(nil)), true},
		{"NonNilPointerToZero", reflect.ValueOf(func() interface{} { i := 0; return &i }()), true},
		{"NonNilPointerToNonZero", reflect.ValueOf(func() interface{} { i := 1; return &i }()), false},
		{"ZeroInt", reflect.ValueOf(0), true},
		{"NonZeroInt", reflect.ValueOf(1), false},
		{"EmptyString", reflect.ValueOf(""), true},
		{"NonEmptyString", reflect.ValueOf("hello"), false},
		{"EmptySlice", reflect.ValueOf([]int{}), true},
		{"NonEmptySlice", reflect.ValueOf([]int{1, 2, 3}), false},
		{"EmptyMap", reflect.ValueOf(map[string]int{}), true},
		{"NonEmptyMap", reflect.ValueOf(map[string]int{"a": 1}), false},
		{"EmptyStruct", reflect.ValueOf(struct{}{}), true},
		{"StructWithZeroFields", reflect.ValueOf(struct{ Name string }{}), true},
		{"StructWithNonZeroFields", reflect.ValueOf(struct{ Name string }{Name: "test"}), false},
		{"NestedStruct", reflect.ValueOf(struct {
			Name   string
			Age    int
			Parent *struct{ Name string }
		}{}), true},
		{"NilFunc", reflect.ValueOf((func())(nil)), true},
		{"NonNilFunc", reflect.ValueOf(func() {}), false},
		{"NilChan", reflect.ValueOf((chan int)(nil)), true},
		{"NonNilChan", reflect.ValueOf(make(chan int)), false},
		{"ZeroArray", reflect.ValueOf([3]int{0, 0, 0}), true},
		{"NonZeroArray", reflect.ValueOf([3]int{0, 1, 0}), false},
		{"StructWithUnexportedField", reflect.ValueOf(struct {
			name string
			Age  int
		}{name: "hidden", Age: 0}), true},
		{"sliceVal", reflect.ValueOf(sliceVal), true},
		{"mapVal", reflect.ValueOf(mapVal), true},
		{"structVal", reflect.ValueOf(structVal), true},
		{"baseInfo", reflect.ValueOf(baseInfo), false},
		{"baseInfo.IntVal", reflect.ValueOf(baseInfo.IntVal), false},
		{"baseInfo.EmptyIntVal", reflect.ValueOf(baseInfo.EmptyIntVal), true},
		{"baseInfo.StrVal", reflect.ValueOf(baseInfo.StrVal), false},
		{"baseInfo.EmptyStrVal", reflect.ValueOf(baseInfo.EmptyStrVal), true},
		{"baseInfo.BoolVal", reflect.ValueOf(baseInfo.BoolVal), false},
		{"baseInfo.EmptyBoolVal", reflect.ValueOf(baseInfo.EmptyBoolVal), true},
		{"baseInfo.Float32Val", reflect.ValueOf(baseInfo.Float32Val), false},
		{"baseInfo.EmptyFloat32Val", reflect.ValueOf(baseInfo.EmptyFloat32Val), true},
		{"baseInfo.Float64Val", reflect.ValueOf(baseInfo.Float64Val), false},
		{"baseInfo.EmptyFloat64Val", reflect.ValueOf(baseInfo.EmptyFloat64Val), true},
		{"baseInfo.TimeVal", reflect.ValueOf(baseInfo.TimeVal), false},
		{"baseInfo.EmptyTimeVal", reflect.ValueOf(baseInfo.EmptyTimeVal), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReallyZeroForReflect(tt.val); got != tt.want {
				t.Errorf("name:%s, IsReallyZeroForReflect() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestComplexNestedCases tests more complex nested data structures
func TestComplexNestedCases(t *testing.T) {
	type InnerStruct struct {
		Field1 string
		Field2 int
		Field3 *string
	}

	type OuterStruct struct {
		Name     string
		Inner    InnerStruct
		InnerPtr *InnerStruct
		Data     []InnerStruct
		Map      map[string]*InnerStruct
	}

	// Test empty complex struct
	emptyComplex := OuterStruct{}
	if !IsReallyZero(emptyComplex) {
		t.Errorf("empty complex struct should be zero")
	}
	// Empty struct with exported fields should be valid
	if !IsReallyValid(emptyComplex) {
		t.Errorf("empty complex struct should be valid since it has exported fields")
	}

	// Test partially filled complex struct
	s := "test"
	partialComplex := OuterStruct{
		Name: "Test",
		Inner: InnerStruct{
			Field1: "",
			Field2: 0,
			Field3: nil,
		},
		InnerPtr: nil,
		Data:     nil,
		Map:      nil,
	}
	if IsReallyZero(partialComplex) {
		t.Errorf("partially filled complex struct should not be zero")
	}
	if !IsReallyValid(partialComplex) {
		t.Errorf("partially filled complex struct should be valid")
	}

	// Test fully filled complex struct
	fullComplex := OuterStruct{
		Name: "Test",
		Inner: InnerStruct{
			Field1: "Value",
			Field2: 42,
			Field3: &s,
		},
		InnerPtr: &InnerStruct{
			Field1: "Inner",
			Field2: 100,
			Field3: nil,
		},
		Data: []InnerStruct{
			{
				Field1: "Item1",
				Field2: 1,
				Field3: nil,
			},
		},
		Map: map[string]*InnerStruct{
			"key": {
				Field1: "MapValue",
				Field2: 200,
				Field3: &s,
			},
		},
	}
	if IsReallyZero(fullComplex) {
		t.Errorf("fully filled complex struct should not be zero")
	}
	if !IsReallyValid(fullComplex) {
		t.Errorf("fully filled complex struct should be valid")
	}
}

// TestIsReallyValidType tests the IsReallyValidType function
func TestIsReallyValidType(t *testing.T) {
	// 基本数据类型测试 - 应该返回 true
	testBasicTypes := []struct {
		name string
		val  interface{}
	}{
		{"bool", true},
		{"int", 42},
		{"int8", int8(8)},
		{"int16", int16(16)},
		{"int32", int32(32)},
		{"int64", int64(64)},
		{"uint", uint(42)},
		{"uint8", uint8(8)},
		{"uint16", uint16(16)},
		{"uint32", uint32(32)},
		{"uint64", uint64(64)},
		{"float32", float32(3.14)},
		{"float64", 3.14},
		{"string", "测试字符串"},
	}

	for _, tt := range testBasicTypes {
		t.Run(fmt.Sprintf("basic_type_%s", tt.name), func(t *testing.T) {
			if !IsReallyValidType(tt.val) {
				t.Errorf("%s 应该是合法类型", tt.name)
			}
		})
	}

	// 容器类型递归测试
	t.Run("容器类型递归测试", func(t *testing.T) {
		// 递归合法类型容器 - 应该返回 true
		validContainers := []struct {
			name string
			val  interface{}
		}{
			{"空切片", []int{}},
			{"整数切片", []int{1, 2, 3}},
			{"字符串切片", []string{"a", "b"}},
			{"整数指针切片", []*int{new(int)}},
			{"基本类型Map", map[string]int{"a": 1}},
			{"嵌套Map", map[int]map[string]float64{1: {"x": 1.0}}},
			{"有效结构体切片", []struct{ A int }{{A: 1}}},
			{"嵌套有效容器", []map[string][]int{{"a": {1, 2}}}},
			{"多层指针", new(int)},
			{"指针切片", &[]string{"a", "b"}},
		}

		for _, tt := range validContainers {
			t.Run("valid_"+tt.name, func(t *testing.T) {
				if !IsReallyValidType(tt.val) {
					t.Errorf("%s 应该是合法类型", tt.name)
				}
			})
		}

		// 递归非法类型容器 - 应该返回 false
		invalidContainers := []struct {
			name string
			val  interface{}
		}{
			{"chan元素切片", []chan int{make(chan int)}},
			{"func元素切片", []func(){func() {}}},
			{"chan值Map", map[string]chan int{"a": make(chan int)}},
			{"非基本类型Key的Map", map[struct{ A int }]int{{A: 1}: 1}},
			{"非法嵌套", []map[string][]chan int{{"a": {make(chan int)}}}},
			{"指向非法类型的指针", func() *chan int { ch := make(chan int); return &ch }()},
			{"包含非法类型切片的结构体", struct{ A []func() }{A: []func(){func() {}}}},
		}

		for _, tt := range invalidContainers {
			t.Run("invalid_"+tt.name, func(t *testing.T) {
				if IsReallyValidType(tt.val) {
					t.Errorf("%s 应该是非法类型", tt.name)
				}
			})
		}
	})

	// 结构体类型测试
	t.Run("结构体类型测试", func(t *testing.T) {
		// time.Time 类型
		if !IsReallyValidType(time.Time{}) {
			t.Error("time.Time 应该是合法类型")
		}
		if !IsReallyValidType(time.Now()) {
			t.Error("time.Time 非零值应该是合法类型")
		}

		// 所有字段类型都合法的结构体
		type AllValidFields struct {
			Int       int
			String    string
			TimeField time.Time
			IntSlice  []int
			StringMap map[string]string
			Nested    struct{ X int }
		}
		if !IsReallyValidType(AllValidFields{}) {
			t.Error("所有字段类型都合法的结构体应该是合法类型")
		}

		// 含有非法类型字段的结构体（导出字段）
		type HasInvalidField struct {
			Valid   int
			Invalid chan int
		}
		if IsReallyValidType(HasInvalidField{}) {
			t.Error("含有非法类型导出字段的结构体应该是非法类型")
		}

		// 含有非法类型字段的结构体（未导出字段）
		type HasInvalidUnexportedField struct {
			Valid   int
			invalid chan int
		}
		if !IsReallyValidType(HasInvalidUnexportedField{}) {
			t.Error("含有非法类型未导出字段的结构体应该是合法类型（不检查未导出字段）")
		}

		// 含有特殊合法类型字段的结构体
		type HasSpecialField struct {
			TimeField time.Time
		}
		if !IsReallyValidType(HasSpecialField{}) {
			t.Error("含有time.Time字段的结构体应该是合法类型")
		}

		// 无导出字段的结构体
		type NoExportedFields struct {
			field int
		}
		if IsReallyValidType(NoExportedFields{}) {
			t.Error("无导出字段的结构体应该是非法类型")
		}

		// 嵌套结构体
		type ValidNested struct {
			Field  int
			Nested struct {
				SubField string
			}
		}
		if !IsReallyValidType(ValidNested{}) {
			t.Error("嵌套合法结构体应该是合法类型")
		}

		type InvalidNested struct {
			Field  int
			Nested struct {
				SubField chan int
			}
		}
		if IsReallyValidType(InvalidNested{}) {
			t.Error("嵌套非法结构体应该是非法类型")
		}
	})

	// 指针类型测试
	t.Run("指针类型测试", func(t *testing.T) {
		// nil 指针
		var intPtr *int = nil
		if IsReallyValidType(intPtr) {
			t.Error("nil 指针应该是非法类型")
		}

		// 指向基本类型的非 nil 指针
		intVal := 42
		if !IsReallyValidType(&intVal) {
			t.Error("指向基本类型的非 nil 指针应该是合法类型")
		}

		// 指向切片的指针
		sliceVal := []string{"a", "b"}
		if !IsReallyValidType(&sliceVal) {
			t.Error("指向切片的指针应该是合法类型")
		}

		// 指向非法类型的指针
		chanVal := make(chan int)
		if IsReallyValidType(&chanVal) {
			t.Error("指向非法类型的指针应该是非法类型")
		}

		// 多层指针
		validPtr := &intVal
		doublePtr := &validPtr
		if !IsReallyValidType(doublePtr) {
			t.Error("多层嵌套的有效指针应该是合法类型")
		}

		invalidPtr := &chanVal
		doubleInvalidPtr := &invalidPtr
		if IsReallyValidType(doubleInvalidPtr) {
			t.Error("多层嵌套的无效指针应该是非法类型")
		}
	})

	// 接口类型测试
	t.Run("接口类型测试", func(t *testing.T) {
		// nil 接口
		var nilInterface interface{} = nil
		if IsReallyValidType(nilInterface) {
			t.Error("nil 接口应该是非法类型")
		}

		// 包含基本类型的接口
		var intInterface interface{} = 42
		if !IsReallyValidType(intInterface) {
			t.Error("包含有效值的接口应该是合法类型")
		}

		// 包含切片的接口
		var sliceInterface interface{} = []int{1, 2, 3}
		if !IsReallyValidType(sliceInterface) {
			t.Error("包含切片的接口应该是合法类型")
		}

		// 包含非法类型的接口
		var chanInterface interface{} = make(chan int)
		if IsReallyValidType(chanInterface) {
			t.Error("包含 channel 的接口应该是非法类型")
		}

		// 包含有嵌套非法类型的接口
		var nestedInvalidInterface interface{} = []chan int{make(chan int)}
		if IsReallyValidType(nestedInvalidInterface) {
			t.Error("包含嵌套非法类型的接口应该是非法类型")
		}
	})

	// 函数中特别举例的示例
	t.Run("函数文档中的示例", func(t *testing.T) {
		// []*map[int]struct{A string} 合法（slice->指针->map[int]->struct）
		complexValid := []*map[int]struct{ A string }{
			{1: {A: "test"}},
		}
		if !IsReallyValidType(complexValid) {
			t.Error("[]*map[int]struct{A string} 应该是合法类型")
		}

		// map[float64]chan struct{} 非法（Value类型含chan）
		invalidMap := map[float64]chan struct{}{
			1.0: make(chan struct{}),
		}
		if IsReallyValidType(invalidMap) {
			t.Error("map[float64]chan struct{} 应该是非法类型")
		}

		// struct{ B int; c []func() } 非法（c字段类型为func且未导出）
		type ComplexInvalid struct {
			B int
			c []func()
		}
		complexInvalid := ComplexInvalid{B: 1, c: []func(){func() {}}}
		// 注意：根据规则，未导出字段不会被检查，所以这个结构体实际是合法的
		if !IsReallyValidType(complexInvalid) {
			t.Error("struct{ B int; c []func() } 应该是合法类型（未导出字段不检查）")
		}

		// struct{ X time.Time } 合法（导出字段类型为特例）
		type HasTimeField struct {
			X time.Time
		}
		if !IsReallyValidType(HasTimeField{}) {
			t.Error("struct{ X time.Time } 应该是合法类型")
		}
	})
}
