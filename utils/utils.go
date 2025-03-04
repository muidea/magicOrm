package utils

import (
	"fmt"
	"reflect"
	"time"
)

// IsReallyValid 判断一个值是否有效
func IsReallyValid(val interface{}) bool {
	if val == nil {
		return false
	}
	rVal := reflect.ValueOf(val)
	return IsReallyValidForReflect(rVal)
}

// IsReallyValidForReflect 判断一个值是否有效
// 规则说明：
// 1. 指针类型：
//   - 若指针为 nil，返回 false
//   - 若指针非 nil，递归检查指向的值是否有效
//
// 2. 结构体（struct）：
//   - 若为 time.Time 类型，直接返回 true
//   - 其他 struct 必须满足：
//     a. 至少有一个导出字段（首字母大写）
//     b. 所有导出字段的类型必须是以下之一：
//   - 基本类型（int/string 等或其指针）
//   - slice/map/struct/time.Time 或其指针
//   - 嵌套字段需递归检查有效性
//   - 零值 struct（如 MyStruct{}）需检查字段有效性，若所有字段无效则返回 false
//
// 3. 切片（slice）：
//   - 若 slice 为 nil，返回 false
//   - 元素类型必须是以下之一：
//   - 基本类型/slice/map/struct/time.Time 或其指针
//   - 需递归检查每个元素的有效性
//
// 4. 映射（map）：
//   - 若 map 为 nil，返回 false
//   - Key 必须是基本类型（int/string 等或其指针）
//   - Value 类型必须是以下之一：
//   - 基本类型/slice/map/struct/time.Time 或其指针
//   - 需递归检查每个 Value 的有效性
//
// 5. 基本类型：
//   - 包括 int/string/bool 等及其指针类型
//   - 自定义类型（如 type MyInt int）按底层 Kind 判断
//
// 6. 其他类型：
//   - 接口类型（interface{}）直接返回 false
//   - 其他未明确类型（如 chan/func）返回 false
func IsReallyValidForReflect(rVal reflect.Value) bool {
	// 首先检查值是否有效
	if !rVal.IsValid() {
		return false
	}

	switch rVal.Kind() {
	case reflect.Ptr:
		if rVal.IsNil() {
			return false
		}
		return IsReallyValidForReflect(rVal.Elem())
	case reflect.Slice, reflect.Map:
		if rVal.IsNil() {
			return false
		}
		// 检查元素类型是否有效
		return IsReallyValidTypeForReflect(rVal.Type().Elem()) &&
			(rVal.Kind() != reflect.Map || IsReallyValidTypeForReflect(rVal.Type().Key()))
	case reflect.Struct:
		// time.Time 类型特殊处理
		if rVal.Type() == reflect.TypeOf(time.Time{}) {
			return true
		}

		// 检查是否有导出字段
		hasExportedField := false

		for i := 0; i < rVal.NumField(); i++ {
			field := rVal.Type().Field(i)

			if field.IsExported() {
				hasExportedField = true

				// 检查字段类型是否有效
				if !IsReallyValidTypeForReflect(field.Type) {
					return false
				}
			}
		}

		return hasExportedField
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}

// IsReallyZero 判断一个值是否为零值
func IsReallyZero(val interface{}) bool {
	if val == nil {
		return true
	}

	rVal := reflect.ValueOf(val)
	return IsReallyZeroForReflect(rVal)
}

// IsReallyZeroForReflect 判断一个 reflect.Value 是否为零值
// 规则说明：
// 1. 指针类型：
//   - 若指针为 nil，直接返回 true（符合 Go 零值定义）
//   - 若指针非 nil，递归检查指向的值是否为零值
//
// 2. 切片（slice）：
//   - 若 slice 为 nil 或长度为零，返回 true
//   - 否则返回 false
//
// 3. 映射（map）：
//   - 若 map 为 nil 或键值对数量为零，返回 true
//   - 否则返回 false
//
// 4. 结构体（struct）：
//   - 若为 time.Time 类型，调用其 IsZero() 方法判断
//   - 若为 ObjectValue 类型：
//     仅检查 Fields 字段长度为零时返回 true（忽略 ID/Name/PkgPath 字段）
//   - 若为 SliceObjectValue 类型：
//     仅检查 Values 字段长度为零时返回 true（忽略 Name/PkgPath 字段）
//   - 其他 struct 类型：
//     a. 所有导出字段（首字母大写）必须为零值
//     b. 递归检查每个导出字段的零值状态
//     c. 若 struct 无导出字段，直接返回 true
//
// 5. 基本类型：
//   - 直接通过 reflect.Value.IsZero() 判断
//
// 6. 其他类型：
//   - 数组：所有元素为零值时返回 true
//   - 接口类型：解包动态值后递归判断
//   - 函数、通道等非标类型：直接返回 false
//
// 异常处理：
//   - 若遇到无法处理的类型（如 unsafe.Pointer），panic 并输出类型信息
func IsReallyZeroForReflect(rVal reflect.Value) bool {
	// 首先检查值是否有效
	if !rVal.IsValid() {
		return true
	}

	switch rVal.Kind() {
	case reflect.Ptr:
		if rVal.IsNil() {
			return true
		}
		return IsReallyZeroForReflect(rVal.Elem())
	case reflect.Slice, reflect.Map:
		return rVal.IsNil() || rVal.Len() == 0
	case reflect.Array:
		for i := 0; i < rVal.Len(); i++ {
			if !IsReallyZeroForReflect(rVal.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		if rVal.Type() == reflect.TypeOf(time.Time{}) {
			return rVal.Interface().(time.Time).IsZero()
		}
		for i := 0; i < rVal.NumField(); i++ {
			field := rVal.Field(i)
			// 只检查可访问的导出字段
			if field.CanInterface() && !IsReallyZeroForReflect(field) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface:
		return rVal.IsNil()
	case reflect.Bool:
		return !rVal.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rVal.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rVal.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rVal.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return rVal.Complex() == complex(0, 0)
	case reflect.String:
		return rVal.String() == ""
	default:
		panic(fmt.Sprintf("Unsupported type: %v", rVal.Type()))
	}
}

// IsReallyValidTypeForReflect 判断一个reflect.Type是否是合法的类型
// 如果rType是一个指针,检查指针对应的类型是否合法
// 合法类型只能是基本的数据类型，或者是slice,map,struct
// 并且如果是struct， 则要求struct有导出字段或者是time.Time
// 如果rType是一个interface，则检查interface对应的原始类型是否满足上述条件
// 其他类型返回false
func IsReallyValidTypeForReflect(rType reflect.Type) bool {
	// 处理指针类型，递归解引用
	for rType != nil && rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}

	if rType == nil {
		return false
	}

	// 处理接口类型，检查其底层类型
	if rType.Kind() == reflect.Interface {
		// 注意：空接口无法直接获取其元素类型
		// 在实际使用时，应该通过传入具体实现的接口值进行检查
		return false
	}

	// 检查基础类型
	switch rType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String:
		return true

	case reflect.Slice, reflect.Map:
		// 检查元素类型是否合法
		elemValid := IsReallyValidTypeForReflect(rType.Elem())
		if !elemValid {
			return false
		}
		// 对于 map，还需检查键类型是否合法
		if rType.Kind() == reflect.Map {
			return IsReallyValidTypeForReflect(rType.Key())
		}
		return true

	case reflect.Struct:
		// 特殊处理 time.Time 类型
		if isTimeType(rType) {
			return true
		}
		// 检查是否有导出字段
		return hasExportedField(rType)

	default:
		return false
	}
}

// 判断是否是 time.Time 类型
func isTimeType(t reflect.Type) bool {
	return t.PkgPath() == "time" && t.Name() == "Time"
}

// 检查结构体是否有导出字段
func hasExportedField(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			return true
		}
	}
	return false
}

