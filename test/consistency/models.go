package consistency

import (
	"time"
)

type BasicTypes struct {
	ID     int       `orm:"id key auto"`
	Bool   bool      `orm:"bool"`
	Int8   int8      `orm:"int8"`
	Int16  int16     `orm:"int16"`
	Int32  int32     `orm:"int32"`
	Int64  int64     `orm:"int64"`
	Int    int       `orm:"int"`
	UInt8  uint8     `orm:"uint8"`
	UInt16 uint16    `orm:"uint16"`
	UInt32 uint32    `orm:"uint32"`
	UInt64 uint64    `orm:"uint64"`
	UInt   uint      `orm:"uint"`
	Float  float32   `orm:"float"`
	Double float64   `orm:"double"`
	Str    string    `orm:"str"`
	Time   time.Time `orm:"time"`
}

type PointerTypes struct {
	ID   int        `orm:"id key auto"`
	Bool *bool      `orm:"bool"`
	Int  *int       `orm:"int"`
	Str  *string    `orm:"str"`
	Time *time.Time `orm:"time"`
}

type SliceTypes struct {
	ID      int         `orm:"id key auto"`
	Bools   []bool      `orm:"bools"`
	Ints    []int       `orm:"ints"`
	Int8s   []int8      `orm:"int8s"`
	Int16s  []int16     `orm:"int16s"`
	Int32s  []int32     `orm:"int32s"`
	Int64s  []int64     `orm:"int64s"`
	UInts   []uint      `orm:"uints"`
	UInt8s  []uint8     `orm:"uint8s"`
	UInt16s []uint16    `orm:"uint16s"`
	UInt32s []uint32    `orm:"uint32s"`
	UInt64s []uint64    `orm:"uint64s"`
	Floats  []float32   `orm:"floats"`
	Doubles []float64   `orm:"doubles"`
	Strs    []string    `orm:"strs"`
	Times   []time.Time `orm:"times"`
}

type SlicePointerTypes struct {
	ID      int          `orm:"id key auto"`
	BoolPtr *[]bool      `orm:"boolPtr"`
	IntPtr  *[]int       `orm:"intPtr"`
	StrPtr  *[]string    `orm:"strPtr"`
	TimePtr *[]time.Time `orm:"timePtr"`
}

type NestedChild struct {
	ID   int    `orm:"id key auto"`
	Name string `orm:"name"`
}

type NestedParent struct {
	ID    int          `orm:"id key auto"`
	Name  string       `orm:"name"`
	Child *NestedChild `orm:"child"`
}

type NestedItem struct {
	ID    int    `orm:"id key auto"`
	Value string `orm:"value"`
}

type NestedSliceParent struct {
	ID    int          `orm:"id key auto"`
	Name  string       `orm:"name"`
	Items []NestedItem `orm:"items"`
}

// NestedSlicePtrParent 用于测试成员属性为 []*T（指针切片）的 Local↔Remote 转换与往返一致性。不考虑 Children 元素为 nil 的场景。
type NestedSlicePtrParent struct {
	ID       int            `orm:"id key auto"`
	Name     string         `orm:"name"`
	Children []*NestedChild `orm:"children"`
}

type DeepLevel1 struct {
	ID    int    `orm:"id key auto"`
	Value string `orm:"value"`
}

type DeepLevel2 struct {
	ID    int         `orm:"id key auto"`
	Level *DeepLevel1 `orm:"level"`
}

type DeepLevel3 struct {
	ID    int         `orm:"id key auto"`
	Level *DeepLevel2 `orm:"level"`
}

type ComplexEntity struct {
	ID       int          `orm:"id key auto"`
	Name     string       `orm:"name"`
	Count    *int         `orm:"count"`
	Flags    []bool       `orm:"flags"`
	Numbers  []int        `orm:"numbers"`
	Child    *NestedChild `orm:"child"`
	Items    []NestedItem `orm:"items"`
	CreateAt time.Time    `orm:"createAt"`
}

type AllInOne struct {
	ID           int           `orm:"id key auto"`
	Bool         bool          `orm:"bool"`
	BoolPtr      *bool         `orm:"boolPtr"`
	Int          int           `orm:"int"`
	IntPtr       *int          `orm:"intPtr"`
	Str          string        `orm:"str"`
	StrPtr       *string       `orm:"strPtr"`
	BoolSlice    []bool        `orm:"boolSlice"`
	IntSlice     []int         `orm:"intSlice"`
	StrSlice     []string      `orm:"strSlice"`
	BoolSlicePtr *[]bool       `orm:"boolSlicePtr"`
	IntSlicePtr  *[]int        `orm:"intSlicePtr"`
	Child        *NestedChild  `orm:"child"`
	Children     []NestedChild `orm:"children"`
	Timestamp    time.Time     `orm:"timestamp"`
	TimePtr      *time.Time    `orm:"timePtr"`
	TimeSlice    []time.Time   `orm:"timeSlice"`
}

