package utils

import (
	"fmt"
	"reflect"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicOrm/models"
)

// 基本数值类型辅助函数

// 基本数值类型定义如下：
// 1. 基本数值类型：bool, int8, int16, int32, int, int64, uint8, uint16, uint32, uint, uint64, float32, float64, string
// 2. 基本数值类型对应的指针类型：*bool, *int8, *int16, *int32, *int, *int64, *uint8, *uint16, *uint32, *uint, *uint64, *float32, *float64, *string
// 3. 基本数值类型的slice或array: []bool, []int8, []int16, []int32, []int, []int64, []uint8, []uint16, []uint32, []uint, []uint64, []float32, []float64, []string
// 4. 基本数值类型指针的slice或array: []*bool, []*int8, []*int16, []*int32, []*int, []*int64, []*uint8, []*uint16, []*uint32, []*uint, []*uint64, []*float32, []*float64, []*string
// 5. time.Time 类型

// IsReallyValidValue 判断判断一个基本数值是否合法
// 1. 必须是是合法的基本数值类型,否则返回false
// 2. 如果是指针类型，则该指针指向的实际值的类型也必须是合法的基本数值类型，并且该指针已经赋值，否则返回false
// 3. 如果是slice/array, 则该slice/array的item类型也必须是合法的基本数值类型，并且该slice已经初始化， 否则返回false
func IsReallyValidValue(val any) bool {
	if val == nil {
		return false
	}

	vVal := reflect.ValueOf(val)
	return IsReallyValidValueForReflect(vVal)
}

func IsReallyValidValueForReflect(vVal reflect.Value) bool {
	switch vVal.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Array:
		return true
	case reflect.Ptr:
		if vVal.IsNil() {
			return false
		}
		return IsReallyValidValueForReflect(vVal.Elem())
	case reflect.Struct:
		if vVal.Type().String() == models.TypeStructTimeName {
			return true
		}
		if !vVal.IsValid() {
			return true
		}
		log.Warnf("IsReallyValidTypeForReflect failed, unsupported type:%s", vVal.Type().String())
	case reflect.Slice:
		if vVal.IsNil() {
			// 未初始化认为不合法
			return false
		}
		if vVal.Len() == 0 {
			return true
		}
		// 只要判断对应的类型合法就行认为合法
		return IsReallyValidTypeForReflect(vVal.Type())
	}

	return false
}

// IsReallyZeroValue 判断一个基本数值类型是否为零值
// 1. 必须是合法的基本数值类型
// 2. 如果是指针类型，则判断该指针指向的实际值是否为零值
func IsReallyZeroValue(val any) bool {
	if val == nil {
		return true
	}

	vVal := reflect.ValueOf(val)
	return IsReallyZeroForReflect(vVal)
}

func IsReallyZeroForReflect(vVal reflect.Value) bool {
	// 1. 基础合法性检查
	if !vVal.IsValid() {
		return true
	}

	// 2. 指针穿透处理：递归获取指针指向的最底层内容
	// 逻辑：如果是指针，只要它是 nil 或者指向的内容是“零值”，就返回 true
	if vVal.Kind() == reflect.Ptr {
		if vVal.IsNil() {
			return true
		}
		return IsReallyZeroForReflect(vVal.Elem())
	}

	// 3. 特殊容器类型处理
	// 对于 Slice, Map, Chan，业务上通常认为长度为 0 即为零值
	switch vVal.Kind() {
	case reflect.Slice, reflect.Map, reflect.Chan:
		return vVal.Len() == 0
	}

	// 4. 通用零值检查
	// reflect.Value.IsZero() 已经内置了对以下类型的优化处理：
	// - 基础类型 (int, float, bool, string)
	// - 结构体 (递归检查所有字段)
	// - 接口 (检查是否为 nil)
	// - time.Time (它会自动调用 time.IsZero() 方法)
	return vVal.IsZero()
}

// IsReallyValidType 判断是否时一个合法的基本数值类型
// 1. 必须是合法的基本数值类型
func IsReallyValidType(val any) bool {
	if val == nil {
		return false
	}

	vType := reflect.TypeOf(val)
	return IsReallyValidTypeForReflect(vType)
}

