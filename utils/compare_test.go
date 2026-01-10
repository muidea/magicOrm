package utils

import (
	"testing"
)

// 测试用结构体定义
type Test struct {
	ID       int
	Name     string
	Age      int
	Score    float64
	PtrField *int
}

type Test2 struct {
	ID       float64
	Name     string
	Age      uint
	Score    float32
	PtrField *float64
}

func TestCompare(t *testing.T) {
	t.Run("基本类型比较", func(t *testing.T) {
		testCases := []struct {
			a, b  any
			match bool
		}{
			{10, 10.0, true},
			{uint(5), int32(5), true},
			{float32(3.14), 3.14, true},
			{10, "10", false},
		}

		for _, tc := range testCases {
			equal, _ := CompareWithNumericConversion(tc.a, tc.b)
			if equal != tc.match {
				t.Errorf("比较失败: %T(%v) vs %T(%v)", tc.a, tc.a, tc.b, tc.b)
			}
		}
	})

	t.Run("指针比较", func(t *testing.T) {
		a := 10
		b := 10.0
		equal, _ := CompareWithNumericConversion(&a, &b)
		if !equal {
			t.Error("指针值比较失败")
		}
	})
}
