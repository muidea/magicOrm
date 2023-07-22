package util

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

// Define the Type enum
const (
	// TypeBooleanValue bool
	TypeBooleanValue = iota + 100
	// TypeBitValue int8
	TypeBitValue
	// TypeSmallIntegerValue int16
	TypeSmallIntegerValue
	// TypeInteger32Value int32
	TypeInteger32Value
	// TypeIntegerValue int
	TypeIntegerValue
	// TypeBigIntegerValue int64
	TypeBigIntegerValue
	// TypePositiveBitValue uint8
	TypePositiveBitValue
	// TypePositiveSmallIntegerValue uint16
	TypePositiveSmallIntegerValue
	// TypePositiveInteger32Value uint32
	TypePositiveInteger32Value
	// TypePositiveIntegerValue uint
	TypePositiveIntegerValue
	// TypePositiveBigIntegerValue uint64
	TypePositiveBigIntegerValue
	// TypeFloatValue float32
	TypeFloatValue
	// TypeDoubleValue float64
	TypeDoubleValue
	// TypeStringValue string
	TypeStringValue
	// TypeDateTimeValue time.Time
	TypeDateTimeValue
	// TypeStructValue struct
	TypeStructValue
	// TypeSliceValue slice
	TypeSliceValue
	// TypeMapValue map
	TypeMapValue
)

func IsInteger(tType reflect.Type) bool {
	switch tType.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		return true
	}

	return false
}

func IsUInteger(tType reflect.Type) bool {
	switch tType.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		return true
	}

	return false
}

func IsFloat(tType reflect.Type) bool {
	switch tType.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	}

	return false
}

func IsBool(tType reflect.Type) bool {
	return tType.Kind() == reflect.Bool
}

func IsString(tType reflect.Type) bool {
	return tType.Kind() == reflect.String
}

func IsDateTime(tType reflect.Type) bool {
	return tType.String() == "time.Time"
}

func IsSlice(tType reflect.Type) bool {
	return tType.Kind() == reflect.Slice
}

func IsStruct(tType reflect.Type) bool {
	return tType.Kind() == reflect.Struct
}

func IsMap(tType reflect.Type) bool {
	return tType.Kind() == reflect.Map
}

func IsBasicType(typeValue int) bool {
	return typeValue < TypeStructValue
}

func IsStructType(typeValue int) bool {
	return typeValue == TypeStructValue
}

func IsSliceType(typeValue int) bool {
	return typeValue == TypeSliceValue
}

func IsMapType(typeVal int) bool {
	return typeVal == TypeMapValue
}

// GetTypeEnum return field type as type constant from reflect.Value
func GetTypeEnum(val reflect.Type) (ret int, err error) {
	switch val.Kind() {
	case reflect.Int8:
		ret = TypeBitValue
	case reflect.Uint8:
		ret = TypePositiveBitValue
	case reflect.Int16:
		ret = TypeSmallIntegerValue
	case reflect.Uint16:
		ret = TypePositiveSmallIntegerValue
	case reflect.Int32:
		ret = TypeInteger32Value
	case reflect.Uint32:
		ret = TypePositiveInteger32Value
	case reflect.Int64:
		ret = TypeBigIntegerValue
	case reflect.Uint64:
		ret = TypePositiveBigIntegerValue
	case reflect.Int:
		ret = TypeIntegerValue
	case reflect.Uint:
		ret = TypePositiveIntegerValue
	case reflect.Float32:
		ret = TypeFloatValue
	case reflect.Float64:
		ret = TypeDoubleValue
	case reflect.Bool:
		ret = TypeBooleanValue
	case reflect.String:
		ret = TypeStringValue
	case reflect.Struct:
		switch val.String() {
		case "time.Time":
			ret = TypeDateTimeValue
		default:
			ret = TypeStructValue
		}
	case reflect.Slice:
		eType := val.Elem()
		if eType.Kind() == reflect.Ptr {
			eType = eType.Elem()
		}
		_, err = GetTypeEnum(eType)
		if err != nil {
			return
		}

		ret = TypeSliceValue
	default:
		err = fmt.Errorf("unsupport type:%v", val.String())
	}

	return
}