func IsReallyValidTypeForReflect(vType reflect.Type) bool {
	switch vType.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	case reflect.Ptr:
		eType := vType.Elem()
		if eType.Kind() == reflect.Ptr {
			// 不允许**这种形式
			return false
		}
		return IsReallyValidTypeForReflect(eType)
	case reflect.Struct:
		if vType.String() == models.TypeStructTimeName {
			return true
		}
	case reflect.Slice, reflect.Array:
		eType := vType.Elem()
		if eType.Kind() == reflect.Slice || eType.Kind() == reflect.Array {
			// 不允许[][]这种形式
			return false
		}

		if eType.Kind() == reflect.Ptr {
			if eType.Elem().Kind() == reflect.Slice || eType.Elem().Kind() == reflect.Array {
				// 不允许[]*[]这种形式
				return false
			}
		}
		return IsReallyValidTypeForReflect(eType)
	}

	return false
}

// IsReallyNil 判断是否是nil
func IsReallyNil(val any) bool {
	if val == nil {
		return true
	}

	vVal := reflect.ValueOf(val)
	return vVal.IsNil()
}

// DeepCopy 深度复制val的值
func DeepCopy(value any) (any, error) {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		// 解引用指针
		elem := val.Elem()
		copiedElem, err := DeepCopy(elem.Interface())
		if err != nil {
			return nil, err
		}
		// 创建新指针并指向复制的值
		newPtr := reflect.New(reflect.TypeOf(copiedElem))
		newPtr.Elem().Set(reflect.ValueOf(copiedElem))
		return newPtr.Interface(), nil
	}

	if val.Kind() == reflect.Slice {
		// 处理slice类型
		length := val.Len()
		newSlice := reflect.MakeSlice(val.Type(), length, length)
		for i := 0; i < length; i++ {
			elem := val.Index(i)
			copiedElem, err := DeepCopy(elem.Interface())
			if err != nil {
				return nil, err
			}
			newSlice.Index(i).Set(reflect.ValueOf(copiedElem))
		}
		return newSlice.Interface(), nil
	}

	// 处理基本类型
	switch v := value.(type) {
	case bool, int8, int16, int32, int, int64,
		uint8, uint16, uint32, uint, uint64,
		float32, float64, string:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", value)
	}
}

func DeepCopyForReflect(src reflect.Value) reflect.Value {
	if !src.IsValid() {
		return reflect.Value{}
	}

	switch src.Kind() {
	case reflect.Ptr:
		// 解引用指针，递归拷贝指向的值
		orig := src.Elem()
		if orig.IsValid() {
			copy := reflect.New(orig.Type())
			copy.Elem().Set(DeepCopyForReflect(orig))
			return copy
		}
		return reflect.Zero(src.Type())

	case reflect.Interface:
		// 处理接口类型
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		valueCopy := DeepCopyForReflect(src.Elem())
		return valueCopy.Convert(src.Type())

	case reflect.Struct:
		// 递归拷贝结构体字段
		dest := reflect.New(src.Type()).Elem()
		for i := 0; i < src.NumField(); i++ {
			if dest.Field(i).CanSet() {
				dest.Field(i).Set(DeepCopyForReflect(src.Field(i)))
			}
		}
		return dest

	case reflect.Slice:
		// 处理切片
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		dest := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
		for i := 0; i < src.Len(); i++ {
			dest.Index(i).Set(DeepCopyForReflect(src.Index(i)))
		}
		return dest

	case reflect.Map:
		// 处理map
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		dest := reflect.MakeMapWithSize(src.Type(), src.Len())
		for _, key := range src.MapKeys() {
			keyCopy := DeepCopyForReflect(key)
			valueCopy := DeepCopyForReflect(src.MapIndex(key))
			dest.SetMapIndex(keyCopy, valueCopy)
		}
		return dest

	default:
		// 基础类型直接拷贝
		if src.CanInterface() {
			return reflect.ValueOf(src.Interface())
		}
		return src
	}
}