func NewBasicTypes() *BasicTypes {
	return &BasicTypes{
		ID:     1,
		Bool:   true,
		Int8:   -8,
		Int16:  -16,
		Int32:  -32,
		Int64:  -64,
		Int:    -100,
		UInt8:  8,
		UInt16: 16,
		UInt32: 32,
		UInt64: 64,
		UInt:   100,
		Float:  3.14,
		Double: 3.14159265358979,
		Str:    "hello world",
		Time:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
}

func NewPointerTypes() *PointerTypes {
	bVal := true
	iVal := 42
	sVal := "pointer test"
	tVal := time.Date(2024, 2, 20, 14, 0, 0, 0, time.UTC)

	return &PointerTypes{
		ID:   1,
		Bool: &bVal,
		Int:  &iVal,
		Str:  &sVal,
		Time: &tVal,
	}
}

func NewSliceTypes() *SliceTypes {
	return &SliceTypes{
		ID:      1,
		Bools:   []bool{true, false, true},
		Ints:    []int{1, 2, 3},
		Int8s:   []int8{-1, 0, 1},
		Int16s:  []int16{-10, 0, 10},
		Int32s:  []int32{-100, 0, 100},
		Int64s:  []int64{-1000, 0, 1000},
		UInts:   []uint{1, 2, 3},
		UInt8s:  []uint8{1, 2, 3},
		UInt16s: []uint16{10, 20, 30},
		UInt32s: []uint32{100, 200, 300},
		UInt64s: []uint64{1000, 2000, 3000},
		Floats:  []float32{1.1, 2.2, 3.3},
		Doubles: []float64{1.11, 2.22, 3.33},
		Strs:    []string{"a", "b", "c"},
		Times:   []time.Time{time.Now(), time.Now().Add(time.Hour)},
	}
}

func NewNestedParent() *NestedParent {
	return &NestedParent{
		ID:   1,
		Name: "parent",
		Child: &NestedChild{
			ID:   10,
			Name: "child",
		},
	}
}

func NewNestedSliceParent() *NestedSliceParent {
	return &NestedSliceParent{
		ID:   1,
		Name: "slice parent",
		Items: []NestedItem{
			{ID: 1, Value: "item1"},
			{ID: 2, Value: "item2"},
			{ID: 3, Value: "item3"},
		},
	}
}

func NewNestedSlicePtrParent() *NestedSlicePtrParent {
	return &NestedSlicePtrParent{
		ID:   1,
		Name: "slice ptr parent",
		Children: []*NestedChild{
			{ID: 1, Name: "child1"},
			{ID: 2, Name: "child2"},
			{ID: 3, Name: "child3"},
			{ID: 4, Name: "child4"},
		},
	}
}

func NewDeepLevel3() *DeepLevel3 {
	return &DeepLevel3{
		ID: 1,
		Level: &DeepLevel2{
			ID: 2,
			Level: &DeepLevel1{
				ID:    3,
				Value: "deepest",
			},
		},
	}
}

func NewComplexEntity() *ComplexEntity {
	count := 100
	return &ComplexEntity{
		ID:       1,
		Name:     "complex",
		Count:    &count,
		Flags:    []bool{true, false, true},
		Numbers:  []int{1, 2, 3, 4, 5},
		Child:    &NestedChild{ID: 10, Name: "nested child"},
		Items:    []NestedItem{{ID: 1, Value: "v1"}, {ID: 2, Value: "v2"}},
		CreateAt: time.Date(2024, 3, 15, 9, 30, 0, 0, time.UTC),
	}
}

func NewAllInOne() *AllInOne {
	bVal := true
	iVal := 999
	sVal := "all in one string"
	bsVal := []bool{true, false}
	isVal := []int{1, 2, 3}
	tVal := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	return &AllInOne{
		ID:           1,
		Bool:         true,
		BoolPtr:      &bVal,
		Int:          100,
		IntPtr:       &iVal,
		Str:          "test string",
		StrPtr:       &sVal,
		BoolSlice:    []bool{true, false, true},
		IntSlice:     []int{10, 20, 30},
		StrSlice:     []string{"x", "y", "z"},
		BoolSlicePtr: &bsVal,
		IntSlicePtr:  &isVal,
		Child:        &NestedChild{ID: 1, Name: "all in one child"},
		Children: []NestedChild{
			{ID: 1, Name: "child1"},
			{ID: 2, Name: "child2"},
		},
		Timestamp: time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC),
		TimePtr:   &tVal,
		TimeSlice: []time.Time{
			time.Date(2024, 5, 2, 10, 0, 0, 0, time.UTC),
			time.Date(2024, 5, 3, 10, 0, 0, 0, time.UTC),
		},
	}
}