// IsNil check value if nil
func IsNil(val reflect.Value) (ret bool) {
	val = reflect.Indirect(val)
	if val.Kind() == reflect.Interface {
		val = reflect.Indirect(val.Elem())
	}

	switch val.Kind() {
	case reflect.Interface:
		ret = val.IsNil()
	case reflect.Slice, reflect.Map:
		ret = false
	case reflect.Invalid:
		ret = true
	default:
		ret = false
	}

	return
}

// isSameStruct check if same
func isSameStruct(firstVal, secondVal reflect.Value) (ret bool, err error) {
	firstNum := firstVal.NumField()
	secondNum := secondVal.NumField()
	if firstNum != secondNum {
		ret = false
		return
	}

	for idx := 0; idx < firstNum; idx++ {
		firstField := firstVal.Field(idx)
		secondField := secondVal.Field(idx)
		ret, err = IsSameVal(firstField, secondField)
		if !ret || err != nil {
			ret = false
			return
		}
	}

	ret = true
	return
}

// IsSameVal is same value
func IsSameVal(firstVal, secondVal reflect.Value) (ret bool, err error) {
	ret = firstVal.Type().String() == secondVal.Type().String()
	if !ret {
		return
	}

	firstIsNil := IsNil(firstVal)
	secondIsNil := IsNil(secondVal)
	if firstIsNil != secondIsNil {
		ret = false
		return
	}
	if firstIsNil {
		ret = true
		return
	}
	firstVal = reflect.Indirect(firstVal)
	secondVal = reflect.Indirect(secondVal)
	typeVal, typeErr := GetTypeEnum(firstVal.Type())
	if typeErr != nil {
		err = typeErr
		ret = false
		return
	}

	if IsStructType(typeVal) {
		ret, err = isSameStruct(firstVal, secondVal)
		return
	}

	if IsBasicType(typeVal) {
		switch typeVal {
		case TypeBooleanValue:
			ret = firstVal.Bool() == secondVal.Bool()
		case TypeStringValue:
			ret = firstVal.String() == secondVal.String()
		case TypeBitValue, TypeSmallIntegerValue, TypeInteger32Value, TypeIntegerValue, TypeBigIntegerValue:
			ret = firstVal.Int() == secondVal.Int()
		case TypePositiveBitValue, TypePositiveSmallIntegerValue, TypePositiveInteger32Value, TypePositiveIntegerValue, TypePositiveBigIntegerValue:
			ret = firstVal.Uint() == secondVal.Uint()
		case TypeFloatValue, TypeDoubleValue:
			ret = math.Abs(firstVal.Float()-secondVal.Float()) <= 0.0001
		case TypeDateTimeValue:
			ret = firstVal.Interface().(time.Time).Sub(secondVal.Interface().(time.Time)) == 0
		default:
			ret = false
			err = fmt.Errorf("illegal value, is a struct value")
		}

		return
	}

	ret = firstVal.Len() == secondVal.Len()
	if !ret {
		return
	}

	for idx := 0; idx < firstVal.Len(); idx++ {
		firstItem := firstVal.Index(idx)
		secondItem := secondVal.Index(idx)
		ret, err = IsSameVal(firstItem, secondItem)
		if !ret || err != nil {
			ret = false
			return
		}
	}

	return
}

// AddSlashes() 函数返回在预定义字符之前添加反斜杠的字符串。
// 预定义字符是：
// 单引号（'）
// 双引号（"）
// 反斜杠（\）
func AddSlashes(str string) string {
	tmpRune := []rune{}
	strRune := []rune(str)
	for _, ch := range strRune {
		switch ch {
		case []rune{'\\'}[0], []rune{'"'}[0], []rune{'\''}[0]:
			tmpRune = append(tmpRune, []rune{'\\'}[0])
			tmpRune = append(tmpRune, ch)
		default:
			tmpRune = append(tmpRune, ch)
		}
	}
	return string(tmpRune)
}

// StripSlashes() 函数删除由 AddSlashes() 函数添加的反斜杠。
func StripSlashes(str string) string {
	dstRune := []rune{}
	strRune := []rune(str)
	strLength := len(strRune)
	for i := 0; i < strLength; i++ {
		if strRune[i] == []rune{'\\'}[0] {
			i++
		}
		dstRune = append(dstRune, strRune[i])
	}
	return string(dstRune)
}