// IsReallyValidType 判断给定值的类型是否属于合法类型集合
// 合法类型定义：
// 1. 基本数据类型：bool、整型(int/uint系列)、浮点型(float32/float64)、string
// 2. 容器类型：
// - slice：递归检查其元素类型是否合法
// - map：Key类型必须是基本数据类型，Value类型需递归检查合法性
// 3. 结构体类型：
// - time.Time 类型直接视为合法（特例）
// - 其他结构体必须满足：
// a. 至少包含一个导出字段（首字母大写）
// b. 所有导出字段的类型需递归检查合法性
// 4. 指针类型：递归检查其指向的原始值类型是否合法
// 5. 接口类型：递归检查其动态值的实际类型是否合法
//
// 特别注意：
// - 下列类型始终非法：
// * 数组(array)、通道(chan)、函数(func)、unsafe.Pointer
// * 包含非法类型嵌套的结构体（如含chan字段的结构体）
// * map的Key类型为非基本类型（如struct/interface等）
// * 任何递归路径中出现非法类型
//
// 示例：
// - []*map[int]struct{A string} 合法（slice->指针->map[int]->struct）
// - map[float64]chan struct{} 非法（Value类型含chan）
// - struct{ B int; c []func() } 非法（c字段类型为func且未导出）
// - struct{ X time.Time } 合法（导出字段类型为特例）
//
// 合法返回true，否则返回false
func IsReallyValidType(val interface{}) bool {
	if val == nil {
		return false
	}

	rType := reflect.TypeOf(val)
	rValue := reflect.ValueOf(val)

	// 特殊处理接口值
	if rType.Kind() == reflect.Interface {
		// 确保接口值有效且非空
		if !rValue.IsValid() || rValue.IsNil() {
			return false
		}

		// 获取接口内的具体值，并递归检查其类型是否合法
		concreteValue := rValue.Elem()
		return IsReallyValidType(concreteValue.Interface())
	}

	// 对于指针类型，需检查指针是否为 nil
	if rType.Kind() == reflect.Ptr {
		if rValue.IsNil() {
			return false
		}
		// 递归检查指针指向的值类型是否合法
		return IsReallyValidType(rValue.Elem().Interface())
	}

	// 基本类型检查
	switch rType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String:
		return true

	case reflect.Slice:
		// 递归检查元素类型是否合法
		elemType := rType.Elem()
		return isTypeValid(elemType)

	case reflect.Map:
		// 检查键类型是否为基本数据类型
		keyType := rType.Key()
		keyValid := isBasicType(keyType)
		if !keyValid {
			return false
		}
		
		// 递归检查值类型是否合法
		valueType := rType.Elem()
		return isTypeValid(valueType)

	case reflect.Struct:
		// 特殊处理 time.Time 类型
		if isTimeType(rType) {
			return true
		}
		
		// 检查结构体是否有导出字段，且所有导出字段类型是否合法
		hasExported := false
		for i := 0; i < rType.NumField(); i++ {
			field := rType.Field(i)
			if field.IsExported() {
				hasExported = true
				if !isTypeValid(field.Type) {
					return false
				}
			}
		}
		return hasExported

	default:
		// 其他类型视为非法（如 array, chan, func 等）
		return false
	}
}

