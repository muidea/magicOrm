package utils

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// CompareWithNumericConversion 支持安全类型转换的比较函数
// 返回值：是否相等，差异描述
func CompareWithNumericConversion(a, b interface{}, opts ...cmp.Option) (bool, string) {
	// 防御性代码：处理nil值比较
	if a == nil && b == nil {
		return true, ""
	}
	if a == nil || b == nil {
		return false, fmt.Sprintf("nil comparison: %v vs %v", a, b)
	}

	// 合并默认选项和用户自定义选项
	defaultOpts := cmp.Options{
		cmp.FilterValues(
			func(x, y interface{}) bool {
				// 仅当两个值都是数字类型时应用比较器
				_, xOk := toFloat64(x)
				_, yOk := toFloat64(y)
				return xOk && yOk
			},
			cmp.Comparer(compareNumbers),
		),
		cmpopts.IgnoreTypes(time.Time{}), // 忽略时间字段比较
	}

	mergedOpts := append(defaultOpts, opts...)

	// 执行比较并返回结果
	if equal := cmp.Equal(a, b, mergedOpts); !equal {
		return false, cmp.Diff(a, b, mergedOpts)
	}
	return true, ""
}

// toFloat64 将数字类型统一转换为float64进行比较
func toFloat64(v interface{}) (float64, bool) {
	rv := reflect.ValueOf(v)

	// 处理nil指针
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return 0, false
	}

	// 自动解引用指针
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// 扩展支持的类型
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	default:
		return 0, false
	}
}

// compareNumbers 实现数字类型自动转换比较
func compareNumbers(x, y interface{}) bool {
	xVal, xOk := toFloat64(x)
	yVal, yOk := toFloat64(y)
	if !xOk || !yOk {
		return false
	}
	// 使用近似相等比较，处理浮点数精度问题
	const epsilon = 0.0001
	diff := xVal - yVal
	return math.Abs(diff) < epsilon
}

/*
func main() {
	// 示例用法
	t1 := &Test{ID: 100, Name: "Alice"}
	t2 := &Test2{ID: 100.0, Name: "Alice"}

	if equal, diff := CompareWithNumericConversion(t1, t2); !equal {
		fmt.Println("差异发现:", diff)
	} else {
		fmt.Println("对象相等")
	}
}

*/