// isTypeValid 检查类型是否合法（辅助函数，通过反射类型判断）
func isTypeValid(rType reflect.Type) bool {
	// 处理指针类型，递归检查
	if rType.Kind() == reflect.Ptr {
		return isTypeValid(rType.Elem())
	}

	// 基本类型检查
	switch rType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String:
		return true

	case reflect.Slice:
		// 递归检查元素类型
		return isTypeValid(rType.Elem())

	case reflect.Map:
		// 检查键类型是否为基本数据类型
		keyValid := isBasicType(rType.Key())
		if !keyValid {
			return false
		}
		// 递归检查值类型
		return isTypeValid(rType.Elem())

	case reflect.Struct:
		// 特殊处理 time.Time 类型
		if isTimeType(rType) {
			return true
		}
		
		// 检查是否有导出字段，且所有导出字段类型是否合法
		hasExported := false
		for i := 0; i < rType.NumField(); i++ {
			field := rType.Field(i)
			if field.IsExported() {
				hasExported = true
				if !isTypeValid(field.Type) {
					return false
				}
			}
		}
		return hasExported

	default:
		// 其他类型视为非法（如 array, chan, func 等）
		return false
	}
}

// isBasicType 检查是否为基本数据类型
func isBasicType(rType reflect.Type) bool {
	switch rType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}
